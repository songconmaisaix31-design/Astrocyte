package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type WorksEventSink struct {
	events *attentionapp.EventService
	clock  func() time.Time
}

var _ workspaceapp.WorksEventSink = (*WorksEventSink)(nil)

func NewWorksEventSink(events *attentionapp.EventService, clock func() time.Time) *WorksEventSink {
	if clock == nil {
		clock = time.Now
	}
	return &WorksEventSink{events: events, clock: clock}
}

func (s *WorksEventSink) EmitWorkEvent(ctx context.Context, v workspacedomain.WorkEvent) error {
	if v.WorkID == "" || v.Kind == "" || v.Source == "" {
		return fmt.Errorf("work event source identity required")
	}
	// Serialize the source envelope without compacting the original payload.
	payload := v.Payload
	if len(payload) == 0 {
		payload = json.RawMessage("null")
	}
	if !json.Valid(payload) {
		return fmt.Errorf("invalid work source payload")
	}
	v.Payload = nil
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// Payload is the final field in WorkEvent's source contract.
	const suffix = `"payload":null}`
	raw = append(raw[:len(raw)-len(suffix)], []byte(`"payload":`)...)
	raw = append(raw, payload...)
	raw = append(raw, '}')
	revision := str(object(payload), "revision")
	if n := object(payload)["revision"]; len(n) > 0 {
		revision = string(n)
	}
	key := "works:" + v.WorkID + ":" + v.Kind + ":" + v.At.UTC().Format(time.RFC3339Nano) + ":revision:" + revision
	ts := int64(0)
	if !v.At.IsZero() {
		ts = v.At.UnixMilli()
	}
	// Works metadata has its own vocabulary. Preserve it pending rather than
	// converting park/done, sync or verify into user veto/task completion.
	record := attentionapp.StoredEvent{OriginKey: key, ReceivedAt: s.clock().UnixMilli(), RawBytes: raw,
		Observation: attentionapp.Observation{Source: v.Source, SourceKind: v.Kind, Scope: v.Scope, TS: ts, Outcome: "unknown", Depth: "unknown", Provenance: "workspace:works"}}
	_, err = s.events.Record(ctx, []attentionapp.StoredEvent{record})
	return err
}
