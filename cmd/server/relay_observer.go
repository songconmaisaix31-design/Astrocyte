package main

import (
	"context"
	"log/slog"
	"time"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// Observations are a lifecycle feature, separate from manual RelayRuntime
// commands. This port cannot create an intent, launch, send or resume a task.
type relayOwnedObserver interface{ ObserveOwned(context.Context) error }

func startRelayObserver(ctx context.Context, service workspace.RelayRuntime, logger *slog.Logger) func() {
	observer, ok := service.(relayOwnedObserver)
	if !ok {
		return func() {}
	}
	observeCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); runRelayObserver(observeCtx, observer, 2*time.Second, logger) }()
	return func() { cancel(); <-done }
}

func runRelayObserver(ctx context.Context, observer relayOwnedObserver, cadence time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(cadence)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := observer.ObserveOwned(bounded)
			cancel()
			if err != nil && ctx.Err() == nil && logger != nil {
				logger.Error("relay observation stopped; prior evidence retained")
			}
			if err != nil {
				return
			}
		}
	}
}
