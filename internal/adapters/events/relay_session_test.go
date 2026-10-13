package events

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func sessionFixture(t *testing.T) RelaySessionBinding {
	t.Helper()
	base := os.Getenv("LOCALAPPDATA")
	if !filepath.IsAbs(base) {
		base = t.TempDir()
	}
	root := filepath.Join(base, "astrocyte-workbench", "v0.2-access", "verification", "relay-session", uuid.NewString())
	archive, workspace := filepath.Join(root, "archive"), filepath.Join(root, "workspace-PRIVATE_CANARY")
	if os.MkdirAll(archive, 0700) != nil || os.MkdirAll(workspace, 0700) != nil {
		t.Fatal("public session fixture unavailable")
	}
	b := RelaySessionBinding{TaskID: uuid.NewString(), TerminalID: uuid.NewString(), SessionID: uuid.NewString(), WorktreeID: uuid.NewString(), ProjectID: "public-project", NativeID: "private-native-PRIVATE_CANARY", Format: "codex", Path: filepath.Join(archive, "public.jsonl"), WorkspaceRoot: workspace, ApprovedRoots: []string{archive}, Owned: true, Authorized: true}
	writeSession(t, b, sessionHeader(b))
	return b
}

func sessionJSON(v any) string { b, _ := json.Marshal(v); return string(b) + "\n" }
func sessionHeader(b RelaySessionBinding) string {
	return sessionJSON(map[string]any{"type": "session_meta", "timestamp": "2026-10-14T01:00:00Z", "payload": map[string]any{"id": b.NativeID, "cwd": b.WorkspaceRoot}})
}
func sessionCall(id string) string {
	return sessionJSON(map[string]any{"type": "response_item", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"type": "function_call", "call_id": id, "name": "Edit", "arguments": "PRIVATE_CANARY prompt/tool arguments"}})
}
func writeSession(t *testing.T, b RelaySessionBinding, s string) {
	t.Helper()
	if os.WriteFile(b.Path, []byte(s), 0600) != nil {
		t.Fatal("public session write unavailable")
	}
}
func appendSession(t *testing.T, b RelaySessionBinding, s string) {
	t.Helper()
	f, e := os.OpenFile(b.Path, os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal("public append unavailable")
	}
	_, e = f.WriteString(s)
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		t.Fatal("public append failed")
	}
}

func TestRelaySessionAppendTailRestartDuplicateAndClosure(t *testing.T) {
	b := sessionFixture(t)
	call := sessionCall("call-PRIVATE_CANARY")
	appendSession(t, b, call[:len(call)/2])
	first, err := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if err != nil || first.Next.Counts.Lines != 1 || first.Next.Counts.Projected != 1 || first.TailBytes == 0 || first.SnapshotComplete || first.SessionComplete {
		t.Fatal("incomplete tail advanced or became complete")
	}
	checkpointJSON, _ := json.Marshal(first.Next)
	if strings.Contains(string(checkpointJSON), "PRIVATE_CANARY") {
		t.Fatal("checkpoint copied private source data")
	}
	var restored RelaySessionCursor
	if json.Unmarshal(checkpointJSON, &restored) != nil {
		t.Fatal("checkpoint restore failed")
	}
	appendSession(t, b, call[len(call)/2:]+call)
	second, err := ReadRelaySession(context.Background(), b, restored)
	if err != nil || second.Next.Counts.Lines != 3 || second.Next.Counts.Projected != 3 || second.Next.Counts.Duplicates != 0 || len(second.Items) != 2 || !second.SnapshotComplete || second.SessionComplete {
		t.Fatal("append/restart/duplicate/exit distinction failed")
	}
	third, err := ReadRelaySession(context.Background(), b, second.Next)
	if err != nil || len(third.Items) != 0 || third.Next.Offset != second.Next.Offset || third.Next.Counts != second.Next.Counts {
		t.Fatal("unchanged read renewed or duplicated evidence")
	}
	// A rewound checkpoint retains previously admitted positions. Identical
	// records at different positions remain separate; rereading positions dedupes.
	rewound := copySessionCursor(third.Next)
	rewound.Offset = 0
	duplicate, err := ReadRelaySession(context.Background(), b, rewound)
	if err != nil || len(duplicate.Items) != 0 || duplicate.Next.Counts.Duplicates != 3 {
		t.Fatal("same-position replay changed stable identities")
	}
	b.ExitConfirmed = true
	b.ArchiveFinalized = true
	b.PrefixComplete = true
	closed, err := ReadRelaySession(context.Background(), b, third.Next)
	if err != nil || !closed.SessionComplete {
		t.Fatal("fixture with explicit closure evidence not reconciled")
	}
}

func TestRelaySessionCompleteLineBeyondOldReaderLimit(t *testing.T) {
	b := sessionFixture(t)
	long := sessionJSON(map[string]any{"type": "response_item", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"type": "function_call", "call_id": "long-call", "name": "Edit", "arguments": strings.Repeat("PRIVATE_CANARY", 200000)}})
	if len(long) <= maxReplayBytes {
		t.Fatal("fixture did not exceed old bounded reader")
	}
	appendSession(t, b, long)
	r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if e != nil || r.Next.Counts.Lines != 2 || r.Next.Counts.Projected != 2 || !r.SnapshotComplete || r.Next.Offset != r.SnapshotBytes {
		t.Fatal("complete long line was truncated")
	}
	for _, item := range r.Items {
		p, _ := json.Marshal(item.Projection)
		var fields map[string]any
		_ = json.Unmarshal(p, &fields)
		if len(fields) != 6 || strings.Contains(string(p), "PRIVATE_CANARY") {
			t.Fatal("long source escaped closed projection")
		}
	}
}

func TestRelaySessionUnknownAndErrorsRemainInDenominator(t *testing.T) {
	b := sessionFixture(t)
	appendSession(t, b, "not JSON PRIVATE_CANARY\n"+sessionJSON(map[string]any{"type": "new_native_kind_PRIVATE_CANARY", "timestamp": "2026-10-14T01:00:02Z", "raw": "PRIVATE_CANARY"})+sessionJSON(map[string]any{"type": "response_item", "payload": map[string]any{"type": "function_call", "name": "Edit"}}))
	b.ExitConfirmed = true
	b.ArchiveFinalized = true
	b.PrefixComplete = true
	r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if e != nil || r.Next.Counts.Lines != 4 || r.Next.Counts.Errors != 1 || r.Next.Counts.Unknown != 2 || r.Next.Counts.Projected != 2 || r.SessionComplete || r.Next.FirstError != "malformed_json" {
		t.Fatal("unknown/error records excluded from coverage")
	}
	b2 := sessionFixture(t)
	b2.Format = "never-seen-CLI"
	r, e = ReadRelaySession(context.Background(), b2, RelaySessionCursor{})
	if e != nil || r.Next.Counts.Lines != 1 || r.Next.Counts.Unknown != 1 || r.HeaderVerified || r.SessionComplete || len(r.Items) != 0 {
		t.Fatal("unknown format inferred support")
	}
}

func TestRelaySessionBoundaryAndMutationFailures(t *testing.T) {
	for _, name := range []string{"unowned", "unauthorized", "outside-root", "wrong-native", "wrong-cwd", "mid-session", "mid-cwd", "typed-cwd", "truncate", "replace", "cursor-task", "cursor-offset", "closed-tail", "hardlink"} {
		t.Run(name, func(t *testing.T) {
			b := sessionFixture(t)
			prior := RelaySessionCursor{}
			var e error
			if name == "truncate" || name == "replace" || strings.HasPrefix(name, "cursor-") {
				r, err := ReadRelaySession(context.Background(), b, prior)
				if err != nil {
					t.Fatal("initial read unavailable")
				}
				prior = r.Next
			}
			switch name {
			case "unowned":
				b.Owned = false
			case "unauthorized":
				b.Authorized = false
			case "outside-root":
				b.ApprovedRoots = []string{b.WorkspaceRoot}
			case "wrong-native":
				b.NativeID = "different-session"
			case "wrong-cwd":
				b.WorkspaceRoot = filepath.Dir(b.WorkspaceRoot)
			case "mid-session":
				appendSession(t, b, sessionJSON(map[string]any{"type": "event_msg", "sessionId": "different", "timestamp": "2026-10-14T01:00:01Z"}))
			case "mid-cwd":
				appendSession(t, b, sessionJSON(map[string]any{"type": "turn_context", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"cwd": filepath.Dir(b.WorkspaceRoot)}}))
			case "typed-cwd":
				appendSession(t, b, sessionJSON(map[string]any{"type": "turn_context", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"cwd": 123}}))
			case "truncate":
				writeSession(t, b, "")
			case "replace":
				replacement := b.Path + ".replacement"
				if os.WriteFile(replacement, []byte(sessionHeader(b)), 0600) != nil || os.Rename(b.Path, b.Path+".retained") != nil || os.Rename(replacement, b.Path) != nil {
					t.Fatal("public replacement fixture failed")
				}
			case "cursor-task":
				b.TaskID = uuid.NewString()
			case "cursor-offset":
				prior.Offset--
			case "closed-tail":
				appendSession(t, b, strings.TrimSuffix(sessionCall("tail"), "\n"))
				b.ArchiveFinalized = true
			case "hardlink":
				if err := os.Link(b.Path, b.Path+".alias"); err != nil {
					t.Skip("NOT_RUN: hardlink unavailable")
				}
			}
			r, err := ReadRelaySession(context.Background(), b, prior)
			e = err
			if name == "closed-tail" {
				if r.Next.Blocked != "closed_incomplete_tail" || r.SessionComplete {
					t.Fatal("closed tail disappeared")
				}
				return
			}
			if e == nil || r.SessionComplete || r.Next.FirstError == "" || strings.Contains(e.Error(), "PRIVATE_CANARY") || strings.Contains(e.Error(), b.Path) {
				t.Fatal("boundary/mutation did not fail closed and safely")
			}
		})
	}
}

type sessionFailRecorder struct {
	delegate       RelaySessionRecorder
	calls          int
	fail           bool
	commitThenFail bool
}

func (s *sessionFailRecorder) Record(ctx context.Context, rows []attention.StoredEvent) ([]attention.StoredEvent, error) {
	s.calls++
	if s.commitThenFail {
		_, err := s.delegate.Record(ctx, rows)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("PRIVATE_CANARY ambiguous committed admission")
	}
	if s.fail {
		return nil, errors.New("PRIVATE_CANARY sink SQL/provider")
	}
	return s.delegate.Record(ctx, rows)
}

func TestRelaySessionEventServiceStableAdmissionAndUnknownFailure(t *testing.T) {
	b := sessionFixture(t)
	appendSession(t, b, sessionCall("record"))
	db, e := sqlite.Open(filepath.Join(filepath.Dir(b.ApprovedRoots[0]), "app.sqlite"))
	if e != nil {
		t.Fatal("isolated app DB unavailable")
	}
	defer db.Close()
	if db.RunMigrations(context.Background(), migrations.FS) != nil {
		t.Fatal("isolated schema unavailable")
	}
	repo := sqlite.NewEventRepository(db)
	port := &sessionFailRecorder{delegate: attention.NewPendingEventService(repo), fail: true}
	capture := NewRelaySessionCapture(port, func() time.Time { return time.Date(2026, 10, 14, 1, 0, 4, 0, time.UTC) })
	failed, err := capture.Capture(context.Background(), b, RelaySessionCursor{})
	if err == nil || !failed.Pending || failed.Checkpoint.Offset != 0 || failed.Checkpoint.FirstError != "admission_unknown" || port.calls != 1 || strings.Contains(err.Error(), "PRIVATE_CANARY") {
		t.Fatal("sink failure lost checkpoint or leaked raw")
	}
	_, err = capture.Capture(context.Background(), b, failed.Checkpoint)
	if err == nil || port.calls != 1 {
		t.Fatal("blocked admission automatically retried")
	}
	port.fail = false
	ok, err := capture.ReconcileAdmission(context.Background(), b, failed)
	if err != nil || ok.Pending || ok.Checkpoint.FirstError != "admission_unknown" || ok.Checkpoint.Offset != ok.Batch.SnapshotBytes {
		t.Fatal("explicit stable-origin reconciliation failed")
	}
	fixtureEnd := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).UnixMilli()
	rows, err := repo.ListEvents(context.Background(), 0, fixtureEnd)
	if err != nil || len(rows) != 2 {
		t.Fatal("pending observations missing")
	}
	for _, row := range rows {
		var projected map[string]any
		_ = json.Unmarshal(row.RawBytes, &projected)
		if len(projected) != 6 || row.Event != nil || row.Observation.Outcome != "unknown" || row.Observation.Depth != "unknown" || strings.Contains(string(row.RawBytes), "PRIVATE_CANARY") {
			t.Fatal("source body/signals crossed admission boundary")
		}
	}
	if _, e = capture.Capture(context.Background(), b, RelaySessionCursor{}); e != nil {
		t.Fatal("stable event replay refused")
	}
	replayed, e := repo.ListEvents(context.Background(), 0, fixtureEnd)
	if e != nil || len(replayed) != 2 || replayed[0].ID != rows[0].ID {
		t.Fatal("replay changed original event identity")
	}
	signals, e := repo.ListSignals(context.Background(), 0, fixtureEnd)
	if e != nil || len(signals) != 0 {
		t.Fatal("capture produced cognitive signals")
	}
	// Unknown acknowledgement can occur after the transaction committed. Keep
	// the old cursor and original origins, with no automatic second Record call.
	port.commitThenFail = true
	unknown, e := capture.Capture(context.Background(), b, RelaySessionCursor{})
	if e == nil || !unknown.Pending || unknown.Checkpoint.Offset != 0 {
		t.Fatal("committed unknown acknowledgement advanced cursor")
	}
	firstCalls := port.calls
	_, e = capture.Capture(context.Background(), b, unknown.Checkpoint)
	if e == nil || port.calls != firstCalls {
		t.Fatal("committed unknown retried automatically")
	}
	port.commitThenFail = false
	reconciled, e := capture.ReconcileAdmission(context.Background(), b, unknown)
	if e != nil || reconciled.Pending || reconciled.Checkpoint.FirstError != "admission_unknown" {
		t.Fatal("committed unknown explicit reconciliation failed")
	}
	confirmed, e := repo.ListEvents(context.Background(), 0, fixtureEnd)
	if e != nil || len(confirmed) != 2 || confirmed[0].ID != rows[0].ID {
		t.Fatal("unknown commit reconciliation replaced original identity")
	}
}
