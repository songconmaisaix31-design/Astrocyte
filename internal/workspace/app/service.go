package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// ProjectService provides project queries.
type ProjectService interface {
	ListProjects(ctx context.Context) (apierrors.ListResult, error)
}

// ProposalService provides proposal queries.
type ProposalService interface {
	ListProposals(ctx context.Context) (apierrors.ListResult, error)
}

// SessionService provides session queries.
type SessionService interface {
	ListSessions(ctx context.Context) (apierrors.ListResult, error)
}

type emptyProjectService struct{}

// NewProjectService returns the S0 project service (empty reads).
func NewProjectService() ProjectService { return &emptyProjectService{} }

func (s *emptyProjectService) ListProjects(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}

type emptyProposalService struct{}

// NewProposalService returns the S0 proposal service (empty reads).
func NewProposalService() ProposalService { return &emptyProposalService{} }

func (s *emptyProposalService) ListProposals(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}

type emptySessionService struct{}

// NewSessionService returns the S0 session service (empty reads).
func NewSessionService() SessionService { return &emptySessionService{} }

func (s *emptySessionService) ListSessions(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}
