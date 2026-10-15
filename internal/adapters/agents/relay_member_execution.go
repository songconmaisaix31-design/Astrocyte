package agents

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 成员事务只读取原控制器内存；命令策略检查仍由 OwnedRelayExecution 单独提供。
func (n *Native) ownedRelayMemberExecution(ctx context.Context, session domain.NativeSession) (RelayOwnedExecution, error) {
	var result RelayOwnedExecution
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if n == nil || session.ID == "" || session.NativeID == "" || session.ProjectID == "" || session.Version == "" || session.CLI != n.id || session.Ownership != "owned" {
		return result, nativeError(apierrors.ScopeDenied, "member execution requires the original owned controller identity")
	}
	p, err := n.get(session)
	if err != nil {
		return result, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.originalSession == nil || p.originalSession.ID != session.ID || p.originalSession.NativeID != session.NativeID || p.originalSession.ProjectID != session.ProjectID || p.originalSession.CLI != session.CLI || p.originalSession.Version != session.Version || p.originalSession.Ownership != "owned" || p.projectID != session.ProjectID || p.nativeID != session.NativeID || p.cli != session.CLI || p.relayEvidence.SessionID != session.ID || p.relayEvidence.ControllerID == "" || p.relayEvidence.StartupOperationID == "" || p.relayEvidence.Provider != session.CLI {
		return result, nativeError(apierrors.DeliveryUnknown, "retained member controller identity is unavailable or differs")
	}
	if p.cmd == nil || p.cmd.Process == nil || p.cmd.Process.Pid <= 0 || p.done == nil || p.startedAt.IsZero() || p.deadlineAt.IsZero() || p.relayEvidence.WorkspaceRoot == "" || !sameRelayMetadataCWD(p.cmd.Dir, p.relayEvidence.WorkspaceRoot) {
		return result, nativeError(apierrors.DeliveryUnknown, "original member process metadata is incomplete")
	}
	result = RelayOwnedExecution{
		ProcessID: p.cmd.Process.Pid, StartedAt: p.startedAt, DeadlineAt: p.deadlineAt,
		CLI: session.CLI, Version: session.Version, LaunchPolicy: "unknown", PermissionPolicy: "unknown",
		EffectiveToolSet: "unknown", DisabledFeatures: []string{}, MetadataStatus: "available",
	}
	if p.permissionPolicy != "" {
		result.PermissionPolicy = p.permissionPolicy
	}
	select {
	case <-p.done:
		result.ExitConfirmed = true
	default:
	}
	return result, nil
}
