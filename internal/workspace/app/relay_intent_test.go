package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayIntentBindsSQLiteSettingsAcrossReload(t *testing.T) {
	for _, level := range []string{"A", "B", "C"} {
		t.Run(level, func(t *testing.T) {
			f := newRelayFixture(t)
			ctx := context.Background()
			db, err := sqlite.Open(filepath.Join(t.TempDir(), "application.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err = db.RunMigrations(ctx, migrations.FS); err != nil {
				t.Fatal(err)
			}
			registry := agents.NewRegistry(filepath.Join(t.TempDir(), "controller"))
			projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
			f.options.Agents = []domain.AgentSpec{{Name: "codex", DriveMode: "structured"}}
			project, err := projects.RegisterProject(ctx, caller(), domain.RegisterProjectCommand{Name: "Public intent settings", Root: f.works.work.Root, SpaceID: "public"})
			if err != nil {
				t.Fatal(err)
			}
			settings := domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: []string{"start"}, AllowDirectory: level == "B" || level == "C", ExpandReferences: level == "C"}
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			assemble := func() app.RelayRuntime {
				return app.NewRelayService(f.store, f.works, db, f.spaces, f.driver, nil, nil, f.options)
			}
			service := assemble()
			intentCmd := app.RelayIntentCommand{WorkID: "work", ProjectID: project.ID, Agent: "codex", Task: "Public settings binding contract"}
			intent, err := service.CreateIntent(ctx, caller(), intentCmd)
			if err != nil {
				t.Fatal(err)
			}
			if intent.Task.SettingsRevision != project.Settings.Revision || intent.Task.PermissionLevel != level || f.spaces.prepares.Load() != 0 || f.driver.launches.Load() != 0 {
				t.Fatal("intent did not bind current SQLite settings without effects")
			}
			if err = f.store.Close(); err != nil {
				t.Fatal(err)
			}
			f.store, err = sqlite.OpenRelayStore(f.root)
			if err != nil {
				t.Fatal(err)
			}
			service = assemble()
			reloaded, err := service.GetTask(ctx, caller(), intent.Task.ID)
			if err != nil || reloaded.Task.SettingsRevision != intent.Task.SettingsRevision || reloaded.Task.PermissionLevel != level {
				t.Fatal("SQLite reload lost original intent permission")
			}
			wrong := app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: level, ProjectSettingsRevision: project.Settings.Revision + 1}
			_, err = service.ConfirmIntent(ctx, caller(), intent.Task.ID, wrong)
			checkCode(t, err, apierrors.VersionConflict)
			// Persist an actual settings revision change with the same A/B/C label.
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			current := app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: level, ProjectSettingsRevision: project.Settings.Revision}
			op := caller()
			rejected, err := service.ConfirmIntent(ctx, op, intent.Task.ID, current)
			checkCode(t, err, apierrors.ApprovalRevoked)
			if rejected.Receipt.Status != "rejected" || rejected.View.Task.SettingsRevision != intent.Task.SettingsRevision || rejected.View.Task.State != domain.RelayReady || f.spaces.prepares.Load() != 0 || f.driver.launches.Load() != 0 {
				t.Fatal("updated confirmation bypassed original intent or had effects")
			}
			_, err = service.ConfirmIntent(ctx, op, intent.Task.ID, current)
			checkCode(t, err, apierrors.ApprovalRevoked)
			fresh, err := service.CreateIntent(ctx, caller(), intentCmd)
			if err != nil || fresh.Task.SettingsRevision != project.Settings.Revision {
				t.Fatal("fresh intent did not bind changed settings")
			}
			// This driver is a contract fixture; actual HTTP/native acceptance is
			// performed by H against its single owned production service.
			confirmed, err := service.ConfirmIntent(ctx, caller(), fresh.Task.ID, current)
			if err != nil || confirmed.Receipt.Status != "accepted" || f.driver.launches.Load() != 1 {
				t.Fatal("current intent could not confirm exactly once")
			}
		})
	}
}

func TestRelayIntentRejectsPermissionLabelChangeWithoutRevisionRewrite(t *testing.T) {
	f := newRelayFixture(t)
	id := f.intent()
	f.projects.project.Settings.AllowDirectory = true
	cmd := confirm()
	cmd.PermissionLevel = "B"
	result, err := f.service.ConfirmIntent(context.Background(), caller(), id, cmd)
	checkCode(t, err, apierrors.ApprovalRevoked)
	if result.View.Task.PermissionLevel != "A" || f.driver.launches.Load() != 0 || f.spaces.prepares.Load() != 0 {
		t.Fatal("changed scope replaced original permission or launched")
	}
}
