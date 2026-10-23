package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func acceptedExperienceDirectory(t *testing.T) string {
	t.Helper()
	base := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if base == "" {
		var err error
		base, err = filepath.Abs(filepath.Join("..", "..", "..", ".verification"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if !filepath.IsAbs(base) {
		t.Fatal("accepted experience verification requires an absolute output root")
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(base, "accepted-experience-actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("actual owned SQLite directory: %s", directory)
	return directory
}

func TestAcceptedExperienceActualTransactionLifecycle(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		name := "commit"
		if rollback {
			name = "actual_constraint_rollback"
		}
		t.Run(name, func(t *testing.T) {
			db := openAttentionDB(t, filepath.Join(acceptedExperienceDirectory(t), "state.sqlite"))
			repository := NewAcceptedExperienceRepository(db)
			attentionRepository := NewAttentionRepository(db)
			service := attention.NewAttentionService(attentionRepository, nil, nil, attention.ServiceOptions{ProjectSpaces: attentionSpaces(db)})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			before := db.Conn().Stats().WaitCount
			err := repository.WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
				// 人工创建实际空间，在同一事务中触发真实主键冲突。
				created, err := service.CreateProjectSpace(tx.Context(), attention.Principal{Kind: "human", ID: "transaction-test-human"}, attention.ProjectSpaceCommand{
					CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
					Title:       "Human-created completion transaction space",
				})
				if err != nil {
					return err
				}
				if rollback {
					return attentionRepository.WithTx(tx.Context(), func(nested attention.AttentionTx) error {
						return nested.SaveProjectSpace(created.Space, 0)
					})
				}
				return nil
			})
			if rollback {
				var serviceError *apierrors.ServiceError
				if !errors.As(err, &serviceError) || serviceError.Code != apierrors.VersionConflict {
					t.Fatal("actual uniqueness failure did not roll back the completion transaction", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if db.Conn().Stats().WaitCount != before {
				t.Fatal("nested completion transaction borrowed a second connection")
			}
			want := 1
			if rollback {
				want = 0
			}
			for _, table := range []string{"attention_project_spaces", "attention_outbox", "attention_receipts"} {
				var count int
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != want {
					t.Fatalf("actual %s count=%d expected=%d: %v", table, count, want, err)
				}
			}
			for _, table := range []string{"attention_engine_events", "attention_engine_experiences", "experience_completion_sources"} {
				var count int
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("transaction lifecycle check created source evidence in %s: %d %v", table, count, err)
				}
			}
		})
	}
}

func TestAcceptedExperienceRejectsActualWrongDatabaseAndClosedStore(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	db := openAttentionDB(t, filepath.Join(directory, "first.sqlite"))
	other := openAttentionDB(t, filepath.Join(directory, "other.sqlite"))
	repository := NewAcceptedExperienceRepository(other)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	called := false
	err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return repository.WithAcceptedExperienceTx(tx.Context(), func(attention.AcceptedExperienceTx) error {
			called = true
			return nil
		})
	})
	if err == nil || called || other.Conn().Stats().WaitCount != 0 {
		t.Fatal("completion repository accepted another database's transaction", err)
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}
	if err := repository.WithAcceptedExperienceTx(ctx, func(attention.AcceptedExperienceTx) error {
		called = true
		return nil
	}); err == nil || called {
		t.Fatal("closed actual completion store accepted a transaction", err)
	}
}

func TestAcceptedExperienceActualCancellationCreatesNoEvidence(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(acceptedExperienceDirectory(t), "state.sqlite"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := NewAcceptedExperienceRepository(db).WithAcceptedExperienceTx(ctx, func(attention.AcceptedExperienceTx) error {
		called = true
		return nil
	})
	if !errors.Is(err, context.Canceled) || called {
		t.Fatal("cancelled completion operation entered an actual transaction", err)
	}
	for _, table := range []string{"attention_engine_events", "attention_engine_experiences", "experience_completion_sources", "attention_receipts"} {
		var count int
		if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("cancelled completion operation wrote %s: %d %v", table, count, err)
		}
	}
}
