package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func TestSwarmActualMemberPreflightRejectsEnclosingWriteTransactions(t *testing.T) {
	db := swarmTestDB(t, filepath.Join(swarmTestDirectory(t), "state.sqlite"))
	db.conn.SetMaxOpenConns(1)
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	s2 := workspaceapp.NewS2Service(NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	bridge, err := NewSwarmApprovalBridge(db, s2)
	if err != nil {
		t.Fatal(err)
	}
	check := swarmapp.GrantCheck{Action: "read_swarm", BindingPurpose: "claim_work_item", BindingIDs: []string{uuid.NewString()}}
	for _, kind := range []string{"workspace", "swarm"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			invoke := func(ctx context.Context) error {
				return bridge.WithAuthorizedGrant(ctx, check, func(context.Context, swarm.Grant) error {
					t.Fatal("member authorization callback ran inside an enclosing write transaction")
					return nil
				})
			}
			var err error
			if kind == "workspace" {
				err = NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspaceapp.WorkspaceTx) error {
					return invoke(tx.Context())
				})
			} else {
				err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
					return invoke(tx.(*swarmTransaction).ctx)
				})
			}
			var rule *swarm.RuleError
			if !errors.As(err, &rule) || rule.Code != string(apierrors.VersionConflict) {
				t.Fatalf("member observation must fail at the enclosing transaction boundary: %v", err)
			}
			if ctx.Err() != nil {
				t.Fatal("member preflight waited for the connection held by its caller")
			}
		})
	}
}
