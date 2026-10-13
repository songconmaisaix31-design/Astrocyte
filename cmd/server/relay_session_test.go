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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// Public archive provider, not an installed model or a claim about real formats.
type relaySessionFixtureDriver struct {
	workspace.RelayDriver
	root            string
	session         domain.NativeSession
	archive         domain.RelaySessionArchive
	launches, reads int
	closed          bool
}

func publicSessionLine(v any) string { data, _ := json.Marshal(v); return string(data) + "\n" }
func publicSessionCall() string {
	return publicSessionLine(map[string]any{"type": "response_item", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"type": "function_call", "name": "Edit", "call_id": "PRIVATE_CANARY", "arguments": "PRIVATE_CANARY prompt/tool/stdout"}})
}
func (d *relaySessionFixtureDriver) Launch(_ context.Context, r workspace.RelayLaunch) (workspace.RelayObservation, error) {
	d.launches++
	d.session = domain.NativeSession{ID: uuid.NewString(), NativeID: "PRIVATE_CANARY native ID", ProjectID: r.Project.ID, CLI: r.Spec.Name, Ownership: "owned", Status: "running"}
	d.archive = domain.RelaySessionArchive{SessionID: d.session.ID, NativeID: d.session.NativeID, WorkspaceRoot: r.Workspace.Root, Root: d.root, Path: filepath.Join(d.root, "public-session.jsonl"), Format: "codex", AppendOnly: true}
	if os.MkdirAll(d.root, 0700) != nil || os.WriteFile(d.archive.Path, []byte(publicSessionLine(map[string]any{"type": "session_meta", "timestamp": "2026-10-14T01:00:00Z", "payload": map[string]any{"id": d.session.NativeID, "cwd": r.Workspace.Root}})+publicSessionCall()), 0600) != nil {
		return workspace.RelayObservation{}, errors.New("public fixture unavailable")
	}
	return workspace.RelayObservation{Session: d.session, Archive: &d.archive}, nil
}
func (d *relaySessionFixtureDriver) Read(context.Context, workspace.RelayLaunch) (workspace.RelayObservation, error) {
	d.reads++
	return workspace.RelayObservation{Session: d.session, Archive: &d.archive, ExitConfirmed: d.closed}, nil
}

func sessionFixtureAppend(t *testing.T, path, line string) {
	t.Helper()
	f, e := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal("public archive append unavailable")
	}
	_, e = f.WriteString(line)
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		t.Fatal("public append failed")
	}
}

func TestRelaySessionActualHTTPAdmissionRestartPermissionAndExitTail(t *testing.T) {
	ctx := context.Background()
	local := filepath.Join(relayTestLocalAppData(t), "astrocyte-workbench", "v0.2-access", "verification", "session-http", uuid.NewString())
	if os.MkdirAll(local, 0700) != nil {
		t.Fatal("isolated state unavailable")
	}
	t.Setenv("LOCALAPPDATA", local)
	root, err := relayRoot()
	if err != nil {
		t.Fatal("isolated root unavailable")
	}
	appDir := filepath.Join(root, "app-state")
	if os.MkdirAll(appDir, 0700) != nil {
		t.Fatal("app dir unavailable")
	}
	db, err := sqlite.Open(filepath.Join(appDir, "state.sqlite"))
	if err != nil {
		t.Fatal("isolated DB unavailable")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("isolated migrations unavailable")
	}
	p := domain.LocalProject{ID: "public-project", Root: filepath.Join(root, "source"), Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "public-session", AllowedActions: []string{"start", "observe", "stop"}}}
	if os.MkdirAll(p.Root, 0700) != nil || db.SaveProject(ctx, p, 0) != nil || db.SaveWorksSnapshot(ctx, domain.WorksSnapshot{Works: []domain.Work{{ID: "public-work", Root: p.Root, Kind: domain.WorkKindFolder}}}) != nil || db.SaveWorkMetadata(ctx, "public-work", domain.WorkHumanMetadata{Revision: 1, Handoff: domain.WorkHandoff{Content: "public controlled handoff", Source: "human"}}, 0) != nil {
		t.Fatal("public fixture unavailable")
	}
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		t.Fatal("isolated relay store unavailable")
	}
	defer store.Close()
	driver := &relaySessionFixtureDriver{root: filepath.Join(root, "public-archive-PRIVATE_CANARY")}
	runtime := workspace.NewRelayService(store, workspace.NewWorksService(db, nil, nil, nil, workspace.WorksServiceOptions{}), db, &relayEOFFixtureWorkspace{relayControlledWorkspace{root: filepath.Join(root, "worktrees")}}, driver, nil, nil, workspace.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "public-session", LaunchCmd: []string{"never-executed-public-fixture"}, DriveMode: "structured"}}, OwnedRoots: agents.RelayOwnedRoots{}})
	repo := sqlite.NewEventRepository(db)
	capture := events.NewRelaySessionCapture(attention.NewPendingEventService(repo), nil)
	state := relaySessionFileState{root: root}
	collector := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	wrapped := &relaySessionRuntime{RelayRuntime: runtime, RelayLifecycle: runtime.(workspace.RelayLifecycle), collector: collector}
	client := &relayPersistentHTTP{t: t, server: httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Relay: wrapped, RelaySessionEvents: collector}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())}
	defer client.server.Close()
	client.request("GET", "/api/v1/auth/session", "", "", 200)
	intent := client.request("POST", "/api/v1/relay/intents", `{"work_id":"public-work","project_id":"public-project","agent":"public-session","task":"public controlled task"}`, "", 200)
	var ready struct {
		View domain.RelayTaskView `json:"view"`
	}
	if json.Unmarshal(intent, &ready) != nil || driver.launches != 0 {
		t.Fatal("intent auto-launched")
	}
	taskID := ready.View.Task.ID
	client.request("POST", "/api/v1/relay/intents/"+taskID+"/confirm", `{"permission_confirmed":false,"permission_level":"A","project_settings_revision":1}`, "", 403)
	client.request("POST", "/api/v1/relay/intents/"+taskID+"/confirm", `{"permission_confirmed":true,"permission_level":"A","project_settings_revision":1}`, "", 200)
	var readSamples []relayContractSample
	query := func(task string) workbench.SessionEvents {
		t.Helper()
		path := "/api/v1/relay/session-events"
		if task != "" {
			path += "?task_id=" + task
		}
		data := client.request("GET", path, "", "", 200)
		readSamples = append(readSamples, relayContractSample{Schema: "RelaySessionEventsV1", Body: append(json.RawMessage(nil), data...)})
		var view workbench.SessionEvents
		if json.Unmarshal(data, &view) != nil {
			t.Fatal("closed read projection unavailable")
		}
		return view
	}
	before := query(taskID)
	if before.Total != 0 || driver.reads != 0 {
		t.Fatal("GET performed capture or native read")
	}
	if wrapped.ObserveOwned(ctx) != nil {
		t.Fatal("owned capture unavailable")
	}
	view := query(taskID)
	if view.Total != 2 || len(view.Coverage) != 1 || view.Coverage[0].Lines != 2 || view.Coverage[0].Gap != "prefix_unverified" || !view.Coverage[0].SnapshotComplete || view.Coverage[0].SessionComplete {
		t.Fatal("local admission/coverage distinction failed")
	}
	if view.Events[0].TS < view.Events[1].TS {
		t.Fatal("source timestamps not descending")
	}
	pagedBytes := client.request("GET", "/api/v1/relay/session-events?task_id="+taskID+"&page=2&page_size=1", "", "", 200)
	var paged workbench.SessionEvents
	if json.Unmarshal(pagedBytes, &paged) != nil || paged.Total != 2 || len(paged.Events) != 1 || paged.Events[0].ID != view.Events[1].ID {
		t.Fatal("actual source timestamp pagination failed")
	}
	readSamples = append(readSamples, relayContractSample{Schema: "RelaySessionEventsV1", Body: append(json.RawMessage(nil), pagedBytes...)})
	checkpointPath, _ := state.path()
	saved, e := os.ReadFile(checkpointPath)
	if e != nil || strings.Contains(string(saved), "PRIVATE_CANARY") {
		t.Fatal("private source leaked to checkpoint")
	}
	reads := driver.reads
	restarted := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	restored, e := restarted.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || restored.Total != 2 || restored.Events[0].ID != view.Events[0].ID || driver.reads != reads {
		t.Fatal("restart read lost stable admission")
	}
	line := publicSessionCall()
	sessionFixtureAppend(t, driver.archive.Path, line[:len(line)/2])
	if restarted.ObserveOwned(ctx) != nil {
		t.Fatal("partial append unavailable")
	}
	partial, e := restarted.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || partial.Total != 2 || partial.Coverage[0].TailBytes == 0 || partial.Coverage[0].SnapshotComplete {
		t.Fatal("unfinished tail published")
	}
	sessionFixtureAppend(t, driver.archive.Path, line[len(line)/2:])
	if restarted.ObserveOwned(ctx) != nil {
		t.Fatal("complete append unavailable")
	}
	driver.closed = true
	if wrapped.ObserveOwned(ctx) != nil {
		t.Fatal("controlled exit unavailable")
	}
	sessionFixtureAppend(t, driver.archive.Path, publicSessionCall()) // late archive flush after actual exit evidence
	if restarted.ObserveOwned(ctx) != nil {
		t.Fatal("closed owned tail not collected")
	}
	after, e := restarted.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || after.Total != 4 || after.Coverage[0].SessionComplete {
		t.Fatal("exit implied finalization or late tail lost")
	}
	rows, e := repo.ListEvents(ctx, 0, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).UnixMilli())
	if e != nil || len(rows) != 4 {
		t.Fatal("existing EventService not used or duplicate admission")
	}
	signals, e := repo.ListSignals(ctx, 0, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).UnixMilli())
	if e != nil || len(signals) != 0 {
		t.Fatal("session capture wrote cognitive signals")
	}
	for _, row := range rows {
		var closed map[string]any
		_ = json.Unmarshal(row.RawBytes, &closed)
		if len(closed) != 6 || row.Event != nil || strings.Contains(string(row.RawBytes), "PRIVATE_CANARY") {
			t.Fatal("six-field admission privacy failed")
		}
	}
	now := time.Now()
	if db.SaveGrant(ctx, domain.ProjectGrant{ProjectID: p.ID, AgentID: "public-session", Actions: []string{"observe"}, RevokedAt: &now}) != nil {
		t.Fatal("controlled revoke unavailable")
	}
	client.request("GET", "/api/v1/relay/session-events?task_id="+taskID, "", "", 403)
	hidden := query("")
	if hidden.Total != 0 || len(hidden.Coverage) != 0 {
		t.Fatal("revoked permission exposed cached admission")
	}
	if restarted.ObserveOwned(ctx) != nil {
		t.Fatal("revoked collector should skip source")
	}
	if driver.launches != 1 {
		t.Fatal("capture launched another task")
	}
	// A second public fixture grant restores the unchanged project revision;
	// then a real filesystem replace failure must not commit candidate offsets.
	if db.SaveGrant(ctx, domain.ProjectGrant{ProjectID: p.ID, AgentID: "public-session", Actions: []string{"observe"}}) != nil {
		t.Fatal("public grant restore unavailable")
	}
	priorBytes, e := os.ReadFile(checkpointPath)
	if e != nil {
		t.Fatal("checkpoint read unavailable")
	}
	writer := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	if os.Rename(checkpointPath, checkpointPath+".retained") != nil || os.Mkdir(checkpointPath, 0700) != nil {
		t.Fatal("public save failure fixture unavailable")
	}
	sessionFixtureAppend(t, driver.archive.Path, publicSessionCall())
	if writer.ObserveOwned(ctx) == nil || writer.blocked != "checkpoint_save_unknown" {
		t.Fatal("metadata failure did not block")
	}
	failedRead, e := writer.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || failedRead.Total != 4 || failedRead.Coverage[0].Gap != "checkpoint_save_unknown" {
		t.Fatal("failed metadata write advanced committed cursor")
	}
	broken := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	if broken.blocked != "checkpoint_unavailable" || broken.ObserveOwned(ctx) == nil {
		t.Fatal("bad checkpoint silently reset")
	}
	if os.Remove(checkpointPath) != nil || os.WriteFile(checkpointPath, priorBytes, 0600) != nil {
		t.Fatal("public checkpoint restore unavailable")
	}
	recovery := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	if recovery.ObserveOwned(ctx) != nil {
		t.Fatal("old checkpoint stable replay unavailable")
	}
	recovered, e := recovery.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || recovered.Total != 5 {
		t.Fatal("restart lost pending admitted record")
	}
	rows, e = repo.ListEvents(ctx, 0, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).UnixMilli())
	if e != nil || len(rows) != 5 {
		t.Fatal("checkpoint retry duplicated existing event identity")
	}
	if os.WriteFile(checkpointPath, []byte("PRIVATE_CANARY invalid JSON"), 0600) != nil {
		t.Fatal("public corruption unavailable")
	}
	corrupt := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), capture, state)
	if corrupt.blocked != "checkpoint_unavailable" || corrupt.ObserveOwned(ctx) == nil {
		t.Fatal("corrupt cursor silently reset")
	}
	var semantic relaySessionCheckpoints
	if json.Unmarshal(priorBytes, &semantic) != nil {
		t.Fatal("public checkpoint decode failed")
	}
	badVersion := semantic[taskID]
	badVersion.Cursor.Version = 99
	semantic[taskID] = badVersion
	badJSON, _ := json.Marshal(semantic)
	if os.WriteFile(checkpointPath, badJSON, 0600) != nil {
		t.Fatal("public semantic corruption unavailable")
	}
	if _, e = state.Load(); e == nil {
		t.Fatal("unknown checkpoint version accepted before capture")
	}
	wrong := relaySessionFileState{root: filepath.Join(root, "snapshots")}
	if _, e = wrong.Load(); e == nil {
		t.Fatal("cursor accepted another storage domain")
	}
	sampleBytes, e := json.Marshal(readSamples)
	if e != nil || os.WriteFile(filepath.Join(local, "session-http-derived.json"), sampleBytes, 0600) != nil {
		t.Fatal("public derived HTTP samples unavailable")
	}
}
