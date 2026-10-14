package sqlite

import (
	"context"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type SwarmApprovalBridge struct {
	db     *DB
	grants workspaceapp.GrantAuthorizer
}

func NewSwarmApprovalBridge(db *DB, grants workspaceapp.GrantAuthorizer) (*SwarmApprovalBridge, error) {
	if db == nil || grants == nil {
		return nil, swarm.Failure(apierrors.ValidationFailed, "real Workspace approval service and same SQLite database required")
	}
	return &SwarmApprovalBridge{db: db, grants: grants}, nil
}

func (b *SwarmApprovalBridge) WithAuthorizedGrant(ctx context.Context, c swarmapp.GrantCheck, fn func(context.Context, swarm.Grant) error) error {
	used := workspace.ResourceBudget{Calls: &c.Resources.Calls, Tokens: &c.Resources.Tokens, MoneyMinor: c.Resources.Money, Currency: c.Resources.Currency}
	check := workspace.GrantCheck{GrantID: c.GrantID, Epoch: c.Epoch, ProposalRevision: c.ProposalRevision, ProjectID: c.ProjectID, Root: c.Root, Action: c.Action, Concurrency: 1, Used: used}
	return b.grants.WithAuthorizedGrant(ctx, check, func(ctx context.Context, g workspace.ApprovalGrant) error {
		conn, err := WorkspaceTransaction(ctx, b.db)
		if err != nil {
			return err
		}
		// 只借用公开端口提供的同一连接，不读取 Workspace 内部表。
		tx := &swarmTransaction{ctx: ctx, exec: conn, db: b.db}
		tx.ctx = context.WithValue(ctx, swarmTransactionKey{}, tx)
		budget := swarm.Budget{Calls: g.Budget.Calls, Tokens: g.Budget.Tokens, Money: g.Budget.MoneyMinor, Currency: g.Budget.Currency}
		grant := swarm.Grant{ID: g.ID, Epoch: g.Epoch, ProposalID: g.ProposalID, ProposalRevision: g.ProposalRevision, ProjectID: g.Scope.ProjectID, Root: c.Root, Goal: fmt.Sprintf("%s@%d", g.Goal.ID, g.Goal.Revision), Actions: g.Scope.Actions, WritePaths: []string{"."}, Budget: budget, Limits: swarm.Limits{MaxWorkItems: g.Stops.MaxWorkItems, MaxAttempts: g.Stops.MaxAttempts, NoProgressSeconds: g.Stops.NoProgressSeconds, MaxActive: g.Scope.MaxConcurrency, Deadline: g.Stops.Deadline}, ExpiresAt: g.ExpiresAt}
		return fn(tx.ctx, grant)
	})
}
