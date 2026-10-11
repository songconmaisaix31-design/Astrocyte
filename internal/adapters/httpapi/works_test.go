package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

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
	read := attentionRequest(s, "GET", "/api/v1/works", "", cookie, "", "", "")
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
		w := attentionRequest(s, "POST", "/api/v1/works/refresh", body, denied.cookie, denied.csrf, denied.bearer, "")
		if w.Code != 403 {
			t.Fatalf("auth accepted: %d", w.Code)
		}
	}
	w := attentionRequest(s, "POST", "/api/v1/works/refresh", command("outside", map[string]any{"roots": []string{filepath.Dir(root)}}), cookie, csrf, "", "")
	if w.Code != 403 {
		t.Fatalf("expanded root: %d %s", w.Code, w.Body.String())
	}
	w = attentionRequest(s, "POST", "/api/v1/works/refresh", command("default", map[string]any{}), cookie, csrf, "", "")
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
	for _, cmd := range []struct {
		path, field string
		value       any
	}{{"attitude", "attitude", "continue"}, {"group", "group", "research"}, {"handoff", "content", "human handoff"}, {"hidden", "hidden", true}} {
		body := command(cmd.path, map[string]any{"expected_revision": revision, cmd.field: cmd.value})
		w = attentionRequest(s, "PUT", "/api/v1/works/"+id+"/"+cmd.path, body, cookie, csrf, "", "")
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", cmd.path, w.Code, w.Body.String())
		}
		var result struct {
			Work domain.Work `json:"work"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		revision = result.Work.Revision
	}
	delivery, _ := json.Marshal(map[string]any{"works": []map[string]any{{"name": "Delivered work", "platform": "codex", "path": filepath.Join(root, "never-read"), "kind": "folder"}}})
	w = attentionRequest(s, "POST", "/api/v1/works/delivery", command("delivery", map[string]any{"content": string(delivery)}), cookie, csrf, "", "")
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
	cookie, _ = bootstrapSession(t, s)
	w = attentionRequest(s, "GET", "/api/v1/works", "", cookie, "", "", "")
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
