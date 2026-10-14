package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 只验证人工绑定参数和持久化，不声称已读取这些配置对应的外部目录。
func TestSourceIdentityAndPublicOnlyBinding(t *testing.T) {
	ctx := context.Background()
	s, db, _ := fixture(t)
	first, err := s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("bind", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "3494358764489275"})
	if err != nil {
		t.Fatal("bind approved source identity")
	}
	r, err := s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("bind-alias", 1), Platform: "bilibili", SourceKind: "uploads", ExternalID: "3494358764489275", Locator: "https://space.bilibili.com/3494358764489275/upload/video?tracking=1"})
	if err != nil || r.Source.ID != first.Source.ID || countRows(t, db, "attention_tracking_sources") != 1 {
		t.Fatal("account aliases duplicated durable binding")
	}
	_, err = s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("second-account", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "12345"})
	if err != nil || countRows(t, db, "attention_tracking_sources") != 2 {
		t.Fatal("multiple explicit binding parameters not retained")
	}
	_, err = s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("douyin-self", 1), Platform: "douyin", SourceKind: "favorites", Locator: "https://www.douyin.com/user/self?showTab=favorite_collection"})
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("mismatch", 1), Platform: "bilibili", SourceKind: "uploads", ExternalID: "1", Locator: "3494358764489275"})
	errorCode(t, err, apierrors.ValidationFailed)
	_, err = s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("folder", 1), Platform: "bilibili", SourceKind: "favorites", Locator: "https://space.bilibili.com/3494358764489275/favlist?fid=456"})
	if err != nil || countRows(t, db, "attention_tracking_sources") != 3 {
		t.Fatal("explicit folder not retained as separate configuration")
	}
}

func TestPublicTrackingSourceCapacity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	_, db, objects := fixture(t)
	python := os.Getenv("ASTROCYTE_LISTING_PYTHON")
	if python == "" {
		python = filepath.Join(os.Getenv("LOCALAPPDATA"), "Astrocyte", "media", "yt-dlp-2026.08.19", "Scripts", "python.exe")
	}
	reader := importers.NewPublicListingReader(python)
	s := NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), objects, ServiceOptions{ListingReader: reader, MaxAttempts: 1, JobTimeout: 3 * time.Minute})
	bound, err := s.BindTrackingSource(ctx, human, BindTrackingSourceCommand{CommandMeta: meta("public-bind", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "3494358764489275"})
	if err != nil {
		t.Fatal("bind authorized public source")
	}
	queued, err := s.SyncTrackingSource(ctx, human, bound.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("public-sync", bound.Source.Version)})
	if err != nil || len(queued.Jobs) != 1 {
		t.Fatal("queue actual public listing observation")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process actual public listing job")
	}
	job, err := s.GetJob(ctx, human, queued.Jobs[0].JobID)
	if err != nil {
		t.Fatal("load actual public listing job")
	}
	if job.Status != "succeeded" {
		code := "unknown"
		if job.Error != nil {
			code = string(job.Error.Code)
		}
		t.Fatalf("public listing did not succeed: status=%s code=%s; raw evidence remains in owned SQLite", job.Status, code)
	}
	observed, err := s.GetTrackingSource(ctx, human, bound.Source.ID)
	if err != nil {
		t.Fatal("load observed public metadata")
	}
	if len(observed.Items) != 100 || !observed.HasMore {
		t.Fatalf("public source cannot validate 100-item truncation: observed=%d has_more=%t", len(observed.Items), observed.HasMore)
	}
	if countRows(t, db, "attention_materials") != 0 {
		t.Fatal("metadata sync imported unselected bodies")
	}
}
