package judgmentadapter

import (
	"context"
	"reflect"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type GoalReader interface {
	ReadGoal(context.Context, domain.Caller, string, int) (domain.Goal, error)
}

type RankingReader interface {
	GetRankingProfile(context.Context, attention.Principal) (attention.RankingProfileDetail, error)
}

func (a *ProjectAuthorization) ReadGoal(ctx context.Context, actor judgment.Actor, projectID string, ref domain.GoalReference) (domain.Goal, error) {
	var goal domain.Goal
	if ref.ID == "" || ref.Revision < 1 {
		return goal, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "judgment requires an explicitly selected fixed Goal reference"}
	}
	if _, err := a.project(ctx, actor, projectID); err != nil {
		return goal, err
	}
	if a.Goals == nil {
		return goal, apierrors.NewUnsupported("judgment_goal_source")
	}
	caller := domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}
	fixed, err := a.Goals.ReadGoal(ctx, caller, ref.ID, ref.Revision)
	if err != nil {
		return goal, err
	}
	current, err := a.Goals.ReadGoal(ctx, caller, ref.ID, 0)
	if err != nil {
		return goal, err
	}
	if fixed.ProjectID != projectID || current.ProjectID != projectID {
		return goal, authorizationDenied("judgment Goal belongs to a different project")
	}
	if fixed.ID != ref.ID || fixed.Revision != ref.Revision || !reflect.DeepEqual(fixed, current) || current.Status != "active" {
		return goal, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment Goal is superseded or no longer active"}
	}
	return fixed, nil
}

func (a *ProjectAuthorization) ReadRanking(ctx context.Context, actor judgment.Actor, projectID, id string, version int) (attention.RankingProfile, error) {
	var profile attention.RankingProfile
	if id == "" || version < 1 {
		return profile, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "judgment requires an explicitly selected fixed ranking profile"}
	}
	if _, err := a.project(ctx, actor, projectID); err != nil {
		return profile, err
	}
	if a.Ranking == nil {
		return profile, apierrors.NewUnsupported("judgment_ranking_source")
	}
	detail, err := a.Ranking.GetRankingProfile(ctx, attention.Principal{ID: actor.ID, Kind: actor.Kind})
	if err != nil {
		return profile, err
	}
	if !detail.Configured || detail.Profile == nil || !detail.Profile.Enabled {
		return profile, &apierrors.ServiceError{Code: apierrors.ApprovalRequired, Message: "judgment ranking profile is unconfigured or disabled"}
	}
	if detail.Profile.ID != id || detail.Profile.Version != version {
		return profile, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment ranking profile version changed"}
	}
	return *detail.Profile, nil
}
