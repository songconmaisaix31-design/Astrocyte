package sqlite

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func swarmGitTest(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	out, err := swarmGit(context.Background(), root, args...)
	if err != nil {
		t.Fatalf("fixed local Git command %q failed: %v", args, err)
	}
	return out
}

func swarmGitCommit(t *testing.T, root string) string {
	t.Helper()
	swarmGitTest(t, root, "add", "--all")
	swarmGitTest(t, root, "-c", "user.name=Swarm verification", "-c", "user.email=swarm-verification@example.invalid", "commit", "-m", "Record local Swarm verification result")
	commit, err := swarmScalar(context.Background(), root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return commit
}

func TestSwarmActualGitWorktreeObjectsAndCommittedScope(t *testing.T) {
	ctx := context.Background()
	dir := swarmTestDirectory(t)
	root := filepath.Join(dir, "project")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	swarmGitTest(t, root, "init", "--initial-branch=main")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Swarm Git verification\n\nThis repository exercises actual worktree evidence.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	base := swarmGitCommit(t, root)
	wt := filepath.Join(dir, "worktree")
	swarmGitTest(t, root, "worktree", "add", "--detach", wt, base)
	canonical, err := swarmCanonical(wt)
	if err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := NewSwarmGitEvidence(store)
	if err != nil {
		t.Fatal(err)
	}
	actualBase, err := evidence.VerifyWorktree(ctx, root, canonical)
	if err != nil {
		t.Fatal(err)
	}
	if actualBase != base {
		t.Fatalf("base=%s actual=%s", base, actualBase)
	}
	if _, err = evidence.VerifyWorktree(ctx, root, root); err == nil {
		t.Fatal("primary project root accepted as independent worktree")
	}
	if err = os.WriteFile(filepath.Join(wt, "README.md"), []byte("# Swarm Git verification\n\nThis repository exercises actual worktree evidence.\n\nA committed improvement is independently verified.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = evidence.VerifyWorktree(ctx, root, canonical); err == nil {
		t.Fatal("dirty worktree was accepted for new claim")
	}
	commit := swarmGitCommit(t, wt)
	diff := swarmGitTest(t, wt, "diff", "--no-ext-diff", "--no-color", "--binary", base, commit, "--")
	diffRef, err := store.Publish(ctx, diff)
	if err != nil {
		t.Fatal(err)
	}
	check := swarmGitTest(t, wt, "diff", "--check", "--no-ext-diff", base, commit, "--")
	checkRef, err := store.Publish(ctx, check)
	if err != nil {
		t.Fatal(err)
	}
	checks := SwarmGitChecks{SchemaVersion: 1, BaseCommit: base, Commit: commit, DiffCheckPassed: true, DiffCheckOutputRef: checkRef}
	data, err := json.Marshal(checks)
	if err != nil {
		t.Fatal(err)
	}
	checksRef, err := store.Publish(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
	l := domain.Lease{Worktree: canonical, BaseCommit: base, WritePaths: []string{"README.md"}}
	a := domain.Artifact{Commit: commit, ObjectRef: diffRef, ChecksRef: checksRef, ChangedPaths: []string{"README.md"}, Inputs: []domain.SourceVersion{{ID: "git", Revision: base}}}
	verification, err := evidence.VerifyArtifact(ctx, l, a)
	if err != nil {
		t.Fatal(err)
	}
	if !verification.ChecksPassed {
		t.Fatal("actual git diff --check did not pass")
	}
	a.ObjectRef = checkRef
	if _, err = evidence.VerifyArtifact(ctx, l, a); err == nil {
		t.Fatal("published object unrelated to commit diff was accepted")
	}
	a.ObjectRef = diffRef
	a.ChangedPaths = []string{"other.md"}
	if _, err = evidence.VerifyArtifact(ctx, l, a); err == nil {
		t.Fatal("committed file omitted from submitted changed paths was accepted")
	}
	other := filepath.Join(dir, "other-project")
	if err = os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	swarmGitTest(t, other, "init", "--initial-branch=main")
	if err = os.WriteFile(filepath.Join(other, "README.md"), []byte("# Independent repository\n"), 0600); err != nil {
		t.Fatal(err)
	}
	swarmGitCommit(t, other)
	otherCanonical, err := swarmCanonical(other)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = evidence.VerifyWorktree(ctx, root, otherCanonical); err == nil {
		t.Fatal("unrelated Git repository was accepted")
	}
}
