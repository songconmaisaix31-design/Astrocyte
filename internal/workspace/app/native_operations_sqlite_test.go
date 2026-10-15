package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestNativeOperationInvalidIdentityDoesNotLaunch(t *testing.T) {
	ctx := context.Background()
	verification, err := filepath.Abs("../../../.verification")
	if err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(verification, "native-operation-actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("owned operation evidence directory=%s", directory)
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: "human", ID: "operation-test-human"}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Human-created operation validation space",
	})
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(filepath.Join(directory, "native"))
	service := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	human := domain.Caller{Kind: "human", ID: "operation-test-human", OperationID: "caller-invalid-operation-input"}
	project, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Public operation validation", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowedActions, settings.ExternalModelCLI = []string{"start"}, "codex"
	if _, err := service.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings}); err != nil {
		t.Fatal(err)
	}
	_, err = service.StartNativeSession(ctx, human, project.ID, domain.NativeCommand{CLI: "codex", Mode: "native"})
	var serviceError *apierrors.ServiceError
	if !errors.As(err, &serviceError) || serviceError.Code != apierrors.ValidationFailed || serviceError.Message != "a verified UUID operation identity is required" {
		t.Fatal("invalid actual command identity reached native launch", err)
	}
	sessions, err := db.ListSessions(ctx, project.ID)
	if err != nil || len(sessions) != 0 {
		t.Fatal("invalid operation created a native session record", err)
	}
	observations, err := registry.SnapshotNative(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range observations {
		if item.Capabilities["start"].Status != "unknown" || item.Startable.Status != "unknown" {
			t.Fatal("invalid operation changed native startup observations")
		}
	}
}
