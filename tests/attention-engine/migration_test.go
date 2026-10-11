package attention_engine_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// M0 migration test: 014 must apply cleanly on a fresh temporary database,
// create the §4 tables, bump the attention schema version and enforce the
// closed status domains. No store code exists yet; this proves the schema lock.

func openEngineDB(t *testing.T) (*sqlite.DB, context.Context) {
	t.Helper()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}
	return db, ctx
}

func TestMigration014CreatesEngineTables(t *testing.T) {
	db, ctx := openEngineDB(t)
	for _, name := range []string{
		"attention_engine_events",
		"attention_engine_signals",
		"attention_engine_nodes",
		"attention_engine_experiences",
		"attention_engine_knowledge",
		"attention_engine_conflicts",
	} {
		var got string
		err := db.Conn().QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&got)
		if err == sql.ErrNoRows {
			t.Errorf("table %s missing", name)
		} else if err != nil {
			t.Errorf("query %s: %v", name, err)
		}
	}
}

func TestMigration014BumpsAttentionVersion(t *testing.T) {
	db, ctx := openEngineDB(t)
	var version int
	if err := db.Conn().QueryRowContext(ctx, "SELECT version FROM foundation_schema WHERE context='attention'").Scan(&version); err != nil {
		t.Fatalf("attention schema version: %v", err)
	}
	if version != 7 {
		t.Errorf("attention schema version = %d, want 7", version)
	}
}

func TestMigration014EnforcesStatusDomains(t *testing.T) {
	db, ctx := openEngineDB(t)
	exec := func(q string, args ...any) error {
		_, err := db.Conn().ExecContext(ctx, q, args...)
		return err
	}
	// Valid statuses are accepted.
	if err := exec("INSERT INTO attention_engine_experiences(id,pattern,verdict,status,data) VALUES('e1','p','avoid','quarantine','{}')"); err != nil {
		t.Fatalf("valid experience status rejected: %v", err)
	}
	if err := exec("INSERT INTO attention_engine_knowledge(id,topic,status,data) VALUES('k1','t','confirmed','{}')"); err != nil {
		t.Fatalf("valid knowledge status rejected: %v", err)
	}
	if err := exec("INSERT INTO attention_engine_conflicts(id,topic,kind,status,data) VALUES('c1','t','knowledge_entry','resolved_merge','{}')"); err != nil {
		t.Fatalf("valid conflict status rejected: %v", err)
	}
	// Closed status domains reject unknown values.
	if err := exec("INSERT INTO attention_engine_experiences(id,pattern,verdict,status,data) VALUES('e2','p','avoid','published','{}')"); err == nil {
		t.Error("unknown experience status accepted")
	}
	if err := exec("INSERT INTO attention_engine_knowledge(id,topic,status,data) VALUES('k2','t','archived','{}')"); err == nil {
		t.Error("unknown knowledge status accepted")
	}
	if err := exec("INSERT INTO attention_engine_conflicts(id,topic,kind,status,data) VALUES('c2','t','x','auto','{}')"); err == nil {
		t.Error("unknown conflict status accepted")
	}
}

func TestMigration014SignalForeignKey(t *testing.T) {
	db, ctx := openEngineDB(t)
	exec := func(q string, args ...any) error {
		_, err := db.Conn().ExecContext(ctx, q, args...)
		return err
	}
	if err := exec("INSERT INTO attention_engine_events(id,source,ts,kind,topic,data) VALUES('ev1','manual',1,'collect','','{}')"); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if err := exec("INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES('ev1','t','implicit_depth',1.0,1,'{}')"); err != nil {
		t.Fatalf("valid signal rejected: %v", err)
	}
	if err := exec("INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES('missing','t','implicit_depth',1.0,1,'{}')"); err == nil {
		t.Error("signal with missing event foreign key accepted")
	}
}

func TestMigration014IdempotentReapply(t *testing.T) {
	db, ctx := openEngineDB(t)
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatalf("second RunMigrations: %v", err)
	}
	var n int
	if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE id='014_attention_engine'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("migration 014 recorded %d times, want 1", n)
	}
}
