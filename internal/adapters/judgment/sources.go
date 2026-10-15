package judgmentadapter

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type GoalReader interface {
	workspace.CurrentGoalReader
	ReadGoal(context.Context, domain.Caller, string, int) (domain.Goal, error)
}

type RankingReader interface {
	GetRankingProfile(context.Context, attention.Principal) (attention.RankingProfileDetail, error)
}

type TasteReader interface {
	ReadTasteProfile(context.Context, attention.Principal, string, int) (attention.TasteProfileSnapshot, error)
}

func (a *ProjectAuthorization) ReadGoal(ctx context.Context, actor judgment.Actor, projectID string, ref domain.GoalReference) (domain.Goal, error) {
	var goal domain.Goal
	if ref.ID == "" || ref.Revision < 1 {
		return goal, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "judgment requires an explicitly selected fixed Goal reference"}
	}
	project, err := a.project(ctx, actor, projectID)
	if err != nil {
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
	current, err := a.Goals.ReadCurrentGoalSnapshot(ctx, caller, projectID, ref.ID, ref.Revision)
	if err != nil {
		return goal, err
	}
	if fixed.ProjectID != projectID || current.Goal.ProjectID != projectID {
		return goal, authorizationDenied("judgment Goal belongs to a different project")
	}
	if fixed.ID != ref.ID || fixed.Revision != ref.Revision || !reflect.DeepEqual(fixed, current.Goal) || current.Goal.Status != "active" || current.ProjectSettingsRevision != project.Settings.Revision {
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

func (a *ProjectAuthorization) ReadDelegationSource(ctx context.Context, actor judgment.Actor, projectID, ruleID, workflowID string, revision int) (judgment.DelegationRule, error) {
	var rule judgment.DelegationRule
	if ruleID == "" || workflowID == "" || revision < 1 {
		return rule, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "taste delegation source requires a fixed rule, workflow and revision"}
	}
	if _, err := a.project(ctx, actor, projectID); err != nil {
		return rule, err
	}
	if a.Rules == nil {
		return rule, apierrors.NewUnsupported("judgment_delegation_source")
	}
	rule, err := a.Rules.LoadDelegationRule(ctx, ruleID)
	if err != nil {
		return rule, err
	}
	if actor.Kind != "human" || rule.OwnerID != actor.ID || rule.ProjectID != projectID || rule.WorkflowID != workflowID {
		return rule, authorizationDenied("taste delegation source belongs to another project, workflow or human owner")
	}
	if rule.Revision != revision {
		return rule, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "taste delegation source revision changed"}
	}
	if err := judgment.ValidateDelegationRule(rule); err != nil {
		return rule, err
	}
	return rule, nil
}

func (a *ProjectAuthorization) readSources(ctx context.Context, actor judgment.Actor, policy judgment.Policy) (*judgment.SourceSnapshot, error) {
	if policy.Sources == nil {
		return nil, nil
	}
	refs := *policy.Sources
	if err := judgment.ValidateSourceReferences(refs); err != nil {
		return nil, err
	}
	if policy.Backend == judgment.JevBackend && !reflect.DeepEqual(policy.Sources, policy.ExternalSources) {
		return nil, authorizationDenied("the selected fixed Goal and complete taste sources have no explicit external-data permission")
	}
	goal, err := a.ReadGoal(ctx, actor, policy.ProjectID, refs.Goal)
	if err != nil {
		return nil, err
	}
	if a.Taste == nil {
		return nil, apierrors.NewUnsupported("judgment_taste_source")
	}
	row, err := a.Taste.ReadTasteProfile(ctx, attention.Principal{ID: actor.ID, Kind: actor.Kind}, refs.Taste.ID, refs.Taste.Version)
	if err != nil {
		return nil, err
	}
	if !row.Configured || row.Fixed == nil || row.Current == nil || !row.Current.Enabled {
		return nil, &apierrors.ServiceError{Code: apierrors.ApprovalRequired, Message: "complete taste profile is unconfigured or disabled"}
	}
	if row.Fixed.ID != refs.Taste.ID || row.Fixed.Version != refs.Taste.Version || !reflect.DeepEqual(row.Fixed, row.Current) {
		return nil, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "complete taste profile revision changed"}
	}
	if err := attention.ValidateTasteProfile(*row.Fixed); err != nil {
		return nil, err
	}
	for _, rubric := range row.Fixed.Rubrics {
		if !slices.Contains([]judgment.QuestionType{judgment.Choice, judgment.Score, judgment.Truth}, judgment.QuestionType(rubric.QuestionType)) {
			return nil, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "complete taste rubric uses an unsupported judgment question type"}
		}
	}
	profile, err := json.Marshal(row.Fixed)
	if err != nil {
		return nil, err
	}
	snapshot := &judgment.SourceSnapshot{References: refs, Goal: goal, Taste: profile, Delegations: []judgment.DelegationRule{}, Examples: []judgment.Evidence{}}
	if refs.Ranking != nil {
		actual, err := a.ReadRanking(ctx, actor, policy.ProjectID, refs.Ranking.ID, refs.Ranking.Version)
		if err != nil {
			return nil, err
		}
		snapshot.Ranking, err = json.Marshal(actual)
		if err != nil {
			return nil, err
		}
	}
	for _, ref := range row.Fixed.Delegations {
		rule, err := a.ReadDelegationSource(ctx, actor, policy.ProjectID, ref.RuleID, ref.WorkflowID, ref.Revision)
		if err != nil {
			return nil, err
		}
		snapshot.Delegations = append(snapshot.Delegations, rule)
	}
	project, err := a.project(ctx, actor, policy.ProjectID)
	if err != nil {
		return nil, err
	}
	for _, example := range row.Fixed.Examples {
		ref := judgment.EvidenceRef{MaterialID: example.MaterialID, Revision: example.Revision, Digest: example.ContentDigest}
		if policy.Backend == judgment.JevBackend && !slices.Contains(policy.ExternalEvidence, ref) {
			return nil, authorizationDenied("a fixed taste example has no explicit external-data permission")
		}
		material, err := a.References.ReadSelected(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false)
		if err != nil {
			return nil, err
		}
		if material.Reference.MaterialID != ref.MaterialID || material.Reference.Revision != ref.Revision || judgment.DigestText(material.Text) != ref.Digest {
			return nil, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "fixed taste example text or actual digest differs"}
		}
		snapshot.Examples = append(snapshot.Examples, judgment.Evidence{Reference: ref, Text: material.Text})
	}
	return snapshot, nil
}
