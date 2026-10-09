package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type transportAttention struct {
	attentionapp.AttentionService
	calls   int
	caller  attentionapp.Principal
	command attentionapp.ImportMaterialCommand
}

func (s *transportAttention) ImportMaterial(_ context.Context, p attentionapp.Principal, c attentionapp.ImportMaterialCommand) (attentionapp.ImportJobResult, error) {
	s.calls++
	s.caller = p
	s.command = c
	return attentionapp.ImportJobResult{SchemaVersion: 1, JobID: "actual-job", Status: "queued"}, nil
}
func (s *transportAttention) GetAttachment(_ context.Context, _ attentionapp.Principal, _ string, _ int, _ string) (attentionapp.AttachmentContent, error) {
	return attentionapp.AttachmentContent{Data: []byte("%PDF-original"), Name: "source.pdf", MediaType: "application/pdf"}, nil
}
func secureServer() (*Server, *transportAttention) {
	s := &transportAttention{}
	return NewServer(Config{Services: Services{Attention: s}, AgentToken: "test-agent-token", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}), s
}
func attentionRequest(s *Server, method, path, body string, cookie *http.Cookie, csrf, bearer, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "127.0.0.1:8787"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Idempotency-Key", "caller-request-key")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	if bearer != "" {
		r.Header.Set("Authorization", bearer)
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, r)
	return w
}
func bootstrapSession(t *testing.T, s *Server) (*http.Cookie, string) {
	t.Helper()
	w := attentionRequest(s, "GET", "/api/v1/auth/session", "", nil, "", "", "")
	if w.Code != 200 {
		t.Fatalf("bootstrap: %d %s", w.Code, w.Body.String())
	}
	var data struct {
		CSRF string `json:"csrf_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data.CSRF == "" || len(w.Result().Cookies()) != 1 {
		t.Fatal("missing CSRF/session")
	}
	return w.Result().Cookies()[0], data.CSRF
}

const importBody = `{"schema_version":1,"request_id":"test-request","expected_version":1,"source_locator":"arxiv:2401.00001v1","source_key":"","content_digest":"","kind":"paper"}`

func TestAttentionTrustedIdentityAndCSRF(t *testing.T) {
	s, app := secureServer()
	cookie, csrf := bootstrapSession(t, s)
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("cookie protection missing")
	}
	cases := []struct {
		name                       string
		cookie                     *http.Cookie
		csrf, bearer, origin, body string
		status                     int
	}{
		{"no session", nil, "", "", "", importBody, 403},
		{"no CSRF", cookie, "", "", "", importBody, 403},
		{"invalid CSRF", cookie, "invalid", "", "", importBody, 403},
		{"different local Origin", cookie, csrf, "", "http://127.0.0.1:9999", importBody, 403},
		{"Agent cannot write with human cookie", cookie, csrf, "Bearer test-agent-token", "", importBody, 403},
		{"invalid bearer cannot fall back to cookie", cookie, csrf, "Bearer invalid", "", importBody, 403},
		{"body actor is rejected", cookie, csrf, "", "", strings.TrimSuffix(importBody, "}") + `,"actor_type":"human"}`, 400},
		{"unknown field", cookie, csrf, "", "", strings.TrimSuffix(importBody, "}") + `,"approve":true}`, 400},
		{"trailing command", cookie, csrf, "", "", importBody + importBody, 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := attentionRequest(s, "POST", "/api/v1/materials/imports", c.body, c.cookie, c.csrf, c.bearer, c.origin)
			if w.Code != c.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
	if app.calls != 0 {
		t.Fatal("denied calls reached application")
	}
	w := attentionRequest(s, "POST", "/api/v1/materials/imports", importBody, cookie, csrf, "", "http://127.0.0.1:8787")
	if w.Code != 202 || app.calls != 1 || app.caller.Kind != "human" || app.caller.ID != "local-human" || app.command.IdempotencyKey != "caller-request-key" {
		t.Fatalf("valid command not correctly authenticated: %d %s %+v", w.Code, w.Body.String(), app)
	}
}
func TestAttentionAgentCannotBootstrapOrReadAnonymously(t *testing.T) {
	s, _ := secureServer()
	if w := attentionRequest(s, "GET", "/api/v1/auth/session", "", nil, "", "Bearer test-agent-token", ""); w.Code != 403 {
		t.Fatal("Agent bootstrapped human session")
	}
	if w := attentionRequest(s, "GET", "/api/v1/jobs", "", nil, "", "", ""); w.Code != 403 {
		t.Fatal("unauthenticated protected read allowed")
	}
	for _, path := range []string{"/api/v1/materials", "/api/v1/materials/m1", "/api/v1/distillations", "/api/v1/opportunities", "/api/v1/materials/m1/revisions/1/attachments/source.pdf", "/api/v1/jobs"} {
		if w := attentionRequest(s, "GET", path, "", nil, "", "Bearer test-agent-token", ""); w.Code != 403 {
			t.Fatalf("unscoped Agent access allowed: %s", path)
		}
	}
	r := httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	r.Host = "127.0.0.1:8787"
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-site bootstrap allowed")
	}
}
func TestAttentionAttachmentOriginalMedia(t *testing.T) {
	s, _ := secureServer()
	cookie, csrf := bootstrapSession(t, s)
	w := attentionRequest(s, "GET", "/api/v1/materials/m1/revisions/1/attachments/source.pdf", "", cookie, csrf, "", "")
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || w.Body.String() != "%PDF-original" || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("wrong attachment response %d %s", w.Code, w.Body.String())
	}
	w = attentionRequest(s, "GET", "/api/v1/materials/m1/revisions/0/attachments/source.pdf", "", cookie, csrf, "", "")
	if w.Code != 400 {
		t.Fatal("invalid revision accepted")
	}
}
