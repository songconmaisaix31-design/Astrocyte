package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(Config{
		Port:   0,
		Logger: logger,
		Services: Services{
			Foundation:    foundation.NewService(1),
			Materials:     attentionapp.NewMaterialService(),
			Opportunities: attentionapp.NewOpportunityService(),
			Projects:      workspaceapp.NewProjectService(),
			Proposals:     workspaceapp.NewProposalService(),
			Sessions:      workspaceapp.NewSessionService(),
			Missions:      swarmapp.NewMissionService(),
		},
	})
}

func doRequest(srv *Server, method, path string) *http.Response {
	req := httptest.NewRequest(method, path, nil)
	req.Host = "127.0.0.1"
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)
	return w.Result()
}

func doRequestWithHost(srv *Server, method, path, host string) *http.Response {
	req := httptest.NewRequest(method, path, nil)
	req.Host = host
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)
	return w.Result()
}

func readJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parse JSON %q: %v", body, err)
	}
	return result
}

// readErrorBody reads the ErrorV1 envelope and returns the error object.
func readErrorBody(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	body := readJSON(t, resp)
	if body["schema_version"] != float64(1) {
		t.Errorf("expected schema_version 1, got %v", body["schema_version"])
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in envelope, got %v", body)
	}
	return errObj
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/health")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := readJSON(t, resp)
	if body["schema_version"] != float64(1) {
		t.Errorf("schema_version: got %v, want 1", body["schema_version"])
	}
	if body["status"] != "ok" {
		t.Errorf("status: got %v, want ok", body["status"])
	}
	if body["service"] != "astrocyte" {
		t.Errorf("service: got %v, want astrocyte", body["service"])
	}
}

func TestFoundationEndpoint(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/foundation")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := readJSON(t, resp)
	if body["schema_version"] != float64(1) {
		t.Errorf("schema_version: got %v, want 1", body["schema_version"])
	}
	if body["stage"] != "S0" {
		t.Errorf("stage: got %v, want S0", body["stage"])
	}
	if body["fixture"] != false {
		t.Errorf("fixture: got %v, want false", body["fixture"])
	}

	caps, ok := body["capabilities"].(map[string]any)
	if !ok {
		t.Fatal("capabilities not a map")
	}
	for _, key := range []string{"imports", "approvals", "execution", "native_resume", "handoff"} {
		if caps[key] != false {
			t.Errorf("capabilities.%s: got %v, want false", key, caps[key])
		}
	}

	storage, ok := body["storage"].(map[string]any)
	if !ok {
		t.Fatal("storage not a map")
	}
	if storage["engine"] != "sqlite" {
		t.Errorf("storage.engine: got %v, want sqlite", storage["engine"])
	}
	if storage["schema_version"] != float64(1) {
		t.Errorf("storage.schema_version: got %v, want 1", storage["schema_version"])
	}
}

func TestListEndpoints_Empty(t *testing.T) {
	srv := newTestServer()

	endpoints := []string{
		"/api/v1/materials",
		"/api/v1/opportunities",
		"/api/v1/projects",
		"/api/v1/proposals",
		"/api/v1/sessions",
		"/api/v1/missions",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			resp := doRequest(srv, "GET", ep)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}

			body := readJSON(t, resp)
			if body["schema_version"] != float64(1) {
				t.Errorf("schema_version: got %v, want 1", body["schema_version"])
			}

			items, ok := body["items"].([]any)
			if !ok {
				t.Fatal("items not an array")
			}
			if len(items) != 0 {
				t.Errorf("expected empty items, got %d", len(items))
			}

			if body["next_cursor"] != nil {
				t.Errorf("expected null next_cursor, got %v", body["next_cursor"])
			}
		})
	}
}

func TestGetMission_NotFound(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/missions/nonexistent-id")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	errObj := readErrorBody(t, resp)
	if errObj["code"] != "not_found" {
		t.Errorf("code: got %v, want not_found", errObj["code"])
	}
	if errObj["retryable"] != false {
		t.Errorf("retryable: got %v, want false", errObj["retryable"])
	}
}

func TestNotImplementedRoutes(t *testing.T) {
	srv := newTestServer()

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/materials/imports"},
		{"POST", "/api/v1/opportunities/abc/reviews"},
		{"POST", "/api/v1/opportunities/abc/admissions"},
		{"POST", "/api/v1/projects"},
		{"POST", "/api/v1/sessions/abc/resume"},
		{"POST", "/api/v1/sessions/abc/handoff"},
		{"POST", "/api/v1/proposals/abc/submit"},
		{"POST", "/api/v1/proposals/abc/approvals"},
		{"POST", "/api/v1/approvals/abc/revoke"},
		{"POST", "/api/v1/missions/abc/pause"},
		{"POST", "/api/v1/missions/abc/cancel"},
		{"POST", "/api/v1/work-items/abc/claims"},
		{"POST", "/api/v1/work-items/abc/artifacts"},
		{"POST", "/api/v1/artifacts/abc/acceptance"},
		{"GET", "/api/v1/events"},
	}

	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			resp := doRequest(srv, r.method, r.path)
			if resp.StatusCode != http.StatusNotImplemented {
				t.Fatalf("expected 501, got %d", resp.StatusCode)
			}

			errObj := readErrorBody(t, resp)
			if errObj["code"] != "unsupported_capability" {
				t.Errorf("code: got %v, want unsupported_capability", errObj["code"])
			}
			if errObj["retryable"] != false {
				t.Errorf("retryable: got %v, want false", errObj["retryable"])
			}
		})
	}
}

func TestNotFoundRoute(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/nonexistent")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	errObj := readErrorBody(t, resp)
	if errObj["code"] != "not_found" {
		t.Errorf("code: got %v, want not_found", errObj["code"])
	}
}

func TestLoopbackMiddleware_RejectsNonLoopback(t *testing.T) {
	srv := newTestServer()
	resp := doRequestWithHost(srv, "GET", "/api/v1/health", "evil.example.com")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for non-loopback host, got %d", resp.StatusCode)
	}
}

func TestLoopbackMiddleware_AllowsLocalhost(t *testing.T) {
	srv := newTestServer()
	resp := doRequestWithHost(srv, "GET", "/api/v1/health", "localhost:8787")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for localhost, got %d", resp.StatusCode)
	}
}

func TestLoopbackMiddleware_Allows127(t *testing.T) {
	srv := newTestServer()
	resp := doRequestWithHost(srv, "GET", "/api/v1/health", "127.0.0.1:8787")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for 127.0.0.1, got %d", resp.StatusCode)
	}
}

func TestLoopbackMiddleware_RejectsNonLoopbackOrigin(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://evil.example.com")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-loopback origin, got %d", w.Code)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	srv := newTestServer()

	resp := doRequest(srv, "GET", "/api/v1/health")
	if rid := resp.Header.Get("X-Request-ID"); rid == "" {
		t.Error("expected X-Request-ID header")
	}

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("X-Request-ID", "client-id-123")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if rid := w.Header().Get("X-Request-ID"); rid != "client-id-123" {
		t.Errorf("expected client-id-123, got %q", rid)
	}
}

func TestContentTypeJSON(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/health")

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

func TestIsLoopbackHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.1:8787", true},
		{"localhost", true},
		{"localhost:8787", true},
		{"::1", true},
		{"[::1]:8787", true},
		{"evil.example.com", false},
		{"192.168.1.1", false},
		{"10.0.0.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := isLoopbackHost(tt.host)
			if got != tt.want {
				t.Errorf("isLoopbackHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
