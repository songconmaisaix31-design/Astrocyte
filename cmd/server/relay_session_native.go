package main

import (
	"context"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (c *relaySessionCollector) nativeSources(ctx context.Context) (map[string]domain.RelayNativeSessionBatch, error) {
	result := map[string]domain.RelayNativeSessionBatch{}
	reader, ok := c.owned.(workspace.RelayOwnedNativeSessionReader)
	if !ok {
		return result, nil
	}
	list, err := reader.ListOwnedRelayNativeSessions(ctx)
	if err != nil {
		return nil, errors.New("owned native sources unavailable")
	}
	for _, source := range list {
		id := source.OwnedSession.TaskID
		if id == "" {
			return nil, errors.New("owned native source identity invalid")
		}
		if _, duplicate := result[id]; duplicate {
			return nil, errors.New("owned native source identity duplicated")
		}
		result[id] = source
	}
	return result, nil
}

func relayNativeCaptureBatch(source domain.RelayNativeSessionBatch) events.RelaySessionNativeBatch {
	owned := source.OwnedSession
	binding := relayCaptureBinding(owned)
	binding.Format, binding.Authorized = owned.Agent, source.SourceStatus == "available" && owned.Gap != "permission_denied"
	result := events.RelaySessionNativeBatch{Binding: binding, BindingID: owned.BindingID, LastSequence: source.LastSequence, Errors: source.Errors, Unknown: source.Unknown, PrefixComplete: source.PrefixComplete, Finalized: source.Finalized, FirstGap: source.FirstGap, Events: []events.RelaySessionNativeEvent{}}
	for _, event := range source.Events {
		result.Events = append(result.Events, events.RelaySessionNativeEvent{Sequence: event.Sequence, Kind: event.Kind, Tokens: event.Tokens, TS: event.TS})
	}
	return result
}

func (c *relaySessionCollector) observeNative(ctx context.Context, source domain.RelayNativeSessionBatch) error {
	owned := source.OwnedSession
	entry := c.entries[owned.TaskID]
	prior := events.RelaySessionNativeCursor{}
	if entry.Native != nil {
		prior = *entry.Native
	}
	if prior.Blocked != "" {
		return errors.New("native capture blocked")
	}
	cursor, captureErr := c.capture.CaptureNative(ctx, relayNativeCaptureBatch(source), prior)
	entry.Native = &cursor
	entry.Coverage = coverageFor(owned)
	entry.Coverage.Source = "native_notifications"
	entry.Coverage.Lines, entry.Coverage.Projected = source.LastSequence, cursor.Counts.Projected
	entry.Coverage.Unknown, entry.Coverage.Errors, entry.Coverage.UnknownTokens = source.Unknown, max(source.Errors, cursor.Counts.Errors), cursor.Counts.UnknownTokens
	entry.Coverage.SnapshotComplete = captureErr == nil && cursor.Sequence == source.LastSequence
	entry.Coverage.Status = "partial"
	switch {
	case captureErr != nil:
		entry.Coverage.Gap = sessionPublicGap(cursor.FirstError)
	case source.FirstGap != "":
		entry.Coverage.Gap = sessionPublicGap(source.FirstGap)
	case source.Unknown > 0:
		entry.Coverage.Gap = "unknown_records"
	case source.PrefixComplete != "confirmed":
		entry.Coverage.Gap = "native_prefix_unverified"
	case source.Finalized != "confirmed":
		entry.Coverage.Gap = "native_source_not_finalized"
	case owned.ExitedAt == nil:
		entry.Coverage.Gap = "exit_unverified"
	default:
		entry.Coverage.Status, entry.Coverage.Gap, entry.Coverage.SessionComplete = "complete", "", true
	}
	if err := c.commitEntry(owned.TaskID, entry); err != nil {
		return err
	}
	if captureErr != nil {
		return errors.New("native session coverage incomplete")
	}
	return nil
}
