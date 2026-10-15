package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestNormalVideoReuseExplicitRefreshAndABAHead(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	old := os.Getenv("ASTROCYTE_TEST_VIDEO_EXPORT_A")
	updated := os.Getenv("ASTROCYTE_TEST_VIDEO_EXPORT_B")
	if !filepath.IsAbs(old) || !filepath.IsAbs(updated) {
		t.Fatal("explicit immutable actual video export paths required")
	}
	for _, file := range []string{old, updated} {
		info, err := os.Lstat(file)
		if err != nil || !info.Mode().IsRegular() {
			t.Fatal("explicit immutable actual video export unavailable")
		}
	}
	reader := &observedSource{Reader: importers.NewReader([]string{filepath.Dir(old), filepath.Dir(updated)})}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
	c := ImportMaterialCommand{CommandMeta: meta("video-A", 1), SourceLocator: "https://www.bilibili.com/video/BV1PReT6EEqR/?spm_id_from=1", Kind: "video", Adapter: "summarize_markdown", LocalFileRef: old}
	a, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("queue actual old video export")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual old video export")
	}
	first, err := s.GetJob(ctx, human, a.JobID)
	if err != nil || first.Status != "succeeded" || first.MaterialID == nil {
		t.Fatal("actual old video export did not persist")
	}
	c.CommandMeta, c.SourceLocator = meta("video-new-form", 1), "https://m.bilibili.com/video/BV1PReT6EEqR?p=1"
	duplicate, err := s.ImportMaterial(ctx, human, c)
	if err != nil || duplicate.JobID != a.JobID || reader.calls.Load() != 1 {
		t.Fatal("normal video alias repeated actual source read")
	}
	c.Refresh, c.LocalFileRef, c.CommandMeta = true, updated, meta("video-refresh-B", 1)
	b, err := s.ImportMaterial(ctx, human, c)
	if err != nil || b.JobID == a.JobID {
		t.Fatal("explicit actual export refresh did not queue new extraction")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual updated video export")
	}
	c.Refresh, c.CommandMeta = false, meta("video-form-after-B", 1)
	latest, err := s.ImportMaterial(ctx, human, c)
	if err != nil || latest.JobID != b.JobID || reader.calls.Load() != 2 {
		t.Fatal("normal form did not reuse latest actual refresh")
	}
	c.LocalFileRef, c.Refresh, c.CommandMeta = old, true, meta("video-refresh-old-A", 1)
	oldA, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("queue explicitly selected immutable old export")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("reread selected actual old export")
	}
	job, err := s.GetJob(ctx, human, oldA.JobID)
	if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
		t.Fatal("actual old export refresh did not complete")
	}
	material, err := s.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil || reader.calls.Load() != 3 || countRows(t, db, "attention_jobs") != 3 || len(material.Revisions) != 2 || material.Material.CurrentRevision != 2 || job.MaterialRevision == nil || *job.MaterialRevision != 1 {
		t.Fatal("actual A/B/A repeated revision or changed current B head")
	}
	if material.Material.ID != *first.MaterialID {
		t.Fatal("same actual video exports created a second material")
	}
	oldBody, err := s.GetContent(ctx, human, material.Material.ID, 1)
	if err != nil {
		t.Fatal("read actual immutable old body")
	}
	newBody, err := s.GetContent(ctx, human, material.Material.ID, 2)
	if err != nil || oldBody.Text == newBody.Text {
		t.Fatal("actual distinct exports lost source history")
	}
	// p2 只使用其实际 URL 配置，不把 p1 导出作为 p2 原件。
	c.SourceLocator = "https://www.bilibili.com/video/BV1PReT6EEqR/?p=2"
	c.Adapter, c.LocalFileRef, c.Refresh, c.CommandMeta = "summarize_url", "", false, meta("video-page-two", 1)
	if domain.CanonicalWebKey(c.SourceLocator) == domain.CanonicalWebKey(material.Material.SourceLocator) {
		t.Fatal("canonical multipart identity reused p1")
	}
	pageTwo, err := s.ImportMaterial(ctx, human, c)
	if err != nil || pageTwo.JobID == oldA.JobID {
		t.Fatal("different video page reused another part")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process actual unavailable p2 extraction configuration")
	}
	blocked, err := s.GetJob(ctx, human, pageTwo.JobID)
	if err != nil || blocked.Status != "failed" || blocked.Error == nil || blocked.Error.Code != apierrors.ValidationFailed || blocked.MaterialID != nil {
		t.Fatal("unconfigured p2 extraction invented material")
	}
}
