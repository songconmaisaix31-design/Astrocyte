package main

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// HTTP selects an existing trusted project context. The body cannot register
// a project or grant an Agent scope; MCP uses its separate startup binding.
type experienceHTTPBridge struct {
	db         *sqlite.DB
	events     attentionapp.EventReader
	projects   workspaceapp.LocalProjects
	registered workspaceapp.RegisteredProjects
}

var _ attentionapp.ExperienceGovernanceService = (*experienceHTTPBridge)(nil)

func (b *experienceHTTPBridge) service(ctx context.Context, p attentionapp.Principal, project string) (*attentionapp.ExperienceService, error) {
	denied := &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Experience requires a trusted registered project context", RequiredAction: "select_registered_project"}
	if p.Kind != "human" || p.ID == "" || project == "" {
		return nil, denied
	}
	caller := workspacedomain.Caller{Kind: p.Kind, ID: p.ID}
	known := false
	if b.projects != nil {
		projects, err := b.projects.ListProjects(ctx, caller)
		if err != nil {
			return nil, err
		}
		for _, v := range projects {
			if project == v.ID || project == v.Root {
				known = true
				break
			}
		}
	}
	if !known && b.registered != nil {
		snapshot, err := b.registered.ListRegisteredProjects(ctx, caller)
		if err != nil {
			return nil, err
		}
		for _, v := range snapshot.Projects {
			if project == v.ProjectKey || project == v.Root {
				known = true
				break
			}
		}
		for _, v := range snapshot.Board {
			if project == v.ID {
				known = true
				break
			}
		}
	}
	if !known {
		return nil, denied
	}
	return attentionapp.NewExperienceService(sqlite.NewExperienceRepository(b.db), b.events, attentionapp.ExperienceOptions{ProjectID: project}), nil
}

func (b *experienceHTTPBridge) ApproveExperience(ctx context.Context, p attentionapp.Principal, id string, c attentionapp.ApproveExperienceCommand) (attentionapp.ExperienceResult, error) {
	s, err := b.service(ctx, p, c.OwnerProject)
	if err != nil {
		return attentionapp.ExperienceResult{}, err
	}
	return s.ApproveExperience(ctx, p, id, c)
}
func (b *experienceHTTPBridge) ShareExperience(ctx context.Context, p attentionapp.Principal, id string, c attentionapp.ShareMemoryCommand) (attentionapp.ExperienceResult, error) {
	s, err := b.service(ctx, p, c.OwnerProject)
	if err != nil {
		return attentionapp.ExperienceResult{}, err
	}
	if _, err = b.service(ctx, p, c.TargetProject); err != nil {
		return attentionapp.ExperienceResult{}, err
	}
	return s.ShareExperience(ctx, p, id, c)
}
func (b *experienceHTTPBridge) DemoteExperience(ctx context.Context, p attentionapp.Principal, id string, c attentionapp.DemoteMemoryCommand) (attentionapp.ExperienceResult, error) {
	s, err := b.service(ctx, p, c.OwnerProject)
	if err != nil {
		return attentionapp.ExperienceResult{}, err
	}
	return s.DemoteExperience(ctx, p, id, c)
}
func (b *experienceHTTPBridge) GetExperience(ctx context.Context, p attentionapp.Principal, id, project string) (attentionapp.StoredExperience, error) {
	s, err := b.service(ctx, p, project)
	if err != nil {
		return attentionapp.StoredExperience{}, err
	}
	return s.GetExperience(ctx, p, id, project)
}
func (b *experienceHTTPBridge) GetExperiences(ctx context.Context, p attentionapp.Principal, project, pattern string) ([]attentionapp.StoredExperience, error) {
	s, err := b.service(ctx, p, project)
	if err != nil {
		return nil, err
	}
	return s.GetExperiences(ctx, p, project, pattern)
}
