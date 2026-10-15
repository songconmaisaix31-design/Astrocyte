package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type transportAttention struct {
	*attentionapp.Service
	db *sqlite.DB
}

func secureServer(t *testing.T) (*Server, *transportAttention) {
	t.Helper()
	directory := t.TempDir()
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	s := &transportAttention{Service: attentionapp.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attentionapp.ServiceOptions{}), db: db}
	return NewServer(Config{Services: Services{Attention: s}, AgentToken: "test-agent-token", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}), s
}

func importPublicPaper(t *testing.T, service *transportAttention) attentionapp.MaterialDetail {
	t.Helper()
	ctx := context.Background()
	human := attentionapp.Principal{ID: "local-human", Kind: "human"}
	key := rand.Text()
	receipt, err := service.ImportMaterial(ctx, human, attentionapp.ImportMaterialCommand{CommandMeta: attentionapp.CommandMeta{SchemaVersion: 1, RequestID: key, IdempotencyKey: key, ExpectedVersion: 1}, Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	job, err := service.GetJob(ctx, human, receipt.JobID)
	if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
		t.Fatalf("实际公开来源导入失败: %+v %v", job, err)
	}
	material, err := service.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	return material
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

const importBody = `{"schema_version":1,"request_id":"test-request","expected_version":1,"source_locator":"https://arxiv.org/abs/2501.12948v1","source_key":"","content_digest":"","kind":"paper"}`

func TestAttentionTrustedIdentityAndCSRF(t *testing.T) {
	s, service := secureServer(t)
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
	var count int
	if err := service.db.Conn().QueryRow("SELECT COUNT(*) FROM attention_jobs").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("denied calls reached application")
	}
	w := attentionRequest(s, "POST", "/api/v1/materials/imports", importBody, cookie, csrf, "", "http://127.0.0.1:8787")
	if w.Code != 202 {
		t.Fatalf("valid command not correctly authenticated: %d %s", w.Code, w.Body.String())
	}
	var callerJSON, commandJSON string
	if err := service.db.Conn().QueryRow("SELECT caller,payload FROM attention_jobs").Scan(&callerJSON, &commandJSON); err != nil {
		t.Fatal(err)
	}
	var caller attentionapp.Principal
	var command attentionapp.ImportMaterialCommand
	if err := json.Unmarshal([]byte(callerJSON), &caller); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(commandJSON), &command); err != nil {
		t.Fatal(err)
	}
	if caller.Kind != "human" || caller.ID != "local-human" || command.RequestID != "test-request" {
		t.Fatalf("persistent caller/command incorrect: %+v %+v", caller, command)
	}
	if replay := attentionRequest(s, "POST", "/api/v1/materials/imports", importBody, cookie, csrf, "", "http://127.0.0.1:8787"); replay.Code != 202 || !bytes.Equal(replay.Body.Bytes(), w.Body.Bytes()) {
		t.Fatal("caller idempotency key lost")
	}
}
func TestAttentionAgentCannotBootstrapOrReadAnonymously(t *testing.T) {
	s, _ := secureServer(t)
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
	s, service := secureServer(t)
	cookie, csrf := bootstrapSession(t, s)
	material := importPublicPaper(t, service)
	path := "/api/v1/materials/" + material.Material.ID + "/revisions/1/attachments/2501.12948v1.pdf"
	original, err := service.GetAttachment(context.Background(), attentionapp.Principal{ID: "local-human", Kind: "human"}, material.Material.ID, 1, "2501.12948v1.pdf")
	if err != nil {
		t.Fatal(err)
	}
	w := attentionRequest(s, "GET", path, "", cookie, csrf, "", "")
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || !bytes.Equal(w.Body.Bytes(), original.Data) || !bytes.HasPrefix(original.Data, []byte("%PDF-")) || len(original.Data) < 1000 || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("wrong attachment response %d bytes=%d", w.Code, w.Body.Len())
	}
	w = attentionRequest(s, "GET", strings.Replace(path, "/revisions/1/", "/revisions/0/", 1), "", cookie, csrf, "", "")
	if w.Code != 400 {
		t.Fatal("invalid revision accepted")
	}
}

func TestAttentionReviewPreservesPublishedCreatedResponse(t *testing.T) {
	s, service := secureServer(t)
	cookie, csrf := bootstrapSession(t, s)
	material := importPublicPaper(t, service)
	unknown := attentionapp.DimensionScore{Reason: "尚未评估"}
	opportunity, err := service.CreateOpportunity(context.Background(), attentionapp.Principal{ID: "local-human", Kind: "human"}, attentionapp.OpportunityCommand{CommandMeta: attentionapp.CommandMeta{SchemaVersion: 1, RequestID: "create-candidate", IdempotencyKey: "create-candidate", ExpectedVersion: 1}, Title: "人工审阅公开论文", EvidenceRefs: []attentionapp.SourceRef{{MaterialID: material.Material.ID, Revision: 1, Locator: material.Material.SourceLocator}}, Dimensions: attentionapp.Dimensions{GoalProgress: unknown, CurrentInterest: unknown, ProjectImprovement: unknown, Originality: unknown}})
	if err != nil {
		t.Fatal(err)
	}
	w := attentionRequest(s, "POST", "/api/v1/opportunities/"+opportunity.Opportunity.ID+"/reviews", `{"schema_version":1,"request_id":"review-request","expected_version":1,"feedback":"later","reason":"return later"}`, cookie, csrf, "", "")
	if w.Code != http.StatusCreated {
		t.Fatalf("review must preserve contract 201: %d %s", w.Code, w.Body.String())
	}
}
