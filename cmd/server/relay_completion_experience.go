package main

import (
	"context"
	"errors"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayCompletionProjectAccess struct {
	projects workspace.LocalProjects
}

var _ attention.ExperienceCompletionReader = (*relaySessionRuntime)(nil)

func (s *relaySessionRuntime) ReadExperienceCompletion(ctx context.Context, principal attention.Principal, projectID, completionID string) (attention.ExperienceCompletionEvidence, error) {
	if s == nil || s.completionSource == nil {
		return attention.ExperienceCompletionEvidence{}, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "actual Relay completion source is unavailable"}
	}
	return s.completionSource.ReadExperienceCompletion(ctx, principal, projectID, completionID)
}

func (s *relaySessionRuntime) AuthorizeExperienceCompletion(ctx context.Context, principal attention.Principal, evidence attention.ExperienceCompletionEvidence) error {
	if s == nil || s.completionSource == nil {
		return &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "actual Relay completion source is unavailable"}
	}
	return s.completionSource.AuthorizeExperienceCompletion(ctx, principal, evidence)
}

// 复用当前主数据库的项目读取和事务 context；Agent action 尚未授权。
func (a relayCompletionProjectAccess) currentOwner(ctx context.Context, principal attention.Principal, projectID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	denied := &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "completion requires the current authenticated human and registered project"}
	if principal.Kind != "human" || strings.TrimSpace(principal.ID) == "" || projectID == "" || strings.TrimSpace(projectID) != projectID {
		return "", denied
	}
	if a.projects == nil {
		return "", &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "current completion project authorization is unavailable"}
	}
	projects, err := a.projects.ListProjects(ctx, domain.Caller{ID: principal.ID, Kind: principal.Kind})
	if err != nil {
		return "", err
	}
	for _, project := range projects {
		if project.ID == projectID {
			return project.ID, nil
		}
	}
	return "", denied
}

func assembleRelayCompletionSource(services *httpapi.Services) error {
	if services == nil || services.Relay == nil || services.LocalProjects == nil {
		return errors.New("completion source requires the original Relay and project services")
	}
	runtime, ok := services.Relay.(*relaySessionRuntime)
	if !ok {
		return errors.New("completion source requires the original production Relay instance")
	}
	reader, ok := runtime.RelayRuntime.(workspace.RelayCompletionEvidenceReader)
	if !ok {
		return errors.New("original Relay completion reader is unavailable")
	}
	source, err := agents.NewRelayCompletionExperienceSource(reader, relayCompletionProjectAccess{projects: services.LocalProjects}.currentOwner)
	if err != nil {
		return err
	}
	runtime.completionSource = source
	return nil
}
