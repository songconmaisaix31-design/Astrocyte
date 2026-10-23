package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func testAttentionSpaces(db *sqlite.DB) projectspacebridge.Attention {
	return projectspacebridge.Attention{Spaces: workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})}
}

func TestWorkspaceSpaceActualProjectDistillationBoundaries(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: "workspace-space-boundary"}
	principal := attention.Principal{Kind: human.Kind, ID: human.ID}
	spaces := testAttentionSpaces(db)
	meta := func() attention.CommandMeta {
		return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
	}
	space, err := spaces.CreateProjectSpace(ctx, principal, attention.ProjectSpaceCommand{CommandMeta: meta(), Title: "Actual public project fixed material scope"})
	if err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(filepath.Join(directory, "native"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Actual public checkout", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	// 原生 CLI 尚未探测；生产 bridge 和 Registry 保留许可与能力前置检查。
	factory := distillers.NewSelectedTextFactory(&selectedTextBridge{projects: projects, processor: registry}, time.Minute)
	service := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attention.ServiceOptions{ProjectSpaces: spaces, ProjectDistillers: factory})
	imported, err := service.ImportMaterial(ctx, principal, attention.ImportMaterialCommand{CommandMeta: meta(), Adapter: "manual", Kind: "text", SourceLocator: "https://github.com/songconmaisaix31-design/swarm-anything/blob/main/README.md", ExportText: string(readme), Title: "Current public repository README"})
	if err != nil {
		t.Fatal(err)
	}
	if processed, err := service.ProcessNextJob(ctx); err != nil || !processed {
		t.Fatal("actual manual public input was not persisted", err)
	}
	importJob, err := service.GetJob(ctx, principal, imported.JobID)
	if err != nil || importJob.Status != "succeeded" || importJob.MaterialID == nil {
		t.Fatal("actual material import did not finish", err)
	}
	material, err := service.GetMaterial(ctx, principal, *importJob.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	command := attention.RequestDistillationCommand{CommandMeta: meta(), Stage: "project", ProjectID: project.ID, CLI: "codex", InputRefs: []attention.SourceRef{{MaterialID: material.Material.ID, Revision: material.Material.CurrentRevision, Locator: material.Revisions[0].SourceLocator}}}
	expect := func(err error, code apierrors.Code) {
		t.Helper()
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != code {
			t.Fatal("actual production boundary returned the wrong failure", code, err)
		}
	}
	_, err = service.RequestDistillation(ctx, principal, command)
	expect(err, apierrors.ApprovalRequired)
	settings := project.Settings
	settings.ExternalModelCLI = "codex"
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	command.CommandMeta = meta()
	_, err = service.RequestDistillation(ctx, principal, command)
	expect(err, apierrors.ScopeDenied)
	_, err = service.ReferenceMaterial(ctx, principal, space.Space.ID, attention.ReferenceMaterialCommand{CommandMeta: meta(), MaterialID: material.Material.ID, Revision: material.Material.CurrentRevision})
	if err != nil {
		t.Fatal("actual fixed membership was not saved", err)
	}
	command.CommandMeta = meta()
	queued, err := service.RequestDistillation(ctx, principal, command)
	if err != nil {
		t.Fatal("actual included source could not record unavailable processing", err)
	}
	job, err := service.GetJob(ctx, principal, queued.JobID)
	if err != nil || job.Status != "failed" {
		t.Fatal("absent native processor was presented as successful", job.Status, err)
	}
	settings = project.Settings
	settings.ExternalModelCLI = ""
	if _, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings}); err != nil {
		t.Fatal(err)
	}
	command.CommandMeta = meta()
	command.Question = "Explicit request after model permission revocation"
	_, err = service.RequestDistillation(ctx, principal, command)
	expect(err, apierrors.ApprovalRequired)
	for _, table := range []string{"attention_distillations", "local_agent_sessions"} {
		var count int
		if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatal("denied or unavailable processing created an execution result", table, count, err)
		}
	}
}
