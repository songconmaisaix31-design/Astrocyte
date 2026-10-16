package httpapi

import (
	"context"
	"encoding/json"
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
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelaySessionHTTPActualUnconfiguredProviderGuards(t *testing.T) {
	parent := filepath.Join("..", "..", "..", "dist", "relay-session-http")
	if selected := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT"); selected != "" {
		if !filepath.IsAbs(selected) || strings.TrimSpace(selected) != selected {
			t.Fatal("实际 HTTP 输出根必须为明确绝对目录")
		}
		parent = filepath.Join(selected, "relay-session-http")
	}
	parent, err := filepath.Abs(parent)
	if err != nil {
		t.Fatal("实际 HTTP 输出目录无法解析")
	}
	if os.MkdirAll(parent, 0700) != nil {
		t.Fatal("实际 HTTP 检查目录不可用")
	}
	root, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("实际 HTTP 独占目录不可用")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal("实际 HTTP 物理目录不可用")
	}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal("实际 HTTP 数据库不可用")
	}
	defer db.Close()
	if db.RunMigrations(context.Background(), migrations.FS) != nil {
		t.Fatal("实际 HTTP 数据库迁移失败")
	}
	registry := agents.NewRegistry(filepath.Join(root, "native"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	token := uuid.NewString()
	server := httptest.NewServer(NewServer(Config{AgentToken: token, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: Services{LocalProjects: projects}}).Handler())
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	request := func(method, path, origin, token string, cookie *http.Cookie) (*http.Response, []byte) {
		t.Helper()
		r, e := http.NewRequest(method, server.URL+path, nil)
		if e != nil {
			t.Fatal("public request unavailable")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		response, e := client.Do(r)
		if e != nil {
			t.Fatal("transport unavailable")
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal("实际 HTTP 响应读取失败")
		}
		return response, body
	}
	bootstrap, _ := request("GET", "/api/v1/auth/session", "", "", nil)
	if bootstrap.StatusCode != 200 || len(bootstrap.Cookies()) != 1 {
		t.Fatal("实际人工 HTTP 会话不可用")
	}
	cookie := bootstrap.Cookies()[0]
	for _, c := range []struct {
		method, path, origin, token string
		cookie                      *http.Cookie
		status                      int
	}{
		{"GET", "/api/v1/relay/session-events", "", "", nil, 403},
		{"GET", "/api/v1/relay/session-events", "https://outside.invalid", "", cookie, 403},
		{"GET", "/api/v1/relay/session-events", "", token, cookie, 403},
		{"HEAD", "/api/v1/relay/session-events", "", "", cookie, 405},
		{"GET", "/api/v1/relay/session-events?raw=true", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?page=1&page=2", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?page_size=101", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?task_id=PRIVATE_CANARY", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?start=2026-10-15T00:00:00Z&end=2026-10-14T00:00:00Z", "", "", cookie, 400},
	} {
		response, body := request(c.method, c.path, c.origin, c.token, c.cookie)
		if response.StatusCode != c.status {
			t.Fatalf("guard status=%d expected=%d", response.StatusCode, c.status)
		}
		if c.status != 403 && response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("实际方法或查询拒绝缺少 no-store")
		}
		if c.method != http.MethodHead {
			var failure apierrors.ErrorEnvelope
			if json.Unmarshal(body, &failure) != nil || failure.Error == nil || failure.Error.RequestID == "" || strings.Contains(string(body), "PRIVATE_CANARY") {
				t.Fatal("实际拒绝缺少安全错误契约")
			}
		}
	}
	response, body := request("GET", "/api/v1/relay/session-events?page_size=1", "", "", cookie)
	var unavailable apierrors.ErrorEnvelope
	if json.Unmarshal(body, &unavailable) != nil || response.StatusCode != 503 || unavailable.Error == nil || unavailable.Error.Code != apierrors.ProviderUnavailable || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("实际未配置会话 provider 没有返回安全503和no-store")
	}
}
