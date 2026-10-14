package repositorycontext

import (
	"context"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ app.ProjectSnapshotProvider = (*Service)(nil)

func (s *Service) CaptureProjectSnapshot(ctx context.Context, c domain.Caller, projectID string, roots []string) (domain.ProjectContextSnapshot, error) {
	var result domain.ProjectContextSnapshot
	q := domain.RepositoryQuery{ProjectID: projectID, Purpose: "build_approved_context_packet", Roots: roots, Budget: s.options.SnapshotBudget}
	before, q, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	after, _, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !SameVersion(before, after) {
		return result, failure(apierrors.ContextStale, "repository changed during project snapshot capture")
	}
	// 固定快照独立保存其真实范围，后续Validate不相信调用方提供的Root。
	generation := uuid.NewString()
	index := domain.RepositoryIndex{Backend: "git-snapshot:" + generation, Version: "git-v1", Generation: generation, Snapshot: before, State: "ready", BuiltAt: before.ObservedAt}
	if err := s.indexes.SaveRepositoryIndex(ctx, index, ""); err != nil {
		return result, err
	}
	result = domain.ProjectContextSnapshot{EnvironmentID: before.EnvironmentID, RepoID: before.RepoID, WorktreeID: before.WorktreeID, ProjectID: projectID, Root: before.Root, Head: before.Head, IndexDigest: before.IndexDigest, WorktreeDigest: before.WorktreeDigest, Generation: generation, ObservedAt: before.ObservedAt}
	return result, nil
}

func (s *Service) ValidateProjectSnapshot(ctx context.Context, c domain.Caller, snapshot domain.ProjectContextSnapshot) error {
	if _, err := uuid.Parse(snapshot.Generation); err != nil {
		return failure(apierrors.ContextStale, "project snapshot generation is invalid")
	}
	stored, err := s.indexes.LoadRepositoryIndex(ctx, snapshot.ProjectID, snapshot.WorktreeID, "git-snapshot:"+snapshot.Generation)
	if err != nil {
		return err
	}
	if stored.Generation != snapshot.Generation || stored.State != "ready" || stored.Version != "git-v1" || snapshot.EnvironmentID != stored.Snapshot.EnvironmentID || snapshot.RepoID != stored.Snapshot.RepoID || snapshot.WorktreeID != stored.Snapshot.WorktreeID || snapshot.Head != stored.Snapshot.Head || snapshot.IndexDigest != stored.Snapshot.IndexDigest || snapshot.WorktreeDigest != stored.Snapshot.WorktreeDigest || !snapshot.ObservedAt.Equal(stored.Snapshot.ObservedAt) {
		return failure(apierrors.ContextStale, "project snapshot is not the recorded Git observation")
	}
	q := domain.RepositoryQuery{ProjectID: snapshot.ProjectID, Purpose: "validate_approved_context_packet", Roots: stored.Snapshot.Roots, Budget: s.options.SnapshotBudget}
	current, _, err := s.capture(ctx, c, q)
	if err != nil {
		return err
	}
	if snapshot.Root != current.Root || !SameVersion(stored.Snapshot, current) {
		return failure(apierrors.ContextStale, "project snapshot differs from the current authorized worktree")
	}
	return nil
}
