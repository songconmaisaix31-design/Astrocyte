package app

import (
	"context"
	"strconv"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type AuthorizedBinding struct {
	ID       string
	Version  int
	PacketID string
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
