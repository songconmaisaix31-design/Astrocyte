package app

import (
	"context"
	"strconv"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type AuthorizedBinding struct {
	ID                 string
	Purpose            string
	Version            int
	PacketID           string
	ProjectID          string
	MemberID           string
	SessionID          string
	NativeSessionID    string
	AdapterID          string
	ControllerID       string
	StartupOperationID string
	WorktreeID         string
	Root               string
}
type authorizedBindingKey struct{}

// 仅由实际 Workspace 授权桥接在同一事务中注入，HTTP 不接受这些事实。
func WithAuthorizedBinding(ctx context.Context, b AuthorizedBinding) context.Context {
	return context.WithValue(ctx, authorizedBindingKey{}, b)
}

func authorizedBinding(ctx context.Context, id, version string) (AuthorizedBinding, error) {
	b, ok := ctx.Value(authorizedBindingKey{}).(AuthorizedBinding)
	if !ok || b.ID != id || b.PacketID == "" || strconv.Itoa(b.Version) != version {
		return b, domain.Failure(apierrors.ContextStale, "current member binding or immutable context version changed")
	}
	return b, nil
}

// 原生桥只消费本次 S2 授权回调中的完整绑定，不接受请求正文中的身份补充。
func AuthorizedBindingForMember(ctx context.Context, p domain.Principal, m Member) (AuthorizedBinding, error) {
	b, err := authorizedBinding(ctx, m.BindingID, m.Version)
	if err != nil {
		return b, err
	}
	if m.Purpose != "claim_work_item" && m.Purpose != "submit_artifact" || b.Purpose != m.Purpose {
		return b, domain.Failure(apierrors.ValidationFailed, "member action purpose is required")
	}
	if p.Kind != "agent" || p.ID == "" || p.ProjectID == "" || p.ID != m.HolderID || p.ID != b.MemberID || p.ProjectID != m.ProjectID || p.ProjectID != b.ProjectID {
		return b, domain.Failure(apierrors.ScopeDenied, "member binding does not belong to the current authenticated project identity")
	}
	if b.SessionID == "" || b.NativeSessionID == "" || b.AdapterID == "" || b.ControllerID == "" || b.StartupOperationID == "" || b.WorktreeID == "" || b.Root == "" || m.SessionID != b.SessionID || m.Worktree != b.Root || m.OperationID == "" {
		return b, domain.Failure(apierrors.ContextStale, "current member binding lacks the exact native session and fixed worktree")
	}
	return b, nil
}

func (s *Service) executionForGrant(ctx context.Context, p domain.Principal, g domain.Grant, member Member) (ExecutionDescriptor, error) {
	var descriptor ExecutionDescriptor
	err := s.grants.WithAuthorizedGrant(ctx, GrantCheck{
		GrantID: g.ID, Epoch: g.Epoch, ProposalRevision: g.ProposalRevision, ProjectID: g.ProjectID,
		Root: g.Root, ExecutionRoot: member.Worktree, Action: "read_swarm", Principal: p,
		BindingIDs: []string{member.BindingID}, BindingPurpose: member.Purpose,
	}, func(ctx context.Context, _ domain.Grant) error {
		var err error
		descriptor, err = s.execution(ctx, p, member)
		return err
	})
	return descriptor, err
}
