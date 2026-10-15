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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
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
	if err != nil || !relayInside(expectedEvidence, evidence) {
		t.Fatal("实际验证中间目录必须使用当前 H 工作目录")
	}
	t.Setenv("LOCALAPPDATA", filepath.Join(canonical, "local"))
	port := 17342
	if raw := os.Getenv("ASTROCYTE_L1_API_PORT"); raw != "" {
		port, err = strconv.Atoi(raw)
		if err != nil || port < 0 || port > 65535 {
			t.Fatal("实际服务端口配置错误")
		}
	}
	if (port == 0 || port == 65535) && os.Getenv("ASTROCYTE_L1_WEB_PORT") == "" {
		t.Fatal("实际端口 0 验证必须明确独立 Web 端口")
	}
	webPort, err := configuredPositiveInt("ASTROCYTE_L1_WEB_PORT", port+1, 65535)
	if err != nil {
		t.Fatal("实际 Web 端口配置错误")
	}
	webURL := "http://127.0.0.1:" + strconv.Itoa(webPort)
	root, err := relayRoot()
	if err != nil {
		t.Fatal("独立接入目录不可用")
	}
	source := filepath.Join(root, "verification", "public-project")
	if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
		if upstream := os.Getenv("ASTROCYTE_L1_PUBLIC_GIT_SOURCE"); upstream != "" {
			baseline := os.Getenv("ASTROCYTE_L1_PUBLIC_GIT_COMMIT")
			canonicalUpstream, err := workbenchCanonical(upstream)
			if err != nil || !filepath.IsAbs(upstream) || !strings.EqualFold(filepath.Clean(upstream), filepath.Clean(canonicalUpstream)) || len(baseline) != 40 {
				t.Fatal("明确公开来源或固定 Git 版本无效")
			}
			if err := os.MkdirAll(filepath.Dir(source), 0700); err != nil {
				t.Fatal("公开项目独占父目录不可用")
			}
			for _, args := range [][]string{{"clone", "--no-hardlinks", "--no-checkout", "--", canonicalUpstream, source}, {"-C", source, "checkout", "--detach", baseline}} {
				command := exec.Command("git", args...)
				command.Stdout, command.Stderr = io.Discard, io.Discard
				if command.Run() != nil {
					t.Fatal("实际公开 Git 固定版本消费失败")
				}
			}
		} else {
			if err := os.MkdirAll(source, 0700); err != nil {
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
		}
	} else if err != nil {
		t.Fatal("公开项目身份不可核实")
	}
	if baseline := os.Getenv("ASTROCYTE_L1_PUBLIC_GIT_COMMIT"); baseline != "" {
		command := exec.Command("git", "-C", source, "rev-parse", "HEAD")
		command.Stderr = io.Discard
		head, err := command.Output()
		if err != nil || strings.TrimSpace(string(head)) != baseline {
			t.Fatal("实际公开 Git 项目没有保持明确固定版本")
		}
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
	schemaVersion, err := db.SchemaVersion(context.Background())
	if err != nil {
		t.Fatal("实际数据库版本不可用")
	}
	registry := agents.NewRegistry(filepath.Join(appDir, "native-sessions"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	projects.ConfigureRegisteredDiscovery(db, nil)
	repo := sqlite.NewEventRepository(db)
	eventService := attention.NewPendingEventService(repo)
	works := workspace.NewWorksService(db, worksscan.NewScanner(), nil, events.NewWorksEventSink(eventService, nil), workspace.WorksServiceOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	snapshot, err := works.RefreshWorks(ctx, domain.Caller{Kind: "human", ID: "local-human", OperationID: uuid.NewString()}, workspace.WorksRefreshCommand{Roots: []string{source}})
	if err != nil || len(snapshot.Works) != 1 || len(snapshot.Failures) != 0 {
		t.Fatal("真实公开项目扫描失败")
	}
	_, _, window, err := relayConfiguration()
	if err != nil {
		t.Fatal("实际登记配置不可用")
	}
	if window != 0 {
		t.Fatal("本轮活性窗口保持未配置")
	}
	nativeNotifications := os.Getenv("ASTROCYTE_L1_NATIVE_NOTIFICATIONS") == "1"
	testWindow := time.Duration(0)
	if nativeNotifications {
		testWindow = 10 * time.Second
	}
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", source)
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT", os.Getenv("ASTROCYTE_L1_PUBLIC_GIT_COMMIT"))
	t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "archive")
	if nativeNotifications {
		t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "native_notifications")
		t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "10")
	}
	services := httpapi.Services{Foundation: foundation.NewAttentionService(schemaVersion), Projects: workspace.NewProjectService(), Sessions: workspace.NewSessionService(), Proposals: workspace.NewProposalService(), Works: works, LocalProjects: projects, RegisteredProjects: projects, Progress: workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, registry, registry), Workbench: sqlite.NewWorkbenchReader(filepath.Join(canonical, "snapshot-unconfigured"))}
	launchContexts, err := relayPublicLaunchContexts(projects)
	if err != nil {
		t.Fatal("实际固定公开上下文配置失败")
	}
	cleanup, err := assembleRelay(&services, db, registry, launchContexts...)
	if err != nil {
		t.Fatal("实际生产组装失败")
	}
	defer func() {
		if err := cleanup(); err != nil {
			t.Error("实际生产组装清理失败")
		}
	}()
	wrapped := services.Relay.(*relaySessionRuntime)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	listener, err := bindHTTPListener(port)
	if err != nil {
		t.Fatal("实际服务端口不可用")
	}
	defer listener.Close()
	actualPort := listener.Addr().(*net.TCPAddr).Port
	if actualPort == webPort {
		t.Fatal("实际 API 与 Web 端口必须独立")
	}
	apiURL := "http://" + listener.Addr().String()
	server := httpapi.NewServer(httpapi.Config{Port: actualPort, AllowedOrigins: []string{webURL}, Logger: logger, Services: services})
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	stopObserver := startRelayObserver(ctx, wrapped, logger)
	defer func() {
		cancel()
		stopObserver()
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer closeCancel()
		if err := server.Shutdown(closeCtx); err != nil {
			t.Error("实际 HTTP 服务关闭失败")
		}
		if err := wrapped.Shutdown(closeCtx); err != nil {
			t.Error("实际原生生命周期关闭失败")
		}
		if err := <-serveDone; err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Error("实际 HTTP 服务异常终止")
		}
	}()
	client := newActualRelayHTTP(t, apiURL)
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
		settings.AllowedActions = []string{"start", "observe", "send", "stop", "resume", "reconcile"}
		client.request(t, "PUT", "/api/v1/local-projects/"+project.ID+"/settings", map[string]any{"expected_revision": project.Settings.Revision, "settings": settings}, &response)
		project = response.Project
	}
	if snapshot.Works[0].Handoff.Content == "" {
		client.request(t, "PUT", "/api/v1/works/"+snapshot.Works[0].ID+"/handoff", map[string]any{"expected_revision": snapshot.Works[0].Revision, "content": "Public L1 project. Only the fixed public task context is authorized. Filesystem and command tools remain disabled. Explain native session capture without inspecting files or history."}, nil)
	}
	deadlineSeconds, err := relayNativeDeadlineSeconds()
	if err != nil {
		t.Fatal("实际进程寿命配置错误")
	}
	public := map[string]any{"api_url": apiURL, "web_url": webURL, "project_id": project.ID, "work_id": snapshot.Works[0].ID, "project_revision": project.Settings.Revision, "agent_name": "codex", "service_started_at": time.Now().UTC(), "test_liveness_window_seconds": int(testWindow / time.Second), "native_notifications": nativeNotifications, "session_deadline_seconds": deadlineSeconds}
	writeActualJSON(t, filepath.Join(evidence, "service-public.json"), public)
	writeActualJSON(t, filepath.Join(evidence, "service-private.json"), map[string]any{"app_root": root, "source_root": source, "db_path": filepath.Join(appDir, "state.sqlite")})
	var executionStore *sqlite.RelayStore
	if os.Getenv("ASTROCYTE_L1_CAPTURE_EXECUTION") == "1" {
		executionStore, err = sqlite.OpenRelayStore(root)
		if err != nil {
			t.Fatal("实际进程证据的只读快照不可用")
		}
		defer executionStore.Close()
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-serveDone:
			serveDone <- err
			t.Fatal("实际 HTTP 服务提前终止")
		case <-ticker.C:
			if executionStore != nil {
				captureActualRelayExecutions(t, ctx, registry, executionStore, evidence)
			}
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

// 仅取当前 Registry 原控制器和 SQLite 实际会话，不从原生 ID 重建进程。
func captureActualRelayExecutions(t *testing.T, ctx context.Context, registry *agents.Registry, store *sqlite.RelayStore, evidence string) {
	t.Helper()
	state, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal("实际进程证据的 SQLite 快照读取失败")
	}
	for taskID, task := range state.Tasks {
		session, ok := state.Sessions[task.TerminalID]
		if !ok || session.NativeID == "" {
			continue
		}
		adapter, err := registry.Adapter(session.CLI)
		if err != nil {
			t.Fatal("实际会话的原控制器不可用")
		}
		reader, ok := adapter.(interface {
			OwnedRelayExecution(context.Context, domain.NativeSession) (agents.RelayOwnedExecution, error)
		})
		if !ok {
			t.Fatal("实际会话的内部执行策略端口缺失")
		}
		execution, err := reader.OwnedRelayExecution(ctx, session)
		if err != nil {
			t.Fatal("实际原控制器进程证据核对失败")
		}
		writeActualJSON(t, filepath.Join(evidence, "execution-"+taskID+"-derived.json"), map[string]any{"task_id": taskID, "observed_at": time.Now().UTC(), "execution": execution})
	}
}

type actualRelayHTTP struct {
	client  *http.Client
	csrf    string
	baseURL string
}

func newActualRelayHTTP(t *testing.T, apiURL string) *actualRelayHTTP {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("实际会话 Cookie 不可用")
	}
	validateActualRelayAPIURL(t, apiURL)
	return &actualRelayHTTP{client: &http.Client{Jar: jar, Timeout: 30 * time.Second}, baseURL: apiURL}
}

func actualRelayAPIURL(t *testing.T, evidence string) string {
	t.Helper()
	var public map[string]json.RawMessage
	readActualJSON(t, filepath.Join(evidence, "service-public.json"), &public)
	var apiURL string
	if err := json.Unmarshal(public["api_url"], &apiURL); err != nil {
		t.Fatal("实际服务地址记录不可解码")
	}
	validateActualRelayAPIURL(t, apiURL)
	return apiURL
}

func validateActualRelayAPIURL(t *testing.T, apiURL string) {
	t.Helper()
	parsed, err := url.Parse(apiURL)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		t.Fatal("实际服务地址必须为明确的 loopback HTTP 地址")
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port < 1 || port > 65535 {
		t.Fatal("实际服务地址缺少已绑定端口")
	}
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
	req, err := http.NewRequest(method, c.baseURL+path, &encoded)
	if err != nil {
		t.Fatal("实际请求不可用")
	}
	req.Header.Set("Origin", c.baseURL)
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
