package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayRepositoryCleanupActualSQLiteLifecycle(t *testing.T) {
	for _, closed := range []bool{false, true} {
		name := "cancel_before_database_close"
		if closed {
			name = "actual_database_failure"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			root := t.TempDir()
			db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
			if err != nil {
				t.Fatal("实际清理数据库打开失败")
			}
			defer func() {
				if db.Close() != nil {
					t.Error("实际清理数据库关闭失败")
				}
			}()
			if db.RunMigrations(ctx, migrations.FS) != nil {
				t.Fatal("实际清理数据库迁移失败")
			}
			environment, err := os.Hostname()
			if err != nil {
				t.Fatal("实际清理部署身份读取失败")
			}
			provider, err := repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: environment}, CacheRoot: filepath.Join(root, "repository-context"), SnapshotBudget: domain.RepositoryBudget{MaxFiles: 20000, MaxBytes: 134217728, MaxResults: 200, MaxDepth: 5}})
			if err != nil {
				t.Fatal("实际清理 provider 创建失败")
			}
			if closed {
				if db.Close() != nil {
					t.Fatal("实际数据库关闭失败")
				}
			}
			results, stop := startRelayRepositoryCleanup(ctx, provider, 1)
			if closed {
				select {
				case err := <-results:
					if err == nil {
						t.Fatal("关闭数据库的真实错误未上送")
					}
				case <-ctx.Done():
					t.Fatal("真实清理错误未及时返回")
				}
				if stop() == nil {
					t.Fatal("真实清理停止未保留原错误")
				}
				return
			}
			if stop() != nil {
				t.Fatal("实际清理取消和等待失败")
			}
			select {
			case err := <-results:
				if err != nil {
					t.Fatal("实际清理取消产生错误")
				}
			default:
				t.Fatal("停止返回时实际 worker 尚未完成")
			}
			if db.Conn().PingContext(ctx) != nil {
				t.Fatal("实际 worker 结束后主数据库不可用")
			}
		})
	}
}
