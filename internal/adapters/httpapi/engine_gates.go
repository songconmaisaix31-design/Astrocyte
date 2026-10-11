package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// registerEngineGates exposes the three human gates (knowledge confirm,
// experience approve, conflict adjudication). The gates are not assembled until
// the W4/W5/W2 stores exist, so an absent service reports unsupported rather
// than a fabricated transition.
func (h *handler) registerEngineGates(mux *http.ServeMux) {
	h.registerEngineGovernance(mux)
	mux.HandleFunc("POST /api/v1/knowledge/{id}/confirm", h.engineGateCommand(
		func(c *attentionapp.ConfirmKnowledgeCommand) *attentionapp.CommandMeta { return &c.CommandMeta },
		http.StatusOK,
		func(g attentionapp.EngineGates, r *http.Request, c attentionapp.ConfirmKnowledgeCommand) (any, error) {
			if c.OwnerProject == "" {
				return nil, invalid("owner_project is required")
			}
			if h.services.KnowledgeGovernance != nil {
				return h.services.KnowledgeGovernance.ConfirmKnowledge(r.Context(), principal(r), r.PathValue("id"), c)
			}
			if g == nil {
				return nil, apierrors.NewUnsupported("knowledge_confirmation")
			}
			return g.ConfirmKnowledge(r.Context(), principal(r), r.PathValue("id"), c)
		}))
	mux.HandleFunc("POST /api/v1/experiences/{id}/approve", h.engineGateCommand(
		func(c *attentionapp.ApproveExperienceCommand) *attentionapp.CommandMeta { return &c.CommandMeta },
		http.StatusOK,
		func(g attentionapp.EngineGates, r *http.Request, c attentionapp.ApproveExperienceCommand) (any, error) {
			if c.OwnerProject == "" {
				return nil, invalid("owner_project is required")
			}
			if h.services.ExperienceGovernance != nil {
				return h.services.ExperienceGovernance.ApproveExperience(r.Context(), principal(r), r.PathValue("id"), c)
			}
			if g == nil {
				return nil, apierrors.NewUnsupported("experience_approval")
			}
			return g.ApproveExperience(r.Context(), principal(r), r.PathValue("id"), c)
		}))
	mux.HandleFunc("POST /api/v1/conflicts/{id}/adjudications", h.engineGateCommand(
		func(c *attentionapp.AdjudicateConflictCommand) *attentionapp.CommandMeta { return &c.CommandMeta },
		http.StatusOK,
		func(g attentionapp.EngineGates, r *http.Request, c attentionapp.AdjudicateConflictCommand) (any, error) {
			if h.services.ConflictGate != nil {
				return h.services.ConflictGate.AdjudicateConflict(r.Context(), principal(r), r.PathValue("id"), c)
			}
			if g == nil {
				return nil, apierrors.NewUnsupported("conflict_adjudication")
			}
			return g.AdjudicateConflict(r.Context(), principal(r), r.PathValue("id"), c)
		}))
}

// engineGateCommand wraps a CSRF command handler with the not-assembled check.
func (h *handler) engineGateCommand[T any](meta func(*T) *attentionapp.CommandMeta, status int, call func(attentionapp.EngineGates, *http.Request, T) (any, error)) http.HandlerFunc {
	return commandHandler(h, meta, status, func(r *http.Request, c T) (any, error) {
		return call(h.services.Gates, r, c)
	})
}
