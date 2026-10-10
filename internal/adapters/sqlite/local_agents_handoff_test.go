package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestLocalContextHandoffRequiresStoppedSameProjectAndSurvivesRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "handoff.sqlite")
	db := openAttentionDB(t, path)
	refs := &localReferences{}
	native := &localNative{}
	service := workspace.NewLocalProjectService(db, refs, agents.ProjectFiles{}, localRegistry{native})
	human := domain.Caller{Kind: "human", ID: "browser-human"}
	register := func(name string) domain.LocalProject {
		t.Helper()
		project, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: name, Root: t.TempDir(), SpaceID: name})
		if err != nil {
			t.Fatal(err)
		}
		settings := project.Settings
		settings.ExternalModelCLI = "codex"
		settings.AllowedActions = []string{"start", "stop", "observe"}
		project, err = service.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: 1, Settings: settings})
		if err != nil {
			t.Fatal(err)
		}
		return project
	}
	project, other := register("selected-project"), register("other-project")
	human.OperationID = uuid.NewString()
	source, err := service.StartNativeSession(ctx, human, project.ID, domain.NativeCommand{CLI: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	handoff := domain.NativeCommand{CLI: "codex", Mode: "context_handoff", SourceSessionID: source.ID, Context: domain.ContextRequest{References: []domain.FixedReference{{MaterialID: "fixed", Revision: 1}}}}
	human.OperationID = uuid.NewString()
	if _, err := service.StartNativeSession(ctx, human, project.ID, handoff); err == nil || native.starts != 1 {
		t.Fatal("unconfirmed source stop launched a handoff")
	}
	native.failStop = true
	if _, err := service.StopNativeSession(ctx, human, project.ID, source.ID); err == nil {
		t.Fatal("fixture did not expose unverified stop")
	}
	human.OperationID = uuid.NewString()
	if _, err := service.StartNativeSession(ctx, human, project.ID, handoff); err == nil || native.starts != 1 {
		t.Fatal("failed stop allowed another native session")
	}
	native.failStop = false
	if stopped, err := service.StopNativeSession(ctx, human, project.ID, source.ID); err != nil || !stopped.StopConfirmed {
		t.Fatalf("positive source stop: %+v %v", stopped, err)
	}
	human.OperationID = uuid.NewString()
	if _, err := service.StartNativeSession(ctx, human, other.ID, handoff); err == nil || native.starts != 1 {
		t.Fatal("human cross-project source silently transferred context")
	}
	human.OperationID = uuid.NewString()
	created, err := service.StartNativeSession(ctx, human, project.ID, handoff)
	if err != nil || native.starts != 2 || native.resumes != 0 {
		t.Fatalf("explicit new-session handoff: %+v %v", created, err)
	}
	if created.ID == source.ID || created.SourceSessionID != source.ID || created.Mode != "context_handoff" || len(created.ContextPacket.Materials) != 1 {
		t.Fatal("handoff identity or fixed selected context was lost")
	}
	if stopped, err := service.StopNativeSession(ctx, human, project.ID, created.ID); err != nil || !stopped.StopConfirmed {
		t.Fatalf("handoff cleanup: %+v %v", stopped, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	defer db.Close()
	restored, err := db.LoadSession(ctx, created.ID)
	if err != nil || restored.Mode != "context_handoff" || restored.SourceSessionID != source.ID || !restored.StopConfirmed {
		t.Fatalf("SQLite restart lost explicit handoff provenance: %+v %v", restored, err)
	}
	if len(restored.ContextPacket.Materials) != 1 || restored.ContextPacket.Materials[0].Text != "" {
		t.Fatal("handoff persisted private context body or lost fixed reference")
	}
}
