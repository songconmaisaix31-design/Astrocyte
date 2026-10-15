package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"reflect"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type AdmissionService struct {
	core    *Service
	repo    AdmissionRepository
	options AdmissionOptions
}

var _ OpportunityAdmissions = (*AdmissionService)(nil)

func NewAdmissionService(core *Service, repository AdmissionRepository, options AdmissionOptions) *AdmissionService {
	return &AdmissionService{core: core, repo: repository, options: options}
}

func (s *AdmissionService) transact(ctx context.Context, p Principal, meta CommandMeta, name string, input any, validate func(AdmissionTx) error, mutate func(AdmissionTx) (AdmissionResult, error)) (AdmissionResult, error) {
	var result AdmissionResult
	if s.core == nil || s.repo == nil {
		return result, serviceError(apierrors.UnsupportedCapability, "Admission persistence is not connected", "connect_admission_repository")
	}
	if s.core.configurationError != nil {
		return result, mapError(s.core.configurationError, meta.RequestID)
	}
	if err := validateMeta(meta); err != nil {
		return result, mapError(err, meta.RequestID)
	}
	digest, err := semanticDigest(input)
	if err != nil {
		return result, err
	}
	err = s.repo.WithAdmissionTx(ctx, func(tx AdmissionTx) error {
		if err := validate(tx); err != nil {
			return err
		}
		caller := p.Kind + ":" + p.ID
		old, err := tx.LoadReceipt(caller, name, meta.IdempotencyKey)
		if err == nil {
			if old.Digest != digest {
				return serviceError(apierrors.VersionConflict, "Idempotency key binds different admission input", "use_new_idempotency_key")
			}
			return json.Unmarshal(old.Result, &result)
		}
		if !isMissing(err) {
			return err
		}
		result, err = mutate(tx)
		if err != nil {
			return err
		}
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return tx.SaveReceipt(caller, name, meta.IdempotencyKey, Receipt{Digest: digest, Result: data})
	})
	return result, mapError(err, meta.RequestID)
}

func admissionFixedOpportunity(row OpportunityDetail, revision int) (Opportunity, error) {
	for _, candidate := range row.Revisions {
		if candidate.Revision == revision {
			return candidate, nil
		}
	}
	return Opportunity{}, apierrors.NewNotFound("opportunity revision", row.Opportunity.ID)
}

func (s *AdmissionService) AdmitOpportunity(ctx context.Context, p Principal, id string, command AdmissionCommand) (AdmissionResult, error) {
	if err := authorize(p, true); err != nil {
		return AdmissionResult{}, err
	}
	return s.admit(ctx, p, id, command, "")
}

// 仅供真实委托判断流程调用；普通 Agent HTTP 凭据不进入该入口。
func (s *AdmissionService) AdmitOpportunityWithJudgment(ctx context.Context, workflowID, decisionID, id string, command AdmissionCommand) (AdmissionResult, error) {
	if workflowID == "" || decisionID == "" {
		return AdmissionResult{}, serviceError(apierrors.ApprovalRequired, "A persisted delegated judgment is required", "provide_authorized_judgment")
	}
	return s.admit(ctx, Principal{Kind: "delegated_judgment", ID: workflowID}, id, command, decisionID)
}

func (s *AdmissionService) judgment(ctx context.Context, p Principal, decisionID, projectID string, candidate Opportunity) (AdmissionJudgmentProof, error) {
	if s.options.Judgment == nil {
		return AdmissionJudgmentProof{}, serviceError(apierrors.ApprovalRequired, "Delegated judgment authorization is not connected", "configure_current_question_delegation")
	}
	proof, err := s.options.Judgment.AuthorizeAdmissionJudgment(ctx, AdmissionJudgmentRequest{WorkflowID: p.ID, DecisionID: decisionID, OpportunityID: candidate.ID, OpportunityRevision: candidate.Revision, ProjectID: projectID, Inputs: slices.Clone(candidate.EvidenceRefs)})
	if err != nil {
		return proof, err
	}
	if proof.WorkflowID != p.ID || proof.DecisionID != decisionID || proof.OwnerID == "" || proof.OpportunityID != candidate.ID || proof.OpportunityRevision != candidate.Revision || proof.ProjectID != projectID || !reflect.DeepEqual(normalizeRefs(proof.Inputs), normalizeRefs(candidate.EvidenceRefs)) {
		return proof, serviceError(apierrors.ScopeDenied, "Judgment does not bind this fixed opportunity input", "review_judgment_scope")
	}
	if proof.Status != "succeeded" || proof.Decision != "admit" || !slices.Contains([]string{"choice", "score", "truth"}, proof.QuestionType) || proof.Purpose != "opportunity_admission" || proof.Refused == nil || *proof.Refused || proof.Truncated == nil || *proof.Truncated || !proof.PolicyActive || !proof.CurrentPermissions || !proof.BudgetVerified {
		return proof, serviceError(apierrors.EvidenceMissing, "Judgment lacks a current authorized complete admission decision", "review_judgment_evidence")
	}
	return proof, nil
}

func (s *AdmissionService) admit(ctx context.Context, p Principal, id string, command AdmissionCommand, decisionID string) (AdmissionResult, error) {
	if command.ExpectedRevision < 1 || strings.TrimSpace(id) == "" {
		return AdmissionResult{}, mapError(domain.ErrInvalid, command.RequestID)
	}
	if s.options.Drafts == nil {
		return AdmissionResult{}, serviceError(apierrors.UnsupportedCapability, "Admission draft consumer is not connected", "connect_workspace_admission_consumer")
	}
	var row OpportunityDetail
	var candidate Opportunity
	var proof AdmissionJudgmentProof
	input := struct {
		ID, DecisionID string
		AdmissionCommand
	}{id, decisionID, command}
	return s.transact(ctx, p, command.CommandMeta, "AdmitOpportunity", input, func(tx AdmissionTx) error {
		var err error
		row, err = tx.LoadOpportunity(id)
		if err != nil {
			return err
		}
		candidate, err = admissionFixedOpportunity(row, command.ExpectedRevision)
		if err != nil {
			return err
		}
		if err := validateOpportunityAccess(tx, candidate); err != nil {
			return err
		}
		if p.Kind == "delegated_judgment" {
			proof, err = s.judgment(tx.Context(), p, decisionID, command.ProjectID, candidate)
			if err != nil {
				return err
			}
		}
		if command.ProjectID != "" {
			if s.options.Projects == nil {
				return serviceError(apierrors.UnsupportedCapability, "Current project authorization is not connected", "connect_project_permission_reader")
			}
			return s.options.Projects.AuthorizeAdmissionProject(tx.Context(), p, command.ProjectID, candidate.EvidenceRefs)
		}
		return nil
	}, func(tx AdmissionTx) (AdmissionResult, error) {
		if row.Opportunity.Version != command.ExpectedVersion || row.Opportunity.Revision != command.ExpectedRevision {
			return AdmissionResult{}, domain.ErrVersion
		}
		if _, found, err := tx.FindAdmission(id, command.ExpectedRevision); err != nil {
			return AdmissionResult{}, err
		} else if found {
			return AdmissionResult{}, domain.ErrVersion
		}
		value := domain.Admission{ID: rand.Text(), AdmittedEventID: rand.Text(), OpportunityID: id, OpportunityRevision: candidate.Revision, OpportunityVersion: row.Opportunity.Version, ProjectID: command.ProjectID, Inputs: domainRefs(candidate.EvidenceRefs), OwnerID: p.ID, OwnerNamespace: "human:" + p.ID, ActorID: p.ID, Authority: domain.AdmissionHuman}
		if p.Kind == "delegated_judgment" {
			value.Authority, value.OwnerID, value.OwnerNamespace = domain.AdmissionDelegatedJudgment, proof.OwnerID, "human:"+proof.OwnerID
			value.DecisionID, value.PolicyVersion, value.Model, value.ModelVersion, value.Backend, value.QuestionType = proof.DecisionID, proof.PolicyVersion, proof.Model, proof.ModelVersion, proof.Backend, proof.QuestionType
			value.Purpose = proof.Purpose
		}
		admission, err := domain.NewAdmission(domain.AdmissionInput{Admission: value, OpportunityState: row.Opportunity.State}, s.core.options.Clock())
		if err != nil {
			return AdmissionResult{}, err
		}
		oldVersion := row.Opportunity.Version
		row.Opportunity.State, row.Opportunity.Version = "admitted", oldVersion+1
		if err := tx.SaveOpportunity(row, oldVersion); err != nil {
			return AdmissionResult{}, err
		}
		if err := tx.SaveAdmission(admission, 0); err != nil {
			return AdmissionResult{}, err
		}
		if err := s.appendEvent(tx, admission.AdmittedEventID, "opportunity_admitted", admission, command.CommandMeta); err != nil {
			return AdmissionResult{}, err
		}
		draft, err := s.options.Drafts.ConsumeAdmittedOpportunity(tx.Context(), admission.AdmittedEventID)
		if err != nil {
			return AdmissionResult{}, err
		}
		if draft.ID == "" || draft.Version < 1 {
			return AdmissionResult{}, serviceError(apierrors.EvidenceMissing, "Workspace did not return a persisted draft identity", "inspect_workspace_admission_receipt")
		}
		return AdmissionResult{SchemaVersion: 1, Admission: admission, Draft: &draft}, nil
	})
}

func (s *AdmissionService) appendEvent(tx AdmissionTx, eventID, kind string, admission domain.Admission, meta CommandMeta) error {
	payload, err := json.Marshal(map[string]any{"admission_id": admission.ID, "opportunity_id": admission.OpportunityID, "opportunity_revision": admission.OpportunityRevision, "owner_namespace": admission.OwnerNamespace, "authority": admission.Authority, "policy_version": admission.PolicyVersion})
	if err != nil {
		return err
	}
	return tx.AppendEvent(OutboxEvent{ID: eventID, Type: "attention." + kind, AggregateID: admission.ID, AggregateVersion: admission.Version, OccurredAt: s.core.options.Clock(), CorrelationID: meta.RequestID, CausationID: meta.IdempotencyKey, Payload: payload})
}

func (s *AdmissionService) RevokeAdmission(ctx context.Context, p Principal, id string, command RevokeAdmissionCommand) (AdmissionResult, error) {
	if err := authorize(p, true); err != nil {
		return AdmissionResult{}, err
	}
	var admission domain.Admission
	input := struct {
		ID      string
		RevokeAdmissionCommand
	}{id, command}
	return s.transact(ctx, p, command.CommandMeta, "RevokeAdmission", input, func(tx AdmissionTx) error {
		var err error
		admission, err = tx.LoadAdmission(id)
		if err != nil {
			return err
		}
		if admission.OwnerID != p.ID {
			return serviceError(apierrors.ScopeDenied, "Admission belongs to another owner namespace", "use_owned_admission")
		}
		return nil
	}, func(tx AdmissionTx) (AdmissionResult, error) {
		oldVersion := admission.Version
		if err := admission.Revoke(command.ExpectedVersion, p.ID, rand.Text(), command.Reason, s.core.options.Clock()); err != nil {
			return AdmissionResult{}, err
		}
		if err := tx.SaveAdmission(admission, oldVersion); err != nil {
			return AdmissionResult{}, err
		}
		row, err := tx.LoadOpportunity(admission.OpportunityID)
		if err != nil {
			return AdmissionResult{}, err
		}
		if row.Opportunity.Revision == admission.OpportunityRevision && row.Opportunity.State == "admitted" {
			old := row.Opportunity.Version
			row.Opportunity.State, row.Opportunity.Version = "withdrawn", old+1
			if err := tx.SaveOpportunity(row, old); err != nil {
				return AdmissionResult{}, err
			}
		}
		if err := s.appendEvent(tx, admission.RevokedEventID, "admission_revoked", admission, command.CommandMeta); err != nil {
			return AdmissionResult{}, err
		}
		return AdmissionResult{SchemaVersion: 1, Admission: admission}, nil
	})
}

func (s *AdmissionService) GetAdmission(ctx context.Context, p Principal, id string) (AdmissionResult, error) {
	if err := authorize(p, true); err != nil {
		return AdmissionResult{}, err
	}
	var result AdmissionResult
	if s.repo == nil {
		return result, serviceError(apierrors.UnsupportedCapability, "Admission repository is not connected", "connect_admission_repository")
	}
	err := s.repo.WithAdmissionTx(ctx, func(tx AdmissionTx) error {
		value, err := tx.LoadAdmission(id)
		if err != nil {
			return err
		}
		if value.OwnerID != p.ID {
			return serviceError(apierrors.ScopeDenied, "Admission belongs to another owner namespace", "use_owned_admission")
		}
		result = AdmissionResult{SchemaVersion: 1, Admission: value}
		return nil
	})
	return result, mapError(err, "")
}

// 只供受信任 Workspace bridge 按实际 outbox ID读取，HTTP 不接收替代 snapshot。
func (s *AdmissionService) ReadAdmittedOpportunitySnapshot(ctx context.Context, eventID string) (AdmittedOpportunitySnapshot, error) {
	var result AdmittedOpportunitySnapshot
	if s.repo == nil {
		return result, serviceError(apierrors.UnsupportedCapability, "Admission repository is not connected", "connect_admission_repository")
	}
	err := s.repo.WithAdmissionTx(ctx, func(tx AdmissionTx) error {
		event, err := tx.LoadAdmissionEvent(eventID)
		if err != nil {
			return err
		}
		if event.Type != "attention.opportunity_admitted" {
			return serviceError(apierrors.ScopeDenied, "Event is not an opportunity admission", "use_actual_admission_event")
		}
		var payload struct {
			AdmissionID   string `json:"admission_id"`
			OpportunityID string `json:"opportunity_id"`
			Revision      int    `json:"opportunity_revision"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}
		admission, err := tx.LoadAdmission(payload.AdmissionID)
		if err != nil {
			return err
		}
		if admission.ID != payload.AdmissionID || admission.AdmittedEventID != eventID || admission.ID != event.AggregateID || event.AggregateVersion != 1 || admission.OpportunityID != payload.OpportunityID || admission.OpportunityRevision != payload.Revision || admission.OwnerNamespace != "human:"+admission.OwnerID {
			return serviceError(apierrors.ScopeDenied, "Admission event identity differs from its fixed source", "inspect_admission_source")
		}
		row, err := tx.LoadOpportunity(admission.OpportunityID)
		if err != nil {
			return err
		}
		candidate, err := admissionFixedOpportunity(row, admission.OpportunityRevision)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(admission.Inputs, domainRefs(candidate.EvidenceRefs)) {
			return domain.ErrEvidence
		}
		if err := validateOpportunityAccess(tx, candidate); err != nil {
			return err
		}
		if admission.Authority == domain.AdmissionDelegatedJudgment {
			proof, err := s.judgment(tx.Context(), Principal{Kind: "delegated_judgment", ID: admission.ActorID}, admission.DecisionID, admission.ProjectID, candidate)
			if err != nil {
				return err
			}
			if proof.OwnerID != admission.OwnerID || proof.PolicyVersion != admission.PolicyVersion || proof.Model != admission.Model || proof.ModelVersion != admission.ModelVersion || proof.Backend != admission.Backend || proof.QuestionType != admission.QuestionType || proof.Purpose != admission.Purpose {
				return serviceError(apierrors.ContextStale, "Current decision differs from the fixed admission authority", "review_current_judgment_authority")
			}
		}
		if admission.ProjectID != "" {
			if s.options.Projects == nil {
				return serviceError(apierrors.UnsupportedCapability, "Current project authorization is not connected", "connect_project_permission_reader")
			}
			p := Principal{Kind: "human", ID: admission.OwnerID}
			if admission.Authority == domain.AdmissionDelegatedJudgment {
				p = Principal{Kind: "delegated_judgment", ID: admission.ActorID}
			}
			if err := s.options.Projects.AuthorizeAdmissionProject(tx.Context(), p, admission.ProjectID, candidate.EvidenceRefs); err != nil {
				return err
			}
		}
		result = AdmittedOpportunitySnapshot{EventID: eventID, AdmissionID: admission.ID, AdmissionVersion: admission.Version, OpportunityID: admission.OpportunityID, OpportunityRevision: admission.OpportunityRevision, OwnerNamespace: admission.OwnerNamespace, Title: candidate.Title, Inputs: slices.Clone(candidate.EvidenceRefs), GoalRefs: slices.Clone(candidate.GoalRefs), ProjectID: admission.ProjectID, Authority: admission.Authority, PolicyVersion: admission.PolicyVersion, Active: admission.Status == domain.AdmissionActive}
		return nil
	})
	return result, mapError(err, "")
}
