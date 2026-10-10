package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// The human revision is the metadata CAS revision, independently of the common
// command envelope. Timestamp, identity, grants and progress are output/domain
// facts and are deliberately absent from the strict input DTO.
type registeredMetadataCommand struct {
	attentionapp.CommandMeta
	Notes    *string `json:"notes"`
	Review   *string `json:"review"`
	Group    *string `json:"group"`
	Intent   *string `json:"intent"`
	Archived *bool   `json:"archived"`
	Revision *int    `json:"revision"`
}

func (h *handler) registerRegisteredProjects(mux *http.ServeMux) {
	s := h.services.RegisteredProjects
	metadata, ok := s.(workspaceapp.RegisteredProjectMetadata)
	if !ok {
		mux.HandleFunc("PUT /api/v1/local-projects/registered/{project_id}/metadata", h.notImplemented("registered_project_metadata"))
	} else {
		mux.HandleFunc("PUT /api/v1/local-projects/registered/{project_id}/metadata", commandHandler(h, func(c *registeredMetadataCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c registeredMetadataCommand) (any, error) {
			if c.Notes == nil || c.Review == nil || c.Group == nil || c.Intent == nil || c.Archived == nil || c.Revision == nil || *c.Revision < 0 {
				return nil, invalid("all manual metadata fields and a non-negative revision are required")
			}
			project, err := metadata.SetRegisteredProjectMetadata(r.Context(), localCaller(r), r.PathValue("project_id"), domain.ProjectHumanMetadata{
				Notes: *c.Notes, Review: *c.Review, Group: *c.Group, Intent: *c.Intent, Archived: *c.Archived, Revision: *c.Revision,
			})
			return localEnvelope("project", project), err
		}))
	}
	if s == nil {
		mux.HandleFunc("GET /api/v1/local-projects/registered", h.notImplemented("registered_project_discovery"))
		mux.HandleFunc("POST /api/v1/local-projects/registered/refresh", h.notImplemented("registered_project_discovery"))
		return
	}
	mux.HandleFunc("GET /api/v1/local-projects/registered", func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := s.ListRegisteredProjects(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localEnvelope("snapshot", snapshot), err)
	})
	mux.HandleFunc("POST /api/v1/local-projects/registered/refresh", commandHandler(h, func(c *attentionapp.CommandMeta) *attentionapp.CommandMeta { return c }, http.StatusOK, func(r *http.Request, _ attentionapp.CommandMeta) (any, error) {
		snapshot, err := s.RefreshRegisteredProjects(r.Context(), localCaller(r))
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "registered project refresh exceeded its deadline", Retryable: true, RequiredAction: "retry_registered_project_refresh"}
		}
		return localEnvelope("snapshot", snapshot), err
	}))
}
