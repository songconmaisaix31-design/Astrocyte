package repositorycontext

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ app.RepositoryProjectionCleaner = (*Service)(nil)

func cleanupEntries(root *os.Root) ([]os.DirEntry, error) {
	dir, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, nil
}

// 调用方使用既有worker；同一Service的互斥锁保护尚未发布的generation。
func (s *Service) CleanupRepositoryContext(ctx context.Context, maxGenerations int) (domain.RepositoryCleanupResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result domain.RepositoryCleanupResult
	if maxGenerations < 1 || maxGenerations > 1000 {
		return result, failure(apierrors.BudgetExhausted, "repository cleanup generation limit is invalid")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	identities := map[string]string{}
	state, err := s.indexes.PrepareRepositoryContextCleanup(ctx, func(ctx context.Context, project domain.LocalProject, index domain.RepositoryIndex) (bool, error) {
		identity, found := identities[project.ID]
		if !found {
			root, err := canonicalRoot(project.Root)
			if err != nil {
				return false, err
			}
			gitdir, err := git(ctx, root, "rev-parse", "--absolute-git-dir")
			if err != nil {
				return false, err
			}
			identity = digest(append(bytes.TrimSpace(gitdir), []byte(root)...))
			identities[project.ID] = identity
		}
		return index.Snapshot.EnvironmentID == s.options.Git.EnvironmentID && identity == index.Snapshot.WorktreeID, nil
	})
	if err != nil {
		return result, err
	}
	result.RetiredIndexes = state.RetiredIndexes
	retained := map[string]bool{}
	for _, index := range state.Retained {
		projection, err := s.projection(index)
		if err != nil {
			return result, err
		}
		retained[projection] = true
	}
	cachePath, err := canonicalRoot(s.options.CacheRoot)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	cache, err := os.OpenRoot(cachePath)
	if err != nil {
		return result, err
	}
	defer cache.Close()
	worktrees, err := cleanupEntries(cache)
	if err != nil {
		return result, err
	}
	for _, entry := range worktrees {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		identity := entry.Name()
		if _, err := hex.DecodeString(identity); len(identity) != 64 || err != nil || strings.ToLower(identity) != identity {
			result.PreservedEntries++
			continue
		}
		info, err := cache.Lstat(identity)
		if err != nil {
			return result, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return result, failure(apierrors.ScopeDenied, "repository cleanup worktree directory is unsafe")
		}
		worktree, err := cache.OpenRoot(identity)
		if err != nil {
			return result, err
		}
		err = s.cleanupWorktree(ctx, cache, worktree, cachePath, identity, retained, maxGenerations, &result)
		closeErr := worktree.Close()
		if err != nil {
			return result, err
		}
		if closeErr != nil {
			return result, closeErr
		}
	}
	return result, nil
}

func (s *Service) cleanupWorktree(ctx context.Context, cache, worktree *os.Root, cachePath, identity string, retained map[string]bool, limit int, result *domain.RepositoryCleanupResult) error {
	entries, err := cleanupEntries(worktree)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := entry.Name()
		generation, err := uuid.Parse(name)
		if err != nil || generation.String() != name {
			result.PreservedEntries++
			continue
		}
		info, err := worktree.Lstat(name)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return failure(apierrors.ScopeDenied, "repository cleanup generation directory is unsafe")
		}
		if retained[filepath.Join(s.options.CacheRoot, identity, name)] {
			continue
		}
		if result.DeletedGenerations >= limit {
			result.HasMore = true
			continue
		}
		// 删除前核实绝对目标位于可信CacheRoot，再由os.Root约束递归操作。
		target := filepath.Join(cachePath, identity, name)
		relative, err := filepath.Rel(cachePath, target)
		if err != nil {
			return err
		}
		if !safeRelative(filepath.ToSlash(relative)) || relative == "." {
			return failure(apierrors.ScopeDenied, "repository cleanup target escapes the cache")
		}
		if err := preserveGraphDiagnostics(ctx, cache, relative); err != nil {
			return err
		}
		if err := cache.RemoveAll(relative); err != nil {
			return err
		}
		result.DeletedGenerations++
	}
	return nil
}

func preserveGraphDiagnostics(ctx context.Context, cache *os.Root, generation string) error {
	graphPath := filepath.Join(generation, ".codegraph")
	info, err := cache.Lstat(graphPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return failure(apierrors.ScopeDenied, "repository diagnostic directory is unsafe")
	}
	graph, err := cache.OpenRoot(graphPath)
	if err != nil {
		return err
	}
	defer graph.Close()
	entries, err := cleanupEntries(graph)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "astrocyte-error-") || !strings.HasSuffix(name, ".log") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimPrefix(name, "astrocyte-error-"), ".log")
		identity, err := uuid.Parse(id)
		if err != nil || identity.String() != id {
			return failure(apierrors.ScopeDenied, "repository diagnostic identity is unsafe")
		}
		info, err := graph.Lstat(name)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return failure(apierrors.ScopeDenied, "repository diagnostic must be a regular file")
		}
		if err := cache.MkdirAll("diagnostics", 0700); err != nil {
			return err
		}
		info, err = cache.Lstat("diagnostics")
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return failure(apierrors.ScopeDenied, "repository diagnostic archive directory is unsafe")
		}
		destination := filepath.Join("diagnostics", name)
		if _, err := cache.Lstat(destination); err == nil {
			return failure(apierrors.VersionConflict, "repository diagnostic archive already exists")
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// 仅移动本平台自己的错误日志，原件在删除旧generation后继续保留。
		if err := cache.Rename(filepath.Join(graphPath, name), destination); err != nil {
			return err
		}
	}
	return nil
}
