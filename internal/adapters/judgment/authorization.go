package judgmentadapter

import (
	"context"
	"slices"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type ProjectAuthorization struct {
	Projects   workspace.LocalProjectRepository
	References workspace.ProjectReferences
	Policies   judgment.PolicyRepository
}

func authorizationDenied(message string) error {
	return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: message}
}

func (a *ProjectAuthorization) project(ctx context.Context, actor judgment.Actor, id string) (domain.LocalProject, error) {
	var project domain.LocalProject
	if a.Projects == nil || a.References == nil || a.Policies == nil {
		return project, apierrors.NewUnsupported("judgment_authorization")
	}
	if actor.ID == "" || (actor.Kind != "human" && actor.Kind != "agent") {
		return project, authorizationDenied("judgment requires a trusted human or project Agent")
	}
	project, err := a.Projects.LoadProject(ctx, id)
	if err != nil {
		return project, err
	}
	if actor.Kind == "agent" {
		if actor.ProjectID != id {
			return project, authorizationDenied("judgment Agent belongs to a different project")
		}
		grant, err := a.Projects.LoadGrant(ctx, id, actor.ID)
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
	input.Evidence = []judgment.Evidence{}
	caller := domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: actor.ProjectID}
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
	if actor.Kind != "human" || actor.ID == "" {
		return authorizationDenied("only an authenticated human can change judgment policy")
	}
	project, err := a.project(ctx, actor, policy.ProjectID)
	if err != nil {
		return err
	}
	if project.Settings.Revision != policy.ProjectSettingsRevision {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "policy project settings changed"}
	}
	// 外发许可逐引用检查当前范围，配置模型名称不会扩大资料权限。
	for _, ref := range policy.ExternalEvidence {
		if _, err := a.References.ReadSelected(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind}, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false); err != nil {
			return err
		}
	}
	return a.Policies.SaveJudgmentPolicy(ctx, policy, expected)
}
