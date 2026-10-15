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
	if c.archiveOnly {
		return result, nil
	}
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
	result.BindingID = ""
	if source.ControllerID == owned.ControllerID && source.StartupOperationID == owned.StartupOperationID {
		result.BindingID, _ = relayNativeConnectionBinding(owned.BindingID, owned.ControllerID, owned.StartupOperationID)
	}
	for _, event := range source.Events {
		result.Events = append(result.Events, events.RelaySessionNativeEvent{Sequence: event.Sequence, Kind: event.Kind, Tokens: event.Tokens, TS: event.TS})
	}
	return result
}

func (c *relaySessionCollector) observeNative(ctx context.Context, source domain.RelayNativeSessionBatch) error {
	owned := source.OwnedSession
	entry := c.entries[owned.TaskID]
	batch := relayNativeCaptureBatch(source)
	if batch.BindingID == "" {
		entry.Coverage = coverageFor(owned)
		entry.Coverage.Source, entry.Coverage.Gap = "native_notifications", "native_controller_binding_unrecorded"
		if err := c.commitEntry(owned.TaskID, entry); err != nil {
			return err
		}
		return errors.New("native controller binding unrecorded")
	}
	prior := events.RelaySessionNativeCursor{}
	if entry.Native != nil {
		if entry.Native.BindingID == batch.BindingID && entry.NativeControllerID == owned.ControllerID && entry.NativeStartupOperationID == owned.StartupOperationID {
			prior = *entry.Native
		} else {
			history := make(map[string]relayNativeCheckpoint, len(entry.NativeHistory)+1)
			for key, checkpoint := range entry.NativeHistory {
				history[key] = checkpoint
			}
			history[entry.Native.BindingID] = relayNativeCheckpoint{*entry.Native, entry.NativeControllerID, entry.NativeStartupOperationID}
			entry.NativeHistory = history
			if checkpoint, exists := history[batch.BindingID]; exists {
				if checkpoint.ControllerID != owned.ControllerID || checkpoint.StartupOperationID != owned.StartupOperationID {
					return errors.New("native historical identity differs")
				}
				prior = checkpoint.Cursor
			}
		}
	}
	if prior.Blocked != "" {
		return errors.New("native capture blocked")
	}
	cursor, captureErr := c.capture.CaptureNative(ctx, batch, prior)
	entry.Native = &cursor
	entry.NativeControllerID, entry.NativeStartupOperationID = owned.ControllerID, owned.StartupOperationID
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
