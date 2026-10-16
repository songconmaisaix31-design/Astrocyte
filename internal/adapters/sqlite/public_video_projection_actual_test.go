//go:build public_video_live

package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func publicVideoPath(t *testing.T, path string, directory bool) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Fatal("explicit absolute path required")
	}
	if err := workbenchPathUnredirected(path); err != nil {
		t.Fatal("redirected video path rejected", err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || !strings.EqualFold(filepath.Clean(path), filepath.Clean(canonical)) {
		t.Fatal("noncanonical video path rejected", err)
	}
	info, err := os.Lstat(path)
	if err != nil || (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		t.Fatal("ordinary video path required", err)
	}
}

func publicVideoJSON(t *testing.T, root, name string, value any) {
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

func publicVideoFiles(t *testing.T, database string) map[string]os.FileInfo {
	t.Helper()
	result := map[string]os.FileInfo{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		path := database + suffix
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) && suffix != "" {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		publicVideoPath(t, path, false)
		result[suffix] = info
	}
	return result
}

func publicVideoFileFacts(before, after map[string]os.FileInfo) map[string]any {
	result := map[string]any{}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		original, oldExists := before[suffix]
		current, exists := after[suffix]
		facts := map[string]any{"before_exists": oldExists, "after_exists": exists}
		if oldExists {
			facts["before_bytes"] = original.Size()
			facts["before_mtime"] = original.ModTime().UTC().Format(time.RFC3339Nano)
		}
		if exists {
			facts["after_bytes"] = current.Size()
			facts["after_mtime"] = current.ModTime().UTC().Format(time.RFC3339Nano)
		}
		if oldExists && exists {
			facts["same_identity"] = os.SameFile(original, current)
			facts["unchanged"] = os.SameFile(original, current) && original.Size() == current.Size() && original.ModTime().Equal(current.ModTime())
		}
		result[suffix] = facts
	}
	return result
}

type publicVideoRows struct {
	version         int
	key             string
	material        string
	revisions       []string
	revisionNumbers []int
	contentDigests  []string
}

func publicVideoReadRows(t *testing.T, source *sql.DB, id string) publicVideoRows {
	t.Helper()
	var result publicVideoRows
	if err := source.QueryRow("SELECT version,source_key,data FROM attention_materials WHERE id=?", id).Scan(&result.version, &result.key, &result.material); err != nil {
		t.Fatal("approved material row unavailable", err)
	}
	rows, err := source.Query("SELECT revision,content_digest,data FROM attention_material_revisions WHERE material_id=? ORDER BY revision", id)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var data string
		var digest string
		var revision int
		if err := rows.Scan(&revision, &digest, &data); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				t.Error(closeErr)
			}
			t.Fatal(err)
		}
		result.revisions = append(result.revisions, data)
		result.revisionNumbers = append(result.revisionNumbers, revision)
		result.contentDigests = append(result.contentDigests, digest)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		t.Fatal(err)
	}
	return result
}

// 只派生已获准的实际视频结果，生产存储和 HTTP 不接提取器或原生处理器。
func TestServePublicVideoProjectionActual(t *testing.T) {
	ctx := context.Background()
	owned := os.Getenv("ASTROCYTE_TEST_VIDEO_OWNED_ROOT")
	evidence := os.Getenv("ASTROCYTE_TEST_VIDEO_EVIDENCE_DIR")
	target := os.Getenv("ASTROCYTE_TEST_VIDEO_TARGET_DIR")
	publication := os.Getenv("ASTROCYTE_TEST_VIDEO_SOURCE_DB")
	id := os.Getenv("ASTROCYTE_TEST_VIDEO_MATERIAL_ID")
	releaseID := os.Getenv("ASTROCYTE_TEST_VIDEO_RELEASE_ID")
	webDir := os.Getenv("ASTROCYTE_WEB_DIR")
	for _, directory := range []string{owned, evidence, webDir} {
		publicVideoPath(t, directory, true)
	}
	publicVideoPath(t, filepath.Join(webDir, "index.html"), false)
	publicVideoPath(t, publication, false)
	if id != "534HWFW7E5GEF75ZY4F7ETPLQ2" || releaseID == "" || strings.TrimSpace(releaseID) != releaseID {
		t.Fatal("fixed approved material and explicit release identity required")
	}
	if !filepath.IsAbs(target) || !strings.EqualFold(filepath.Dir(filepath.Clean(target)), filepath.Clean(owned)) || !strings.EqualFold(filepath.Dir(filepath.Clean(evidence)), filepath.Clean(owned)) || strings.EqualFold(target, evidence) {
		t.Fatal("new target and separate evidence directory must be direct owned children")
	}
	if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("target already exists or cannot be inspected")
	}
	for _, pair := range [][2]string{{owned, filepath.Dir(publication)}, {filepath.Dir(publication), owned}} {
		relative, err := filepath.Rel(pair[0], pair[1])
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			t.Fatal("source and owned target roots must be separate")
		}
	}
	filesBefore := publicVideoFiles(t, publication)
	locator := &url.URL{Scheme: "file", Path: filepath.ToSlash(publication)}
	if filepath.VolumeName(publication) != "" {
		locator.Path = "/" + locator.Path
	}
	query := locator.Query()
	query.Set("mode", "ro")
	query.Set("_pragma", "query_only(ON)")
	locator.RawQuery = query.Encode()
	source, err := sql.Open("sqlite", locator.String())
	if err != nil {
		t.Fatal(err)
	}
	source.SetMaxOpenConns(1)
	defer func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	}()
	var queryOnly int
	var locking string
	if err := source.QueryRow("PRAGMA query_only").Scan(&queryOnly); err != nil || queryOnly != 1 {
		t.Fatal("source must enforce query_only", err)
	}
	if err := source.QueryRow("PRAGMA locking_mode").Scan(&locking); err != nil || locking != "normal" {
		t.Fatal("source must use normal SQLite locks", err)
	}
	originalRows := publicVideoReadRows(t, source, id)
	var original attention.MaterialDetail
	if err := json.Unmarshal([]byte(originalRows.material), &original); err != nil {
		t.Fatal(err)
	}
	const publicLocator = "https://www.bilibili.com/video/BV1PReT6EEqR/"
	if originalRows.version != 2 || original.Material.Version != 2 || original.Material.ID != id || original.Material.Kind != "video" || original.Material.SourceLocator != publicLocator || original.Material.CurrentRevision != 2 || len(original.Material.SourceSpans) != 952 || len(original.Revisions) != 2 || len(originalRows.revisions) != 2 {
		t.Fatal("original approved two-revision video provenance unavailable")
	}
	private := map[string]bool{}
	for _, revision := range original.Revisions {
		for _, attachment := range revision.Attachments {
			if attachment.MediaType == attention.NativePrivateAttachmentMediaType || attachment.Name == "whisper-original.json" || attachment.Name == "summarize-pdf-original.json" {
				private[attachment.ObjectRef] = true
			}
		}
	}
	refs := map[string]bool{}
	for index, revision := range original.Revisions {
		var stored attention.MaterialRevision
		if err := json.Unmarshal([]byte(originalRows.revisions[index]), &stored); err != nil || !reflect.DeepEqual(stored, revision) {
			t.Fatal("immutable revision row differs from aggregate", err)
		}
		if revision.MaterialID != id || revision.Revision != index+1 || originalRows.revisionNumbers[index] != revision.Revision || originalRows.contentDigests[index] != revision.ContentDigest || revision.SourceKey != originalRows.key || revision.SourceLocator != publicLocator || revision.ContentDigest != revision.ObjectRef || len(revision.SourceSpans) != []int{0, 952}[index] || private[revision.ObjectRef] {
			t.Fatal("original revision identity, digest or positions differ")
		}
		for _, span := range revision.SourceSpans {
			if !strings.HasPrefix(span, publicLocator) {
				t.Fatal("original source position locator differs")
			}
		}
		refs[revision.ObjectRef] = true
		for _, attachment := range revision.Attachments {
			refs[attachment.ObjectRef] = true
		}
	}
	sourceObjects := filepath.Join(filepath.Dir(publication), "objects")
	publicVideoPath(t, sourceObjects, true)
	publicBytes := map[string][]byte{}
	objectFiles := map[string]os.FileInfo{}
	privateFacts := []map[string]any{}
	for ref := range refs {
		decoded, err := hex.DecodeString(ref)
		if err != nil || len(decoded) != sha256.Size || ref != strings.ToLower(ref) {
			t.Fatal("invalid immutable object reference")
		}
		path := filepath.Join(sourceObjects, ref)
		publicVideoPath(t, path, false)
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		if private[ref] {
			privateFacts = append(privateFacts, map[string]any{"object_ref": ref, "source_pointer": path, "exists": true, "regular_file": true, "bytes": info.Size(), "body_read": false, "copied": false})
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil || int64(len(data)) != info.Size() || fmt.Sprintf("%x", sha256.Sum256(data)) != ref {
			t.Fatal("public original object bytes or digest differ", err)
		}
		publicBytes[ref] = data
		objectFiles[ref] = info
	}
	if len(publicBytes) != 5 || len(privateFacts) != 1 {
		t.Fatal("approved public/private object counts differ")
	}
	verifySource := func(name string) {
		if !reflect.DeepEqual(publicVideoReadRows(t, source, id), originalRows) {
			t.Fatal("original material or revision rows changed")
		}
		after := publicVideoFiles(t, publication)
		publicVideoJSON(t, evidence, name, map[string]any{"rows_unchanged": true, "source_mode": "ro", "query_only": true, "normal_sqlite_locks": true, "sqlite_files": publicVideoFileFacts(filesBefore, after), "shm_metadata_recorded_separately": true})
		for _, suffix := range []string{"", "-wal"} {
			before, beforeExists := filesBefore[suffix]
			current, exists := after[suffix]
			if beforeExists != exists || (exists && (!os.SameFile(before, current) || before.Size() != current.Size() || !before.ModTime().Equal(current.ModTime()))) {
				t.Fatal("source main or WAL identity, size or mtime changed")
			}
		}
		for ref, before := range objectFiles {
			path := filepath.Join(sourceObjects, ref)
			publicVideoPath(t, path, false)
			current, err := os.Lstat(path)
			if err != nil || !os.SameFile(before, current) || before.Size() != current.Size() || !before.ModTime().Equal(current.ModTime()) {
				t.Fatal("source public object metadata changed", err)
			}
			data, err := os.ReadFile(path)
			if err != nil || !reflect.DeepEqual(data, publicBytes[ref]) || fmt.Sprintf("%x", sha256.Sum256(data)) != ref {
				t.Fatal("source public object bytes changed", err)
			}
		}
	}
	sourceVerifiedAtExit := false
	defer func() {
		if !sourceVerifiedAtExit {
			verifySource("source-early-exit.json")
		}
	}()
	verifySource("source-before-target.json")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	publicVideoPath(t, target, true)
	dataDir := filepath.Join(target, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	targetDB := filepath.Join(dataDir, "state.sqlite")
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
	objectStore, err := objects.New(filepath.Join(dataDir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	publicFacts := []map[string]any{}
	for ref, data := range publicBytes {
		published, err := objectStore.Publish(ctx, data)
		if err != nil || published != ref {
			t.Fatal("public object projection changed original digest", err)
		}
		publicFacts = append(publicFacts, map[string]any{"object_ref": ref, "bytes": len(data), "source_pointer": filepath.Join(sourceObjects, ref), "target_pointer": filepath.Join(dataDir, "objects", ref), "actual_digest_matches": true})
	}
	projection := attention.MaterialDetail{SchemaVersion: original.SchemaVersion, Material: original.Material, Revisions: original.Revisions, Distillations: []attention.Distillation{}, Uses: []attention.Usage{}}
	projection.Material.Version = 1
	if err := NewAttentionRepository(db).WithTx(ctx, func(tx attention.AttentionTx) error { return tx.SaveMaterial(projection, 0) }); err != nil {
		t.Fatal(err)
	}
	verifyStored := func() {
		if err := NewAttentionRepository(db).WithTx(ctx, func(tx attention.AttentionTx) error {
			stored, err := tx.LoadMaterial(id)
			if err == nil && !reflect.DeepEqual(stored, projection) {
				return errors.New("target material fields or immutable revisions changed")
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
		for table, expected := range map[string]int{"attention_materials": 1, "attention_material_revisions": 2, "attention_jobs": 0, "attention_receipts": 0, "attention_outbox": 0, "attention_distillations": 0, "local_agent_grants": 0, "local_agent_sessions": 0} {
			var count int
			if err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != expected {
				t.Fatal("target row count differs", table, err)
			}
		}
		for ref, expected := range publicBytes {
			actual, err := objectStore.Read(ctx, ref)
			if err != nil || !reflect.DeepEqual(expected, actual) {
				t.Fatal("target public immutable bytes differ", err)
			}
		}
		for ref := range private {
			if _, err := os.Lstat(filepath.Join(dataDir, "objects", ref)); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("private object bytes must be absent from target")
			}
		}
	}
	verifyStored()
	logFile, err := os.OpenFile(filepath.Join(evidence, "http.log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
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
		schema, err := db.SchemaVersion(ctx)
		if err != nil {
			t.Fatal(err)
		}
		materials := attention.NewAttentionService(NewAttentionRepository(db), nil, objectStore, attention.ServiceOptions{})
		projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
		projects.ConfigureRegisteredDiscovery(db, nil)
		services := httpapi.Services{Foundation: foundation.NewAttentionService(schema), Attention: materials, Materials: materials, Opportunities: materials, Tracking: materials, RankingProfile: materials,
			LocalProjects: projects, RegisteredProjects: projects, Progress: workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, nil, nil),
			Projects: workspace.NewProjectService(), Proposals: workspace.NewProposalService(), Sessions: workspace.NewSessionService(),
			WorkspaceS2: workspace.NewS2Service(NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)}
		server := httpapi.NewServer(httpapi.Config{WebDir: webDir, Logger: logger, Services: services})
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		return server, "http://" + listener.Addr().String(), done
	}
	closed := map[*httpapi.Server]bool{}
	stopHTTP := func(server *httpapi.Server, done <-chan error) {
		if closed[server] {
			return
		}
		closed[server] = true
		shutdown, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			t.Fatal(err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("production server did not shut down normally", err)
		}
	}
	readHTTP := func(origin string) attention.MaterialDetail {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Jar: jar, Timeout: 10 * time.Second}
		defer client.CloseIdleConnections()
		get := func(path string) []byte {
			response, err := client.Get(origin + path)
			if err != nil {
				t.Fatal(err)
			}
			data, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if response.StatusCode != http.StatusOK || readErr != nil || closeErr != nil {
				t.Fatal("production human GET failed", path, response.StatusCode, readErr, closeErr)
			}
			return data
		}
		var session struct {
			ActorKind string `json:"actor_kind"`
			CSRF      string `json:"csrf_token"`
		}
		if err := json.Unmarshal(get("/api/v1/auth/session"), &session); err != nil || session.ActorKind != "human" || session.CSRF == "" {
			t.Fatal("production human bootstrap failed", err)
		}
		// 核查当前页面所用既有生产服务响应。
		for _, path := range []string{"/health", "/foundation", "/projects", "/proposals", "/sessions", "/project-spaces", "/goals", "/local-projects", "/local-projects/registered", "/materials", "/opportunities", "/distillations", "/jobs", "/material-domains", "/attention-ranking-profile", "/distillations/processor"} {
			get("/api/v1" + path)
		}
		var detail attention.MaterialDetail
		if err := json.Unmarshal(get("/api/v1/materials/"+id), &detail); err != nil || detail.Material.ID != id || detail.Material.Version != 1 || detail.Material.CurrentRevision != 2 || detail.Material.Title != projection.Material.Title || !reflect.DeepEqual(detail.Material.SourceSpans, projection.Material.SourceSpans) || !reflect.DeepEqual(detail.Revisions, projection.Revisions) {
			t.Fatal("HTTP material differs from original identity and immutable revisions", err)
		}
		for _, revision := range detail.Revisions {
			var content attention.ContentResult
			path := fmt.Sprintf("/api/v1/materials/%s/revisions/%d", id, revision.Revision)
			if err := json.Unmarshal(get(path+"/content"), &content); err != nil || content.ContentDigest != revision.ContentDigest || content.Text != string(publicBytes[revision.ObjectRef]) {
				t.Fatal("HTTP content differs from actual immutable bytes", err)
			}
			for _, attachment := range revision.Attachments {
				if private[attachment.ObjectRef] {
					continue
				}
				data := get(path + "/attachments/" + url.PathEscape(attachment.Name))
				if !reflect.DeepEqual(data, publicBytes[attachment.ObjectRef]) || fmt.Sprintf("%x", sha256.Sum256(data)) != attachment.ObjectRef {
					t.Fatal("HTTP public attachment differs from original immutable bytes")
				}
			}
		}
		response, err := client.Get(origin + "/attention")
		if err != nil {
			t.Fatal(err)
		}
		_, readErr := io.Copy(io.Discard, response.Body)
		closeErr := response.Body.Close()
		if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") || readErr != nil || closeErr != nil {
			t.Fatal("actual frontend static page unavailable", response.StatusCode, readErr, closeErr)
		}
		return detail
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
	verifySource("source-after-http1.json")
	db, err = Open(targetDB)
	if err != nil {
		t.Fatal(err)
	}
	verifyStored()
	second, secondURL, secondDone := startHTTP()
	defer stopHTTP(second, secondDone)
	if !reflect.DeepEqual(firstResult, readHTTP(secondURL)) {
		t.Fatal("same-database HTTP2 material differs")
	}
	verifyStored()
	verifySource("source-ready-http2.json")
	publicVideoJSON(t, evidence, "ready.json", map[string]any{"pid": os.Getpid(), "release_id": releaseID, "source_db": publication, "target_db": targetDB, "material_id": id, "source_aggregate_version": originalRows.version, "target_cas_version": 1,
		"source_rows_unchanged": true, "source_main_wal_unchanged": true, "shm_metadata_recorded_separately": true, "source_read_only": true, "query_only": true,
		"revision_counts": []map[string]int{{"revision": 1, "source_spans": 0}, {"revision": 2, "source_spans": 952}}, "public_objects": publicFacts, "private_objects": privateFacts,
		"http1_url": firstURL, "http1_closed": true, "http2_url": secondURL, "same_target_db": true, "frontend_dir": webDir,
		"grant_count": 0, "native_session_count": 0, "new_asr_calls": 0, "new_extractions": 0, "new_model_turns": 0, "private_object_bodies_read": 0, "private_objects_copied": 0})
	t.Log("HTTP1 closed; same-target HTTP2 ready; safe pointers and counts are in ready.json")
	deadline := time.NewTimer(2 * time.Hour)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-secondDone:
			secondDone <- err
			t.Fatal("HTTP2 ended before the explicit read-only window release", err)
		case <-deadline.C:
			t.Fatal("read-only window release not received within two hours")
		case <-ticker.C:
			release, err := os.ReadFile(filepath.Join(evidence, "release.request"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil || strings.TrimSpace(string(release)) != releaseID {
				t.Fatal("invalid read-only window release", err)
			}
			verifyStored()
			verifySource("source-after-f-window.json")
			sourceVerifiedAtExit = true
			stopHTTP(second, secondDone)
			publicVideoJSON(t, evidence, "final.json", map[string]any{"source_rows_unchanged": true, "source_main_wal_unchanged": true, "shm_metadata_recorded_separately": true, "target_rows_unchanged": true, "public_object_bytes_unchanged": true, "read_only_window_released": true, "http2_closed": true, "private_object_bodies_read": 0, "private_objects_copied": 0, "new_asr_calls": 0, "new_extractions": 0, "new_model_turns": 0})
			return
		}
	}
}
