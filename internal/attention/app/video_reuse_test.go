package app

import (
	"context"
	"testing"
	"time"
)

func TestNormalVideoReuseExplicitRefreshAndABAHead(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	calls := 0
	text := "A"
	now := s.options.Clock()
	s.options.Clock = func() time.Time { now = now.Add(time.Millisecond); return now }
	s.sources = sourceFunc(func(_ context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		calls++
		return ImportedSource{SourceKey: "https://www.bilibili.com/video/BV1PReT6EEqR/", SourceLocator: c.SourceLocator, Kind: "video", Text: text}, nil
	})
	c := ImportMaterialCommand{CommandMeta: meta("video-A", 1), SourceLocator: "https://www.bilibili.com/video/BV1PReT6EEqR/?spm_id_from=1", Kind: "video", Adapter: "summarize_url"}
	a, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	c.CommandMeta = meta("video-new-form", 1)
	c.SourceLocator = "https://m.bilibili.com/video/BV1PReT6EEqR?p=1"
	duplicate, err := s.ImportMaterial(ctx, human, c)
	if err != nil || duplicate.JobID != a.JobID || calls != 1 {
		t.Fatal("normal alias reread video", err)
	}
	text = "B"
	c.Refresh = true
	c.CommandMeta = meta("video-refresh-B", 1)
	b, err := s.ImportMaterial(ctx, human, c)
	if err != nil || b.JobID == a.JobID {
		t.Fatal("refresh did not queue new extraction", err)
	}
	s.ProcessNextJob(ctx)
	c.Refresh = false
	c.CommandMeta = meta("video-form-after-B", 1)
	latest, err := s.ImportMaterial(ctx, human, c)
	if err != nil || latest.JobID != b.JobID || calls != 2 {
		t.Fatal("normal form did not reuse latest refresh", err)
	}
	text = "A"
	c.Refresh = true
	c.CommandMeta = meta("video-refresh-old-A", 1)
	oldA, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	job, _ := s.GetJob(ctx, human, oldA.JobID)
	material, _ := s.GetMaterial(ctx, human, *job.MaterialID)
	if calls != 3 || len(repo.state.Jobs) != 3 || len(material.Revisions) != 2 || material.Material.CurrentRevision != 2 || *job.MaterialRevision != 1 {
		t.Fatalf("A/B/A duplicated revision or moved head: calls%d job%+v material%+v", calls, job, material.Material)
	}
	// Multipart p=2 is a different extraction, unlike tracking and p=1 aliases.
	c.Refresh = false
	c.CommandMeta = meta("video-page-two", 1)
	c.SourceLocator = "https://www.bilibili.com/video/BV1PReT6EEqR/?p=2"
	pageTwo, err := s.ImportMaterial(ctx, human, c)
	if err != nil || pageTwo.JobID == oldA.JobID {
		t.Fatal("different video page reused another part", err)
	}
}
