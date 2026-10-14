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
	"net/http/cookiejar"
	"os"
	"path/filepath"
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
	server := httpapi.NewServer(httpapi.Config{Port: address.Port, AllowedOrigins: []string{webOrigin}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: httpapi.Services{Foundation: foundation.NewAttentionService(schemaVersion), LocalProjects: projects}})
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
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("实际客户端 session 初始化失败")
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
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
	request("GET", "/api/v1/auth/session", "http://127.0.0.1:0", "", nil, 403, nil)
	var bootstrap struct {
		CSRF string `json:"csrf_token"`
	}
	request("GET", "/api/v1/auth/session", apiURL, "", nil, 200, &bootstrap)
	if bootstrap.CSRF == "" {
		t.Fatal("实际 bootstrap 缺少 CSRF")
	}
	request("GET", "/api/v1/local-projects", webOrigin, "", nil, 200, nil)
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
	readonly, err := sql.Open("sqlite", "file:"+filepath.ToSlash(databasePath)+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		t.Fatal("独立只读 SQL 连接失败")
	}
	defer readonly.Close()
	var nativeSessions int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM local_agent_sessions").Scan(&nativeSessions); err != nil || nativeSessions != 0 {
		t.Fatal("本轮原生会话零调用状态无法核实")
	}
	if listener.Addr().String() != address.String() || server.Addr() != address.String() {
		t.Fatal("HTTP 服务没有保持原始 listener 地址")
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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
