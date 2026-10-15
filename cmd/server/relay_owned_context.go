package main

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只映射同一Registry原控制器缓存，私有身份不进入HTTP或公开配置。
type relayOwnedContext struct {
	registry *agents.Registry
}

func (r relayOwnedContext) ReadOwnedNativeSessionRecord(ctx context.Context, id string) (domain.OwnedNativeSessionRecord, error) {
	if r.registry == nil {
		return domain.OwnedNativeSessionRecord{}, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	original, err := r.registry.ReadOwnedNativeSession(ctx, id)
	if err != nil {
		return domain.OwnedNativeSessionRecord{}, err
	}
	adapter, err := r.registry.Adapter(original.Session.CLI)
	if err != nil {
		return domain.OwnedNativeSessionRecord{}, err
	}
	liveness := "live"
	if original.Evidence.ExitConfirmed {
		liveness = "exited"
	}
	evidence := original.Evidence
	return domain.OwnedNativeSessionRecord{
		Session: original.Session, ControllerID: evidence.ControllerID, StartupOperationID: evidence.StartupOperationID,
		EvidenceSessionID: evidence.SessionID, EvidenceNativeID: evidence.NativeID, EvidenceAdapterID: evidence.Provider,
		WorkspaceRoot: evidence.WorkspaceRoot, Liveness: liveness, Capabilities: adapter.Capabilities(),
	}, nil
}
