//go:build public_progress_live

package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func publicProgressOrdinaryPath(t *testing.T, path string, directory bool) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Fatal("an explicit absolute path is required")
	}
	if err := workbenchPathUnredirected(path); err != nil {
		t.Fatal("redirected projection path rejected", err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || !strings.EqualFold(filepath.Clean(path), filepath.Clean(canonical)) {
		t.Fatal("noncanonical projection path rejected", err)
	}
	info, err := os.Lstat(path)
	if err != nil || (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		t.Fatal("ordinary projection path required", err)
	}
}

func publicProgressSourceFiles(t *testing.T) map[string]os.FileInfo {
	t.Helper()
	files := map[string]os.FileInfo{}
	for _, suffix := range []string{"", "-wal"} {
		path := os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB") + suffix
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) && suffix != "" {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		publicProgressOrdinaryPath(t, path, false)
		files[path] = info
	}
	return files
}

func publicProgressFilesUnchanged(t *testing.T, before map[string]os.FileInfo) {
	t.Helper()
	after := publicProgressSourceFiles(t)
	if len(after) != len(before) {
		t.Fatal("source SQLite file membership changed")
	}
	for path, original := range before {
		current, exists := after[path]
		if !exists || !os.SameFile(original, current) || original.Size() != current.Size() || original.ModTime() != current.ModTime() {
			t.Fatal("source SQLite identity, size or timestamp changed")
		}
	}
}

func publicProgressWriteSafeJSON(t *testing.T, root, name string, result any) {
	t.Helper()
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(root, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		t.Fatal(err)
	}
}

// 仅托管既有公开结果；实际生产 handler、鉴权和 SQLite 端口不接原生 Registry。
func TestServePublicProgressProjectionActual(t *testing.T) {
	ctx := context.Background()
	ownedRoot := os.Getenv("ASTROCYTE_TEST_PROGRESS_OWNED_ROOT")
	publicProgressOrdinaryPath(t, ownedRoot, true)
	releaseID := os.Getenv("ASTROCYTE_TEST_PROGRESS_RELEASE_ID")
	if strings.TrimSpace(releaseID) == "" || strings.TrimSpace(releaseID) != releaseID {
		t.Fatal("explicit read-only window release identity required")
	}
	webDir := os.Getenv("ASTROCYTE_WEB_DIR")
	publicProgressOrdinaryPath(t, webDir, true)
	publicProgressOrdinaryPath(t, filepath.Join(webDir, "index.html"), false)
	target := os.Getenv("ASTROCYTE_TEST_PROGRESS_TARGET_DIR")
	if !filepath.IsAbs(target) || !strings.EqualFold(filepath.Dir(filepath.Clean(target)), filepath.Clean(ownedRoot)) {
		t.Fatal("target must be a new direct child of the explicit owned root")
	}
	if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("target already exists or cannot be inspected")
	}
	sourceDir := filepath.Dir(os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB"))
	for _, pair := range [][2]string{{ownedRoot, sourceDir}, {sourceDir, ownedRoot}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			t.Fatal("owned projection root and original source directory must be separate")
		}
	}
	// 所有守卫通过后，才创建新目标；原库连接始终使用 mode=ro。
	sourceFiles := publicProgressSourceFiles(t)
	sourceRecord, projectJSON, projectVersion := readPublicProgress(t)
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	publicProgressOrdinaryPath(t, target, true)
	targetDB := filepath.Join(target, "state.sqlite")
	db, err := Open(targetDB)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?)", sourceRecord.ProjectID, projectVersion, projectJSON); err != nil {
		t.Fatal(err)
	}
	projection := sourceRecord
	projection.Revision = 1
	if err := db.SaveProjectProgress(ctx, projection, 0); err != nil {
		t.Fatal(err)
	}
	verifyStored := func() {
		loaded, err := db.LoadProjectProgress(ctx, projection.ProjectID)
		if err != nil || !reflect.DeepEqual(loaded, projection) {
			t.Fatal("derived progress differs from original fields and local CAS revision", err)
		}
		var storedProject string
		var version int
		if err := db.conn.QueryRowContext(ctx, "SELECT version,data FROM local_agent_projects WHERE id=?", projection.ProjectID).Scan(&version, &storedProject); err != nil || version != projectVersion || storedProject != projectJSON {
			t.Fatal("original project metadata changed", err)
		}
		for table, expected := range map[string]int{"local_agent_projects": 1, "local_project_progress": 1, "local_agent_grants": 0, "local_agent_sessions": 0} {
			var count int
			if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != expected {
				t.Fatal("projection row count differs", table, err)
			}
		}
	}
	verifySource := func() {
		current, currentProject, currentVersion := readPublicProgress(t)
		if !reflect.DeepEqual(current, sourceRecord) || currentProject != projectJSON || currentVersion != projectVersion {
			t.Fatal("original source rows changed")
		}
		publicProgressFilesUnchanged(t, sourceFiles)
	}
	verifyStored()
	logFile, err := os.OpenFile(filepath.Join(ownedRoot, "http.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			t.Error(err)
		}
	}()
	logger := slog.New(slog.NewTextHandler(logFile, nil))
	startHTTP := func() (*httpapi.Server, string, chan error) {
		schemaVersion, err := db.SchemaVersion(ctx)
		if err != nil {
			t.Fatal(err)
		}
		objectStore, err := objects.New(filepath.Join(target, "objects"))
		if err != nil {
			t.Fatal(err)
		}
		materials := attention.NewAttentionService(NewAttentionRepository(db), nil, objectStore, attention.ServiceOptions{})
		projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
		projects.ConfigureRegisteredDiscovery(db, nil)
		progress := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, nil, nil)
		services := httpapi.Services{Foundation: foundation.NewAttentionService(schemaVersion), LocalProjects: projects, RegisteredProjects: projects, Progress: progress,
			Attention: materials, Materials: materials, Opportunities: materials, Tracking: materials,
			Projects: workspace.NewProjectService(), Proposals: workspace.NewProposalService(), Sessions: workspace.NewSessionService(),
			WorkspaceS2: workspace.NewS2Service(NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)}
		server := httpapi.NewServer(httpapi.Config{WebDir: webDir, Logger: logger, Services: services})
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		served := make(chan error, 1)
		go func() { served <- server.Serve(listener) }()
		return server, "http://" + listener.Addr().String(), served
	}
	closed := map[*httpapi.Server]bool{}
	stopHTTP := func(server *httpapi.Server, served <-chan error) {
		if closed[server] {
			return
		}
		closed[server] = true
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			t.Fatal(err)
		}
		if err := <-served; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("production HTTP did not close normally", err)
		}
	}
	readHTTP := func(origin string) domain.ProjectProgress {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Jar: jar, Timeout: 10 * time.Second}
		defer client.CloseIdleConnections()
		get := func(path string, result any) {
			response, err := client.Get(origin + path)
			if err != nil {
				t.Fatal(err)
			}
			decodeErr := json.NewDecoder(response.Body).Decode(result)
			closeErr := response.Body.Close()
			if response.StatusCode != http.StatusOK || decodeErr != nil || closeErr != nil {
				t.Fatal("production authenticated GET failed", path, response.StatusCode, decodeErr, closeErr)
			}
		}
		var session struct {
			ActorKind string `json:"actor_kind"`
			CSRF      string `json:"csrf_token"`
		}
		get("/api/v1/auth/session", &session)
		if session.ActorKind != "human" || session.CSRF == "" {
			t.Fatal("real human bootstrap failed")
		}
		// 当前页面会同时读取这些既有生产服务，均从新自有库返回真实状态。
		for _, path := range []string{"/health", "/foundation", "/projects", "/proposals", "/sessions", "/project-spaces", "/goals", "/local-projects/registered"} {
			var result any
			get("/api/v1"+path, &result)
		}
		var projects struct {
			Items []domain.LocalProject `json:"items"`
		}
		get("/api/v1/local-projects", &projects)
		if len(projects.Items) != 1 || projects.Items[0].ID != projection.ProjectID {
			t.Fatal("HTTP project list differs from fixed original identity")
		}
		var result struct {
			Progress domain.ProjectProgress `json:"progress"`
		}
		get("/api/v1/local-projects/"+projection.ProjectID+"/progress", &result)
		expected := projection
		expected.SchemaVersion = 0
		expected.PendingOperation = ""
		expected.Operations = nil
		expected.ObservedAt = projection.ObservedAt.UTC().Truncate(time.Second)
		if !reflect.DeepEqual(result.Progress, expected) {
			t.Fatal("production HTTP progress projection changed an original public field")
		}
		response, err := client.Get(origin + "/workspace")
		if err != nil {
			t.Fatal(err)
		}
		_, readErr := io.Copy(io.Discard, response.Body)
		closeErr := response.Body.Close()
		if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") || readErr != nil || closeErr != nil {
			t.Fatal("current real frontend static page unavailable", response.StatusCode, readErr, closeErr)
		}
		return result.Progress
	}
	first, firstURL, firstDone := startHTTP()
	defer stopHTTP(first, firstDone)
	firstResult := readHTTP(firstURL)
	stopHTTP(first, firstDone)
	if connection, err := net.DialTimeout("tcp4", strings.TrimPrefix(firstURL, "http://"), time.Second); err == nil {
		if err := connection.Close(); err != nil {
			t.Fatal(err)
		}
		t.Fatal("HTTP1 listener remained open")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = nil
	verifySource()
	db, err = Open(targetDB)
	if err != nil {
		t.Fatal(err)
	}
	verifyStored()
	second, secondURL, secondDone := startHTTP()
	defer stopHTTP(second, secondDone)
	if !reflect.DeepEqual(firstResult, readHTTP(secondURL)) {
		t.Fatal("HTTP2 changed the same-database public progress")
	}
	verifyStored()
	verifySource()
	publicProgressWriteSafeJSON(t, ownedRoot, "ready.json", map[string]any{
		"release_id": releaseID, "pid": os.Getpid(),
		"source_db": os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB"), "target_db": targetDB,
		"source_read_only": true, "source_rows_unchanged": true, "source_files_unchanged": true,
		"source_revision": sourceRecord.Revision, "target_cas_revision": projection.Revision,
		"project_id": projection.ProjectID, "native_id": projection.NativeID, "model": *projection.Model,
		"evidence_count": len(projection.Evidence), "original_operation_count": len(projection.Operations),
		"http1_url": firstURL, "http1_closed": true, "http2_url": secondURL, "same_target_db": true,
		"frontend_dir": webDir, "native_session_count": 0, "grant_count": 0, "new_model_turns": 0,
		"original_source_completeness": "unknown", "effective_tool_set": "unknown",
	})
	t.Log("HTTP1 closed; HTTP2 serves the original public progress from a new owned projection; ready.json contains safe local pointers")
	deadline := time.NewTimer(2 * time.Hour)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-secondDone:
			secondDone <- err
			t.Fatal("owned HTTP2 ended before explicit read-only window release", err)
		case <-deadline.C:
			t.Fatal("read-only window release was not received within two hours")
		case <-ticker.C:
			release, err := os.ReadFile(filepath.Join(ownedRoot, "release.request"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil || strings.TrimSpace(string(release)) != releaseID {
				t.Fatal("invalid explicit window release", err)
			}
			verifyStored()
			verifySource()
			stopHTTP(second, secondDone)
			publicProgressWriteSafeJSON(t, ownedRoot, "final.json", map[string]any{"source_rows_unchanged": true, "source_files_unchanged": true, "target_rows_unchanged": true, "read_only_window_released": true, "http2_closed": true, "new_model_turns": 0})
			return
		}
	}
}
