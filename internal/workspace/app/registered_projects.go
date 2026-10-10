package app

import (
	"context"

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
	return s.discoveryRepo.LoadRegisteredProjectDiscovery(ctx)
}

func (s *localProjectService) RefreshRegisteredProjects(ctx context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	if err := human(c); err != nil {
		return domain.ProjectDiscoverySnapshot{}, err
	}
	if s.discoveryRepo == nil || s.discoverySource == nil {
		return domain.ProjectDiscoverySnapshot{}, projectError(apierrors.UnsupportedCapability, "registered project discovery is not connected")
	}
	// Serialize refreshes without locking native session/settings operations.
	s.discoveryMu.Lock()
	defer s.discoveryMu.Unlock()
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
		return old, nil
	}
	if current.ObservedAt == nil || (current.Status != "complete" && current.Status != "partial") {
		return old, projectError(apierrors.EvidenceMissing, "registered source did not provide an observation")
	}
	if err := s.discoveryRepo.SaveRegisteredProjectDiscovery(ctx, current); err != nil {
		return current, err
	}
	return current, nil
}
