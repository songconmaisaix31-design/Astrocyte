package main

import (
	"context"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// 清理结果进入原主生命周期；取消后等待实际 worker 返回再关闭主数据库。
func startRelayRepositoryCleanup(ctx context.Context, provider workspace.RepositoryCleanupWorker, maximum int) (<-chan error, func() error) {
	workerCtx, cancel := context.WithCancel(ctx)
	results := make(chan error, 1)
	done := make(chan struct{})
	var result error
	go func() {
		result = provider.RunRepositoryCleanup(workerCtx, maximum)
		results <- result
		close(done)
	}()
	return results, func() error {
		cancel()
		<-done
		return result
	}
}
