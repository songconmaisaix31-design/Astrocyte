package app

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type gitHubRepositoryService struct {
	gate     chan struct{}
	store    GitHubRepositoryStore
	projects LocalProjectRepository
	source   PublicGitHubSource
	cloner   RepositoryCloner
	spaces   RepositorySpaces
}

func NewGitHubRepositoryService(store GitHubRepositoryStore, projects LocalProjectRepository, source PublicGitHubSource, cloner RepositoryCloner, spaces RepositorySpaces) *gitHubRepositoryService {
	return &gitHubRepositoryService{gate: make(chan struct{}, 1), store: store, projects: projects, source: source, cloner: cloner, spaces: spaces}
}

func notFound(err error) bool {
	var e *apierrors.ServiceError
	return errors.As(err, &e) && e.Code == apierrors.NotFound
}

func (s *gitHubRepositoryService) ListGitHubRepositories(ctx context.Context, c domain.Caller) ([]domain.GitHubRepository, error) {
	if err := human(c); err != nil {
		return nil, err
	}
	return s.store.ListGitHubRepositories(ctx)
}

func (s *gitHubRepositoryService) enter(ctx context.Context) (func(), error) {
	select {
	case s.gate <- struct{}{}:
		return func() { <-s.gate }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *gitHubRepositoryService) SyncGitHubRepositories(ctx context.Context, c domain.Caller, cmd domain.RepositorySyncCommand) ([]domain.GitHubRepository, error) {
	if err := human(c); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	exit, err := s.enter(ctx)
	if err != nil {
		return nil, err
	}
	defer exit()
	if s.source == nil {
		return nil, projectError(apierrors.UnsupportedCapability, "public repository metadata source is not connected")
	}
	metadata, err := s.source.FetchPublicRepositories(ctx, cmd.Input)
	if err != nil {
		// Preserve the last successful observation and mark it stale, rather
		// than replacing it with an invented empty account or fresh timestamp.
		var e *apierrors.ServiceError
		if errors.As(err, &e) && e.Code == apierrors.ValidationFailed {
			return nil, err
		}
		input := strings.TrimSuffix(strings.Trim(strings.TrimPrefix(strings.TrimSpace(cmd.Input), "https://github.com/"), "/"), ".git")
		old, loadErr := s.store.ListGitHubRepositories(ctx)
		if loadErr != nil {
			return nil, errors.Join(err, loadErr)
		}
		for _, item := range old {
			if !strings.EqualFold(input, item.Metadata.FullName) && !strings.EqualFold(input, strings.Split(item.Metadata.FullName, "/")[0]) {
				continue
			}
			expected := item.Revision
			item.Revision++
			item.SyncStatus = "stale"
			item.LastError = "public_metadata_sync_failed"
			if saveErr := s.store.SaveGitHubRepository(ctx, item, expected); saveErr != nil {
				return nil, errors.Join(err, saveErr)
			}
		}
		return nil, err
	}
	if len(metadata) > 100 {
		return nil, projectError(apierrors.BudgetExhausted, "public metadata page exceeds 100 repositories")
	}
	result := []domain.GitHubRepository{}
	seen := map[int64]bool{}
	for _, m := range metadata {
		m = domain.NormalizeGitHubMetadata(m)
		if m.GitHubID <= 0 || len(strings.Split(m.FullName, "/")) != 2 || m.CloneURL != "https://github.com/"+m.FullName+".git" || m.HTMLURL != "https://github.com/"+m.FullName {
			return nil, projectError(apierrors.EvidenceMissing, "public source returned invalid metadata")
		}
		if seen[m.GitHubID] {
			continue
		}
		seen[m.GitHubID] = true
		id := fmt.Sprintf("github-%d", m.GitHubID)
		item, err := s.store.LoadGitHubRepository(ctx, id)
		if err != nil && !notFound(err) {
			return nil, err
		}
		if notFound(err) {
			item = domain.GitHubRepository{ID: id, CloneStatus: "not_placed"}
		}
		expected := item.Revision
		if item.MetadataRevision == 0 || !reflect.DeepEqual(item.Metadata, m) {
			item.MetadataRevision++
		}
		item.Metadata = m
		item.Revision++
		item.SyncStatus = "synced"
		item.SyncedAt = time.Now().UTC()
		if item.LastError == "public_metadata_sync_failed" {
			item.LastError = ""
		}
		if err = s.store.SaveGitHubRepository(ctx, item, expected); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *gitHubRepositoryService) PlaceGitHubRepository(ctx context.Context, c domain.Caller, id string, cmd domain.RepositoryPlacementCommand) (domain.GitHubRepository, error) {
	var item domain.GitHubRepository
	if err := human(c); err != nil {
		return item, err
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	exit, err := s.enter(ctx)
	if err != nil {
		return item, err
	}
	defer exit()
	if s.cloner == nil || s.spaces == nil {
		return item, projectError(apierrors.UnsupportedCapability, "repository checkout or top-level space bridge is not connected")
	}
	if cmd.SpaceID == "" {
		return item, projectError(apierrors.ValidationFailed, "human placement requires an existing top-level space")
	}
	// The existing Attention bridge verifies space existence/access. A folder,
	// ranking change or Agent request cannot substitute for this human action.
	if err = s.spaces.ValidateRepositorySpace(ctx, c, cmd.SpaceID); err != nil {
		return item, err
	}
	item, err = s.store.LoadGitHubRepository(ctx, id)
	if err != nil {
		return item, err
	}
	if item.SpaceID != "" && item.SpaceID != cmd.SpaceID {
		return item, projectError(apierrors.ScopeDenied, "repository was already placed into another space")
	}
	if item.CloneStatus == "ready" {
		// A repeated placement remains successful even after the transport's old
		// revision. It never clones again or rewrites a project's permissions.
		p, err := s.projects.LoadProject(ctx, item.ProjectID)
		if err != nil {
			return item, err
		}
		if p.SpaceID != cmd.SpaceID || p.Root != item.Root {
			return item, projectError(apierrors.EvidenceMissing, "placed project differs from its checkout")
		}
		return item, nil
	}
	if cmd.ExpectedRevision != item.Revision {
		return item, projectError(apierrors.VersionConflict, "repository observation changed; reload before placement")
	}
	if item.CloneAttempts >= 3 {
		return item, projectError(apierrors.BudgetExhausted, "repository checkout retry limit reached; inspect retained checkout stages")
	}
	expected := item.Revision
	item.Revision++
	item.SpaceID = cmd.SpaceID
	item.ProjectID = id
	item.CloneStatus = "cloning"
	item.CloneAttempts++
	item.LastError = ""
	if err = s.store.SaveGitHubRepository(ctx, item, expected); err != nil {
		return item, err
	}
	clone, err := s.cloner.CloneRepository(ctx, id, item.Metadata.CloneURL)
	if err != nil {
		expected = item.Revision
		item.Revision++
		item.CloneStatus = "failed"
		item.LastError = "checkout_failed_review_or_retry"
		if ctx.Err() != nil {
			item.CloneStatus = "interrupted"
			item.LastError = "checkout_interrupted_human_retry_required"
		}
		// Persist the known failure after cancellation as well; process death
		// instead leaves the durable cloning row for an explicit recovery call.
		saveCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		// Keep the safe ServiceError first for HTTP errors.As extraction, while
		// preserving the clone diagnostic and any persistence failure for callers.
		return item, errors.Join(&apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "public repository checkout did not complete", Retryable: item.CloneAttempts < 3, RequiredAction: item.LastError}, err, s.store.SaveGitHubRepository(saveCtx, item, expected))
	}
	if clone.Root == "" || clone.Head == "" {
		return item, projectError(apierrors.EvidenceMissing, "checkout returned no actual root or HEAD")
	}
	p, err := s.projects.LoadProject(ctx, id)
	if err != nil && !notFound(err) {
		return item, err
	}
	if notFound(err) {
		p = domain.LocalProject{ID: id, Name: item.Metadata.FullName, Root: clone.Root, SpaceID: cmd.SpaceID, CreatedAt: time.Now().UTC(), Settings: domain.ProjectSettings{Revision: 1, HistoryRoots: map[string]string{}, AllowedSubdirs: []string{}, AllowedActions: []string{}, AllowedTools: []string{}}}
		if err = s.projects.SaveProject(ctx, p, 0); err != nil {
			return item, err
		}
	} else if p.Root != clone.Root || p.SpaceID != cmd.SpaceID {
		return item, projectError(apierrors.ScopeDenied, "existing project differs from this placement")
	}
	// A crash after rename/project save is recoverable: the adapter verifies the
	// same root/origin/HEAD and this code reuses the project without any grants.
	expected = item.Revision
	item.Revision++
	item.CloneStatus = "ready"
	item.Root = clone.Root
	item.Head = clone.Head
	item.LastError = ""
	err = s.store.SaveGitHubRepository(ctx, item, expected)
	return item, err
}
