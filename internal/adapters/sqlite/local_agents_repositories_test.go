package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type publicMetadataFixture struct {
	calls    int
	err      error
	metadata []domain.GitHubMetadata
}

func (s *publicMetadataFixture) FetchPublicRepositories(context.Context, string) ([]domain.GitHubMetadata, error) {
	s.calls++
	return s.metadata, s.err
}

type placementSpaces struct{ calls int }

func (s *placementSpaces) ValidateRepositorySpace(_ context.Context, _ domain.Caller, id string) error {
	s.calls++
	if id != "swarm-top" {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	return nil
}

type checkoutFixture struct {
	calls int
	err   error
	base  string
}

func (s *checkoutFixture) CloneRepository(_ context.Context, id, _ string) (domain.RepositoryClone, error) {
	s.calls++
	if s.err != nil {
		return domain.RepositoryClone{}, s.err
	}
	root := filepath.Join(s.base, id)
	err := os.MkdirAll(root, 0700)
	return domain.RepositoryClone{Root: root, Head: "0123456789012345678901234567890123456789"}, err
}
func repositoryMetadata() domain.GitHubMetadata {
	return domain.GitHubMetadata{GitHubID: 123, FullName: "steipete/summarize", HTMLURL: "https://github.com/steipete/summarize", CloneURL: "https://github.com/steipete/summarize.git"}
}

func TestGitHubMetadataNoCheckoutPlacementRestartRepeatAndPermissions(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	source := &publicMetadataFixture{metadata: []domain.GitHubMetadata{repositoryMetadata(), repositoryMetadata()}}
	cloner := &checkoutFixture{base: filepath.Join(t.TempDir(), "managed")}
	spaces := &placementSpaces{}
	s := workspace.NewGitHubRepositoryService(db, db, source, cloner, spaces)
	human := domain.Caller{Kind: "human", ID: "browser-human"}
	for _, caller := range []domain.Caller{{}, {Kind: "human"}, {Kind: "agent", ID: "a", ProjectID: "github-123"}} {
		if _, err := s.SyncGitHubRepositories(ctx, caller, domain.RepositorySyncCommand{Input: "steipete/summarize"}); err == nil {
			t.Fatal("nonhuman sync accepted")
		}
		if _, err := s.PlaceGitHubRepository(ctx, caller, "github-123", domain.RepositoryPlacementCommand{SpaceID: "swarm-top"}); err == nil {
			t.Fatal("nonhuman placement accepted")
		}
		if _, err := s.ListGitHubRepositories(ctx, caller); err == nil {
			t.Fatal("nonhuman list accepted")
		}
	}
	if source.calls != 0 || spaces.calls != 0 || cloner.calls != 0 {
		t.Fatal("denied caller performed work")
	}
	items, err := s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"})
	if err != nil || len(items) != 1 || items[0].MetadataRevision != 1 || items[0].CloneStatus != "not_placed" {
		t.Fatalf("sync %+v %v", items, err)
	}
	first := items[0]
	if _, err = os.Stat(cloner.base); !os.IsNotExist(err) {
		t.Fatal("metadata wrote code directory")
	}
	projects, err := db.ListProjects(ctx)
	if err != nil || len(projects) != 0 {
		t.Fatal("metadata created native project")
	}
	items, err = s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"})
	if err != nil || items[0].MetadataRevision != 1 || items[0].Revision <= first.Revision {
		t.Fatal("repeat metadata created another content revision", err)
	}
	if _, err = s.PlaceGitHubRepository(ctx, human, "github-123", domain.RepositoryPlacementCommand{SpaceID: "folder", ExpectedRevision: items[0].Revision}); err == nil || cloner.calls != 0 {
		t.Fatal("invalid space cloned")
	}
	if _, err = s.PlaceGitHubRepository(ctx, human, "github-123", domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: first.Revision}); err == nil || cloner.calls != 0 {
		t.Fatal("stale first placement cloned")
	}
	placed, err := s.PlaceGitHubRepository(ctx, human, "github-123", domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: items[0].Revision})
	if err != nil || placed.CloneStatus != "ready" || placed.Root == "" || placed.Head == "" || placed.CloneAttempts != 1 || cloner.calls != 1 {
		t.Fatalf("placement %+v %v", placed, err)
	}
	p, err := db.LoadProject(ctx, placed.ProjectID)
	if err != nil || p.Settings.AllowDirectory || p.Settings.ExpandReferences || p.Settings.ExternalModelCLI != "" || p.Settings.AllowAgentControl || len(p.Settings.AllowedActions) != 0 {
		t.Fatal("clone widened permissions", err)
	}
	native := &localNative{}
	locals := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{native})
	if _, err = locals.ReadProjectContext(ctx, human, p.ID, domain.ContextRequest{Files: []string{"README.md"}}); err == nil {
		t.Fatal("clone implicitly granted B")
	}
	if _, err = locals.StartNativeSession(ctx, human, p.ID, domain.NativeCommand{CLI: "codex"}); err == nil {
		t.Fatal("clone implicitly granted model/start")
	}
	if _, err = locals.ReadProjectContext(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: p.ID}, p.ID, domain.ContextRequest{}); err == nil {
		t.Fatal("clone implicitly granted Agent identity")
	}
	if native.starts != 0 {
		t.Fatal("clone launched native Agent")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	defer db.Close()
	s = workspace.NewGitHubRepositoryService(db, db, source, cloner, spaces)
	loaded, err := s.ListGitHubRepositories(ctx, human)
	if err != nil || len(loaded) != 1 || loaded[0].Root != placed.Root || loaded[0].Head != placed.Head || source.calls != 2 {
		t.Fatal("restart GET lost actual checkout or executed source", err)
	}
	again, err := s.PlaceGitHubRepository(ctx, human, p.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: first.Revision})
	if err != nil || again.Revision != placed.Revision || cloner.calls != 1 {
		t.Fatal("repeat placement re-cloned or mutated permissions", err)
	}
	source.err = errors.New("public API failed")
	if _, err = s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"}); err == nil {
		t.Fatal("metadata failure reported success")
	}
	stale, err := db.LoadGitHubRepository(ctx, p.ID)
	if err != nil || stale.SyncStatus != "stale" || stale.CloneStatus != "ready" || !stale.SyncedAt.Equal(loaded[0].SyncedAt) || stale.Root != placed.Root {
		t.Fatal("failure erased successful observation", err)
	}
}

func TestGitHubPlacementFailureExplicitBoundedRetryAndInterruptedRecovery(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	source := &publicMetadataFixture{metadata: []domain.GitHubMetadata{repositoryMetadata()}}
	cloner := &checkoutFixture{base: t.TempDir(), err: errors.New("interrupted")}
	spaces := &placementSpaces{}
	s := workspace.NewGitHubRepositoryService(db, db, source, cloner, spaces)
	human := domain.Caller{Kind: "human", ID: "browser-human"}
	items, err := s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := s.PlaceGitHubRepository(ctx, human, items[0].ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: items[0].Revision})
	if err == nil || failed.CloneStatus != "failed" || failed.CloneAttempts != 1 {
		t.Fatal("first failure missing")
	}
	projects, _ := db.ListProjects(ctx)
	if len(projects) != 0 {
		t.Fatal("failed clone granted project")
	}
	// Simulate process loss after the durable cloning state. An explicit human
	// call after reconstructing the service can resume; GET never retries it.
	interrupted := failed
	interrupted.Revision++
	interrupted.CloneStatus = "cloning"
	if err = db.SaveGitHubRepository(ctx, interrupted, failed.Revision); err != nil {
		t.Fatal(err)
	}
	s = workspace.NewGitHubRepositoryService(db, db, source, cloner, spaces)
	cached, err := s.ListGitHubRepositories(ctx, human)
	if err != nil || cached[0].CloneStatus != "cloning" || cloner.calls != 1 {
		t.Fatal("GET retried interrupted clone")
	}
	cloner.err = nil
	ready, err := s.PlaceGitHubRepository(ctx, human, interrupted.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: interrupted.Revision})
	if err != nil || ready.CloneStatus != "ready" || ready.CloneAttempts != 2 {
		t.Fatal("explicit interrupted recovery failed", err)
	}
	// A separate repository proves the three-attempt bound without laundering
	// exhausted status through metadata re-sync.
	m := repositoryMetadata()
	m.GitHubID = 124
	source.metadata = []domain.GitHubMetadata{m}
	cloner.err = errors.New("network fails")
	items, err = s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"})
	if err != nil {
		t.Fatal(err)
	}
	item := items[0]
	for i := 0; i < 3; i++ {
		item, err = s.PlaceGitHubRepository(ctx, human, item.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: item.Revision})
		if err == nil {
			t.Fatal("failed attempt accepted")
		}
	}
	before := cloner.calls
	if _, err = s.PlaceGitHubRepository(ctx, human, item.ID, domain.RepositoryPlacementCommand{SpaceID: "swarm-top", ExpectedRevision: item.Revision}); err == nil || cloner.calls != before {
		t.Fatal("clone retry bound bypassed")
	}
	items, err = s.SyncGitHubRepositories(ctx, human, domain.RepositorySyncCommand{Input: "steipete/summarize"})
	if err != nil || items[0].CloneAttempts != 3 {
		t.Fatal("metadata reset attempts", err)
	}
}
