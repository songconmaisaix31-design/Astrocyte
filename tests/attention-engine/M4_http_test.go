package attention_engine_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM4ActualHTTPGovernanceAndColdStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "knowledge.sqlite")
	db, _, service := openM4(t, path)
	defer func() { db.Close() }()
	source := seedM4Source(t, db, "source", "p1")
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "summary", Sources: []domain.EventRef{source}, TS: 2000}); err != nil {
		t.Fatal(err)
	}
	makeHTTP := func() *httptest.Server {
		server := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{KnowledgeGovernance: service}, AgentToken: "test-agent", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		return httptest.NewServer(server.Handler())
	}
	server := makeHTTP()
	defer func() { server.Close() }()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	bootstrap := func() string {
		response, err := client.Get(server.URL + "/api/v1/auth/session")
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var body struct {
			CSRF string `json:"csrf_token"`
		}
		if err = json.NewDecoder(response.Body).Decode(&body); err != nil || response.StatusCode != 200 || body.CSRF == "" {
			t.Fatalf("session=%d csrf=%q err=%v", response.StatusCode, body.CSRF, err)
		}
		return body.CSRF
	}
	csrf := bootstrap()
	request := func(method, path string, body any, token, bearer string, want int) {
		t.Helper()
		data, _ := json.Marshal(body)
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", token)
		req.Header.Set("Idempotency-Key", path+string(data))
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		payload, _ := io.ReadAll(response.Body)
		if response.StatusCode != want {
			t.Fatalf("%s %s = %d want %d: %s", method, path, response.StatusCode, want, payload)
		}
	}
	confirm := app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(1, "http-confirm"), OwnerProject: "p1"}
	request("POST", "/api/v1/knowledge/k/confirm", confirm, "", "", 403)
	request("POST", "/api/v1/knowledge/k/confirm", confirm, csrf, "test-agent", 403)
	request("GET", "/api/v1/knowledge/k?project_id=p1", nil, "", "", 200)
	request("GET", "/api/v1/knowledge/k?project_id=p2", nil, "", "", 403)
	request("POST", "/api/v1/knowledge/k/references", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(1, "http-before-confirm"), ProjectID: "p1"}, csrf, "", 403)
	request("POST", "/api/v1/knowledge/k/confirm", confirm, csrf, "", 200)
	request("POST", "/api/v1/knowledge/k/shares", app.ShareMemoryCommand{CommandMeta: m4Meta(2, "http-share"), OwnerProject: "p1", TargetProject: "p2"}, csrf, "", 200)
	request("POST", "/api/v1/knowledge/k/references", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(3, "http-reference"), ProjectID: "p2"}, csrf, "", 200)
	server.Close()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, _, service = openM4(t, path)
	server = makeHTTP()
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Jar: jar}
	csrf = bootstrap()
	request("GET", "/api/v1/knowledge/k?project_id=p2", nil, "", "", 200)
	request("GET", "/api/v1/knowledge/k?project_id=p3", nil, "", "", 403)
	if _, err := db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	var owner, target, actor, kind, project string
	if err := db.Conn().QueryRow(`SELECT owner_project,target_project,actor_id,actor_kind FROM attention_engine_knowledge_shares WHERE entry_id='k'`).Scan(&owner, &target, &actor, &kind); err != nil {
		t.Fatal(err)
	}
	if owner != "p1" || target != "p2" || actor != "local-human" || kind != "human" {
		t.Fatalf("actual HTTP audit: %s %s %s %s", owner, target, actor, kind)
	}
	if err := db.Conn().QueryRow(`SELECT project_id FROM attention_engine_knowledge_references WHERE entry_id='k'`).Scan(&project); err != nil || project != "p2" {
		t.Fatalf("actual HTTP reference: %s %v", project, err)
	}
}
