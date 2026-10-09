package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// InventoryProvider returns a cached CLI inventory. Snapshot must not execute
// commands, inspect credentials or scan project/session directories.
type InventoryProvider interface {
	Snapshot(ctx context.Context) ([]domain.LocalAgent, error)
}

// LocalAgentInventory is a human-readable installation/capability query, not
// permission to start or resume an Agent or access its private native state.
type LocalAgentInventory interface {
	ListLocalAgents(ctx context.Context) (apierrors.ListResult, error)
}
