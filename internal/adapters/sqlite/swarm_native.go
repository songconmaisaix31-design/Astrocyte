package sqlite

import (
	"context"
	"errors"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type SwarmNativeBridge struct {
	repo   *SwarmRepository
	source workspace.RelayMemberEvidenceSource
}

var _ swarm.MemberAuthorizer = (*SwarmNativeBridge)(nil)
var _ swarm.ExecutionDescriptorReader = (*SwarmNativeBridge)(nil)
var _ swarm.ExecutionOwnership = (*SwarmNativeBridge)(nil)

func missingSwarmLease(err error) bool {
	var missing *apierrors.ServiceError
	return errors.As(err, &missing) && missing.Code == apierrors.NotFound
}

func NewSwarmNativeBridge(repo *SwarmRepository, source workspace.RelayMemberEvidenceSource) (*SwarmNativeBridge, error) {
	if repo == nil || source == nil {
		return nil, domain.Failure(apierrors.UnsupportedCapability, "actual Swarm ledger and original Registry evidence source are required")
	}
	return &SwarmNativeBridge{repo: repo, source: source}, nil
}

func swarmMatchNativeTuple(e workspace.RelayMemberEvidence, purpose, session, native, adapter, controller, startup, operation, project, root string) error {
	if session == "" || native == "" || adapter == "" || controller == "" || startup == "" || operation == "" || project == "" || root == "" {
		return domain.Failure(apierrors.EvidenceMissing, "original bound controller and input operation tuple is incomplete")
	}
	if startup == operation {
		return domain.Failure(apierrors.ScopeDenied, "startup operation cannot replace the reserved input operation")
	}
	if e.Purpose != purpose || e.SessionID != session || e.NativeID != native || e.CLI != adapter || e.ControllerID != controller || e.StartupOperationID != startup || e.InputOperationID != operation || e.ProjectID != project || e.CWD != root {
		return domain.Failure(apierrors.ContextStale, "original Registry evidence differs from the authorized binding tuple")
	}
	if !e.ControllerHeld || e.CLIVersion == "" {
		return domain.Failure(apierrors.EvidenceMissing, "original Registry controller ownership is not established")
	}
	if purpose == "claim_work_item" {
		if e.ExitConfirmed {
			return domain.Failure(apierrors.DeliveryUnknown, "new claim requires the original live controller")
		}
	} else if purpose == "submit_artifact" {
		if !e.InputAccepted || e.AcceptedTurnID == nil || *e.AcceptedTurnID == "" || e.InputAcceptedAt == nil || e.InputAcceptedAt.IsZero() {
			return domain.Failure(apierrors.DeliveryUnknown, "submission requires the actual accepted original input turn")
		}
	} else {
		return domain.Failure(apierrors.ValidationFailed, "native member evidence requires an explicit protected action")
	}
	return nil
}

func swarmMatchPreparedInput(prepared workspace.RelayPreparedInput, e workspace.RelayMemberEvidence, p domain.Principal) error {
	if err := swarmMatchNativeTuple(e, "claim_work_item", prepared.Session.ID, prepared.Session.NativeID, prepared.Session.CLI, prepared.ControllerID, prepared.StartupOperationID, prepared.OperationID, prepared.Session.ProjectID, prepared.WorkspaceRoot); err != nil {
		return err
	}
	if p.Kind != "agent" || p.ID == "" || p.ProjectID == "" || prepared.ActorKind != "agent" || prepared.ActorID != p.ID || prepared.ActorProjectID != p.ProjectID || prepared.Session.ProjectID != p.ProjectID {
		return domain.Failure(apierrors.ScopeDenied, "prepared input actor differs from the current authenticated member")
	}
	if prepared.PreparedAt.IsZero() || prepared.Session.Ownership != "owned" || prepared.Session.StopConfirmed {
		return domain.Failure(apierrors.EvidenceMissing, "actual unattempted preparation and owned session evidence are required")
	}
	if e.InputAccepted || e.AcceptedTurnID != nil || e.InputAcceptedAt != nil || e.Outcome != nil || e.FailureCode != nil || e.ResultSource != "" || e.ResultObservedAt != nil {
		return domain.Failure(apierrors.DeliveryUnknown, "a new lease cannot be added after an original input acceptance or result")
	}
	return nil
}

func (b *SwarmNativeBridge) memberEvidence(ctx context.Context, p domain.Principal, m swarm.Member) (workspace.RelayMemberEvidence, error) {
	authorized, err := swarm.AuthorizedBindingForMember(ctx, p, m)
	if err != nil {
		return workspace.RelayMemberEvidence{}, err
	}
	if authorized.ControllerID == "" || authorized.StartupOperationID == "" {
		return workspace.RelayMemberEvidence{}, domain.Failure(apierrors.EvidenceMissing, "current Workspace binding lacks the original private controller tuple")
	}
	e, err := b.source.ReadRelayMemberEvidence(ctx, m.SessionID, m.OperationID, m.Purpose)
	if err != nil {
		return e, err
	}
	err = swarmMatchNativeTuple(e, m.Purpose, authorized.SessionID, authorized.NativeSessionID, authorized.AdapterID, authorized.ControllerID, authorized.StartupOperationID, m.OperationID, authorized.ProjectID, authorized.Root)
	if err != nil {
		return e, err
	}
	if m.Purpose == "claim_work_item" {
		l, err := b.repo.LeaseForOperation(ctx, m.OperationID, m.SessionID)
		if missingSwarmLease(err) {
			reader, ok := b.source.(workspace.RelayPreparedInputReader)
			if !ok {
				return e, domain.Failure(apierrors.UnsupportedCapability, "actual original Registry prepared input reader is unavailable")
			}
			prepared, err := reader.ReadPreparedNativeSession(ctx, m.SessionID, m.OperationID)
			if err != nil {
				return e, err
			}
			return e, swarmMatchPreparedInput(prepared, e, p)
		}
		if err != nil {
			return e, err
		}
		if l.BindingID != authorized.ID || l.MemberVersion != m.Version || l.HolderID != p.ID {
			return e, domain.Failure(apierrors.LeaseLost, "existing input lease differs from the current member binding")
		}
		if err = swarmMatchNativeTuple(e, m.Purpose, l.SessionID, l.NativeSessionID, l.AdapterID, l.ControllerID, l.StartupOperationID, l.OperationID, l.ProjectID, l.Worktree); err != nil {
			return e, err
		}
	}
	return e, nil
}

func (b *SwarmNativeBridge) CheckMember(ctx context.Context, p domain.Principal, m swarm.Member) error {
	_, err := b.memberEvidence(ctx, p, m)
	return err
}

func swarmNativeDescriptor(e workspace.RelayMemberEvidence) swarm.ExecutionDescriptor {
	return swarm.ExecutionDescriptor{Known: e.ToolsKnown && e.TargetsKnown, Capabilities: slices.Clone(e.Capabilities), Tools: slices.Clone(e.Tools), ExternalTargets: slices.Clone(e.ExternalTargets)}
}

func (b *SwarmNativeBridge) DescribeExecution(ctx context.Context, p domain.Principal, m swarm.Member) (swarm.ExecutionDescriptor, error) {
	e, err := b.memberEvidence(ctx, p, m)
	if err != nil {
		return swarm.ExecutionDescriptor{}, err
	}
	return swarmNativeDescriptor(e), nil
}

func swarmMatchLeasedNative(e workspace.RelayMemberEvidence, l domain.Lease) error {
	return swarmMatchNativeTuple(e, "submit_artifact", l.SessionID, l.NativeSessionID, l.AdapterID, l.ControllerID, l.StartupOperationID, l.OperationID, l.ProjectID, l.Worktree)
}

func swarmNativeExecutionFact(e workspace.RelayMemberEvidence) (swarm.ExecutionFact, error) {
	fact := swarm.ExecutionFact{OperationID: e.InputOperationID, SessionID: e.SessionID, StopConfirmed: e.ExitConfirmed}
	if e.Outcome == nil && e.FailureCode == nil && e.ResultSource == "" && e.ResultObservedAt == nil {
		return fact, nil
	}
	if e.Purpose != "submit_artifact" || e.CLI != "codex" || !e.InputAccepted || e.AcceptedTurnID == nil || *e.AcceptedTurnID == "" || e.InputAcceptedAt == nil || e.InputAcceptedAt.IsZero() || e.Outcome == nil || e.ResultSource != "codex_turn_completed" || e.ResultObservedAt == nil || e.ResultObservedAt.IsZero() {
		return fact, domain.Failure(apierrors.EvidenceMissing, "original accepted input result lacks its actual native observation")
	}
	if !slices.Contains([]string{"completed", "failed", "interrupted"}, *e.Outcome) || e.FailureCode != nil && (*e.FailureCode != "native_failure" || *e.Outcome == "completed") {
		return fact, domain.Failure(apierrors.EvidenceMissing, "original input result contains unsupported or contradictory outcome evidence")
	}
	// 原 input 结果、控制器退出与调用预算分别保存；响应计数不补充未知用量。
	fact.Outcome = *e.Outcome
	if e.FailureCode != nil {
		fact.FailureCode = *e.FailureCode
	}
	return fact, nil
}

func (b *SwarmNativeBridge) Reconcile(ctx context.Context, operationID, sessionID string) (swarm.ExecutionFact, error) {
	l, err := b.repo.LeaseForOperation(ctx, operationID, sessionID)
	if err != nil {
		return swarm.ExecutionFact{}, err
	}
	e, err := b.source.ReadRelayMemberEvidence(ctx, sessionID, operationID, "submit_artifact")
	if err != nil {
		return swarm.ExecutionFact{}, err
	}
	if err = swarmMatchLeasedNative(e, l); err != nil {
		return swarm.ExecutionFact{}, err
	}
	return swarmNativeExecutionFact(e)
}

func (b *SwarmNativeBridge) Stop(ctx context.Context, operationID, sessionID string) (swarm.ExecutionFact, error) {
	l, err := b.repo.LeaseForOperation(ctx, operationID, sessionID)
	if err != nil {
		return swarm.ExecutionFact{}, err
	}
	e, err := b.source.ReadRelayMemberEvidence(ctx, sessionID, operationID, "submit_artifact")
	if err != nil {
		return swarm.ExecutionFact{}, err
	}
	if err = swarmMatchLeasedNative(e, l); err != nil {
		return swarm.ExecutionFact{}, err
	}
	if !e.ExitConfirmed && !e.CurrentInput {
		return swarm.ExecutionFact{}, domain.Failure(apierrors.VersionConflict, "original controller is occupied by a different input operation")
	}
	if !e.ExitConfirmed {
		e, err = b.source.StopRelayMemberEvidence(ctx, sessionID, operationID)
		if err != nil {
			return swarm.ExecutionFact{}, err
		}
		if err = swarmMatchLeasedNative(e, l); err != nil {
			return swarm.ExecutionFact{}, err
		}
	}
	return swarmNativeExecutionFact(e)
}
