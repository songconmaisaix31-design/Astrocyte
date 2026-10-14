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
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 此测试启动真实服务；派发由操作人员通过既有 HTTP 接口明确执行。
func TestRelaySessionActualService(t *testing.T) {
	selected := os.Getenv("ASTROCYTE_L1_ACTUAL_ROOT")
	if selected == "" {
		t.Skip("NOT_RUN: explicit actual-service directory required")
	}
	parent := filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification")
	canonical, err := workbenchCanonical(selected)
	if err != nil || !filepath.IsAbs(selected) || !relayInside(parent, canonical) || canonical == parent {
		t.Fatal("实际服务目录不在指定本机范围")
	}
	if err := os.MkdirAll(canonical, 0700); err != nil {
		t.Fatal("实际服务目录不可用")
	}
	evidence := os.Getenv("ASTROCYTE_L1_EVIDENCE_DIR")
	expectedEvidence, err := filepath.Abs(filepath.Join("..", "..", "dist", "l1-live"))
	if err != nil || evidence != expectedEvidence {
		t.Fatal("实际验证中间目录必须使用当前 H 工作目录")
	}
	source := filepath.Join(canonical, "public-project")
	if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(source, 0700); err != nil {
			t.Fatal("公开项目目录不可用")
		}
		if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("# L1 public project\nThis project validates native session capture with fixed public context and disabled tools.\n"), 0600); err != nil {
			t.Fatal("公开项目文件不可用")
		}
		for _, args := range [][]string{{"init", "--initial-branch=main"}, {"add", "README.md"}, {"-c", "user.name=L1 Verification", "-c", "user.email=l1@example.invalid", "commit", "-m", "Create public session verification project"}} {
			cmd := exec.Command("git", append([]string{"-C", source}, args...)...)
			cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
			if err := cmd.Run(); err != nil {
				t.Fatal("公开 Git 项目初始化失败")
			}
		}
	} else if err != nil {
		t.Fatal("公开项目身份不可核实")
	}
	t.Setenv("LOCALAPPDATA", filepath.Join(canonical, "local"))
	root, err := relayRoot()
	if err != nil {
		t.Fatal("独立接入目录不可用")
	}
	appDir := filepath.Join(root, "app-state")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		t.Fatal("独立应用目录不可用")
	}
	db, err := sqlite.Open(filepath.Join(appDir, "state.sqlite"))
	if err != nil {
		t.Fatal("独立应用数据库不可用")
	}
	defer db.Close()
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal("既有数据库初始化失败")
	}
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		t.Fatal("独立运行时数据库不可用")
	}
	defer store.Close()
	registry := agents.NewRegistry(filepath.Join(appDir, "native-sessions"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	projects.ConfigureRegisteredDiscovery(db, nil)
	repo := sqlite.NewEventRepository(db)
	eventService := attention.NewPendingEventService(repo)
	sink := events.NewRelayEventSink(eventService, nil)
	works := workspace.NewWorksService(db, worksscan.NewScanner(), nil, events.NewWorksEventSink(eventService, nil), workspace.WorksServiceOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	snapshot, err := works.RefreshWorks(ctx, domain.Caller{Kind: "human", ID: "local-human", OperationID: uuid.NewString()}, workspace.WorksRefreshCommand{Roots: []string{source}})
	if err != nil || len(snapshot.Works) != 1 || len(snapshot.Failures) != 0 {
		t.Fatal("真实公开项目扫描失败")
	}
	_, specs, window, err := relayConfiguration()
	if err != nil {
		t.Fatal("实际登记配置不可用")
	}
	if window != 0 {
		t.Fatal("本轮活性窗口保持未配置")
	}
	runtime := workspace.NewRelayService(store, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{}), sink, workspace.RelayServiceOptions{Agents: specs, OwnedRoots: agents.RelayOwnedRoots{}})
	collector := newRelaySessionCollector(runtime.(workspace.RelayOwnedSessionReader), events.NewRelaySessionCapture(eventService, nil), relaySessionFileState{root: root}, sqlite.NewAccessSessionReader(db))
	if collector.blocked != "" {
		t.Fatal("实际采集游标不可用")
	}
	wrapped := &relaySessionRuntime{RelayRuntime: runtime, RelayLifecycle: runtime.(workspace.RelayLifecycle), collector: collector}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httpapi.NewServer(httpapi.Config{Port: 17342, AllowedOrigins: []string{"http://127.0.0.1:17343"}, Logger: logger, Services: httpapi.Services{Relay: wrapped, RelaySessionEvents: collector, Works: works, LocalProjects: projects, RegisteredProjects: projects, Workbench: sqlite.NewWorkbenchReader(filepath.Join(canonical, "snapshot-unconfigured"))}})
	listener, err := net.Listen("tcp", "127.0.0.1:17342")
	if err != nil {
		t.Fatal("实际服务端口不可用")
	}
	transport := &http.Server{Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second}
	serveDone := make(chan error, 1)
	go func() { serveDone <- transport.Serve(listener) }()
	stopObserver := startRelayObserver(ctx, wrapped, logger)
	defer func() {
		cancel()
		stopObserver()
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer closeCancel()
		if err := transport.Shutdown(closeCtx); err != nil {
			t.Error("实际 HTTP 服务关闭失败")
		}
		if err := wrapped.Shutdown(closeCtx); err != nil {
			t.Error("实际原生生命周期关闭失败")
		}
		if err := <-serveDone; err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Error("实际 HTTP 服务异常终止")
		}
	}()
	client := newActualRelayHTTP(t)
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	var project domain.LocalProject
	list, err := projects.ListProjects(ctx, domain.Caller{Kind: "human", ID: "local-human"})
	if err != nil {
		t.Fatal("实际项目查询失败")
	}
	if len(list) == 0 {
		var response struct {
			Project domain.LocalProject `json:"project"`
		}
		client.request(t, "POST", "/api/v1/local-projects", map[string]any{"name": "L1 public project", "root": source, "space_id": "l1-public-context"}, &response)
		project = response.Project
	} else if len(list) == 1 && list[0].Root == source {
		project = list[0]
	} else {
		t.Fatal("实际目录含非本轮持有项目")
	}
	if project.Settings.ExternalModelCLI == "" {
		var response struct {
			Project domain.LocalProject `json:"project"`
		}
		settings := project.Settings
		settings.ExternalModelCLI = "codex"
		settings.AllowedActions = []string{"start", "observe", "send", "stop", "resume"}
		client.request(t, "PUT", "/api/v1/local-projects/"+project.ID+"/settings", map[string]any{"expected_revision": project.Settings.Revision, "settings": settings}, &response)
		project = response.Project
	}
	public := map[string]any{"api_url": "http://127.0.0.1:17342", "project_id": project.ID, "work_id": snapshot.Works[0].ID, "project_revision": project.Settings.Revision, "agent_name": "codex", "service_started_at": time.Now().UTC()}
	writeActualJSON(t, filepath.Join(evidence, "service-public.json"), public)
	writeActualJSON(t, filepath.Join(evidence, "service-private.json"), map[string]any{"app_root": root, "source_root": source, "db_path": filepath.Join(appDir, "state.sqlite")})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-serveDone:
			serveDone <- err
			t.Fatal("实际 HTTP 服务提前终止")
		case <-ticker.C:
			_, err := os.Stat(filepath.Join(evidence, "shutdown"))
			if err == nil {
				return
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatal("实际服务关闭请求读取失败")
			}
		}
	}
}

type actualRelayHTTP struct {
	client *http.Client
	csrf   string
}

func newActualRelayHTTP(t *testing.T) *actualRelayHTTP {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("实际会话 Cookie 不可用")
	}
	return &actualRelayHTTP{client: &http.Client{Jar: jar, Timeout: 30 * time.Second}}
}

func (c *actualRelayHTTP) request(t *testing.T, method, path string, body map[string]any, result any) {
	t.Helper()
	var encoded bytes.Buffer
	operation := uuid.NewString()
	if body != nil {
		if path != "/api/v1/relay/intents" && !bytes.Contains([]byte(path), []byte("/relay/")) {
			body["schema_version"], body["request_id"], body["expected_version"] = 1, operation, 1
		}
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			t.Fatal("实际请求编码失败")
		}
	}
	req, err := http.NewRequest(method, "http://127.0.0.1:17342"+path, &encoded)
	if err != nil {
		t.Fatal("实际请求不可用")
	}
	req.Header.Set("Origin", "http://127.0.0.1:17342")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", c.csrf)
		req.Header.Set("Idempotency-Key", operation)
	}
	response, err := c.client.Do(req)
	if err != nil {
		t.Fatal("实际 HTTP 连接失败")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var failure struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(response.Body).Decode(&failure); err != nil {
			t.Fatal("实际 HTTP 错误响应不可解码")
		}
		t.Fatalf("HTTP status=%d code=%s", response.StatusCode, failure.Error.Code)
	}
	if path == "/api/v1/auth/session" {
		var bootstrap struct {
			CSRF string `json:"csrf_token"`
		}
		if err := json.NewDecoder(response.Body).Decode(&bootstrap); err != nil {
			t.Fatal("实际会话启动失败")
		}
		c.csrf = bootstrap.CSRF
		return
	}
	if result != nil {
		if err := json.NewDecoder(response.Body).Decode(result); err != nil {
			t.Fatal("实际 HTTP 响应解码失败")
		}
	}
}

func writeActualJSON(t *testing.T, path string, value any) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal("实际验证文件不可用")
	}
	if err := json.NewEncoder(f).Encode(value); err != nil {
		f.Close()
		t.Fatal("实际验证文件编码失败")
	}
	if err := f.Close(); err != nil {
		t.Fatal("实际验证文件关闭失败")
	}
}
