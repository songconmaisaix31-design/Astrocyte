package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// RelaySessionBinding is a trusted, private selection, never an HTTP DTO. The
// caller must obtain ownership and archive permission separately; no discovery
// or ReadContext expansion is performed by this adapter.
type RelaySessionBinding struct {
	TaskID, TerminalID, SessionID, WorktreeID, ProjectID string
	NativeID                                             string `json:"-"`
	Format                                               string
	Path, WorkspaceRoot                                  string   `json:"-"`
	ApprovedRoots                                        []string `json:"-"`
	Owned, Authorized                                    bool
	ExitConfirmed, ArchiveFinalized, PrefixComplete      bool
}

// RelaySessionProjection is the closed, derived serialization boundary. An
// unknown source timestamp has no projection rather than invented activity.
type RelaySessionProjection struct {
	ID     string `json:"event_id"`
	Kind   string `json:"kind"`
	Topic  string `json:"topic"`
	Tokens *int   `json:"tokens"`
	Scope  string `json:"scope"`
	TS     string `json:"ts"`
}

type RelaySessionCounts struct {
	Lines, Projected, Duplicates, Unknown, Errors, UnknownTokens int64
}

// The host may serialize this cursor in its authorized local state. It contains
// no original line, prompt, native event ID, file path or terminal output. This
// adapter does not write RelayStore, a cursor database, or a raw spool.
type RelaySessionCursor struct {
	Version      int
	BindingID    string
	FileIdentity string
	Offset       int64
	Counts       RelaySessionCounts
	Seen         map[string]RelaySessionProjection
	FirstError   string
	Blocked      string
}

type RelaySessionItem struct {
	Origin     string
	Projection RelaySessionProjection
}

type RelaySessionBatch struct {
	// Next is a proposal until all Items have been admitted. A failed admission
	// leaves the original checkpoint intact and requires explicit reconciliation.
	Next                                RelaySessionCursor
	Items                               []RelaySessionItem
	SnapshotBytes, Remaining, TailBytes int64
	SnapshotComplete, SessionComplete   bool
	HeaderVerified                      bool
}

func sessionFailure(code string) error { return errors.New("relay session capture: " + code) }

func sessionBindingID(b RelaySessionBinding) string {
	// Namespace UUIDs identify events, not a hash-based completion/proof system.
	data, _ := json.Marshal([]string{b.TaskID, b.TerminalID, b.SessionID, b.WorktreeID, b.ProjectID, b.NativeID, b.Format})
	return uuid.NewSHA1(uuid.NameSpaceURL, data).String()
}

// BindingID lets the entrypoint check a cached admission against the current
// private selection without exporting source identifiers or paths.
func (b RelaySessionBinding) BindingID() string { return sessionBindingID(b) }

func sessionOrigin(bindingID, fileIdentity string, offset int64) (string, string) {
	// Correlation IDs (call_id/message.id) can span multiple native records.
	// Only a selected append-only file position identifies this record.
	id := uuid.NewSHA1(uuid.MustParse(bindingID), []byte(fmt.Sprintf("%s:byte:%d", fileIdentity, offset))).String()
	return "relay-session:" + bindingID + ":" + id, id
}

func copySessionCursor(v RelaySessionCursor) RelaySessionCursor {
	out := v
	out.Seen = make(map[string]RelaySessionProjection, len(v.Seen))
	for k, p := range v.Seen {
		if p.Tokens != nil {
			n := *p.Tokens
			p.Tokens = &n
		}
		out.Seen[k] = p
	}
	return out
}

func markSessionError(c *RelaySessionCursor, code string, blocked bool) {
	c.Counts.Errors++
	if c.FirstError == "" {
		c.FirstError = code
	}
	if blocked {
		c.Blocked = code
	}
}

type RelaySessionRecorder interface {
	Record(context.Context, []attention.StoredEvent) ([]attention.StoredEvent, error)
}

// RelaySessionCapture combines the parser with the existing EventService. It
// records pending source observations only; no Event or Signal is synthesized.
type RelaySessionCapture struct {
	recorder RelaySessionRecorder
	clock    func() time.Time
}

func NewRelaySessionCapture(recorder RelaySessionRecorder, clock func() time.Time) *RelaySessionCapture {
	if clock == nil {
		clock = time.Now
	}
	return &RelaySessionCapture{recorder: recorder, clock: clock}
}

type RelaySessionAdmission struct {
	Batch           RelaySessionBatch
	Checkpoint      RelaySessionCursor
	AdmissionErrors int64
	// Pending is derived only and is retained for an explicit retry of event
	// admission, never of a model action. No automatic retry is made here.
	Pending bool
}

func (s *RelaySessionCapture) Capture(ctx context.Context, b RelaySessionBinding, prior RelaySessionCursor) (RelaySessionAdmission, error) {
	batch, err := ReadRelaySession(ctx, b, prior)
	result := RelaySessionAdmission{Batch: batch, Checkpoint: copySessionCursor(prior)}
	if err != nil {
		// No candidate from a failed read is committed. Keep the offset that
		// precedes these not-yet-admitted records, plus the first static error.
		if result.Checkpoint.Blocked == "" {
			code := batch.Next.FirstError
			if code == "" {
				code = "source_read_failed"
			}
			markSessionError(&result.Checkpoint, code, true)
		}
		result.Batch.SessionComplete = false
		return result, err
	}
	if len(batch.Items) == 0 {
		result.Checkpoint = batch.Next
		return result, nil
	}
	if err := s.admit(ctx, b, batch.Items); err != nil {
		markSessionError(&result.Checkpoint, "admission_unknown", true)
		result.Pending = true
		result.AdmissionErrors++
		result.Batch.SessionComplete = false
		return result, sessionFailure("admission_unknown")
	}
	result.Checkpoint = batch.Next
	return result, nil
}

// ReconcileAdmission must be invoked explicitly by the authorized host. It
// uses the original derived origins after a sink error, without rereading raw
// data, changing identity or driving the Agent. Source permission is rechecked.
func (s *RelaySessionCapture) ReconcileAdmission(ctx context.Context, b RelaySessionBinding, pending RelaySessionAdmission) (RelaySessionAdmission, error) {
	if !pending.Pending || !b.Owned || !b.Authorized || pending.Batch.Next.BindingID != sessionBindingID(b) {
		return pending, sessionFailure("reconciliation_denied")
	}
	if err := validateSessionSelection(b); err != nil {
		return pending, err
	}
	f, err := os.Open(b.Path)
	if err != nil {
		return pending, sessionFailure("reconciliation_source_unavailable")
	}
	info, statErr := f.Stat()
	identity, identityErr := "", statErr
	if statErr == nil {
		identity, identityErr = relaySessionFileIdentity(f, info)
	}
	_ = f.Close()
	if identityErr != nil || identity != pending.Batch.Next.FileIdentity {
		return pending, sessionFailure("reconciliation_source_changed")
	}
	if err := s.admit(ctx, b, pending.Batch.Items); err != nil {
		markSessionError(&pending.Checkpoint, "admission_unknown", true)
		pending.AdmissionErrors++
		return pending, sessionFailure("admission_unknown")
	}
	next := copySessionCursor(pending.Batch.Next)
	if next.FirstError == "" {
		next.FirstError = pending.Checkpoint.FirstError
	}
	next.Counts.Errors += pending.AdmissionErrors
	next.Blocked = ""
	pending.Checkpoint, pending.Pending = next, false
	// Historical first RED is retained. A reconciled receipt is not full-session
	// acceptance and this method cannot clear source completeness gaps.
	pending.Batch.SessionComplete = false
	return pending, nil
}

func (s *RelaySessionCapture) admit(ctx context.Context, b RelaySessionBinding, items []RelaySessionItem) error {
	if s.recorder == nil {
		return sessionFailure("admission_unavailable")
	}
	rows := make([]attention.StoredEvent, 0, len(items))
	for _, item := range items {
		p := item.Projection
		payload, err := json.Marshal(p)
		if err != nil {
			return sessionFailure("projection_invalid")
		}
		ts, err := time.Parse(time.RFC3339Nano, p.TS)
		if err != nil {
			return sessionFailure("projection_invalid")
		}
		rows = append(rows, attention.StoredEvent{OriginKey: item.Origin, ReceivedAt: s.clock().UnixMilli(), RawBytes: payload,
			Observation: attention.Observation{Source: b.Format, SourceKind: p.Kind, Session: b.SessionID, Scope: b.ProjectID, TS: ts.UnixMilli(), Tokens: p.Tokens, Outcome: "unknown", Depth: "unknown", Provenance: "workspace:relay_session"}})
	}
	admitted, err := s.recorder.Record(ctx, rows)
	if err != nil {
		return err
	}
	if len(admitted) != len(rows) {
		return sessionFailure("admission_acknowledgement_unknown")
	}
	for i := range rows {
		if admitted[i].ID == "" || admitted[i].OriginKey != rows[i].OriginKey {
			return sessionFailure("admission_acknowledgement_unknown")
		}
	}
	return nil
}
