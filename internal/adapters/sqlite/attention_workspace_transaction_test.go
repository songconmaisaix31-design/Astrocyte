package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func TestAttentionUsesActualWorkspaceTransaction(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		name := "commit"
		if rollback {
			name = "actual_constraint_rollback"
		}
		t.Run(name, func(t *testing.T) {
			db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
			repository := NewAttentionRepository(db)
			service := attention.NewAttentionService(repository, nil, nil, attention.ServiceOptions{ProjectSpaces: attentionSpaces(db)})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			before := db.Conn().Stats().WaitCount
			var created attention.ProjectSpaceResult
			err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
				var err error
				created, err = service.CreateProjectSpace(tx.Context(), attention.Principal{Kind: "human", ID: "transaction-test-human"}, attention.ProjectSpaceCommand{
					CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
					Title:       "Human-created shared transaction space",
				})
				if err != nil {
					return err
				}
				if rollback {
					// 同一真实空间再次 INSERT，触发实际 SQLite 主键约束。
					return repository.WithTx(tx.Context(), func(nested attention.AttentionTx) error {
						return nested.SaveProjectSpace(created.Space, 0)
					})
				}
				return nil
			})
			if rollback {
				var serviceError *apierrors.ServiceError
				if !errors.As(err, &serviceError) || serviceError.Code != apierrors.VersionConflict {
					t.Fatal("actual nested uniqueness failure did not reach the outer transaction", err)
				}
			} else if err != nil || created.Space.ID == "" {
				t.Fatal("actual nested human command failed", err)
			}
			if db.Conn().Stats().WaitCount != before {
				t.Fatal("nested attention command borrowed a second SQLite connection")
			}
			want := 1
			if rollback {
				want = 0
			}
			for _, table := range []string{"attention_project_spaces", "attention_outbox", "attention_receipts"} {
				var count int
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != want {
					t.Fatalf("actual %s transaction result count=%d expected=%d: %v", table, count, want, err)
				}
			}
		})
	}
}

func TestAttentionRejectsActualWorkspaceTransactionFromAnotherDB(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "first.sqlite"))
	other := openAttentionDB(t, filepath.Join(t.TempDir(), "other.sqlite"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	called := false
	err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return NewAttentionRepository(other).WithTx(tx.Context(), func(nested attention.AttentionTx) error {
			called = true
			_, err := nested.ListMaterials()
			return err
		})
	})
	if err == nil || called || other.Conn().Stats().WaitCount != 0 {
		t.Fatal("another database was used inside the current workspace transaction")
	}
	if err := NewAttentionRepository(other).WithTx(ctx, func(tx attention.AttentionTx) error {
		_, err := tx.ListMaterials()
		return err
	}); err != nil {
		t.Fatal("absent workspace context did not permit a normal independent transaction", err)
	}
}
