package app

import (
	"context"
	"crypto/rand"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *Service) ListMaterialDomains(ctx context.Context, p Principal) (apierrors.ListResult, error) {
	if err := authorize(p, false); err != nil {
		return apierrors.EmptyList(), err
	}
	var result []MaterialDomain
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error { var err error; result, err = tx.ListMaterialDomains(); return err })
	return listResult(result), mapError(err, "")
}

func (s *Service) CreateMaterialDomain(ctx context.Context, p Principal, c MaterialDomainCommand) (MaterialDomainResult, error) {
	return command(s, ctx, p, c.CommandMeta, "CreateMaterialDomain", c, func(tx AttentionTx) (MaterialDomainResult, error) {
		if c.ExpectedVersion != 1 || strings.TrimSpace(c.Title) == "" {
			return MaterialDomainResult{}, domain.ErrInvalid
		}
		d := MaterialDomain{ID: rand.Text(), Version: 1, Title: strings.TrimSpace(c.Title), Description: c.Description, CreatedAt: s.options.Clock()}
		if err := tx.SaveMaterialDomain(d, 0); err != nil {
			return MaterialDomainResult{}, err
		}
		if err := s.event(tx, "domain_created", d.ID, d.Version, c.CommandMeta, map[string]any{"domain_id": d.ID}); err != nil {
			return MaterialDomainResult{}, err
		}
		return MaterialDomainResult{SchemaVersion: 1, Domain: d}, nil
	})
}

func (s *Service) ReviseMaterialDomain(ctx context.Context, p Principal, id string, c MaterialDomainCommand) (MaterialDomainResult, error) {
	input := struct {
		ID string
		MaterialDomainCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "ReviseMaterialDomain", input, func(tx AttentionTx) (MaterialDomainResult, error) {
		if strings.TrimSpace(c.Title) == "" {
			return MaterialDomainResult{}, domain.ErrInvalid
		}
		d, err := tx.LoadMaterialDomain(id)
		if err != nil {
			return MaterialDomainResult{}, err
		}
		if d.Version != c.ExpectedVersion {
			return MaterialDomainResult{}, domain.ErrVersion
		}
		if d.Title == strings.TrimSpace(c.Title) && d.Description == c.Description {
			return MaterialDomainResult{SchemaVersion: 1, Domain: d}, nil
		}
		d.Title = strings.TrimSpace(c.Title)
		d.Description = c.Description
		d.Version++
		if err = tx.SaveMaterialDomain(d, c.ExpectedVersion); err != nil {
			return MaterialDomainResult{}, err
		}
		if err = s.event(tx, "domain_revised", id, d.Version, c.CommandMeta, map[string]any{"domain_id": id}); err != nil {
			return MaterialDomainResult{}, err
		}
		return MaterialDomainResult{SchemaVersion: 1, Domain: d}, nil
	})
}

func (s *Service) SetMaterialDomains(ctx context.Context, p Principal, id string, c SetMaterialDomainsCommand) (MaterialDetail, error) {
	c.DomainIDs = slices.Clone(c.DomainIDs)
	slices.Sort(c.DomainIDs)
	c.DomainIDs = nonNil(slices.Compact(c.DomainIDs))
	input := struct {
		ID string
		SetMaterialDomainsCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "SetMaterialDomains", input, func(tx AttentionTx) (MaterialDetail, error) {
		row, err := tx.LoadMaterial(id)
		if err != nil {
			return MaterialDetail{}, err
		}
		if row.Material.Version != c.ExpectedVersion {
			return MaterialDetail{}, domain.ErrVersion
		}
		for _, domainID := range c.DomainIDs {
			if _, err = tx.LoadMaterialDomain(domainID); err != nil {
				return MaterialDetail{}, err
			}
		}
		if !slices.Equal(row.Material.DomainIDs, c.DomainIDs) {
			row.Material.DomainIDs = c.DomainIDs
			if err = s.humanSignal(tx, &row, p, "classify", c.CommandMeta); err != nil {
				return MaterialDetail{}, err
			}
			if err = s.event(tx, "material_classified", id, row.Material.Version, c.CommandMeta, map[string]any{"material_id": id, "domain_ids": c.DomainIDs}); err != nil {
				return MaterialDetail{}, err
			}
		}
		row, err = materialProjection(tx, row)
		if err != nil {
			return MaterialDetail{}, err
		}
		return s.projectAttention(row), nil
	})
}

func (s *Service) ListProjectSpaces(ctx context.Context, p Principal) (apierrors.ListResult, error) {
	if err := authorize(p, false); err != nil {
		return apierrors.EmptyList(), err
	}
	var result []ProjectSpace
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		var err error
		result, err = tx.ListProjectSpaces()
		if err != nil {
			return err
		}
		for i := range result {
			result[i].MaterialRefs = nonNil(result[i].MaterialRefs)
		}
		return nil
	})
	return listResult(result), mapError(err, "")
}

func (s *Service) CreateProjectSpace(ctx context.Context, p Principal, c ProjectSpaceCommand) (ProjectSpaceResult, error) {
	return command(s, ctx, p, c.CommandMeta, "CreateProjectSpace", c, func(tx AttentionTx) (ProjectSpaceResult, error) {
		if c.ExpectedVersion != 1 || strings.TrimSpace(c.Title) == "" {
			return ProjectSpaceResult{}, domain.ErrInvalid
		}
		space := ProjectSpace{ID: rand.Text(), Version: 1, Title: strings.TrimSpace(c.Title), MaterialRefs: []SourceRef{}, CreatedAt: s.options.Clock()}
		if err := tx.SaveProjectSpace(space, 0); err != nil {
			return ProjectSpaceResult{}, err
		}
		if err := s.event(tx, "project_space_created", space.ID, space.Version, c.CommandMeta, map[string]any{"space_id": space.ID}); err != nil {
			return ProjectSpaceResult{}, err
		}
		return ProjectSpaceResult{SchemaVersion: 1, Space: space}, nil
	})
}

func (s *Service) GetProjectSpace(ctx context.Context, p Principal, id string) (ProjectSpaceResult, error) {
	if err := authorize(p, false); err != nil {
		return ProjectSpaceResult{}, err
	}
	var space ProjectSpace
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error { var err error; space, err = tx.LoadProjectSpace(id); return err })
	space.MaterialRefs = nonNil(space.MaterialRefs)
	return ProjectSpaceResult{SchemaVersion: 1, Space: space}, mapError(err, "")
}

func (s *Service) ReferenceMaterial(ctx context.Context, p Principal, id string, c ReferenceMaterialCommand) (ProjectSpaceResult, error) {
	input := struct {
		ID string
		ReferenceMaterialCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "ReferenceMaterial", input, func(tx AttentionTx) (ProjectSpaceResult, error) {
		if c.MaterialID == "" || c.Revision < 1 {
			return ProjectSpaceResult{}, domain.ErrInvalid
		}
		space, err := tx.LoadProjectSpace(id)
		if err != nil {
			return ProjectSpaceResult{}, err
		}
		if space.Version != c.ExpectedVersion {
			return ProjectSpaceResult{}, domain.ErrVersion
		}
		row, err := tx.LoadMaterial(c.MaterialID)
		if err != nil {
			return ProjectSpaceResult{}, err
		}
		if row.Material.Lifecycle == "withdrawn" {
			return ProjectSpaceResult{}, serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
		}
		var reference SourceRef
		for _, revision := range row.Revisions {
			if revision.Revision == c.Revision {
				reference = SourceRef{MaterialID: c.MaterialID, Revision: c.Revision, Locator: revision.SourceLocator}
				break
			}
		}
		if reference.MaterialID == "" {
			return ProjectSpaceResult{}, apierrors.NewNotFound("material_revision", c.MaterialID)
		}
		space.MaterialRefs = nonNil(slices.Clone(space.MaterialRefs))
		replaced := false
		for i, ref := range space.MaterialRefs {
			if ref.MaterialID != c.MaterialID {
				continue
			}
			if ref.Revision == c.Revision && ref.Locator == reference.Locator && ref.Span == nil {
				return ProjectSpaceResult{SchemaVersion: 1, Space: space}, nil
			}
			space.MaterialRefs[i] = reference
			replaced = true
			break
		}
		if !replaced {
			space.MaterialRefs = append(space.MaterialRefs, reference)
		}
		space.Version++
		if err = tx.SaveProjectSpace(space, c.ExpectedVersion); err != nil {
			return ProjectSpaceResult{}, err
		}
		if err = s.humanSignal(tx, &row, p, "mention", c.CommandMeta); err != nil {
			return ProjectSpaceResult{}, err
		}
		if err = s.event(tx, "project_space_referenced", id, space.Version, c.CommandMeta, map[string]any{"space_id": id, "material_id": c.MaterialID, "revision": c.Revision}); err != nil {
			return ProjectSpaceResult{}, err
		}
		return ProjectSpaceResult{SchemaVersion: 1, Space: space}, nil
	})
}

func (s *Service) RemoveMaterialReference(ctx context.Context, p Principal, id string, c RemoveMaterialReferenceCommand) (ProjectSpaceResult, error) {
	input := struct {
		ID string
		RemoveMaterialReferenceCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "RemoveMaterialReference", input, func(tx AttentionTx) (ProjectSpaceResult, error) {
		if c.MaterialID == "" {
			return ProjectSpaceResult{}, domain.ErrInvalid
		}
		space, err := tx.LoadProjectSpace(id)
		if err != nil {
			return ProjectSpaceResult{}, err
		}
		if space.Version != c.ExpectedVersion {
			return ProjectSpaceResult{}, domain.ErrVersion
		}
		oldLength := len(space.MaterialRefs)
		space.MaterialRefs = nonNil(slices.DeleteFunc(slices.Clone(space.MaterialRefs), func(ref SourceRef) bool { return ref.MaterialID == c.MaterialID }))
		if oldLength == len(space.MaterialRefs) {
			return ProjectSpaceResult{SchemaVersion: 1, Space: space}, nil
		}
		space.Version++
		if err = tx.SaveProjectSpace(space, c.ExpectedVersion); err != nil {
			return ProjectSpaceResult{}, err
		}
		if err = s.event(tx, "project_space_reference_removed", id, space.Version, c.CommandMeta, map[string]any{"space_id": id, "material_id": c.MaterialID}); err != nil {
			return ProjectSpaceResult{}, err
		}
		return ProjectSpaceResult{SchemaVersion: 1, Space: space}, nil
	})
}

func (s *Service) humanSignal(tx AttentionTx, row *MaterialDetail, p Principal, action string, m CommandMeta) error {
	old := row.Material.Version
	row.Uses = append(row.Uses, Usage{ID: rand.Text(), MaterialID: row.Material.ID, ActorID: p.ID, ActorKind: "human", Action: action, OccurredAt: s.options.Clock()})
	row.Material.HumanUsageCount++
	row.Material.Version++
	if action == "adopt" || action == "project_reuse" {
		row.Material.LongTermValue++
	}
	if err := tx.SaveMaterial(*row, old); err != nil {
		return err
	}
	return s.event(tx, "material_used", row.Material.ID, row.Material.Version, m, map[string]any{"material_id": row.Material.ID, "actor_kind": "human", "action": action})
}
