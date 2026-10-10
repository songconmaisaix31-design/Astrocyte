package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type localAgentService struct {
	provider InventoryProvider
	native   []NativeRegistryObservations
}

// NewLocalAgentService exposes only the provider's cached inventory. Native
// execution, session discovery and project access are deliberately separate.
func NewLocalAgentService(provider InventoryProvider, native ...NativeRegistryObservations) LocalAgentInventory {
	return &localAgentService{provider: provider, native: native}
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
	for _, observer := range s.native {
		if observer == nil {
			continue
		}
		observations, err := observer.SnapshotNative(ctx)
		if err != nil {
			return apierrors.ListResult{}, err
		}
		for _, observed := range observations {
			for i := range items {
				if items[i].ID != observed.ID {
					continue
				}
				items[i].NativeAdapterRegistered = observed.NativeAdapterRegistered
				if observed.Version != nil {
					items[i].Version = observed.Version
				}
				items[i].Configured = observed.Configured
				items[i].Startable = observed.Startable
				for name, capability := range observed.Capabilities {
					items[i].Capabilities[name] = capability
				}
			}
		}
	}
	result := apierrors.EmptyList()
	for _, item := range items {
		result.Items = append(result.Items, item)
	}
	return result, nil
}
