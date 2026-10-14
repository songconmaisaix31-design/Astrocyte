package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayOwnedNativeFixture struct {
	*relayNativeRecording
	root                string
	workspace           string
	selected            bool
	wrongIdentity       bool
	badSource           string
	appendOnly          bool
	wrongSessionProject bool
	wrongSessionCLI     bool
	finalized           bool
	sourceCalls         atomic.Int32
}

func (a *relayOwnedNativeFixture) Start(ctx context.Context, r domain.NativeRequest) (domain.NativeSession, error) {
	s, err := a.relayNativeRecording.Start(ctx, r)
	a.workspace = r.Project.Root
	s.HistoryPath = filepath.Join(a.root, s.ID+".jsonl")
	if a.wrongSessionProject {
		s.ProjectID = "foreign-project"
	}
	if a.wrongSessionCLI {
		s.CLI = "foreign-cli"
	}
	if err == nil {
		err = os.WriteFile(s.HistoryPath, []byte("{\"public_fixture\":true}\n"), 0600)
	}
	return s, err
}
func (a *relayOwnedNativeFixture) OwnedRelayArchive(_ context.Context, s domain.NativeSession) (domain.RelaySessionArchive, error) {
	a.sourceCalls.Add(1)
	if !a.selected {
		return domain.RelaySessionArchive{}, os.ErrNotExist
	}
	archive := domain.RelaySessionArchive{SessionID: s.ID, NativeID: s.NativeID, WorkspaceRoot: a.workspace, Path: s.HistoryPath, Root: a.root, Format: "public_jsonl", Finalized: a.finalized, AppendOnly: a.appendOnly}
	if a.wrongIdentity {
		archive.NativeID = "different-native"
	}
	switch a.badSource {
	case "cwd":
		archive.WorkspaceRoot = a.root
	case "session":
		archive.SessionID = "foreign-session"
	case "outside_root":
		archive.Root = filepath.Join(a.root, "other-directory")
	case "relative":
		archive.Path = "relative-session.jsonl"
	case "format":
		archive.Format = ""
	}
	return archive, nil
}

type relayOwnedRegistry struct{ adapter *relayOwnedNativeFixture }

type relayOwnedRootCounter struct {
	app.RelayOwnedRoots
	calls atomic.Int32
}

func (r *relayOwnedRootCounter) CanonicalRoot(root string) (string, error) {
	r.calls.Add(1)
	return r.RelayOwnedRoots.CanonicalRoot(root)
}

func (r relayOwnedRegistry) Adapter(string) (app.NativeAdapter, error) { return r.adapter, nil }
func (r relayOwnedRegistry) List() []string                            { return []string{"alpha"} }

func ownedReader(t *testing.T, service app.RelayRuntime) app.RelayOwnedSessionReader {
	t.Helper()
	r, ok := service.(app.RelayOwnedSessionReader)
	if !ok {
		t.Fatal("owned read port absent")
	}
	return r
}
func findOwned(t *testing.T, r app.RelayOwnedSessionReader, id string) domain.RelayOwnedSession {
	t.Helper()
	all, err := r.ListOwnedRelaySessions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range all {
		if b.TaskID == id {
			return b
		}
	}
	t.Fatal("owned binding missing")
	return domain.RelayOwnedSession{}
}
func newOwnedFixture(t *testing.T, selected bool) (*relayFixture, *relayOwnedNativeFixture, app.RelayDriver, *relayCountingStore) {
	t.Helper()
	f := newRelayFixture(t)
	f.options.Clock = time.Now
	f.options.OwnedRoots = agents.RelayOwnedRoots{}
	a := &relayOwnedNativeFixture{relayNativeRecording: &relayNativeRecording{}, root: t.TempDir(), selected: selected}
	d := agents.NewRelayNativeDriver(relayOwnedRegistry{adapter: a})
	c := &relayCountingStore{RelayStore: f.store}
	f.service = f.assemble(c, d, nil)
	return f, a, d, c
}

func TestRelayOwnedReaderReadonlyPrivateAndColdReload(t *testing.T) {
	f, a, d, c := newOwnedFixture(t, true)
	ctx := context.Background()
	id := f.intent()
	if _, e := f.service.ConfirmIntent(ctx, caller(), id, confirm()); e != nil {
		t.Fatal(e)
	}
	first := findOwned(t, ownedReader(t, f.service), id)
	if first.BindingID == "" || first.SessionID == "" || first.NativeID == "" || first.SourceStatus != "available" || first.ArchivePath == "" {
		t.Fatalf("selected fixture source not bound: %s", first.Gap)
	}
	if first.ArchiveFinalized != "unknown" || first.ArchiveAppendOnly != "unknown" {
		t.Fatal("source selection invented archive closure")
	}
	c.updates.Store(0)
	calls := a.sourceCalls.Load()
	before, _ := f.store.Snapshot(ctx)
	stateJSON, _ := json.Marshal(before)
	for i := 0; i < 8; i++ {
		b := findOwned(t, ownedReader(t, f.service), id)
		if b.BindingID != first.BindingID {
			t.Fatal("read invented binding generation")
		}
	}
	after, _ := f.store.Snapshot(ctx)
	stateAfter, _ := json.Marshal(after)
	if c.updates.Load() != 0 || calls != a.sourceCalls.Load() || string(stateJSON) != string(stateAfter) || len(a.starts) != 1 || len(a.resumes) != 0 {
		t.Fatal("read wrote state/receipt, resolved source, or controlled process")
	}
	public, _ := json.Marshal(first)
	for _, private := range []string{first.WorkspaceRoot, first.ArchivePath, first.ArchiveRoot, first.SessionID, first.NativeID} {
		encoded, _ := json.Marshal(private)
		if strings.Contains(string(public), string(encoded)) {
			t.Fatal("private source escaped owned DTO")
		}
	}
	view, err := f.service.GetTask(ctx, caller(), id)
	if err != nil {
		t.Fatal(err)
	}
	transport, _ := json.Marshal(view)
	for _, private := range []string{first.ArchivePath, first.ArchiveRoot, first.SessionID, first.NativeID} {
		encoded, _ := json.Marshal(private)
		if strings.Contains(string(transport), string(encoded)) {
			t.Fatal("native source reference entered existing task transport")
		}
	}
	if e := f.store.Close(); e != nil {
		t.Fatal(e)
	}
	var e error
	f.store, e = sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	f.service = f.assemble(f.store, d, nil)
	reloaded := findOwned(t, ownedReader(t, f.service), id)
	if reloaded.BindingID != first.BindingID || reloaded.ArchivePath != first.ArchivePath || reloaded.NativeID != first.NativeID {
		t.Fatal("private binding did not reload")
	}
	state, e := f.store.Snapshot(ctx)
	if e != nil || state.Sessions[first.TerminalID].HistoryPath != first.ArchivePath {
		t.Fatal("private HistoryPath did not survive isolated wire")
	}
}

func TestRelayOwnedResumeAndReleasedTailSurviveSameTerminalNewTask(t *testing.T) {
	f, a, d, _ := newOwnedFixture(t, true)
	ctx := context.Background()
	id := f.intent()
	if _, e := f.service.ConfirmIntent(ctx, caller(), id, confirm()); e != nil {
		t.Fatal(e)
	}
	first := findOwned(t, ownedReader(t, f.service), id)
	if _, e := f.service.Stop(ctx, caller(), id); e != nil {
		t.Fatal(e)
	}
	stopped := findOwned(t, ownedReader(t, f.service), id)
	if stopped.ExitedAt == nil || stopped.ArchiveFinalized != "unknown" {
		t.Fatal("exit failed or pretended final archive flush")
	}
	if _, e := f.service.Resume(ctx, caller(), id); e != nil {
		t.Fatal(e)
	}
	resumed := findOwned(t, ownedReader(t, f.service), id)
	if resumed.BindingID != first.BindingID || resumed.SessionID != first.SessionID || resumed.ExitedAt != nil || len(a.starts) != 1 || len(a.resumes) != 1 || a.resumes[0].Command.Message != "" {
		t.Fatal("resume re-created binding or replayed input")
	}
	if _, e := f.service.Stop(ctx, caller(), id); e != nil {
		t.Fatal(e)
	}
	if _, e := f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true}); e != nil {
		t.Fatal(e)
	}
	if _, e := f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "done", To: "released"}); e != nil {
		t.Fatal(e)
	}
	second := f.intent()
	if _, e := f.service.ConfirmIntent(ctx, caller(), second, confirm()); e != nil {
		t.Fatal(e)
	}
	current := findOwned(t, ownedReader(t, f.service), second)
	if current.TerminalID != first.TerminalID || current.WorktreeID != first.WorktreeID || current.SessionID == first.SessionID || current.BindingID == first.BindingID {
		t.Fatal("new task did not retain terminal but change native binding")
	}
	if e := f.store.Close(); e != nil {
		t.Fatal(e)
	}
	var e error
	f.store, e = sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	f.service = f.assemble(f.store, d, nil)
	tail := findOwned(t, ownedReader(t, f.service), id)
	if tail.ExitedAt == nil || tail.BindingID != first.BindingID || tail.SessionID != first.SessionID || tail.ArchivePath != first.ArchivePath || tail.SourceStatus != "available" {
		t.Fatal("released archive tail lost on terminal reuse or restart")
	}
	// The test writes only this public fixture, after the old task is closed.
	file, e := os.OpenFile(tail.ArchivePath, os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	_, e = file.WriteString("{\"public_tail\":true}\n")
	_ = file.Close()
	if e != nil {
		t.Fatal(e)
	}
	if final := findOwned(t, ownedReader(t, f.service), id); final.ArchivePath != tail.ArchivePath || final.ArchiveFinalized != "unknown" {
		t.Fatal("exit tail selected another file or falsely finalized")
	}
}

func TestRelayOwnedPermissionsAndWorkspaceChangesHidePrivateSources(t *testing.T) {
	for _, reason := range []string{"revoked_grant", "expired_grant", "revision", "action", "source", "workspace", "terminal"} {
		t.Run(reason, func(t *testing.T) {
			f, a, d, c := newOwnedFixture(t, true)
			roots := &relayOwnedRootCounter{RelayOwnedRoots: agents.RelayOwnedRoots{}}
			f.options.OwnedRoots = roots
			f.service = f.assemble(c, d, nil)
			ctx := context.Background()
			id := f.intent()
			if _, e := f.service.ConfirmIntent(ctx, caller(), id, confirm()); e != nil {
				t.Fatal(e)
			}
			switch reason {
			case "revoked_grant":
				now := time.Now()
				f.projects.grant = &domain.ProjectGrant{RevokedAt: &now}
			case "expired_grant":
				past := time.Now().Add(-time.Second)
				f.projects.grant = &domain.ProjectGrant{ExpiresAt: &past}
			case "revision":
				f.projects.project.Settings.Revision++
			case "action":
				f.projects.project.Settings.AllowedActions = []string{"start"}
			case "source":
				f.projects.project.Root = t.TempDir()
			case "workspace":
				_ = f.store.Update(ctx, func(st *app.RelayState) error {
					task := st.Tasks[id]
					w := st.Workspaces[task.WorktreeID]
					w.Root = t.TempDir()
					st.Workspaces[w.ID] = w
					return nil
				})
			case "terminal":
				_ = f.store.Update(ctx, func(st *app.RelayState) error {
					task := st.Tasks[id]
					term := st.Terminals[task.TerminalID]
					term.Agent = "other"
					st.Terminals[term.ID] = term
					return nil
				})
			}
			c.updates.Store(0)
			calls := a.sourceCalls.Load()
			b := findOwned(t, ownedReader(t, f.service), id)
			if b.SourceStatus != "gap" || b.Gap == "" || b.WorkspaceRoot != "" || b.ArchivePath != "" || b.ArchiveRoot != "" || b.SessionID != "" || b.NativeID != "" || a.sourceCalls.Load() != calls || c.updates.Load() != 0 || roots.calls.Load() != 0 {
				t.Fatal("denied binding exposed/read source or wrote state")
			}
		})
	}
}

func TestRelayOwnedMissingOrConflictingSourceIsGapNeverGuessedFromHistory(t *testing.T) {
	f, a, _, c := newOwnedFixture(t, false)
	ctx := context.Background()
	id := f.intent()
	if _, e := f.service.ConfirmIntent(ctx, caller(), id, confirm()); e != nil {
		t.Fatal(e)
	}
	b := findOwned(t, ownedReader(t, f.service), id)
	if b.SourceStatus != "unknown" || b.ArchivePath != "" || b.Gap != "archive_not_selected" {
		t.Fatal("HistoryPath alone inferred an authorized archive")
	}
	a.selected = true
	c.updates.Store(0)
	if e := f.service.(app.RelayLifecycle).ObserveOwned(ctx); e != nil {
		t.Fatal(e)
	}
	b = findOwned(t, ownedReader(t, f.service), id)
	if b.SourceStatus != "available" {
		t.Fatal("new positive metadata without a fresh activity event was missed")
	}
	a.finalized = true
	a.appendOnly = true
	if e := f.service.(app.RelayLifecycle).ObserveOwned(ctx); e != nil {
		t.Fatal(e)
	}
	b = findOwned(t, ownedReader(t, f.service), id)
	if b.ArchiveFinalized != "confirmed" || b.ArchiveAppendOnly != "confirmed" || b.ExitedAt != nil {
		t.Fatal("explicit finalization evidence depended on process exit")
	}
	a.wrongIdentity = true
	if e := f.service.(app.RelayLifecycle).ObserveOwned(ctx); e != nil {
		t.Fatal(e)
	}
	b = findOwned(t, ownedReader(t, f.service), id)
	if b.SourceStatus != "gap" || b.ArchivePath != "" || b.ArchiveFinalized != "unknown" {
		t.Fatal("conflicting source was allowed or kept closure claims")
	}
	updates := c.updates.Load()
	if e := f.service.(app.RelayLifecycle).ObserveOwned(ctx); e != nil {
		t.Fatal(e)
	}
	if c.updates.Load() != updates {
		t.Fatal("unchanged source conflict rewrote state")
	}
}

func TestRelayOwnedPositiveSourceMustMatchExactSessionCWDAndRoot(t *testing.T) {
	for _, reason := range []string{"cwd", "session", "outside_root", "relative", "format"} {
		t.Run(reason, func(t *testing.T) {
			f, a, _, _ := newOwnedFixture(t, true)
			a.badSource = reason
			id := f.intent()
			if _, err := f.service.ConfirmIntent(context.Background(), caller(), id, confirm()); err != nil {
				t.Fatal(err)
			}
			b := findOwned(t, ownedReader(t, f.service), id)
			if b.SourceStatus != "gap" || b.Gap != "archive_identity_changed" || b.ArchivePath != "" || b.SessionID != "" {
				t.Fatal("positive source bypassed identity/CWD/root checks")
			}
		})
	}
}

func TestRelayOwnedLegacyAndTerminalOnlySourcesStayExplicitGaps(t *testing.T) {
	for _, mode := range []string{"legacy", "terminal"} {
		t.Run(mode, func(t *testing.T) {
			f := newRelayFixture(t)
			ctx := context.Background()
			if mode == "terminal" {
				f.options.Agents[0].DriveMode = "terminal"
				f.service = f.assemble(f.store, nil, f.driver)
			}
			id := f.intent()
			if _, err := f.service.ConfirmIntent(ctx, caller(), id, confirm()); err != nil {
				t.Fatal(err)
			}
			if err := f.store.Update(ctx, func(st *app.RelayState) error {
				if mode == "legacy" {
					delete(st.OwnedSessions, id)
				} else {
					task := st.Tasks[id]
					delete(st.Sessions, task.TerminalID)
					bound := st.OwnedSessions[id]
					bound.SessionID, bound.NativeID, bound.Ownership = "", "", ""
					st.OwnedSessions[id] = bound
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			before, _ := f.store.Snapshot(ctx)
			wire, _ := json.Marshal(before)
			b := findOwned(t, ownedReader(t, f.service), id)
			if b.SourceStatus != "gap" || b.ArchivePath != "" || b.BindingID == "" && mode != "legacy" {
				t.Fatal("legacy/terminal coverage invented native source")
			}
			after, _ := f.store.Snapshot(ctx)
			afterWire, _ := json.Marshal(after)
			if string(wire) != string(afterWire) {
				t.Fatal("read backfilled old binding")
			}
		})
	}
}

func TestRelayOwnedArchiveCannotOverrideWrongControllerProjectOrCLI(t *testing.T) {
	for _, reason := range []string{"project", "cli"} {
		t.Run(reason, func(t *testing.T) {
			f, a, _, _ := newOwnedFixture(t, true)
			a.wrongSessionProject = reason == "project"
			a.wrongSessionCLI = reason == "cli"
			id := f.intent()
			if _, err := f.service.ConfirmIntent(context.Background(), caller(), id, confirm()); err != nil {
				t.Fatal(err)
			}
			b := findOwned(t, ownedReader(t, f.service), id)
			if b.SourceStatus != "gap" || b.ArchivePath != "" || b.SessionID != "" || b.NativeID != "" {
				t.Fatal("matching Archive IDs/CWD overrode rejected native ownership")
			}
			st, err := f.store.Snapshot(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			stored := st.OwnedSessions[id]
			if stored.SourceStatus != "gap" || stored.Gap != "session_ownership_unverified" || stored.ArchivePath != "" {
				t.Fatal("archive source cleared ownership gap in persistent state")
			}
		})
	}
}

func TestRelayOwnedDifferentAgentHandoffKeepsClosedBindingAndCurrentPermission(t *testing.T) {
	f, _, _, _ := newOwnedFixture(t, true)
	ctx := context.Background()
	first := f.intent()
	if _, err := f.service.ConfirmIntent(ctx, caller(), first, confirm()); err != nil {
		t.Fatal(err)
	}
	old := findOwned(t, ownedReader(t, f.service), first)
	if _, err := f.service.Stop(ctx, caller(), first); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Transition(ctx, caller(), first, app.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Transition(ctx, caller(), first, app.RelayTransitionCommand{From: "done", To: "released"}); err != nil {
		t.Fatal(err)
	}
	f.projects.project.Settings.ExternalModelCLI = "beta"
	second := f.intent()
	if _, err := f.service.ConfirmIntent(ctx, caller(), second, confirm()); err != nil {
		t.Fatal(err)
	}
	current := findOwned(t, ownedReader(t, f.service), second)
	if current.TerminalID == old.TerminalID || current.WorktreeID != old.WorktreeID {
		t.Fatal("different Agent reused old terminal or lost neutral workspace")
	}
	state, err := f.store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tail := state.OwnedSessions[first]
	if tail.ExitedAt == nil || tail.BindingID != old.BindingID || tail.ArchivePath != old.ArchivePath || state.Terminals[old.TerminalID].Agent != "alpha" {
		t.Fatal("different Agent overwrote old closed identity")
	}
	denied := findOwned(t, ownedReader(t, f.service), first)
	if denied.Gap != "permission_denied" || denied.ArchivePath != "" {
		t.Fatal("closed source bypassed current Agent authorization")
	}
}

func TestRelayOwnedReadChecksCurrentControllerProjectAndCLI(t *testing.T) {
	for _, reason := range []string{"project", "cli"} {
		t.Run(reason, func(t *testing.T) {
			f, _, _, _ := newOwnedFixture(t, true)
			ctx := context.Background()
			id := f.intent()
			if _, err := f.service.ConfirmIntent(ctx, caller(), id, confirm()); err != nil {
				t.Fatal(err)
			}
			if err := f.store.Update(ctx, func(st *app.RelayState) error {
				task := st.Tasks[id]
				session := st.Sessions[task.TerminalID]
				if reason == "project" {
					session.ProjectID = "foreign"
				} else {
					session.CLI = "foreign"
				}
				st.Sessions[task.TerminalID] = session
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			b := findOwned(t, ownedReader(t, f.service), id)
			if b.SourceStatus != "gap" || b.ArchivePath != "" || b.SessionID != "" {
				t.Fatal("read trusted stale bound identity after current controller mismatch")
			}
		})
	}
}
