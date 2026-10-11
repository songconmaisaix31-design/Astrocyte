package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func receipt(origin, source, sourceKind, provenance string, raw []byte, o nativeObservation, receivedAt int64) app.StoredEvent {
	v := app.StoredEvent{OriginKey: origin, ReceivedAt: receivedAt, RawBytes: append([]byte(nil), raw...), Observation: app.Observation{Source: source, SourceKind: sourceKind, Session: o.session, Scope: o.cwd, TS: o.ts, Tokens: o.tokens, Outcome: o.outcome, Depth: o.depth, Provenance: provenance}}
	if o.kind != "unknown" {
		v.Event = &domain.Event{Source: source, Session: o.session, TS: o.ts, Kind: o.kind, Scope: o.cwd, Raw: append(json.RawMessage(nil), raw...)}
		// The §4 integer projection is not a token observation. Nil in the
		// authoritative Observation remains unknown, and tokens do not scale s_i.
		if o.tokens != nil {
			v.Event.Tokens = *o.tokens
		}
	}
	return v
}

func MapS1Outbox(raw json.RawMessage, receivedAt int64) (app.StoredEvent, error) {
	v, err := outboxRecord(raw)
	if err != nil {
		return app.StoredEvent{}, err
	}
	source := domain.EventSourceManual
	if str(object(v.Payload), "actor_kind") == "agent" {
		source = domain.EventSourceMCP
	}
	o := s1Record(v)
	// These are outcomes of the original S1 job, not agent-task outcomes.
	switch v.Type {
	case "attention.job_succeeded":
		o.outcome = "success"
	case "attention.job_failed":
		p := object(v.Payload)
		if string(p["delivery_unknown"]) != "true" && str(object(p["cause"]), "code") != "delivery_unknown" {
			o.outcome = "failure"
		}
	}
	return receipt("s1:outbox:"+v.ID, source, v.Type, "s1:attention_outbox", raw, o, receivedAt), nil
}

func MapCachedMetadata(raw json.RawMessage, receivedAt int64) (app.StoredEvent, error) {
	source, o, err := cachedMetadata(raw)
	if err != nil {
		return app.StoredEvent{}, err
	}
	// A refreshed cache is a new metadata observation, not new session activity.
	// Its original observed_at distinguishes receipts without changing source TS.
	origin := "s1:session:" + source + ":" + o.cwd + ":" + o.session + ":observed:" + str(object(raw), "observed_at")
	return receipt(origin, source, "native_session_header", "s1:local_project_discovery:registered", raw, o, receivedAt), nil
}

// ReplayNativeFile requires the exact separately authorized file and root.
// It is not used for the metadata-only acceptance while raw scope is undecided.
func ReplayNativeFile(ctx context.Context, source, root, path, project string, limit int, receivedAt int64) ([]app.StoredEvent, bool, error) {
	lines, partial, err := readNativeFile(ctx, source, root, path, project, limit)
	if err != nil {
		return nil, partial, err
	}
	result := make([]app.StoredEvent, 0, len(lines))
	for _, line := range lines {
		kind := str(object(line.raw), "type")
		origin := "native:" + source + ":" + line.observation.session + ":line:" + strconv.Itoa(line.line)
		result = append(result, nativeReceipt(origin, source, kind, "native:selected_file", line.raw, line.observation, receivedAt))
	}
	return result, partial, nil
}

func MapNativeRecord(source, origin, project string, raw json.RawMessage, receivedAt int64) (app.StoredEvent, error) {
	return MapAssociatedNativeRecord(source, origin, project, raw, receivedAt, NativeAssociation{})
}

// MapAssociatedNativeRecord accepts only association already observed in this file.
func MapAssociatedNativeRecord(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation) (app.StoredEvent, error) {
	o, err := nativeRecord(source, raw)
	if err != nil {
		return app.StoredEvent{}, err
	}
	if o.cwd != "" && !sameReplayProject(o.cwd, project) {
		return app.StoredEvent{}, fmt.Errorf("native record leaves approved project")
	}
	if association.Scope != "" && !sameReplayProject(association.Scope, project) {
		return app.StoredEvent{}, fmt.Errorf("native association leaves approved project")
	}
	if association.Session != "" && o.session != "" && association.Session != o.session {
		return app.StoredEvent{}, fmt.Errorf("native association changes session")
	}
	if o.session == "" {
		o.session = association.Session
	}
	if o.cwd == "" {
		o.cwd = association.Scope
	}
	if o.cwd != "" {
		o.cwd = project
	}
	return nativeReceipt(origin, source, str(object(raw), "type"), "native:selected_record", raw, o, receivedAt), nil
}
