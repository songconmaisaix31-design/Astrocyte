package httpapi

import (
	"net/http"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// registerPaperSearch wires public academic metadata search. Until a domain
// search adapter is assembled the route returns 501 unsupported_capability.
// Search never imports, extracts full text or calls a model; selection and
// import reuse the existing ImportMaterial/job/revision path.
func (h *handler) registerPaperSearch(mux *http.ServeMux) {
	s := h.services.PaperSearch
	if s == nil {
		mux.HandleFunc("POST /api/v1/paper/search", h.notImplemented("paper_search"))
		return
	}
	mux.HandleFunc("POST /api/v1/paper/search", commandHandler(h, func(c *attentionapp.PaperSearchCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attentionapp.PaperSearchCommand) (any, error) {
		return s.SearchPapers(r.Context(), principal(r), c)
	}))
}
