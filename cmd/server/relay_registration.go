package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayTrustedRegistration struct {
	mu       sync.Mutex
	root     string
	identity os.FileInfo
	window   time.Duration
}

func (s *relaySessionRuntime) RegisterTrustedAgents(ctx context.Context, specs []domain.AgentSpec) error {
	registrar, ok := s.RelayRuntime.(workspace.RelayTrustedAgentRegistrar)
	if !ok {
		return &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	return registrar.RegisterTrustedAgents(ctx, specs)
}

// 只重读启动时固定的可信登记；HTTP不能提供程序、路径或启动参数。
func (s *relaySessionRuntime) ReloadTrustedAgents(ctx context.Context) error {
	registration := s.registration
	if registration == nil {
		return &apierrors.ServiceError{Code: apierrors.UnsupportedCapability}
	}
	registration.mu.Lock()
	defer registration.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := relayRoot()
	if err != nil || root != registration.root {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "trusted registration root changed"}
	}
	identity, err := os.Stat(root)
	if err != nil || !os.SameFile(registration.identity, identity) {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "trusted registration root identity changed"}
	}
	configuredRoot, specs, window, err := relayConfiguration()
	if err != nil {
		return err
	}
	if configuredRoot != registration.root || window != registration.window {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "trusted registration startup policy changed"}
	}
	return s.RegisterTrustedAgents(ctx, specs)
}
