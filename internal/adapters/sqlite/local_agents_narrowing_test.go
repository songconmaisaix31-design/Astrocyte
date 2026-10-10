package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

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
	settings.AllowedActions = []string{"start", "resume", "send", "stop"}
	p, err = service.SetProjectSettings(ctx, human, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: settings})
	if err != nil {
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
	if _, err := service.ResumeNativeSession(ctx, human, p.ID, session.ID, domain.NativeCommand{}); err == nil {
		t.Fatal("B revoked but native full history restored")
	}
	if native.sends != 0 || native.resumes != 1 {
		t.Fatalf("denied operations reached native adapter: sends=%d resumes=%d", native.sends, native.resumes)
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
	}
}
