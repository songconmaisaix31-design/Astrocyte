package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestV1RCommonDirRootAndArbitraryWorktreeName(t *testing.T) {
	// A disposable repository with the actual McPet identity and arbitrary
	// worktree spelling exercises common-dir recognition without private data.
	base := t.TempDir()
	root := filepath.Join(base, "mcpet")
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
	git("remote", "add", "origin", "https://github.com/songconmaisaix31-design/mcpet")
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
	git("remote", "set-url", "origin", "https://github.com/fixture/Astrocyte.git")
	if got := ObserveRepository(worktree); got.Topic != "" || got.Status != "repository_identity_conflict" {
		t.Fatal("conflicting Git identities assigned a topic")
	}
	if got := ObserveRepository(filepath.Join(base, "missing", "mcpet")); got.Topic != "" {
		t.Fatal("path spelling invented a McPet source")
	}
}

func TestV1RLocalMcPetCommonDirAndExistingWorktrees(t *testing.T) {
	if os.Getenv("ASTROCYTE_V1R_LOCAL_SOURCE") != "1" {
		t.Skip("opt-in authorized local McPet identity; no Git log capture")
	}
	root := "C:/Users/DW/orca/mcpet"
	original := ObserveRepository(root)
	common, err := os.Stat(original.CommonDir)
	if err != nil || original.Topic != "McPet" || original.Remote != "https://github.com/songconmaisaix31-design/mcpet" {
		t.Fatal("authorized local McPet identity mismatch")
	}
	data, err := exec.Command("git", "-C", root, "worktree", "list", "--porcelain").Output()
	if err != nil {
		t.Fatal("local worktree inventory failed")
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		observed := ObserveRepository(strings.TrimPrefix(line, "worktree "))
		info, err := os.Stat(observed.CommonDir)
		if err != nil || !os.SameFile(common, info) || observed.Topic != "McPet" {
			t.Fatal("existing McPet worktree common-dir mismatch")
		}
		count++
	}
	if count == 0 {
		t.Fatal("local McPet worktree inventory empty")
	}
	t.Logf("authorized local common-dir roots checked: %d", count)
}
