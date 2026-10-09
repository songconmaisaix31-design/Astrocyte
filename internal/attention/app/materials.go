package app

import (
	"context"
	"crypto/rand"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *Service) machineRead(tx AttentionTx, row *MaterialDetail, p Principal) error {
	version := row.Material.Version
	row.Uses = append(row.Uses, Usage{ID: rand.Text(), MaterialID: row.Material.ID, ActorID: p.ID, ActorKind: "agent", Action: "read", OccurredAt: s.options.Clock()})
	row.Material.AgentUsageCount++
	row.Material.Version++
	if err := tx.SaveMaterial(*row, version); err != nil {
		return err
	}
	return s.event(tx, "material_used", row.Material.ID, row.Material.Version, CommandMeta{}, map[string]any{"material_id": row.Material.ID, "actor_kind": "agent", "action": "read"})
}

func (s *Service) RecordUse(ctx context.Context, p Principal, id string, c RecordUseCommand) (VersionResult, error) {
	input := struct {
		ID string
		RecordUseCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "RecordUse", input, func(tx AttentionTx) (VersionResult, error) {
		if !slices.Contains([]string{"annotate", "reread", "adopt", "mention", "project_reuse"}, c.Action) {
			return VersionResult{}, domain.ErrInvalid
		}
		row, err := tx.LoadMaterial(id)
		if err != nil {
			return VersionResult{}, err
		}
		if row.Material.Version != c.ExpectedVersion {
			return VersionResult{}, domain.ErrVersion
		}
		if row.Material.Lifecycle == "withdrawn" {
			return VersionResult{}, serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
		}
		if err = s.humanSignal(tx, &row, p, c.Action, c.CommandMeta); err != nil {
			return VersionResult{}, err
		}
		return VersionResult{SchemaVersion: 1, ID: id, Version: row.Material.Version}, nil
	})
}

func (s *Service) UpdateMaterial(ctx context.Context, p Principal, id string, c UpdateMaterialCommand) (MaterialDetail, error) {
	input := struct {
		ID string
		UpdateMaterialCommand
	}{id, c}
	row, err := command(s, ctx, p, c.CommandMeta, "UpdateMaterial", input, func(tx AttentionTx) (MaterialDetail, error) {
		if !slices.Contains([]string{"active", "archived", "withdrawn"}, c.Lifecycle) {
			return MaterialDetail{}, domain.ErrInvalid
		}
		row, err := tx.LoadMaterial(id)
		if err != nil {
			return MaterialDetail{}, err
		}
		if row.Material.Version != c.ExpectedVersion {
			return MaterialDetail{}, domain.ErrVersion
		}
		row.Material.Pinned = c.Pinned
		row.Material.CollectionReason = c.CollectionReason
		row.Material.Lifecycle = c.Lifecycle
		row.Material.Version++
		// Pinning and lifecycle are independent metadata, never content versions
		// or implicit reread events. Collection reasons remain human-authored.
		if err = tx.SaveMaterial(row, c.ExpectedVersion); err != nil {
			return MaterialDetail{}, err
		}
		if err = s.event(tx, "material_updated", id, row.Material.Version, c.CommandMeta, map[string]any{"material_id": id, "lifecycle": c.Lifecycle, "pinned": c.Pinned}); err != nil {
			return MaterialDetail{}, err
		}
		row, err = materialProjection(tx, row)
		if err != nil {
			return MaterialDetail{}, err
		}
		return s.projectAttention(row), nil
	})
	return row, err
}

func (s *Service) GetAttachment(ctx context.Context, p Principal, id string, revision int, name string) (AttachmentContent, error) {
	var attachment AttachmentRef
	if err := authorize(p, false); err != nil {
		return AttachmentContent{}, err
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		row, err := tx.LoadMaterial(id)
		if err != nil {
			return err
		}
		if row.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
		}
		for _, r := range row.Revisions {
			if r.Revision != revision {
				continue
			}
			for _, a := range r.Attachments {
				if a.Name == name {
					attachment = a
					if p.Kind == "agent" {
						return s.machineRead(tx, &row, p)
					}
					return nil
				}
			}
		}
		return apierrors.NewNotFound("attachment", name)
	})
	if err != nil {
		return AttachmentContent{}, mapError(err, "")
	}
	data, err := s.objects.Read(ctx, attachment.ObjectRef)
	if err == nil {
		err = s.checkMaterialAccess(ctx, id)
	}
	if err != nil {
		return AttachmentContent{}, mapError(err, "")
	}
	return AttachmentContent{Data: data, MediaType: attachment.MediaType, Name: attachment.Name}, mapError(err, "")
}
