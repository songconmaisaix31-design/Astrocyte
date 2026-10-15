package sqlite

import (
	"context"
	"database/sql"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type storedSwarmLease struct {
	domain.Lease
	PrivateNativeSessionID    string `json:"private_native_session_id"`
	PrivateAdapterID          string `json:"private_adapter_id"`
	PrivateControllerID       string `json:"private_controller_id"`
	PrivateStartupOperationID string `json:"private_startup_operation_id"`
}

func storeSwarmLease(l domain.Lease) storedSwarmLease {
	return storedSwarmLease{Lease: l, PrivateNativeSessionID: l.NativeSessionID, PrivateAdapterID: l.AdapterID, PrivateControllerID: l.ControllerID, PrivateStartupOperationID: l.StartupOperationID}
}

func (s storedSwarmLease) restore() domain.Lease {
	l := s.Lease
	l.NativeSessionID = s.PrivateNativeSessionID
	l.AdapterID = s.PrivateAdapterID
	l.ControllerID = s.PrivateControllerID
	l.StartupOperationID = s.PrivateStartupOperationID
	return l
}

func swarmLoadLease(row *sql.Row, id string) (domain.Lease, error) {
	stored, err := swarmDecode[storedSwarmLease](row, "lease", id)
	if err != nil {
		return domain.Lease{}, err
	}
	return stored.restore(), nil
}

// 仅供受信任的原执行者对账桥读取；该查询不授予停止权限。
func (r *SwarmRepository) LeaseForOperation(ctx context.Context, operationID, sessionID string) (domain.Lease, error) {
	if operationID == "" || sessionID == "" {
		return domain.Lease{}, domain.Failure(apierrors.ValidationFailed, "original operation and session are required")
	}
	executor, err := r.executor(ctx)
	if err != nil {
		return domain.Lease{}, err
	}
	l, err := swarmLoadLease(executor.QueryRowContext(ctx, "SELECT data FROM swarm_leases WHERE operation_id=?", operationID), operationID)
	if err != nil {
		return l, err
	}
	if l.OperationID != operationID || l.SessionID != sessionID {
		return domain.Lease{}, domain.Failure(apierrors.ScopeDenied, "operation does not belong to the original leased session")
	}
	return l, nil
}
