package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// MaterialService provides material queries.
type MaterialService interface {
	ListMaterials(ctx context.Context) (apierrors.ListResult, error)
}

// OpportunityService provides opportunity queries.
type OpportunityService interface {
	ListOpportunities(ctx context.Context) (apierrors.ListResult, error)
}

type emptyMaterialService struct{}

// NewMaterialService returns the S0 material service (empty reads).
func NewMaterialService() MaterialService { return &emptyMaterialService{} }

func (s *emptyMaterialService) ListMaterials(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}

type emptyOpportunityService struct{}

// NewOpportunityService returns the S0 opportunity service (empty reads).
func NewOpportunityService() OpportunityService { return &emptyOpportunityService{} }

func (s *emptyOpportunityService) ListOpportunities(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}
