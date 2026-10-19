package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExplicitProjectDiscoveryConfiguration(t *testing.T) {
	// 克隆当前真实公开仓库，分支初始化不依赖调用方是否处于 detached HEAD。
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	sourceRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	base := filepath.Join(sourceRoot, ".verification", "discovery-config")
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(base, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gitCommand := func(args ...string) string {
		t.Helper()
		command := exec.CommandContext(ctx, "git", append([]string{"-c", "core.hooksPath="}, args...)...)
		for _, value := range os.Environ() {
			key, _, _ := strings.Cut(value, "=")
			if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "GIT_TERMINAL_PROMPT=0")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("actual Git initialization %v: %v: %s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	head := gitCommand("-C", sourceRoot, "rev-parse", "HEAD")
	root := filepath.Join(directory, "public-project")
	gitCommand("clone", "--shared", "--no-checkout", "--", sourceRoot, root)
	branch := "discovery-validation"
	gitCommand("-C", root, "switch", "--create", branch, head)
	t.Logf("actual discovery repository=%s head=%s branch=%s", root, head, branch)
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
			if *git.Head != head || *git.Branch != branch || *git.Dirty {
				t.Fatalf("actual initialized Git state changed: %+v", git)
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
