package repositorycontext

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type IndexStore interface {
	LoadRepositoryIndex(context.Context, string, string, string) (domain.RepositoryIndex, error)
	SaveRepositoryIndex(context.Context, domain.RepositoryIndex, string) error
}

type Options struct {
	Git            Git
	Graph          CodeGraph
	AOCI           string
	CacheRoot      string
	SnapshotBudget domain.RepositoryBudget
}

type Service struct {
	mu       sync.Mutex
	projects app.RepositoryProjectReader
	indexes  IndexStore
	options  Options
}

func New(projects app.RepositoryProjectReader, indexes IndexStore, options Options) (*Service, error) {
	if projects == nil || indexes == nil || options.CacheRoot == "" || options.Git.EnvironmentID == "" {
		return nil, failure(apierrors.ValidationFailed, "repository context configuration is incomplete")
	}
	if err := validateBudget(options.SnapshotBudget); err != nil {
		return nil, err
	}
	cache, err := filepath.Abs(options.CacheRoot)
	if err != nil {
		return nil, err
	}
	options.CacheRoot = cache
	return &Service{projects: projects, indexes: indexes, options: options}, nil
}

func relativeRoots(p domain.LocalProject, roots []string) ([]string, error) {
	result := []string{}
	for _, value := range roots {
		if filepath.IsAbs(value) {
			rel, err := filepath.Rel(p.Root, value)
			if err != nil {
				return nil, err
			}
			value = filepath.ToSlash(rel)
		}
		if !safeRelative(value) || privatePath(value) {
			return nil, failure(apierrors.ScopeDenied, "repository root is outside the project")
		}
		result = append(result, value)
	}
	slices.Sort(result)
	return slices.Compact(result), nil
}

func (s *Service) authorize(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.LocalProject, domain.RepositoryQuery, error) {
	p, err := s.projects.LoadRepositoryProject(ctx, q.ProjectID)
	if err != nil {
		return p, q, err
	}
	if c.ID == "" || (c.Kind != "human" && (c.Kind != "agent" || c.ProjectID != p.ID)) {
		return p, q, failure(apierrors.ScopeDenied, "caller has no project identity")
	}
	if c.Kind == "agent" {
		grant, err := s.projects.LoadRepositoryGrant(ctx, p.ID, c.ID)
		if err != nil {
			return p, q, err
		}
		if grant.ProjectID != p.ID || grant.AgentID != c.ID || grant.RevokedAt != nil || grant.ExpiresAt != nil && !grant.ExpiresAt.After(time.Now()) {
			return p, q, failure(apierrors.ApprovalRevoked, "project authorization is revoked or expired")
		}
		if !slices.Contains(grant.Actions, "read_context") {
			return p, q, failure(apierrors.ScopeDenied, "project authorization does not permit context access")
		}
	}
	if !p.Settings.AllowDirectory || len(p.Settings.AllowedSubdirs) == 0 {
		return p, q, failure(apierrors.ScopeDenied, "project directory access is disabled")
	}
	if strings.TrimSpace(q.Purpose) == "" || len(q.Purpose) > 512 {
		return p, q, failure(apierrors.ValidationFailed, "a bounded query purpose is required")
	}
	if err := validateBudget(q.Budget); err != nil {
		return p, q, err
	}
	limit := s.options.SnapshotBudget
	if q.Budget.MaxFiles > limit.MaxFiles || q.Budget.MaxBytes > limit.MaxBytes || q.Budget.MaxResults > limit.MaxResults || q.Budget.MaxDepth > limit.MaxDepth {
		return p, q, failure(apierrors.BudgetExhausted, "requested repository budget exceeds configured limits")
	}
	q.Roots, err = relativeRoots(p, q.Roots)
	if err != nil {
		return p, q, err
	}
	allowed, err := relativeRoots(p, p.Settings.AllowedSubdirs)
	if err != nil {
		return p, q, err
	}
	if len(q.Roots) == 0 {
		return p, q, failure(apierrors.ScopeDenied, "repository roots are required")
	}
	for _, root := range q.Roots {
		if !under(root, allowed) {
			return p, q, failure(apierrors.ScopeDenied, "query root exceeds current project authorization")
		}
	}
	return p, q, nil
}

func (s *Service) capture(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.RepositoryVersion, domain.RepositoryQuery, error) {
	p, q, err := s.authorize(ctx, c, q)
	if err != nil {
		return domain.RepositoryVersion{}, q, err
	}
	v, err := s.options.Git.Capture(ctx, p, q.Roots, q.Budget)
	return v, q, err
}

func (s *Service) projection(index domain.RepositoryIndex) (string, error) {
	if _, err := uuid.Parse(index.Generation); err != nil {
		return "", failure(apierrors.IndexStale, "index generation is invalid")
	}
	if len(index.Snapshot.WorktreeID) != 64 || strings.ContainsAny(index.Snapshot.WorktreeID, "/\\.") {
		return "", failure(apierrors.IndexStale, "index worktree identity is invalid")
	}
	return filepath.Join(s.options.CacheRoot, index.Snapshot.WorktreeID, index.Generation), nil
}

func (s *Service) UpdateRepositoryIndex(ctx context.Context, c domain.Caller, q domain.RepositoryQuery, action string) (domain.RepositoryIndex, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result domain.RepositoryIndex
	if c.Kind != "human" || c.ID == "" {
		return result, failure(apierrors.ScopeDenied, "explicit index maintenance requires the human control plane")
	}
	if !slices.Contains([]string{"index", "sync", "rebuild"}, action) {
		return result, failure(apierrors.ValidationFailed, "unknown index maintenance action")
	}
	before, q, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	previous, loadErr := s.indexes.LoadRepositoryIndex(ctx, q.ProjectID, before.WorktreeID, "codegraph")
	expected := ""
	if loadErr == nil {
		expected = previous.Generation
	} else {
		var serviceErr *apierrors.ServiceError
		if !errors.As(loadErr, &serviceErr) || serviceErr.Code != apierrors.NotFound {
			return result, loadErr
		}
	}
	if action == "index" && expected != "" {
		return result, failure(apierrors.VersionConflict, "an index already exists; choose sync or rebuild explicitly")
	}
	if action == "sync" && (loadErr != nil || previous.State != "ready" || !slices.Equal(previous.Snapshot.Roots, before.Roots) || previous.Snapshot.SettingsRevision != before.SettingsRevision) {
		return result, failure(apierrors.IndexStale, "incremental sync requires the previous authorized scope")
	}
	result = domain.RepositoryIndex{Backend: "codegraph", Version: "1.0.1", Generation: uuid.NewString(), Snapshot: before, State: "building"}
	root, err := s.projection(result)
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return result, err
	}
	// 每次维护使用独立generation；失败不会改写上次已发布的投影。
	if action == "sync" {
		oldRoot, err := s.projection(previous)
		if err != nil {
			return result, err
		}
		if err := copyGraph(oldRoot, root); err != nil {
			return result, err
		}
	}
	if err := mirrorSources(ctx, before, root, q.Budget); err != nil {
		return result, err
	}
	if _, err := s.options.Graph.invoke(ctx, root, action, q); err != nil {
		return result, err
	}
	after, _, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !SameVersion(before, after) {
		return result, failure(apierrors.IndexStale, "repository changed during index maintenance")
	}
	result.State = "ready"
	result.BuiltAt = time.Now().UTC()
	if err := s.indexes.SaveRepositoryIndex(ctx, result, expected); err != nil {
		return result, err
	}
	return result, nil
}

func copyGraph(from, to string) error {
	root, err := os.OpenRoot(filepath.Join(from, ".codegraph"))
	if err != nil {
		return err
	}
	defer root.Close()
	// SDK关闭连接以后SQLite主文件已经checkpoint，复制时拒绝活跃WAL。
	if info, err := root.Stat("codegraph.db-wal"); err == nil && info.Size() > 0 {
		return failure(apierrors.IndexStale, "CodeGraph still has uncheckpointed writes")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	budget := int64(512 * 1024 * 1024)
	return copyFile(root, "codegraph.db", filepath.Join(to, ".codegraph", "codegraph.db"), &budget)
}

func mirrorSources(ctx context.Context, v domain.RepositoryVersion, target string, b domain.RepositoryBudget) error {
	args := append([]string{"ls-files", "-z", "--cached", "--others", "--exclude-standard", "--"}, v.Roots...)
	listed, err := git(ctx, v.Root, args...)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(v.Root)
	if err != nil {
		return err
	}
	defer root.Close()
	remaining := b.MaxBytes
	seen := map[string]bool{}
	for _, item := range bytes.Split(bytes.TrimSuffix(listed, []byte{0}), []byte{0}) {
		file := string(item)
		if file == "" || seen[file] {
			continue
		}
		seen[file] = true
		if len(seen) > b.MaxFiles {
			return failure(apierrors.BudgetExhausted, "projection source count exceeds budget")
		}
		if !safeRelative(file) || !under(file, v.Roots) {
			return failure(apierrors.ScopeDenied, "projection source path is unsafe")
		}
		if privatePath(file) {
			continue
		}
		if _, err := root.Lstat(file); errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		if err := copyFile(root, file, filepath.Join(target, filepath.FromSlash(file)), &remaining); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) read(ctx context.Context, c domain.Caller, q domain.RepositoryQuery, action string) (domain.RepositoryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := domain.RepositoryResult{Symbols: []domain.RepositorySymbol{}, Relations: []domain.RepositoryRelation{}, Coverage: []string{}, Gaps: []string{"runtime_dynamic_calls_not_complete", "test_coverage_not_proven", "semantic_embedding_not_available"}}
	before, q, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	index, err := s.indexes.LoadRepositoryIndex(ctx, q.ProjectID, before.WorktreeID, "codegraph")
	if err != nil {
		return result, err
	}
	if index.State != "ready" || index.Version != "1.0.1" || !SameVersion(index.Snapshot, before) {
		return result, failure(apierrors.IndexStale, "repository index does not match the current authorized snapshot")
	}
	root, err := s.projection(index)
	if err != nil {
		return result, err
	}
	response, err := s.options.Graph.invoke(ctx, root, action, q)
	if err != nil {
		return result, err
	}
	for _, symbol := range response.Symbols {
		if !safeRelative(symbol.Span.Path) || privatePath(symbol.Span.Path) || !under(symbol.Span.Path, q.Roots) || symbol.Span.StartLine < 1 || symbol.Span.EndLine < symbol.Span.StartLine {
			return result, failure(apierrors.ScopeDenied, "CodeGraph returned a reference outside the approved source scope")
		}
	}
	after, _, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !SameVersion(before, after) {
		return result, failure(apierrors.IndexStale, "repository changed during query")
	}
	result.Index = index
	result.Symbols = response.Symbols
	result.Relations = response.Relations
	result.Coverage = slices.Clone(q.Roots)
	return result, nil
}

func (s *Service) SearchRepositoryCode(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.RepositoryResult, error) {
	if strings.TrimSpace(q.Text) == "" || len(q.Text) > 1024 {
		return domain.RepositoryResult{}, failure(apierrors.ValidationFailed, "a bounded code query is required")
	}
	return s.read(ctx, c, q, "query")
}

func (s *Service) QueryRepositoryImpact(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.RepositoryResult, error) {
	if q.SymbolID == "" || len(q.SymbolID) > 512 {
		return domain.RepositoryResult{}, failure(apierrors.ValidationFailed, "a bounded symbol identity is required")
	}
	return s.read(ctx, c, q, "impact")
}
