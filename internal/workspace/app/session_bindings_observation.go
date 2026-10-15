package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type ObservedSessionBindingAuthorizer interface {
	ObserveSessionBindings(context.Context, domain.GrantCheck, []string) (*SessionBindingObservation, error)
	WithObservedSessionBindings(context.Context, domain.GrantCheck, *SessionBindingObservation, func(context.Context, domain.ApprovalGrant, []domain.SessionBinding) error) error
}

type WorkspaceObservationGuard interface {
	RequireWorkspaceObservation(context.Context) error
}

// 观察只用于原请求的最终复核，不能替代当前授权，也不进入HTTP或持久缓存。
type SessionBindingObservation struct {
	service *S2Service
	request context.Context
	check   domain.GrantCheck
	entries []sessionBindingObservationEntry
}

type sessionBindingObservationEntry struct {
	binding domain.SessionBinding
	packet  domain.ImmutableContextPacket
	facts   domain.OwnedSessionFacts
}
