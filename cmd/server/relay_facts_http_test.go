package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayFactsActualHTTPSQLiteReadOnlyAndPermission(t *testing.T) {
	ctx := context.Background()
	parent, err := filepath.Abs(filepath.Join("..", "..", "dist", "relay-facts"))
	if err != nil || os.MkdirAll(parent, 0700) != nil {
		t.Fatal("独占验证目录不可用")
	}
	root, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("独占验证目录创建失败")
	}
	source := filepath.Join(root, "public-git")
	if os.MkdirAll(source, 0700) != nil || os.WriteFile(filepath.Join(source, "README.md"), []byte("# Public Relay facts\n\nRead persisted action facts without starting a controller.\n"), 0600) != nil {
		t.Fatal("公开项目创建失败")
	}
	for _, args := range [][]string{{"init", source}, {"-C", source, "add", "README.md"}, {"-C", source, "-c", "user.name=Relay public facts", "-c", "user.email=relay-facts@example.invalid", "-c", "core.hooksPath=", "commit", "-m", "Document public Relay facts"}} {
		command := exec.Command("git", args...)
		command.Stdout, command.Stderr = io.Discard, io.Discard
		if command.Run() != nil {
			t.Fatal("公开 Git 项目提交失败")
		}
	}
	db, err := sqlite.Open(filepath.Join(root, "application.sqlite"))
	if err != nil {
		t.Fatal("真实 SQLite 打开失败")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("真实 SQLite 迁移失败")
	}
	t.Setenv("LOCALAPPDATA", root)
	stateRoot, err := sqlite.RelayStateRoot()
	if err != nil {
		t.Fatal("独占 Relay 根不可用")
	}
	store, err := sqlite.OpenRelayStore(stateRoot)
	if err != nil {
		t.Fatal("真实 Relay SQLite 打开失败")
	}
	defer store.Close()
	registry := agents.NewRegistry(filepath.Join(root, "native-sessions"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	caller := func() domain.Caller {
		return domain.Caller{Kind: "human", ID: "local-human", OperationID: uuid.NewString()}
	}
	project, err := projects.RegisterProject(ctx, caller(), domain.RegisterProjectCommand{Name: "Public facts documentation", Root: source, SpaceID: "public"})
	if err != nil {
		t.Fatalf("公开项目注册失败: %v", err)
	}
	settings := domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: []string{"observe"}}
	project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal("当前项目权限配置失败")
	}
	works := workspace.NewWorksService(db, nil, worksscan.Delivery{}, nil, workspace.WorksServiceOptions{})
	delivery, err := json.Marshal(map[string]any{"works": []any{map[string]any{"platform": "codex", "name": "Public facts documentation", "path": source, "kind": "folder", "handoff": "Read public facts documentation"}}})
	if err != nil {
		t.Fatal("公开任务编码失败")
	}
	album, err := works.ImportWorksDelivery(ctx, caller(), workspace.WorksDeliveryCommand{Content: delivery})
	if err != nil || len(album.Works) != 1 {
		t.Fatal("公开任务导入失败")
	}
	runtime := workspace.NewRelayService(store, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), nil, nil, workspace.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}})
	lifecycle, ok := runtime.(workspace.RelayLifecycle)
	if !ok {
		t.Fatal("生产生命周期接口缺失")
	}
	wrapped := &relaySessionRuntime{RelayRuntime: runtime, RelayLifecycle: lifecycle}
	intent, err := runtime.CreateIntent(ctx, caller(), workspace.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "codex", Task: "Read public facts documentation"})
	if err != nil {
		t.Fatal("真实公开意图创建失败")
	}
	listener, err := bindHTTPListener(0)
	if err != nil {
		t.Fatal("真实 Go HTTP 绑定失败")
	}
	server := httpapi.NewServer(httpapi.Config{Port: listener.Addr().(*net.TCPAddr).Port, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: httpapi.Services{Relay: wrapped}})
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	client := newActualRelayHTTP(t, "http://"+listener.Addr().String())
	defer func() {
		client.client.CloseIdleConnections()
		shutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if server.Shutdown(shutdown) != nil || !errors.Is(<-serveDone, http.ErrServerClosed) {
			t.Error("真实 HTTP 正常关闭失败")
		}
	}()
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	// 未派发的真实 ready 任务拒绝完成，生成原业务拒绝收据和结果事实。
	_, err = runtime.Transition(ctx, caller(), intent.Task.ID, workspace.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true})
	if err == nil {
		t.Fatal("未派发任务被错误完成")
	}
	before, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal("真实事实快照读取失败")
	}
	var response struct {
		SchemaVersion int                           `json:"schema_version"`
		TaskID        string                        `json:"task_id"`
		Facts         []domain.RelayRuntimeFact     `json:"facts"`
		Relations     []domain.RelaySessionRelation `json:"session_relations"`
		Ownership     domain.RelayCurrentOwnership  `json:"current_ownership"`
	}
	path := "/api/v1/relay/tasks/" + intent.Task.ID + "/facts"
	var public map[string]json.RawMessage
	client.request(t, "GET", path, nil, &public)
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatal("实际安全响应编码失败")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&response) != nil || decoder.Decode(new(any)) != io.EOF {
		t.Fatal("事实响应包含契约之外的字段")
	}
	expected := []domain.RelayRuntimeFact{}
	for _, fact := range before.RuntimeFacts {
		if fact.TaskID == intent.Task.ID {
			expected = append(expected, fact)
		}
	}
	if response.SchemaVersion != 1 || response.TaskID != intent.Task.ID || response.Ownership.Status != "unknown" || response.Ownership.Reason != "controller_binding_unrecorded" || len(response.Facts) != len(expected) || len(response.Relations) != 0 {
		t.Fatal("安全事实响应未对应真实 store")
	}
	rejected := false
	for _, fact := range response.Facts {
		if fact.Kind == "action_result" && fact.Status == "rejected" && fact.Reason == "version_conflict" {
			rejected = true
		}
	}
	if !rejected {
		t.Fatal("原拒绝收据对应的实际结果事实缺失")
	}
	for _, fact := range response.Facts {
		matched := false
		for _, stored := range before.RuntimeFacts {
			if reflect.DeepEqual(fact, stored) {
				matched = true
			}
		}
		if !matched {
			t.Fatal("HTTP fact_id 没有真实存储对应")
		}
	}
	after, err := store.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("GET 修改了真实 Relay store")
	}
	requestStatus := func(origin string, expected int) {
		request, err := http.NewRequest("GET", client.baseURL+path, nil)
		if err != nil {
			t.Fatal("真实请求创建失败")
		}
		request.Header.Set("Origin", origin)
		result, err := client.client.Do(request)
		if err != nil {
			t.Fatal("真实请求失败，没有重试")
		}
		defer result.Body.Close()
		if result.StatusCode != expected {
			t.Fatalf("HTTP status=%d，预期%d", result.StatusCode, expected)
		}
	}
	requestStatus("http://127.0.0.1:0", 403)
	if _, err := projects.GrantProjectAgent(ctx, caller(), project.ID, domain.ProjectGrant{AgentID: "codex", Actions: []string{"start"}}); err != nil {
		t.Fatal("真实权限变更失败")
	}
	requestStatus(client.baseURL, 403)
	final, err := store.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(before, final) || len(final.Sessions) != 0 || len(final.Terminals) != 0 || len(final.Workspaces) != 0 {
		t.Fatal("读取或拒绝产生原生进程状态或 store 写入")
	}
	prepareBody, err := json.Marshal(workspace.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: "A", ProjectSettingsRevision: project.Settings.Revision})
	if err != nil {
		t.Fatal("实际初始化命令编码失败")
	}
	prepareRequest, err := http.NewRequest("POST", client.baseURL+"/api/v1/relay/intents/"+intent.Task.ID+"/prepare", bytes.NewReader(prepareBody))
	if err != nil {
		t.Fatal("实际初始化请求创建失败")
	}
	prepareRequest.Header.Set("Origin", client.baseURL)
	prepareRequest.Header.Set("Content-Type", "application/json")
	prepareRequest.Header.Set("X-CSRF-Token", client.csrf)
	prepareRequest.Header.Set("Idempotency-Key", uuid.NewString())
	prepareResponse, err := client.client.Do(prepareRequest)
	if err != nil {
		t.Fatal("实际初始化请求失败，没有重试")
	}
	prepareResponse.Body.Close()
	if prepareResponse.StatusCode != http.StatusForbidden {
		t.Fatal("原Prepare包装没有重核当前start权限")
	}
	prepared, err := store.Snapshot(ctx)
	if err != nil || prepared.Tasks[intent.Task.ID].State != "ready" || len(prepared.Sessions) != 0 || len(prepared.Terminals) != 0 || len(prepared.Workspaces) != 0 {
		t.Fatal("被拒绝的初始化改变了原生会话或任务状态")
	}
	writeActualJSON(t, filepath.Join(root, "safe-result.json"), map[string]any{"facts": len(response.Facts), "fact_ids_located": true, "get_readonly": true, "origin_rejected": true, "permission_rechecked": true, "native_sessions": 0})
}
