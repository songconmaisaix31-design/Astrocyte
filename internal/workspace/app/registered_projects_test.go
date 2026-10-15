package app_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRegisteredProjectsCachedGitActivity(t *testing.T) {
	// 时间和路径均为聚合函数的明确输入，不声明存在原生会话或来源观察。
	commit := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	creation := commit.Add(time.Hour)
	projects := []domain.RegisteredProject{{
		Root:         "/aggregation-input",
		Git:          domain.ProjectGitObservation{Status: "known", LastCommitAt: &commit},
		Contributors: []domain.ProjectContributor{{CLI: "codex", Source: "native_session_header", CreatedAt: &creation}},
	}}
	board := domain.AggregateProjects(projects)
	if len(board) != 1 || board[0].LastActivityAt == nil || !board[0].LastActivityAt.Equal(commit) {
		t.Fatal("Git commit input did not determine aggregate activity")
	}
	if !board[0].Contributors[0].CreatedAt.Equal(creation) || board[0].Contributors[0].ActivityAt != nil {
		t.Fatal("session creation input became activity")
	}
}

func TestRegisteredProjectsExistingSQLiteCache(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	source, err := agents.NewScopedRegisteredProjects([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	service := app.NewLocalProjectService(nil, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	human := domain.Caller{Kind: "human", ID: "cache-test-human"}
	snapshot, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || snapshot.Status != "complete" || len(snapshot.Projects) != 1 || len(snapshot.Board) != 1 {
		t.Fatal("actual scoped Git discovery failed", err)
	}
	if snapshot.Projects[0].Root != root || snapshot.Projects[0].Git.Status != "known" || snapshot.Projects[0].Git.LastCommitAt == nil {
		t.Fatal("actual approved Git commit observation is missing")
	}
	metadata := domain.ProjectHumanMetadata{Notes: "cache test human notes", Archived: true}
	annotated, err := service.SetRegisteredProjectMetadata(ctx, human, snapshot.Board[0].ID, metadata)
	if err != nil || annotated.Human.Revision != 1 || annotated.Human.UpdatedAt == nil {
		t.Fatal("actual metadata write failed", err)
	}
	// 只清空可重建的 Board 投影，全部实际来源观察保持原值。
	for i := range snapshot.Board {
		snapshot.Board[i].LastActivityAt = nil
	}
	if err := db.SaveRegisteredProjectDiscovery(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	var before string
	if err := db.Conn().QueryRow("SELECT data FROM local_project_discovery WHERE id='registered'").Scan(&before); err != nil {
		t.Fatal(err)
	}
	var persisted domain.ProjectDiscoverySnapshot
	if err := json.Unmarshal([]byte(before), &persisted); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Conn().Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	// 缓存读取实例没有发现来源，真实 SQLite 同时禁止写入。
	service.ConfigureRegisteredDiscovery(db, nil)
	got, err := service.ListRegisteredProjects(ctx, human)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Projects, persisted.Projects) || !reflect.DeepEqual(got.Sources, persisted.Sources) || !reflect.DeepEqual(got.Failures, persisted.Failures) || got.Status != persisted.Status || !reflect.DeepEqual(got.ObservedAt, persisted.ObservedAt) {
		t.Fatal("cache GET changed discovery observations")
	}
	if len(got.Board) != 1 || !reflect.DeepEqual(got.Board[0].Human, annotated.Human) {
		t.Fatal("cache GET lost actual human metadata")
	}
	oldKnown, newKnown, recovered := 0, 0, 0
	old := map[string]*time.Time{}
	for _, p := range persisted.Board {
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
			t.Fatal("cached aggregate activity differs from actual observations")
		}
		if p.LastActivityAt != nil {
			newKnown++
			if old[p.ID] == nil {
				recovered++
			}
		}
	}
	if len(got.Board) == 0 || recovered == 0 {
		t.Fatal("actual cache did not exercise missing aggregate Git activity")
	}
	var after string
	if err := db.Conn().QueryRow("SELECT data FROM local_project_discovery WHERE id='registered'").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatal("persisted cache changed")
	}
	t.Logf("actual roots=%d groups=%d old_activity=%d new_activity=%d recovered=%d; source disconnected; SQLite query_only; persisted cache unchanged", len(persisted.Projects), len(got.Board), oldKnown, newKnown, recovered)
}
