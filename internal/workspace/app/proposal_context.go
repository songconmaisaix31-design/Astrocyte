package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func packetGrant(tx WorkspaceTx, c domain.Caller, id string, epoch int) (g domain.ApprovalGrant, p domain.LocalProject, err error) {
	g, err = tx.LoadApproval(id)
	if err != nil {
		return
	}
	if g.Epoch != epoch {
		err = domain.WorkspaceError(apierrors.VersionConflict, "packet grant epoch changed")
		return
	}
	if g.Status != "active" || !g.ExpiresAt.After(time.Now()) || !g.Stops.Deadline.After(time.Now()) {
		err = domain.WorkspaceError(apierrors.ApprovalRevoked, "packet approval is inactive")
		return
	}
	p, err = projectAccess(tx, c, g.Scope.ProjectID)
	if err != nil {
		return
	}
	if g.Scope.SettingsRevision != p.Settings.Revision {
		err = domain.WorkspaceError(apierrors.ApprovalRevoked, "project settings changed since context approval")
		return
	}
	if !has(g.Scope.Actions, "read_context") {
		err = domain.WorkspaceError(apierrors.ScopeDenied, "approval does not permit context access")
		return
	}
	err = ValidateProjectActionScope(p, ProjectActionScope{ProjectID: p.ID, SettingsRevision: g.Scope.SettingsRevision, Action: "read_context", Roots: g.Scope.Roots})
	return
}

func (s *S2Service) BuildContextPacket(ctx context.Context, c domain.Caller, cmd BuildContextPacketCommand) (packet domain.ImmutableContextPacket, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if s.snapshots == nil || s.objects == nil {
		return packet, domain.WorkspaceError(apierrors.UnsupportedCapability, "fixed project snapshot and object store are required")
	}
	if cmd.ExpectedVersion != 0 || cmd.MaxBytes < 1 || cmd.MaxBytes > 2*1024*1024 {
		return packet, domain.WorkspaceError(apierrors.ValidationFailed, "new packet requires zero version and explicit bounded max_bytes")
	}
	var grant domain.ApprovalGrant
	var project domain.LocalProject
	var proposal domain.Proposal
	var goal domain.Goal
	err = s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		var err error
		grant, project, err = packetGrant(tx, c, cmd.GrantID, cmd.GrantEpoch)
		if err != nil {
			return err
		}
		proposal, err = tx.LoadProposal(grant.ProposalID, grant.ProposalRevision)
		if err != nil {
			return err
		}
		goal, err = tx.LoadGoal(grant.Goal.ID, grant.Goal.Revision)
		return err
	})
	if err != nil {
		return
	}
	snapshot, err := s.snapshots.CaptureProjectSnapshot(ctx, c, project.ID, grant.Scope.Roots)
	if err != nil {
		return packet, err
	}
	if snapshot.ProjectID != project.ID || snapshot.EnvironmentID == "" || snapshot.RepoID == "" || snapshot.WorktreeID == "" || snapshot.Head == "" || snapshot.IndexDigest == "" || snapshot.WorktreeDigest == "" || snapshot.Generation == "" || snapshot.ObservedAt.IsZero() {
		return packet, domain.WorkspaceError(apierrors.EvidenceMissing, "snapshot provider omitted fixed project identity")
	}
	packet = domain.ImmutableContextPacket{ID: uuid.NewString(), SchemaVersion: 1, ProjectID: project.ID, ProjectSettingsRevision: project.Settings.Revision, PredecessorID: cmd.PredecessorID, GrantID: grant.ID, GrantEpoch: grant.Epoch, ProposalID: proposal.ID, ProposalRevision: proposal.Revision, Goal: goal, Snapshot: snapshot, Scope: grant.Scope, Stops: grant.Stops, Sources: []domain.PacketSource{}, Decisions: cmd.Decisions, Results: cmd.Results, Checks: cmd.Checks, Failures: cmd.Failures, OpenQuestions: cmd.OpenQuestions, NextSteps: cmd.NextSteps, CreatedAt: time.Now().UTC()}
	remaining := cmd.MaxBytes
	for _, ref := range proposal.Content.Inputs {
		if s.references == nil {
			return packet, domain.WorkspaceError(apierrors.UnsupportedCapability, "project reference reader is required")
		}
		material, err := s.references.ReadSelected(ctx, c, project.SpaceID, ref, false)
		if err != nil {
			return packet, err
		}
		if material.Reference != ref {
			return packet, domain.WorkspaceError(apierrors.ContextStale, "source fixed revision differs")
		}
		remaining -= len(material.Text)
		if remaining < 0 {
			return packet, domain.WorkspaceError(apierrors.BudgetExhausted, "packet source exceeds explicit budget")
		}
		refObject, err := s.objects.Publish(ctx, []byte(material.Text))
		if err != nil {
			return packet, err
		}
		packet.Sources = append(packet.Sources, domain.PacketSource{Reference: ref, ObjectRef: refObject, Title: material.Title, Locations: []string{}})
	}
	data, err := json.Marshal(packet)
	if err != nil {
		return packet, err
	}
	if len(data) > remaining {
		return packet, domain.WorkspaceError(apierrors.BudgetExhausted, "packet constraints and sources exceed budget")
	}
	packet.ObjectRef, err = s.objects.Publish(ctx, data)
	if err != nil {
		return packet, err
	}
	if err = s.snapshots.ValidateProjectSnapshot(ctx, c, snapshot); err != nil {
		return packet, err
	}
	err = s.command(ctx, c, cmd.WorkspaceCommandMeta, "workspace.context.build", cmd, &packet, func(tx WorkspaceTx) error {
		_, current, err := packetGrant(tx, c, grant.ID, grant.Epoch)
		if err != nil {
			return err
		}
		if current.Settings.Revision != project.Settings.Revision {
			return domain.WorkspaceError(apierrors.ContextStale, "project permissions changed while building packet")
		}
		for _, root := range grant.Scope.Roots {
			if err = s.validateProjectRoot(current, root); err != nil {
				return err
			}
		}
		for _, source := range packet.Sources {
			material, err := s.references.ReadSelected(tx.Context(), c, current.SpaceID, source.Reference, false)
			if err != nil {
				return err
			}
			if material.Reference != source.Reference {
				return domain.WorkspaceError(apierrors.ContextStale, "packet source revision changed before commit")
			}
		}
		if packet.PredecessorID != "" {
			previous, err := tx.LoadContextPacket(packet.PredecessorID)
			if err != nil {
				return err
			}
			if previous.ProjectID != packet.ProjectID {
				return domain.WorkspaceError(apierrors.ScopeDenied, "packet predecessor belongs to another project")
			}
		}
		if err = s.snapshots.ValidateProjectSnapshot(tx.Context(), c, snapshot); err != nil {
			return err
		}
		if err = tx.SaveContextPacket(packet); err != nil {
			return err
		}
		return workspaceEvent(tx, cmd.WorkspaceCommandMeta, "context_packet_built", packet.ID, 1, map[string]any{"packet_id": packet.ID, "snapshot_digest": packet.Snapshot.WorktreeDigest, "predecessor_id": packet.PredecessorID})
	})
	return
}

func (s *S2Service) ReadContextPacket(ctx context.Context, c domain.Caller, id string) (packet domain.ImmutableContextPacket, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	err = s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		var err error
		packet, err = tx.LoadContextPacket(id)
		if err != nil {
			return err
		}
		_, err = projectAccess(tx, c, packet.ProjectID)
		return err
	})
	return
}
