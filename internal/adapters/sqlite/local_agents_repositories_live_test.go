package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type persistedRepositorySpaces struct{ db *DB }

func (s persistedRepositorySpaces) ValidateRepositorySpace(ctx context.Context, c domain.Caller, id string) error {
	return NewAttentionRepository(s.db).WithTx(ctx, func(tx attention.AttentionTx) error { _, err := tx.LoadProjectSpace(id); return err })
}

type countedRepositoryCloner struct {
	inner workspace.RepositoryCloner
	calls int
}

func (c *countedRepositoryCloner) CloneRepository(ctx context.Context, id, url string) (domain.RepositoryClone, error) {
	c.calls++
	return c.inner.CloneRepository(ctx, id, url)
}

// Opt-in only: real anonymous metadata and clone of the user's public upstream;
// no models, media, personal browser or credential helpers participate.
func TestGitHubPublicActualCloneColdStoreAndRepeat(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_GITHUB_REPOSITORY") != "1" {
		t.Skip("actual public GitHub clone opt-in disabled")
	}
	base, err := os.MkdirTemp("", "astrocyte-github-public-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("retained evidence directory: %s", base)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	dbPath := filepath.Join(base, "state.sqlite")
	db := openAttentionDB(t, dbPath)
	if err = NewAttentionRepository(db).WithTx(ctx, func(tx attention.AttentionTx) error {
		return tx.SaveProjectSpace(attention.ProjectSpace{ID: "swarm-top", Version: 1, Title: "蜂群空间", MaterialRefs: []attention.SourceRef{}}, 0)
	}); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(base, "repositories")
	cloner := &countedRepositoryCloner{inner: agents.NewManagedRepositoryCloner(managed)}
	s := workspace.NewGitHubRepositoryService(db, db, agents.NewPublicGitHubSource(), cloner, persistedRepositorySpaces{db})
	human := domain.Caller{Kind: "human", ID: "authorized-public-verification"}
	items, err := s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "https://github.com/steipete/summarize"})
	if err != nil || len(items) != 1 {
		t.Fatalf("real metadata %+v %v", items, err)
	}
	if _, err = os.Stat(managed); !os.IsNotExist(err) {
		t.Fatal("real metadata created checkout directory")
	}
	item := items[0]
	t.Logf("metadata only id=%s revision=%d metadata_revision=%d full_name=%s source=%s unknown_fields=%v", item.ID, item.Revision, item.MetadataRevision, item.Metadata.FullName, item.Metadata.MetadataSource, item.Metadata.UnknownFields)
	placed, err := s.PlaceGitHubRepository(ctx, human, item.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: item.Revision})
	if err != nil {
		t.Fatalf("actual clone %+v %v", placed, err)
	}
	if placed.CloneStatus != "ready" || len(placed.Head) != 40 || cloner.calls != 1 {
		t.Fatal("actual checkout root/HEAD unavailable")
	}
	if _, err = os.Stat(filepath.Join(placed.Root, ".git", "HEAD")); err != nil {
		t.Fatal("actual Git HEAD absent", err)
	}
	project, err := db.LoadProject(ctx, placed.ProjectID)
	if err != nil || project.Settings.AllowDirectory || project.Settings.ExpandReferences || project.Settings.AllowAgentControl || project.Settings.ExternalModelCLI != "" || len(project.Settings.AllowedActions) != 0 {
		t.Fatal("actual clone widened permission", err)
	}
	t.Logf("actual checkout root=%s HEAD=%s attempts=%d", placed.Root, placed.Head, placed.CloneAttempts)
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	s = workspace.NewGitHubRepositoryService(db, db, agents.NewPublicGitHubSource(), cloner, persistedRepositorySpaces{db})
	loaded, err := s.ListGitHubRepositories(ctx, human)
	if err != nil || len(loaded) != 1 || loaded[0].Head != placed.Head || loaded[0].Root != placed.Root {
		t.Fatal("cold-store lost checkout", err)
	}
	if _, err = s.PlaceGitHubRepository(ctx, human, item.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: item.Revision}); err != nil || cloner.calls != 1 {
		t.Fatal("cold repeat re-cloned", err)
	}
	// Simulate loss between atomic publication and final row save. The adapter
	// verifies the existing checkout without fetch/checkout and preserves HEAD.
	recovering := loaded[0]
	recovering.Revision++
	recovering.CloneStatus = "cloning"
	if err = db.SaveGitHubRepository(ctx, recovering, loaded[0].Revision); err != nil {
		t.Fatal(err)
	}
	recovered, err := s.PlaceGitHubRepository(ctx, human, item.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: recovering.Revision})
	if err != nil || recovered.Head != placed.Head || recovered.Root != placed.Root || recovered.CloneAttempts != 2 {
		t.Fatal("published checkout recovery failed", err)
	}
	t.Log("cold restart/repeat and published-checkout recovery PASS; A-only, no model/control/native sessions")
}
