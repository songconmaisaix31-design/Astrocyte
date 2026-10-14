package sqlite

import (
	"context"
	"errors"
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

type swarmApprovalReader interface {
	ReadApproval(context.Context, workspace.Caller, string) (workspace.ApprovalGrant, error)
}

func (b *SwarmApprovalBridge) WithMissionRead(ctx context.Context, p swarm.Principal, id string, fn func(context.Context) error) error {
	reader, ok := b.grants.(swarmApprovalReader)
	if !ok {
		return swarm.Failure(apierrors.UnsupportedCapability, "actual Workspace current project reader is unavailable")
	}
	return NewWorkspaceRepository(b.db).WithWorkspaceTx(ctx, func(wtx workspaceapp.WorkspaceTx) error {
		ctx := wtx.Context()
		if _, err := reader.ReadApproval(ctx, workspace.Caller{ID: p.ID, Kind: p.Kind, ProjectID: p.ProjectID}, id); err != nil {
			return err
		}
		conn, err := WorkspaceTransaction(ctx, b.db)
		if err != nil {
			return err
		}
		tx := &swarmTransaction{ctx: ctx, exec: conn, db: b.db}
		tx.ctx = context.WithValue(ctx, swarmTransactionKey{}, tx)
		return fn(tx.ctx)
	})
}

func NewSwarmApprovalBridge(db *DB, grants workspaceapp.GrantAuthorizer) (*SwarmApprovalBridge, error) {
	if db == nil || grants == nil {
		return nil, swarm.Failure(apierrors.ValidationFailed, "real Workspace approval service and same SQLite database required")
	}
	return &SwarmApprovalBridge{db: db, grants: grants}, nil
}

func (b *SwarmApprovalBridge) WithAuthorizedGrant(ctx context.Context, c swarmapp.GrantCheck, fn func(context.Context, swarm.Grant) error) error {
	return NewWorkspaceRepository(b.db).WithWorkspaceTx(ctx, func(wtx workspaceapp.WorkspaceTx) error {
		ctx := wtx.Context()
		conn, err := WorkspaceTransaction(ctx, b.db)
		if err != nil {
			return err
		}
		tx := &swarmTransaction{ctx: ctx, exec: conn, db: b.db}
		tx.ctx = context.WithValue(ctx, swarmTransactionKey{}, tx)
		g, err := b.grants.QueryGrant(ctx, c.GrantID)
		if err != nil {
			return err
		}
		resources := swarm.Resources{}
		if g.Budget.MoneyMinor != nil {
			n := int64(0)
			resources.Money = &n
			resources.Currency = g.Budget.Currency
		}
		m, err := tx.MissionByGrant(c.GrantID)
		if err == nil {
			resources, err = m.Used.Add(m.Reserved)
			if err != nil {
				return err
			}
		} else {
			var missing *apierrors.ServiceError
			if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
				return err
			}
		}
		leases, err := tx.ActiveLeases(c.ProjectID)
		if err != nil {
			return err
		}
		concurrency := len(leases)
		var original *swarm.Lease
		for i := range leases {
			if leases[i].OperationID == c.OperationID {
				original = &leases[i]
				break
			}
		}
		if c.Action == "claim_work_item" {
			if !c.ExecutionKnown || c.OperationID == "" {
				return swarm.Failure(apierrors.EvidenceMissing, "actual native execution policy is required before reservation")
			}
			if original == nil {
				concurrency++
				resources, err = resources.Add(c.Resources)
				if err != nil {
					return err
				}
			}
		}
		if c.Action == "submit_artifact" {
			if !c.ExecutionKnown || original == nil {
				return swarm.Failure(apierrors.LeaseLost, "submission requires the original bound operation")
			}
			resources, err = resources.Sub(original.Reserved)
			if err != nil {
				return err
			}
			resources, err = resources.Add(c.Resources)
			if err != nil {
				return err
			}
		}
		used := workspace.ResourceBudget{Calls: &resources.Calls, Tokens: &resources.Tokens, MoneyMinor: resources.Money, Currency: resources.Currency}
		check := workspace.GrantCheck{Caller: workspace.Caller{ID: c.Principal.ID, Kind: c.Principal.Kind, ProjectID: c.Principal.ProjectID}, GrantID: c.GrantID, Epoch: c.Epoch, ProposalRevision: c.ProposalRevision, ProjectID: c.ProjectID, Root: c.Root, Action: c.Action, Concurrency: concurrency, Used: used, Tools: c.Tools, ExternalTargets: c.ExternalTargets}
		invoke := func(ctx context.Context, g workspace.ApprovalGrant) error {
			budget := swarm.Budget{Calls: g.Budget.Calls, Tokens: g.Budget.Tokens, Money: g.Budget.MoneyMinor, Currency: g.Budget.Currency}
			grant := swarm.Grant{ID: g.ID, Epoch: g.Epoch, ProposalID: g.ProposalID, ProposalRevision: g.ProposalRevision, ProjectID: g.Scope.ProjectID, Root: c.Root, Goal: fmt.Sprintf("%s@%d", g.Goal.ID, g.Goal.Revision), Actions: g.Scope.Actions, WritePaths: g.Scope.WritePaths, Budget: budget, Limits: swarm.Limits{MaxWorkItems: g.Stops.MaxWorkItems, MaxAttempts: g.Stops.MaxAttempts, MaxNoProgress: g.Stops.MaxNoProgressRounds, NoProgressSeconds: g.Stops.NoProgressSeconds, MaxActive: g.Scope.MaxConcurrency, Deadline: g.Stops.Deadline}, ExpiresAt: g.ExpiresAt}
			return fn(context.WithValue(ctx, swarmTransactionKey{}, tx), grant)
		}
		if c.Action == "claim_work_item" || c.Action == "submit_artifact" {
			if len(c.BindingIDs) != 1 || c.BindingIDs[0] == "" {
				return swarm.Failure(apierrors.EvidenceMissing, "one exact current Workspace member binding is required")
			}
			authorizer, ok := b.grants.(workspaceapp.SessionBindingAuthorizer)
			if !ok {
				return swarm.Failure(apierrors.UnsupportedCapability, "actual Workspace member authorization is unavailable")
			}
			return authorizer.WithAuthorizedSessionBindings(tx.ctx, check, c.BindingIDs, func(ctx context.Context, g workspace.ApprovalGrant, bindings []workspace.SessionBinding) error {
				if len(bindings) != 1 || bindings[0].ID != c.BindingIDs[0] || bindings[0].MemberID != c.Principal.ID {
					return swarm.Failure(apierrors.ScopeDenied, "binding belongs to another authenticated member")
				}
				return invoke(ctx, g)
			})
		}
		return b.grants.WithAuthorizedGrant(tx.ctx, check, invoke)
	})
}
