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

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func worksHTTPRequest(s *Server, method, path, body string, cookie *http.Cookie, csrf, bearer, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "127.0.0.1:8787"
	r.Header.Set("Content-Type", "application/json")
	var meta struct {
		RequestID string `json:"request_id"`
	}
	_ = json.Unmarshal([]byte(body), &meta)
	r.Header.Set("Idempotency-Key", meta.RequestID)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	r.Header.Set("X-CSRF-Token", csrf)
	r.Header.Set("Authorization", bearer)
	r.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(w, r)
	return w
}

type worksRegisteredRoots struct {
	root      string
	refreshes int
}

func (r *worksRegisteredRoots) ListRegisteredProjects(context.Context, domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	return domain.ProjectDiscoverySnapshot{Projects: []domain.RegisteredProject{{Root: r.root}}}, nil
}
func (r *worksRegisteredRoots) RefreshRegisteredProjects(context.Context, domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	r.refreshes++
	return domain.ProjectDiscoverySnapshot{}, nil
}

func TestWorksHTTPActualStoresRegisteredBoundsAndRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "metadata.txt"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	roots := &worksRegisteredRoots{root: root}
	makeServer := func() *Server {
		return NewServer(Config{Services: Services{Works: workspaceapp.NewWorksService(db, worksscan.NewScanner(), worksscan.Delivery{}, nil, workspaceapp.WorksServiceOptions{}), RegisteredProjects: roots}, AgentToken: "test-agent", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	}
	s := makeServer()
	cookie, csrf := bootstrapSession(t, s)
	command := func(request string, fields map[string]any) string {
		fields["schema_version"], fields["request_id"], fields["expected_version"] = 1, request, 1
		b, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	read := worksHTTPRequest(s, "GET", "/api/v1/works", "", cookie, "", "", "")
	if read.Code != 200 || roots.refreshes != 0 {
		t.Fatalf("cache GET: %d %s", read.Code, read.Body.String())
	}
	var snapshot struct {
		Snapshot domain.WorksSnapshot `json:"snapshot"`
	}
	if err = json.Unmarshal(read.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Snapshot.Status != "unknown" || snapshot.Snapshot.ObservedAt != nil || len(snapshot.Snapshot.Works) != 0 {
		t.Fatalf("invented fresh observation: %+v", snapshot)
	}
	body := command("refresh", map[string]any{"roots": []string{root}})
	for _, denied := range []struct {
		cookie       *http.Cookie
		csrf, bearer string
	}{{nil, csrf, ""}, {cookie, "", ""}, {cookie, csrf, "Bearer test-agent"}} {
		w := worksHTTPRequest(s, "POST", "/api/v1/works/refresh", body, denied.cookie, denied.csrf, denied.bearer, "")
		if w.Code != 403 {
			t.Fatalf("auth accepted: %d", w.Code)
		}
	}
	w := worksHTTPRequest(s, "POST", "/api/v1/works/refresh", command("outside", map[string]any{"roots": []string{filepath.Dir(root)}}), cookie, csrf, "", "")
	if w.Code != 403 {
		t.Fatalf("expanded root: %d %s", w.Code, w.Body.String())
	}
	w = worksHTTPRequest(s, "POST", "/api/v1/works/refresh", command("default", map[string]any{}), cookie, csrf, "", "")
	if w.Code != 200 {
		t.Fatalf("default refresh: %d %s", w.Code, w.Body.String())
	}
	if err = json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Snapshot.Works) != 1 || !snapshot.Snapshot.Works[0].Files.Verified || snapshot.Snapshot.Works[0].Files.FileCount != 1 || snapshot.Snapshot.Works[0].Platform != "unknown" || roots.refreshes != 0 {
		t.Fatalf("scan facts: %+v", snapshot)
	}
	id := snapshot.Snapshot.Works[0].ID
	revision := snapshot.Snapshot.Works[0].Revision
	if err = os.WriteFile(filepath.Join(root, "new-observation.txt"), []byte("later fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	defaultReply := w.Body.String()
	w = worksHTTPRequest(s, "POST", "/api/v1/works/refresh", command("default", map[string]any{}), cookie, csrf, "", "")
	if w.Code != 200 || w.Body.String() != defaultReply {
		t.Fatalf("refresh replay changed result: %d %s", w.Code, w.Body.String())
	}
	for _, cmd := range []struct {
		path, field string
		value       any
	}{{"attitude", "attitude", "continue"}, {"group", "group", "research"}, {"handoff", "content", "human handoff"}, {"hidden", "hidden", true}} {
		body := command(cmd.path, map[string]any{"expected_revision": revision, cmd.field: cmd.value})
		w = worksHTTPRequest(s, "PUT", "/api/v1/works/"+id+"/"+cmd.path, body, cookie, csrf, "", "")
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", cmd.path, w.Code, w.Body.String())
		}
		accepted := w.Body.String()
		w = worksHTTPRequest(s, "PUT", "/api/v1/works/"+id+"/"+cmd.path, body, cookie, csrf, "", "")
		if w.Code != 200 || w.Body.String() != accepted {
			t.Fatalf("metadata replay: %d %s", w.Code, w.Body.String())
		}
		var result struct {
			Work domain.Work `json:"work"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		revision = result.Work.Revision
	}
	w = worksHTTPRequest(s, "PUT", "/api/v1/works/"+id+"/attitude", command("attitude", map[string]any{"expected_revision": 0, "attitude": "park"}), cookie, csrf, "", "")
	if w.Code != 409 {
		t.Fatalf("different input reused receipt: %d %s", w.Code, w.Body.String())
	}
	delivery, _ := json.Marshal(map[string]any{"works": []map[string]any{{"name": "Delivered work", "platform": "codex", "path": filepath.Join(root, "never-read"), "kind": "folder"}}})
	w = worksHTTPRequest(s, "POST", "/api/v1/works/delivery", command("delivery", map[string]any{"content": string(delivery)}), cookie, csrf, "", "")
	if w.Code != 200 {
		t.Fatalf("delivery: %d %s", w.Code, w.Body.String())
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	s = makeServer()
	cookie, csrf = bootstrapSession(t, s)
	w = worksHTTPRequest(s, "POST", "/api/v1/works/refresh", command("default", map[string]any{}), cookie, csrf, "", "")
	if w.Code != 200 || w.Body.String() != defaultReply {
		t.Fatalf("restart receipt replay: %d %s", w.Code, w.Body.String())
	}
	w = worksHTTPRequest(s, "GET", "/api/v1/works", "", cookie, "", "", "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if err = json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Snapshot.Works) != 2 {
		t.Fatalf("restart work count: %+v", snapshot)
	}
	for _, work := range snapshot.Snapshot.Works {
		if work.ID == id && (work.Revision != revision || work.Attitude != "continue" || work.Group != "research" || !work.Hidden || work.Handoff.Content != "human handoff" || work.Handoff.Source != "human") {
			t.Fatalf("restart metadata: %+v", work)
		}
		if work.ID != id && work.Files.Verified {
			t.Fatalf("JSON delivery invented verification: %+v", work)
		}
	}
	if roots.refreshes != 0 {
		t.Fatal("HTTP GET/refresh unexpectedly started registered discovery")
	}
}

func TestEngineIndependentServicesActivateSessionGuard(t *testing.T) {
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	for _, services := range []Services{
		{KnowledgeGovernance: attentionapp.NewKnowledgeService(sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db)), attentionapp.KnowledgeServiceOptions{})},
		{ExperienceGovernance: attentionapp.NewExperienceService(sqlite.NewExperienceRepository(db), sqlite.NewEventRepository(db), attentionapp.ExperienceOptions{ProjectID: "project-a"})},
		{Works: workspaceapp.NewWorksService(db, nil, worksscan.Delivery{}, nil, workspaceapp.WorksServiceOptions{})},
	} {
		s := NewServer(Config{Services: services, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		w := attentionRequest(s, "GET", "/api/v1/works", "", nil, "", "", "")
		if w.Code != 403 {
			t.Fatalf("standalone service bypassed auth: %d", w.Code)
		}
		cookie, csrf := bootstrapSession(t, s)
		if cookie == nil || csrf == "" {
			t.Fatal("standalone service has no session")
		}
		w = attentionRequest(s, "POST", "/api/v1/knowledge/missing/confirm", `{"schema_version":1,"request_id":"confirm","expected_version":1,"owner_project":"project-a"}`, cookie, "", "", "")
		if w.Code != 403 {
			t.Fatalf("standalone service bypassed CSRF: %d", w.Code)
		}
	}
}
