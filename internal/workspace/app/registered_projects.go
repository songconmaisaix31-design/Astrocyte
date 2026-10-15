package app

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RegisteredProjects = (*localProjectService)(nil)

// 并发使用前由启动入口装配；观察缓存不会创建项目、空间或改变授权。
func (s *localProjectService) ConfigureRegisteredDiscovery(repo RegisteredProjectRepository, source RegisteredProjectSource) {
	s.discoveryRepo, s.discoverySource = repo, source
}

func (s *localProjectService) ListRegisteredProjects(ctx context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	if err := human(c); err != nil {
		return domain.ProjectDiscoverySnapshot{}, err
	}
	if s.discoveryRepo == nil {
		return domain.ProjectDiscoverySnapshot{}, projectError(apierrors.UnsupportedCapability, "registered project cache is not connected")
	}
	snapshot, err := s.discoveryRepo.LoadRegisteredProjectDiscovery(ctx)
	if err != nil {
		return snapshot, err
	}
	return s.withProjectMetadata(ctx, snapshot)
}

func (s *localProjectService) RefreshRegisteredProjects(ctx context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	if err := human(c); err != nil {
		return domain.ProjectDiscoverySnapshot{}, err
	}
	if s.discoveryRepo == nil || s.discoverySource == nil {
		return domain.ProjectDiscoverySnapshot{}, projectError(apierrors.UnsupportedCapability, "registered project discovery is not connected")
	}
	// 请求期限包含等待其他刷新的时间。
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	// 刷新顺序执行，原生会话和设置操作继续独立处理。
	select {
	case s.discoveryGate <- struct{}{}:
	case <-ctx.Done():
		return domain.ProjectDiscoverySnapshot{}, ctx.Err()
	}
	defer func() { <-s.discoveryGate }()
	if err := ctx.Err(); err != nil {
		return domain.ProjectDiscoverySnapshot{}, err
	}
	old, err := s.discoveryRepo.LoadRegisteredProjectDiscovery(ctx)
	if err != nil {
		return old, err
	}
	current, err := s.discoverySource.DiscoverRegistered(ctx)
	if err != nil {
		old.Status = "unknown"
		if old.ObservedAt != nil {
			old.Status = "stale"
		}
		old.Failures = []domain.ProjectDiscoveryFailure{{Root: "", Reason: "registered_source_unavailable"}}
		// 取消仍保留此前的持久快照，并传播实际错误。
		if ctx.Err() != nil {
			return old, errors.Join(err, ctx.Err())
		}
		if saveErr := s.discoveryRepo.SaveRegisteredProjectDiscovery(ctx, old); saveErr != nil {
			return old, errors.Join(err, saveErr)
		}
		projected, metadataErr := s.withProjectMetadata(ctx, old)
		return projected, errors.Join(err, metadataErr)
	}
	if current.ObservedAt == nil || (current.Status != "complete" && current.Status != "partial") {
		return old, projectError(apierrors.EvidenceMissing, "registered source did not provide an observation")
	}
	// 暂时无法观察 Git 时保留已经确认的目录身份和人工备注；当前未知字段仍未知。
	for i := range current.Projects {
		if current.Projects[i].ProjectKey != "" {
			continue
		}
		for _, previous := range old.Projects {
			if previous.Root == current.Projects[i].Root && previous.ProjectKey != "" {
				current.Projects[i].ProjectKey = previous.ProjectKey
				break
			}
		}
	}
	if err := s.discoveryRepo.SaveRegisteredProjectDiscovery(ctx, current); err != nil {
		return current, err
	}
	return s.withProjectMetadata(ctx, current)
}

// 缓存读取在内存重建聚合，来源调用、写入和原生操作均不发生。
func (s *localProjectService) withProjectMetadata(ctx context.Context, snapshot domain.ProjectDiscoverySnapshot) (domain.ProjectDiscoverySnapshot, error) {
	snapshot.Board = domain.AggregateProjects(snapshot.Projects)
	if snapshot.Sources == nil {
		snapshot.Sources = []domain.ProjectSourceObservation{}
	}
	if repo, ok := s.discoveryRepo.(RegisteredProjectMetadataRepository); ok {
		for i := range snapshot.Board {
			metadata, err := repo.LoadProjectMetadata(ctx, snapshot.Board[i].ID)
			if err != nil {
				return snapshot, err
			}
			snapshot.Board[i].Human = metadata
		}
	}
	return snapshot, nil
}

func (s *localProjectService) SetRegisteredProjectMetadata(ctx context.Context, c domain.Caller, id string, metadata domain.ProjectHumanMetadata) (domain.ProjectSummary, error) {
	if err := human(c); err != nil {
		return domain.ProjectSummary{}, err
	}
	// 人工内容与 revision 不授予项目访问或执行权限。
	if metadata.Revision < 0 || metadata.UpdatedAt != nil || !utf8.ValidString(metadata.Notes+metadata.Review+metadata.Group+metadata.Intent) || len(metadata.Notes) > 64*1024 || len(metadata.Review) > 64*1024 || len(metadata.Group) > 200 || len(metadata.Intent) > 4096 {
		return domain.ProjectSummary{}, projectError(apierrors.ValidationFailed, "human project metadata exceeds bounds or supplies server-owned updated_at")
	}
	repo, ok := s.discoveryRepo.(RegisteredProjectMetadataRepository)
	if !ok {
		return domain.ProjectSummary{}, projectError(apierrors.UnsupportedCapability, "human project metadata store is not connected")
	}
	snapshot, err := s.ListRegisteredProjects(ctx, c)
	if err != nil {
		return domain.ProjectSummary{}, err
	}
	for _, p := range snapshot.Board {
		if p.ID != id {
			continue
		}
		expected := metadata.Revision
		stamp := time.Now().UTC()
		metadata.UpdatedAt = &stamp
		metadata.Revision = expected + 1
		if err := repo.SaveProjectMetadata(ctx, id, metadata, expected); err != nil {
			return domain.ProjectSummary{}, err
		}
		p.Human = metadata
		return p, nil
	}
	return domain.ProjectSummary{}, projectError(apierrors.NotFound, "project is absent from registered observations")
}
