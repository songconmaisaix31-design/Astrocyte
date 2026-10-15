//go:build video_live

package importers_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func videoRequiredEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Fatal("actual video verification requires explicit configuration", key)
	}
	return value
}

// 包装只保存实际失败原件，所有读取始终交生产 Reader。
type videoDiagnosticReader struct {
	*importers.Reader
	directory string
}

func (r videoDiagnosticReader) ReadSource(ctx context.Context, command app.ImportMaterialCommand) (app.ImportedSource, error) {
	source, err := r.Reader.ReadSource(ctx, command)
	var failure *importers.ExtractionFailure
	if errors.As(err, &failure) {
		for name, data := range map[string][]byte{"first-failure-cli.json": failure.CLIJSON, "first-failure-cli.stderr.txt": failure.CLIStderr, "first-failure-media.json": failure.MediaJSON} {
			file, writeErr := os.OpenFile(filepath.Join(r.directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if writeErr != nil {
				return source, errors.Join(err, writeErr)
			}
			_, writeErr = file.Write(data)
			closeErr := file.Close()
			if writeErr != nil || closeErr != nil {
				return source, errors.Join(err, writeErr, closeErr)
			}
		}
	}
	return source, err
}

func TestVideoProductionHTTPRefreshTimelineRestart(t *testing.T) {
	ctx := context.Background()
	locator := "https://www.bilibili.com/video/BV1PReT6EEqR/"
	root := videoRequiredEnv(t, "ASTROCYTE_TEST_EXPORT_DIR")
	oldFile := videoRequiredEnv(t, "ASTROCYTE_TEST_VIDEO_ORIGINAL_TEXT")
	oldText, err := os.ReadFile(oldFile)
	if err != nil || len(oldText) != 32908 {
		t.Fatal("approved prior original unavailable or changed", err)
	}
	if _, err := os.Stat(filepath.Join(root, "state.sqlite")); !os.IsNotExist(err) {
		t.Fatal("preserve prior evidence; provide a fresh verification directory")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	toolRoot := filepath.Join(root, "tool-temp")
	if err := os.MkdirAll(toolRoot, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEMP", toolRoot)
	t.Setenv("TMP", toolRoot)
	extractor, err := importers.NewSummarizeExtractorWithOptions(
		videoRequiredEnv(t, "ASTROCYTE_TEST_SUMMARIZE_NODE"), videoRequiredEnv(t, "ASTROCYTE_TEST_SUMMARIZE_CLI"),
		importers.SummarizeOptions{
			YtDlpPath: videoRequiredEnv(t, "ASTROCYTE_TEST_YT_DLP"), FFmpegPath: videoRequiredEnv(t, "ASTROCYTE_TEST_FFMPEG"),
			WhisperBinary: videoRequiredEnv(t, "ASTROCYTE_TEST_WHISPER"), WhisperModel: videoRequiredEnv(t, "ASTROCYTE_TEST_WHISPER_MODEL"),
			ProxyURL: os.Getenv("ASTROCYTE_TEST_SUMMARIZE_PROXY_URL"), Timeout: 30 * time.Minute,
		})
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader([]string{filepath.Dir(oldFile)})
	reader.Summarize = extractor
	var db *sqlite.DB
	var service *app.Service
	var store *objects.Store
	var server *httptest.Server
	var client *http.Client
	var csrf string
	start := func() {
		var err error
		db, err = sqlite.Open(filepath.Join(root, "state.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		if err = db.RunMigrations(ctx, migrations.FS); err != nil {
			t.Fatal(err)
		}
		store, err = objects.New(filepath.Join(root, "objects"))
		if err != nil {
			t.Fatal(err)
		}
		service = app.NewAttentionService(sqlite.NewAttentionRepository(db), videoDiagnosticReader{Reader: reader, directory: root}, store, app.ServiceOptions{JobTimeout: 30 * time.Minute, MaxAttempts: 1})
		server = httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Attention: service}}).Handler())
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client = &http.Client{Jar: jar, Timeout: time.Minute}
	}
	stop := func() {
		server.Close()
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
	request := func(method, path string, body any, result any, expected int) {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r, err := http.NewRequestWithContext(ctx, method, server.URL+path, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", rand.Text())
		r.Header.Set("X-CSRF-Token", csrf)
		response, err := client.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != expected {
			t.Fatal("production HTTP status", method, path, response.StatusCode)
		}
		if err := json.NewDecoder(response.Body).Decode(result); err != nil {
			t.Fatal(err)
		}
	}
	bootstrap := func() {
		var session struct {
			CSRF string `json:"csrf_token"`
		}
		request("GET", "/api/v1/auth/session", nil, &session, 200)
		csrf = session.CSRF
		if csrf == "" {
			t.Fatal("actual session bootstrap missing CSRF")
		}
	}
	start()
	t.Cleanup(func() {
		if server != nil {
			server.Close()
		}
		if db != nil {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	bootstrap()
	importCommand := func(adapter, file string, refresh bool) app.ImportMaterialCommand {
		return app.ImportMaterialCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: rand.Text(), ExpectedVersion: 1}, Adapter: adapter, Kind: "video", SourceLocator: locator, LocalFileRef: file, Refresh: refresh}
	}
	process := func(receipt app.ImportJobResult) app.Job {
		worked, err := service.ProcessNextJob(ctx)
		if err != nil || !worked {
			t.Fatal("actual production import worker", worked, err)
		}
		var job app.Job
		request("GET", "/api/v1/jobs/"+receipt.JobID, nil, &job, 200)
		data, err := json.MarshalIndent(job, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, receipt.JobID+"-job.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		if job.Status != "succeeded" || job.MaterialID == nil || job.MaterialRevision == nil {
			t.Fatal("actual import did not succeed; local job evidence retained", job.Status, job.JobID)
		}
		return job
	}
	var oldReceipt app.ImportJobResult
	request("POST", "/api/v1/materials/imports", importCommand("summarize_markdown", oldFile, false), &oldReceipt, 202)
	oldJob := process(oldReceipt)
	var timedReceipt app.ImportJobResult
	request("POST", "/api/v1/materials/imports", importCommand("summarize_url", "", true), &timedReceipt, 202)
	t.Log("actual public timed extraction queued; native originals retained locally")
	timedJob := process(timedReceipt)
	var material app.MaterialDetail
	request("GET", "/api/v1/materials/"+*timedJob.MaterialID, nil, &material, 200)
	if *oldJob.MaterialID != *timedJob.MaterialID || len(material.Revisions) != 2 || material.Material.CurrentRevision != 2 {
		t.Fatal("actual prior and timed exports did not produce same-source changed revisions", len(material.Revisions), material.Material.CurrentRevision)
	}
	prior, timed := material.Revisions[0], material.Revisions[1]
	if prior.SourceKey != timed.SourceKey || timed.SourceLocator != locator || timed.Provenance.Source != locator || !strings.Contains(timed.Provenance.Version, "0.25.1") || len(timed.SourceSpans) == 0 || len(prior.SourceSpans) != 0 {
		t.Fatal("source/version/timeline evidence missing")
	}
	actualText, err := store.Read(ctx, timed.ObjectRef)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(actualText, oldText) {
		t.Fatal("actual timed export content unchanged; changed-version assertion not satisfied")
	}
	digest := sha256.Sum256(actualText)
	if timed.ContentDigest != hex.EncodeToString(digest[:]) {
		t.Fatal("actual content digest mismatch")
	}
	attachmentBytes := map[string][]byte{}
	for _, attachment := range timed.Attachments {
		data, err := store.Read(ctx, attachment.ObjectRef)
		if err != nil {
			t.Fatal(err)
		}
		attachmentBytes[attachment.Name] = data
		if err := os.WriteFile(filepath.Join(root, filepath.Base(attachment.Name)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if len(attachmentBytes["whisper-original.json"]) == 0 {
		t.Fatal("native whisper.cpp JSON original missing")
	}
	var native struct {
		Transcription []struct {
			Offsets struct{ From, To int64 }
			Text    string
		}
	}
	if err := json.Unmarshal(attachmentBytes["whisper-original.json"], &native); err != nil {
		t.Fatal(err)
	}
	if len(native.Transcription) != len(timed.SourceSpans) {
		t.Fatal("actual native time segment count differs from source spans")
	}
	if err := os.WriteFile(filepath.Join(root, "video-original-text.txt"), actualText, 0600); err != nil {
		t.Fatal(err)
	}
	duplicateCommand := importCommand("summarize_url", "", false)
	duplicateCommand.SourceLocator += "?spm_id_from=333.337.search-card.all.click"
	var duplicate app.ImportJobResult
	request("POST", "/api/v1/materials/imports", duplicateCommand, &duplicate, 202)
	if duplicate.JobID != timedReceipt.JobID {
		t.Fatal("tracking alias without refresh repeated actual extraction")
	}
	stop()
	start()
	bootstrap()
	var after app.MaterialDetail
	request("GET", "/api/v1/materials/"+*timedJob.MaterialID, nil, &after, 200)
	if !reflect.DeepEqual(after, material) {
		t.Fatal("HTTP/SQLite restart changed actual material revisions or timeline")
	}
	for _, revision := range after.Revisions {
		data, err := store.Read(ctx, revision.ObjectRef)
		if err != nil || (revision.Revision == 1 && !bytes.Equal(data, oldText)) || (revision.Revision == 2 && !bytes.Equal(data, actualText)) {
			t.Fatal("restart changed original content object", revision.Revision, err)
		}
		for _, attachment := range revision.Attachments {
			data, err := store.Read(ctx, attachment.ObjectRef)
			if err != nil || (revision.Revision == 2 && !bytes.Equal(data, attachmentBytes[attachment.Name])) {
				t.Fatal("restart changed original attachment", attachment.Name, err)
			}
		}
	}
	for table, expected := range map[string]int{"attention_materials": 1, "attention_material_revisions": 2, "attention_jobs": 2, "attention_distillations": 0, "attention_opportunities": 0} {
		var count int
		if err := db.Conn().QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != expected {
			t.Fatal("actual persistent row count", table, count, err)
		}
	}
	var missionEvents int
	if err := db.Conn().QueryRow("SELECT count(*) FROM attention_outbox WHERE type LIKE '%mission%'").Scan(&missionEvents); err != nil || missionEvents != 0 {
		t.Fatal("video import created Mission event", err)
	}
	if len(after.Distillations) != 0 || len(after.Uses) != 0 || after.Material.HumanUsageCount != 0 || after.Material.AgentUsageCount != 0 {
		t.Fatal("video import created Agent distillation or usage")
	}
	t.Logf("actual HTTP/SQLite source refresh/restart: revisions=%d old_bytes=%d timed_bytes=%d spans=%d attachments=%d no_agent_model_calls=true asr_model=ggml-base-multilingual no_mission=true evidence=%s", len(after.Revisions), len(oldText), len(actualText), len(timed.SourceSpans), len(timed.Attachments), root)
}
