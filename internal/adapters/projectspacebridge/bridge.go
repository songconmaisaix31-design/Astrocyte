package projectspacebridge

import (
	"context"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// Attention 显式消费 Workspace 的空间，不拥有另一份创建或存储逻辑。
type Attention struct{ Spaces workspace.ProjectSpaces }

func ToAttention(space domain.ProjectSpace) attention.ProjectSpace {
	refs := make([]attention.SourceRef, 0, len(space.MaterialRefs))
	for _, ref := range space.MaterialRefs {
		refs = append(refs, attention.SourceRef{MaterialID: ref.MaterialID, Revision: ref.Revision, Locator: ref.Locator, Span: ref.Span})
	}
	return attention.ProjectSpace{ID: space.ID, Version: space.Version, Title: space.Title, MaterialRefs: refs, CreatedAt: space.CreatedAt}
}

func ToWorkspace(space attention.ProjectSpace) domain.ProjectSpace {
	refs := make([]domain.ProjectSpaceReference, 0, len(space.MaterialRefs))
	for _, ref := range space.MaterialRefs {
		refs = append(refs, domain.ProjectSpaceReference{MaterialID: ref.MaterialID, Revision: ref.Revision, Locator: ref.Locator, Span: ref.Span})
	}
	return domain.ProjectSpace{ID: space.ID, Version: space.Version, Title: space.Title, MaterialRefs: refs, CreatedAt: space.CreatedAt}
}

func (b Attention) ListProjectSpaces(ctx context.Context, p attention.Principal) (apierrors.ListResult, error) {
	if b.Spaces == nil {
		return apierrors.EmptyList(), apierrors.NewUnsupported("project_spaces")
	}
	result, err := b.Spaces.ListProjectSpaces(ctx, domain.Caller{ID: p.ID, Kind: p.Kind})
	if err != nil {
		return result, err
	}
	for index, item := range result.Items {
		space, ok := item.(domain.ProjectSpace)
		if !ok {
			return apierrors.EmptyList(), fmt.Errorf("Workspace returned an invalid project space")
		}
		result.Items[index] = ToAttention(space)
	}
	return result, nil
}

func (b Attention) GetProjectSpace(ctx context.Context, p attention.Principal, id string) (attention.ProjectSpaceResult, error) {
	if b.Spaces == nil {
		return attention.ProjectSpaceResult{}, apierrors.NewUnsupported("project_spaces")
	}
	result, err := b.Spaces.GetProjectSpace(ctx, domain.Caller{ID: p.ID, Kind: p.Kind}, id)
	return attention.ProjectSpaceResult{SchemaVersion: result.SchemaVersion, Space: ToAttention(result.Space)}, err
}

func (b Attention) CreateProjectSpace(ctx context.Context, p attention.Principal, command attention.ProjectSpaceCommand) (attention.ProjectSpaceResult, error) {
	if b.Spaces == nil {
		return attention.ProjectSpaceResult{}, apierrors.NewUnsupported("project_spaces")
	}
	result, err := b.Spaces.CreateProjectSpace(ctx, domain.Caller{ID: p.ID, Kind: p.Kind}, workspace.ProjectSpaceCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: command.SchemaVersion, RequestID: command.RequestID, ExpectedVersion: command.ExpectedVersion, IdempotencyKey: command.IdempotencyKey}, Title: command.Title})
	return attention.ProjectSpaceResult{SchemaVersion: result.SchemaVersion, Space: ToAttention(result.Space)}, err
}
