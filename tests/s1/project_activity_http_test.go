package s1_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// This opt-in runs the normal SQLite/application/HTTP/session stack over a
// controller-approved existing temporary store. No entrypoint discovery or
// attention workers are started, and the connection forbids SQL writes.
func TestActualCachedProjectActivityHTTP(t *testing.T) {
	path := os.Getenv("ASTROCYTE_TEST_PROJECT_ACTIVITY_HTTP")
	if path == "" {
		t.Skip("approved existing temporary SQLite path not provided")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(os.TempDir(), absolute)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		t.Fatal("actual cache must remain within the explicitly named temporary store")
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatal("actual cache must be an existing ordinary SQLite file", err)
	}
	db, err := sqlite.Open(absolute)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Conn().Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	rows := func(query string) [][]any {
		t.Helper()
		result, err := db.Conn().Query(query)
		if err != nil {
			t.Fatal(err)
		}
		defer result.Close()
		columns, err := result.Columns()
		if err != nil {
			t.Fatal(err)
		}
		var values [][]any
		for result.Next() {
			row := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range row {
				pointers[i] = &row[i]
			}
			if err := result.Scan(pointers...); err != nil {
				t.Fatal(err)
			}
			values = append(values, row)
		}
		if err := result.Err(); err != nil {
			t.Fatal(err)
		}
		return values
	}
	queries := []string{
		"SELECT * FROM local_project_discovery ORDER BY id",
		"SELECT * FROM local_project_metadata ORDER BY project_id",
		"SELECT * FROM local_agent_grants ORDER BY id",
		"SELECT * FROM attention_jobs ORDER BY id",
	}
	before := make([][][]any, len(queries))
	for i, query := range queries {
		before[i] = rows(query)
	}
	if len(before[1]) == 0 {
		t.Fatal("existing actual human notes are required for this acceptance")
	}
	prior, err := db.LoadRegisteredProjectDiscovery(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	service := workspaceapp.NewLocalProjectService(nil, nil, nil, nil)
	// A nil source has no discovery capability; the repository is the actual
	// SQLite adapter, including actual persisted human metadata.
	service.ConfigureRegisteredDiscovery(db, nil)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	server := httpapi.NewServer(httpapi.Config{Port: port, Services: httpapi.Services{RegisteredProjects: service}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	ended := make(chan error, 1)
	go func() { ended <- server.ListenAndServe() }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			t.Error(err)
		}
		if err := <-ended; err != http.ErrServerClosed {
			t.Error("owned HTTP server exit", err)
		}
	}()
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://" + server.Addr()
	var bootstrap *http.Response
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		bootstrap, err = client.Get(baseURL + "/api/v1/auth/session")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond) // transport readiness only
	}
	if err != nil {
		t.Fatal(err)
	}
	bootstrap.Body.Close()
	if bootstrap.StatusCode != http.StatusOK || len(bootstrap.Cookies()) == 0 {
		t.Fatal("normal loopback human session was not established")
	}
	request, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/local-projects/registered", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range bootstrap.Cookies() {
		request.AddCookie(cookie)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("actual cached HTTP GET status", response.StatusCode)
	}
	var result struct {
		Snapshot domain.ProjectDiscoverySnapshot `json:"snapshot"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	got := result.Snapshot
	if len(got.Projects) != 111 || len(got.Board) != 71 || !reflect.DeepEqual(got.ObservedAt, prior.ObservedAt) || !reflect.DeepEqual(got.Sources, prior.Sources) {
		t.Fatal("actual cache observations changed or the approved 111/71 store was not used")
	}
	known, humanRecords := 0, 0
	for _, project := range got.Board {
		if project.LastActivityAt != nil {
			known++
		}
		metadata, err := db.LoadProjectMetadata(context.Background(), project.ID)
		if err != nil || !reflect.DeepEqual(project.Human, metadata) {
			t.Fatal("actual persisted human fields differ from HTTP", err)
		}
		if metadata.Revision > 0 {
			humanRecords++
		}
	}
	if known != 17 || humanRecords == 0 {
		t.Fatalf("actual cached activity/human records: known=%d human=%d", known, humanRecords)
	}
	for i, query := range queries {
		if !reflect.DeepEqual(before[i], rows(query)) {
			t.Fatal("cached HTTP GET changed persisted discovery, human fields, grants or jobs")
		}
	}
	t.Logf("task_live_cached_http: roots=%d groups=%d known_activity=%d actual_human_records=%d; SQLite query_only, discovery source absent, persisted rows unchanged", len(got.Projects), len(got.Board), known, humanRecords)
}
