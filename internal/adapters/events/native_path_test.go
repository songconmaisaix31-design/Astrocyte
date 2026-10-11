package events

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func writeReplayHeader(t *testing.T, path, project string) {
	t.Helper()
	header, err := json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": "owned-fixture", "cwd": project}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, append(header, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestNativeReplayRejectsOutsideAndWrongProject(t *testing.T) {
	root, outside, project := t.TempDir(), t.TempDir(), t.TempDir()
	path := filepath.Join(root, "selected.jsonl")
	out := filepath.Join(outside, "selected.jsonl")
	writeReplayHeader(t, path, project)
	writeReplayHeader(t, out, project)
	for name, v := range map[string]struct{ path, project string }{
		"outside_root":  {out, project},
		"wrong_project": {path, outside},
	} {
		t.Run(name, func(t *testing.T) {
			records, partial, err := ReplayNativeFile(context.Background(), "codex", root, v.path, v.project, 128, 1)
			if err == nil || len(records) != 0 || partial {
				t.Fatalf("unapproved scope accepted: count=%d partial=%v error=%v", len(records), partial, err)
			}
		})
	}
}

func TestNativeReplayRejectsOwnedSymlinks(t *testing.T) {
	base, project := t.TempDir(), t.TempDir()
	root := filepath.Join(base, "approved")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, "selected.jsonl")
	writeReplayHeader(t, path, project)
	for _, kind := range []string{"root", "ancestor", "file"} {
		t.Run(kind, func(t *testing.T) {
			link := filepath.Join(root, kind+"-link")
			destination := target
			if kind == "file" {
				destination = path
			}
			if err := os.Symlink(destination, link); err != nil {
				if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
					t.Skip("NOT_RUN: ordinary symlink fixture lacks Windows privilege")
				}
				t.Fatal(err)
			}
			selectedRoot, selectedPath := root, link
			if kind == "root" {
				selectedRoot, selectedPath = link, filepath.Join(link, "selected.jsonl")
			} else if kind == "ancestor" {
				selectedPath = filepath.Join(link, "selected.jsonl")
			}
			records, partial, err := ReplayNativeFile(context.Background(), "codex", selectedRoot, selectedPath, project, 128, 1)
			if err == nil || len(records) != 0 || partial || !strings.Contains(err.Error(), "symbolic links") {
				t.Fatalf("symlink accepted: count=%d partial=%v error=%v", len(records), partial, err)
			}
		})
	}
}
