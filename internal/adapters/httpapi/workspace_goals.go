package httpapi

import (
	"net/http"
	"strconv"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func (h *handler) registerWorkspaceGoals(mux *http.ServeMux, s *workspace.S2Service) {
	mux.HandleFunc("GET /api/v1/goals", func(w http.ResponseWriter, r *http.Request) {
		goals, err := s.ListGoalsFor(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localList(goals), err)
	})
	for _, pattern := range []string{"POST /api/v1/goals", "PUT /api/v1/goals/{id}"} {
		mux.HandleFunc(pattern, workspaceCommand(h, func(c *workspace.GoalCommand) *workspace.WorkspaceCommandMeta { return &c.WorkspaceCommandMeta }, func(r *http.Request, c workspace.GoalCommand) (any, error) {
			g, err := s.PutGoal(r.Context(), localCaller(r), r.PathValue("id"), c)
			return localEnvelope("goal", g), err
		}))
	}
	mux.HandleFunc("GET /api/v1/goals/{id}", func(w http.ResponseWriter, r *http.Request) {
		revision := 0
		var err error
		if raw := r.URL.Query().Get("revision"); raw != "" {
			revision, err = strconv.Atoi(raw)
			if err != nil || revision < 1 {
				err = invalid("revision must be positive")
			}
		}
		var g domain.Goal
		if err == nil {
			g, err = s.ReadGoal(r.Context(), localCaller(r), r.PathValue("id"), revision)
		}
		h.attentionResult(w, r, http.StatusOK, localEnvelope("goal", g), err)
	})
}
