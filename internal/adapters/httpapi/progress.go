package httpapi

import (
	"net/http"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type inferProgressCommand struct {
	attentionapp.CommandMeta
	CLI string `json:"cli,omitempty"`
}

// registerProgress wires model-inferred project progress. Until a domain
// inference service is assembled the routes return 501 unsupported_capability.
// Inference reads only approved fixed TASK/STATUS files and attaches source,
// version and freshness evidence; it is never authoritative project state.
func (h *handler) registerProgress(mux *http.ServeMux) {
	s := h.services.Progress
	if s == nil {
		mux.HandleFunc("GET /api/v1/local-projects/{id}/progress", h.notImplemented("project_progress"))
		mux.HandleFunc("POST /api/v1/local-projects/{id}/progress/infer", h.notImplemented("project_progress"))
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects/{id}/progress", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.GetProjectProgress(r.Context(), localCaller(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("progress", p), err)
	})
	mux.HandleFunc("POST /api/v1/local-projects/{id}/progress/infer", commandHandler(h, func(c *inferProgressCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c inferProgressCommand) (any, error) {
		p, err := s.InferProjectProgress(r.Context(), localCaller(r), r.PathValue("id"), workspaceapp.InferProgressCommand{CLI: c.CLI})
		return localEnvelope("progress", p), err
	}))
}
