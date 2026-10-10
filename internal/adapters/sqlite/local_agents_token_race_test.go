package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type pausedTokenRepository struct {
	workspace.LocalProjectRepository
	entered, release chan struct{}
}

func (r pausedTokenRepository) SaveGrant(ctx context.Context, g domain.ProjectGrant) error {
	if g.TokenDigest != "" {
		close(r.entered)
		<-r.release
	}
	return r.LocalProjectRepository.SaveGrant(ctx, g)
}

func TestConcurrentTokenIssuanceCannotResurrectRevokedGrant(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	repo := pausedTokenRepository{LocalProjectRepository: db, entered: make(chan struct{}), release: make(chan struct{})}
	service := workspace.NewLocalProjectService(repo, &localReferences{}, agents.ProjectFiles{}, localRegistry{&localNative{}})
	human := domain.Caller{Kind: "human", ID: "human"}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "revocation race", Root: t.TempDir(), SpaceID: "space"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GrantProjectAgent(ctx, human, p.ID, domain.ProjectGrant{AgentID: "worker", Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	tokenResult := make(chan domain.AgentToken, 1)
	issueErr := make(chan error, 1)
	go func() {
		token, err := service.IssueProjectAgentToken(ctx, human, p.ID, "worker")
		tokenResult <- token
		issueErr <- err
	}()
	select {
	case <-repo.entered:
	case <-time.After(time.Second):
		t.Fatal("issuance never reached persistent write")
	}
	revoked := make(chan error, 1)
	go func() { _, err := service.RevokeProjectAgent(ctx, human, p.ID, "worker"); revoked <- err }()
	select {
	case err := <-revoked:
		close(repo.release)
		t.Fatalf("revoke raced the older token write: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(repo.release)
	if err := <-issueErr; err != nil {
		t.Fatal(err)
	}
	if err := <-revoked; err != nil {
		t.Fatal(err)
	}
	issued := <-tokenResult
	if _, err := service.AuthenticateAgentToken(ctx, issued.Token); err == nil {
		t.Fatal("concurrent issuance resurrected revoked token")
	}
	grant, err := db.LoadGrant(ctx, p.ID, "worker")
	if err != nil || grant.RevokedAt == nil || grant.TokenDigest != "" {
		t.Fatal("durable grant revocation was overwritten", err)
	}
}
