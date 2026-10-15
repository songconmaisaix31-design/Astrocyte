package judgment

import (
	"context"
	"errors"
	"math"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type DelegationService struct {
	repo          DelegationRepository
	core          *Service
	access        DelegationPolicyAccess
	opportunities OpportunityReader
}

func NewDelegationService(repo DelegationRepository, core *Service, access DelegationPolicyAccess, opportunities OpportunityReader) (*DelegationService, error) {
	if repo == nil || core == nil || access == nil || opportunities == nil {
		return nil, invalid("delegation requires actual persistence, judgment, authorization and opportunity sources")
	}
	return &DelegationService{repo: repo, core: core, access: access, opportunities: opportunities}, nil
}

func ValidateDelegationRule(rule DelegationRule) error {
	if _, err := uuid.Parse(rule.ID); err != nil {
		return invalid("delegation rule id must be a UUID")
	}
	if rule.ProjectID == "" || strings.TrimSpace(rule.WorkflowID) == "" || rule.OwnerID == "" || rule.Revision < 1 || rule.PolicyRevision < 1 || rule.Language == "" || rule.Purpose != AdmissionPurpose {
		return invalid("delegation requires fixed project, workflow, human owner, versions and admission purpose")
	}
	if !slices.Contains([]string{"disabled", "active", "paused", "revoked"}, rule.State) {
		return invalid("delegation state must be explicit")
	}
	if len(rule.Materials) == 0 || len(rule.ReviewConditions) == 0 || !slices.Contains(rule.ReviewConditions, "fixed_inputs") {
		return invalid("delegation requires explicit fixed material scope and fixed_inputs review")
	}
	seen := map[EvidenceRef]bool{}
	for _, ref := range rule.Materials {
		if err := validRef(ref); err != nil {
			return err
		}
		if seen[ref] {
			return invalid("delegation material references must be unique")
		}
		seen[ref] = true
	}
	conditions := map[string]bool{}
	for _, condition := range rule.ReviewConditions {
		if !slices.Contains([]string{"fixed_inputs", "human_adopt"}, condition) || conditions[condition] {
			return invalid("delegation review conditions must be supported and unique")
		}
		conditions[condition] = true
	}
	if rule.Question.Type == Choice {
		if rule.AdmissionChoice == nil || rule.AdmissionValueMin != nil || rule.AdmissionValueMax != nil {
			return invalid("choice delegation requires an explicit admission option")
		}
		found := false
		for _, option := range rule.Question.Options {
			found = found || option.ID == *rule.AdmissionChoice
		}
		if !found {
			return invalid("admission option is outside the configured question")
		}
	} else {
		if rule.AdmissionChoice != nil || rule.AdmissionValueMin == nil || rule.AdmissionValueMax == nil {
			return invalid("truth and score delegation require explicit acceptance bounds")
		}
		limit := 1.0
		if rule.Question.Type == Score {
			limit = float64(len(rule.Question.Levels) - 1)
		}
		for _, bound := range []*float64{rule.AdmissionValueMin, rule.AdmissionValueMax} {
			if math.IsNaN(*bound) || math.IsInf(*bound, 0) || *bound < 0 || *bound > limit {
				return invalid("delegation acceptance is outside the question range")
			}
		}
		if *rule.AdmissionValueMin > *rule.AdmissionValueMax {
			return invalid("delegation acceptance bounds are reversed")
		}
	}
	return nil
}

func (s *DelegationService) policy(ctx context.Context, actor Actor, rule DelegationRule) (Policy, error) {
	policy, err := s.access.ReadPolicy(ctx, actor, rule.ProjectID, rule.Question.Type)
	if err != nil {
		return policy, err
	}
	if policy.Revision != rule.PolicyRevision {
		return policy, conflict("delegation targets a different judgment policy revision")
	}
	if rule.State == "active" && policy.Sources == nil {
		return policy, &apierrors.ServiceError{Code: apierrors.ApprovalRequired, Message: "automatic use requires the explicitly selected current Goal and complete taste profile"}
	}
	if !slices.Contains(policy.Languages, rule.Language) {
		return policy, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "delegation language is outside the explicitly selected policy"}
	}
	return policy, nil
}

func delegationRequest(policy Policy, rule DelegationRule, refs []EvidenceRef, binding *DecisionContext) JudgmentRequest {
	return JudgmentRequest{SchemaVersion: "1.0", ProjectID: rule.ProjectID, WorktreeID: policy.WorktreeID, SnapshotID: policy.SnapshotID, Question: rule.Question, EvidenceRefs: refs, Language: rule.Language, Backend: policy.Backend, ExternalScope: policy.ExternalScope, Versions: policy.Versions, Budget: policy.Budget, Context: binding, Sources: policy.Sources}
}

func (s *DelegationService) SetRule(ctx context.Context, actor Actor, rule DelegationRule, expected int) error {
	if actor.Kind != "human" || actor.ID == "" || actor.ProjectID != "" && actor.ProjectID != rule.ProjectID {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "only the authenticated human owner can configure delegation"}
	}
	if rule.OwnerID != "" && rule.OwnerID != actor.ID {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "delegation owner must be the actual human caller"}
	}
	rule.OwnerID = actor.ID
	if expected < 0 || rule.Revision != expected+1 {
		return conflict("delegation requires the next exact rule revision")
	}
	if err := ValidateDelegationRule(rule); err != nil {
		return err
	}
	return s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		if expected > 0 {
			old, err := s.repo.LoadDelegationRule(ctx, rule.ID)
			if err != nil {
				return err
			}
			if old.Revision != expected || old.OwnerID != actor.ID || old.ProjectID != rule.ProjectID || old.WorkflowID != rule.WorkflowID || old.State == "revoked" && rule.State != "revoked" {
				return conflict("delegation identity, revision or terminal revoked state changed")
			}
		}
		policy, err := s.policy(ctx, actor, rule)
		if err != nil {
			return err
		}
		request := delegationRequest(policy, rule, rule.Materials, nil)
		if err := ValidateRequest(request); err != nil {
			return err
		}
		if rule.State == "active" {
			if policy.Backend == JevBackend && (rule.AllowOpportunityExternal == nil || !*rule.AllowOpportunityExternal) {
				return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "cloud delegation requires explicit opportunity metadata external permission"}
			}
			if _, err := s.core.resolve(ctx, actor, request); err != nil {
				return err
			}
		}
		return s.repo.SaveDelegationRule(ctx, rule, expected)
	})
}

func (s *DelegationService) ReadRule(ctx context.Context, actor Actor, id string) (DelegationRule, error) {
	rule, err := s.repo.LoadDelegationRule(ctx, id)
	if err != nil {
		return rule, err
	}
	if actor.Kind != "human" || actor.ID != rule.OwnerID || actor.ProjectID != "" && actor.ProjectID != rule.ProjectID {
		return rule, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "delegation configuration belongs to its human owner"}
	}
	if _, err := s.access.ReadPolicy(ctx, actor, rule.ProjectID, rule.Question.Type); err != nil {
		return rule, err
	}
	return rule, nil
}

func (s *DelegationService) ListRules(ctx context.Context, actor Actor, projectID string) ([]DelegationRule, error) {
	if actor.Kind != "human" || actor.ID == "" || actor.ProjectID != "" && actor.ProjectID != projectID {
		return nil, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "delegation configuration requires the actual project human"}
	}
	rules, err := s.repo.ListDelegationRules(ctx, projectID)
	if err != nil {
		return nil, err
	}
	result := []DelegationRule{}
	for _, rule := range rules {
		if rule.OwnerID == actor.ID {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (s *DelegationService) SetState(ctx context.Context, actor Actor, id string, expected int, state string) (DelegationRule, error) {
	var rule DelegationRule
	err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		var err error
		rule, err = s.repo.LoadDelegationRule(ctx, id)
		if err != nil {
			return err
		}
		if actor.Kind != "human" || actor.ID != rule.OwnerID || actor.ProjectID != "" && actor.ProjectID != rule.ProjectID {
			return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "only the actual delegation owner can change its state"}
		}
		if rule.Revision != expected || rule.State == "revoked" && state != "revoked" {
			return conflict("delegation revision or terminal revoked state changed")
		}
		rule.Revision++
		rule.State = state
		if err := ValidateDelegationRule(rule); err != nil {
			return err
		}
		if state == "active" {
			return s.SetRule(ctx, actor, rule, expected)
		}
		return s.repo.SaveDelegationRule(ctx, rule, expected)
	})
	return rule, err
}

func opportunityRefs(opportunity OpportunitySnapshot) []EvidenceRef {
	refs := make([]EvidenceRef, len(opportunity.Evidence))
	for i, entry := range opportunity.Evidence {
		refs[i] = entry.Reference
	}
	return refs
}

func sameOpportunity(a, b OpportunitySnapshot) bool {
	a.State, b.State = "", ""
	a.Version, b.Version = 0, 0
	return reflect.DeepEqual(a, b)
}

func (s *DelegationService) currentOpportunity(ctx context.Context, rule DelegationRule, id string, revision int) (OpportunitySnapshot, error) {
	opportunity, err := s.opportunities.ReadJudgmentOpportunity(ctx, Actor{ID: rule.OwnerID, Kind: "human", ProjectID: rule.ProjectID}, id, revision)
	if err != nil {
		return opportunity, err
	}
	if opportunity.ID != id || opportunity.Revision != revision || opportunity.Version < 1 || len(opportunity.MissingEvidence) != 0 || len(opportunity.Evidence) == 0 {
		return opportunity, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "delegation requires the actual complete fixed opportunity"}
	}
	for _, ref := range opportunityRefs(opportunity) {
		if !slices.Contains(rule.Materials, ref) {
			return opportunity, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "opportunity evidence exceeds the exact human delegation scope"}
		}
	}
	return opportunity, nil
}

func (s *DelegationService) Evaluate(ctx context.Context, actor Actor, ruleID string, ruleRevision int, opportunityID string, opportunityRevision int) (JudgmentResult, error) {
	var result JudgmentResult
	rule, err := s.repo.LoadDelegationRule(ctx, ruleID)
	if err != nil {
		return result, err
	}
	if rule.State != "active" || rule.Revision != ruleRevision || actor.Kind != "human" || actor.ID != rule.OwnerID || actor.ProjectID != "" && actor.ProjectID != rule.ProjectID {
		return result, &apierrors.ServiceError{Code: apierrors.ApprovalRequired, Message: "current explicit human delegation and its owner workflow are required"}
	}
	policy, err := s.policy(ctx, actor, rule)
	if err != nil {
		return result, err
	}
	opportunity, err := s.currentOpportunity(ctx, rule, opportunityID, opportunityRevision)
	if err != nil {
		return result, err
	}
	if opportunity.State != "ready_for_review" {
		return result, conflict("only the current ready_for_review opportunity can enter delegation")
	}
	if policy.Backend == JevBackend && (rule.AllowOpportunityExternal == nil || !*rule.AllowOpportunityExternal) {
		return result, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "opportunity external permission is absent"}
	}
	bindingContext := DecisionContext{Purpose: AdmissionPurpose, RuleID: rule.ID, RuleVersion: rule.Revision, WorkflowID: rule.WorkflowID, OwnerID: rule.OwnerID, Opportunity: opportunity}
	result, err = s.core.Evaluate(ctx, actor, delegationRequest(policy, rule, opportunityRefs(opportunity), &bindingContext))
	if err != nil {
		return result, err
	}
	err = s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		currentRule, err := s.repo.LoadDelegationRule(ctx, rule.ID)
		if err != nil {
			return err
		}
		current, err := s.currentOpportunity(ctx, currentRule, opportunityID, opportunityRevision)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(rule, currentRule) || current.State != "ready_for_review" || !sameOpportunity(opportunity, current) {
			return conflict("delegation or fixed opportunity changed during judgment")
		}
		if _, err := s.core.Revalidate(ctx, actor, result.ID); err != nil {
			return err
		}
		old, err := s.repo.LoadAdmissionBinding(ctx, result.ID)
		if err == nil {
			if old.ResultVersion != result.Version || old.ProjectID != rule.ProjectID || old.Actor != actor || !reflect.DeepEqual(old.Context, bindingContext) {
				return conflict("cached judgment has a different immutable admission binding")
			}
			return nil
		}
		var missing *apierrors.ServiceError
		if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
			return err
		}
		return s.repo.SaveAdmissionBinding(ctx, AdmissionBinding{ResultID: result.ID, ResultVersion: result.Version, ProjectID: rule.ProjectID, Actor: actor, Context: bindingContext, CreatedAt: time.Now().UTC()})
	})
	return result, err
}

func (s *DelegationService) Authorize(ctx context.Context, workflowID, resultID, projectID, opportunityID string, revision int) (Record, DelegationRule, AdmissionBinding, error) {
	var record Record
	var rule DelegationRule
	var binding AdmissionBinding
	err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		var err error
		binding, err = s.repo.LoadAdmissionBinding(ctx, resultID)
		if err != nil {
			return err
		}
		c := binding.Context
		if binding.ProjectID != projectID || c.Purpose != AdmissionPurpose || c.WorkflowID != workflowID || c.Opportunity.ID != opportunityID || c.Opportunity.Revision != revision {
			return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "fixed admission binding does not match this workflow and opportunity"}
		}
		rule, err = s.repo.LoadDelegationRule(ctx, c.RuleID)
		if err != nil {
			return err
		}
		if rule.State != "active" || rule.Revision != c.RuleVersion || rule.OwnerID != c.OwnerID || rule.WorkflowID != workflowID || rule.ProjectID != projectID {
			return &apierrors.ServiceError{Code: apierrors.ApprovalRevoked, Message: "delegation is paused, revoked or has changed version"}
		}
		current, err := s.currentOpportunity(ctx, rule, opportunityID, revision)
		if err != nil {
			return err
		}
		if !slices.Contains([]string{"ready_for_review", "admitted"}, current.State) || !sameOpportunity(current, c.Opportunity) {
			return conflict("actual fixed opportunity changed before admission use")
		}
		policy, err := s.policy(ctx, binding.Actor, rule)
		if err != nil {
			return err
		}
		record, err = s.core.Revalidate(ctx, binding.Actor, resultID)
		if err != nil {
			return err
		}
		if record.Request.Context == nil || !reflect.DeepEqual(*record.Request.Context, c) || record.Request.Versions != policy.Versions || record.Result.Version != binding.ResultVersion || !record.Cacheable || record.Result.Decision != Recommend || record.Result.Truncated || incompleteOptions(record.Result.Options) {
			return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "actual judgment does not provide a complete current admission recommendation"}
		}
		if record.Result.Backend != RuleBackend && (record.Result.Usage == nil || record.Result.Usage.Truncated == nil || *record.Result.Usage.Truncated) {
			return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "automatic admission requires a known actual model truncation status"}
		}
		if rule.Question.Type == Choice {
			if record.Result.Selected == nil || *record.Result.Selected != *rule.AdmissionChoice {
				return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "actual judgment did not select the configured admission option"}
			}
		} else if record.Result.Value == nil || *record.Result.Value < *rule.AdmissionValueMin || *record.Result.Value > *rule.AdmissionValueMax {
			return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "actual judgment is outside explicit delegation acceptance bounds"}
		}
		if slices.Contains(rule.ReviewConditions, "human_adopt") {
			history, err := s.core.FeedbackHistory(ctx, Actor{ID: rule.OwnerID, Kind: "human", ProjectID: projectID}, resultID)
			if err != nil {
				return err
			}
			if len(history) == 0 || history[len(history)-1].ActorID != rule.OwnerID || history[len(history)-1].Kind != "adopt" || history[len(history)-1].ResultVersion != record.Result.Version {
				return &apierrors.ServiceError{Code: apierrors.ApprovalRequired, Message: "delegation requires current fixed-version human adoption feedback"}
			}
		}
		return nil
	})
	return record, rule, binding, err
}

func (s *DelegationService) AuthorizeForOwner(ctx context.Context, actor Actor, resultID string) (Record, DelegationRule, AdmissionBinding, error) {
	binding, err := s.repo.LoadAdmissionBinding(ctx, resultID)
	if err != nil {
		return Record{}, DelegationRule{}, binding, err
	}
	if actor.Kind != "human" || actor.ID != binding.Context.OwnerID || actor.ProjectID != "" && actor.ProjectID != binding.ProjectID {
		return Record{}, DelegationRule{}, binding, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "delegated admission requires the actual configured human workflow owner"}
	}
	return s.Authorize(ctx, binding.Context.WorkflowID, resultID, binding.ProjectID, binding.Context.Opportunity.ID, binding.Context.Opportunity.Revision)
}
