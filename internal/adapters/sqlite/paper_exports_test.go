package sqlite

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type paperExportTransport struct{}

func (paperExportTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := "%PDF-contract-local-original"
	if req.URL.Host == "export.arxiv.org" {
		body = `<feed><entry><id>https://arxiv.org/abs/2504.16054v1</id><title>Contract-local paper</title><summary>Contract-local metadata</summary></entry></feed>`
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: req}, nil
}

func TestOfficialPaperAndExportKeepOneMaterialAcrossSQLiteRestart(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state.sqlite")
	db := openAttentionDB(t, dbPath)
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	// Contract-local HTTP only; no real source/model/media operation runs.
	reader.Arxiv.Client = &http.Client{Transport: paperExportTransport{}}
	s := app.NewAttentionService(NewAttentionRepository(db), reader, store, app.ServiceOptions{})
	human := app.Principal{ID: "paper-export-test", Kind: "human"}
	command := func(key, adapter, locator, text string) app.ImportMaterialCommand {
		return app.ImportMaterialCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: key, IdempotencyKey: key, ExpectedVersion: 1}, Kind: "paper", Adapter: adapter, SourceLocator: locator, ExportText: text}
	}
	importJob := func(c app.ImportMaterialCommand) app.Job {
		t.Helper()
		receipt, err := s.ImportMaterial(ctx, human, c)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.ProcessNextJob(ctx); err != nil {
			t.Fatal(err)
		}
		job, err := s.GetJob(ctx, human, receipt.JobID)
		if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
			t.Fatal("paper job failed", job, err)
		}
		return job
	}
	a := importJob(command("official-A", "arxiv", "2504.16054v1", ""))
	oldContent, err := s.GetContent(ctx, human, *a.MaterialID, 1)
	if err != nil {
		t.Fatal(err)
	}
	locator := "https://arxiv.org/html/2504.16054v1"
	export := func(text string) string {
		raw, err := json.Marshal(map[string]any{"input": map[string]any{"url": locator}, "extracted": map[string]any{"title": "Contract-local paper", "content": text}})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	bExport := export("different genuine-export-shaped contract-local bytes")
	b := importJob(command("export-B", "summarize_json", locator, bExport))
	if *a.MaterialID != *b.MaterialID || b.MaterialRevision == nil || *b.MaterialRevision != 2 {
		t.Fatal("supported adapters split paper identity or erased changed bytes", a, b)
	}
	second := importJob(command("export-B-new-form", "summarize_json", locator, bExport))
	if second.JobID != b.JobID {
		t.Fatal("same exact export started a new job")
	}
	oldA := importJob(command("export-old-A", "summarize_json", locator, export(oldContent.Text)))
	if oldA.MaterialRevision == nil || *oldA.MaterialRevision != 1 {
		t.Fatal("old exact content was not reused", oldA)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	s = app.NewAttentionService(NewAttentionRepository(db), reader, store, app.ServiceOptions{})
	detail, err := s.GetMaterial(ctx, human, *a.MaterialID)
	if err != nil || detail.Material.CurrentRevision != 2 || len(detail.Revisions) != 2 || detail.Revisions[1].SourceKey != "arxiv:2504.16054" || detail.Revisions[1].SourceLocator != locator || detail.Revisions[1].Provenance.Mode != "existing_json_export" {
		t.Fatal("restart lost one material, B head or truthful export provenance", detail, err)
	}
	attachment, err := s.GetAttachment(ctx, human, *a.MaterialID, 2, "summarize.json")
	if err != nil || string(attachment.Data) != bExport {
		t.Fatal("original export not retained", err)
	}
	var materialCount int
	if err = db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM attention_materials").Scan(&materialCount); err != nil || materialCount != 1 {
		t.Fatal("duplicate paper materials persisted", materialCount, err)
	}
}
