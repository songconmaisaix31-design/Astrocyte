package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

type DelegationRuleCommand struct {
	ExpectedRevision int                     `json:"expected_revision"`
	Rule             judgment.DelegationRule `json:"rule"`
}

type DelegationStateCommand struct {
	ExpectedRevision int    `json:"expected_revision"`
	State            string `json:"state"`
}

type DelegationJudgmentCommand struct {
	RuleRevision        int    `json:"rule_revision"`
	OpportunityID       string `json:"opportunity_id"`
	OpportunityRevision int    `json:"opportunity_revision"`
}

func (h *handler) registerDelegation(mux *http.ServeMux, services JudgmentServices) {
	const prefix = "/api/v1/local-projects/{project_id}"
	mux.HandleFunc("GET "+prefix+"/judgment-policies/{question_type}", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var policy judgment.Policy
		if err == nil {
			if services.Policies == nil {
				err = judgmentUnavailable()
			} else {
				policy, err = services.Policies.ReadPolicy(r.Context(), actor, actor.ProjectID, judgment.QuestionType(r.PathValue("question_type")))
			}
		}
		h.judgmentResult(w, r, localEnvelope("policy", publicJudgmentPolicy(policy)), err)
	})
	mux.HandleFunc("GET "+prefix+"/delegation-rules", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var rules []judgment.DelegationRule
		if err == nil {
			if services.Delegation == nil {
				err = judgmentUnavailable()
			} else {
				rules, err = services.Delegation.ListRules(r.Context(), actor, actor.ProjectID)
			}
		}
		h.judgmentResult(w, r, localList(rules), err)
	})
	mux.HandleFunc("GET "+prefix+"/delegation-rules/{rule_id}", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var rule judgment.DelegationRule
		if err == nil {
			if services.Delegation == nil {
				err = judgmentUnavailable()
			} else {
				rule, err = services.Delegation.ReadRule(r.Context(), actor, r.PathValue("rule_id"))
			}
		}
		h.judgmentResult(w, r, localEnvelope("delegation_rule", rule), err)
	})
	mux.HandleFunc("PUT "+prefix+"/delegation-rules/{rule_id}", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command DelegationRuleCommand
		if err == nil {
			command, err = decodeAttention[DelegationRuleCommand](w, r)
		}
		if err == nil && (command.Rule.ProjectID != actor.ProjectID || command.Rule.ID != r.PathValue("rule_id")) {
			err = &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Delegation rule body and path identities must match"}
		}
		var rule judgment.DelegationRule
		if err == nil {
			if services.Delegation == nil {
				err = judgmentUnavailable()
			} else {
				err = services.Delegation.SetRule(r.Context(), actor, command.Rule, command.ExpectedRevision)
				if err == nil {
					rule, err = services.Delegation.ReadRule(r.Context(), actor, command.Rule.ID)
				}
			}
		}
		h.judgmentResult(w, r, localEnvelope("delegation_rule", rule), err)
	})
	mux.HandleFunc("POST "+prefix+"/delegation-rules/{rule_id}/state", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command DelegationStateCommand
		if err == nil {
			command, err = decodeAttention[DelegationStateCommand](w, r)
		}
		var rule judgment.DelegationRule
		if err == nil {
			if services.Delegation == nil {
				err = judgmentUnavailable()
			} else {
				rule, err = services.Delegation.SetState(r.Context(), actor, r.PathValue("rule_id"), command.ExpectedRevision, command.State)
			}
		}
		h.judgmentResult(w, r, localEnvelope("delegation_rule", rule), err)
	})
	mux.HandleFunc("POST "+prefix+"/delegation-rules/{rule_id}/judgments", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command DelegationJudgmentCommand
		if err == nil {
			command, err = decodeAttention[DelegationJudgmentCommand](w, r)
		}
		var result judgment.JudgmentResult
		if err == nil {
			if services.Delegation == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Delegation.Evaluate(r.Context(), actor, r.PathValue("rule_id"), command.RuleRevision, command.OpportunityID, command.OpportunityRevision)
			}
		}
		h.judgmentResult(w, r, localEnvelope("judgment", result), err)
	})
	mux.HandleFunc("POST "+prefix+"/judgments/{decision_id}/admissions", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command judgment.DelegatedAdmissionCommand
		if err == nil {
			command, err = decodeAttention[judgment.DelegatedAdmissionCommand](w, r)
		}
		command.IdempotencyKey = r.Header.Get("Idempotency-Key")
		var receipt judgment.DelegatedAdmissionReceipt
		if err == nil {
			if services.Admissions == nil {
				err = judgmentUnavailable()
			} else {
				receipt, err = services.Admissions.AdmitJudgment(r.Context(), actor, r.PathValue("decision_id"), command)
			}
		}
		h.judgmentResult(w, r, localEnvelope("admission", receipt), err)
	})
}
