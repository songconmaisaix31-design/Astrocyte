package httpapi

import (
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func (h *handler) registerOpenViking(mux *http.ServeMux) {
	s := h.services.OpenViking
	unavailable := func(w http.ResponseWriter, r *http.Request) bool {
		if s != nil {
			return false
		}
		h.attentionResult(w, r, http.StatusServiceUnavailable, nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "OpenViking 未配置", RequiredAction: "configure_openviking_on_server"})
		return true
	}
	mux.HandleFunc("GET /api/v1/openviking/status", func(w http.ResponseWriter, r *http.Request) {
		if s == nil {
			h.attentionResult(w, r, http.StatusOK, attentionapp.OpenVikingStatus{SchemaVersion: 1, Message: "OpenViking 未配置"}, nil)
			return
		}
		result, err := s.Status(r.Context(), principal(r))
		h.attentionResult(w, r, http.StatusOK, result, err)
	})
	mux.HandleFunc("GET /api/v1/project-spaces/{id}/openviking/submissions", func(w http.ResponseWriter, r *http.Request) {
		if unavailable(w, r) {
			return
		}
		items, err := s.List(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, map[string]any{"schema_version": 1, "items": items}, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces/{id}/openviking/submissions", func(w http.ResponseWriter, r *http.Request) {
		if unavailable(w, r) {
			return
		}
		c, err := decodeAttention[attentionapp.OpenVikingSubmitCommand](w, r)
		if err != nil {
			h.attentionResult(w, r, 0, nil, err)
			return
		}
		c.IdempotencyKey = r.Header.Get("Idempotency-Key")
		if s == nil {
			return
		}
		row, err := s.Submit(r.Context(), principal(r), r.PathValue("id"), c)
		h.attentionResult(w, r, http.StatusAccepted, map[string]any{"schema_version": 1, "submission": row}, err)
	})
	mux.HandleFunc("POST /api/v1/openviking/submissions/{id}/refresh", func(w http.ResponseWriter, r *http.Request) {
		if unavailable(w, r) {
			return
		}
		row, task, err := s.Refresh(r.Context(), principal(r), r.PathValue("id"))
		h.attentionResult(w, r, http.StatusOK, map[string]any{"schema_version": 1, "submission": row, "task": task}, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces/{id}/openviking/search", func(w http.ResponseWriter, r *http.Request) {
		if unavailable(w, r) {
			return
		}
		body, err := decodeAttention[attentionapp.OpenVikingSearchCommand](w, r)
		if err != nil {
			h.attentionResult(w, r, 0, nil, err)
			return
		}
		items, err := s.Search(r.Context(), principal(r), r.PathValue("id"), body)
		h.attentionResult(w, r, http.StatusOK, map[string]any{"schema_version": 1, "items": items}, err)
	})
	mux.HandleFunc("POST /api/v1/project-spaces/{id}/openviking/content", func(w http.ResponseWriter, r *http.Request) {
		if unavailable(w, r) {
			return
		}
		body, err := decodeAttention[attentionapp.OpenVikingContentCommand](w, r)
		if err != nil {
			h.attentionResult(w, r, 0, nil, err)
			return
		}
		content, err := s.Content(r.Context(), principal(r), r.PathValue("id"), body)
		h.attentionResult(w, r, http.StatusOK, map[string]any{"schema_version": 1, "content": content}, err)
	})
}
