package app_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestFixedArxivImportReusesJobAndPreservesVersions(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	reader := &observedSource{Reader: importers.NewReader(nil)}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
	c := ImportMaterialCommand{CommandMeta: meta("fixed-first", 1), Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v1", Adapter: "arxiv"}
	first, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("queue official fixed paper")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process official fixed paper")
	}
	firstJob, err := s.GetJob(ctx, human, first.JobID)
	if err != nil || firstJob.Status != "succeeded" || firstJob.MaterialID == nil || firstJob.MaterialRevision == nil || *firstJob.MaterialRevision != 1 {
		t.Fatal("first official Atom/PDF import failed")
	}
	before, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil {
		t.Fatal("load official material")
	}
	oldContent, err := s.GetContent(ctx, human, *firstJob.MaterialID, 1)
	if err != nil || oldContent.Text == "" {
		t.Fatal("read official first revision")
	}
	oldPDF, err := s.GetAttachment(ctx, human, *firstJob.MaterialID, 1, "2501.12948v1.pdf")
	if err != nil || len(oldPDF.Data) == 0 {
		t.Fatal("read official first PDF")
	}
	publications := objects.publishes.Load()
	c.CommandMeta = meta("fixed-new-form", 1)
	second, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("repeat fixed paper form")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process fixed reuse")
	}
	if second.JobID != first.JobID || second.Status != "succeeded" || reader.calls.Load() != 1 || countRows(t, db, "attention_jobs") != 1 {
		t.Fatal("fixed revision repeated external source work")
	}
	after, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || !reflect.DeepEqual(before, after) || objects.publishes.Load() != publications {
		t.Fatal("reuse changed source history or republished objects")
	}
	original := c
	original.CommandMeta = meta("fixed-first", 1)
	replayed, err := s.ImportMaterial(ctx, human, original)
	if err != nil || replayed != first {
		t.Fatal("original queued receipt changed")
	}
	original.SourceLocator = "https://arxiv.org/abs/2501.12948v2"
	_, err = s.ImportMaterial(ctx, human, original)
	errorCode(t, err, apierrors.VersionConflict)
	c.CommandMeta = meta("fixed-v2", 1)
	c.SourceLocator = original.SourceLocator
	next, err := s.ImportMaterial(ctx, human, c)
	if err != nil || next.JobID == first.JobID {
		t.Fatal("explicit second fixed revision did not create work")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process official second revision")
	}
	nextJob, err := s.GetJob(ctx, human, next.JobID)
	if err != nil || nextJob.Status != "succeeded" || nextJob.MaterialID == nil || *nextJob.MaterialID != *firstJob.MaterialID || nextJob.MaterialRevision == nil || *nextJob.MaterialRevision != 2 {
		t.Fatal("second official version did not append to same material")
	}
	after, err = s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || reader.calls.Load() != 2 || countRows(t, db, "attention_jobs") != 2 || countRows(t, db, "attention_materials") != 1 || len(after.Revisions) != 2 || after.Material.CurrentRevision != 2 || !reflect.DeepEqual(after.Revisions[0], before.Revisions[0]) {
		t.Fatal("new official version lost first revision")
	}
	content, err := s.GetContent(ctx, human, *firstJob.MaterialID, 1)
	if err != nil || content.Text != oldContent.Text {
		t.Fatal("old official source text lost")
	}
	pdf, err := s.GetAttachment(ctx, human, *firstJob.MaterialID, 1, "2501.12948v1.pdf")
	if err != nil || !reflect.DeepEqual(pdf.Data, oldPDF.Data) {
		t.Fatal("old original PDF attachment lost")
	}
	c.CommandMeta = meta("fixed-v1-again", 1)
	c.SourceLocator = "https://arxiv.org/abs/2501.12948v1"
	reused, err := s.ImportMaterial(ctx, human, c)
	if err != nil || reused.JobID != first.JobID {
		t.Fatal("older fixed job replaced")
	}
	final, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || !reflect.DeepEqual(final, after) || reader.calls.Load() != 2 {
		t.Fatal("old fixed job reuse moved current head")
	}
}

func TestFixedArxivAliasesReuseSingleSourceRead(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	reader := &observedSource{Reader: importers.NewReader(nil)}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
	aliases := []string{"2501.12948v1", "arXiv:2501.12948v1", "https://arxiv.org/abs/2501.12948v1", "https://arxiv.org/pdf/2501.12948v1", "https://arxiv.org/pdf/2501.12948v1.pdf", "https://arxiv.org/html/2501.12948v1", "https://www.arxiv.org/abs/2501.12948v1?tracking=1#section", "http://export.arxiv.org/abs/2501.12948v1"}
	var first ImportJobResult
	for i, locator := range aliases {
		c := ImportMaterialCommand{CommandMeta: meta(locator, 1), Kind: "paper", SourceLocator: locator}
		if i%2 != 0 {
			c.Adapter = "arxiv"
		}
		result, err := s.ImportMaterial(ctx, human, c)
		if err != nil {
			t.Fatal("queue equivalent fixed alias")
		}
		if i == 0 {
			first = result
		}
		if result != first {
			t.Fatal("fixed alias created a different queued job")
		}
	}
	if _, err := s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process actual official alias source")
	}
	completed, err := s.ImportMaterial(ctx, human, ImportMaterialCommand{CommandMeta: meta("completed-alias", 1), Kind: "paper", SourceLocator: aliases[4], Adapter: "arxiv"})
	if err != nil || completed.JobID != first.JobID || completed.Status != "succeeded" {
		t.Fatal("fixed alias did not reuse actual completed result")
	}
	worked, err := s.ProcessNextJob(ctx)
	if err != nil || worked || reader.calls.Load() != 1 || countRows(t, db, "attention_jobs") != 1 || countRows(t, db, "attention_materials") != 1 || countRows(t, db, "attention_receipts") != len(aliases)+1 {
		t.Fatal("aliases repeated external work or lost receipts")
	}
}

func TestImportFixedArxivBoundaryQueuesDurably(t *testing.T) {
	inputs := []ImportMaterialCommand{
		{SourceLocator: "https://arxiv.org/abs/2401.01234v1", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/pdf/2401.01234v1.pdf", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://www.arxiv.org/html/2401.01234v1?tracking=1#s2", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "http://export.arxiv.org/abs/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "2401.01234v12", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "arXiv:hep-th/9901001v2", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/math.GT/0309136v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234?version=v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://example.test/abs/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org.example.test/abs/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://reader@arxiv.org/abs/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "ftp://arxiv.org/abs/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/other/2401.01234v1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/arbitraryv1", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v0", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v01", Adapter: "arxiv", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v1", Adapter: "manual", Kind: "paper"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v1", Adapter: "summarize_json", Kind: "paper", LocalFileRef: "mutable.json"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v1", Kind: "paper", LocalFileRef: "mutable.md"},
		{SourceLocator: "https://arxiv.org/abs/2401.01234v1", Kind: "text"},
	}
	for i, c := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s, db, _ := fixture(t)
			c.CommandMeta = meta("one", 1)
			first, err := s.ImportMaterial(context.Background(), human, c)
			if err != nil {
				t.Fatal("first boundary request not queued")
			}
			c.CommandMeta = meta("two", 1)
			second, err := s.ImportMaterial(context.Background(), human, c)
			if err != nil || first.JobID != second.JobID || countRows(t, db, "attention_jobs") != 1 || countRows(t, db, "attention_receipts") != 2 {
				t.Fatal("boundary requests duplicated durable work or lost receipts")
			}
		})
	}
}

// 明确 manual 导出参数与真实可变本地文件；不声明远端来源发生了更新。
func TestMutableImportCommandsReadUpdatedContent(t *testing.T) {
	for _, local := range []bool{false, true} {
		name := "explicit manual export"
		if local {
			name = "actual local export with fixed locator"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			_, db, objects := fixture(t)
			dir, err := os.MkdirTemp(evidenceRoot(t), "mutable-export-")
			if err != nil {
				t.Fatal("create owned export directory")
			}
			reader := &observedSource{Reader: importers.NewReader([]string{dir})}
			s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
			text := "first manual source bytes"
			file := filepath.Join(dir, "export.md")
			if err = os.WriteFile(file, []byte(text), 0600); err != nil {
				t.Fatal("write first selected manual export")
			}
			c := importCommand("mutable-first", text)
			if local {
				c.LocalFileRef, c.ExportText = file, ""
			}
			first, err := s.ImportMaterial(ctx, human, c)
			if err != nil {
				t.Fatal("queue first manual export")
			}
			if _, err = s.ProcessNextJob(ctx); err != nil {
				t.Fatal("read first manual export")
			}
			firstJob, err := s.GetJob(ctx, human, first.JobID)
			if err != nil || firstJob.MaterialID == nil {
				t.Fatal("first manual export not persisted")
			}
			text = "updated manual source bytes"
			if err = os.WriteFile(file, []byte(text), 0600); err != nil {
				t.Fatal("update selected manual export")
			}
			if !local {
				c.ExportText = text
			}
			c.CommandMeta, c.Refresh = meta("mutable-again", 1), true
			second, err := s.ImportMaterial(ctx, human, c)
			if err != nil || second.JobID == first.JobID {
				t.Fatal("explicit refresh did not schedule a mutable read")
			}
			if _, err = s.ProcessNextJob(ctx); err != nil {
				t.Fatal("read changed manual export")
			}
			row, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
			if err != nil || reader.calls.Load() != 2 || len(row.Revisions) != 2 || row.Material.CurrentRevision != 2 {
				t.Fatal("mutable export update did not append revision")
			}
			old, err := s.GetContent(ctx, human, *firstJob.MaterialID, 1)
			if err != nil || old.Text != "first manual source bytes" {
				t.Fatal("mutable export lost first version")
			}
			current, err := s.GetContent(ctx, human, *firstJob.MaterialID, 2)
			if err != nil || current.Text != text {
				t.Fatal("new revision differs from updated actual export")
			}
		})
	}
}
