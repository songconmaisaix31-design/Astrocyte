package agents

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"golang.org/x/sys/windows"
)

func actualShortPath(t *testing.T, path string) string {
	t.Helper()
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(ptr, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || n >= uint32(len(buf)) {
		t.Skipf("NOT_RUN: real Windows 8.3 path unavailable: %v", err)
	}
	short := windows.UTF16ToString(buf[:n])
	long, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(short, long) {
		t.Skip("NOT_RUN: fixture volume has no usable Windows 8.3 alias")
	}
	t.Logf("actual filesystem alias: %s -> %s", short, long)
	return short
}

func TestWindowsActualShortPathScopes(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "approved project directory")
	history := filepath.Join(base, "fixture metadata directory")
	for _, dir := range []string{filepath.Join(root, "allowed"), filepath.Join(root, "other"), history} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"package.json", "allowed/a.txt", "other/b.txt", "allowed/.env"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte("fixture only"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	shortRoot, shortHistory := actualShortPath(t, root), actualShortPath(t, history)
	longRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("identity_and_scope", func(t *testing.T) {
		if rootKey(shortRoot) != rootKey(longRoot) || !inside(shortRoot, filepath.Join(longRoot, "allowed", "a.txt")) || !inside(longRoot, filepath.Join(shortRoot, "allowed", "a.txt")) {
			t.Fatal("real 8.3 alias is treated as a different approved directory")
		}
		p := domain.LocalProject{Root: shortRoot, Settings: domain.ProjectSettings{AllowDirectory: true, AllowedSubdirs: []string{"allowed"}}}
		files, err := ReadProjectFiles(context.Background(), p, []string{"allowed/a.txt"})
		if err != nil || len(files) != 1 || files[0].Text != "fixture only" {
			t.Fatalf("approved alias read: %+v %v", files, err)
		}
		for _, name := range []string{"other/b.txt", "../outside.txt", "allowed/.env"} {
			if _, err := ReadProjectFiles(context.Background(), p, []string{name}); err == nil {
				t.Fatalf("unapproved path accepted: %s", name)
			}
		}
	})
	t.Run("bounded_metadata_and_registration", func(t *testing.T) {
		for i, cwd := range []string{shortRoot, longRoot} {
			data, _ := json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": "same-session", "cwd": cwd}})
			if err := os.WriteFile(filepath.Join(history, []string{"short.jsonl", "long.jsonl"}[i]), append(data, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
		}
		s := registeredFixture(t, []map[string]any{{"id": "repo1", "path": shortRoot}, {"id": "repo1", "path": longRoot}}, []map[string]any{})
		s.metadataSources = func() []projectMetadataSource { return []projectMetadataSource{{"codex", shortHistory, 3}} }
		result, err := s.DiscoverRegistered(context.Background())
		if err != nil || len(result.Projects) != 1 || len(result.Projects[0].Contributors) != 1 || result.Sources[0].HeadersExamined != 2 {
			t.Fatalf("short/long registration and header dedup: %+v %v", result, err)
		}
		snapshot := domain.ProjectDiscoverySnapshot{Status: "complete"}
		s.observeProjectMetadata(context.Background(), &snapshot)
		if len(snapshot.Projects) != 1 || rootKey(snapshot.Projects[0].Root) != rootKey(longRoot) || len(snapshot.Projects[0].Contributors) != 1 {
			t.Fatalf("standalone short cwd marker: %+v", snapshot)
		}
		missingShort := filepath.Join(shortRoot, "removed-checkout")
		missingLong := filepath.Join(longRoot, "removed-checkout")
		if rootKey(missingShort) != rootKey(missingLong) || !inside(shortRoot, missingLong) {
			t.Fatal("missing metadata path loses existing ancestor alias")
		}
		s = registeredFixture(t, []map[string]any{{"id": "repo1", "path": missingShort}}, []map[string]any{{"repoId": "repo1", "hostId": "local", "path": longRoot}})
		result, err = s.DiscoverRegistered(context.Background())
		if err != nil || len(result.Projects) != 1 || rootKey(result.Projects[0].ParentRoot) != rootKey(missingLong) || len(result.Failures) != 1 {
			t.Fatalf("missing parent association: %+v %v", result, err)
		}
	})
	t.Run("secret_aliases", func(t *testing.T) {
		secret := filepath.Join(root, ".codex")
		if err := os.Mkdir(secret, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(secret, "package.json"), []byte("fixture only"), 0600); err != nil {
			t.Fatal(err)
		}
		credential := filepath.Join(root, "credentials.json")
		if err := os.WriteFile(credential, []byte("fixture only"), 0600); err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{filepath.Join(secret, "package.json"), credential} {
			short := actualShortPath(t, path)
			rel, err := filepath.Rel(shortRoot, short)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ValidateProjectPath(shortRoot, rel); err == nil {
				t.Fatalf("short secret name accepted: %s", rel)
			}
		}
		if _, err := metadataProjectRoot(actualShortPath(t, secret)); err == nil {
			t.Fatal("native metadata alias treated as project root")
		}
	})
}

func TestWindowsRedirectedPathsRemainDenied(t *testing.T) {
	base := t.TempDir()
	root, target := filepath.Join(base, "approved"), filepath.Join(base, "target")
	for _, dir := range []string{root, filepath.Join(target, "history"), filepath.Join(target, "project")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"project/package.json", "project/fixture.txt"} {
		if err := os.WriteFile(filepath.Join(target, filepath.FromSlash(name)), []byte("fixture only"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(root, "redirect")
	if err := os.Symlink(target, link); err != nil {
		t.Logf("NOT_RUN: actual symlink creation unavailable (%v); checking real junction instead", err)
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
			t.Skipf("NOT_RUN: junction creation unavailable: %v %s", err, out)
		}
	}
	if rootKey(link) == rootKey(target) {
		t.Fatal("8.3 spelling normalization resolved a filesystem redirect")
	}
	if _, err := ValidateProjectPath(root, "redirect/project/fixture.txt"); err == nil {
		t.Fatal("redirected relative file accepted")
	}
	if _, err := metadataProjectRoot(filepath.Join(link, "project")); err == nil {
		t.Fatal("redirected project ancestor accepted")
	}
	for _, source := range []string{link, filepath.Join(link, "history")} {
		s := &RegisteredProjects{metadataSources: func() []projectMetadataSource { return []projectMetadataSource{{"codex", source, 3}} }}
		snapshot := domain.ProjectDiscoverySnapshot{Status: "complete"}
		s.observeProjectMetadata(context.Background(), &snapshot)
		if snapshot.Sources[0].Reason != "known_source_unavailable_or_redirected" || snapshot.Sources[0].EntriesExamined != 0 {
			t.Fatalf("redirected source or ancestor accepted: %+v", snapshot)
		}
	}
}
