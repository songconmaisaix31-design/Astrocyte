package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// MissionService provides mission queries.
type MissionService interface {
	ListMissions(ctx context.Context) (apierrors.ListResult, error)
	GetMission(ctx context.Context, id string) (any, error)
}

type emptyMissionService struct{}

// NewMissionService returns the S0 mission service (empty reads, 404 on get).
func NewMissionService() MissionService { return &emptyMissionService{} }

func (s *emptyMissionService) ListMissions(_ context.Context) (apierrors.ListResult, error) {
	return apierrors.EmptyList(), nil
}

func (s *emptyMissionService) GetMission(_ context.Context, id string) (any, error) {
	return nil, apierrors.NewNotFound("mission", id)
}
