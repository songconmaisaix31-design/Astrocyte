package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// Explicit live acceptance uses the real Reader and original Service job with
// SQLite/object persistence. No Distiller is configured, so no model call runs.
func TestSummarizeVideoLiveServiceRestart(t *testing.T) {
	locator := os.Getenv("ASTROCYTE_TEST_VIDEO_URL")
	if locator == "" {
		t.Skip("explicit approved public video opt-in required")
	}
	root := os.Getenv("ASTROCYTE_TEST_EXPORT_DIR")
	if root == "" {
		root = t.TempDir()
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	extractor, err := importers.NewSummarizeExtractorWithOptions(os.Getenv("ASTROCYTE_TEST_SUMMARIZE_NODE"), os.Getenv("ASTROCYTE_TEST_SUMMARIZE_CLI"), importers.SummarizeOptions{YtDlpPath: os.Getenv("ASTROCYTE_TEST_YT_DLP"), FFmpegPath: os.Getenv("ASTROCYTE_TEST_FFMPEG"), WhisperBinary: os.Getenv("ASTROCYTE_TEST_WHISPER"), WhisperModel: os.Getenv("ASTROCYTE_TEST_WHISPER_MODEL"), Timeout: 30 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	reader.Summarize = extractor
	ctx := context.Background()
	// Save first failure diagnostics in the selected external test export directory.
	readerPort := &liveDiagnosticReader{Reader: reader, Dir: root}
	dbPath := filepath.Join(root, "state.sqlite")
	db := openAttentionDB(t, dbPath)
	store, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	service := app.NewAttentionService(NewAttentionRepository(db), readerPort, store, app.ServiceOptions{JobTimeout: 30 * time.Minute, MaxAttempts: 1})
	human := app.Principal{ID: "live-public-video", Kind: "human"}
	cmd := app.ImportMaterialCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "live-video-import", IdempotencyKey: "live-video-import", ExpectedVersion: 1}, Kind: "video", Adapter: "summarize_url", SourceLocator: locator}
	receipt, err := service.ImportMaterial(ctx, human, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	job, err := service.GetJob(ctx, human, receipt.JobID)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(job, "", "  ")
	os.WriteFile(filepath.Join(root, "video-job.json"), data, 0600)
	if job.Status != "succeeded" || job.MaterialID == nil {
		t.Fatalf("actual live job=%+v", job)
	}
	material, err := service.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	content, err := service.GetContent(ctx, human, *job.MaterialID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(content.Text) < 1000 || !strings.Contains(content.Provenance.Version, "0.25.1") || len(material.Distillations) != 0 {
		t.Fatal("missing original content/version or accidental model output")
	}
	rev := material.Revisions[0]
	for _, a := range rev.Attachments {
		got, err := service.GetAttachment(ctx, human, *job.MaterialID, 1, a.Name)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, filepath.Base(a.Name)), got.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(root, "video-original-text.txt"), []byte(content.Text), 0600)
	data, _ = json.MarshalIndent(material, "", "  ")
	os.WriteFile(filepath.Join(root, "video-material.json"), data, 0600)
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	service = app.NewAttentionService(NewAttentionRepository(db), readerPort, store, app.ServiceOptions{})
	after, err := service.GetContent(ctx, human, *job.MaterialID, 1)
	if err != nil || after.Text != content.Text || after.Provenance != content.Provenance {
		t.Fatal("restart lost original", err)
	}
	// The same submitted command replays its existing job; a fresh command for
	// an unversioned remote URL deliberately refreshes under the existing Service.
	duplicate, err := service.ImportMaterial(ctx, human, cmd)
	if err != nil || duplicate.JobID != receipt.JobID {
		t.Fatalf("duplicate source reran downloader: %+v %v", duplicate, err)
	}
	t.Logf("source=%s bytes=%d mode=%s attachments=%d spans=%d duplicate_job=%s no_distillations=true artifacts=%s", locator, len(content.Text), content.Provenance.Mode, len(rev.Attachments), len(rev.SourceSpans), duplicate.JobID, root)
}

type liveDiagnosticReader struct {
	Reader *importers.Reader
	Dir    string
}

func (r *liveDiagnosticReader) ReadSource(ctx context.Context, c app.ImportMaterialCommand) (app.ImportedSource, error) {
	s, err := r.Reader.ReadSource(ctx, c)
	var failed *importers.ExtractionFailure
	if errors.As(err, &failed) {
		os.WriteFile(filepath.Join(r.Dir, "first-failure-cli.json"), failed.CLIJSON, 0600)
		os.WriteFile(filepath.Join(r.Dir, "first-failure-cli.stderr.txt"), failed.CLIStderr, 0600)
		os.WriteFile(filepath.Join(r.Dir, "first-failure-media.json"), failed.MediaJSON, 0600)
	}
	return s, err
}
