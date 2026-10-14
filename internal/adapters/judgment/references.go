package judgmentadapter

import (
	"context"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type AttentionReferences struct {
	Attention attention.AttentionProjectReferences
}

var _ workspace.ProjectReferences = AttentionReferences{}

func (b AttentionReferences) ReadSelected(ctx context.Context, caller domain.Caller, space string, ref domain.FixedReference, expanded bool) (domain.ContextMaterial, error) {
	result, err := b.Attention.ReadProjectReference(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, space, attention.SourceRef{MaterialID: ref.MaterialID, Revision: ref.Revision}, expanded)
	if err != nil {
		return domain.ContextMaterial{}, err
	}
	return domain.ContextMaterial{Reference: domain.FixedReference{MaterialID: result.Ref.MaterialID, Revision: result.Ref.Revision}, Title: result.Title, Text: result.Text, Expanded: expanded}, nil
}

func (b AttentionReferences) Linked(ctx context.Context, caller domain.Caller, space string, ref domain.FixedReference) ([]domain.FixedReference, error) {
	refs, err := b.Attention.LinkedProjectReferences(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, space, attention.SourceRef{MaterialID: ref.MaterialID, Revision: ref.Revision})
	if err != nil {
		return nil, err
	}
	result := []domain.FixedReference{}
	for _, ref := range refs {
		result = append(result, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision})
	}
	return result, nil
}
