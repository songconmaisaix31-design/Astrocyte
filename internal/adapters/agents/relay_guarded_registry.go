package agents

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 受信组装选择固定保护；HTTP 不能设置回调、原生来源或进程参数。
func NewRelayGuardedRegistry(stateRoot string, guard RelayBeforeFirstInput) (*Registry, error) {
	if guard == nil {
		return nil, nativeError(apierrors.ValidationFailed, "native source checkpoint is required")
	}
	registry := NewRegistry(stateRoot)
	registry.adapters["codex"].(*Native).relayInputGuard = guard
	return registry, nil
}

// 只使用原控制器内存中的固定正文；数据库重载不能补造已失去的输入上下文。
func (p *nativeProcess) relayInputSession(session domain.NativeSession) (domain.NativeSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.originalSession == nil {
		return domain.NativeSession{}, nativeError(apierrors.DeliveryUnknown, "original input context is unavailable")
	}
	original := p.originalSession
	if original.ID != session.ID || original.NativeID != session.NativeID || original.ProjectID != session.ProjectID || original.CLI != session.CLI || original.Ownership != "owned" || session.Ownership != "owned" {
		return domain.NativeSession{}, nativeError(apierrors.ScopeDenied, "input differs from the original owned session")
	}
	return cloneOriginalNativeSession(*original)
}
