package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type repositoryProbe struct {
	reads, syncs, placements int
	caller                   domain.Caller
	placement                domain.RepositoryPlacementCommand
}

func (p *repositoryProbe) ListGitHubRepositories(_ context.Context, c domain.Caller) ([]domain.GitHubRepository, error) {
	p.reads++
	p.caller = c
	return []domain.GitHubRepository{}, nil
}
func (p *repositoryProbe) SyncGitHubRepositories(_ context.Context, c domain.Caller, _ domain.RepositorySyncCommand) ([]domain.GitHubRepository, error) {
	p.syncs++
	p.caller = c
	return []domain.GitHubRepository{}, nil
}
func (p *repositoryProbe) PlaceGitHubRepository(_ context.Context, c domain.Caller, _ string, cmd domain.RepositoryPlacementCommand) (domain.GitHubRepository, error) {
	p.placements++
	p.caller = c
	p.placement = cmd
	return domain.GitHubRepository{}, nil
}

func TestRepositoryTransportHumanCommandsAndCacheOnlyGET(t *testing.T) {
	p := &repositoryProbe{}
	s := NewServer(Config{Services: Services{GitHubRepositories: p}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	path := "/api/v1/github-repositories"
	for _, auth := range []string{"", "Bearer unscoped-agent"} {
		w := attentionRequest(s, "GET", path, "", nil, "", auth, "")
		if w.Code != http.StatusForbidden || p.reads != 0 {
			t.Fatalf("unauthorized repository cache read: %d", w.Code)
		}
	}
	cookie, csrf := bootstrapSession(t, s)
	for range 2 {
		w := attentionRequest(s, "GET", path, "", cookie, "", "", "")
		if w.Code != http.StatusOK || p.syncs != 0 || p.placements != 0 {
			t.Fatalf("cache GET caused effect: %d %+v", w.Code, p)
		}
	}
	body := `{"schema_version":1,"request_id":"place","expected_version":7,"space_id":"top-space"}`
	w := attentionRequest(s, "POST", path+"/github-1/placement", body, cookie, "", "", "")
	if w.Code != http.StatusForbidden || p.placements != 0 {
		t.Fatalf("placement without CSRF: %d", w.Code)
	}
	w = attentionRequest(s, "POST", path+"/github-1/placement", body, cookie, csrf, "", "")
	if w.Code != http.StatusOK || p.placements != 1 || p.placement.ExpectedRevision != 7 || p.placement.SpaceID != "top-space" || p.caller.Kind != "human" || p.caller.OperationID != "caller-request-key" {
		t.Fatalf("placement identity/revision mismatch: %d %+v", w.Code, p)
	}
	w = attentionRequest(s, "POST", path+"/sync", `{"schema_version":1,"request_id":"sync","expected_version":1,"input":"example/repository","actor_kind":"human"}`, cookie, csrf, "", "")
	if w.Code != http.StatusBadRequest || p.syncs != 0 {
		t.Fatalf("body supplied trusted identity: %d", w.Code)
	}
}
