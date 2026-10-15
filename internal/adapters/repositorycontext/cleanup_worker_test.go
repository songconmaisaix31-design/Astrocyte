package repositorycontext_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func cleanupService(t *testing.T, invalidCache bool) *repositorycontext.Service {
	t.Helper()
	parent, err := filepath.Abs("../../../.verification/s4/cleanup-worker-tests")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp(parent, "actual-sqlite-")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(root, "projections")
	if invalidCache {
		if err := os.WriteFile(cache, []byte("This owned regular file cannot be a projection directory."), 0600); err != nil {
			t.Fatal(err)
		}
	}
	service, err := repositorycontext.New(db, db, repositorycontext.Options{
		Git: repositorycontext.Git{EnvironmentID: "s4-worker-empty-sqlite"}, CacheRoot: cache,
		SnapshotBudget: domain.RepositoryBudget{MaxFiles: 100, MaxBytes: 1024 * 1024, MaxResults: 100, MaxDepth: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestCleanupWorkerSingleOwnerMergedWakeAndCancellation(t *testing.T) {
	service := cleanupService(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	results := make(chan error, 2)
	for range 2 {
		go func() { results <- service.RunRepositoryCleanup(ctx, 1) }()
	}
	var err error
	select {
	case err = <-results:
	case <-ctx.Done():
		t.Fatal("concurrent worker start did not reject its duplicate")
	}
	var serviceError *apierrors.ServiceError
	if !errors.As(err, &serviceError) || serviceError.Code != apierrors.VersionConflict {
		t.Fatal("concurrent worker start", err)
	}
	for range 1000 {
		service.WakeRepositoryCleanup()
	}
	cancel()
	select {
	case err := <-results:
		if err != nil {
			t.Fatal("worker cancellation", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not finish before SQLite shutdown")
	}
}

func TestCleanupWorkerReturnsActualFilesystemFailure(t *testing.T) {
	service := cleanupService(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := service.RunRepositoryCleanup(ctx, 1)
	var serviceError *apierrors.ServiceError
	if !errors.As(err, &serviceError) || serviceError.Code != apierrors.ScopeDenied {
		t.Fatal("actual unsafe cache directory was accepted", err)
	}
}
