package main

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relaySessionEntry struct {
	Cursor                   events.RelaySessionCursor
	Coverage                 workbench.SessionCoverage
	Native                   *events.RelaySessionNativeCursor `json:",omitempty"`
	NativeControllerID       string                           `json:",omitempty"`
	NativeStartupOperationID string                           `json:",omitempty"`
	NativeHistory            map[string]relayNativeCheckpoint `json:",omitempty"`
}

type relayNativeCheckpoint struct {
	Cursor             events.RelaySessionNativeCursor
	ControllerID       string
	StartupOperationID string
}
type relaySessionCheckpoints map[string]relaySessionEntry
type relaySessionState interface {
	Load() (relaySessionCheckpoints, error)
	Save(relaySessionCheckpoints) error
}

// Entry-point bridge only: no cross-context app import or second runtime writer.
type relaySessionCollector struct {
	mu         sync.Mutex
	owned      workspace.RelayOwnedSessionReader
	capture    *events.RelaySessionCapture
	state      relaySessionState
	admissions workbench.SessionAdmissions
	entries    relaySessionCheckpoints
	blocked    string
}

func newRelaySessionCollector(owned workspace.RelayOwnedSessionReader, capture *events.RelaySessionCapture, state relaySessionState, admissions workbench.SessionAdmissions) *relaySessionCollector {
	c := &relaySessionCollector{owned: owned, capture: capture, state: state, admissions: admissions, entries: relaySessionCheckpoints{}}
	loaded, err := state.Load()
	if err != nil {
		c.blocked = "checkpoint_unavailable"
	} else if loaded != nil {
		c.entries = loaded
	}
	return c
}

func relayCaptureBinding(v domain.RelayOwnedSession) events.RelaySessionBinding {
	return events.RelaySessionBinding{TaskID: v.TaskID, TerminalID: v.TerminalID, SessionID: v.SessionID, WorktreeID: v.WorktreeID, ProjectID: v.ProjectID, NativeID: v.NativeID, Format: v.ArchiveFormat, Path: v.ArchivePath, WorkspaceRoot: v.WorkspaceRoot, ApprovedRoots: []string{v.ArchiveRoot}, Owned: v.Ownership == "owned", Authorized: v.SourceStatus == "available", ExitConfirmed: v.ExitedAt != nil, ArchiveFinalized: v.ArchiveFinalized == "confirmed"}
}

func sessionPublicGap(code string) string {
	switch code {
	case "native_controller_binding_unrecorded":
		return code
	case "native_source_not_finalized", "native_prefix_unverified", "native_source_unverifiable", "native_sequence_invalid", "native_sequence_gap", "native_source_error", "invalid_native_record", "source_extent_unsupported", "native_notification_source_not_armed", "native_start_metadata_decode_failed", "native_start_identity_mismatch", "native_notification_method_invalid", "native_notification_params_invalid", "native_notification_cwd_mismatch", "native_notification_identity_unverified", "native_notification_before_start_identity", "native_notification_semantics_unknown", "native_notification_after_close_or_capacity", "native_token_usage_invalid", "native_stream_ended_without_shutdown_ack", "native_resume_raw_events_unavailable", "native_shutdown_ack_unavailable", "native_reader_end_unconfirmed":
		return code
	case "archive_metadata_unavailable", "archive_metadata_unsupported", "archive_metadata_unverifiable", "archive_metadata_identity_unverified", "archive_metadata_path_unavailable":
		return code
	case "", "archive_not_selected", "binding_not_recorded", "permission_denied", "workspace_identity_changed", "workspace_unavailable", "session_ownership_unverified", "session_identity_changed", "archive_identity_changed", "prefix_unverified", "append_only_unverified", "source_not_finalized", "exit_unverified", "capture_pending", "unknown_records", "checkpoint_unavailable", "checkpoint_save_unknown", "admission_unknown", "malformed_json", "closed_incomplete_tail", "session_header_unverifiable", "session_or_cwd_changed", "source_replaced", "source_truncated", "checkpoint_binding_mismatch", "source_changed_during_read", "selection_denied", "capture_cancelled", "header_read_failed", "source_read_failed", "source_identity_unverifiable_or_aliased", "source_open_failed", "source_not_regular", "checkpoint_invalid", "checkpoint_not_line_boundary":
		return code
	default:
		return "source_unavailable"
	}
}

func coverageFor(v domain.RelayOwnedSession) workbench.SessionCoverage {
	source := "unknown"
	if v.SourceStatus == "available" && v.ArchiveFormat != "" {
		source = "archive"
	}
	return workbench.SessionCoverage{Source: source, TaskID: v.TaskID, TerminalID: v.TerminalID, WorktreeID: v.WorktreeID, ProjectID: v.ProjectID, Status: "unknown", Gap: sessionPublicGap(v.Gap)}
}

func (c *relaySessionCollector) ObserveOwned(ctx context.Context) error {
	bindings, err := c.owned.ListOwnedRelaySessions(ctx)
	if err != nil {
		return errors.New("session ownership read unavailable")
	}
	nativeSources, err := c.nativeSources(ctx)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.blocked != "" {
		return errors.New("session checkpoints blocked")
	}
	for _, v := range bindings {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if v.Gap == "permission_denied" || v.Gap == "binding_not_recorded" {
			continue
		}
		if native, ok := nativeSources[v.TaskID]; ok && (native.SourceStatus == "available" || c.entries[v.TaskID].Native != nil) {
			if err := c.observeNative(ctx, native); err != nil {
				return err
			}
			continue
		}
		if v.SourceStatus != "available" {
			entry := c.entries[v.TaskID]
			coverage := coverageFor(v)
			if entry.Cursor.FirstError != "" {
				coverage = entry.Coverage
				coverage.Gap = sessionPublicGap(entry.Cursor.FirstError)
				coverage.SessionComplete = false
			}
			entry.Coverage = coverage
			if c.commitEntry(v.TaskID, entry) != nil {
				return errors.New("session checkpoint save unknown")
			}
			continue
		}
		priorEntry := c.entries[v.TaskID]
		entry := priorEntry
		if entry.Cursor.Blocked != "" {
			continue
		} // preserve first unknown; no automatic reconciliation
		result, readErr := c.capture.Capture(ctx, relayCaptureBinding(v), entry.Cursor)
		entry.Cursor = result.Checkpoint
		entry.Coverage = coverageFor(v)
		counts := result.Batch.Next.Counts
		counts.Errors += result.AdmissionErrors
		entry.Coverage.Lines, entry.Coverage.Projected, entry.Coverage.Duplicates = counts.Lines, counts.Projected, counts.Duplicates
		entry.Coverage.Unknown, entry.Coverage.Errors, entry.Coverage.UnknownTokens = counts.Unknown, counts.Errors, counts.UnknownTokens
		entry.Coverage.RemainingBytes, entry.Coverage.TailBytes = result.Batch.Remaining, result.Batch.TailBytes
		entry.Coverage.SnapshotComplete = result.Batch.SnapshotComplete && readErr == nil
		entry.Coverage.SessionComplete = result.Batch.SessionComplete && readErr == nil
		entry.Coverage.Status = "partial"
		switch {
		case readErr != nil:
			entry.Coverage.Gap = sessionPublicGap(entry.Cursor.FirstError)
		case counts.Errors > 0:
			entry.Coverage.Gap = sessionPublicGap(entry.Cursor.FirstError)
		case counts.Unknown > 0:
			entry.Coverage.Gap = "unknown_records"
		case !result.Batch.SnapshotComplete:
			entry.Coverage.Gap = "capture_pending"
		case v.ArchiveAppendOnly != "confirmed":
			entry.Coverage.Gap = "append_only_unverified"
		case !relayCaptureBinding(v).PrefixComplete:
			entry.Coverage.Gap = "prefix_unverified"
		case !relayCaptureBinding(v).ArchiveFinalized:
			entry.Coverage.Gap = "source_not_finalized"
		case v.ExitedAt == nil:
			entry.Coverage.Gap = "exit_unverified"
		default:
			entry.Coverage.Status, entry.Coverage.Gap = "complete", ""
		}
		if c.commitEntry(v.TaskID, entry) != nil {
			return errors.New("session checkpoint save unknown")
		}
		if readErr != nil {
			return errors.New("session coverage incomplete")
		}
	}
	return nil
}

func (c *relaySessionCollector) commitEntry(taskID string, entry relaySessionEntry) error {
	prior, existed := c.entries[taskID]
	if existed && reflect.DeepEqual(prior, entry) {
		return nil
	}
	c.entries[taskID] = entry
	// Admission precedes metadata; silent polls must not create writes.
	if c.state.Save(c.entries) != nil {
		if existed {
			c.entries[taskID] = prior
		} else {
			delete(c.entries, taskID)
		}
		c.blocked = "checkpoint_save_unknown"
		return errors.New("session checkpoint save unknown")
	}
	return nil
}

func (c *relaySessionCollector) ReadSessionEvents(ctx context.Context, q workbench.SessionEventsQuery) (workbench.SessionEvents, error) {
	view := workbench.SessionEvents{Events: []workbench.Event{}, Coverage: []workbench.SessionCoverage{}, Page: q.Page, PageSize: q.PageSize}
	if q.Page < 1 || q.Page > 1000000 || q.PageSize < 1 || q.PageSize > 100 {
		return view, &apierrors.ServiceError{Code: apierrors.ValidationFailed}
	}
	bindings, err := c.owned.ListOwnedRelaySessions(ctx) // every GET rechecks grant/revision, without archive I/O
	if err != nil {
		return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
	}
	nativeSources, err := c.nativeSources(ctx)
	if err != nil {
		return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	found := false
	scopes := []workbench.SessionAdmissionScope{}
	start, end := time.Time{}, time.Time{}
	if q.Start != "" {
		start, err = time.Parse(time.RFC3339Nano, q.Start)
		if err != nil {
			return view, &apierrors.ServiceError{Code: apierrors.ValidationFailed}
		}
	}
	if q.End != "" {
		end, err = time.Parse(time.RFC3339Nano, q.End)
		if err != nil {
			return view, &apierrors.ServiceError{Code: apierrors.ValidationFailed}
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return view, &apierrors.ServiceError{Code: apierrors.ValidationFailed}
	}
	for _, v := range bindings {
		if q.TaskID != "" && q.TaskID != v.TaskID {
			continue
		}
		if v.Gap == "permission_denied" || v.Gap == "binding_not_recorded" {
			continue
		}
		if current := c.entries[v.TaskID]; current.Native != nil {
			if native, exists := nativeSources[v.TaskID]; exists {
				if native.OwnedSession.Gap == "permission_denied" || native.OwnedSession.Gap == "binding_not_recorded" {
					continue
				}
				v = native.OwnedSession
			}
		}
		found = true
		coverage := coverageFor(v)
		entry, ok := c.entries[v.TaskID]
		if ok && v.SourceStatus == "available" {
			coverage = entry.Coverage
			coverage.TaskID, coverage.TerminalID, coverage.WorktreeID, coverage.ProjectID = v.TaskID, v.TerminalID, v.WorktreeID, v.ProjectID
			coverage.Gap = sessionPublicGap(coverage.Gap)
			if entry.Cursor.BindingID != relayCaptureBinding(v).BindingID() {
				coverage = coverageFor(v)
				coverage.Gap = "session_identity_changed"
			}
		}
		nativeSelected := entry.Native != nil
		if nativeSelected {
			coverage = entry.Coverage
			coverage.Source = "native_notifications"
			currentBinding, bindingErr := relayNativeConnectionBinding(v.BindingID, v.ControllerID, v.StartupOperationID)
			if bindingErr != nil || entry.Native.BindingID != currentBinding {
				coverage = coverageFor(v)
				coverage.Gap = "session_identity_changed"
			}
			if native, exists := nativeSources[v.TaskID]; !exists || relayNativeCaptureBatch(native).BindingID != entry.Native.BindingID || native.OwnedSession.Gap == "permission_denied" {
				coverage = coverageFor(v)
				coverage.Gap = "native_source_unverifiable"
			} else {
				coverage.Unknown, coverage.Errors = max(coverage.Unknown, native.Unknown), max(coverage.Errors, native.Errors)
				if native.LastSequence != entry.Native.Sequence {
					coverage.SnapshotComplete, coverage.SessionComplete, coverage.Gap = false, false, "capture_pending"
				}
				if entry.Native.FirstError != "" {
					coverage.SessionComplete, coverage.Gap = false, sessionPublicGap(entry.Native.FirstError)
				} else if native.FirstGap != "" {
					coverage.SessionComplete, coverage.Gap = false, sessionPublicGap(native.FirstGap)
				}
				if native.Finalized != "confirmed" || native.PrefixComplete != "confirmed" {
					coverage.SessionComplete = false
				}
				if !coverage.SessionComplete {
					coverage.Status = "partial"
				}
			}
		}
		if coverage.Source == "" {
			coverage.Source = coverageFor(v).Source
		}
		if c.blocked != "" {
			coverage.Status, coverage.Gap, coverage.SessionComplete = "unknown", c.blocked, false
		}
		if coverage.Gap == "" && !ok {
			coverage.Gap = "capture_pending"
		}
		view.Coverage = append(view.Coverage, coverage)
		if v.SourceStatus != "gap" || (v.DriveMode == "terminal" && v.Gap == "session_ownership_unverified") {
			nativeID := ""
			nativeBinding := ""
			if v.Ownership == "owned" {
				nativeID = v.SessionID
				if v.SourceStatus == "available" {
					nativeBinding = relayCaptureBinding(v).BindingID()
				}
				if nativeSelected {
					nativeBinding = entry.Native.BindingID
				}
			}
			taskReader, ok := c.owned.(interface {
				ListEvents(context.Context, domain.Caller, string) ([]domain.RelayEvent, error)
			})
			if !ok {
				return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
			}
			committed, e := taskReader.ListEvents(ctx, domain.Caller{Kind: "human", ID: "local-human"}, v.TaskID)
			if e != nil {
				return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
			}
			ids := make([]string, 0, len(committed))
			for _, event := range committed {
				if event.TaskID == v.TaskID && event.Scope == v.ProjectID && event.TerminalID == v.TerminalID {
					if nativeSelected && event.Source == "native" {
						continue
					}
					ids = append(ids, event.ID)
				}
			}
			scopes = append(scopes, workbench.SessionAdmissionScope{TaskID: v.TaskID, ProjectID: v.ProjectID, TerminalID: v.TerminalID, NativeSessionID: nativeID, NativeBindingID: nativeBinding, RelayEventIDs: ids, NativeNotifications: nativeSelected})
			if nativeSelected {
				history, err := c.nativeHistoricalBindings(ctx, v, entry)
				if err != nil {
					return view, &apierrors.ServiceError{Code: apierrors.EvidenceMissing}
				}
				for _, binding := range history {
					scopes = append(scopes, workbench.SessionAdmissionScope{TaskID: v.TaskID, ProjectID: v.ProjectID, TerminalID: v.TerminalID, NativeSessionID: nativeID, NativeBindingID: binding, NativeNotifications: true})
				}
			}
		}
	}
	if q.TaskID != "" && !found {
		return view, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	sort.Slice(view.Coverage, func(i, j int) bool { return view.Coverage[i].TaskID < view.Coverage[j].TaskID })
	if c.admissions == nil {
		return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
	}
	view.Events, view.Total, err = c.admissions.ReadCommittedSessionEvents(ctx, scopes, q)
	if err != nil {
		return view, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable}
	}
	return view, nil
}

type relaySessionRuntime struct {
	workspace.RelayRuntime
	workspace.RelayLifecycle
	collector *relaySessionCollector
}

func (s *relaySessionRuntime) ListRuntimeFacts(ctx context.Context, caller domain.Caller, id string) ([]domain.RelayRuntimeFact, []domain.RelaySessionRelation, error) {
	reader, ok := s.RelayRuntime.(workspace.RelayRuntimeFactsReader)
	if !ok {
		return nil, nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	return reader.ListRuntimeFacts(ctx, caller, id)
}

func (s *relaySessionRuntime) GetCurrentOwnership(ctx context.Context, caller domain.Caller, id string) (domain.RelayCurrentOwnership, error) {
	reader, ok := s.RelayRuntime.(workspace.RelayCurrentOwnershipReader)
	if !ok {
		return domain.RelayCurrentOwnership{}, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	return reader.GetCurrentOwnership(ctx, caller, id)
}

func (s *relaySessionRuntime) ListOwnedTurnBindings(ctx context.Context, id string) ([]domain.RelayTurnBinding, error) {
	reader, ok := s.RelayRuntime.(workspace.RelayOwnedTurnBindingReader)
	if !ok {
		return nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	return reader.ListOwnedTurnBindings(ctx, id)
}

func (s *relaySessionRuntime) ObserveOwned(ctx context.Context) error {
	baseErr := s.RelayLifecycle.ObserveOwned(ctx)
	if baseErr != nil {
		return baseErr
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return s.collector.ObserveOwned(ctx)
}

func (s *relaySessionRuntime) Shutdown(ctx context.Context) error {
	err := s.RelayLifecycle.Shutdown(ctx)
	if err != nil {
		return err
	}
	if ctx.Err() != nil {
		return errors.Join(err, ctx.Err())
	}
	// The periodic observer has been canceled and joined by the entrypoint.
	// Drain the currently available owned archive extent after containment.
	return s.collector.ObserveOwned(ctx)
}
