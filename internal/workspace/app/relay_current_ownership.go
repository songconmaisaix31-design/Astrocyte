package app

import (
	"context"
	"errors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayCurrentOwnershipReader = (*relayService)(nil)

func (s *relayService) GetCurrentOwnership(ctx context.Context, caller domain.Caller, id string) (domain.RelayCurrentOwnership, error) {
	unknown := func(reason string) domain.RelayCurrentOwnership {
		return domain.RelayCurrentOwnership{Status: "unknown", Reason: reason}
	}
	st, err := s.snapshot(ctx, caller)
	if err != nil {
		return domain.RelayCurrentOwnership{}, err
	}
	task, exists := st.Tasks[id]
	if !exists {
		return domain.RelayCurrentOwnership{}, apierrors.NewNotFound("relay task", id)
	}
	p, err := s.permission(ctx, caller, task, "observe")
	if err != nil {
		return domain.RelayCurrentOwnership{}, err
	}
	stored := st.ControllerEvidence[id]
	if task.TerminalID == "" || stored.ControllerID == "" || stored.StartupOperationID == "" {
		return unknown("controller_binding_unrecorded"), nil
	}
	workspace, workspaceExists := st.Workspaces[task.WorktreeID]
	terminal, terminalExists := st.Terminals[task.TerminalID]
	if !workspaceExists || !terminalExists || workspace.ProjectID != task.ProjectID || terminal.WorktreeID != task.WorktreeID || terminal.Agent != task.Agent || !sameRelayRoot(workspace.SourceRoot, p.Root) {
		return domain.RelayCurrentOwnership{}, relayError(apierrors.ScopeDenied, "current controller workspace identity differs")
	}
	if !s.existingOwnedRoot(workspace.SourceRoot) || !s.existingOwnedRoot(workspace.Root) {
		return unknown("workspace_identity_unverified"), nil
	}
	driver := s.structured
	if terminal.DriveMode == "terminal" {
		driver = s.terminal
	}
	reader, supported := driver.(RelayControllerEvidenceReader)
	if !supported {
		return unknown("controller_metadata_unsupported"), nil
	}
	evidence, err := reader.CurrentControllerEvidence(ctx, RelayLaunch{Spec: domain.AgentSpec{Name: task.Agent}, Project: p, Workspace: workspace, Terminal: terminal, Session: st.Sessions[terminal.ID], ControllerID: stored.ControllerID})
	if err != nil {
		var failure *apierrors.ServiceError
		if errors.As(err, &failure) && (failure.Code == apierrors.DeliveryUnknown || failure.Code == apierrors.UnsupportedCapability || failure.Code == apierrors.EvidenceMissing) {
			return unknown("controller_not_held"), nil
		}
		return domain.RelayCurrentOwnership{}, err
	}
	if evidence.ControllerID != stored.ControllerID || evidence.StartupOperationID != stored.StartupOperationID || evidence.SessionID != stored.SessionID || evidence.NativeID != stored.NativeID || evidence.Provider != task.Agent || !sameRelayRoot(evidence.WorkspaceRoot, workspace.Root) {
		return domain.RelayCurrentOwnership{}, relayError(apierrors.ScopeDenied, "current controller differs from its original recorded identity")
	}
	if evidence.ExitConfirmed {
		return domain.RelayCurrentOwnership{Status: "confirmed", Reason: "controller_exited"}, nil
	}
	return domain.RelayCurrentOwnership{Status: "confirmed"}, nil
}
