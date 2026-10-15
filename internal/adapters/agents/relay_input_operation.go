package agents

import (
	"context"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 私有执行证据供 S2 / S3 原绑定检查；不授权 Member，也不作为 HTTP DTO。
type RelayOwnedInputOperation struct {
	Session       domain.NativeSession           `json:"-"`
	Controller    domain.RelayControllerEvidence `json:"-"`
	Execution     RelayOwnedExecution            `json:"-"`
	OperationID   string                         `json:"-"`
	Purpose       string                         `json:"-"`
	InputAccepted bool                           `json:"-"`
	Turn          *domain.RelayTurnBinding       `json:"-"`
	ToolsKnown    bool                           `json:"-"`
	TargetsKnown  bool                           `json:"-"`
	CallsKnown    bool                           `json:"-"`
	TokensKnown   bool                           `json:"-"`
}

// Claim 在输入前检查当前控制器；Submit 检查实际已接受输入，允许原控制器已退出。
func (r *Registry) ReadOwnedInputOperation(ctx context.Context, sessionID, operationID, purpose string) (RelayOwnedInputOperation, error) {
	var result RelayOwnedInputOperation
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if _, err := uuid.Parse(operationID); err != nil || sessionID == "" {
		return result, nativeError(apierrors.ValidationFailed, "original session and UUID input operation are required")
	}
	if purpose != "claim_work_item" && purpose != "submit_artifact" {
		return result, nativeError(apierrors.ValidationFailed, "native membership purpose is unsupported")
	}
	original, err := r.ReadOwnedNativeSession(ctx, sessionID)
	if err != nil {
		return result, err
	}
	if original.Session.CLI != "codex" {
		return result, nativeError(apierrors.UnsupportedCapability, "this provider has no verified original input turn binding")
	}
	adapter, err := r.Adapter(original.Session.CLI)
	if err != nil {
		return result, err
	}
	native := adapter.(*Native)
	execution, err := native.OwnedRelayExecution(ctx, original.Session)
	if err != nil {
		return result, err
	}
	if execution.ExitConfirmed != original.Evidence.ExitConfirmed {
		return result, nativeError(apierrors.DeliveryUnknown, "original controller changed during membership observation")
	}
	if purpose == "claim_work_item" && execution.ExitConfirmed {
		return result, nativeError(apierrors.DeliveryUnknown, "work item claim requires the original live controller")
	}
	process, err := native.get(original.Session)
	if err != nil {
		return result, err
	}
	process.mu.Lock()
	occupied := process.status == "blocked" || process.truncated || process.status == "running" && process.inputOperationID != operationID
	process.mu.Unlock()
	if purpose == "claim_work_item" && occupied {
		return result, nativeError(apierrors.VersionConflict, "original controller is occupied by another input or blocked")
	}
	result = RelayOwnedInputOperation{Session: original.Session, Controller: original.Evidence, Execution: execution, OperationID: operationID, Purpose: purpose}
	for _, turn := range original.Evidence.Turns {
		if turn.OperationID != operationID {
			continue
		}
		if result.Turn != nil || turn.SessionID != original.Session.ID || turn.NativeID != original.Session.NativeID || turn.ControllerID != original.Evidence.ControllerID || turn.StartupOperationID != original.Evidence.StartupOperationID || turn.TurnID == "" || turn.AcceptedAt.IsZero() {
			return RelayOwnedInputOperation{}, nativeError(apierrors.ScopeDenied, "original input turn differs from its retained controller tuple")
		}
		copy := turn
		result.Turn = &copy
		result.InputAccepted = true
	}
	if purpose == "submit_artifact" && !result.InputAccepted {
		return RelayOwnedInputOperation{}, nativeError(apierrors.DeliveryUnknown, "artifact submission has no accepted original input operation")
	}
	// 原生回合受理和启动策略不证明有效工具、外部目标或本次完整 usage。
	return result, nil
}

// 调用方先检查原 binding / epoch / grant；此端口只终止准确已持有的原输入。
func (r *Registry) StopOwnedInputOperation(ctx context.Context, sessionID, operationID string) (RelayOwnedInputOperation, error) {
	original, err := r.ReadOwnedInputOperation(ctx, sessionID, operationID, "submit_artifact")
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	if original.Execution.ExitConfirmed {
		return original, nil
	}
	adapter, err := r.Adapter(original.Session.CLI)
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	native := adapter.(*Native)
	process, err := native.get(original.Session)
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	process.operationMu.Lock()
	defer process.operationMu.Unlock()
	process.mu.Lock()
	current := process.inputOperationID == operationID
	process.mu.Unlock()
	if !current {
		return RelayOwnedInputOperation{}, nativeError(apierrors.VersionConflict, "original controller now belongs to another input operation")
	}
	if _, err := native.StopRelay(ctx, original.Session); err != nil {
		return RelayOwnedInputOperation{}, err
	}
	return r.ReadOwnedInputOperation(ctx, sessionID, operationID, "submit_artifact")
}
