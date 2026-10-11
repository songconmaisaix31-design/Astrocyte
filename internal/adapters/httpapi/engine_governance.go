package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerEngineGovernance(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/knowledge/{id}", func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project_id")
		var result any
		var err error
		if project == "" {
			err = invalid("project_id is required")
		} else if h.services.KnowledgeGovernance == nil {
			err = apierrors.NewUnsupported("knowledge_read")
		} else {
			entry, e := h.services.KnowledgeGovernance.GetKnowledge(r.Context(), principal(r), project, r.PathValue("id"))
			result = localEnvelope("knowledge", entry)
			err = e
		}
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/experiences/{id}", func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project_id")
		var result any
		var err error
		if project == "" {
			err = invalid("project_id is required")
		} else if h.services.ExperienceGovernance == nil {
			err = apierrors.NewUnsupported("experience_read")
		} else {
			entry, e := h.services.ExperienceGovernance.GetExperience(r.Context(), principal(r), r.PathValue("id"), project)
			result = localEnvelope("experience", entry)
			err = e
		}
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/knowledge/{id}/shares", commandHandler(h, func(c *attentionapp.ShareMemoryCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.ShareMemoryCommand) (any, error) {
		if h.services.KnowledgeGovernance == nil {
			return nil, apierrors.NewUnsupported("knowledge_share")
		}
		return h.services.KnowledgeGovernance.ShareKnowledge(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/experiences/{id}/shares", commandHandler(h, func(c *attentionapp.ShareMemoryCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.ShareMemoryCommand) (any, error) {
		if h.services.ExperienceGovernance == nil {
			return nil, apierrors.NewUnsupported("experience_share")
		}
		return h.services.ExperienceGovernance.ShareExperience(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/knowledge/{id}/demote", commandHandler(h, func(c *attentionapp.DemoteMemoryCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.DemoteMemoryCommand) (any, error) {
		if h.services.KnowledgeGovernance == nil {
			return nil, apierrors.NewUnsupported("knowledge_demote")
		}
		return h.services.KnowledgeGovernance.DemoteKnowledge(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/experiences/{id}/demote", commandHandler(h, func(c *attentionapp.DemoteMemoryCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.DemoteMemoryCommand) (any, error) {
		if h.services.ExperienceGovernance == nil {
			return nil, apierrors.NewUnsupported("experience_demote")
		}
		return h.services.ExperienceGovernance.DemoteExperience(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/knowledge/{id}/references", commandHandler(h, func(c *attentionapp.ReferenceKnowledgeCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.ReferenceKnowledgeCommand) (any, error) {
		if h.services.KnowledgeGovernance == nil {
			return nil, apierrors.NewUnsupported("knowledge_reference")
		}
		return h.services.KnowledgeGovernance.ReferenceKnowledge(r.Context(), principal(r), r.PathValue("id"), c)
	}))
}
