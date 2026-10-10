package httpapi

import (
	"net/http"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type repositorySyncCommand struct {
	attentionapp.CommandMeta
	Input string `json:"input"`
}

type repositoryPlacementCommand struct {
	attentionapp.CommandMeta
	SpaceID string `json:"space_id"`
}

func (h *handler) registerRepositories(mux *http.ServeMux) {
	s := h.services.GitHubRepositories
	if s == nil {
		for _, route := range []string{"GET /api/v1/github-repositories", "POST /api/v1/github-repositories/sync", "POST /api/v1/github-repositories/{id}/placement"} {
			mux.HandleFunc(route, h.notImplemented("public_github_repositories"))
		}
		return
	}
	mux.HandleFunc("GET /api/v1/github-repositories", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.ListGitHubRepositories(r.Context(), localCaller(r))
		h.attentionResult(w, r, http.StatusOK, localList(items), err)
	})
	mux.HandleFunc("POST /api/v1/github-repositories/sync", commandHandler(h, func(c *repositorySyncCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c repositorySyncCommand) (any, error) {
		items, err := s.SyncGitHubRepositories(r.Context(), localCaller(r), domain.RepositorySyncCommand{Input: c.Input})
		return localList(items), err
	}))
	mux.HandleFunc("POST /api/v1/github-repositories/{id}/placement", commandHandler(h, func(c *repositoryPlacementCommand) *attentionapp.CommandMeta { return &c.CommandMeta }, http.StatusOK, func(r *http.Request, c repositoryPlacementCommand) (any, error) {
		item, err := s.PlaceGitHubRepository(r.Context(), localCaller(r), r.PathValue("id"), domain.RepositoryPlacementCommand{SpaceID: c.SpaceID, ExpectedRevision: c.ExpectedVersion})
		return localEnvelope("repository", item), err
	}))
}
