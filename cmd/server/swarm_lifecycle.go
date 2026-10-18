package main

import (
	"context"
	"time"

	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
)

func startSwarmLifecycle(ctx context.Context, service *swarm.Service) (<-chan error, func()) {
	lifecycleCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	failures := make(chan error, 1)
	go func() {
		defer close(done)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			bounded, release := context.WithTimeout(lifecycleCtx, 10*time.Second)
			err := service.EnforceStopConditions(bounded)
			release()
			if lifecycleCtx.Err() != nil {
				return
			}
			if err != nil {
				failures <- err
				return
			}
			select {
			case <-lifecycleCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return failures, func() { cancel(); <-done }
}
