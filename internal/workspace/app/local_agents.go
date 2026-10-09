package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type localAgentService struct{ provider InventoryProvider }

// NewLocalAgentService exposes only the provider's cached inventory. Native
// execution, session discovery and project access are deliberately separate.
func NewLocalAgentService(provider InventoryProvider) LocalAgentInventory {
	return &localAgentService{provider: provider}
}

func (s *localAgentService) ListLocalAgents(ctx context.Context) (apierrors.ListResult, error) {
	if err := ctx.Err(); err != nil {
		return apierrors.ListResult{}, err
	}
	if s.provider == nil {
		return apierrors.ListResult{}, &apierrors.ServiceError{
			Code:           apierrors.UnsupportedCapability,
			Message:        "local Agent inventory provider is not connected",
			RequiredAction: "connect_local_agent_inventory_provider",
		}
	}
	items, err := s.provider.Snapshot(ctx)
	if err != nil {
		return apierrors.ListResult{}, err
	}
	result := apierrors.EmptyList()
	for _, item := range items {
		result.Items = append(result.Items, item)
	}
	return result, nil
}
