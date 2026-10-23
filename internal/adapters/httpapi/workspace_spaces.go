package httpapi

import (
	"net/http"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func (h *handler) registerWorkspaceSpaces(mux *http.ServeMux) {
	spaces := h.services.ProjectSpaces
	if spaces == nil {
		return
	}
	mux.HandleFunc("GET /api/v1/project-spaces", func(w http.ResponseWriter, r *http.Request) {
		result, err := spaces.ListProjectSpaces(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces", commandHandler(h, func(c *attention.ProjectSpaceCommand) *attention.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, c attention.ProjectSpaceCommand) (any, error) {
		return spaces.CreateProjectSpace(r.Context(), localCaller(r), workspace.ProjectSpaceCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: c.SchemaVersion, RequestID: c.RequestID, ExpectedVersion: c.ExpectedVersion, IdempotencyKey: c.IdempotencyKey}, Title: c.Title})
	}))
	mux.HandleFunc("GET /api/v1/project-spaces/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := spaces.GetProjectSpace(r.Context(), localCaller(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	if h.services.Attention == nil {
		mux.HandleFunc("POST /api/v1/project-spaces/{id}/references", h.notImplemented("project_material_access"))
		mux.HandleFunc("POST /api/v1/project-spaces/{id}/references/remove", h.notImplemented("project_material_access"))
	}
}
