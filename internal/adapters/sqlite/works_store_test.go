package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// ensureWorksTables stands in for the W0-owned migration. It mirrors the
// proposed 014_works.sql schema so the storage layer is testable before the
// migration lands.
func ensureWorksTables(t *testing.T, db *DB) {
	t.Helper()
	if _, err := db.conn.ExecContext(context.Background(),
		"CREATE TABLE IF NOT EXISTS works_album (key TEXT PRIMARY KEY, data TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.ExecContext(context.Background(),
		"CREATE TABLE IF NOT EXISTS works_metadata (work_id TEXT PRIMARY KEY, data TEXT NOT NULL, revision INTEGER NOT NULL CHECK (revision >= 1))"); err != nil {
		t.Fatal(err)
	}
}

func TestWorksSnapshotRoundTrip(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	ensureWorksTables(t, db)

	if got, err := db.LoadWorksSnapshot(ctx); err != nil || len(got.Works) != 0 || got.Status != "" {
		t.Fatalf("fresh album should be empty, got %+v %v", got, err)
	}

	mtime := time.Now().UTC()
	snapshot := domain.WorksSnapshot{
		Works:     []domain.Work{{ID: "w1", Platform: "codex", Name: "app", Files: domain.WorkFileFact{Mtime: &mtime, Verified: true}}},
		Platforms: []domain.PlatformStat{{Platform: "codex", Works: 1}},
		Status:    "complete",
		Failures:  []domain.WorkDiscoveryFailure{},
	}
	if err := db.SaveWorksSnapshot(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.LoadWorksSnapshot(ctx)
	if err != nil || len(loaded.Works) != 1 || loaded.Works[0].Platform != "codex" || loaded.Status != "complete" {
		t.Fatalf("round trip %+v %v", loaded, err)
	}
}

func TestWorkMetadataRoundTripAndCAS(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	ensureWorksTables(t, db)

	if got, err := db.LoadWorkMetadata(ctx, "w1"); err != nil || got.Revision != 0 {
		t.Fatalf("absent metadata should be zero, got %+v %v", got, err)
	}

	stamp := time.Now().UTC()
	metadata := domain.WorkHumanMetadata{Attitude: domain.AttitudeContinue, Group: "research", Handoff: domain.WorkHandoff{Content: "notes", Source: domain.HandoffSourceHuman, UpdatedAt: &stamp}, UpdatedAt: &stamp, Revision: 1}
	if err := db.SaveWorkMetadata(ctx, "w1", metadata, 0); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.LoadWorkMetadata(ctx, "w1")
	if err != nil || loaded.Attitude != domain.AttitudeContinue || loaded.Revision != 1 || loaded.Handoff.Content != "notes" {
		t.Fatalf("round trip %+v %v", loaded, err)
	}

	if err := db.SaveWorkMetadata(ctx, "w1", metadata, 0); err == nil {
		t.Fatal("stale insert accepted")
	}
	metadata.Revision = 2
	metadata.Attitude = domain.AttitudeDone
	if err := db.SaveWorkMetadata(ctx, "w1", metadata, 1); err != nil {
		t.Fatal(err)
	}
	if loaded, err := db.LoadWorkMetadata(ctx, "w1"); err != nil || loaded.Attitude != domain.AttitudeDone || loaded.Revision != 2 {
		t.Fatalf("update %+v %v", loaded, err)
	}
}
