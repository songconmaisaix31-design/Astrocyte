package s2_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func verifyWorkspaceCommitNotifications(t *testing.T, ctx context.Context, db, otherDB *sqlite.DB, service *workspace.S2Service, human domain.Caller, projectID string) {
	t.Helper()
	repo := sqlite.NewWorkspaceRepository(db)
	create := func(txctx context.Context, title string) (domain.Goal, error) {
		return service.PutGoal(txctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(0), Goal: domain.Goal{ProjectID: projectID, Title: title, Status: "active", SuccessCriteria: []string{"Notify only after the actual outer SQLite commit"}}})
	}
	t.Run("notifications follow the outer commit and connection release", func(t *testing.T) {
		var saved domain.Goal
		var transactionContext context.Context
		var notifications []int
		notify := func(order int) func() {
			return func() {
				readctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				current, err := service.ReadGoal(readctx, human, saved.ID, saved.Revision)
				if err != nil || !reflect.DeepEqual(current, saved) {
					t.Errorf("notification could not read the committed goal through the released single connection: %v", err)
				}
				notifications = append(notifications, order)
			}
		}
		err := repo.WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
			transactionContext = tx.Context()
			var err error
			saved, err = create(tx.Context(), "Verify commit notification ordering on the public project")
			if err != nil {
				return err
			}
			if err := db.AfterWorkspaceCommit(tx.Context(), notify(1)); err != nil {
				return err
			}
			if err := repo.WithWorkspaceTx(tx.Context(), func(inner workspace.WorkspaceTx) error {
				return db.AfterWorkspaceCommit(inner.Context(), notify(2))
			}); err != nil {
				return err
			}
			if len(notifications) != 0 {
				t.Fatal("notification ran before the outer transaction committed")
			}
			return nil
		})
		if err != nil || !reflect.DeepEqual(notifications, []int{1, 2}) {
			t.Fatalf("outer commit notifications: %v, %v", notifications, err)
		}
		if err := db.AfterWorkspaceCommit(transactionContext, notify(3)); err == nil {
			t.Fatal("finished transaction accepted a notification")
		}
		if err := db.AfterWorkspaceCommit(ctx, notify(3)); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(notifications, []int{1, 2, 3}) {
			t.Fatalf("successful save without an outer transaction did not notify immediately: %v", notifications)
		}
	})
	t.Run("callback failure rolls back without notification", func(t *testing.T) {
		var saved domain.Goal
		notifications := 0
		cause := errors.New("abort actual goal save before its outer commit")
		err := repo.WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
			var err error
			saved, err = create(tx.Context(), "Verify an aborted save never wakes cleanup")
			if err != nil {
				return err
			}
			if err := db.AfterWorkspaceCommit(tx.Context(), func() { notifications++ }); err != nil {
				return err
			}
			return cause
		})
		if !errors.Is(err, cause) || notifications != 0 {
			t.Fatalf("failed transaction notified: count=%d error=%v", notifications, err)
		}
		_, err = service.ReadGoal(ctx, human, saved.ID, 0)
		requireCode(t, err, apierrors.NotFound)
	})
	t.Run("actual deferred constraint failure at commit does not notify", func(t *testing.T) {
		var saved domain.Goal
		notifications := 0
		orphanID := uuid.NewString()
		err := repo.WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
			conn, err := sqlite.WorkspaceTransaction(tx.Context(), db)
			if err != nil {
				return err
			}
			if _, err := conn.ExecContext(tx.Context(), "PRAGMA defer_foreign_keys = ON"); err != nil {
				return err
			}
			saved, err = create(tx.Context(), "Verify a failed SQLite commit never wakes cleanup")
			if err != nil {
				return err
			}
			if err := db.AfterWorkspaceCommit(tx.Context(), func() { notifications++ }); err != nil {
				return err
			}
			_, err = conn.ExecContext(tx.Context(), "INSERT INTO workspace_goal_revisions(id,revision,data) SELECT ?,revision,data FROM workspace_goal_revisions WHERE id=? AND revision=?", orphanID, saved.ID, saved.Revision)
			if err != nil {
				t.Fatalf("deferred foreign key failed before the actual COMMIT: %v", err)
			}
			return nil
		})
		if err == nil || notifications != 0 {
			t.Fatalf("failed actual commit notified: count=%d error=%v", notifications, err)
		}
		_, err = service.ReadGoal(ctx, human, saved.ID, 0)
		requireCode(t, err, apierrors.NotFound)
		var remaining int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM workspace_goal_revisions WHERE id=?", orphanID).Scan(&remaining); err != nil || remaining != 0 {
			t.Fatalf("failed actual commit retained a revision: count=%d error=%v", remaining, err)
		}
	})
	t.Run("different database and nil notifications are rejected", func(t *testing.T) {
		notifications := 0
		if err := db.AfterWorkspaceCommit(ctx, nil); err == nil {
			t.Fatal("nil notification was accepted outside a transaction")
		}
		err := repo.WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
			if err := db.AfterWorkspaceCommit(tx.Context(), nil); err == nil {
				t.Fatal("nil notification was accepted in a transaction")
			}
			if err := otherDB.AfterWorkspaceCommit(tx.Context(), func() { notifications++ }); err == nil {
				t.Fatal("different database accepted a commit notification")
			}
			return nil
		})
		if err != nil || notifications != 0 {
			t.Fatalf("invalid registration notified: count=%d error=%v", notifications, err)
		}
	})
}
