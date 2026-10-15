package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestWorkbenchDataDirRejectedBeforeMigrations(t *testing.T) {
	t.Setenv("LOCALAPPDATA", workbenchOwnedActualRoot(t))
	t.Setenv("ASTROCYTE_DATA_DIR", "explicit")
	sourceDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "owned-source")
	if err := os.MkdirAll(sourceDir, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(sourceDir, "state.sqlite")
	workbenchCreateActualSQLite(t, source)
	encoded, err := json.Marshal([]string{source})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", string(encoded))
	for _, dir := range []string{sourceDir, filepath.Join(workbenchRoot(), "snapshots"), filepath.Join(workbenchRoot(), "snapshots", "generation")} {
		if _, e := guardWorkbenchDataDir(dir); e == nil {
			t.Fatal("protected directory accepted")
		}
	}
	own := filepath.Join(workbenchRoot(), "app-state")
	if _, e := guardWorkbenchDataDir(own); e != nil {
		t.Fatal(e)
	}
	t.Setenv("ASTROCYTE_DATA_DIR", "")
	t.Setenv("ASTROCYTE_WORKBENCH", "1")
	dir, e := guardWorkbenchDataDir("other")
	if e != nil || dir != own {
		t.Fatal("runtime app state not isolated")
	}
	if e = os.MkdirAll(filepath.Join(workbenchRoot(), "snapshots"), 0700); e != nil {
		t.Fatal(e)
	}
}

func TestWorkbenchRejectsApplicationDBHardlinkToSnapshot(t *testing.T) {
	t.Setenv("LOCALAPPDATA", workbenchOwnedActualRoot(t))
	t.Setenv("ASTROCYTE_DATA_DIR", "explicit")
	t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", "")
	own := filepath.Join(workbenchRoot(), "app-state")
	copyDir := filepath.Join(workbenchRoot(), "snapshots", "owned-generation")
	for _, dir := range []string{own, copyDir} {
		if e := os.MkdirAll(dir, 0700); e != nil {
			t.Fatal(e)
		}
	}
	copyFile := filepath.Join(copyDir, "snapshot.sqlite")
	workbenchCreateActualSQLite(t, copyFile)
	if e := os.Link(copyFile, filepath.Join(own, "state.sqlite")); e != nil {
		t.Fatal(e)
	}
	if _, e := guardWorkbenchDataDir(own); e == nil {
		t.Fatal("snapshot hardlink accepted as writable app DB")
	}
}

func workbenchOwnedActualRoot(t *testing.T) string {
	t.Helper()
	parent, err := filepath.Abs(filepath.Join("..", "..", ".verification", "workbench-guard"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp(parent, "h-workbench-config-")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func workbenchCreateActualSQLite(t *testing.T, path string) {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
