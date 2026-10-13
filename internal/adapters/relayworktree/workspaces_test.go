package relayworktree

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func testGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	if err := cmd.Run(); err != nil {
		t.Fatalf("public test Git command failed: %v", err)
	}
}
func cleanRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	testGit(t, root, "init", "-q")
	if e := os.WriteFile(filepath.Join(root, "public.txt"), []byte("committed public baseline\n"), 0600); e != nil {
		t.Fatal(e)
	}
	testGit(t, root, "add", "public.txt")
	testGit(t, root, "-c", "user.name=Public Test", "-c", "user.email=public-test@example.invalid", "commit", "-qm", "public baseline")
	return root
}

func TestNeutralWorkspacePrecedesAgentAndReusePreservesChanges(t *testing.T) {
	source := cleanRepo(t)
	service := New(t.TempDir(), Options{AllowCommittedGit: true})
	work := domain.Work{Root: source, Kind: domain.WorkKindFolder}
	id := uuid.NewString()
	w, e := service.Prepare(context.Background(), work, domain.RelayWorkspace{ID: id})
	if e != nil {
		t.Fatal(e)
	}
	if !w.Prepared || w.Branch != "relay/"+id || strings.Contains(w.Root, "codex") || strings.Contains(w.Root, "claude") {
		t.Fatal("workspace derived from an Agent")
	}
	baseline, e := service.Observe(context.Background(), w)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(w.Root, "public.txt"), []byte("first unfinished edit\n"), 0600); e != nil {
		t.Fatal(e)
	}
	first, e := service.Observe(context.Background(), w)
	if e != nil || first == baseline {
		t.Fatal("tracked edit not observed")
	}
	if e = os.WriteFile(filepath.Join(w.Root, "public.txt"), []byte("second unfinished edit\n"), 0600); e != nil {
		t.Fatal(e)
	}
	second, e := service.Observe(context.Background(), w)
	if e != nil || second == first {
		t.Fatal("continued dirty edit not observed")
	}
	if e = os.WriteFile(filepath.Join(w.Root, "new-public.txt"), []byte("untracked relay state\n"), 0600); e != nil {
		t.Fatal(e)
	}
	again, e := service.Prepare(context.Background(), work, w)
	if e != nil {
		t.Fatal(e)
	}
	if again.Root != w.Root || again.ID != w.ID {
		t.Fatal("handoff rebuilt neutral workspace")
	}
	data, e := os.ReadFile(filepath.Join(w.Root, "new-public.txt"))
	if e != nil || string(data) != "untracked relay state\n" {
		t.Fatal("handoff discarded untracked relay file")
	}
	sourceBytes, e := os.ReadFile(filepath.Join(source, "public.txt"))
	if e != nil || string(sourceBytes) != "committed public baseline\n" {
		t.Fatal("neutral workspace changed source tree")
	}
}

func TestNeutralWorkspacePendingScopeDoesNotDiscardSource(t *testing.T) {
	source := cleanRepo(t)
	work := domain.Work{Root: source, Kind: domain.WorkKindFolder}
	service := New(t.TempDir(), Options{})
	if _, e := service.Prepare(context.Background(), work, domain.RelayWorkspace{ID: uuid.NewString()}); e == nil {
		t.Fatal("unconfigured scope silently chosen")
	}
	service = New(t.TempDir(), Options{AllowCommittedGit: true})
	if e := os.WriteFile(filepath.Join(source, "private-untracked.txt"), []byte("preserve source"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := service.Prepare(context.Background(), work, domain.RelayWorkspace{ID: uuid.NewString()}); e == nil {
		t.Fatal("undecided untracked source silently excluded")
	}
	if data, e := os.ReadFile(filepath.Join(source, "private-untracked.txt")); e != nil || string(data) != "preserve source" {
		t.Fatal("source data lost")
	}
	work.Kind = domain.WorkKindFile
	if _, e := service.Prepare(context.Background(), work, domain.RelayWorkspace{ID: uuid.NewString()}); e == nil {
		t.Fatal("undecided single-file scope implemented")
	}
}
