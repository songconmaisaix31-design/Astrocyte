package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type projectCleanupNotifier interface {
	WakeRepositoryCleanup()
}

type projectCommitCallbacks interface {
	AfterWorkspaceCommit(context.Context, func()) error
}

// 启动装配传入索引和查询共用的实际 provider；回调只唤醒原清理生命周期。
func (s *localProjectService) ConfigureRepositoryCleanup(notifier interface{ WakeRepositoryCleanup() }) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if notifier == nil {
		return projectError(apierrors.ValidationFailed, "actual repository cleanup notifier is required")
	}
	if _, ok := s.repo.(projectCommitCallbacks); !ok {
		return projectError(apierrors.UnsupportedCapability, "project repository cannot notify after its actual database commit")
	}
	s.repositoryCleanup = notifier
	return nil
}

func (s *localProjectService) wakeRepositoryCleanupAfterCommit(ctx context.Context) error {
	if s.repositoryCleanup == nil {
		return nil
	}
	callbacks, ok := s.repo.(projectCommitCallbacks)
	if !ok {
		return projectError(apierrors.UnsupportedCapability, "project commit callback is unavailable")
	}
	return callbacks.AfterWorkspaceCommit(ctx, s.repositoryCleanup.WakeRepositoryCleanup)
}
