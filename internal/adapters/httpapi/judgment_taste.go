package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func judgmentTasteHuman(r *http.Request) error {
	caller := localCaller(r)
	if caller.ID == "" || caller.Kind != "human" {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Complete taste configuration requires the authenticated human session"}
	}
	return nil
}

func (h *handler) registerJudgmentTaste(mux *http.ServeMux, services JudgmentServices) {
	const path = "/api/v1/attention-taste-profile"
	mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		err := judgmentTasteHuman(r)
		var result attention.TasteProfileDetail
		if err == nil {
			if services.Taste == nil {
				err = judgmentUnavailable()
			} else {
				result, err = services.Taste.GetTasteProfile(r.Context(), principal(r))
			}
		}
		h.judgmentResult(w, r, result, err)
	})
	put := commandHandler(h, func(c *attention.UpdateTasteProfileCommand) *attention.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c attention.UpdateTasteProfileCommand) (any, error) {
		if err := judgmentTasteHuman(r); err != nil {
			return nil, err
		}
		if services.Taste == nil {
			return nil, judgmentUnavailable()
		}
		return services.Taste.UpdateTasteProfile(r.Context(), principal(r), c)
	})
	mux.HandleFunc("PUT "+path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		put(w, r)
	})
}
