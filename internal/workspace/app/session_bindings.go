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
	if err = s.requireWorkspaceObservation(ctx); err != nil {
		return binding, err
	}
	if s.snapshots == nil {
		return binding, domain.WorkspaceError(apierrors.UnsupportedCapability, "current project snapshot provider is required")
	}
	observedPacket, err := s.ReadContextPacket(ctx, c, cmd.PacketID)
	if err != nil {
		return binding, err
	}
	facts, err := s.owned.ReadOwnedSessionFacts(ctx, c, cmd.SessionID, OwnedPurposeBind)
	if err != nil {
		return binding, err
	}
	if !facts.Owned || facts.SessionID != cmd.SessionID || facts.NativeID == "" || facts.AdapterID == "" || facts.WorktreeID == "" || facts.ControllerID == "" || facts.StartupOperationID == "" || facts.Liveness != "live" {
		return binding, domain.WorkspaceError(apierrors.EvidenceMissing, "native provider did not confirm current live owned session")
	}
	if err = s.snapshots.ValidateProjectSnapshot(ctx, c, observedPacket.Snapshot); err != nil {
		return binding, err
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
		if packet.ObjectRef != observedPacket.ObjectRef || packet.Snapshot != observedPacket.Snapshot {
			return domain.WorkspaceError(apierrors.ContextStale, "session packet changed after its actual project observation")
		}
		if err = validateProjectSnapshotReference(tx.Context(), s.snapshots, c, packet.Snapshot); err != nil {
			return err
		}
		if err = s.owned.ValidateOwnedSessionFacts(tx.Context(), c, facts, OwnedPurposeBind); err != nil {
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
	observed, err := s.ObserveSessionBindings(ctx, check, ids)
	if err != nil {
		return err
	}
	return s.WithObservedSessionBindings(ctx, check, observed, fn)
}
