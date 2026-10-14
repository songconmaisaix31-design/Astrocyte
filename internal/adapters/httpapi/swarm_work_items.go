package httpapi

import (
	"net/http"

	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func (h *handler) registerSwarmWorkItems(mux *http.ServeMux, s *swarm.Service) {
	mux.HandleFunc("GET /api/v1/artifacts/{id}/diff", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("mission_id")
		var err error
		var out swarm.ArtifactDiff
		if id == "" {
			err = invalid("mission_id is required")
		} else {
			out, err = s.ReadArtifactDiff(r.Context(), swarmCaller(r), id, r.PathValue("id"))
		}
		h.swarmResult(w, r, out, err)
	})
	mux.HandleFunc("POST /api/v1/work-items/{id}/claims", swarmCommand(h, func(c *swarm.ClaimCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.ClaimCommand) (any, error) {
		if c.WorkItemID != "" && c.WorkItemID != r.PathValue("id") {
			return nil, invalid("work_item_id differs from route")
		}
		c.WorkItemID = r.PathValue("id")
		l, err := s.ClaimWorkItem(r.Context(), swarmCaller(r), c)
		return localEnvelope("lease", l), err
	}))
	mux.HandleFunc("POST /api/v1/work-items/{id}/artifacts", swarmCommand(h, func(c *swarm.SubmitCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.SubmitCommand) (any, error) {
		if c.WorkItemID != "" && c.WorkItemID != r.PathValue("id") {
			return nil, invalid("work_item_id differs from route")
		}
		c.WorkItemID = r.PathValue("id")
		a, err := s.SubmitArtifact(r.Context(), swarmCaller(r), c)
		return localEnvelope("artifact", a), err
	}))
	mux.HandleFunc("POST /api/v1/artifacts/{id}/acceptance", swarmCommand(h, func(c *swarm.AcceptCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.AcceptCommand) (any, error) {
		if c.ArtifactID != "" && c.ArtifactID != r.PathValue("id") {
			return nil, invalid("artifact_id differs from route")
		}
		c.ArtifactID = r.PathValue("id")
		a, err := s.AcceptArtifact(r.Context(), swarmCaller(r), c)
		return localEnvelope("artifact", a), err
	}))
	mux.HandleFunc("POST /api/v1/artifacts/{id}/rejection", swarmCommand(h, func(c *swarm.RejectCommand) *domain.Meta { return &c.Meta }, func(r *http.Request, c swarm.RejectCommand) (any, error) {
		if c.ArtifactID != "" && c.ArtifactID != r.PathValue("id") {
			return nil, invalid("artifact_id differs from route")
		}
		c.ArtifactID = r.PathValue("id")
		out, err := s.RejectArtifact(r.Context(), swarmCaller(r), c)
		return localEnvelope("artifact", out), err
	}))
}
