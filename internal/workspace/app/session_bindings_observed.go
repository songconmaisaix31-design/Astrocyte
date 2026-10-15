package app

import (
	"context"
	"slices"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (s *S2Service) requireWorkspaceObservation(ctx context.Context) error {
	guard, ok := s.repo.(WorkspaceObservationGuard)
	if !ok {
		return domain.WorkspaceError(apierrors.UnsupportedCapability, "actual Workspace transaction observation guard is required")
	}
	return guard.RequireWorkspaceObservation(ctx)
}

func validateProjectSnapshotReference(ctx context.Context, snapshots ProjectSnapshotProvider, c domain.Caller, snapshot domain.ProjectContextSnapshot) error {
	validator, ok := snapshots.(interface {
		ValidateProjectSnapshotReference(context.Context, domain.Caller, domain.ProjectContextSnapshot) error
	})
	if !ok {
		return domain.WorkspaceError(apierrors.UnsupportedCapability, "actual transaction-local project snapshot reference validator is required")
	}
	return validator.ValidateProjectSnapshotReference(ctx, c, snapshot)
}

func (s *S2Service) ObserveSessionBindings(ctx context.Context, check domain.GrantCheck, ids []string) (observed *SessionBindingObservation, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if check.Action != OwnedPurposeClaim && check.Action != OwnedPurposeSubmit || len(ids) == 0 {
		return nil, domain.WorkspaceError(apierrors.ValidationFailed, "binding observation requires an explicit protected member purpose and binding")
	}
	if err = s.requireWorkspaceObservation(ctx); err != nil {
		return nil, err
	}
	if s.owned == nil {
		return nil, domain.WorkspaceError(apierrors.UnsupportedCapability, "actual owned Registry provider is required")
	}
	observed = &SessionBindingObservation{service: s, request: ctx, check: check}
	observed.check.Tools = slices.Clone(check.Tools)
	observed.check.ExternalTargets = slices.Clone(check.ExternalTargets)
	observed.check.WritePaths = slices.Clone(check.WritePaths)
	err = s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		for _, id := range ids {
			binding, err := tx.LoadSessionBinding(id)
			if err != nil {
				return err
			}
			if binding.Status != "bound" || binding.ProjectID != check.ProjectID || binding.GrantID != check.GrantID || binding.GrantEpoch != check.Epoch {
				return domain.WorkspaceError(apierrors.ApprovalRevoked, "member binding is outside requested approval")
			}
			packet, err := tx.LoadContextPacket(binding.PacketID)
			if err != nil {
				return err
			}
			if _, _, err = packetGrant(tx, check.Caller, binding.GrantID, binding.GrantEpoch); err != nil {
				return err
			}
			if packet.ProjectID != binding.ProjectID || packet.GrantID != binding.GrantID || packet.GrantEpoch != binding.GrantEpoch || packet.Snapshot.ProjectID != binding.ProjectID || packet.Snapshot.WorktreeID != binding.WorktreeID {
				return domain.WorkspaceError(apierrors.ContextStale, "member binding differs from its immutable context packet")
			}
			observed.entries = append(observed.entries, sessionBindingObservationEntry{binding: binding, packet: packet})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for i := range observed.entries {
		entry := &observed.entries[i]
		entry.facts, err = s.owned.ReadOwnedSessionFacts(ctx, check.Caller, entry.binding.SessionID, check.Action)
		if err != nil {
			return nil, err
		}
		if err = matchOwnedBinding(entry.binding, entry.packet, entry.facts, check.Action); err != nil {
			return nil, err
		}
	}
	return observed, nil
}

func matchOwnedBinding(binding domain.SessionBinding, packet domain.ImmutableContextPacket, facts domain.OwnedSessionFacts, purpose string) error {
	if !facts.Owned || facts.SessionID != binding.SessionID || facts.ProjectID != binding.ProjectID || facts.NativeID != binding.NativeSessionID || facts.WorktreeID != binding.WorktreeID || facts.Root != packet.Snapshot.Root || facts.AdapterID != binding.AdapterID || binding.ControllerID == "" || binding.StartupOperationID == "" || facts.ControllerID != binding.ControllerID || facts.StartupOperationID != binding.StartupOperationID {
		return domain.WorkspaceError(apierrors.EvidenceMissing, "current Registry controller tuple does not match member binding")
	}
	if facts.Liveness != "live" && (purpose != OwnedPurposeSubmit || facts.Liveness != "exited") {
		return domain.WorkspaceError(apierrors.EvidenceMissing, "controller lifecycle does not permit the protected action")
	}
	return nil
}

func (s *S2Service) WithObservedSessionBindings(ctx context.Context, check domain.GrantCheck, observed *SessionBindingObservation, fn func(context.Context, domain.ApprovalGrant, []domain.SessionBinding) error) (err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if observed == nil || observed.service != s || observed.request == nil || len(observed.entries) == 0 || observed.request.Done() != ctx.Done() {
		return domain.WorkspaceError(apierrors.EvidenceMissing, "binding observation must belong to this service and request")
	}
	if err = observed.request.Err(); err != nil {
		return err
	}
	before := observed.check
	if before.Caller != check.Caller || before.GrantID != check.GrantID || before.Epoch != check.Epoch || before.ProposalRevision != check.ProposalRevision || before.ProjectID != check.ProjectID || before.Root != check.Root || before.Action != check.Action || !slices.Equal(before.Tools, check.Tools) || !slices.Equal(before.ExternalTargets, check.ExternalTargets) || !slices.Equal(before.WritePaths, check.WritePaths) {
		return domain.WorkspaceError(apierrors.ContextStale, "protected identity or execution scope changed after actual observation")
	}
	return s.WithAuthorizedGrant(ctx, check, func(txctx context.Context, grant domain.ApprovalGrant) error {
		return s.repo.WithWorkspaceTx(txctx, func(tx WorkspaceTx) error {
			bindings := make([]domain.SessionBinding, 0, len(observed.entries))
			for _, entry := range observed.entries {
				binding, err := tx.LoadSessionBinding(entry.binding.ID)
				if err != nil {
					return err
				}
				previous := entry.binding
				if binding.Version != previous.Version || binding.Status != "bound" || binding.GrantID != grant.ID || binding.GrantEpoch != grant.Epoch || binding.ProjectID != grant.Scope.ProjectID || binding.MemberID != previous.MemberID || binding.SessionID != previous.SessionID || binding.PacketID != previous.PacketID || binding.AdapterID != previous.AdapterID || binding.NativeSessionID != previous.NativeSessionID || binding.ControllerID != previous.ControllerID || binding.StartupOperationID != previous.StartupOperationID || binding.WorktreeID != previous.WorktreeID {
					return domain.WorkspaceError(apierrors.ApprovalRevoked, "member binding changed after actual observation")
				}
				packet, err := tx.LoadContextPacket(binding.PacketID)
				if err != nil {
					return err
				}
				if packet.ObjectRef != entry.packet.ObjectRef || packet.Snapshot != entry.packet.Snapshot {
					return domain.WorkspaceError(apierrors.ContextStale, "immutable binding packet changed after observation")
				}
				if err = s.owned.ValidateOwnedSessionFacts(tx.Context(), check.Caller, entry.facts, check.Action); err != nil {
					return err
				}
				member, err := tx.LoadProjectGrant(binding.ProjectID, binding.MemberID)
				if err != nil {
					return err
				}
				if member.RevokedAt != nil || member.ExpiresAt != nil && !member.ExpiresAt.After(time.Now()) {
					return domain.WorkspaceError(apierrors.ApprovalRevoked, "member project grant is inactive")
				}
				if !has(member.Actions, check.Action) {
					return domain.WorkspaceError(apierrors.ScopeDenied, "member action is not authorized")
				}
				session, err := tx.LoadNativeSession(binding.SessionID)
				if err != nil {
					return err
				}
				if session.NativeID != binding.NativeSessionID || session.CLI != binding.AdapterID || session.ProjectID != binding.ProjectID || session.Ownership != "owned" || session.StopConfirmed && check.Action != OwnedPurposeSubmit {
					return domain.WorkspaceError(apierrors.EvidenceMissing, "bound native session is no longer permitted for this action")
				}
				bindings = append(bindings, binding)
			}
			return fn(tx.Context(), grant, bindings)
		})
	})
}
