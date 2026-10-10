package app

import (
	"context"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type projectScopedDistiller interface {
	Distiller
	ProjectSpaceID(context.Context) (string, error)
}

func (s *Service) resolveDistiller(ctx context.Context, p Principal, c RequestDistillationCommand) (Distiller, string, error) {
	if c.ProjectID == "" && c.CLI == "" {
		return s.options.Distiller, "", nil
	}
	if c.ProjectID == "" || c.CLI == "" || s.options.ProjectDistillers == nil {
		return nil, "", serviceError(apierrors.ScopeDenied, "Choose a project and its permitted configured CLI", "select_permitted_project_cli")
	}
	processor, err := s.options.ProjectDistillers.Resolve(ctx, p, c.ProjectID, c.CLI)
	if err != nil {
		return nil, "", err
	}
	scoped, ok := processor.(projectScopedDistiller)
	if !ok {
		return nil, "", serviceError(apierrors.ScopeDenied, "Selected project processor cannot verify fixed reference scope", "configure_project_reference_scope")
	}
	spaceID, err := scoped.ProjectSpaceID(ctx)
	return processor, spaceID, err
}
