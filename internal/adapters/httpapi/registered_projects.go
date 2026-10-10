package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerRegisteredProjects(mux *http.ServeMux) {
	s := h.services.RegisteredProjects
	if s == nil {
		mux.HandleFunc("GET /api/v1/local-projects/registered", h.notImplemented("registered_project_discovery"))
		mux.HandleFunc("POST /api/v1/local-projects/registered/refresh", h.notImplemented("registered_project_discovery"))
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects/registered", func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := s.ListRegisteredProjects(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("snapshot", snapshot), err)
	})
	refresh := commandHandler(h, func(c *attentionapp.CommandMeta) *attentionapp.CommandMeta { return c }, http.StatusOK, func(r *http.Request, _ attentionapp.CommandMeta) (any, error) {
		snapshot, err := s.RefreshRegisteredProjects(r.Context(), localCaller(r))
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "registered project refresh exceeded its deadline", Retryable: true, RequiredAction: "retry_registered_project_refresh"}
		}
		return localEnvelope("snapshot", snapshot), err
	})
	mux.HandleFunc("POST /api/v1/local-projects/registered/refresh", func(w http.ResponseWriter, r *http.Request) {
		// Bound the entire request, including a wait for another refresh. The
		// adapter's own deadline alone starts only after that wait completes.
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		refresh(w, r.WithContext(ctx))
	})
}
