package repositorycontext

import (
	"context"
	"errors"
	"sync"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

var _ app.RepositoryCleanupWorker = (*Service)(nil)

type cleanupWorker struct {
	mu      sync.Mutex
	running bool
	wake    chan struct{}
}

// 唤醒可以合并；待清理事实始终来自项目、索引元数据和真实目录。
func (s *Service) WakeRepositoryCleanup() {
	select {
	case s.cleanup.wake <- struct{}{}:
	default:
	}
}

// H持有此方法的生命周期，取消后等待其返回，再关闭原SQLite。
func (s *Service) RunRepositoryCleanup(ctx context.Context, maxGenerations int) error {
	if maxGenerations < 1 || maxGenerations > 1000 {
		return failure(apierrors.BudgetExhausted, "repository cleanup generation limit is invalid")
	}
	s.cleanup.mu.Lock()
	if s.cleanup.running {
		s.cleanup.mu.Unlock()
		return failure(apierrors.VersionConflict, "repository cleanup worker is already running")
	}
	s.cleanup.running = true
	s.cleanup.mu.Unlock()
	defer func() {
		s.cleanup.mu.Lock()
		s.cleanup.running = false
		s.cleanup.mu.Unlock()
	}()
	for {
		if ctx.Err() != nil {
			return nil
		}
		result, err := s.CleanupRepositoryContext(ctx, maxGenerations)
		if err != nil {
			if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
				return nil
			}
			return err
		}
		if result.HasMore {
			if result.DeletedGenerations == 0 {
				return failure(apierrors.EvidenceMissing, "repository cleanup made no progress with remaining generations")
			}
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-s.cleanup.wake:
		}
	}
}
