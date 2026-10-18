package httpapi_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/openviking"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestOpenVikingActualHTTPUnconfiguredAndRefused(t *testing.T) {
	ctx := context.Background()
	output := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if output == "" || !filepath.IsAbs(output) {
		t.Fatal("需要明确的本次验证输出目录")
	}
	directory, err := os.MkdirTemp(output, "openviking-actual-")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	attention := app.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, app.ServiceOptions{})
	for _, configured := range []bool{false, true} {
		var gateway app.OpenVikingGateway
		if configured {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := listener.Addr().String()
			if err := listener.Close(); err != nil {
				t.Fatal(err)
			}
			gateway, err = openviking.NewClient(openviking.Config{BaseURL: "http://" + address})
			if err != nil {
				t.Fatal(err)
			}
		}
		service, err := app.NewOpenVikingService(ctx, attention, db, gateway)
		if err != nil {
			t.Fatal(err)
		}
		transport := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Attention: attention, Materials: attention, OpenViking: service}, AgentToken: rand.Text(), Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		server := httptest.NewServer(transport.Handler())
		t.Cleanup(server.Close)
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
		bootstrap, err := client.Get(server.URL + "/api/v1/auth/session")
		if err != nil {
			t.Fatal(err)
		}
		var session struct {
			CSRF string `json:"csrf_token"`
		}
		if err := json.NewDecoder(bootstrap.Body).Decode(&session); err != nil {
			t.Fatal(err)
		}
		bootstrap.Body.Close()
		if bootstrap.StatusCode != 200 || session.CSRF == "" {
			t.Fatal("真实人类会话失败")
		}
		request := func(method, path string, body any, csrf, origin string) (int, map[string]json.RawMessage) {
			t.Helper()
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(method, server.URL+"/api/v1"+path, bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrf)
			req.Header.Set("Origin", origin)
			req.Header.Set("Idempotency-Key", rand.Text())
			response, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var data map[string]json.RawMessage
			if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
				t.Fatal(err)
			}
			return response.StatusCode, data
		}
		status, data := request("GET", "/openviking/status", nil, "", server.URL)
		if configured {
			if status != 503 {
				t.Fatal("真实连接拒绝必须返回503", status)
			}
		} else {
			var configuredValue, connected bool
			if err := json.Unmarshal(data["configured"], &configuredValue); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data["connected"], &connected); err != nil {
				t.Fatal(err)
			}
			if status != 200 || configuredValue || connected {
				t.Fatal("未配置状态错误", status)
			}
		}
		status, data = request("GET", "/materials", nil, "", server.URL)
		var materials []any
		if err := json.Unmarshal(data["items"], &materials); err != nil {
			t.Fatal(err)
		}
		if status != 200 || len(materials) != 0 {
			t.Fatal("核心资料空读取失败", status)
		}
		status, data = request("POST", "/project-spaces", map[string]any{"schema_version": 1, "request_id": rand.Text(), "expected_version": 1, "title": "OpenViking 真实边界验证"}, session.CSRF, server.URL)
		var space app.ProjectSpace
		if err := json.Unmarshal(data["space"], &space); err != nil {
			t.Fatal(err)
		}
		if status != 201 || space.ID == "" {
			t.Fatal("真实项目空间创建失败", status)
		}
		path := "/project-spaces/" + space.ID + "/openviking/submissions"
		status, data = request("GET", path, nil, "", server.URL)
		var rows []any
		if err := json.Unmarshal(data["items"], &rows); err != nil {
			t.Fatal(err)
		}
		if status != 200 || len(rows) != 0 {
			t.Fatal("真实提交集合应为空", status)
		}
		for _, denied := range []struct{ csrf, origin string }{{"", server.URL}, {session.CSRF, "https://outside.invalid"}} {
			status, data = request("POST", path, map[string]any{}, denied.csrf, denied.origin)
			var failure apierrors.ServiceError
			if err := json.Unmarshal(data["error"], &failure); err != nil {
				t.Fatal(err)
			}
			if status != 403 || failure.Code != apierrors.ScopeDenied {
				t.Fatal("真实权限拒绝失败", status)
			}
		}
		for _, action := range []string{"search", "content"} {
			body := map[string]any{"query": "项目资料"}
			if action == "content" {
				body = map[string]any{"uri": "viking://resources/astrocyte/projects/" + space.ID + "/README.md", "level": 2}
			}
			status, _ = request("POST", "/project-spaces/"+space.ID+"/openviking/"+action, body, session.CSRF, server.URL)
			if status != 503 {
				t.Fatal("未配置或真实连接拒绝的远端动作应返回503", status)
			}
		}
		for _, body := range []map[string]any{
			{"uri": "viking://resources/astrocyte/projects/" + space.ID + "/README.md"},
			{"uri": "viking://resources/astrocyte/projects/" + space.ID + "/README.md", "level": nil},
		} {
			status, _ = request("POST", "/project-spaces/"+space.ID+"/openviking/content", body, session.CSRF, server.URL)
			if status != 400 {
				t.Fatal("缺失或null层级必须在远端请求前拒绝", status)
			}
		}
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM attention_openviking_submissions").Scan(&count); err != nil || count != 0 {
			t.Fatal("拒绝路径产生了提交记录", count, err)
		}
		server.Close()
	}
	t.Log("真实 HTTP/SQLite 未配置、连接拒绝、空集合、Origin/CSRF 与核心资料读取通过；未调用成功上游")
}
