package judgmentadapter

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type ProjectAuthorization struct {
	Projects   workspace.RepositoryProjectReader
	References workspace.ProjectReferences
	Policies   judgment.PolicyRepository
	Snapshots  workspace.ProjectSnapshotProvider
	Goals      GoalReader
	Ranking    RankingReader
}

func authorizationDenied(message string) error {
	return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: message}
}

func (a *ProjectAuthorization) project(ctx context.Context, actor judgment.Actor, id string) (domain.LocalProject, error) {
	var project domain.LocalProject
	if a.Projects == nil || a.References == nil || a.Policies == nil || a.Snapshots == nil {
		return project, apierrors.NewUnsupported("judgment_authorization")
	}
	if actor.ID == "" || (actor.Kind != "human" && actor.Kind != "agent") {
		return project, authorizationDenied("judgment requires a trusted human or project Agent")
	}
	project, err := a.Projects.LoadRepositoryProject(ctx, id)
	if err != nil {
		return project, err
	}
	if actor.Kind == "agent" {
		if actor.ProjectID != id {
			return project, authorizationDenied("judgment Agent belongs to a different project")
		}
		grant, err := a.Projects.LoadRepositoryGrant(ctx, id, actor.ID)
		if err != nil {
			return project, err
		}
		if grant.RevokedAt != nil || (grant.ExpiresAt != nil && !grant.ExpiresAt.After(time.Now())) || !slices.Contains(grant.Actions, "read_context") {
			return project, authorizationDenied("judgment project Agent read permission is revoked, expired or absent")
		}
	}
	return project, nil
}

func (a *ProjectAuthorization) Resolve(ctx context.Context, actor judgment.Actor, request judgment.JudgmentRequest) (judgment.AuthorizedInput, error) {
	var input judgment.AuthorizedInput
	project, err := a.project(ctx, actor, request.ProjectID)
	if err != nil {
		return input, err
	}
	policy, err := a.Policies.LoadJudgmentPolicy(ctx, project.ID, request.Question.Type)
	if err != nil {
		return input, err
	}
	if project.Settings.Revision != policy.ProjectSettingsRevision {
		return input, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment policy targets old project settings"}
	}
	input.Policy, input.WorktreeID, input.SnapshotID = policy, policy.WorktreeID, policy.SnapshotID
	caller := domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}
	if err := a.Snapshots.ValidateProjectSnapshot(ctx, caller, policy.Snapshot); err != nil {
		return input, err
	}
	if actor.Kind == "agent" {
		grant, err := a.Projects.LoadRepositoryGrant(ctx, project.ID, actor.ID)
		if err != nil {
			return input, err
		}
		data, err := json.Marshal(grant)
		if err != nil {
			return input, err
		}
		input.AccessVersion = judgment.DigestText(string(data))
	} else {
		input.AccessVersion = policy.Versions.AccessEpoch
	}
	input.Evidence = []judgment.Evidence{}
	for _, ref := range request.EvidenceRefs {
		material, err := a.References.ReadSelected(ctx, caller, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false)
		if err != nil {
			return input, err
		}
		if material.Reference.MaterialID != ref.MaterialID || material.Reference.Revision != ref.Revision || judgment.DigestText(material.Text) != ref.Digest {
			return input, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment evidence fixed revision or actual digest differs"}
		}
		if request.Backend == judgment.JevBackend && !slices.Contains(policy.ExternalEvidence, ref) {
			return input, authorizationDenied("fixed material revision has no explicit human external-data permission")
		}
		input.Evidence = append(input.Evidence, judgment.Evidence{Reference: ref, Text: material.Text})
	}
	return input, nil
}

func (a *ProjectAuthorization) SetPolicy(ctx context.Context, actor judgment.Actor, policy judgment.Policy, expected int) error {
	if a.Policies == nil {
		return apierrors.NewUnsupported("judgment_policy_persistence")
	}
	return a.Policies.WithJudgmentTx(ctx, func(ctx context.Context) error {
		return a.setPolicy(ctx, actor, policy, expected)
	})
}

func (a *ProjectAuthorization) setPolicy(ctx context.Context, actor judgment.Actor, policy judgment.Policy, expected int) error {
	if actor.Kind != "human" || actor.ID == "" {
		return authorizationDenied("only an authenticated human can change judgment policy")
	}
	project, err := a.project(ctx, actor, policy.ProjectID)
	if err != nil {
		return err
	}
	if policy.Snapshot.Root == "" {
		policy.Snapshot.Root = project.Root
	}
	if err := judgment.ValidatePolicy(policy); err != nil {
		return err
	}
	if project.Settings.Revision != policy.ProjectSettingsRevision {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "policy project settings changed"}
	}
	if err := a.Snapshots.ValidateProjectSnapshot(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind}, policy.Snapshot); err != nil {
		return err
	}
	// 外发许可逐引用检查当前范围，配置模型名称不会扩大资料权限。
	for _, ref := range policy.ExternalEvidence {
		material, err := a.References.ReadSelected(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind}, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false)
		if err != nil {
			return err
		}
		if judgment.DigestText(material.Text) != ref.Digest {
			return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "external permission requires the actual fixed evidence digest"}
		}
	}
	return a.Policies.SaveJudgmentPolicy(ctx, policy, expected)
}

func (a *ProjectAuthorization) Snapshot(ctx context.Context, actor judgment.Actor, projectID string, files []string) (domain.ProjectContextSnapshot, error) {
	project, err := a.project(ctx, actor, projectID)
	if err != nil {
		return domain.ProjectContextSnapshot{}, err
	}
	return a.Snapshots.CaptureProjectSnapshot(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}, project.ID, files)
}

func (a *ProjectAuthorization) AuthorizeRecord(ctx context.Context, actor judgment.Actor, request judgment.JudgmentRequest) error {
	project, err := a.project(ctx, actor, request.ProjectID)
	if err != nil {
		return err
	}
	for _, ref := range request.EvidenceRefs {
		material, err := a.References.ReadSelected(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false)
		if err != nil {
			return err
		}
		if judgment.DigestText(material.Text) != ref.Digest {
			return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "historical judgment evidence digest differs"}
		}
	}
	return nil
}

func (a *ProjectAuthorization) ReadPolicy(ctx context.Context, actor judgment.Actor, projectID string, questionType judgment.QuestionType) (judgment.Policy, error) {
	if _, err := a.project(ctx, actor, projectID); err != nil {
		return judgment.Policy{}, err
	}
	return a.Policies.LoadJudgmentPolicy(ctx, projectID, questionType)
}
