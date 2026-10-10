package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// progressTable stands in for the W0-owned migration. It mirrors the proposed
// 013_project_progress schema so the storage layer is testable before the
// migration lands.
func ensureProgressTable(t *testing.T, db *DB) {
	t.Helper()
	if _, err := db.conn.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS local_project_progress (project_id TEXT PRIMARY KEY, data TEXT NOT NULL, revision INTEGER NOT NULL CHECK (revision >= 1))"); err != nil {
		t.Fatal(err)
	}
}

func TestProgressStorageRoundTripAndCAS(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	ensureProgressTable(t, db)

	if got, err := db.LoadProjectProgress(ctx, "missing"); err != nil || got.Status != "" || got.Revision != 0 {
		t.Fatalf("absent project should be zero value, got %+v %v", got, err)
	}

	percent := 50
	record := domain.ProjectProgress{SchemaVersion: 1, ProjectID: "p1", Status: "in_progress", Percent: &percent, Source: "agent_inferred", Evidence: []domain.ProgressEvidence{{SourcePath: "TASK.md", Kind: "task", Version: "v1"}}, NativeID: "nid", ObservedAt: time.Now().UTC(), Revision: 1, Operations: map[string]domain.ProgressOperation{"op1": {Action: "infer", Status: "accepted", CreatedAt: time.Now().UTC()}}}
	if err := db.SaveProjectProgress(ctx, record, 0); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.LoadProjectProgress(ctx, "p1")
	if err != nil || loaded.Status != "in_progress" || loaded.Source != "agent_inferred" || loaded.Revision != 1 || loaded.NativeID != "nid" || len(loaded.Evidence) != 1 || loaded.Operations["op1"].Status != "accepted" {
		t.Fatalf("round trip %+v %v", loaded, err)
	}

	if err := db.SaveProjectProgress(ctx, record, 0); err == nil {
		t.Fatal("stale insert accepted")
	}
	record.Revision = 2
	record.Status = "review"
	if err := db.SaveProjectProgress(ctx, record, 1); err != nil {
		t.Fatal(err)
	}
	if loaded, err := db.LoadProjectProgress(ctx, "p1"); err != nil || loaded.Status != "review" || loaded.Revision != 2 {
		t.Fatalf("update %+v %v", loaded, err)
	}
	// Progress storage must not create grants, projects or sessions.
	projects, err := db.ListProjects(ctx)
	if err != nil || len(projects) != 0 {
		t.Fatal("progress storage touched project rows")
	}
	if _, err := db.LoadGrant(ctx, "p1", "a"); err == nil {
		t.Fatal("progress storage created a grant")
	}
}
