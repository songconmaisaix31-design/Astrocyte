package app

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RegisteredProjects = (*localProjectService)(nil)

// ConfigureRegisteredDiscovery is entrypoint-only wiring before concurrent use.
// Observations use a separate cache, never create a project/space or modify grants.
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
	// One request budget includes waiting for another refresh. The source's
	// own deadline alone would otherwise allow an unbounded queue wait.
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	// Serialize refreshes without locking native session/settings operations.
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
		// A caller cancellation must not lose the prior durable snapshot. Return
		// the cancellation separately; it is not a completed refresh.
		if ctx.Err() != nil {
			return old, ctx.Err()
		}
		if saveErr := s.discoveryRepo.SaveRegisteredProjectDiscovery(ctx, old); saveErr != nil {
			return old, saveErr
		}
		return s.withProjectMetadata(ctx, old)
	}
	if current.ObservedAt == nil || (current.Status != "complete" && current.Status != "partial") {
		return old, projectError(apierrors.EvidenceMissing, "registered source did not provide an observation")
	}
	// A temporary Git observation failure must not rename an already observed
	// worktree family and orphan its human annotations. This carries identity
	// only; current activity/commit unknowns stay unknown.
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

// Pure cache reads: no source calls, writes or native operations. Rebuild old
// snapshots' aggregation in memory so upgrading does not require a disk scan.
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
	// User content and revisions never grant project permissions or execution.
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
