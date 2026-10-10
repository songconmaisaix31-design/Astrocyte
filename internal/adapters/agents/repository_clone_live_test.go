package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Separate from real GitHub API acceptance: a successful Git operation must not
// turn an anonymous API rate limit into a successful metadata sync claim.
func TestManagedRepositoryPublicActualCloneAndRecovery(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_GITHUB_CLONE") != "1" {
		t.Skip("actual public clone opt-in disabled")
	}
	base, err := os.MkdirTemp("", "astrocyte-github-clone-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("retained actual checkout directory: %s", base)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	cloner := NewManagedRepositoryCloner(filepath.Join(base, "repositories"))
	actual, err := cloner.CloneRepository(ctx, "github-1", "https://github.com/steipete/summarize.git")
	if err != nil {
		t.Fatal("actual public Git clone failed", err)
	}
	if !gitObjectID.MatchString(actual.Head) || actual.Root == "" {
		t.Fatal("actual HEAD/root unavailable")
	}
	t.Logf("actual public Git HEAD=%s root=%s; github-1 is test target only, no API metadata asserted", actual.Head, actual.Root)
	before, err := os.Stat(filepath.Join(actual.Root, ".git", "HEAD"))
	if err != nil {
		t.Fatal(err)
	}
	// Reconstructed adapter uses the existing checkout; no fetch/checkout or
	// writes occur, so even HEAD mtime remains the original observation.
	again, err := NewManagedRepositoryCloner(filepath.Join(base, "repositories")).CloneRepository(ctx, "github-1", "https://github.com/steipete/summarize.git")
	if err != nil || again != actual {
		t.Fatal("actual recovery mismatch", err)
	}
	after, err := os.Stat(filepath.Join(actual.Root, ".git", "HEAD"))
	if err != nil || !after.ModTime().Equal(before.ModTime()) {
		t.Fatal("repeat checkout modified HEAD", err)
	}
	if _, err = cloner.CloneRepository(ctx, "github-1", "https://github.com/root-project/root.git"); err == nil {
		t.Fatal("existing different origin adopted")
	}
}
