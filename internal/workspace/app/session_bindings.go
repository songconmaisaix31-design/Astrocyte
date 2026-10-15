package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (s *S2Service) BindSession(ctx context.Context, c domain.Caller, cmd BindSessionCommand) (binding domain.SessionBinding, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if err = human(c); err != nil {
		return
	}
	if s.owned == nil {
		return binding, domain.WorkspaceError(apierrors.UnsupportedCapability, "current owned native session provider is unavailable")
	}
	if cmd.ExpectedVersion != 0 || cmd.MemberID == "" {
		return binding, domain.WorkspaceError(apierrors.ValidationFailed, "new binding requires zero version and member identity")
	}
	facts, err := s.owned.ReadOwnedSessionFacts(ctx, c, cmd.SessionID, OwnedPurposeBind)
	if err != nil {
		return binding, err
	}
	if !facts.Owned || facts.SessionID != cmd.SessionID || facts.NativeID == "" || facts.AdapterID == "" || facts.WorktreeID == "" || facts.ControllerID == "" || facts.StartupOperationID == "" || facts.Liveness != "live" {
		return binding, domain.WorkspaceError(apierrors.EvidenceMissing, "native provider did not confirm current live owned session")
	}
	err = s.command(ctx, c, cmd.WorkspaceCommandMeta, "workspace.session.bind", cmd, &binding, func(tx WorkspaceTx) error {
		packet, err := tx.LoadContextPacket(cmd.PacketID)
		if err != nil {
			return err
		}
		grant, project, err := packetGrant(tx, c, packet.GrantID, packet.GrantEpoch)
		if err != nil {
			return err
		}
		if s.snapshots == nil {
			return domain.WorkspaceError(apierrors.UnsupportedCapability, "current project snapshot provider is required")
		}
		if err = s.snapshots.ValidateProjectSnapshot(tx.Context(), c, packet.Snapshot); err != nil {
			return err
		}
		session, err := tx.LoadNativeSession(cmd.SessionID)
		if err != nil {
			return err
		}
		if session.ProjectID != project.ID || facts.ProjectID != project.ID || session.NativeID != facts.NativeID || session.CLI != facts.AdapterID || facts.WorktreeID != packet.Snapshot.WorktreeID || facts.Root != packet.Snapshot.Root || session.Ownership != "owned" || session.StopConfirmed {
			return domain.WorkspaceError(apierrors.ScopeDenied, "owned native session does not match fixed packet project and worktree")
		}
		memberGrant, err := tx.LoadProjectGrant(project.ID, cmd.MemberID)
		if err != nil {
			return err
		}
		if memberGrant.RevokedAt != nil || memberGrant.ExpiresAt != nil && !memberGrant.ExpiresAt.After(time.Now()) {
			return domain.WorkspaceError(apierrors.ApprovalRevoked, "member project grant is inactive")
		}
		now := time.Now().UTC()
		binding = domain.SessionBinding{ID: uuid.NewString(), Version: 1, ProjectID: project.ID, SessionID: session.ID, AdapterID: facts.AdapterID, NativeSessionID: facts.NativeID, ControllerID: facts.ControllerID, StartupOperationID: facts.StartupOperationID, WorktreeID: facts.WorktreeID, GrantID: grant.ID, GrantEpoch: grant.Epoch, PacketID: packet.ID, MemberID: cmd.MemberID, Status: "bound", Capabilities: facts.Capabilities, CreatedAt: now, UpdatedAt: now}
		return tx.SaveSessionBinding(binding, 0)
	})
	return
}

func (s *S2Service) WithAuthorizedSessionBindings(ctx context.Context, check domain.GrantCheck, ids []string, fn func(context.Context, domain.ApprovalGrant, []domain.SessionBinding) error) error {
	if check.Action != OwnedPurposeClaim && check.Action != OwnedPurposeSubmit {
		return NormalizeWorkspaceError(domain.WorkspaceError(apierrors.ValidationFailed, "protected binding requires claim_work_item or submit_artifact purpose"))
	}
	return s.WithAuthorizedGrant(ctx, check, func(txctx context.Context, g domain.ApprovalGrant) error {
		return s.repo.WithWorkspaceTx(txctx, func(tx WorkspaceTx) error {
			bindings := make([]domain.SessionBinding, 0, len(ids))
			for _, id := range ids {
				if s.owned == nil {
					return domain.WorkspaceError(apierrors.UnsupportedCapability, "current owned registry facts are required for protected member actions")
				}
				b, err := tx.LoadSessionBinding(id)
				if err != nil {
					return err
				}
				if b.Status != "bound" || b.GrantID != g.ID || b.GrantEpoch != g.Epoch || b.ProjectID != g.Scope.ProjectID {
					return domain.WorkspaceError(apierrors.ApprovalRevoked, "member binding is outside current approval")
				}
				facts, err := s.owned.ReadOwnedSessionFacts(txctx, check.Caller, b.SessionID, check.Action)
				if err != nil {
					return err
				}
				if !facts.Owned || facts.ProjectID != b.ProjectID || facts.NativeID != b.NativeSessionID || facts.WorktreeID != b.WorktreeID || facts.AdapterID != b.AdapterID || b.ControllerID == "" || b.StartupOperationID == "" || facts.ControllerID != b.ControllerID || facts.StartupOperationID != b.StartupOperationID {
					return domain.WorkspaceError(apierrors.EvidenceMissing, "current Registry controller tuple does not match member binding")
				}
				if facts.Liveness != "live" && (check.Action != OwnedPurposeSubmit || facts.Liveness != "exited") {
					return domain.WorkspaceError(apierrors.EvidenceMissing, "controller lifecycle does not permit the protected action")
				}
				member, err := tx.LoadProjectGrant(b.ProjectID, b.MemberID)
				if err != nil {
					return err
				}
				if member.RevokedAt != nil || member.ExpiresAt != nil && !member.ExpiresAt.After(time.Now()) {
					return domain.WorkspaceError(apierrors.ApprovalRevoked, "member project grant is inactive")
				}
				if !has(member.Actions, check.Action) {
					return domain.WorkspaceError(apierrors.ScopeDenied, "member action is not authorized")
				}
				session, err := tx.LoadNativeSession(b.SessionID)
				if err != nil {
					return err
				}
				if session.NativeID != b.NativeSessionID || session.ProjectID != b.ProjectID || session.Ownership != "owned" || session.StopConfirmed && check.Action != OwnedPurposeSubmit {
					return domain.WorkspaceError(apierrors.EvidenceMissing, "bound native session is no longer permitted for this action")
				}
				bindings = append(bindings, b)
			}
			return fn(txctx, g, bindings)
		})
	})
}
