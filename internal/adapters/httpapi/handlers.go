package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// handler holds the application services and implements HTTP handlers.
type handler struct {
	services Services
	logger   any
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// writeError writes a structured error response in the ErrorV1 envelope:
// {"schema_version": 1, "error": {...}}
func writeError(w http.ResponseWriter, r *http.Request, status int, svcErr *apierrors.ServiceError) {
	if svcErr.RequestID == "" {
		if rid, ok := r.Context().Value(requestIDKey).(string); ok {
			svcErr.RequestID = rid
		}
	}
	writeJSON(w, status, apierrors.Wrap(svcErr))
}

// handleHealth responds to GET /api/v1/health.
func (h *handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := h.services.Foundation.Health(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, &apierrors.ServiceError{
			Code:      apierrors.InternalError,
			Message:   err.Error(),
			Retryable: true,
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleFoundation responds to GET /api/v1/foundation.
func (h *handler) handleFoundation(w http.ResponseWriter, r *http.Request) {
	resp, err := h.services.Foundation.Foundation(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, &apierrors.ServiceError{
			Code:      apierrors.InternalError,
			Message:   err.Error(),
			Retryable: true,
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleListMaterials responds to GET /api/v1/materials.
func (h *handler) handleListMaterials(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Materials.ListMaterials(r.Context())
	h.handleList(w, r, result, err)
}

// handleListOpportunities responds to GET /api/v1/opportunities.
func (h *handler) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Opportunities.ListOpportunities(r.Context())
	h.handleList(w, r, result, err)
}

// handleListProjects responds to GET /api/v1/projects.
func (h *handler) handleListProjects(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Projects.ListProjects(r.Context())
	h.handleList(w, r, result, err)
}

// handleListProposals responds to GET /api/v1/proposals.
func (h *handler) handleListProposals(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Proposals.ListProposals(r.Context())
	h.handleList(w, r, result, err)
}

// handleListSessions responds to GET /api/v1/sessions.
func (h *handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Sessions.ListSessions(r.Context())
	h.handleList(w, r, result, err)
}

// handleListMissions responds to GET /api/v1/missions.
func (h *handler) handleListMissions(w http.ResponseWriter, r *http.Request) {
	result, err := h.services.Missions.ListMissions(r.Context())
	h.handleList(w, r, result, err)
}

// handleGetMission responds to GET /api/v1/missions/{id}.
func (h *handler) handleGetMission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := h.services.Missions.GetMission(r.Context(), id)
	if err != nil {
		if svcErr, ok := err.(*apierrors.ServiceError); ok {
			switch svcErr.Code {
			case apierrors.NotFound:
				writeError(w, r, http.StatusNotFound, svcErr)
			default:
				writeError(w, r, http.StatusInternalServerError, svcErr)
			}
			return
		}
		writeError(w, r, http.StatusInternalServerError, &apierrors.ServiceError{
			Code:      apierrors.InternalError,
			Message:   err.Error(),
			Retryable: true,
		})
		return
	}
}

// handleList is a helper for list endpoints.
func (h *handler) handleList(w http.ResponseWriter, r *http.Request, result apierrors.ListResult, err error) {
	if err != nil {
		if svcErr, ok := err.(*apierrors.ServiceError); ok {
			writeError(w, r, http.StatusInternalServerError, svcErr)
			return
		}
		writeError(w, r, http.StatusInternalServerError, &apierrors.ServiceError{
			Code:      apierrors.InternalError,
			Message:   err.Error(),
			Retryable: true,
		})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// notImplemented returns a handler that always responds with 501
// unsupported_capability for the named feature.
func (h *handler) notImplemented(capability string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotImplemented, apierrors.NewUnsupported(capability))
	}
}

// handleAPINotFound responds to unmatched /api/v1/* routes with JSON 404.
func (h *handler) handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusNotFound, &apierrors.ServiceError{
		Code:           apierrors.NotFound,
		Message:        "API route not found",
		Retryable:      false,
		RequiredAction: "check_path_and_method",
	})
}
