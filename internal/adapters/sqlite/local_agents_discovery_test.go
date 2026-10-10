package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type registeredSourceFixture struct {
	calls    int
	err      error
	snapshot domain.ProjectDiscoverySnapshot
}

func (s *registeredSourceFixture) DiscoverRegistered(context.Context) (domain.ProjectDiscoverySnapshot, error) {
	s.calls++
	return s.snapshot, s.err
}

func TestRegisteredProjectCacheRestartStaleAndNoAuthority(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, dbPath)
	source := &registeredSourceFixture{}
	service := workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{&localNative{}})
	service.ConfigureRegisteredDiscovery(db, source)
	human := domain.Caller{Kind: "human", ID: "browser-human"}
	initial, err := service.ListRegisteredProjects(ctx, human)
	if err != nil || initial.Status != "unknown" || initial.ObservedAt != nil || source.calls != 0 || initial.Projects == nil || initial.Failures == nil {
		t.Fatalf("initial %+v %v", initial, err)
	}
	for _, caller := range []domain.Caller{{Kind: "agent", ID: "agent", ProjectID: "p"}, {Kind: "human"}, {}} {
		if _, err := service.ListRegisteredProjects(ctx, caller); err == nil {
			t.Fatal("cache visible without human identity")
		}
		if _, err := service.RefreshRegisteredProjects(ctx, caller); err == nil {
			t.Fatal("nonhuman refresh accepted")
		}
	}
	if source.calls != 0 {
		t.Fatal("denied refresh executed source")
	}
	root := t.TempDir()
	stamp := time.Now().Add(-time.Hour).UTC()
	source.snapshot = domain.ProjectDiscoverySnapshot{Status: "complete", ObservedAt: &stamp, Projects: []domain.RegisteredProject{{Root: root, Name: "Actual observed folder", Source: "orca_registered", Git: domain.ProjectGitObservation{Status: "unknown"}, Activity: domain.ProjectActivityObservation{Status: "unknown"}}}, Failures: []domain.ProjectDiscoveryFailure{}}
	observed, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || observed.Status != "complete" || source.calls != 1 {
		t.Fatalf("refresh %+v %v", observed, err)
	}
	projects, err := service.ListProjects(ctx, human)
	if err != nil || len(projects) != 0 {
		t.Fatal("discovery granted project access or created a space")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	service = workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{&localNative{}})
	service.ConfigureRegisteredDiscovery(db, source)
	restarted, err := service.ListRegisteredProjects(ctx, human)
	if err != nil || restarted.Status != "complete" || !restarted.ObservedAt.Equal(stamp) || len(restarted.Projects) != 1 || source.calls != 1 {
		t.Fatalf("cold restart %+v %v", restarted, err)
	}
	source.err = errors.New("private diagnostic must not enter cache")
	stale, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || stale.Status != "stale" || len(stale.Projects) != 1 || !stale.ObservedAt.Equal(stamp) || len(stale.Failures) != 1 || stale.Failures[0].Reason != "registered_source_unavailable" {
		t.Fatalf("stale %+v %v", stale, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	service = workspace.NewLocalProjectService(db, &localReferences{}, agents.ProjectFiles{}, localRegistry{&localNative{}})
	service.ConfigureRegisteredDiscovery(db, source)
	stale, err = service.ListRegisteredProjects(ctx, human)
	if err != nil || stale.Status != "stale" || !stale.ObservedAt.Equal(stamp) || source.calls != 2 {
		t.Fatalf("stale restart %+v %v", stale, err)
	}
	// A new source observation updates cache without changing project settings.
	source.err = nil
	newStamp := stamp.Add(30 * time.Minute)
	source.snapshot.ObservedAt = &newStamp
	if _, err = service.RefreshRegisteredProjects(ctx, human); err != nil {
		t.Fatal(err)
	}
	projects, err = service.ListProjects(ctx, human)
	if err != nil || len(projects) != 0 {
		t.Fatal("refresh created executable projects")
	}
}

func TestRegisteredSourceUnavailableBeforeFirstObservation(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	source := &registeredSourceFixture{err: errors.New("disconnected")}
	service := workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	got, err := service.RefreshRegisteredProjects(context.Background(), domain.Caller{Kind: "human", ID: "h"})
	if err != nil || got.Status != "unknown" || got.ObservedAt != nil || len(got.Projects) != 0 || len(got.Failures) != 1 {
		t.Fatalf("unobserved %+v %v", got, err)
	}
}
