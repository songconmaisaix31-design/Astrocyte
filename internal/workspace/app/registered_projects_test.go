package app_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	_ "modernc.org/sqlite"
)

type cachedProjectRepository struct {
	snapshot domain.ProjectDiscoverySnapshot
	metadata domain.ProjectHumanMetadata
	writes   int
}

func (r *cachedProjectRepository) LoadRegisteredProjectDiscovery(context.Context) (domain.ProjectDiscoverySnapshot, error) {
	return r.snapshot, nil
}
func (r *cachedProjectRepository) SaveRegisteredProjectDiscovery(context.Context, domain.ProjectDiscoverySnapshot) error {
	r.writes++
	return nil
}
func (r *cachedProjectRepository) LoadProjectMetadata(context.Context, string) (domain.ProjectHumanMetadata, error) {
	return r.metadata, nil
}
func (r *cachedProjectRepository) SaveProjectMetadata(context.Context, string, domain.ProjectHumanMetadata, int) error {
	r.writes++
	return nil
}

type forbiddenProjectScan struct{ calls int }

func (s *forbiddenProjectScan) DiscoverRegistered(context.Context) (domain.ProjectDiscoverySnapshot, error) {
	s.calls++
	return domain.ProjectDiscoverySnapshot{}, nil
}

func cachedProjectList(t *testing.T, snapshot domain.ProjectDiscoverySnapshot) domain.ProjectDiscoverySnapshot {
	t.Helper()
	repo := &cachedProjectRepository{snapshot: snapshot, metadata: domain.ProjectHumanMetadata{Notes: "keep notes", Archived: true, Revision: 2}}
	source := &forbiddenProjectScan{}
	service := app.NewLocalProjectService(nil, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(repo, source)
	got, err := service.ListRegisteredProjects(context.Background(), domain.Caller{Kind: "human", ID: "test-human"})
	if err != nil {
		t.Fatal(err)
	}
	if source.calls != 0 || repo.writes != 0 {
		t.Fatalf("cache GET scanned or wrote: scans=%d writes=%d", source.calls, repo.writes)
	}
	if !reflect.DeepEqual(got.Projects, snapshot.Projects) || !reflect.DeepEqual(got.Sources, snapshot.Sources) || !reflect.DeepEqual(got.Failures, snapshot.Failures) || got.Status != snapshot.Status || !reflect.DeepEqual(got.ObservedAt, snapshot.ObservedAt) {
		t.Fatal("cache GET changed discovery observations")
	}
	for _, p := range got.Board {
		if p.Human != repo.metadata {
			t.Fatal("cache GET lost human metadata")
		}
	}
	return got
}

func TestRegisteredProjectsCachedGitActivity(t *testing.T) {
	commit := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	creation := commit.Add(time.Hour)
	snapshot := domain.ProjectDiscoverySnapshot{
		Status: "partial", ObservedAt: &creation,
		Projects: []domain.RegisteredProject{{Root: "/native", Git: domain.ProjectGitObservation{Status: "known", LastCommitAt: &commit}, Contributors: []domain.ProjectContributor{{CLI: "codex", Source: "native_session_header", CreatedAt: &creation}}}},
		Board:    []domain.ProjectSummary{{LastActivityAt: nil}},
		Sources:  []domain.ProjectSourceObservation{{CLI: "codex", Status: "partial", HeadersExamined: 15, RetainedAssociations: 1}},
		Failures: []domain.ProjectDiscoveryFailure{{Reason: "sample_bound"}},
	}
	got := cachedProjectList(t, snapshot)
	if len(got.Board) != 1 || got.Board[0].LastActivityAt == nil || !got.Board[0].LastActivityAt.Equal(commit) {
		t.Fatalf("old cached board was not reprojected: %+v", got.Board)
	}
	if !got.Board[0].Contributors[0].CreatedAt.Equal(creation) || got.Board[0].Contributors[0].ActivityAt != nil {
		t.Fatal("session creation became activity")
	}
}

// Opt-in reads only an already observed SQLite cache; no adapters or CLI scans.
func TestRegisteredProjectsExistingSQLiteCache(t *testing.T) {
	path := os.Getenv("ASTROCYTE_TEST_PROJECT_ACTIVITY_CACHE")
	if path == "" {
		t.Skip("existing cache path not provided")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	dsn := (&url.URL{Scheme: "file", Path: "/" + strings.TrimPrefix(filepath.ToSlash(absolute), "/"), RawQuery: "mode=ro"}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var raw string
	if err := db.QueryRow("SELECT data FROM local_project_discovery WHERE id='registered'").Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var snapshot domain.ProjectDiscoverySnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		t.Fatal(err)
	}
	got := cachedProjectList(t, snapshot)
	oldKnown, newKnown, recovered := 0, 0, 0
	old := map[string]*time.Time{}
	for _, p := range snapshot.Board {
		old[p.ID] = p.LastActivityAt
		if p.LastActivityAt != nil {
			oldKnown++
		}
	}
	for _, p := range got.Board {
		var want *time.Time
		for _, observation := range p.Observations {
			for _, stamp := range []*time.Time{observation.Git.LastCommitAt, observation.Activity.LastActivityAt} {
				if stamp != nil && (want == nil || stamp.After(*want)) {
					want = stamp
				}
			}
			for _, c := range observation.Contributors {
				if stamp := c.ActivityAt; c.Source != "native_session_header" && stamp != nil && (want == nil || stamp.After(*want)) {
					want = stamp
				}
			}
		}
		if !reflect.DeepEqual(p.LastActivityAt, want) {
			t.Fatalf("cached group %s activity=%v want=%v", p.ID, p.LastActivityAt, want)
		}
		if p.LastActivityAt != nil {
			newKnown++
			if old[p.ID] == nil {
				recovered++
			}
		}
	}
	if len(got.Board) == 0 || recovered == 0 {
		t.Fatal("existing cache did not exercise missing Git activity")
	}
	var after string
	if err := db.QueryRow("SELECT data FROM local_project_discovery WHERE id='registered'").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != raw {
		t.Fatal("persisted cache changed")
	}
	t.Logf("cache-only roots=%d groups=%d old_activity=%d new_activity=%d recovered=%d; scans=0 writes=0 persisted cache unchanged", len(snapshot.Projects), len(got.Board), oldKnown, newKnown, recovered)
}
