package main

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// Composition only: Attention owns membership/material access, Workspace owns
// project/action/model approval. Cross-app types are converted at this entrypoint.
type projectReferenceBridge struct {
	attention attentionapp.AttentionProjectReferences
}

func (b *projectReferenceBridge) ReadSelected(ctx context.Context, caller domain.Caller, spaceID string, ref domain.FixedReference, expanded bool) (domain.ContextMaterial, error) {
	if b.attention == nil {
		return domain.ContextMaterial{}, apierrors.NewUnsupported("project_material_access")
	}
	snapshot, err := b.attention.ReadProjectReference(ctx, attentionapp.Principal{ID: caller.ID, Kind: caller.Kind}, spaceID, attentionapp.SourceRef{MaterialID: ref.MaterialID, Revision: ref.Revision}, expanded)
	if err != nil {
		return domain.ContextMaterial{}, err
	}
	return domain.ContextMaterial{Reference: domain.FixedReference{MaterialID: snapshot.Ref.MaterialID, Revision: snapshot.Ref.Revision}, Title: snapshot.Title, Text: snapshot.Text}, nil
}
func (b *projectReferenceBridge) Linked(ctx context.Context, caller domain.Caller, spaceID string, ref domain.FixedReference) ([]domain.FixedReference, error) {
	if b.attention == nil {
		return nil, apierrors.NewUnsupported("project_material_access")
	}
	refs, err := b.attention.LinkedProjectReferences(ctx, attentionapp.Principal{ID: caller.ID, Kind: caller.Kind}, spaceID, attentionapp.SourceRef{MaterialID: ref.MaterialID, Revision: ref.Revision})
	if err != nil {
		return nil, err
	}
	result := make([]domain.FixedReference, 0, len(refs))
	for _, r := range refs {
		result = append(result, domain.FixedReference{MaterialID: r.MaterialID, Revision: r.Revision})
	}
	return result, nil
}

type selectedTextBridge struct {
	projects  workspaceapp.LocalProjects
	processor workspaceapp.TextProcessor
}

func (b *selectedTextBridge) ProjectSpaceID(ctx context.Context, p attentionapp.Principal, projectID, cli string) (string, error) {
	if b.projects == nil {
		return "", apierrors.NewUnsupported("project_model_processor")
	}
	project, err := b.projects.CheckProjectModel(ctx, domain.Caller{ID: p.ID, Kind: p.Kind, ProjectID: projectID}, projectID, cli)
	if err != nil {
		return "", err
	}
	return project.SpaceID, nil
}

func (b *selectedTextBridge) ConfigurationID(ctx context.Context, p attentionapp.Principal, projectID, cli string) (string, error) {
	if b.projects == nil || b.processor == nil {
		return "", apierrors.NewUnsupported("project_model_processor")
	}
	if _, err := b.projects.CheckProjectModel(ctx, domain.Caller{ID: p.ID, Kind: p.Kind, ProjectID: projectID}, projectID, cli); err != nil {
		return "", err
	}
	return b.processor.ConfigurationID(ctx, cli)
}
func (b *selectedTextBridge) ProcessSelectedText(ctx context.Context, p attentionapp.Principal, r attentionapp.SelectedTextRequest) (attentionapp.SelectedTextResult, error) {
	if b.projects == nil || b.processor == nil {
		return attentionapp.SelectedTextResult{}, apierrors.NewUnsupported("project_model_processor")
	}
	if _, err := b.projects.CheckProjectModel(ctx, domain.Caller{ID: p.ID, Kind: p.Kind, ProjectID: r.ProjectID}, r.ProjectID, r.CLI); err != nil {
		return attentionapp.SelectedTextResult{}, err
	}
	result, err := b.processor.ProcessSelectedText(ctx, domain.TextRequest{CLI: r.CLI, Prompt: r.Prompt, DeadlineSeconds: r.DeadlineSeconds, OutputLimit: r.OutputLimit})
	if err != nil {
		return attentionapp.SelectedTextResult{}, err
	}
	return attentionapp.SelectedTextResult{Text: result.Text, NativeID: result.NativeID, Version: result.Version, Model: result.Model}, nil
}

var _ workspaceapp.ProjectReferences = (*projectReferenceBridge)(nil)
var _ attentionapp.SelectedTextProcessor = (*selectedTextBridge)(nil)

type repositorySpaceBridge struct {
	attention interface {
		GetProjectSpace(context.Context, attentionapp.Principal, string) (attentionapp.ProjectSpaceResult, error)
	}
}

// Existing ProjectSpaces are the human's top-level development spaces. This
// bridge only checks existence/access; it never creates spaces or grants.
func (b *repositorySpaceBridge) ValidateRepositorySpace(ctx context.Context, c domain.Caller, spaceID string) error {
	if c.Kind != "human" || c.ID == "" {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "repository placement requires the human's selected top-level space", RequiredAction: "human_select_repository_space"}
	}
	if b.attention == nil {
		return apierrors.NewUnsupported("repository_space_placement")
	}
	_, err := b.attention.GetProjectSpace(ctx, attentionapp.Principal{ID: c.ID, Kind: c.Kind}, spaceID)
	return err
}

var _ workspaceapp.RepositorySpaces = (*repositorySpaceBridge)(nil)
