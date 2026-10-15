package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayTrustedRegistrationActualSQLiteHTTP(t *testing.T) {
	ctx := context.Background()
	parent, err := filepath.Abs(filepath.Join("..", "..", "dist", "relay-registration"))
	if err != nil || os.MkdirAll(parent, 0700) != nil {
		t.Fatal("登记实际检查目录不可用")
	}
	local, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("登记独占检查目录不可用")
	}
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "")
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", "")
	t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "archive")
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "")
	root, err := relayRoot()
	if err != nil || os.MkdirAll(root, 0700) != nil {
		t.Fatal("实际可信登记根不可用")
	}
	registrationPath := filepath.Join(root, "agents.json")
	if os.WriteFile(registrationPath, []byte("[]"), 0600) != nil {
		t.Fatal("实际空登记写入失败")
	}
	db, err := sqlite.Open(filepath.Join(local, "application.sqlite"))
	if err != nil {
		t.Fatal("实际应用数据库打开失败")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("实际应用数据库迁移失败")
	}
	registry := agents.NewRegistry(filepath.Join(local, "native-sessions"))
	services := httpapi.Services{}
	cleanup, err := assembleRelay(&services, db, registry)
	if err != nil {
		t.Fatal("原生产Relay装配失败")
	}
	defer func() {
		if cleanup() != nil {
			t.Error("原生产Relay关闭失败")
		}
	}()
	runtime, ok := services.Relay.(*relaySessionRuntime)
	if !ok || runtime.registration == nil {
		t.Fatal("原运行时缺少可信登记重载")
	}
	originalRuntime := runtime.RelayRuntime
	listener, err := bindHTTPListener(0)
	if err != nil {
		t.Fatal("实际HTTP绑定失败")
	}
	server := httpapi.NewServer(httpapi.Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: services})
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	client := newActualRelayHTTP(t, "http://"+listener.Addr().String())
	defer func() {
		client.client.CloseIdleConnections()
		shutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if server.Shutdown(shutdown) != nil || !errors.Is(<-serveDone, http.ErrServerClosed) {
			t.Error("实际HTTP正常关闭失败")
		}
	}()
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	type agentList struct {
		Agents []struct {
			Name string `json:"name"`
		} `json:"agents"`
	}
	var empty agentList
	client.request(t, "GET", "/api/v1/relay/agents", nil, &empty)
	if len(empty.Agents) != 0 {
		t.Fatal("同一服务空登记基线不为空")
	}
	specs := []domain.AgentSpec{{Name: "codex", LaunchCmd: []string{"codex"}, DriveMode: "structured"}}
	encoded, err := json.Marshal(specs)
	if err != nil || os.WriteFile(registrationPath, encoded, 0600) != nil {
		t.Fatal("本机Codex可信登记写入失败")
	}
	// GET只读取旧完整快照，原observer才执行本机可信重载。
	client.request(t, "GET", "/api/v1/relay/agents", nil, &empty)
	if len(empty.Agents) != 0 {
		t.Fatal("HTTP读取触发了可信登记重载")
	}
	if runtime.ObserveOwned(ctx) != nil {
		t.Fatal("原observer同实例登记激活失败")
	}
	checkActive := func() {
		t.Helper()
		var active agentList
		client.request(t, "GET", "/api/v1/relay/agents", nil, &active)
		if len(active.Agents) != 1 || active.Agents[0].Name != "codex" || services.Relay != runtime || runtime.RelayRuntime != originalRuntime {
			t.Fatal("可信登记没有保留同一服务实例与完整快照")
		}
	}
	checkActive()
	var group sync.WaitGroup
	failures := make(chan error, 8)
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := runtime.ReloadTrustedAgents(ctx); err != nil {
				failures <- err
				return
			}
			response, err := client.client.Get(client.baseURL + "/api/v1/relay/agents")
			if err != nil {
				failures <- err
				return
			}
			var active agentList
			decodeErr := json.NewDecoder(response.Body).Decode(&active)
			closeErr := response.Body.Close()
			if response.StatusCode != http.StatusOK || decodeErr != nil || closeErr != nil || len(active.Agents) != 1 || active.Agents[0].Name != "codex" {
				failures <- errors.New("concurrent registration read differs")
			}
		}()
	}
	group.Wait()
	close(failures)
	for range failures {
		t.Fatal("并发同实例重载或实际HTTP读取失败")
	}
	if os.WriteFile(registrationPath, append(encoded, []byte(" {}")...), 0600) != nil || runtime.ReloadTrustedAgents(ctx) == nil {
		t.Fatal("错误登记没有立即拒绝")
	}
	checkActive()
	if os.WriteFile(registrationPath, encoded, 0600) != nil {
		t.Fatal("恢复准确可信登记失败")
	}
	t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "7")
	if runtime.ReloadTrustedAgents(ctx) == nil {
		t.Fatal("登记重载擅自改变启动活性策略")
	}
	t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "")
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if !errors.Is(runtime.ReloadTrustedAgents(cancelled), context.Canceled) {
		t.Fatal("取消的登记重载未立即停止")
	}
	checkActive()
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		t.Fatal("实际Relay库核查失败")
	}
	defer store.Close()
	state, err := store.Snapshot(ctx)
	if err != nil || len(state.Tasks) != 0 || len(state.Sessions) != 0 || len(state.Terminals) != 0 || len(state.Workspaces) != 0 || len(state.Receipts) != 0 {
		t.Fatal("可信登记产生了原生任务或输入状态")
	}
	writeActualJSON(t, filepath.Join(local, "safe-result.json"), map[string]any{"same_service": true, "get_readonly": true, "concurrent_reloads": 8, "invalid_registration_rejected": true, "startup_policy_preserved": true, "native_sessions": 0, "model_inputs": 0})
}
