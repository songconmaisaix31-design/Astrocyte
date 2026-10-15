package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayIntentBindsSQLiteSettingsAcrossReload(t *testing.T) {
	for _, level := range []string{"A", "B", "C"} {
		t.Run(level, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			source := filepath.Join(root, "public-git")
			if err := os.MkdirAll(source, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("# Public Relay permission check\n\nThis repository documents the intent permission gate. No model task is dispatched by this check.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"init", source}, {"-C", source, "add", "README.md"}, {"-C", source, "-c", "user.name=Relay permission check", "-c", "user.email=relay-permissions@example.invalid", "-c", "core.hooksPath=", "commit", "-m", "Document the public permission check"}} {
				if err := exec.Command("git", args...).Run(); err != nil {
					t.Fatal("public documentation Git creation failed")
				}
			}
			t.Setenv("LOCALAPPDATA", filepath.Join(root, "local-state"))
			stateRoot, err := sqlite.RelayStateRoot()
			if err != nil {
				t.Fatal(err)
			}
			store, err := sqlite.OpenRelayStore(stateRoot)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			db, err := sqlite.Open(filepath.Join(root, "application.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err = db.RunMigrations(ctx, migrations.FS); err != nil {
				t.Fatal(err)
			}
			registry := agents.NewRegistry(filepath.Join(root, "controller"))
			projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
			project, err := projects.RegisterProject(ctx, caller(), domain.RegisterProjectCommand{Name: "Public Relay permission check", Root: source, SpaceID: "public"})
			if err != nil {
				t.Fatal(err)
			}
			settings := domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: []string{"start"}, AllowDirectory: level == "B" || level == "C", ExpandReferences: level == "C"}
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			works := app.NewWorksService(db, nil, worksscan.Delivery{}, nil, app.WorksServiceOptions{})
			delivery, err := json.Marshal(map[string]any{"works": []any{map[string]any{"platform": "codex", "name": "Public Relay permission check", "path": project.Root, "kind": "folder", "handoff": "Use only the public README. This check verifies permissions without dispatching a model task."}}})
			if err != nil {
				t.Fatal(err)
			}
			album, err := works.ImportWorksDelivery(ctx, caller(), app.WorksDeliveryCommand{Content: delivery})
			if err != nil || len(album.Works) != 1 {
				t.Fatal("actual public works import failed")
			}
			worktrees := filepath.Join(root, "worktrees")
			spaces := relayworktree.New(worktrees, relayworktree.Options{AllowCommittedGit: true})
			driver := agents.NewRelayNativeDriver(registry, agents.RelayNativeDriverOptions{CaptureNativeNotifications: true})
			assemble := func() app.RelayRuntime {
				return app.NewRelayService(store, works, db, spaces, driver, nil, nil, app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}})
			}
			service := assemble()
			intentCmd := app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "codex", Task: "Check the public README permission binding"}
			intent, err := service.CreateIntent(ctx, caller(), intentCmd)
			if err != nil {
				t.Fatal(err)
			}
			if intent.Task.SettingsRevision != project.Settings.Revision || intent.Task.PermissionLevel != level {
				t.Fatal("intent did not bind current SQLite settings")
			}
			if err = store.Close(); err != nil {
				t.Fatal(err)
			}
			store, err = sqlite.OpenRelayStore(stateRoot)
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
			// 通过生产项目服务真实保存一次版本变化，A/B/C 标签保持相同。
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			current := app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: level, ProjectSettingsRevision: project.Settings.Revision}
			op := caller()
			rejected, err := service.ConfirmIntent(ctx, op, intent.Task.ID, current)
			checkCode(t, err, apierrors.ApprovalRevoked)
			if rejected.Receipt.Status != "rejected" || rejected.View.Task.SettingsRevision != intent.Task.SettingsRevision || rejected.View.Task.State != domain.RelayReady {
				t.Fatal("updated confirmation replaced original intent")
			}
			_, err = service.ConfirmIntent(ctx, op, intent.Task.ID, current)
			checkCode(t, err, apierrors.ApprovalRevoked)
			// 真实 HTTP guard、项目来源和 SQLite；拒绝请求不进入原生驱动。
			relayIntentHTTPDenied(t, service, intentCmd, intent.Task.ID, current, project.Settings.Revision, level)
			state, err := store.Snapshot(ctx)
			if err != nil || len(state.Workspaces) != 0 || len(state.Terminals) != 0 || len(state.Sessions) != 0 || len(state.OwnedSessions) != 0 || len(state.PendingActions) != 0 {
				t.Fatal("denied intent created execution state")
			}
			if _, err = os.Stat(worktrees); !os.IsNotExist(err) {
				t.Fatal("denied intent prepared a real worktree")
			}
			// 成功派发由 H/F 的独占真实控制器服务执行，本检查不启动 CLI 或模型。
		})
	}
}

func relayIntentHTTPDenied(t *testing.T, service app.RelayRuntime, cmd app.RelayIntentCommand, staleID string, confirm app.RelayConfirmCommand, revision int, level string) {
	t.Helper()
	server := httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Relay: service}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())
	defer server.Close()
	bootstrap, err := http.Get(server.URL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	err = json.NewDecoder(bootstrap.Body).Decode(&session)
	_ = bootstrap.Body.Close()
	if err != nil || len(bootstrap.Cookies()) != 1 || session.CSRF == "" {
		t.Fatal("actual HTTP session bootstrap failed")
	}
	post := func(path string, input any, status int, output any) {
		t.Helper()
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader(string(body)))
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(bootstrap.Cookies()[0])
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", session.CSRF)
		req.Header.Set("Idempotency-Key", uuid.NewString())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != status || json.NewDecoder(resp.Body).Decode(output) != nil {
			t.Fatalf("actual relay HTTP status %d, expected %d", resp.StatusCode, status)
		}
	}
	var created struct {
		View struct {
			Task domain.RelayTask `json:"task"`
		} `json:"view"`
	}
	post("/api/v1/relay/intents", cmd, http.StatusOK, &created)
	if created.View.Task.SettingsRevision != revision || created.View.Task.PermissionLevel != level {
		t.Fatal("HTTP intent did not expose real current SQLite revision/scope")
	}
	var denied struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	post("/api/v1/relay/intents/"+staleID+"/confirm", confirm, http.StatusForbidden, &denied)
	if denied.Error.Code != string(apierrors.ApprovalRevoked) {
		t.Fatal("actual HTTP stale intent did not preserve approval_revoked")
	}
}
