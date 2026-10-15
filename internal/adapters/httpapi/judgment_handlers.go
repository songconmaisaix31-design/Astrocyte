package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type JudgmentPolicies interface {
	SetPolicy(context.Context, judgment.Actor, judgment.Policy, int) error
	Snapshot(context.Context, judgment.Actor, string, []string) (domain.ProjectContextSnapshot, error)
	ReadPolicy(context.Context, judgment.Actor, string, judgment.QuestionType) (judgment.Policy, error)
}

type JudgmentServices struct {
	Judgments  *judgment.Service
	Policies   JudgmentPolicies
	Delegation *judgment.DelegationService
	Admissions judgment.DelegatedAdmissions
}

type JudgmentFeedbackCommand struct {
	ResultVersion int    `json:"result_version"`
	Kind          string `json:"kind"`
	Reason        string `json:"reason"`
}

type JudgmentPolicyCommand struct {
	ExpectedRevision int             `json:"expected_revision"`
	Policy           judgment.Policy `json:"policy"`
}

type JudgmentSnapshotCommand struct {
	Files []string `json:"files"`
}

func judgmentCaller(r *http.Request, humanOnly bool) (judgment.Actor, error) {
	caller := localCaller(r)
	project := r.PathValue("project_id")
	if project == "" || caller.ID == "" || (caller.Kind != "human" && caller.Kind != "agent") || (humanOnly && caller.Kind != "human") || (caller.Kind == "agent" && caller.ProjectID != project) {
		return judgment.Actor{}, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Current project identity is required for this judgment operation"}
	}
	return judgment.Actor{ID: caller.ID, Kind: caller.Kind, ProjectID: project}, nil
}

func judgmentUnavailable() error {
	return &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "Judgment service is unavailable", RequiredAction: "configure_judgment_services"}
}

func (h *handler) judgmentResult(w http.ResponseWriter, r *http.Request, result any, err error) {
	if err == nil {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, result)
		return
	}
	var serviceError *apierrors.ServiceError
	if !errors.As(err, &serviceError) {
		serviceError = &apierrors.ServiceError{Code: apierrors.InternalError, Message: "Judgment operation failed", RequiredAction: "inspect_original_error_with_request_id"}
	}
	status := http.StatusInternalServerError
	switch serviceError.Code {
	case apierrors.NotFound:
		status = http.StatusNotFound
	case apierrors.ValidationFailed, apierrors.EvidenceMissing:
		status = http.StatusBadRequest
	case apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.ApprovalRevoked:
		status = http.StatusForbidden
	case apierrors.VersionConflict, apierrors.ContextStale, apierrors.IndexStale, apierrors.DeliveryUnknown:
		status = http.StatusConflict
	case apierrors.BudgetExhausted:
		status = http.StatusTooManyRequests
	case apierrors.ProviderUnavailable:
		status = http.StatusServiceUnavailable
	case apierrors.UnsupportedCapability:
		status = http.StatusNotImplemented
	}
	writeError(w, r, status, serviceError)
}

func (h *handler) registerJudgment(mux *http.ServeMux, services JudgmentServices) {
	h.registerDelegation(mux, services)
	const prefix = "/api/v1/local-projects/{project_id}"
	mux.HandleFunc("POST "+prefix+"/judgments", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, false)
		var request judgment.JudgmentRequest
		if err == nil {
			request, err = decodeAttention[judgment.JudgmentRequest](w, r)
		}
		if err == nil && request.ProjectID != actor.ProjectID {
			err = &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Judgment body and path project must match"}
		}
		if err == nil && request.Context != nil {
			err = invalid("Bound judgments require the configured delegation workflow")
		}
		var result judgment.JudgmentResult
		if err == nil {
			if services.Judgments == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Judgments.Evaluate(r.Context(), actor, request)
				var structured *apierrors.ServiceError
				if err != nil && !errors.As(err, &structured) && (result.ReasonCode == "backend_unavailable" || result.ReasonCode == "backend_call_failed") {
					err = &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "The selected judgment backend failed", RequiredAction: "inspect_the_selected_backend_before_an_explicit_new_request"}
				}
			}
		}
		h.judgmentResult(w, r, localEnvelope("judgment", result), err)
	})
	mux.HandleFunc("GET "+prefix+"/judgments/{decision_id}", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, false)
		var result judgment.JudgmentResult
		if err == nil {
			if services.Judgments == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Judgments.Read(r.Context(), actor, r.PathValue("decision_id"))
			}
		}
		h.judgmentResult(w, r, localEnvelope("judgment", result), err)
	})
	mux.HandleFunc("POST "+prefix+"/judgments/{decision_id}/feedback", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command JudgmentFeedbackCommand
		if err == nil {
			command, err = decodeAttention[JudgmentFeedbackCommand](w, r)
		}
		var result judgment.Feedback
		if err == nil {
			if services.Judgments == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Judgments.Feedback(r.Context(), actor, r.PathValue("decision_id"), command.ResultVersion, command.Kind, command.Reason)
			}
		}
		h.judgmentResult(w, r, localEnvelope("feedback", result), err)
	})
	mux.HandleFunc("GET "+prefix+"/judgments/{decision_id}/feedback", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, false)
		var result []judgment.Feedback
		if err == nil {
			if services.Judgments == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Judgments.FeedbackHistory(r.Context(), actor, r.PathValue("decision_id"))
			}
		}
		h.judgmentResult(w, r, localList(result), err)
	})
	mux.HandleFunc("PUT "+prefix+"/judgment-policies/{question_type}", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command JudgmentPolicyCommand
		if err == nil {
			command, err = decodeAttention[JudgmentPolicyCommand](w, r)
		}
		if err == nil && (command.ExpectedRevision < 0 || command.Policy.ProjectID != actor.ProjectID || command.Policy.QuestionType != judgment.QuestionType(r.PathValue("question_type"))) {
			err = invalid("Policy project, question type and expected revision must match the request path")
		}
		if err == nil {
			if services.Policies == nil {
				err = judgmentUnavailable()
			} else {
				err = services.Policies.SetPolicy(r.Context(), actor, command.Policy, command.ExpectedRevision)
			}
		}
		h.judgmentResult(w, r, localEnvelope("policy", command.Policy), err)
	})
	mux.HandleFunc("POST "+prefix+"/judgment-snapshots", func(w http.ResponseWriter, r *http.Request) {
		actor, err := judgmentCaller(r, true)
		var command JudgmentSnapshotCommand
		if err == nil {
			command, err = decodeAttention[JudgmentSnapshotCommand](w, r)
		}
		var snapshot domain.ProjectContextSnapshot
		if err == nil {
			if services.Policies == nil {
				err = judgmentUnavailable()
			} else {
				snapshot, err = services.Policies.Snapshot(r.Context(), actor, actor.ProjectID, command.Files)
			}
		}
		h.judgmentResult(w, r, localEnvelope("snapshot", snapshot), err)
	})
}
