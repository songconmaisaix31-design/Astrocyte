package main

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relaySessionEntry struct {
	Cursor   events.RelaySessionCursor
	Coverage workbench.SessionCoverage
}
type relaySessionCheckpoints map[string]relaySessionEntry
type relaySessionState interface {
	Load() (relaySessionCheckpoints, error)
	Save(relaySessionCheckpoints) error
}

// Entry-point bridge only: no cross-context app import or second runtime writer.
type relaySessionCollector struct {
	mu      sync.Mutex
	owned   workspace.RelayOwnedSessionReader
	capture *events.RelaySessionCapture
	state   relaySessionState
	entries relaySessionCheckpoints
	blocked string
}

func newRelaySessionCollector(owned workspace.RelayOwnedSessionReader, capture *events.RelaySessionCapture, state relaySessionState) *relaySessionCollector {
	c := &relaySessionCollector{owned: owned, capture: capture, state: state, entries: relaySessionCheckpoints{}}
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
	case "", "archive_not_selected", "binding_not_recorded", "permission_denied", "workspace_identity_changed", "workspace_unavailable", "session_ownership_unverified", "session_identity_changed", "archive_identity_changed", "prefix_unverified", "append_only_unverified", "source_not_finalized", "exit_unverified", "capture_pending", "unknown_records", "checkpoint_unavailable", "checkpoint_save_unknown", "admission_unknown", "malformed_json", "closed_incomplete_tail", "session_header_unverifiable", "session_or_cwd_changed", "source_replaced", "source_truncated", "checkpoint_binding_mismatch", "source_changed_during_read", "selection_denied", "capture_cancelled", "header_read_failed", "source_read_failed", "source_identity_unverifiable_or_aliased", "source_open_failed", "source_not_regular", "checkpoint_invalid", "checkpoint_not_line_boundary":
		return code
	default:
		return "source_unavailable"
	}
}

func coverageFor(v domain.RelayOwnedSession) workbench.SessionCoverage {
	return workbench.SessionCoverage{TaskID: v.TaskID, TerminalID: v.TerminalID, WorktreeID: v.WorktreeID, ProjectID: v.ProjectID, Status: "unknown", Gap: sessionPublicGap(v.Gap)}
}

func (c *relaySessionCollector) ObserveOwned(ctx context.Context) error {
	bindings, err := c.owned.ListOwnedRelaySessions(ctx)
	if err != nil {
		return errors.New("session ownership read unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.blocked != "" {
		return errors.New("session checkpoints blocked")
	}
	failed := false
	for _, v := range bindings {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if v.SourceStatus != "available" {
			continue
		}
		priorEntry, existed := c.entries[v.TaskID]
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
			failed = true
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
		c.entries[v.TaskID] = entry
		// EventService admission precedes this metadata commit. A failed write
		// blocks this process; restart replays old origins through EventRepository.
		if c.state.Save(c.entries) != nil {
			if existed {
				c.entries[v.TaskID] = priorEntry
			} else {
				delete(c.entries, v.TaskID)
			}
			c.blocked = "checkpoint_save_unknown"
			return errors.New("session checkpoint save unknown")
		}
	}
	if failed {
		return errors.New("session coverage incomplete")
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
	c.mu.Lock()
	defer c.mu.Unlock()
	found := false
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
	for _, v := range bindings {
		if q.TaskID != "" && q.TaskID != v.TaskID {
			continue
		}
		if v.Gap == "permission_denied" {
			continue
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
		if c.blocked != "" {
			coverage.Status, coverage.Gap, coverage.SessionComplete = "unknown", c.blocked, false
		}
		if coverage.Gap == "" && !ok {
			coverage.Gap = "capture_pending"
		}
		view.Coverage = append(view.Coverage, coverage)
		if !ok || v.SourceStatus != "available" || v.Ownership != "owned" || entry.Cursor.BindingID != relayCaptureBinding(v).BindingID() {
			continue
		}
		for _, p := range entry.Cursor.Seen {
			ts, e := time.Parse(time.RFC3339Nano, p.TS)
			id, idErr := uuid.Parse(p.ID)
			if e != nil || idErr != nil || id == uuid.Nil || p.Scope != v.ProjectID || p.Topic != "relay:"+v.TaskID || (!start.IsZero() && ts.Before(start)) || (!end.IsZero() && ts.After(end)) {
				continue
			}
			kind := p.Kind
			switch kind {
			case "native_session_header", "native_activity", "edit_file", "search", "read_paper", "commit":
			default:
				kind = "native_activity"
			}
			var tokens *int64
			if p.Tokens != nil && *p.Tokens >= 0 {
				n := int64(*p.Tokens)
				tokens = &n
			}
			view.Events = append(view.Events, workbench.Event{ID: id.String(), Kind: kind, Topic: "relay:" + v.TaskID, Tokens: tokens, Scope: v.ProjectID, TS: ts.UTC().Format(time.RFC3339Nano)})
		}
	}
	if q.TaskID != "" && !found {
		return view, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	sort.Slice(view.Coverage, func(i, j int) bool { return view.Coverage[i].TaskID < view.Coverage[j].TaskID })
	sort.Slice(view.Events, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, view.Events[i].TS)
		b, _ := time.Parse(time.RFC3339Nano, view.Events[j].TS)
		if a.Equal(b) {
			return view.Events[i].ID > view.Events[j].ID
		}
		return a.After(b)
	})
	view.Total = len(view.Events)
	offset := (q.Page - 1) * q.PageSize
	if offset >= view.Total {
		view.Events = []workbench.Event{}
	} else {
		last := offset + q.PageSize
		if last > view.Total {
			last = view.Total
		}
		view.Events = view.Events[offset:last]
	}
	return view, nil
}

type relaySessionRuntime struct {
	workspace.RelayRuntime
	workspace.RelayLifecycle
	collector *relaySessionCollector
}

func (s *relaySessionRuntime) ObserveOwned(ctx context.Context) error {
	baseErr := s.RelayLifecycle.ObserveOwned(ctx)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return errors.Join(baseErr, s.collector.ObserveOwned(ctx))
}

func (s *relaySessionRuntime) Shutdown(ctx context.Context) error {
	err := s.RelayLifecycle.Shutdown(ctx)
	if ctx.Err() != nil {
		return errors.Join(err, ctx.Err())
	}
	// The periodic observer has been canceled and joined by the entrypoint.
	// Drain the currently available owned archive extent after containment.
	return errors.Join(err, s.collector.ObserveOwned(ctx))
}
