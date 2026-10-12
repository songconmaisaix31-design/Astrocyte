package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestV1RCommonDirRootAndArbitraryWorktreeName(t *testing.T) {
	// A real disposable Git repository/worktree tests the existing authorized
	// maimai-monster identity; no actual candidate McPet source is read here.
	base := t.TempDir()
	root := filepath.Join(base, "maimai-monster")
	worktree := filepath.Join(base, "not-a-project-name")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		if b, err := command.CombinedOutput(); err != nil {
			t.Fatalf("synthetic Git setup failed: %v %s", err, b)
		}
	}
	git("init", "-q")
	git("-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "commit", "--allow-empty", "-m", "synthetic fixture")
	git("worktree", "add", "--detach", worktree, "HEAD")
	mapper := NewNativeSemanticMapper()
	a, b := mapper.Repository(root), mapper.Repository(worktree)
	left, x := os.Stat(a.CommonDir)
	right, y := os.Stat(b.CommonDir)
	if a.Topic != "McPet" || b.Topic != "McPet" || x != nil || y != nil || !os.SameFile(left, right) || a.Status != "observed_recognized" || b.Status != "observed_recognized" {
		t.Fatalf("worktree did not reuse actual common-dir identity: %s/%s", a.Topic, b.Topic)
	}
	// Neither ordinary cwd spelling nor an unavailable McPet-looking path is
	// authority, and conflicting known remote/common identities stay unresolved.
	git("remote", "add", "origin", "https://github.com/fixture/Astrocyte.git")
	if got := ObserveRepository(worktree); got.Topic != "" || got.Status != "repository_identity_conflict" {
		t.Fatal("conflicting Git identities assigned a topic")
	}
	if got := ObserveRepository(filepath.Join(base, "missing", "mcpet")); got.Topic != "" {
		t.Fatal("path spelling invented a McPet source")
	}
}
