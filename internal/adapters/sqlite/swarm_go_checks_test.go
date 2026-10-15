package sqlite

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestSwarmActualFixedGoChecksApprovedPublicBaseline(t *testing.T) {
	ctx := context.Background()
	dir := swarmTestDirectory(t)
	root := filepath.Join(dir, "public-project")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(root, os.DirFS(filepath.Join("..", "..", "..", "tests", "s3", "testdata", "public-stats"))); err != nil {
		t.Fatal(err)
	}
	swarmGitTest(t, root, "init", "--initial-branch=main")
	baseline := swarmGitCommit(t, root)
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	checks, err := NewSwarmGoProjectChecks(root, baseline, store)
	if err != nil {
		t.Fatal(err)
	}
	result, err := checks.VerifyProjectChecks(ctx, domain.Lease{Worktree: root, BaseCommit: baseline, ExpiresAt: time.Now().Add(2 * time.Minute)}, domain.Artifact{Commit: baseline})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || result.ReceiptRef == "" {
		t.Fatal("actual correct baseline checks did not pass")
	}
	data, err := store.Read(ctx, result.ReceiptRef)
	if err != nil {
		t.Fatal(err)
	}
	var receipt SwarmProjectChecksReceipt
	if err = json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Baseline != baseline || receipt.Commit != baseline || len(receipt.Checks) != 2 {
		t.Fatal("fixed checks lost baseline or command provenance")
	}
	for _, r := range receipt.Checks {
		if !r.Passed || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) {
			t.Fatal("invalid actual check receipt")
		}
		if _, err = store.Read(ctx, r.StdoutRef); err != nil {
			t.Fatal(err)
		}
		if _, err = store.Read(ctx, r.StderrRef); err != nil {
			t.Fatal(err)
		}
	}
	other := filepath.Join(dir, "unapproved-project")
	if err = os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	swarmGitTest(t, other, "init", "--initial-branch=main")
	if err = os.WriteFile(filepath.Join(other, "README.md"), []byte("# Independent source\n"), 0600); err != nil {
		t.Fatal(err)
	}
	commit := swarmGitCommit(t, other)
	if _, err = checks.VerifyProjectChecks(ctx, domain.Lease{Worktree: other, ExpiresAt: time.Now().Add(time.Minute)}, domain.Artifact{Commit: commit}); err == nil {
		t.Fatal("fixed approved project checks executed another repository")
	}
}
