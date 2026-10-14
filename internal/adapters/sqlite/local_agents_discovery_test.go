package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type observedRegisteredSource struct {
	source workspace.RegisteredProjectSource
	calls  atomic.Int32
}

func (s *observedRegisteredSource) DiscoverRegistered(ctx context.Context) (domain.ProjectDiscoverySnapshot, error) {
	s.calls.Add(1)
	return s.source.DiscoverRegistered(ctx)
}

func actualRegisteredSource(t *testing.T, root string) *observedRegisteredSource {
	t.Helper()
	source, err := agents.NewScopedRegisteredProjects([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	return &observedRegisteredSource{source: source}
}

func TestRegisteredProjectCacheRestartStaleAndNoAuthority(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, dbPath)
	root := filepath.Join(t.TempDir(), "observed-directory")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	source := actualRegisteredSource(t, root)
	service := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	human := domain.Caller{Kind: "human", ID: "browser-human"}
	initial, err := service.ListRegisteredProjects(ctx, human)
	if err != nil || initial.Status != "unknown" || initial.ObservedAt != nil || source.calls.Load() != 0 || initial.Projects == nil || initial.Failures == nil {
		t.Fatal("initial cache observations are invalid", err)
	}
	for _, caller := range []domain.Caller{{Kind: "agent", ID: "agent", ProjectID: "p"}, {Kind: "human"}, {}} {
		if _, err := service.ListRegisteredProjects(ctx, caller); err == nil {
			t.Fatal("cache visible without human identity")
		}
		if _, err := service.RefreshRegisteredProjects(ctx, caller); err == nil {
			t.Fatal("nonhuman refresh accepted")
		}
	}
	if source.calls.Load() != 0 {
		t.Fatal("denied refresh executed source")
	}
	observed, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || observed.Status != "complete" || observed.ObservedAt == nil || len(observed.Projects) != 1 || source.calls.Load() != 1 {
		t.Fatal("actual directory discovery failed", err)
	}
	if observed.Projects[0].Root != root || observed.Projects[0].Source != "explicit_directory" || observed.Projects[0].Git.Status != "unavailable" || observed.Projects[0].Activity.Status != "unknown" {
		t.Fatal("actual empty directory observations were invented")
	}
	projects, err := service.ListProjects(ctx, human)
	if err != nil || len(projects) != 0 {
		t.Fatal("discovery granted project access or created a space")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	service = workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	restarted, err := service.ListRegisteredProjects(ctx, human)
	if err != nil || !reflect.DeepEqual(restarted, observed) || source.calls.Load() != 1 {
		t.Fatal("cold restart changed actual observations or scanned", err)
	}
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	stale, err := service.RefreshRegisteredProjects(ctx, human)
	var sourceError *apierrors.ServiceError
	if !errors.As(err, &sourceError) || sourceError.Code != apierrors.ScopeDenied || sourceError.Message != "project root is unavailable" || sourceError.RequiredAction != "select_explicit_project_scope" || stale.Status != "stale" || len(stale.Projects) != 1 || !stale.ObservedAt.Equal(*observed.ObservedAt) || len(stale.Failures) != 1 || stale.Failures[0].Reason != "registered_source_unavailable" {
		t.Fatal("actual source disappearance was suppressed or lost the durable observation")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	service = workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	restarted, err = service.ListRegisteredProjects(ctx, human)
	if err != nil || !reflect.DeepEqual(restarted, stale) || source.calls.Load() != 2 {
		t.Fatal("stale restart changed observations or scanned", err)
	}
	// 恢复同一实际目录后重新观察，不改写此前来源的时间。
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	refreshed, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || refreshed.Status != "complete" || refreshed.ObservedAt.Before(*observed.ObservedAt) || source.calls.Load() != 3 {
		t.Fatal("actual restored directory was not refreshed", err)
	}
	projects, err = service.ListProjects(ctx, human)
	if err != nil || len(projects) != 0 {
		t.Fatal("refresh created executable projects")
	}
}

func TestRegisteredSourceUnavailableBeforeFirstObservation(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	root := filepath.Join(t.TempDir(), "disappearing-directory")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	source := actualRegisteredSource(t, root)
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	service := workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	got, err := service.RefreshRegisteredProjects(context.Background(), domain.Caller{Kind: "human", ID: "h"})
	var sourceError *apierrors.ServiceError
	if !errors.As(err, &sourceError) || sourceError.Code != apierrors.ScopeDenied || sourceError.Message != "project root is unavailable" || sourceError.RequiredAction != "select_explicit_project_scope" || got.Status != "unknown" || got.ObservedAt != nil || len(got.Projects) != 0 || len(got.Failures) != 1 || source.calls.Load() != 1 {
		t.Fatal("actual unavailable source was reported as an observation")
	}
}

func TestRegisteredRefreshWaitCancellationDoesNotBlockShutdown(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	source := actualRegisteredSource(t, t.TempDir())
	service := workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	human := domain.Caller{Kind: "human", ID: "h"}
	// 实际事务占用生产单连接，首个刷新因此持有刷新锁并等待 SQLite。
	tx, err := db.Conn().Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
	})
	priorWaits := db.Conn().Stats().WaitCount
	firstCtx, firstCancel := context.WithCancel(context.Background())
	defer firstCancel()
	firstDone := make(chan error, 1)
	go func() { _, err := service.RefreshRegisteredProjects(firstCtx, human); firstDone <- err }()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for db.Conn().Stats().WaitCount == priorWaits {
		select {
		case err := <-firstDone:
			t.Fatal("first refresh ended before waiting for the actual connection", err)
		case <-deadline.C:
			t.Fatal("first refresh did not wait for the actual SQLite connection")
		case <-tick.C:
		}
	}
	waitingCtx, cancel := context.WithCancel(context.Background())
	waitDone := make(chan error, 1)
	go func() { _, err := service.RefreshRegisteredProjects(waitingCtx, human); waitDone <- err }()
	cancel()
	select {
	case err := <-waitDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("waiting refresh lost its actual cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled refresh waited behind the active refresh")
	}
	if source.calls.Load() != 0 || db.Conn().Stats().WaitCount != priorWaits+1 {
		t.Fatal("cancelled waiter started a discovery or another SQLite wait")
	}
	firstCancel()
	select {
	case err := <-firstDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("active refresh lost its actual cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("active cancelled refresh did not exit")
	}
	if source.calls.Load() != 0 {
		t.Fatal("cancelled SQLite wait reached discovery")
	}
}
