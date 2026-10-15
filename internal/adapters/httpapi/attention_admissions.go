package httpapi

import (
	"net/http"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerAdmissions(mux *http.ServeMux, service attention.OpportunityAdmissions) {
	if service == nil {
		mux.HandleFunc("POST /api/v1/opportunities/{id}/admissions", h.notImplemented("opportunity_admission"))
		mux.HandleFunc("GET /api/v1/admissions/{id}", h.notImplemented("opportunity_admission"))
		mux.HandleFunc("POST /api/v1/admissions/{id}/revocations", h.notImplemented("opportunity_admission"))
		return
	}
	mux.HandleFunc("POST /api/v1/opportunities/{id}/admissions", commandHandler(h, func(c *attention.AdmissionCommand) *attention.CommandMeta { return &c.CommandMeta }, http.StatusCreated, func(r *http.Request, command attention.AdmissionCommand) (any, error) {
		return service.AdmitOpportunity(r.Context(), principal(r), r.PathValue("id"), command)
	}))
	mux.HandleFunc("GET /api/v1/admissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := service.GetAdmission(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("POST /api/v1/admissions/{id}/revocations", commandHandler(h, func(c *attention.RevokeAdmissionCommand) *attention.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, command attention.RevokeAdmissionCommand) (any, error) {
		return service.RevokeAdmission(r.Context(), principal(r), r.PathValue("id"), command)
	}))
}
