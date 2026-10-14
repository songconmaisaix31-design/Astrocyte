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

// 使用生产适配器验证权限拒绝；工作状态和退出成功由 H/F 的真实控制器验证。
func TestRelayHumanTransitionProductionPermissionGates(t *testing.T) {
	for _, level := range []string{"A", "B", "C"} {
		t.Run(level, func(t *testing.T) {
			ctx := context.Background()
			evidence, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "relay-l1"))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.MkdirAll(evidence, 0700); err != nil {
				t.Fatal(err)
			}
			root, err := os.MkdirTemp(evidence, "transition-permissions-")
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(root, "public-git")
			if err = os.MkdirAll(source, 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(source, "README.md"), []byte("# Public Relay permission documentation\n\nHuman completion and release require observe permission, explicit confirmation and verified lifecycle state. Event publication reconciliation requires reconcile permission.\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"init", source}, {"-C", source, "add", "README.md"}, {"-C", source, "-c", "user.name=Relay permission documentation", "-c", "user.email=relay-permissions@example.invalid", "-c", "core.hooksPath=", "commit", "-m", "Document public human transition permissions"}} {
				if err = exec.Command("git", args...).Run(); err != nil {
					t.Fatal("public permission documentation commit failed")
				}
			}
			db, err := sqlite.Open(filepath.Join(root, "application.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err = db.RunMigrations(ctx, migrations.FS); err != nil {
				t.Fatal(err)
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
			registry := agents.NewRegistry(filepath.Join(root, "controller"))
			projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
			project, err := projects.RegisterProject(ctx, caller(), domain.RegisterProjectCommand{Name: "Public transition permission documentation", Root: source, SpaceID: "public"})
			if err != nil {
				t.Fatal(err)
			}
			settings := domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: []string{"observe"}, AllowDirectory: level == "B" || level == "C", ExpandReferences: level == "C"}
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			works := app.NewWorksService(db, nil, worksscan.Delivery{}, nil, app.WorksServiceOptions{})
			delivery, err := json.Marshal(map[string]any{"works": []any{map[string]any{"platform": "codex", "name": "Public transition permission documentation", "path": project.Root, "kind": "folder", "handoff": "Check the permissions documented in the public README."}}})
			if err != nil {
				t.Fatal(err)
			}
			album, err := works.ImportWorksDelivery(ctx, caller(), app.WorksDeliveryCommand{Content: delivery})
			if err != nil || len(album.Works) != 1 {
				t.Fatal("public work import failed")
			}
			service := app.NewRelayService(store, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), nil, nil, app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}})
			intent, err := service.CreateIntent(ctx, caller(), app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "codex", Task: "Check the documented human transition permissions"})
			if err != nil || intent.Task.SettingsRevision != project.Settings.Revision || intent.Task.PermissionLevel != level {
				t.Fatal("original intent permission binding failed")
			}
			cmd := app.RelayTransitionCommand{From: domain.RelayWorking, To: domain.RelayDone, CompletionConfirmed: true}
			// observe 许可通过后，由真实 ready 状态拒绝未派发任务的完成请求。
			_, err = service.Transition(ctx, caller(), intent.Task.ID, cmd)
			checkCode(t, err, apierrors.VersionConflict)
			relayTransitionPermissionHTTP(t, service, intent.Task.ID, cmd, http.StatusConflict, apierrors.VersionConflict)
			_, err = service.(app.RelayLifecycle).ReconcileEvents(ctx, caller(), intent.Task.ID)
			checkCode(t, err, apierrors.ApprovalRequired)
			if level == "A" {
				relayFactsAcrossSQLiteConnections(t, ctx, store, stateRoot, service, works, db, registry, root, intent.Task.ID)
			}
			before, err := store.Snapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.Transition(ctx, domain.Caller{Kind: "agent", ID: "codex", OperationID: uuid.NewString()}, intent.Task.ID, cmd)
			checkCode(t, err, apierrors.ScopeDenied)
			after, err := store.Snapshot(ctx)
			if err != nil || len(after.Receipts) != len(before.Receipts) {
				t.Fatal("Agent transition wrote a receipt")
			}
			if _, err = projects.GrantProjectAgent(ctx, caller(), project.ID, domain.ProjectGrant{AgentID: "codex", Actions: []string{"start"}}); err != nil {
				t.Fatal(err)
			}
			_, err = service.Transition(ctx, caller(), intent.Task.ID, cmd)
			checkCode(t, err, apierrors.ScopeDenied)
			if _, err = projects.RevokeProjectAgent(ctx, caller(), project.ID, "codex"); err != nil {
				t.Fatal(err)
			}
			_, err = service.Transition(ctx, caller(), intent.Task.ID, cmd)
			checkCode(t, err, apierrors.ApprovalRevoked)
			relayTransitionPermissionHTTP(t, service, intent.Task.ID, cmd, http.StatusForbidden, apierrors.ApprovalRevoked)
			project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.Transition(ctx, caller(), intent.Task.ID, cmd)
			checkCode(t, err, apierrors.ApprovalRevoked)
			reloaded, err := service.GetTask(ctx, caller(), intent.Task.ID)
			if err != nil || reloaded.Task.State != domain.RelayReady || reloaded.Task.SettingsRevision != intent.Task.SettingsRevision || reloaded.Task.PermissionLevel != level {
				t.Fatal("rejected transition changed the original intent")
			}
			state, err := store.Snapshot(ctx)
			if err != nil || len(state.Workspaces) != 0 || len(state.Terminals) != 0 || len(state.Sessions) != 0 || len(state.OwnedSessions) != 0 || len(state.PendingActions) != 0 {
				t.Fatal("permission check created execution state")
			}
			if level == "A" {
				handoffService := app.NewRelayService(store, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), nil, nil, app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured"}, {Name: "claude", DriveMode: "structured"}}})
				settings.ExternalModelCLI = "claude"
				project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
				if err != nil {
					t.Fatal(err)
				}
				cases := []struct {
					name, from string
					confirmed  bool
					code       apierrors.Code
					status     int
				}{
					{"missing_confirmation", intent.Task.ID, false, apierrors.ValidationFailed, http.StatusBadRequest},
					{"missing_source", "", true, apierrors.ValidationFailed, http.StatusBadRequest},
					{"invalid_source", "invalid", true, apierrors.ValidationFailed, http.StatusBadRequest},
					{"absent_source", uuid.NewString(), true, apierrors.NotFound, http.StatusNotFound},
					{"unreleased_source", intent.Task.ID, true, apierrors.VersionConflict, http.StatusConflict},
				}
				for _, test := range cases {
					command := app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "claude", Task: "Read the public human handoff documentation", HandoffFromTaskID: test.from, HandoffContextConfirmed: test.confirmed}
					_, err = handoffService.CreateIntent(ctx, caller(), command)
					checkCode(t, err, test.code)
					relayPermissionHTTP(t, handoffService, "/api/v1/relay/intents", command, test.status, test.code)
				}
				command := app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "claude", Task: "Read public documentation in a new session"}
				ordinary, err := handoffService.CreateIntent(ctx, caller(), command)
				if err != nil || ordinary.Task.State != domain.RelayReady {
					t.Fatal("ordinary intent without a handoff declaration failed")
				}
				after, err := store.Snapshot(ctx)
				if err != nil || len(after.Tasks) != 2 || len(after.HandoffDeclarations) != 0 || len(after.ControllerEvidence) != 0 || len(after.SessionRelations) != 0 || len(after.Workspaces) != 0 {
					t.Fatal("rejected handoff created a relationship or execution")
				}
				t.Log("production HTTP and SQLite rejected five invalid handoff declarations; an omitted declaration created only an ordinary ready intent")
			}
		})
	}
}

func relayTransitionPermissionHTTP(t *testing.T, service app.RelayRuntime, id string, cmd app.RelayTransitionCommand, status int, code apierrors.Code) {
	relayPermissionHTTP(t, service, "/api/v1/relay/tasks/"+id+"/transition", cmd, status, code)
}

func relayPermissionHTTP(t *testing.T, service app.RelayRuntime, path string, cmd any, status int, code apierrors.Code) {
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
		t.Fatal("HTTP session bootstrap failed")
	}
	body, err := json.Marshal(cmd)
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
	var denied struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if resp.StatusCode != status || json.NewDecoder(resp.Body).Decode(&denied) != nil || denied.Error.Code != string(code) {
		t.Fatal("production HTTP transition did not preserve the expected denial")
	}
}
