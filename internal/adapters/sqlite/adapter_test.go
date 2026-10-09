package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Verify the file was created.
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file not created: %v", err)
	}
}

func TestRunMigrations_Initial(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	migrationsFS := fstest.MapFS{
		"001_initial.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE IF NOT EXISTS test_table (id TEXT PRIMARY KEY);
				INSERT OR IGNORE INTO _migrations (id) VALUES ('001_initial');
			`),
		},
	}

	ctx := context.Background()
	if err := db.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	// Verify migration was recorded.
	var id string
	err = db.Conn().QueryRowContext(ctx, "SELECT id FROM _migrations WHERE id = '001_initial'").Scan(&id)
	if err != nil {
		t.Fatalf("migration not recorded: %v", err)
	}
	if id != "001_initial" {
		t.Errorf("unexpected migration id: %q", id)
	}

	// Verify test_table was created.
	_, err = db.Conn().ExecContext(ctx, "INSERT INTO test_table (id) VALUES ('test')")
	if err != nil {
		t.Fatalf("test_table not created: %v", err)
	}
}

func TestRunMigrations_Reopen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	migrationsFS := fstest.MapFS{
		"001_initial.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE IF NOT EXISTS test_table (id TEXT PRIMARY KEY, value TEXT);
				INSERT OR IGNORE INTO _migrations (id) VALUES ('001_initial');
			`),
		},
		"002_second.sql": &fstest.MapFile{
			Data: []byte(`
				ALTER TABLE test_table ADD COLUMN extra TEXT;
				INSERT OR IGNORE INTO _migrations (id) VALUES ('002_second');
			`),
		},
	}

	ctx := context.Background()

	// First open: apply 001.
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	mig1 := fstest.MapFS{"001_initial.sql": migrationsFS["001_initial.sql"]}
	if err := db.RunMigrations(ctx, mig1); err != nil {
		t.Fatalf("RunMigrations (first): %v", err)
	}
	db.Close()

	// Reopen: apply both 001 (skip) and 002.
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	defer db2.Close()
	if err := db2.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations (second): %v", err)
	}

	// Verify 002 was applied.
	var id string
	err = db2.Conn().QueryRowContext(ctx, "SELECT id FROM _migrations WHERE id = '002_second'").Scan(&id)
	if err != nil {
		t.Fatalf("002 migration not recorded: %v", err)
	}

	// Verify new column exists.
	_, err = db2.Conn().ExecContext(ctx, "INSERT INTO test_table (id, value, extra) VALUES ('a', 'b', 'c')")
	if err != nil {
		t.Fatalf("new column not available: %v", err)
	}
}

func TestRunMigrations_FailureAtomicity(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Apply a valid migration first.
	validFS := fstest.MapFS{
		"001_valid.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE IF NOT EXISTS good_table (id TEXT PRIMARY KEY);
				INSERT OR IGNORE INTO _migrations (id) VALUES ('001_valid');
			`),
		},
	}
	if err := db.RunMigrations(ctx, validFS); err != nil {
		t.Fatalf("RunMigrations (valid): %v", err)
	}

	// Apply a migration that will fail (syntax error).
	badFS := fstest.MapFS{
		"002_bad.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE bad_table (id TEXT PRIMARY KEY);
				THIS IS INVALID SQL;
				INSERT OR IGNORE INTO _migrations (id) VALUES ('002_bad');
			`),
		},
	}
	err = db.RunMigrations(ctx, badFS)
	if err == nil {
		t.Fatal("expected migration to fail")
	}

	// Verify 002 was NOT recorded (rolled back).
	var count int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE id = '002_bad'").Scan(&count)
	if err != nil {
		t.Fatalf("query migrations: %v", err)
	}
	if count != 0 {
		t.Errorf("failed migration was recorded, expected rollback")
	}

	// Verify 001 is still present.
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations WHERE id = '001_valid'").Scan(&count)
	if err != nil {
		t.Fatalf("query migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("valid migration was lost")
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	migrationsFS := fstest.MapFS{
		"001_initial.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE IF NOT EXISTS test_table (id TEXT PRIMARY KEY);
				INSERT OR IGNORE INTO _migrations (id) VALUES ('001_initial');
			`),
		},
	}

	ctx := context.Background()

	// Run twice; second should be a no-op.
	if err := db.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations (first): %v", err)
	}
	if err := db.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations (second): %v", err)
	}

	var count int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM _migrations").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration, got %d", count)
	}
}

func TestSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Before migrations: version 0.
	v, err := db.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v != 0 {
		t.Errorf("expected version 0, got %d", v)
	}

	// After migration with foundation_schema.
	migrationsFS := fstest.MapFS{
		"001_initial.sql": &fstest.MapFile{
			Data: []byte(`
				CREATE TABLE IF NOT EXISTS foundation_schema (
					context TEXT PRIMARY KEY,
					version INTEGER NOT NULL,
					updated_at TEXT NOT NULL DEFAULT 'now'
				);
				INSERT OR IGNORE INTO foundation_schema (context, version) VALUES ('foundation', 1);
				INSERT OR IGNORE INTO _migrations (id) VALUES ('001_initial');
			`),
		},
	}
	if err := db.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	v, err = db.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v != 1 {
		t.Errorf("expected version 1, got %d", v)
	}
}

func TestRunMigrations_RealFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Use the real migrations directory.
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	migrationsFS := os.DirFS(migrationsDir)

	if err := db.RunMigrations(ctx, migrationsFS); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	// Verify foundation_schema table exists and has version 1.
	var version int
	err = db.Conn().QueryRowContext(ctx,
		"SELECT version FROM foundation_schema WHERE context = 'foundation'").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("foundation_schema row not found")
		}
		t.Fatalf("query: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
}
