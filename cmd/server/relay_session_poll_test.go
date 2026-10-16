package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type relaySessionOwnedStub struct{ bindings []domain.RelayOwnedSession }

func (s *relaySessionOwnedStub) ListOwnedRelaySessions(context.Context) ([]domain.RelayOwnedSession, error) {
	return s.bindings, nil
}

type relaySessionMemoryState struct {
	entries relaySessionCheckpoints
	saves   int
}

func (s *relaySessionMemoryState) Load() (relaySessionCheckpoints, error) { return s.entries, nil }
func (s *relaySessionMemoryState) Save(v relaySessionCheckpoints) error {
	s.saves++
	data, _ := json.Marshal(v)
	return json.Unmarshal(data, &s.entries)
}

func TestRelaySessionFirstGapPersistsOnceAndSilentPollWritesNothing(t *testing.T) {
	owned := &relaySessionOwnedStub{bindings: []domain.RelayOwnedSession{{TaskID: uuid.NewString(), TerminalID: uuid.NewString(), WorktreeID: uuid.NewString(), ProjectID: "public-project", SourceStatus: "unknown", Gap: "archive_not_selected"}}}
	state := &relaySessionMemoryState{}
	collector := newRelaySessionCollector(owned, nil, state, nil)
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 1 {
		t.Fatal("first source gap not recorded exactly once")
	}
	for i := 0; i < 3; i++ {
		if collector.ObserveOwned(context.Background()) != nil || state.saves != 1 {
			t.Fatal("silent gap poll rewrote metadata")
		}
	}
	owned.bindings[0].Gap = "archive_metadata_unsupported"
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 2 {
		t.Fatal("changed gap not recorded exactly once")
	}
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 2 {
		t.Fatal("unchanged gap rewrote metadata")
	}
}

type relaySessionActualAssembly struct {
	services   httpapi.Services
	db         *sqlite.DB
	root       string
	closeRelay func() error
}

func openRelaySessionActualAssembly(t *testing.T, root string) *relaySessionActualAssembly {
	t.Helper()
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "archive")
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", "")
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "")
	access, err := relayRoot()
	if err != nil || os.MkdirAll(filepath.Join(access, "app-state"), 0700) != nil {
		t.Fatal("实际独占 Relay 接入目录不可用")
	}
	db, err := sqlite.Open(filepath.Join(access, "app-state", "state.sqlite"))
	if err != nil {
		t.Fatal("实际会话数据库打开失败")
	}
	actual := &relaySessionActualAssembly{db: db, root: access}
	t.Cleanup(func() {
		if actual.closeRelay != nil && actual.closeRelay() != nil {
			t.Error("实际 Relay 关闭失败")
		}
		if actual.db != nil && actual.db.Close() != nil {
			t.Error("实际会话数据库关闭失败")
		}
	})
	if db.RunMigrations(context.Background(), migrations.FS) != nil {
		t.Fatal("实际会话数据库迁移失败")
	}
	originals, err := objects.New(filepath.Join(access, "objects"))
	if err != nil {
		t.Fatal("实际会话对象库不可用")
	}
	registry := agents.NewRegistry(filepath.Join(access, "native"))
	actual.services.LocalProjects = workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	actual.closeRelay, err = assembleRelay(&actual.services, db, registry, originals)
	if err != nil {
		t.Fatal("实际空登记 Relay 生产装配失败")
	}
	return actual
}

func (a *relaySessionActualAssembly) close(t *testing.T) {
	t.Helper()
	if a.closeRelay != nil {
		if a.closeRelay() != nil {
			t.Fatal("实际 Relay 正常关闭失败")
		}
		a.closeRelay = nil
	}
	if a.db != nil {
		if a.db.Close() != nil {
			t.Fatal("实际会话数据库正常关闭失败")
		}
		a.db = nil
	}
}

func TestRelaySessionActualEmptyRegistrationPollAndReopen(t *testing.T) {
	ctx := context.Background()
	root := relayActualTestDirectory(t)
	actual := openRelaySessionActualAssembly(t, root)
	assertEmpty := func() {
		t.Helper()
		runtime := actual.services.Relay.(*relaySessionRuntime)
		owned, err := runtime.RelayRuntime.(workspace.RelayOwnedSessionReader).ListOwnedRelaySessions(ctx)
		if err != nil || len(owned) != 0 {
			t.Fatal("空登记基线出现 owned 会话或读取失败")
		}
		for i := 0; i < 3; i++ {
			if runtime.collector.ObserveOwned(ctx) != nil {
				t.Fatal("真实空登记 poll 失败")
			}
			if _, err := os.Stat(filepath.Join(actual.root, "app-state", "session-cursor.json")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("真实空登记 poll 创建了 checkpoint 或目录读取失败")
			}
		}
		entries, err := (relaySessionFileState{root: actual.root}).Load()
		if err != nil || len(entries) != 0 {
			t.Fatal("实际 FileState 空读取失败")
		}
		view, err := actual.services.RelaySessionEvents.ReadSessionEvents(ctx, workbench.SessionEventsQuery{Page: 1, PageSize: 1})
		if err != nil || len(view.Events) != 0 || len(view.Coverage) != 0 || view.Total != 0 || view.Page != 1 || view.PageSize != 1 {
			t.Fatal("实际空会话读取没有保留分页或空集合")
		}
		for _, table := range []string{"attention_engine_events", "local_agent_sessions", "workspace_session_bindings"} {
			var count int
			if actual.db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count) != nil || count != 0 {
				t.Fatal("真实空登记检查创建了事件或会话")
			}
		}
	}
	assertEmpty()
	actual.close(t)
	actual = openRelaySessionActualAssembly(t, root)
	assertEmpty()
}

func TestRelaySessionHTTPActualCollectorGuardsAndClosedStore(t *testing.T) {
	actual := openRelaySessionActualAssembly(t, relayActualTestDirectory(t))
	token := uuid.NewString()
	api := httpapi.NewServer(httpapi.Config{AgentToken: token, Services: actual.services, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	server := httptest.NewServer(api.Handler())
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	request := func(method, path, origin, bearer string, cookie *http.Cookie) (*http.Response, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, nil)
		if err != nil {
			t.Fatal("实际会话 HTTP 请求创建失败")
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal("实际会话 HTTP 请求失败")
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal("实际会话 HTTP 响应读取失败")
		}
		return response, body
	}
	bootstrap, _ := request(http.MethodGet, "/api/v1/auth/session", "", "", nil)
	if bootstrap.StatusCode != http.StatusOK || len(bootstrap.Cookies()) != 1 {
		t.Fatal("实际人工会话不可用")
	}
	cookie := bootstrap.Cookies()[0]
	assertGuards := func() {
		t.Helper()
		for _, item := range []struct {
			method, path, origin, token string
			cookie                      *http.Cookie
			status                      int
			code                        apierrors.Code
		}{
			{"GET", "/api/v1/relay/session-events", "", "", nil, 403, apierrors.ScopeDenied},
			{"GET", "/api/v1/relay/session-events", "https://outside.invalid", "", cookie, 403, apierrors.ScopeDenied},
			{"GET", "/api/v1/relay/session-events", "", token, cookie, 403, apierrors.ScopeDenied},
			{"HEAD", "/api/v1/relay/session-events", "", "", cookie, 405, apierrors.UnsupportedCapability},
			{"GET", "/api/v1/relay/session-events?raw=true", "", "", cookie, 400, apierrors.ValidationFailed},
			{"GET", "/api/v1/relay/session-events?page=1&page=2", "", "", cookie, 400, apierrors.ValidationFailed},
			{"GET", "/api/v1/relay/session-events?page_size=101", "", "", cookie, 400, apierrors.ValidationFailed},
			{"GET", "/api/v1/relay/session-events?task_id=PRIVATE_CANARY", "", "", cookie, 400, apierrors.ValidationFailed},
			{"GET", "/api/v1/relay/session-events?start=2026-10-15T00:00:00Z&end=2026-10-14T00:00:00Z", "", "", cookie, 400, apierrors.ValidationFailed},
		} {
			response, body := request(item.method, item.path, item.origin, item.token, item.cookie)
			if response.StatusCode != item.status {
				t.Fatalf("实际 guard 状态=%d，期望=%d", response.StatusCode, item.status)
			}
			if item.status != 403 && response.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("实际方法或查询拒绝缺少 no-store")
			}
			if item.method != http.MethodHead {
				var envelope apierrors.ErrorEnvelope
				if json.Unmarshal(body, &envelope) != nil || envelope.Error == nil || envelope.Error.Code != item.code || envelope.Error.RequestID == "" || strings.Contains(string(body), "PRIVATE_CANARY") {
					t.Fatal("实际会话错误没有保留安全契约")
				}
			}
		}
	}
	assertGuards()
	response, body := request("GET", "/api/v1/relay/session-events?page_size=1", "", "", cookie)
	var view workbench.SessionEvents
	if json.Unmarshal(body, &view) != nil || response.StatusCode != 200 || response.Header.Get("Cache-Control") != "no-store" || view.Events == nil || view.Coverage == nil || len(view.Events) != 0 || len(view.Coverage) != 0 || view.Total != 0 || view.Page != 1 || view.PageSize != 1 {
		t.Fatal("实际 collector 空结果或 no-store 读取失败")
	}
	if actual.closeRelay() != nil {
		t.Fatal("实际 owned store 正常关闭失败")
	}
	actual.closeRelay = nil
	response, body = request("GET", "/api/v1/relay/session-events?page_size=1", "", "", cookie)
	var unavailable apierrors.ErrorEnvelope
	if json.Unmarshal(body, &unavailable) != nil || response.StatusCode != 503 || unavailable.Error == nil || unavailable.Error.Code != apierrors.ProviderUnavailable || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("真实关闭存储没有产生准确读取失败")
	}
	assertGuards()
}
