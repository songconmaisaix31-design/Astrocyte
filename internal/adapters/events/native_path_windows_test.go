package events

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func replayShortPath(t *testing.T, path string) string {
	t.Helper()
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || n >= uint32(len(buf)) {
		t.Fatalf("GetShortPathName: length=%d error=%v", n, err)
	}
	short := windows.UTF16ToString(buf[:n])
	if strings.EqualFold(short, path) {
		t.Skip("NOT_RUN: owned fixture volume does not generate an 8.3 alias")
	}
	t.Logf("actual owned Windows alias: %s -> %s", short, path)
	return short
}

func replayFixture(t *testing.T, source, path, project string) ([]byte, []byte) {
	t.Helper()
	var header []byte
	var err error
	tool := []byte(`{ "type":"response_item", "timestamp":"2026-10-10T00:00:01.123Z", "payload":{"type":"function_call","name":"apply_patch","arguments":"owned fixture only"}}`)
	if source == "codex" {
		header, err = json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": "fixture-session", "cwd": project}})
	} else {
		header, err = json.Marshal(map[string]any{"type": "system", "sessionId": "fixture-session", "cwd": project})
		tool = []byte(`{ "type":"assistant", "timestamp":"2026-10-10T00:00:01.123Z", "message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"owned.txt"}}]}}`)
	}
	if err != nil {
		t.Fatal(err)
	}
	content := append(append(append([]byte{}, header...), '\n'), append(tool, '\n')...)
	if err = os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	return header, tool
}

func TestNativeReplayWindowsShortPathAliases(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "Approved long replay parent")
	root := filepath.Join(parent, "Approved long replay root")
	project := filepath.Join(t.TempDir(), "Approved long project root")
	for _, dir := range []string{root, project} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(root, "Selected long session filename.jsonl")
	replayFixture(t, "codex", path, project)
	shortParent, shortRoot := replayShortPath(t, parent), replayShortPath(t, root)
	shortFile, shortProject := replayShortPath(t, path), replayShortPath(t, project)
	for _, source := range []string{"codex", "claude_code"} {
		for name, v := range map[string]struct{ root, path, project, header string }{
			"short_root_and_file":  {shortRoot, shortFile, project, project},
			"short_root_long_file": {shortRoot, path, project, project},
			"long_root_short_file": {root, shortFile, project, project},
			"short_ancestor":       {filepath.Join(shortParent, filepath.Base(root)), path, project, project},
			"short_project":        {root, path, shortProject, project},
			"short_header":         {root, path, project, shortProject},
			"case_alias":           {strings.ToUpper(root), strings.ToUpper(path), strings.ToUpper(project), project},
		} {
			t.Run(source+"/"+name, func(t *testing.T) {
				header, tool := replayFixture(t, source, path, v.header)
				records, partial, err := ReplayNativeFile(context.Background(), source, v.root, v.path, v.project, 128, 1791590400500)
				if err != nil || partial || len(records) != 2 {
					t.Fatalf("legal alias rejected: count=%d partial=%v error=%v", len(records), partial, err)
				}
				if !bytes.Equal(records[0].RawBytes, header) || !bytes.Equal(records[1].RawBytes, tool) || records[0].Event != nil || records[1].Event == nil || records[1].Event.Kind != "edit_file" || records[1].Event.Scope != v.project {
					t.Fatal("alias changed raw bytes, pending mapping or approved scope")
				}
				for _, record := range records {
					if record.Observation.Tokens != nil || record.Observation.Depth != "unknown" || record.Observation.Outcome != "unknown" {
						t.Fatal("alias invented missing observations")
					}
				}
				bounded, partial, err := ReplayNativeFile(context.Background(), source, v.root, v.path, v.project, 1, 1791590400500)
				if err != nil || !partial || len(bounded) != 1 || bounded[0].Event != nil || !bytes.Equal(bounded[0].RawBytes, header) {
					t.Fatal("alias changed selected-file record bound or raw header")
				}
				mapped, err := MapNativeRecord(source, "fixture-header", v.project, header, 1791590400500)
				if err != nil || mapped.Event != nil || mapped.Observation.Scope != v.project || !bytes.Equal(mapped.RawBytes, header) {
					t.Fatalf("selected-record project alias rejected or changed: %v", err)
				}
			})
		}
	}
}

func replayJunction(t *testing.T, link, target string) {
	t.Helper()
	output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", filepath.FromSlash(link), filepath.FromSlash(target)).CombinedOutput()
	if err != nil {
		t.Fatalf("create owned junction: %v: %s", err, output)
	}
	linkInfo, err := os.Stat(link)
	if err != nil {
		t.Fatal(err)
	}
	targetInfo, err := os.Stat(target)
	if err != nil || !os.SameFile(linkInfo, targetInfo) {
		t.Fatalf("owned junction must point at target: %v", err)
	}
}

func TestReplayJunctionUsesActualWindowsPaths(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "Owned target with spaces")
	link := filepath.Join(base, "Owned junction with spaces")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	// 明确的正斜杠参数经过同一 helper，实际创建的 junction 由文件身份核实。
	replayJunction(t, filepath.ToSlash(link), filepath.ToSlash(target))
}

func TestNativeReplayWindowsRejectsOwnedJunctions(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "Approved replay root")
	target := filepath.Join(root, "Owned inner target")
	outside := t.TempDir()
	project := t.TempDir()
	for _, dir := range []string{target, filepath.Join(target, "child")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{target, filepath.Join(target, "child"), outside} {
		replayFixture(t, "codex", filepath.Join(dir, "selected.jsonl"), project)
	}
	rootLink := filepath.Join(base, "Owned root junction")
	innerLink := filepath.Join(root, "Owned inner junction")
	outerLink := filepath.Join(root, "Owned outer junction")
	for link, destination := range map[string]string{rootLink: target, innerLink: target, outerLink: outside} {
		replayJunction(t, link, destination)
	}
	for name, v := range map[string]struct{ root, path string }{
		"terminal_root":              {rootLink, filepath.Join(rootLink, "selected.jsonl")},
		"root_ancestor":              {filepath.Join(rootLink, "child"), filepath.Join(rootLink, "child", "selected.jsonl")},
		"file_ancestor_inside_root":  {root, filepath.Join(innerLink, "selected.jsonl")},
		"file_ancestor_outside_root": {root, filepath.Join(outerLink, "selected.jsonl")},
		"terminal_file_reparse":      {root, innerLink},
	} {
		t.Run(name, func(t *testing.T) {
			records, partial, err := ReplayNativeFile(context.Background(), "codex", v.root, v.path, project, 128, 1)
			if err == nil || len(records) != 0 || partial || !strings.Contains(err.Error(), "symbolic links") {
				t.Fatalf("real redirect accepted or obscured: count=%d partial=%v error=%v", len(records), partial, err)
			}
		})
	}
	t.Run("short_named_junction_ancestor", func(t *testing.T) {
		link := replayShortPath(t, innerLink)
		_, _, err := ReplayNativeFile(context.Background(), "codex", root, filepath.Join(link, "selected.jsonl"), project, 128, 1)
		if err == nil || !strings.Contains(err.Error(), "symbolic links") {
			t.Fatalf("8.3 name hid real junction: %v", err)
		}
	})
	t.Run("project_header_redirect", func(t *testing.T) {
		link := filepath.Join(base, "Owned project junction")
		replayJunction(t, link, project)
		path := filepath.Join(root, "project-header.jsonl")
		header, _ := replayFixture(t, "codex", path, link)
		if _, _, err := ReplayNativeFile(context.Background(), "codex", root, path, project, 128, 1); err == nil {
			t.Fatal("junction project header accepted as a spelling alias")
		}
		if _, err := MapNativeRecord("codex", "fixture-header", project, header, 1); err == nil {
			t.Fatal("selected-record junction header accepted as a spelling alias")
		}
	})
}
