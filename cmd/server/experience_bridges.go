package main

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// HTTP 使用已注册的可信项目；MCP 使用独立的启动绑定。
type experienceHTTPBridge struct {
	db         *sqlite.DB
	events     attentionapp.EventReader
	projects   workspaceapp.LocalProjects
	registered workspaceapp.RegisteredProjects

	completionSources     attentionapp.ExperienceCompletionReader
	completionInterpreter attentionapp.CompletionExperienceInterpreter
}

var _ attentionapp.ExperienceGovernanceService = (*experienceHTTPBridge)(nil)
var _ attentionapp.CompletionExperienceGovernance = (*experienceHTTPBridge)(nil)

// 完成经验的 owner 始终使用 LocalProject.ID，不接收目录或发现页身份替代。
func (b *experienceHTTPBridge) completionService(ctx context.Context, p attentionapp.Principal, projectID string) (*attentionapp.AcceptedExperienceService, error) {
	denied := &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Completion experience requires a current human and actual registered local project identity", RequiredAction: "select_registered_project"}
	if p.Kind != "human" || p.ID == "" || projectID == "" || b.projects == nil {
		return nil, denied
	}
	projects, err := b.projects.ListProjects(ctx, workspacedomain.Caller{Kind: p.Kind, ID: p.ID})
	if err != nil {
		return nil, err
	}
	known := false
	for _, project := range projects {
		if project.ID == projectID {
			known = true
			break
		}
	}
	if !known {
		return nil, denied
	}
	if b.db == nil {
		return nil, apierrors.NewUnsupported("actual_completion_experience_source")
	}
	repository := sqlite.NewAcceptedExperienceRepository(b.db)
	governance := attentionapp.NewExperienceService(repository, b.events, attentionapp.ExperienceOptions{ProjectID: projectID})
	return attentionapp.NewAcceptedExperienceService(governance, repository, b.completionSources, b.completionInterpreter), nil
}

func (b *experienceHTTPBridge) CaptureCompletionExperience(ctx context.Context, p attentionapp.Principal, command attentionapp.CaptureCompletionExperienceCommand) (attentionapp.StoredExperience, error) {
	service, err := b.completionService(ctx, p, command.LocalProjectID)
	if err != nil {
		return attentionapp.StoredExperience{}, err
	}
	return service.CaptureCompletionExperience(ctx, p, command)
}

func (b *experienceHTTPBridge) ResolveExperienceEvidence(ctx context.Context, p attentionapp.Principal, id, projectID string) ([]attentionapp.SafeExperienceEvidence, error) {
	service, err := b.completionService(ctx, p, projectID)
	if err != nil {
		return nil, err
	}
	return service.ResolveEvidence(ctx, p, id, projectID)
}

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
	if b.db == nil {
		return nil, apierrors.NewUnsupported("experience_persistence")
	}
	return attentionapp.NewExperienceService(sqlite.NewAcceptedExperienceRepository(b.db), b.events, attentionapp.ExperienceOptions{ProjectID: project}), nil
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
