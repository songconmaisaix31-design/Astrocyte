package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestNativeResumeRetainsCumulativeScopeAndNarrowingDeniesOldContext(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	native := &localNative{}
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{native})
	human := domain.Caller{Kind: "human", ID: "human"}
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "original.txt"), []byte("formerly approved B context"), 0600)
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "scope test", Root: root, SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	settings := p.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	settings.ExternalModelCLI = "codex"
	settings.AllowAgentControl = true
	settings.AllowedActions = []string{"start", "resume", "send", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GrantProjectAgent(ctx, human, p.ID, domain.ProjectGrant{AgentID: "scoped", Actions: []string{"observe", "read_context"}}); err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	session, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex", Context: domain.ContextRequest{Files: []string{"original.txt"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.StopNativeSession(ctx, human, p.ID, session.ID); err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	session, err = service.ResumeNativeSession(ctx, human, p.ID, session.ID, domain.NativeCommand{})
	if err != nil {
		t.Fatal(err)
	}
	if len(session.ContextPacket.Files) != 1 || session.ContextPacket.Files[0].Path != "original.txt" {
		t.Fatal("empty resume forgot prior native context scope")
	}
	settings = p.Settings
	settings.AllowDirectory = false
	settings.AllowedSubdirs = []string{}
	if _, err := service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings}); err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	if _, err := service.SendNativeMessage(ctx, human, p.ID, session.ID, "reexport old context"); err == nil {
		t.Fatal("B revoked after resume but old context exported")
	}
	if _, err := service.ReadNativeContext(ctx, human, p.ID, session.ID); err == nil {
		t.Fatal("B revoked but historical native context still readable")
	}
	if _, err := service.ObserveNativeSession(ctx, human, p.ID, session.ID); err == nil {
		t.Fatal("B revoked but native model event output still readable")
	}
	if _, err := service.ObserveNativeSession(ctx, domain.Caller{Kind: "agent", ID: "scoped", ProjectID: p.ID}, p.ID, session.ID); err == nil {
		t.Fatal("scoped Agent could read old B output after permission narrowing")
	}
	if _, err := service.ResumeNativeSession(ctx, human, p.ID, session.ID, domain.NativeCommand{}); err == nil {
		t.Fatal("B revoked but native full history restored")
	}
	if native.sends != 0 || native.resumes != 1 {
		t.Fatalf("denied operations reached native adapter: sends=%d resumes=%d", native.sends, native.resumes)
	}
}

type blockedNativeObservation struct {
	*localNative
	entered, release chan struct{}
	block            bool
}

func (n *blockedNativeObservation) Observe(ctx context.Context, s domain.NativeSession) (domain.NativeObservation, error) {
	if n.block {
		n.block = false
		close(n.entered)
		<-n.release
	}
	return domain.NativeObservation{Status: "idle", Events: []domain.NativeEvent{}}, nil
}

type observationRegistry struct{ adapter *blockedNativeObservation }

func (r observationRegistry) Adapter(string) (workspace.NativeAdapter, error) { return r.adapter, nil }
func (observationRegistry) List() []string                                    { return []string{"codex"} }

func TestConcurrentObservationCannotEraseAcceptedDeliveryReceipt(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	n := &blockedNativeObservation{localNative: &localNative{}, entered: make(chan struct{}), release: make(chan struct{}), block: true}
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, observationRegistry{n})
	human := domain.Caller{Kind: "human", ID: "human"}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "receipt serialization", Root: t.TempDir(), SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	settings := p.Settings
	settings.ExternalModelCLI, settings.AllowedActions = "codex", []string{"start", "send", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	session, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	observed := make(chan error, 1)
	go func() { _, err := service.ObserveNativeSession(ctx, human, p.ID, session.ID); observed <- err }()
	select {
	case <-n.entered:
	case <-time.After(time.Second):
		t.Fatal("observation did not start")
	}
	sender := human
	sender.OperationID = uuid.NewString()
	sent := make(chan error, 1)
	go func() {
		_, err := service.SendNativeMessage(ctx, sender, p.ID, session.ID, "one authorized message")
		sent <- err
	}()
	select {
	case err := <-sent:
		close(n.release)
		t.Fatalf("send raced unsynchronized stale observation: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(n.release)
	if err := <-observed; err != nil {
		t.Fatal(err)
	}
	if err := <-sent; err != nil {
		t.Fatal(err)
	}
	saved, err := db.LoadSession(ctx, session.ID)
	if err != nil || saved.Operations[sender.OperationID].Status != "accepted" {
		t.Fatal("observation erased durable accepted receipt", err)
	}
	if _, err := service.SendNativeMessage(ctx, sender, p.ID, session.ID, "one authorized message"); err != nil {
		t.Fatal(err)
	}
	if n.sends != 1 {
		t.Fatal("accepted HTTP retry sent another native turn")
	}
}

func TestProjectOldRowsNormalizeAndUnknownHistoryTimeIsNull(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	// Write an old-format row directly to verify read normalization, not only
	// the modern registration/save constructor.
	_, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?)", "old", 1, `{"id":"old","settings":{"revision":1,"history_roots":null}}`)
	if err != nil {
		t.Fatal(err)
	}
	items, err := db.ListProjects(ctx)
	if err != nil || len(items) != 1 || items[0].Settings.HistoryRoots == nil || items[0].Settings.AllowedActions == nil || items[0].Settings.AllowedSubdirs == nil || items[0].Settings.AllowedTools == nil {
		t.Fatal("old settings remain null", err)
	}
	p, err := db.LoadProject(ctx, "old")
	if err != nil || p.Settings.HistoryRoots == nil {
		t.Fatal("old single-project settings remain null", err)
	}
	data, err := json.Marshal(domain.NativeSession{Ownership: "external_observed"})
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if value, exists := document["updated_at"]; !exists || value != nil {
		t.Fatal("unknown historical time became a fake date")
	}
}

func TestNativeResumeCannotCreateSecondProjectWriter(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	native := &localNative{}
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{native})
	human := domain.Caller{Kind: "human", ID: "human"}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "writer bound", Root: t.TempDir(), SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	settings := p.Settings
	settings.ExternalModelCLI = "codex"
	settings.AllowedActions = []string{"start", "resume", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	first, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.StopNativeSession(ctx, human, p.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	if _, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"}); err != nil {
		t.Fatal(err)
	}
	human.OperationID = uuid.NewString()
	if _, err := service.ResumeNativeSession(ctx, human, p.ID, first.ID, domain.NativeCommand{}); err == nil {
		t.Fatal("resume launched a second project writer")
	}
	if native.resumes != 0 {
		t.Fatal("denied resume reached native adapter")
	}
}

func TestNativeUnavailableLaunchDoesNotLeaveGhostWriter(t *testing.T) {
	t.Setenv("PATH", "") // no native CLI/model process can be launched
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, agents.NewRegistry(t.TempDir()))
	human := domain.Caller{Kind: "human", ID: "human"}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "unavailable", Root: t.TempDir(), SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	settings := p.Settings
	settings.ExternalModelCLI = "codex"
	settings.AllowedActions = []string{"start", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		human.OperationID = uuid.NewString()
		session, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"})
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != apierrors.ProviderUnavailable {
			t.Fatalf("unavailable launch %d: %v", i, err)
		}
		if !session.StopConfirmed || session.Ownership != "unstarted" || session.PendingOperation != "" || session.Operations[human.OperationID].Status != "failed" {
			t.Fatalf("unstarted failure left ghost ownership: %+v", session)
		}
		if _, err := service.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"}); err == nil {
			t.Fatal("duplicate failed native start falsely succeeded")
		}
	}
}

func TestProvenUnstartedResumePreservesOwnedOriginalSession(t *testing.T) {
	t.Setenv("PATH", "")
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, agents.NewRegistry(t.TempDir()))
	human := domain.Caller{Kind: "human", ID: "human"}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "original ID retained", Root: t.TempDir(), SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	settings := p.Settings
	settings.ExternalModelCLI, settings.AllowedActions = "codex", []string{"resume", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	// A stored positive-stop fixture, not evidence of a task-live native ID.
	session := domain.NativeSession{ID: uuid.NewString(), ProjectID: p.ID, CLI: "codex", NativeID: "offline-original-id", Ownership: "owned", StopConfirmed: true, Status: "stopped", ContextPacket: domain.ContextPacket{SchemaVersion: 1, ProjectID: p.ID, SettingsRevision: p.Settings.Revision}}
	if err := db.SaveSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		human.OperationID = uuid.NewString()
		result, err := service.ResumeNativeSession(ctx, human, p.ID, session.ID, domain.NativeCommand{})
		var failure *domain.NativePrelaunchFailure
		if !errors.As(err, &failure) {
			t.Fatalf("resume %d did not reach proven unavailable prelaunch: %v", i, err)
		}
		if result.NativeID != session.NativeID || result.Ownership != "owned" || !result.StopConfirmed || result.PendingOperation != "" || result.Operations[human.OperationID].Status != "failed" {
			t.Fatalf("unstarted resume destroyed original durable session: %+v", result)
		}
	}
}
