package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestConcurrentTokenIssuanceCannotResurrectRevokedGrant(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	db := openAttentionDB(t, filepath.Join(acceptedExperienceDirectory(t), "state.sqlite"))
	service := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	human := domain.Caller{Kind: "human", ID: uuid.NewString()}
	space, err := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{}).CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "Actual token revocation race",
	})
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	p, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Current public checkout token race", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GrantProjectAgent(ctx, human, p.ID, domain.ProjectGrant{AgentID: "worker", Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	// 真实事务占用生产单连接，使令牌签发持有服务锁并等待读取原 Grant。
	tx, err := db.Conn().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	released := false
	t.Cleanup(func() {
		if !released {
			if err := tx.Rollback(); err != nil {
				t.Error(err)
			}
		}
	})
	priorWaits := db.Conn().Stats().WaitCount
	tokenResult := make(chan domain.AgentToken, 1)
	issueErr := make(chan error, 1)
	go func() {
		token, err := service.IssueProjectAgentToken(ctx, human, p.ID, "worker")
		tokenResult <- token
		issueErr <- err
	}()
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for db.Conn().Stats().WaitCount == priorWaits {
		select {
		case err := <-issueErr:
			t.Fatal("issuance ended before waiting for the actual SQLite connection", err)
		case <-deadline.C:
			t.Fatal("issuance never reached persistent storage")
		case <-tick.C:
		}
	}
	revoked := make(chan error, 1)
	go func() { _, err := service.RevokeProjectAgent(ctx, human, p.ID, "worker"); revoked <- err }()
	select {
	case err := <-revoked:
		t.Fatalf("revoke raced the older token write: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	if db.Conn().Stats().WaitCount != priorWaits+1 {
		t.Fatal("revocation bypassed the original issuance lock and waited for SQLite")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	released = true
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
