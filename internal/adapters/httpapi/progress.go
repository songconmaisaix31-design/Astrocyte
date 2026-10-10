package httpapi

import (
	"net/http"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type progressSetCommand struct {
	attentionapp.CommandMeta
	domain.ProgressInput
}

type progressInferCommand struct {
	attentionapp.CommandMeta
	domain.ProgressCommand
}

// registerProgress wires project stage reading/writing and model-inferred
// progress. Reading is a durable-cache GET; human writes and inference are
// separate human-initiated operations. A missing service is an explicit 501,
// never an empty success. Evidence is file provenance only; a status never
// grants or expands project permission.
func (h *handler) registerProgress(mux *http.ServeMux) {
	s := h.services.Progress
	if s == nil {
		for _, route := range []string{
			"GET /api/v1/local-projects/{id}/progress",
			"PUT /api/v1/local-projects/{id}/progress",
			"POST /api/v1/local-projects/{id}/progress/infer",
		} {
			mux.HandleFunc(route, h.notImplemented("project_progress"))
		}
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects/{id}/progress", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.GetProjectProgress(r.Context(), localCaller(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("progress", p), err)
	})
	mux.HandleFunc("PUT /api/v1/local-projects/{id}/progress", commandHandler(h, func(c *progressSetCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c progressSetCommand) (any, error) {
		p, err := s.SetProjectProgress(r.Context(), localCaller(r), r.PathValue("id"), c.ProgressInput)
		return localEnvelope("progress", p), err
	}))
	mux.HandleFunc("POST /api/v1/local-projects/{id}/progress/infer", commandHandler(h, func(c *progressInferCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c progressInferCommand) (any, error) {
		p, err := s.InferProjectProgress(r.Context(), localCaller(r), r.PathValue("id"), c.ProgressCommand)
		return localEnvelope("progress", p), err
	}))
}
