package agents

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ domain.RelayMemberEvidenceSource = (*Registry)(nil)

func (r *Registry) ReadRelayMemberEvidence(ctx context.Context, sessionID, inputOperationID, purpose string) (domain.RelayMemberEvidence, error) {
	if r == nil {
		return domain.RelayMemberEvidence{}, nativeError(apierrors.UnsupportedCapability, "original Registry is not connected")
	}
	original, err := r.ReadOwnedInputOperation(ctx, sessionID, inputOperationID, purpose)
	if err != nil {
		return domain.RelayMemberEvidence{}, err
	}
	return r.portableMemberEvidence(ctx, original)
}

func (r *Registry) StopRelayMemberEvidence(ctx context.Context, sessionID, inputOperationID string) (domain.RelayMemberEvidence, error) {
	if r == nil {
		return domain.RelayMemberEvidence{}, nativeError(apierrors.UnsupportedCapability, "original Registry is not connected")
	}
	original, err := r.StopOwnedInputOperation(ctx, sessionID, inputOperationID)
	if err != nil {
		return domain.RelayMemberEvidence{}, err
	}
	return r.portableMemberEvidence(ctx, original)
}

func (r *Registry) portableMemberEvidence(ctx context.Context, original RelayOwnedInputOperation) (domain.RelayMemberEvidence, error) {
	var result domain.RelayMemberEvidence
	if err := ctx.Err(); err != nil {
		return result, err
	}
	adapter, err := r.Adapter(original.Session.CLI)
	if err != nil {
		return result, err
	}
	native := adapter.(*Native)
	process, err := native.get(original.Session)
	if err != nil {
		return result, err
	}
	process.mu.Lock()
	matches := process.relayEvidence.ControllerID == original.Controller.ControllerID && process.relayEvidence.StartupOperationID == original.Controller.StartupOperationID
	currentInput := process.inputOperationID == original.OperationID
	exited := false
	select {
	case <-process.done:
		exited = true
	default:
	}
	process.mu.Unlock()
	if !matches || exited != original.Execution.ExitConfirmed {
		return result, nativeError(apierrors.DeliveryUnknown, "original controller changed before portable membership observation")
	}
	result = domain.RelayMemberEvidence{
		Purpose: original.Purpose, SessionID: original.Session.ID, NativeID: original.Session.NativeID,
		ControllerID: original.Controller.ControllerID, StartupOperationID: original.Controller.StartupOperationID,
		InputOperationID: original.OperationID, ProjectID: original.Session.ProjectID, CWD: original.Controller.WorkspaceRoot,
		CLI: original.Session.CLI, CLIVersion: original.Session.Version,
		ControllerHeld: true, CurrentInput: currentInput, ExitConfirmed: original.Execution.ExitConfirmed,
		InputAccepted: original.InputAccepted,
	}
	if !original.Execution.StartedAt.IsZero() {
		started := original.Execution.StartedAt
		result.ControllerStartedAt = &started
	}
	if original.Turn != nil {
		turnID, acceptedAt := original.Turn.TurnID, original.Turn.AcceptedAt
		result.AcceptedTurnID, result.InputAcceptedAt = &turnID, &acceptedAt
		if original.Turn.StartedAt != nil {
			started := *original.Turn.StartedAt
			result.TurnStartedAt = &started
		}
		if original.Turn.ResultObservedAt != nil && original.Turn.ResultSource == "codex_turn_completed" {
			outcome, when := original.Turn.Outcome, *original.Turn.ResultObservedAt
			result.Outcome, result.ResultObservedAt, result.ResultSource = &outcome, &when, original.Turn.ResultSource
			if original.Turn.FailureCode != "" {
				code := original.Turn.FailureCode
				result.FailureCode = &code
			}
		}
	}
	// 不从读取时间生成退出时间，不将原控制器持有记录当作完整活性证明。
	// 当前没有完整工具、目标或单输入 usage 来源，保持 false / nil。
	return result, nil
}
