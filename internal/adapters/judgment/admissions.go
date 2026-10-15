package judgmentadapter

import (
	"context"
	"reflect"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type AttentionOpportunities struct {
	Attention interface {
		GetOpportunity(context.Context, attention.Principal, string) (attention.OpportunityDetail, error)
		GetContent(context.Context, attention.Principal, string, int) (attention.ContentResult, error)
	}
}

func (a AttentionOpportunities) ReadJudgmentOpportunity(ctx context.Context, actor judgment.Actor, id string, revision int) (judgment.OpportunitySnapshot, error) {
	var snapshot judgment.OpportunitySnapshot
	if a.Attention == nil {
		return snapshot, apierrors.NewUnsupported("judgment_opportunity_source")
	}
	principal := attention.Principal{ID: actor.ID, Kind: actor.Kind}
	row, err := a.Attention.GetOpportunity(ctx, principal, id)
	if err != nil {
		return snapshot, err
	}
	opportunity := row.Opportunity
	if opportunity.ID != id || opportunity.Revision != revision {
		return snapshot, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment requires the current actual fixed opportunity revision"}
	}
	snapshot = judgment.OpportunitySnapshot{ID: opportunity.ID, Revision: opportunity.Revision, Version: opportunity.Version, State: opportunity.State, Title: opportunity.Title, NextStep: opportunity.NextStep, GoalRefs: slices.Clone(opportunity.GoalRefs), MissingEvidence: slices.Clone(opportunity.MissingEvidence), Evidence: []judgment.OpportunityEvidence{}}
	for _, ref := range opportunity.EvidenceRefs {
		content, err := a.Attention.GetContent(ctx, principal, ref.MaterialID, ref.Revision)
		if err != nil {
			return snapshot, err
		}
		snapshot.Evidence = append(snapshot.Evidence, judgment.OpportunityEvidence{Reference: judgment.EvidenceRef{MaterialID: ref.MaterialID, Revision: ref.Revision, Digest: judgment.DigestText(content.Text)}, Locator: ref.Locator, Span: ref.Span})
	}
	return snapshot, nil
}

type AdmissionAuthorizer struct{ Delegation *judgment.DelegationService }

var _ attention.AdmissionJudgmentAuthorizer = AdmissionAuthorizer{}

func (a AdmissionAuthorizer) AuthorizeAdmissionJudgment(ctx context.Context, request attention.AdmissionJudgmentRequest) (attention.AdmissionJudgmentProof, error) {
	var proof attention.AdmissionJudgmentProof
	if a.Delegation == nil {
		return proof, apierrors.NewUnsupported("delegated_admission_judgment")
	}
	record, rule, binding, err := a.Delegation.Authorize(ctx, request.WorkflowID, request.DecisionID, request.ProjectID, request.OpportunityID, request.OpportunityRevision)
	if err != nil {
		return proof, err
	}
	inputs := make([]attention.SourceRef, len(binding.Context.Opportunity.Evidence))
	for i, ref := range binding.Context.Opportunity.Evidence {
		inputs[i] = attention.SourceRef{MaterialID: ref.Reference.MaterialID, Revision: ref.Reference.Revision, Locator: ref.Locator, Span: ref.Span}
	}
	if !reflect.DeepEqual(inputs, request.Inputs) {
		return proof, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "actual admission input differs from the immutable judgment binding"}
	}
	model, version := "", ""
	if record.Result.Backend == judgment.RuleBackend {
		model, version = "rule:"+record.Request.Question.Rule, "evidence-rules/v1"
	} else {
		if record.Result.ProviderModel == nil || record.Result.ModelRevision == nil {
			return proof, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "actual judgment model identity is missing"}
		}
		model, version = *record.Result.ProviderModel, *record.Result.ModelRevision
	}
	refused, truncated := false, record.Result.Truncated
	proof = attention.AdmissionJudgmentProof{WorkflowID: rule.WorkflowID, DecisionID: record.Result.ID, OwnerID: rule.OwnerID, OpportunityID: binding.Context.Opportunity.ID, OpportunityRevision: binding.Context.Opportunity.Revision, ProjectID: binding.ProjectID, Inputs: inputs, Status: "succeeded", Decision: "admit", QuestionType: string(record.Result.QuestionType), Purpose: judgment.AdmissionPurpose, Backend: string(record.Result.Backend), Model: model, ModelVersion: version, PolicyVersion: record.Result.Versions.Policy, Refused: &refused, Truncated: &truncated, PolicyActive: true, CurrentPermissions: true, BudgetVerified: true}
	return proof, nil
}

type AdmissionWorkflow struct {
	Delegation *judgment.DelegationService
	Admissions interface {
		AdmitOpportunityWithJudgment(context.Context, string, string, string, attention.AdmissionCommand) (attention.AdmissionResult, error)
	}
}

func (a AdmissionWorkflow) AdmitJudgment(ctx context.Context, actor judgment.Actor, id string, command judgment.DelegatedAdmissionCommand) (judgment.DelegatedAdmissionReceipt, error) {
	var receipt judgment.DelegatedAdmissionReceipt
	if a.Delegation == nil || a.Admissions == nil {
		return receipt, apierrors.NewUnsupported("delegated_admission_workflow")
	}
	if command.SchemaVersion != 1 || command.RequestID == "" || command.IdempotencyKey == "" || command.ExpectedOpportunityVersion < 1 {
		return receipt, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "delegated admission requires explicit command identity and opportunity version"}
	}
	_, rule, binding, err := a.Delegation.AuthorizeForOwner(ctx, actor, id)
	if err != nil {
		return receipt, err
	}
	result, err := a.Admissions.AdmitOpportunityWithJudgment(ctx, rule.WorkflowID, id, binding.Context.Opportunity.ID, attention.AdmissionCommand{CommandMeta: attention.CommandMeta{SchemaVersion: command.SchemaVersion, RequestID: command.RequestID, IdempotencyKey: command.IdempotencyKey, ExpectedVersion: command.ExpectedOpportunityVersion}, ExpectedRevision: binding.Context.Opportunity.Revision, ProjectID: binding.ProjectID})
	if err != nil {
		return receipt, err
	}
	if result.Draft == nil || result.Draft.ID == "" {
		return receipt, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "actual delegated admission did not create its fixed proposal draft"}
	}
	return judgment.DelegatedAdmissionReceipt{AdmissionID: result.Admission.ID, AdmissionVersion: result.Admission.Version, ProposalID: result.Draft.ID, ProposalVersion: result.Draft.Version}, nil
}

type AdmissionProjects struct {
	Authorization *ProjectAuthorization
	Rules         judgment.DelegationRepository
}

func (a AdmissionProjects) AuthorizeAdmissionProject(ctx context.Context, principal attention.Principal, projectID string, refs []attention.SourceRef) error {
	if a.Authorization == nil || a.Rules == nil {
		return apierrors.NewUnsupported("admission_current_project_access")
	}
	actor := judgment.Actor{ID: principal.ID, Kind: principal.Kind, ProjectID: projectID}
	if principal.Kind == "delegated_judgment" {
		rules, err := a.Rules.ListDelegationRules(ctx, projectID)
		if err != nil {
			return err
		}
		found := false
		for _, rule := range rules {
			if rule.WorkflowID == principal.ID && rule.State == "active" {
				actor.ID, actor.Kind, found = rule.OwnerID, "human", true
			}
		}
		if !found {
			return &apierrors.ServiceError{Code: apierrors.ApprovalRevoked, Message: "the actual project delegation workflow is absent or inactive"}
		}
	}
	project, err := a.Authorization.project(ctx, actor, projectID)
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if _, err := a.Authorization.References.ReadSelected(ctx, domain.Caller{ID: actor.ID, Kind: actor.Kind, ProjectID: projectID}, project.SpaceID, domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}, false); err != nil {
			return err
		}
	}
	return nil
}
