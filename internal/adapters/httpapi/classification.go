package httpapi

import (
	"net/http"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerClassification(mux *http.ServeMux) {
	s := h.services.Attention
	mux.HandleFunc("GET /api/v1/material-domains", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.ListMaterialDomains(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/material-domains", commandHandler(h, func(c *attentionapp.MaterialDomainCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, c attentionapp.MaterialDomainCommand) (any, error) {
		return s.CreateMaterialDomain(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("PATCH /api/v1/material-domains/{id}", commandHandler(h, func(c *attentionapp.MaterialDomainCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.MaterialDomainCommand) (any, error) {
		return s.ReviseMaterialDomain(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("PUT /api/v1/materials/{id}/domains", commandHandler(h, func(c *attentionapp.SetMaterialDomainsCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.SetMaterialDomainsCommand) (any, error) {
		return s.SetMaterialDomains(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("GET /api/v1/project-spaces", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.ListProjectSpaces(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces", commandHandler(h, func(c *attentionapp.ProjectSpaceCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, c attentionapp.ProjectSpaceCommand) (any, error) {
		return s.CreateProjectSpace(r.Context(), principal(r), c)
	}))
	mux.HandleFunc("GET /api/v1/project-spaces/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.GetProjectSpace(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces/{id}/references", commandHandler(h, func(c *attentionapp.ReferenceMaterialCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.ReferenceMaterialCommand) (any, error) {
		return s.ReferenceMaterial(r.Context(), principal(r), r.PathValue("id"), c)
	}))
	mux.HandleFunc("POST /api/v1/project-spaces/{id}/references/remove", commandHandler(h, func(c *attentionapp.RemoveMaterialReferenceCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.RemoveMaterialReferenceCommand) (any, error) {
		return s.RemoveMaterialReference(r.Context(), principal(r), r.PathValue("id"), c)
	}))
}
