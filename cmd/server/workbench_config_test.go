package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
)

func TestWorkbenchDataDirRejectedBeforeMigrations(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("ASTROCYTE_DATA_DIR", "explicit")
	for _, dir := range []string{filepath.Dir(sqlite.WorkbenchFormalSource), filepath.Join(workbenchRoot(), "snapshots"), filepath.Join(workbenchRoot(), "snapshots", "generation")} {
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
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Setenv("ASTROCYTE_DATA_DIR", "explicit")
	own := filepath.Join(workbenchRoot(), "app-state")
	copyDir := filepath.Join(workbenchRoot(), "snapshots", "fixture")
	for _, dir := range []string{own, copyDir} {
		if e := os.MkdirAll(dir, 0700); e != nil {
			t.Fatal(e)
		}
	}
	copyFile := filepath.Join(copyDir, "snapshot.sqlite")
	if e := os.WriteFile(copyFile, []byte("fixture"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Link(copyFile, filepath.Join(own, "state.sqlite")); e != nil {
		t.Fatal(e)
	}
	if _, e := guardWorkbenchDataDir(own); e == nil {
		t.Fatal("snapshot hardlink accepted as writable app DB")
	}
}
