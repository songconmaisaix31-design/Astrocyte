package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func newTestServer() *Server {
	return newTestServerWithConfig(Config{})
}

func newTestServerWithConfig(extra Config) *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{
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
	}
	if extra.WebDir != "" {
		cfg.WebDir = extra.WebDir
	}
	return NewServer(cfg)
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

// --- Health and Foundation ---

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

func TestFoundationEndpoint_Structure(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/foundation")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := readJSON(t, resp)
	if body["schema_version"] != float64(1) {
		t.Errorf("schema_version: got %v, want 1", body["schema_version"])
	}
	// Verify structure has expected top-level keys; detailed field values
	// are tested in foundation/service_test.go.
	for _, key := range []string{"stage", "capabilities", "storage", "fixture"} {
		if _, ok := body[key]; !ok {
			t.Errorf("missing key %q in foundation response", key)
		}
	}
}

// --- List endpoints ---

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

// --- Single resource ---

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

// --- 501 routes ---

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

func TestAPINotFound_JSON404(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/nonexistent")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	errObj := readErrorBody(t, resp)
	if errObj["code"] != "not_found" {
		t.Errorf("code: got %v, want not_found", errObj["code"])
	}
	if errObj["required_action"] != "check_path_and_method" {
		t.Errorf("required_action: got %v, want check_path_and_method", errObj["required_action"])
	}
}

func TestAPINotFound_BareAPI_JSON404(t *testing.T) {
	srv := newTestServer()

	// Bare /api (no trailing slash) should return JSON 404
	resp := doRequest(srv, "GET", "/api")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for /api, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type for /api: got %q, want application/json", ct)
	}
	errObj := readErrorBody(t, resp)
	if errObj["code"] != "not_found" {
		t.Errorf("code for /api: got %v, want not_found", errObj["code"])
	}

	// Bare /api/v1 (no trailing slash) should also return JSON 404
	resp = doRequest(srv, "GET", "/api/v1")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for /api/v1, got %d", resp.StatusCode)
	}
	ct = resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type for /api/v1: got %q, want application/json", ct)
	}
	errObj = readErrorBody(t, resp)
	if errObj["code"] != "not_found" {
		t.Errorf("code for /api/v1: got %v, want not_found", errObj["code"])
	}
}

// --- Loopback middleware with ErrorV1 ---

func TestLoopbackMiddleware_RejectsNonLoopback_ErrorV1(t *testing.T) {
	srv := newTestServer()
	resp := doRequestWithHost(srv, "GET", "/api/v1/health", "evil.example.com")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	errObj := readErrorBody(t, resp)
	if errObj["code"] != "scope_denied" {
		t.Errorf("code: got %v, want scope_denied", errObj["code"])
	}
	if errObj["retryable"] != false {
		t.Errorf("retryable: got %v, want false", errObj["retryable"])
	}
	if errObj["request_id"] == nil || errObj["request_id"] == "" {
		t.Error("expected nonempty request_id in ErrorV1")
	}
	if errObj["required_action"] == nil || errObj["required_action"] == "" {
		t.Error("expected nonempty required_action in ErrorV1")
	}

	// X-Request-ID header should be set.
	if rid := resp.Header.Get("X-Request-ID"); rid == "" {
		t.Error("expected X-Request-ID header on rejected response")
	}
}

func TestLoopbackMiddleware_RejectsNonLoopbackOrigin_ErrorV1(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("Origin", "http://evil.example.com")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}

	errObj, ok := env["error"].(map[string]any)
	if !ok {
		t.Fatal("missing error object in rejection response")
	}
	if errObj["code"] != "scope_denied" {
		t.Errorf("code: got %v, want scope_denied", errObj["code"])
	}
	if errObj["request_id"] == nil || errObj["request_id"] == "" {
		t.Error("expected nonempty request_id")
	}
	if errObj["required_action"] == nil || errObj["required_action"] == "" {
		t.Error("expected nonempty required_action")
	}
}

func TestLoopbackMiddleware_AllowsLoopbackHosts(t *testing.T) {
	srv := newTestServer()
	for _, host := range []string{"localhost:8787", "127.0.0.1:8787"} {
		t.Run(host, func(t *testing.T) {
			resp := doRequestWithHost(srv, "GET", "/api/v1/health", host)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", host, resp.StatusCode)
			}
		})
	}
}

// --- Request ID ---

func TestRequestIDMiddleware(t *testing.T) {
	srv := newTestServer()

	// Auto-generated.
	resp := doRequest(srv, "GET", "/api/v1/health")
	if rid := resp.Header.Get("X-Request-ID"); rid == "" {
		t.Error("expected X-Request-ID header")
	}

	// Client-supplied.
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req.Host = "127.0.0.1"
	req.Header.Set("X-Request-ID", "client-id-123")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if rid := w.Header().Get("X-Request-ID"); rid != "client-id-123" {
		t.Errorf("expected client-id-123, got %q", rid)
	}
}

// --- Content-Type ---

func TestContentTypeJSON(t *testing.T) {
	srv := newTestServer()
	resp := doRequest(srv, "GET", "/api/v1/health")

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

// --- isLoopbackHost ---

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

// --- Static file serving (SPA) ---

func setupWebDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create index.html
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>SPA</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a static asset
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestStaticServing_ServesExistingFile(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	resp := doRequest(srv, "GET", "/assets/style.css")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "body{}" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestStaticServing_SPAServesIndexForNonAPI(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	// Non-existent path should get index.html (SPA fallback)
	resp := doRequest(srv, "GET", "/some/spa/route")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>SPA</html>" {
		t.Errorf("expected index.html content, got %q", body)
	}
}

func TestStaticServing_RootServesIndex(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	resp := doRequest(srv, "GET", "/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>SPA</html>" {
		t.Errorf("expected index.html content, got %q", body)
	}
}

func TestStaticServing_APIRoutesNotServedByStatic(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	// /api/v1/health should still go to API handler
	resp := doRequest(srv, "GET", "/api/v1/health")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := readJSON(t, resp)
	if body["status"] != "ok" {
		t.Errorf("expected health ok, got %v", body["status"])
	}

	// Unknown /api/v1/* should return JSON 404, not index.html
	resp = doRequest(srv, "GET", "/api/v1/nonexistent")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	// Bare /api should return JSON 404, not index.html
	resp = doRequest(srv, "GET", "/api")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for bare /api, got %d", resp.StatusCode)
	}
	ct = resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type for bare /api: got %q, want application/json", ct)
	}
}

func TestStaticServing_NoDirectoryListing(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	// Request for directory should serve index.html (SPA fallback), not listing
	resp := doRequest(srv, "GET", "/assets/")
	body, _ := io.ReadAll(resp.Body)
	// Should be SPA index, not a directory listing
	if string(body) != "<html>SPA</html>" {
		t.Errorf("expected SPA fallback for directory, got %q", body)
	}
}

func TestStaticServing_TraversalPrevented(t *testing.T) {
	webDir := setupWebDir(t)
	srv := newTestServerWithConfig(Config{WebDir: webDir})

	// Path traversal attempt: Go's http.ServeFile returns 400 for .. in paths,
	// or our SPA fallback serves index.html. Either way, no real file is exposed.
	resp := doRequest(srv, "GET", "/../etc/passwd")
	body, _ := io.ReadAll(resp.Body)

	// Acceptable: 400 Bad Request (Go rejects ..) or 200 with SPA index
	if resp.StatusCode == http.StatusBadRequest {
		return // correctly rejected
	}
	if string(body) == "<html>SPA</html>" {
		return // SPA fallback, no real file exposed
	}
	// Must not contain passwd-like content
	if resp.StatusCode == http.StatusOK && len(body) > 0 && string(body) != "<html>SPA</html>" {
		t.Errorf("traversal may have served real file: status=%d body_len=%d", resp.StatusCode, len(body))
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
