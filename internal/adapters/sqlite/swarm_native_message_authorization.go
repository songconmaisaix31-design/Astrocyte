package sqlite

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarmdomain "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type SwarmNativeMessageAuthorizer struct {
	db      *DB
	grants  *SwarmApprovalBridge
	service *swarm.Service
}

var _ workspace.NativeMessageAuthorizer = (*SwarmNativeMessageAuthorizer)(nil)

func NewSwarmNativeMessageAuthorizer(db *DB, grants *SwarmApprovalBridge, service *swarm.Service) (*SwarmNativeMessageAuthorizer, error) {
	if db == nil || grants == nil || grants.db != db || service == nil {
		return nil, errors.New("native send authorization requires the same actual database, approval bridge and Swarm service")
	}
	return &SwarmNativeMessageAuthorizer{db: db, grants: grants, service: service}, nil
}

func swarmPreparedNativeInput(prepared workspace.NativeMessagePreparation) swarm.PreparedMemberInput {
	i := prepared.Input
	return swarm.PreparedMemberInput{OperationID: i.OperationID, SessionID: i.Session.ID, NativeSessionID: i.Session.NativeID,
		AdapterID: i.Session.CLI, AdapterVersion: i.Session.Version, ControllerID: i.ControllerID, StartupOperationID: i.StartupOperationID,
		ProjectID: i.Session.ProjectID, Root: i.WorkspaceRoot, ActorID: i.ActorID, ActorKind: i.ActorKind, ActorProjectID: i.ActorProjectID, PreparedAt: i.PreparedAt}
}

func (a *SwarmNativeMessageAuthorizer) validateRequest(ctx context.Context, request workspace.NativeMessageAuthorization, fn func(context.Context) error) error {
	if err := a.db.RequireNativeMessageOutsideTx(ctx); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("native send authorization callback is required")
	}
	if err := workspace.ValidateNativeMessageScope(request.Preparation.Scope); err != nil {
		return err
	}
	i, c := request.Preparation.Input, request.Caller
	if c.ID == "" || c.Kind != "human" && c.Kind != "agent" || i.ActorID != c.ID || i.ActorKind != c.Kind || i.ActorProjectID != c.ProjectID || i.OperationID == "" || i.OperationID != c.OperationID || i.Session.ID == "" || i.Session.ProjectID == "" {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "native input authorization requires its original authenticated operation"}
	}
	return nil
}

func (a *SwarmNativeMessageAuthorizer) withMemberDeadline(ctx context.Context, c domain.Caller, fn func(context.Context) error) error {
	if c.Kind != "agent" {
		return fn(ctx)
	}
	g, err := a.db.LoadGrant(ctx, c.ProjectID, c.ID)
	if err != nil {
		return err
	}
	if g.RevokedAt != nil || g.ExpiresAt != nil && !time.Now().Before(*g.ExpiresAt) {
		return &apierrors.ServiceError{Code: apierrors.ApprovalRevoked, Message: "original member project grant is inactive"}
	}
	if g.ExpiresAt == nil {
		return fn(ctx)
	}
	bounded, cancel := context.WithDeadline(ctx, *g.ExpiresAt)
	defer cancel()
	return fn(bounded)
}

func (a *SwarmNativeMessageAuthorizer) withProject(ctx context.Context, request workspace.NativeMessageAuthorization, fn func(context.Context) error) error {
	if request.Preparation.Lease != (workspace.NativeMessageLease{}) {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "project input cannot carry a Swarm execution lease"}
	}
	return a.db.WithNativeMessageTx(ctx, func(ctx context.Context) error {
		if err := a.db.requireProjectNativeMessageScope(ctx, request.Preparation.Input.Session.ID); err != nil {
			return err
		}
		return a.withMemberDeadline(ctx, request.Caller, fn)
	})
}

func nativeMessageAuthorizationError(err error) error {
	var denied *swarmdomain.RuleError
	if errors.As(err, &denied) {
		return &apierrors.ServiceError{Code: apierrors.Code(denied.Code), Message: denied.Message, RequiredAction: "review_project_permissions_and_native_state"}
	}
	return workspace.NormalizeWorkspaceError(err)
}

func (a *SwarmNativeMessageAuthorizer) WithNativeMessagePreparation(ctx context.Context, request workspace.NativeMessageAuthorization, fn func(context.Context) error) (err error) {
	defer func() { err = nativeMessageAuthorizationError(err) }()
	if err := a.validateRequest(ctx, request, fn); err != nil {
		return err
	}
	if request.Preparation.Scope.Kind == "project" {
		return a.withProject(ctx, request, fn)
	}
	input := swarmPreparedNativeInput(request.Preparation)
	p := swarmdomain.Principal{ID: input.ActorID, Kind: input.ActorKind, ProjectID: input.ActorProjectID}
	var binding domain.SessionBinding
	var grant domain.ApprovalGrant
	if err := NewWorkspaceRepository(a.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		var err error
		binding, err = tx.LoadSessionBinding(request.Preparation.Scope.BindingID)
		if err != nil {
			return err
		}
		grant, err = tx.LoadApproval(binding.GrantID)
		return err
	}); err != nil {
		return err
	}
	member := swarm.Member{Purpose: "claim_work_item", BindingID: binding.ID, Version: strconv.Itoa(binding.Version), HolderID: p.ID, ProjectID: input.ProjectID, SessionID: input.SessionID, Worktree: input.Root, OperationID: input.OperationID}
	return a.grants.WithAuthorizedGrant(ctx, swarm.GrantCheck{GrantID: grant.ID, Epoch: binding.GrantEpoch, ProposalRevision: grant.ProposalRevision, ProjectID: input.ProjectID, Root: input.Root, ExecutionRoot: input.Root, Action: "read_swarm", BindingPurpose: member.Purpose, BindingIDs: []string{binding.ID}, Principal: p}, func(ctx context.Context, _ swarmdomain.Grant) error {
		if err := swarm.ValidatePreparedMemberInput(ctx, p, member, input); err != nil {
			return err
		}
		return a.withMemberDeadline(ctx, request.Caller, fn)
	})
}

func (a *SwarmNativeMessageAuthorizer) WithNativeMessageSend(ctx context.Context, request workspace.NativeMessageAuthorization, fn func(context.Context) error) (err error) {
	defer func() { err = nativeMessageAuthorizationError(err) }()
	if err := a.validateRequest(ctx, request, fn); err != nil {
		return err
	}
	if request.Preparation.Scope.Kind == "project" {
		return a.withProject(ctx, request, fn)
	}
	p := request.Preparation
	return a.service.WithAuthorizedNativeSend(ctx, swarm.NativeSendCheck{Input: swarmPreparedNativeInput(p), BindingID: p.Scope.BindingID, WorkItemID: p.Lease.WorkItemID, MissionID: p.Lease.MissionID, Fence: p.Lease.Fence}, func(ctx context.Context) error {
		return a.withMemberDeadline(ctx, request.Caller, fn)
	})
}
