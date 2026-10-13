package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// A real public helper executable, not a model client. Input evidence stays in
// the isolated fixture worktree; its stdout canary must never become an HTTP DTO.
func TestRelayHTTPEOFProcess(t *testing.T) {
	if os.Getenv("ASTROCYTE_HTTP_EOF_HELPER") != "1" {
		return
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil || os.WriteFile("public-http-eof-input.txt", input, 0600) != nil {
		os.Exit(7)
	}
	_, _ = io.WriteString(os.Stdout, "PRIVATE_CANARY EOF observed\n")
}

type relayEOFFixtureWorkspace struct{ relayControlledWorkspace }

func (w *relayEOFFixtureWorkspace) Prepare(ctx context.Context, work domain.Work, v domain.RelayWorkspace) (domain.RelayWorkspace, error) {
	v, err := w.relayControlledWorkspace.Prepare(ctx, work, v)
	if err == nil {
		err = os.MkdirAll(v.Root, 0700)
	}
	return v, err
}

func TestRelayRealHTTPEOFFirstInputAndUnsupportedReceipt(t *testing.T) {
	t.Setenv("ASTROCYTE_HTTP_EOF_HELPER", "1")
	ctx := context.Background()
	evidence := filepath.Join(relayTestLocalAppData(t), "astrocyte-workbench", "v0.2-access", "verification")
	// Keep this child appdata under the access root without repeating the long
	// verification/eof-http/UUID scaffold. Windows process cwd has a tighter
	// bound than file operations; an already-isolated caller must remain usable.
	fixtureParent := filepath.Dir(evidence)
	if os.MkdirAll(evidence, 0700) != nil {
		t.Fatal("isolated EOF evidence directory unavailable")
	}
	local, err := os.MkdirTemp(fixtureParent, "")
	if err != nil {
		t.Fatal("isolated EOF state unavailable")
	}
	t.Setenv("LOCALAPPDATA", local)
	root, err := relayRoot()
	if err != nil {
		t.Fatal("isolated EOF root unavailable")
	}
	appDir := filepath.Join(root, "app-state")
	if os.MkdirAll(appDir, 0700) != nil {
		t.Fatal("isolated app directory unavailable")
	}
	db, err := sqlite.Open(filepath.Join(appDir, "state.sqlite"))
	if err != nil {
		t.Fatal("isolated app DB unavailable")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("isolated app schema unavailable")
	}
	p := domain.LocalProject{ID: "eof-project", Root: filepath.Join(root, "source"), Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "fixture-eof", AllowedActions: []string{"start", "send", "observe", "stop"}}}
	if db.SaveProject(ctx, p, 0) != nil || db.SaveWorksSnapshot(ctx, domain.WorksSnapshot{Works: []domain.Work{{ID: "eof-work", Root: p.Root, Kind: domain.WorkKindFolder}}}) != nil || db.SaveWorkMetadata(ctx, "eof-work", domain.WorkHumanMetadata{Revision: 1, Handoff: domain.WorkHandoff{Content: "complete HTTP manual\nsecond line", Source: "human"}}, 0) != nil {
		t.Fatal("public EOF fixture unavailable")
	}
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		t.Fatal("isolated relay DB unavailable")
	}
	defer store.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal("public helper executable unavailable")
	}
	terminal := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{LaunchPolicy: func(_ context.Context, r workspace.RelayLaunch) error {
		if len(r.Spec.LaunchCmd) != 2 || r.Spec.LaunchCmd[0] != executable || r.Spec.LaunchCmd[1] != "-test.run=^TestRelayHTTPEOFProcess$" {
			return errors.New("only public fixture executable allowed")
		}
		return nil
	}})
	neutral := &relayEOFFixtureWorkspace{relayControlledWorkspace{root: filepath.Join(root, "worktrees")}}
	runtime := workspace.NewRelayService(store, workspace.NewWorksService(db, nil, nil, nil, workspace.WorksServiceOptions{}), db, neutral, nil, terminal, events.NewRelayEventSink(attention.NewPendingEventService(sqlite.NewEventRepository(db)), time.Now), workspace.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "fixture-eof", LaunchCmd: []string{executable, "-test.run=^TestRelayHTTPEOFProcess$"}, DriveMode: "terminal", InputMode: "stdin_eof"}}, Clock: time.Now})
	lifecycle := runtime.(workspace.RelayLifecycle)
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = lifecycle.Shutdown(cleanup)
	}()
	client := &relayPersistentHTTP{t: t, server: httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Relay: runtime}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())}
	defer client.server.Close()
	client.request("GET", "/api/v1/auth/session", "", "", 200)
	b := client.request("GET", "/api/v1/relay/agents", "", "", 200)
	var registry struct {
		Agents []domain.AgentSpec `json:"agents"`
	}
	if json.Unmarshal(b, &registry) != nil || len(registry.Agents) != 1 || registry.Agents[0].InputMode != "stdin_eof" {
		t.Fatal("actual runtime registry query lost optional input mode")
	}
	b = client.request("POST", "/api/v1/relay/intents", `{"work_id":"eof-work","project_id":"eof-project","agent":"fixture-eof","task":"explicit HTTP task"}`, "", 200)
	var task struct {
		View domain.RelayTaskView `json:"view"`
	}
	if json.Unmarshal(b, &task) != nil || task.View.Task.State != "ready" || neutral.prepared != 0 {
		t.Fatal("EOF intent launched before human confirmation")
	}
	id := task.View.Task.ID
	client.request("POST", "/api/v1/relay/intents/"+id+"/confirm", `{"permission_confirmed":true,"permission_level":"A","project_settings_revision":1}`, "", 200)
	b = client.request("GET", "/api/v1/relay/tasks/"+id, "", "", 200)
	if json.Unmarshal(b, &task) != nil || task.View.Workspace == nil || task.View.Terminal == nil {
		t.Fatal("EOF dispatch view unavailable")
	}
	inputFile := filepath.Join(task.View.Workspace.Root, "public-http-eof-input.txt")
	want := "HANDOFF_MANUAL:\ncomplete HTTP manual\nsecond line\n\nTASK:\nexplicit HTTP task"
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, e := os.ReadFile(inputFile)
		if e == nil && string(data) == want {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	before, e := os.ReadFile(inputFile)
	if e != nil || string(before) != want {
		t.Fatal("real helper did not receive complete first input followed by EOF")
	}
	key := uuid.NewString()
	b = client.request("POST", "/api/v1/relay/tasks/"+id+"/send", `{"input":"must never append"}`, key, 501)
	client.samples = append(client.samples, relayContractSample{"ErrorV1", append(json.RawMessage(nil), b...)})
	var failure struct {
		Error struct {
			Code      string `json:"code"`
			Retryable bool   `json:"retryable"`
		} `json:"error"`
	}
	if json.Unmarshal(b, &failure) != nil || failure.Error.Code != "unsupported_capability" || failure.Error.Retryable {
		t.Fatal("EOF follow-up did not return safe unsupported result")
	}
	client.request("POST", "/api/v1/relay/tasks/"+id+"/send", `{"input":"must never append"}`, key, 501)
	client.request("POST", "/api/v1/relay/tasks/"+id+"/read", `{}`, "", 200)
	client.request("POST", "/api/v1/relay/tasks/"+id+"/stop", `{}`, "", 200)
	b = client.request("GET", "/api/v1/relay/tasks/"+id+"/receipts", "", "", 200)
	var result struct {
		Receipts []domain.RelayReceipt `json:"receipts"`
	}
	if json.Unmarshal(b, &result) != nil {
		t.Fatal("EOF receipts unavailable")
	}
	denied, read, stop := 0, 0, 0
	for _, r := range result.Receipts {
		if r.OperationID == key && r.Status == "rejected" && r.Error != nil && r.Error.Code == "unsupported_capability" {
			denied++
		}
		if r.Action == "read" && r.Status == "accepted" {
			read++
		}
		if r.Action == "stop" && r.Status == "accepted" {
			stop++
		}
	}
	if denied != 1 || read != 1 || stop != 1 || neutral.prepared != 1 {
		t.Fatal("EOF first refusal or explicit action receipts changed or second workspace launched")
	}
	after, e := os.ReadFile(inputFile)
	if e != nil || string(after) != string(before) {
		t.Fatal("EOF rejected input changed original delivery")
	}
	b = client.request("GET", "/api/v1/relay/tasks/"+id, "", "", 200)
	if json.Unmarshal(b, &task) != nil || task.View.Terminal.Liveness != "exited" || task.View.Task.State == "done" || task.View.Task.State == "released" {
		t.Fatal("owned EOF exit was confused with task completion")
	}
	client.request("GET", "/api/v1/relay/tasks/"+id+"/events", "", "", 200)
	encoded, e := json.Marshal(client.samples)
	if e != nil || os.WriteFile(filepath.Join(evidence, "http-eof-contract-samples.json"), encoded, 0600) != nil {
		t.Fatal("EOF public response evidence unavailable")
	}
}
