package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitProjectDiscoveryConfiguration(t *testing.T) {
	// 实际工作树及空目录用于观察文件和 Git 生命周期。
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "..", ".."))
	for _, raw := range []string{"", "null", "{}", "[", `["relative"]`, `["` + filepath.ToSlash(filepath.Join(t.TempDir(), "missing")) + `"]`} {
		t.Setenv("ASTROCYTE_PROJECT_DISCOVERY_ROOTS", raw)
		if _, err := resolveProjectDiscovery(); err == nil {
			t.Fatalf("invalid discovery configuration accepted: %q", raw)
		}
	}
	for _, roots := range [][]string{{}, {root}, {t.TempDir()}} {
		raw, err := json.Marshal(roots)
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("ASTROCYTE_PROJECT_DISCOVERY_ROOTS", string(raw))
		source, err := resolveProjectDiscovery()
		if err != nil {
			t.Fatal(err)
		}
		observed, err := source.DiscoverRegistered(context.Background())
		if err != nil || len(observed.Projects) != len(roots) {
			t.Fatalf("explicit roots lost: %+v %v", observed, err)
		}
		if len(roots) == 1 && roots[0] == root {
			git := observed.Projects[0].Git
			if git.Status != "known" || git.Head == nil || git.Branch == nil || git.LastCommitAt == nil || git.Dirty == nil {
				t.Fatalf("actual worktree Git state missing: %+v", git)
			}
		}
	}
	raw, err := json.Marshal([]string{root, root})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASTROCYTE_PROJECT_DISCOVERY_ROOTS", string(raw))
	if _, err := resolveProjectDiscovery(); err == nil {
		t.Fatal("duplicate explicit roots accepted")
	}
	removed := t.TempDir()
	raw, err = json.Marshal([]string{removed})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASTROCYTE_PROJECT_DISCOVERY_ROOTS", string(raw))
	source, err := resolveProjectDiscovery()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(removed); err != nil {
		t.Fatal(err)
	}
	if _, err := source.DiscoverRegistered(context.Background()); err == nil {
		t.Fatal("removed explicit root did not fail")
	}
}
