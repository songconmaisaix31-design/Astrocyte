package app

import (
	"context"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

var _ AttentionProjectReferences = (*Service)(nil)

func fixedReferenceEqual(a, b SourceRef) bool {
	return a.MaterialID == b.MaterialID && a.Revision == b.Revision
}

func loadFixedReference(tx AttentionTx, ref SourceRef) (MaterialDetail, MaterialRevision, error) {
	if ref.MaterialID == "" || ref.Revision < 1 {
		return MaterialDetail{}, MaterialRevision{}, domain.ErrInvalid
	}
	material, err := tx.LoadMaterial(ref.MaterialID)
	if err != nil {
		return material, MaterialRevision{}, err
	}
	if material.Material.Lifecycle == "withdrawn" {
		return material, MaterialRevision{}, serviceError(apierrors.ScopeDenied, "Source material was withdrawn", "choose_active_project_reference")
	}
	for _, revision := range material.Revisions {
		if revision.Revision == ref.Revision {
			if ref.Locator != "" && ref.Locator != revision.SourceLocator {
				return material, revision, domain.ErrEvidence
			}
			if ref.Span != nil && !slices.Contains(revision.SourceSpans, *ref.Span) {
				return material, revision, domain.ErrEvidence
			}
			return material, revision, nil
		}
	}
	return material, MaterialRevision{}, apierrors.NewNotFound("material_revision", ref.MaterialID)
}

func linkedProjectReferences(tx AttentionTx, space ProjectSpace, root SourceRef) ([]SourceRef, error) {
	membership := false
	for _, member := range space.MaterialRefs {
		if fixedReferenceEqual(member, root) {
			membership = true
			root = member
			break
		}
	}
	if !membership {
		return nil, serviceError(apierrors.ScopeDenied, "Root is not a fixed reference in this project space", "select_included_project_reference")
	}
	if _, _, err := loadFixedReference(tx, root); err != nil {
		return nil, err
	}
	all, err := tx.ListDistillations()
	if err != nil {
		return nil, err
	}
	result := []SourceRef{}
	for _, record := range all {
		if record.Status != "succeeded" {
			continue
		}
		matches := false
		for _, input := range record.InputRefs {
			if fixedReferenceEqual(input, root) {
				matches = true
				break
			}
		}
		if !matches {
			continue
		}
		for _, ref := range record.RelatedRefs {
			if fixedReferenceEqual(root, ref) {
				continue
			}
			if _, _, err := loadFixedReference(tx, ref); err != nil {
				if isRestricted(err) || isMissing(err) {
					continue
				}
				return nil, err
			}
			seen := false
			for _, old := range result {
				if fixedReferenceEqual(old, ref) {
					seen = true
					break
				}
			}
			if !seen {
				result = append(result, ref)
			}
		}
	}
	return result, nil
}

func projectReferenceAllowed(tx AttentionTx, spaceID string, ref SourceRef, expanded bool) error {
	space, err := tx.LoadProjectSpace(spaceID)
	if err != nil {
		return err
	}
	for _, member := range space.MaterialRefs {
		if fixedReferenceEqual(member, ref) {
			_, _, err = loadFixedReference(tx, member)
			return err
		}
	}
	if expanded {
		for _, root := range space.MaterialRefs {
			linked, err := linkedProjectReferences(tx, space, root)
			if err != nil {
				if isRestricted(err) || isMissing(err) {
					continue
				}
				return err
			}
			for _, candidate := range linked {
				if fixedReferenceEqual(candidate, ref) {
					return nil
				}
			}
		}
	}
	return serviceError(apierrors.ScopeDenied, "This fixed revision is outside the selected project reference scope", "include_fixed_reference_or_enable_linked_scope")
}

func (s *Service) LinkedProjectReferences(ctx context.Context, p Principal, spaceID string, root SourceRef) ([]SourceRef, error) {
	if p.Kind != "agent" || p.ID == "" {
		return nil, serviceError(apierrors.ScopeDenied, "Scoped project reads require a trusted Agent principal", "use_project_scoped_agent")
	}
	var result []SourceRef
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		space, err := tx.LoadProjectSpace(spaceID)
		if err != nil {
			return err
		}
		result, err = linkedProjectReferences(tx, space, root)
		return err
	})
	return nonNil(result), mapError(err, "")
}

func (s *Service) ReadProjectReference(ctx context.Context, p Principal, spaceID string, ref SourceRef, expanded bool) (SourceSnapshot, error) {
	if p.Kind != "agent" || p.ID == "" {
		return SourceSnapshot{}, serviceError(apierrors.ScopeDenied, "Scoped project reads require a trusted Agent principal", "use_project_scoped_agent")
	}
	if s.objects == nil {
		return SourceSnapshot{}, serviceError(apierrors.ProviderUnavailable, "Original object storage is unavailable", "configure_object_store")
	}
	var material MaterialDetail
	var revision MaterialRevision
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		if err := projectReferenceAllowed(tx, spaceID, ref, expanded); err != nil {
			return err
		}
		var err error
		material, revision, err = loadFixedReference(tx, ref)
		return err
	})
	if err != nil {
		return SourceSnapshot{}, mapError(err, "")
	}
	data, err := s.objects.Read(ctx, revision.ObjectRef)
	if err != nil {
		return SourceSnapshot{}, mapError(err, "")
	}
	// Membership/lifecycle can change during object I/O. Check again before
	// publishing bytes and recording Agent use; queries never emit human heat.
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		if err := projectReferenceAllowed(tx, spaceID, ref, expanded); err != nil {
			return err
		}
		row, _, err := loadFixedReference(tx, ref)
		if err != nil {
			return err
		}
		return s.machineRead(tx, &row, p)
	})
	if err != nil {
		return SourceSnapshot{}, mapError(err, "")
	}
	ref.Locator = revision.SourceLocator
	return SourceSnapshot{Ref: ref, SourceKey: revision.SourceKey, Title: material.Material.Title, Text: string(data), Summary: revision.Summary, ContentDigest: revision.ContentDigest, Provenance: revision.Provenance}, nil
}
