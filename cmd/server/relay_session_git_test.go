package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// An actual public executable modifies one public tracked file after a local
// gate, then waits for containment. It is never an installed CLI/model client.
func TestRelaySessionGitProcess(t *testing.T) {
	if os.Getenv("ASTROCYTE_PUBLIC_SESSION_GIT_HELPER") != "1" {
		return
	}
	if _, e := bufio.NewReader(os.Stdin).ReadString('\n'); e != nil {
		os.Exit(8)
	}
	deadline := time.Now().Add(15 * time.Second)
	for {
		if _, e := os.Stat(filepath.Join(".git", "public-go")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			os.Exit(9)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if os.WriteFile("public.txt", []byte("public helper changed tracked file\n"), 0600) != nil || os.WriteFile(filepath.Join(".git", "public-ready"), []byte("public ready"), 0600) != nil {
		os.Exit(10)
	}
	_, _ = io.WriteString(os.Stdout, "PRIVATE_CANARY public helper output\n")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

type relaySessionGitWorkspace struct {
	root  string
	probe *relayworktree.Workspaces
}

func (w *relaySessionGitWorkspace) Prepare(ctx context.Context, _ domain.Work, v domain.RelayWorkspace) (domain.RelayWorkspace, error) {
	v.Root = filepath.Join(w.root, v.ID)
	v.Prepared = true
	v.Branch = "public-fixture"
	if os.MkdirAll(v.Root, 0700) != nil || os.WriteFile(filepath.Join(v.Root, "public.txt"), []byte("public committed fixture baseline\n"), 0600) != nil {
		return v, errors.New("public Git workspace unavailable")
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "public.txt"}, {"-c", "user.name=Public Fixture", "-c", "user.email=public@example.invalid", "commit", "-qm", "public baseline"}} {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-c", "core.hooksPath=", "-C", v.Root}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
		cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
		if cmd.Run() != nil {
			return v, errors.New("public fixture Git command failed")
		}
	}
	return v, nil
}
func (w *relaySessionGitWorkspace) Observe(ctx context.Context, v domain.RelayWorkspace) (string, error) {
	return w.probe.Observe(ctx, v)
}

func TestRelaySessionTerminalGitHelperCommittedHTTP(t *testing.T) {
	ctx := context.Background()
	t.Setenv("ASTROCYTE_PUBLIC_SESSION_GIT_HELPER", "1")
	parent := filepath.Join(relayTestLocalAppData(t), "astrocyte-workbench", "v0.2-access")
	if os.MkdirAll(parent, 0700) != nil {
		t.Fatal("public fixture parent unavailable")
	}
	local, e := os.MkdirTemp(parent, "l1g-")
	if e != nil {
		t.Fatal("short isolated appdata unavailable")
	}
	t.Setenv("LOCALAPPDATA", local)
	root, e := relayRoot()
	if e != nil {
		t.Fatal("isolated root unavailable")
	}
	appDir := filepath.Join(root, "app-state")
	if os.MkdirAll(appDir, 0700) != nil {
		t.Fatal("isolated app unavailable")
	}
	db, e := sqlite.Open(filepath.Join(appDir, "state.sqlite"))
	if e != nil {
		t.Fatal("isolated DB unavailable")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("isolated schema unavailable")
	}
	p := domain.LocalProject{ID: "public-git-project", Root: filepath.Join(root, "source"), Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "public-git-helper", AllowedActions: []string{"start", "observe", "stop"}}}
	if os.MkdirAll(p.Root, 0700) != nil || db.SaveProject(ctx, p, 0) != nil || db.SaveWorksSnapshot(ctx, domain.WorksSnapshot{Works: []domain.Work{{ID: "public-git-work", Root: p.Root, Kind: domain.WorkKindFolder}}}) != nil || db.SaveWorkMetadata(ctx, "public-git-work", domain.WorkHumanMetadata{Revision: 1, Handoff: domain.WorkHandoff{Content: "public helper handoff", Source: "human"}}, 0) != nil {
		t.Fatal("public fixture unavailable")
	}
	store, e := sqlite.OpenRelayStore(root)
	if e != nil {
		t.Fatal("isolated runtime store unavailable")
	}
	defer store.Close()
	executable, e := os.Executable()
	if e != nil {
		t.Fatal("public executable unavailable")
	}
	driver := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{LaunchPolicy: func(_ context.Context, r workspace.RelayLaunch) error {
		if len(r.Spec.LaunchCmd) != 2 || r.Spec.LaunchCmd[0] != executable || r.Spec.LaunchCmd[1] != "-test.run=^TestRelaySessionGitProcess$" {
			return errors.New("only public helper allowed")
		}
		return nil
	}})
	neutralRoot := filepath.Join(root, "worktrees")
	neutral := &relaySessionGitWorkspace{root: neutralRoot, probe: relayworktree.New(neutralRoot, relayworktree.Options{})}
	repo := sqlite.NewEventRepository(db)
	eventService := attention.NewPendingEventService(repo)
	runtime := workspace.NewRelayService(store, workspace.NewWorksService(db, nil, nil, nil, workspace.WorksServiceOptions{}), db, neutral, nil, driver, events.NewRelayEventSink(eventService, nil), workspace.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "public-git-helper", LaunchCmd: []string{executable, "-test.run=^TestRelaySessionGitProcess$"}, DriveMode: "terminal"}}, OwnedRoots: agents.RelayOwnedRoots{}})
	collector := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), events.NewRelaySessionCapture(eventService, nil), relaySessionFileState{root: root}, sqlite.NewAccessSessionReader(db))
	wrapped := &relaySessionRuntime{RelayRuntime: runtime, RelayLifecycle: runtime.(workspace.RelayLifecycle), collector: collector}
	defer func() {
		bounded, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = wrapped.Shutdown(bounded)
	}()
	client := &relayPersistentHTTP{t: t, server: httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Relay: wrapped, RelaySessionEvents: collector}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())}
	defer client.server.Close()
	client.request("GET", "/api/v1/auth/session", "", "", 200)
	intent := client.request("POST", "/api/v1/relay/intents", `{"work_id":"public-git-work","project_id":"public-git-project","agent":"public-git-helper","task":"public helper controlled edit"}`, "", 200)
	var ready struct {
		View domain.RelayTaskView `json:"view"`
	}
	if json.Unmarshal(intent, &ready) != nil {
		t.Fatal("public intent unavailable")
	}
	accepted := client.request("POST", "/api/v1/relay/intents/"+ready.View.Task.ID+"/confirm", `{"permission_confirmed":true,"permission_level":"A","project_settings_revision":1}`, "", 200)
	var active struct {
		View domain.RelayTaskView `json:"view"`
	}
	if json.Unmarshal(accepted, &active) != nil {
		t.Fatal("public launch unavailable")
	}
	workspaceRoot := filepath.Join(neutralRoot, active.View.Workspace.ID)
	if wrapped.ObserveOwned(ctx) != nil {
		t.Fatal("initial public Git baseline observation unavailable")
	}
	if os.WriteFile(filepath.Join(workspaceRoot, ".git", "public-go"), []byte("public go"), 0600) != nil {
		t.Fatal("public helper gate unavailable")
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, e = os.Stat(filepath.Join(workspaceRoot, ".git", "public-ready")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("public helper did not edit tracked fixture")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if wrapped.ObserveOwned(ctx) != nil {
		t.Fatal("existing terminal Git observation unavailable")
	}
	data := client.request("GET", "/api/v1/relay/session-events?task_id="+ready.View.Task.ID, "", "", 200)
	var view workbench.SessionEvents
	if json.Unmarshal(data, &view) != nil || view.Total < 2 || len(view.Coverage) != 1 || view.Coverage[0].SessionComplete || view.Coverage[0].Status != "unknown" {
		t.Fatal("terminal Git receipt missing or native completeness invented")
	}
	gitEvents := 0
	for _, event := range view.Events {
		receipt, e := repo.GetEvent(ctx, event.ID)
		if e != nil || receipt.ID != event.ID || receipt.Event != nil {
			t.Fatal("Git HTTP ID not a pending committed receipt")
		}
		if event.Kind == "git_activity" {
			gitEvents++
		}
	}
	if gitEvents < 2 {
		t.Fatal("helper edit not visible through existing Git sink")
	}
	current, e := runtime.GetTask(ctx, domain.Caller{Kind: "human", ID: "local-human"}, ready.View.Task.ID)
	if e != nil || current.Task.State == domain.RelayDone {
		t.Fatal("Git activity changed task to done")
	}
	samples, _ := json.Marshal([]relayContractSample{{Schema: "RelaySessionEventsV1", Body: data}})
	if os.WriteFile(filepath.Join(local, "terminal-git-http-derived.json"), samples, 0600) != nil {
		t.Fatal("public Git HTTP sample unavailable")
	}
}
