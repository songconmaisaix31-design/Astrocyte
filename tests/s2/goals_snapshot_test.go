package s2_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestActualCurrentGoalTransactionAndInactiveStates(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".verification", "s2-first")
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(base, "current-goal-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("local database: %s", filepath.Join(dir, "state.sqlite"))
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	otherDB, err := sqlite.Open(filepath.Join(dir, "other.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := otherDB.Close(); err != nil {
			t.Error(err)
		}
	}()
	human := domain.Caller{Kind: "human", ID: uuid.NewString()}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public current target transaction"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Astrocyte current target transaction", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	service := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	goal, err := service.PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "Verify the current target before saving a judgment", Status: "active", SuccessCriteria: []string{"An unsuccessful database callback leaves the original goal unchanged"}}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := service.ReadCurrentGoalSnapshot(ctx, human, project.ID, goal.ID, goal.Revision)
	if err != nil {
		t.Fatal(err)
	}
	change := workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(goal.Revision), Goal: goal}
	change.Goal.Status = "paused"
	rollback := errors.New("abort actual goal change after its database write")
	err = service.WithCurrentGoalSnapshot(ctx, human, snapshot, func(txctx context.Context, current domain.CurrentGoalSnapshot) error {
		if !reflect.DeepEqual(current.Goal, goal) {
			t.Fatal("current target differs from its immutable original revision")
		}
		if _, err := sqlite.WorkspaceTransaction(txctx, otherDB); err == nil {
			t.Fatal("goal callback accepted a different SQLite database")
		}
		changed, err := service.PutGoal(txctx, human, goal.ID, change)
		if err != nil {
			return err
		}
		if changed.Revision != goal.Revision+1 || changed.Status != "paused" {
			t.Fatal("actual nested goal change did not execute")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("actual callback did not return its abort cause: %v", err)
	}
	unchanged, err := service.ReadGoal(ctx, human, goal.ID, 0)
	if err != nil || !reflect.DeepEqual(unchanged, goal) {
		t.Fatalf("aborted goal callback left a partial new revision: %v", err)
	}
	current, err := service.PutGoal(ctx, human, goal.ID, change)
	if err != nil || current.Status != "paused" || current.Revision != goal.Revision+1 {
		t.Fatalf("rolled back command receipt was retained: %v", err)
	}
	for _, status := range []string{"paused", "completed", "archived"} {
		if current.Status != status {
			update := current
			update.Status = status
			current, err = service.PutGoal(ctx, human, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(current.Revision), Goal: update})
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err = service.ReadCurrentGoalSnapshot(ctx, human, project.ID, goal.ID, current.Revision)
		requireCode(t, err, apierrors.ScopeDenied)
		inactive := snapshot
		inactive.Goal = current
		err = service.WithCurrentGoalSnapshot(ctx, human, inactive, func(context.Context, domain.CurrentGoalSnapshot) error {
			t.Fatal("inactive target callback executed")
			return nil
		})
		requireCode(t, err, apierrors.ScopeDenied)
	}
	historical, err := service.ReadGoal(ctx, human, goal.ID, goal.Revision)
	if err != nil || !reflect.DeepEqual(historical, goal) {
		t.Fatalf("later target states changed the original goal: %v", err)
	}
}
