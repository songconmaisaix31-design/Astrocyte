package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 只读此前公开 README 验证的准确记录；原生身份和推断收据保持原始来源。
func openPublicProgressSource(t *testing.T) *DB {
	t.Helper()
	publication := os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB")
	if !filepath.IsAbs(publication) {
		t.Fatal("select the absolute SQLite path of the actual public README progress result")
	}
	if err := workbenchPathUnredirected(publication); err != nil {
		t.Fatal("public progress source contains a redirected path", err)
	}
	canonical, err := filepath.EvalSymlinks(publication)
	if err != nil || !strings.EqualFold(filepath.Clean(canonical), filepath.Clean(publication)) {
		t.Fatal("public progress source is not its explicit canonical path", err)
	}
	info, err := os.Lstat(publication)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatal("public progress source must be an existing regular SQLite file", err)
	}
	locator := &url.URL{Scheme: "file", Path: filepath.ToSlash(publication)}
	if filepath.VolumeName(publication) != "" {
		locator.Path = "/" + locator.Path
	}
	query := locator.Query()
	query.Set("mode", "ro")
	query.Set("_pragma", "query_only(ON)")
	locator.RawQuery = query.Encode()
	conn, err := sql.Open("sqlite", locator.String())
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(1)
	return &DB{conn: conn, dbPath: publication}
}

func readPublicProgress(t *testing.T) (domain.ProjectProgress, string, int) {
	t.Helper()
	source := openPublicProgressSource(t)
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	return readPublicProgressFrom(t, source)
}

func readPublicProgressFrom(t *testing.T, source *DB) (domain.ProjectProgress, string, int) {
	t.Helper()
	const projectID = "25b3f37a-a96c-4082-9a59-6ed6707fbf70"
	record, err := source.LoadProjectProgress(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	if record.ProjectID != projectID || record.Source != "agent_inferred" || record.NativeID != "01a13a73-1de2-7761-aaa3-5cbd8fb433cc" || record.Model == nil || *record.Model != "gpt-6.1-sol" || record.Status == "" || len(record.Evidence) == 0 || len(record.Operations) == 0 {
		t.Fatal("actual public progress provenance is unavailable")
	}
	for _, operation := range record.Operations {
		if operation.Action != "infer" || operation.Status != "accepted" || operation.Result == nil || operation.Result.NativeID != record.NativeID || len(operation.Result.Evidence) == 0 {
			t.Fatal("actual accepted public inference receipt is unavailable")
		}
	}
	var projectJSON string
	var version int
	if err := source.conn.QueryRow("SELECT version,data FROM local_agent_projects WHERE id=?", projectID).Scan(&version, &projectJSON); err != nil {
		t.Fatal(err)
	}
	var project domain.LocalProject
	if err := json.Unmarshal([]byte(projectJSON), &project); err != nil || project.ID != projectID || project.Settings.Revision != version {
		t.Fatal("actual registered source project identity differs", err)
	}
	return record, projectJSON, version
}

func TestProgressStorageRoundTripAndCAS(t *testing.T) {
	ctx := context.Background()
	record, projectJSON, projectVersion := readPublicProgress(t)
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	if got, err := db.LoadProjectProgress(ctx, uuid.NewString()); err != nil || got.Status != "" || got.Revision != 0 {
		t.Fatalf("absent project should be zero value, got %+v %v", got, err)
	}
	// 自有副本保留原项目元数据；新的进度投影从本库 CAS revision 1 开始。
	if _, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?)", record.ProjectID, projectVersion, projectJSON); err != nil {
		t.Fatal(err)
	}
	record.Revision = 1
	if err := db.SaveProjectProgress(ctx, record, 0); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.LoadProjectProgress(ctx, record.ProjectID)
	if err != nil || !reflect.DeepEqual(loaded, record) {
		t.Fatalf("actual public observation round trip changed: %v", err)
	}

	if err := db.SaveProjectProgress(ctx, record, 0); err == nil {
		t.Fatal("stale insert accepted")
	}
	percent := 50
	service := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, nil, nil)
	updated, err := service.SetProjectProgress(ctx, domain.Caller{Kind: "human", ID: "local-human"}, record.ProjectID, domain.ProgressInput{Status: "review", Percent: &percent})
	if err != nil || updated.Revision != 2 || updated.Source != "human" {
		t.Fatal("actual human progress update failed", err)
	}
	loaded, err = db.LoadProjectProgress(ctx, record.ProjectID)
	if err != nil || loaded.Status != "review" || loaded.Revision != 2 || loaded.Source != "human" || loaded.Percent == nil || *loaded.Percent != percent || loaded.NativeID != "" || !reflect.DeepEqual(loaded.Operations, record.Operations) {
		t.Fatal("actual human update lost its source or immutable original inference receipt", err)
	}
	projects, err := db.ListProjects(ctx)
	if err != nil || len(projects) != 1 || projects[0].ID != record.ProjectID {
		t.Fatal("progress storage created or changed the actual source project")
	}
	_, err = db.LoadGrant(ctx, record.ProjectID, "ungranted-input")
	var missing *apierrors.ServiceError
	if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
		t.Fatal("progress storage created a grant or lost the actual missing result", err)
	}
	for _, table := range []string{"local_agent_sessions", "local_agent_grants"} {
		var count int
		if err := db.conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("progress persistence created native work or permission", table, count, err)
		}
	}
}

func progressReceiptPath(t *testing.T, path string, directory bool) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Fatal("explicit absolute receipt path required")
	}
	if err := workbenchPathUnredirected(path); err != nil {
		t.Fatal("redirected receipt path rejected", err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || !strings.EqualFold(filepath.Clean(path), filepath.Clean(canonical)) {
		t.Fatal("noncanonical receipt path rejected", err)
	}
	info, err := os.Lstat(path)
	if err != nil || (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		t.Fatal("ordinary receipt path required", err)
	}
}

func progressReceiptJSON(t *testing.T, root, name string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(root, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		t.Fatal(err)
	}
}

func progressReceiptFiles(t *testing.T, path string) map[string]os.FileInfo {
	t.Helper()
	result := map[string]os.FileInfo{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		file := path + suffix
		info, err := os.Lstat(file)
		if errors.Is(err, os.ErrNotExist) && suffix != "" {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		progressReceiptPath(t, file, false)
		result[suffix] = info
	}
	return result
}

func progressReceiptCompareFiles(t *testing.T, before, after map[string]os.FileInfo) map[string]any {
	t.Helper()
	result := map[string]any{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		original, originalExists := before[suffix]
		current, currentExists := after[suffix]
		unchanged := originalExists == currentExists
		facts := map[string]any{"before_exists": originalExists, "after_exists": currentExists}
		if originalExists {
			facts["before_bytes"] = original.Size()
			facts["before_mtime"] = original.ModTime().UTC().Format(time.RFC3339Nano)
		}
		if currentExists {
			facts["after_bytes"] = current.Size()
			facts["after_mtime"] = current.ModTime().UTC().Format(time.RFC3339Nano)
		}
		if originalExists && currentExists {
			facts["same_identity"] = os.SameFile(original, current)
			unchanged = os.SameFile(original, current) && original.Size() == current.Size() && original.ModTime().Equal(current.ModTime())
		}
		facts["unchanged"] = unchanged
		result[suffix] = facts
		if suffix != "-shm" && !unchanged {
			t.Fatal("source main/WAL identity, size or mtime changed")
		}
	}
	return result
}

// 人工阶段写入和同库重启均须保留真实已受理回执，原操作重读只返回原结果。
func TestProgressHumanSetPreservesOperationReceipts(t *testing.T) {
	ctx := context.Background()
	root := os.Getenv("ASTROCYTE_TEST_PROGRESS_RECEIPTS_ROOT")
	progressReceiptPath(t, root, true)
	publication := os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB")
	progressReceiptPath(t, publication, false)
	for _, pair := range [][2]string{{root, filepath.Dir(publication)}, {filepath.Dir(publication), root}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			t.Fatal("source and owned receipt root must be separate")
		}
	}
	target := filepath.Join(root, "projection")
	nativeRoot := filepath.Join(root, "native-state")
	for _, path := range []string{target, nativeRoot} {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("receipt target and unused native state root must be new")
		}
	}
	filesBefore := progressReceiptFiles(t, publication)
	source := openPublicProgressSource(t)
	sourceClosed := false
	defer func() {
		if !sourceClosed {
			if err := source.Close(); err != nil {
				t.Error(err)
			}
			progressReceiptJSON(t, root, "source-early-close-files.json", progressReceiptCompareFiles(t, filesBefore, progressReceiptFiles(t, publication)))
		}
	}()
	var queryOnly int
	var locking string
	if err := source.conn.QueryRow("PRAGMA query_only").Scan(&queryOnly); err != nil || queryOnly != 1 {
		t.Fatal("source query_only is not enforced", err)
	}
	if err := source.conn.QueryRow("PRAGMA locking_mode").Scan(&locking); err != nil || locking != "normal" {
		t.Fatal("source must use normal SQLite locks", err)
	}
	original, projectJSON, projectVersion := readPublicProgressFrom(t, source)
	var project domain.LocalProject
	if err := json.Unmarshal([]byte(projectJSON), &project); err != nil || project.Settings.ExternalModelCLI == "" {
		t.Fatal("original model consent metadata unavailable", err)
	}
	var operationID string
	var receipt domain.ProgressOperation
	if len(original.Operations) != 1 {
		t.Fatal("exact original accepted operation count differs")
	}
	for key, operation := range original.Operations {
		if _, err := uuid.Parse(key); err != nil || operation.Action != "infer" || operation.Status != "accepted" || operation.CreatedAt.IsZero() || operation.FinishedAt == nil || operation.FinishedAt.IsZero() || operation.Result == nil || len(operation.Result.Evidence) == 0 || operation.Result.Evidence[0].SourcePath == "" || operation.Result.ObservedAt.IsZero() {
			t.Fatal("original accepted operation precondition unavailable", err)
		}
		operationID, receipt = key, operation
	}
	if original.Status != receipt.Result.Status || original.Summary != receipt.Result.Summary || !reflect.DeepEqual(original.Percent, receipt.Result.Percent) || !reflect.DeepEqual(original.Evidence, receipt.Result.Evidence) || original.NativeID != receipt.Result.NativeID || !reflect.DeepEqual(original.Model, receipt.Result.Model) || !original.ObservedAt.Equal(receipt.Result.ObservedAt) {
		t.Fatal("original observation and accepted result differ")
	}
	var sourceProgressJSON string
	var sourceProgressRevision int
	if err := source.conn.QueryRow("SELECT revision,data FROM local_project_progress WHERE project_id=?", original.ProjectID).Scan(&sourceProgressRevision, &sourceProgressJSON); err != nil || sourceProgressRevision != original.Revision {
		t.Fatal("original stored progress revision differs", err)
	}
	progressReceiptJSON(t, root, "source-before-target.json", progressReceiptCompareFiles(t, filesBefore, progressReceiptFiles(t, publication)))
	for _, path := range []string{target, nativeRoot} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		progressReceiptPath(t, path, true)
	}
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
	if _, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?)", original.ProjectID, projectVersion, projectJSON); err != nil {
		t.Fatal(err)
	}
	projection := original
	projection.Revision = 1
	if err := db.SaveProjectProgress(ctx, projection, 0); err != nil {
		t.Fatal(err)
	}
	logFile, err := os.OpenFile(filepath.Join(root, "http.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			t.Error(err)
		}
	}()
	logger := slog.New(slog.NewTextHandler(logFile, nil))
	start := func() (*httpapi.Server, string, chan error, *agents.Registry, []domain.LocalAgent) {
		// 仅注册真实协议实现；未 Probe 的配置不能进入付费处理路径。
		registry := agents.NewRegistry(nativeRoot)
		if _, err := registry.Adapter(project.Settings.ExternalModelCLI); err != nil {
			t.Fatal("original CLI has no actual protocol adapter", err)
		}
		observations, err := registry.SnapshotNative(ctx)
		if err != nil {
			t.Fatal(err)
		}
		projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
		progress := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, registry, registry)
		server := httpapi.NewServer(httpapi.Config{Logger: logger, Services: httpapi.Services{LocalProjects: projects, Progress: progress}})
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		return server, "http://" + listener.Addr().String(), done, registry, observations
	}
	closed := map[*httpapi.Server]bool{}
	stop := func(server *httpapi.Server, done <-chan error) {
		if closed[server] {
			return
		}
		closed[server] = true
		shutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			t.Fatal(err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("actual receipt server did not shut down normally", err)
		}
	}
	bootstrap := func(origin string) (*http.Client, string) {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
		t.Cleanup(client.CloseIdleConnections)
		response, err := client.Get(origin + "/api/v1/auth/session")
		if err != nil {
			t.Fatal(err)
		}
		var session struct {
			Kind string `json:"actor_kind"`
			CSRF string `json:"csrf_token"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&session)
		if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK || session.Kind != "human" || session.CSRF == "" {
			t.Fatal("actual human bootstrap failed", err)
		}
		return client, session.CSRF
	}
	request := func(client *http.Client, csrf, origin, method, suffix, key string, body any) domain.ProjectProgress {
		var data []byte
		if body != nil {
			var err error
			data, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, origin+"/api/v1/local-projects/"+original.ProjectID+"/progress"+suffix, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if method != http.MethodGet {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrf)
			req.Header.Set("Idempotency-Key", key)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Progress json.RawMessage `json:"progress"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&envelope)
		if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK {
			t.Fatal("real receipt HTTP operation failed", method, response.StatusCode, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(envelope.Progress, &fields); err != nil {
			t.Fatal(err)
		}
		if _, exists := fields["operations"]; exists {
			t.Fatal("HTTP disclosed private operation receipts")
		}
		if _, exists := fields["pending_operation"]; exists {
			t.Fatal("HTTP disclosed the private pending operation")
		}
		var result domain.ProjectProgress
		if err := json.Unmarshal(envelope.Progress, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	assertWire := func(actual, expected domain.ProjectProgress) {
		expected.SchemaVersion = 0
		expected.Operations = nil
		expected.PendingOperation = ""
		expected.ObservedAt = expected.ObservedAt.UTC().Truncate(time.Second)
		expected.Evidence = append([]domain.ProgressEvidence{}, expected.Evidence...)
		if !reflect.DeepEqual(actual, expected) {
			t.Fatal("HTTP public fields differ from the actual original result or human write")
		}
	}
	assertStored := func(expected domain.ProjectProgress) {
		actual, err := db.LoadProjectProgress(ctx, original.ProjectID)
		if err != nil || !reflect.DeepEqual(actual, expected) || !reflect.DeepEqual(actual.Operations, original.Operations) || actual.PendingOperation != original.PendingOperation || actual.Operations[operationID].Status != "accepted" || actual.Operations[operationID].Result == nil {
			t.Fatal("human write or accepted replay changed the actual stored receipt", err)
		}
		var version int
		var storedProject string
		if err := db.conn.QueryRow("SELECT version,data FROM local_agent_projects WHERE id=?", original.ProjectID).Scan(&version, &storedProject); err != nil || version != projectVersion || storedProject != projectJSON {
			t.Fatal("original project metadata changed", err)
		}
		for table, expectedCount := range map[string]int{"local_agent_projects": 1, "local_project_progress": 1, "local_agent_grants": 0, "local_agent_sessions": 0} {
			var count int
			if err := db.conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != expectedCount {
				t.Fatal("unexpected actual target row count", table, err)
			}
		}
	}
	assertUnusedNative := func(registry *agents.Registry, before []domain.LocalAgent) {
		after, err := registry.SnapshotNative(ctx)
		if err != nil || !reflect.DeepEqual(after, before) {
			t.Fatal("accepted replay changed native observations", err)
		}
		entries, err := os.ReadDir(nativeRoot)
		if err != nil || len(entries) != 0 {
			t.Fatal("accepted replay created native state", err)
		}
	}
	first, firstURL, firstDone, firstRegistry, firstNative := start()
	defer stop(first, firstDone)
	client, csrf := bootstrap(firstURL)
	assertWire(request(client, csrf, firstURL, http.MethodGet, "", "", nil), projection)
	assertStored(projection)
	request(client, csrf, firstURL, http.MethodPut, "", uuid.NewString(), map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 1, "status": "done"})
	human, err := db.LoadProjectProgress(ctx, original.ProjectID)
	if err != nil || human.Source != "human" || human.Status != "done" || human.Revision != 2 {
		t.Fatal("real human PUT lost its own stage", err)
	}
	assertStored(human)
	assertWire(request(client, csrf, firstURL, http.MethodGet, "", "", nil), human)
	expectedReplay := projection
	expectedReplay.Revision = human.Revision
	expectedReplay.Warning = human.Warning
	inferBody := map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": human.Revision, "files": []string{receipt.Result.Evidence[0].SourcePath}}
	assertWire(request(client, csrf, firstURL, http.MethodPost, "/infer", operationID, inferBody), expectedReplay)
	assertStored(human)
	assertUnusedNative(firstRegistry, firstNative)
	stop(first, firstDone)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = nil
	db, err = Open(targetDB)
	if err != nil {
		t.Fatal(err)
	}
	second, secondURL, secondDone, secondRegistry, secondNative := start()
	defer stop(second, secondDone)
	client, csrf = bootstrap(secondURL)
	assertStored(human)
	assertWire(request(client, csrf, secondURL, http.MethodGet, "", "", nil), human)
	inferBody["request_id"] = uuid.NewString()
	assertWire(request(client, csrf, secondURL, http.MethodPost, "/infer", operationID, inferBody), expectedReplay)
	assertStored(human)
	assertUnusedNative(secondRegistry, secondNative)
	stop(second, secondDone)
	current, currentProject, currentVersion := readPublicProgressFrom(t, source)
	var currentJSON string
	var currentRevision int
	if err := source.conn.QueryRow("SELECT revision,data FROM local_project_progress WHERE project_id=?", original.ProjectID).Scan(&currentRevision, &currentJSON); err != nil || currentRevision != sourceProgressRevision || currentJSON != sourceProgressJSON || !reflect.DeepEqual(current, original) || currentProject != projectJSON || currentVersion != projectVersion {
		t.Fatal("original source rows changed", err)
	}
	progressReceiptJSON(t, root, "source-before-close.json", progressReceiptCompareFiles(t, filesBefore, progressReceiptFiles(t, publication)))
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	sourceClosed = true
	progressReceiptJSON(t, root, "source-after-close.json", map[string]any{"observed_utc": time.Now().UTC(), "metadata_only": true, "source_sqlite_closed": true, "source_sqlite_reopened": false, "files": progressReceiptCompareFiles(t, filesBefore, progressReceiptFiles(t, publication))})
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = nil
	progressReceiptJSON(t, root, "result.json", map[string]any{"source_db": publication, "target_db": targetDB, "project_id": original.ProjectID, "native_id": original.NativeID, "model": original.Model, "source_revision": original.Revision, "target_initial_cas_revision": 1, "target_human_revision": human.Revision, "original_operation_id": operationID, "original_operation_count": len(original.Operations), "operations_unchanged": true, "pending_operation_unchanged": true, "human_storage_preserved_after_replay": true, "same_db_restart": true, "http1_url": firstURL, "http2_url": secondURL, "both_http_closed": true, "source_rows_unchanged_before_close": true, "source_main_wal_unchanged_after_close": true, "shm_metadata_recorded_separately": true, "registry_observations_unchanged": true, "native_state_entries": 0, "new_model_turns": 0, "new_native_sessions": 0, "grant_count": 0})
	t.Log("real HTTP human write preserved the original receipt; original-key replay and same-database restart passed without native execution")
}
