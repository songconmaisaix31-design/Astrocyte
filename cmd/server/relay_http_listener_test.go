package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayHTTPAtomicListenerActualSQLiteSessionAndShutdown(t *testing.T) {
	ctx := context.Background()
	parent, err := filepath.Abs(filepath.Join("..", "..", "dist", "http-listener"))
	if err != nil {
		t.Fatal("本轮验证目录无法解析")
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal("本轮验证目录不可用")
	}
	root, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("本轮独占目录不可用")
	}
	databasePath := filepath.Join(root, "state.sqlite")
	db, err := sqlite.Open(databasePath)
	if err != nil {
		t.Fatal("实际 SQLite 打开失败")
	}
	defer db.Close()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal("实际 SQLite 迁移失败")
	}
	schemaVersion, err := db.SchemaVersion(ctx)
	if err != nil {
		t.Fatal("实际 schema 版本不可用")
	}
	projectRoot := filepath.Join(root, "public-project")
	if err := os.MkdirAll(projectRoot, 0700); err != nil {
		t.Fatal("本次公开项目目录不可用")
	}
	registry := agents.NewRegistry(filepath.Join(root, "native-sessions"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	t.Setenv("ASTROCYTE_PORT", "0")
	configuredPort, err := resolvePort()
	if err != nil || configuredPort != 0 {
		t.Fatal("实际 main 端口配置没有接受显式 0")
	}
	listener, err := bindHTTPListener(configuredPort)
	if err != nil {
		t.Fatal("Go 实际端口 0 绑定失败")
	}
	defer listener.Close()
	address := listener.Addr().(*net.TCPAddr)
	if !address.IP.IsLoopback() || address.Port == 0 {
		t.Fatal("端口 0 没有获得实际 loopback 地址")
	}
	var announcement bytes.Buffer
	if err := announceHTTPListener(&announcement, listener); err != nil {
		t.Fatal("实际监听地址公告失败")
	}
	var announced map[string]string
	decoder := json.NewDecoder(&announcement)
	if err := decoder.Decode(&announced); err != nil || decoder.Decode(new(any)) != io.EOF || len(announced) != 2 || announced["type"] != "http_listening" || announced["api_url"] != "http://"+address.String() {
		t.Fatal("实际地址公告包含错误或额外信息")
	}
	if competing, err := bindHTTPListener(address.Port); err == nil {
		competing.Close()
		t.Fatal("已持有监听端口允许第二次绑定")
	}
	apiURL := announced["api_url"]
	webOrigin := "http://127.0.0.1:17345"
	services := httpapi.Services{Foundation: foundation.NewAttentionService(schemaVersion), LocalProjects: projects, WorkspaceS2: workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil), RepositoryContext: &httpapi.RepositoryContextServices{}}
	if err := assembleSwarmLedger(&services, db); err != nil {
		t.Fatal("实际同数据库批准桥与Swarm账本装配失败")
	}
	server := httpapi.NewServer(httpapi.Config{Port: address.Port, AllowedOrigins: []string{webOrigin}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: services})
	if err := server.Serve(nil); err == nil {
		t.Fatal("缺少 listener 没有立即失败")
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	closed := false
	defer func() {
		if closed {
			return
		}
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			t.Error("实际 HTTP 关闭失败")
		}
		select {
		case err := <-serveDone:
			if !errors.Is(err, http.ErrServerClosed) {
				t.Error("实际 Serve 未正常终止")
			}
		case <-shutdownCtx.Done():
			t.Error("实际 Serve 关闭未完成")
		}
	}()
	writeActualJSON(t, filepath.Join(root, "service-public.json"), map[string]any{"api_url": apiURL})
	client := newActualRelayHTTP(t, actualRelayAPIURL(t, root)).client
	defer client.CloseIdleConnections()
	request := func(method, path, origin, csrf string, body any, status int, target any) {
		t.Helper()
		var encoded bytes.Buffer
		if body != nil {
			if err := json.NewEncoder(&encoded).Encode(body); err != nil {
				t.Fatal("实际请求编码失败")
			}
		}
		req, err := http.NewRequest(method, apiURL+path, &encoded)
		if err != nil {
			t.Fatal("实际请求创建失败")
		}
		req.Header.Set("Origin", origin)
		req.Header.Set("X-CSRF-Token", csrf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", uuid.NewString())
		response, err := client.Do(req)
		if err != nil {
			t.Fatal("实际 HTTP 请求失败，没有重试")
		}
		defer response.Body.Close()
		if response.StatusCode != status {
			t.Fatalf("实际 HTTP 状态 %d，预期 %d", response.StatusCode, status)
		}
		if target != nil {
			if err := json.NewDecoder(response.Body).Decode(target); err != nil {
				t.Fatal("实际 HTTP 响应解码失败")
			}
		}
	}
	request("GET", "/api/v1/health", apiURL, "", nil, 200, nil)
	request("GET", "/api/v1/local-projects", apiURL, "", nil, 403, nil)
	request("GET", "/api/v1/missions", apiURL, "", nil, 403, nil)
	request("GET", "/api/v1/auth/session", "http://127.0.0.1:0", "", nil, 403, nil)
	var bootstrap struct {
		CSRF string `json:"csrf_token"`
	}
	request("GET", "/api/v1/auth/session", apiURL, "", nil, 200, &bootstrap)
	if bootstrap.CSRF == "" {
		t.Fatal("实际 bootstrap 缺少 CSRF")
	}
	request("GET", "/api/v1/local-projects", webOrigin, "", nil, 200, nil)
	var missions struct {
		SchemaVersion int               `json:"schema_version"`
		Items         []json.RawMessage `json:"items"`
	}
	request("GET", "/api/v1/missions", webOrigin, "", nil, 200, &missions)
	if missions.SchemaVersion != 1 || missions.Items == nil || len(missions.Items) != 0 {
		t.Fatal("实际空Swarm账本响应不符合公开集合契约")
	}
	missingGrant := map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 0, "grant_id": uuid.NewString(), "epoch": 1, "proposal_revision": 1, "project_id": uuid.NewString(), "root": projectRoot}
	request("POST", "/api/v1/missions", webOrigin, "", missingGrant, 403, nil)
	request("POST", "/api/v1/missions", webOrigin, bootstrap.CSRF, missingGrant, 404, nil)
	command := map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 1, "name": "Atomic HTTP public project", "root": projectRoot, "space_id": "http-listener-public"}
	request("POST", "/api/v1/local-projects", apiURL, "", command, 403, nil)
	request("POST", "/api/v1/local-projects", "http://127.0.0.1:0", bootstrap.CSRF, command, 403, nil)
	before, err := db.ListProjects(ctx)
	if err != nil || len(before) != 0 {
		t.Fatal("Origin 或 CSRF 拒绝请求产生了 SQLite 写入")
	}
	var registered struct {
		Project domain.LocalProject `json:"project"`
	}
	request("POST", "/api/v1/local-projects", webOrigin, bootstrap.CSRF, command, 200, &registered)
	actual, err := db.LoadProject(ctx, registered.Project.ID)
	if err != nil || actual.ID == "" || actual.Root != projectRoot {
		t.Fatal("实际 HTTP 写入没有对应真实 SQLite 项目")
	}
	settingsCommand := map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 1, "expected_revision": actual.Settings.Revision, "settings": actual.Settings}
	request("PUT", "/api/v1/local-projects/"+actual.ID+"/settings", apiURL, bootstrap.CSRF, settingsCommand, 200, &registered)
	updated, err := db.LoadProject(ctx, actual.ID)
	if err != nil || updated.Settings.Revision != actual.Settings.Revision+1 {
		t.Fatal("实际 API Origin 和 CSRF 写入没有持久化新版本")
	}
	repositoryPath := "/api/v1/local-projects/" + actual.ID + "/repository-context"
	mapQuery := "/map?purpose=inspect-public-project&root=.&max_files=10&max_bytes=4096&max_results=10&max_depth=1"
	var unavailable map[string]json.RawMessage
	request("GET", repositoryPath+mapQuery, webOrigin, "", nil, 503, &unavailable)
	safeError, err := json.Marshal(unavailable)
	if err != nil || bytes.Contains(safeError, []byte(projectRoot)) || bytes.Contains(safeError, []byte(databasePath)) {
		t.Fatal("缺配置响应包含本机原始路径")
	}
	request("GET", repositoryPath+mapQuery+"&max_files=5", webOrigin, "", nil, 400, nil)
	request("GET", repositoryPath+mapQuery+"&unknown=value", webOrigin, "", nil, 400, nil)
	request("GET", repositoryPath+mapQuery+"&%GG=value", webOrigin, "", nil, 400, nil)
	query := map[string]any{"schema_version": 1, "purpose": "inspect-public-project", "roots": []string{"."}, "budget": map[string]int{"max_files": 10, "max_bytes": 4096, "max_results": 10, "max_depth": 1}}
	request("POST", repositoryPath+"/search", webOrigin, "", query, 403, nil)
	request("POST", repositoryPath+"/search", webOrigin, bootstrap.CSRF, query, 503, nil)
	request("POST", repositoryPath+"/impact", webOrigin, bootstrap.CSRF, query, 503, nil)
	request("POST", repositoryPath+"/map/refresh", webOrigin, bootstrap.CSRF, query, 503, nil)
	query["action"] = "index"
	request("POST", repositoryPath+"/index", webOrigin, bootstrap.CSRF, query, 503, nil)
	afterRepositoryReads, err := db.LoadProject(ctx, actual.ID)
	if err != nil || !reflect.DeepEqual(updated, afterRepositoryReads) {
		t.Fatal("缺配置的仓库上下文请求修改了真实项目")
	}
	readonly, err := sql.Open("sqlite", "file:"+filepath.ToSlash(databasePath)+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		t.Fatal("独立只读 SQL 连接失败")
	}
	defer readonly.Close()
	var nativeSessions int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM local_agent_sessions").Scan(&nativeSessions); err != nil || nativeSessions != 0 {
		t.Fatal("本轮原生会话零调用状态无法核实")
	}
	var missionCount int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM swarm_missions").Scan(&missionCount); err != nil || missionCount != 0 {
		t.Fatal("未批准的请求创建了实际Mission")
	}
	if listener.Addr().String() != address.String() || server.Addr() != address.String() {
		t.Fatal("HTTP 服务没有保持原始 listener 地址")
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client.CloseIdleConnections()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatal("实际 HTTP 正常关闭失败")
	}
	if err := <-serveDone; !errors.Is(err, http.ErrServerClosed) {
		t.Fatal("实际 Serve 未返回正常关闭结果")
	}
	closed = true
	if connection, err := net.DialTimeout("tcp", address.String(), time.Second); err == nil {
		connection.Close()
		t.Fatal("正常关闭后仍可连接原监听地址")
	}
	writeActualJSON(t, filepath.Join(root, "safe-result.json"), map[string]any{"api_url": apiURL, "project_id": actual.ID, "atomic_listener": true, "duplicate_bind_rejected": true, "actual_sqlite_write": true, "native_sessions": nativeSessions, "origin_csrf": "PASS", "normal_shutdown": true})
}
