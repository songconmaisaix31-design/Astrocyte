package app_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestProgressFailureKeepsActualSQLiteCancellation(t *testing.T) {
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	service := app.NewProjectProgressService(nil, db, nil, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// 语义错误是明确调用参数，不声明发生了模型调用或交付。
	semantic := &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "caller-declared unknown outcome"}
	for _, invoke := range []func() error{
		func() error {
			_, err := app.SettleProgressFailureForTest(service, ctx, domain.ProjectProgress{}, uuid.NewString(), "unknown", semantic)
			return err
		},
		func() error {
			_, err := app.UnknownComputedProgressForTest(service, ctx, uuid.NewString(), uuid.NewString(), semantic)
			return err
		},
	} {
		err := invoke()
		if !errors.Is(err, semantic) || !errors.Is(err, context.Canceled) {
			t.Fatal("actual SQLite cancellation or original semantic error was suppressed")
		}
	}
	var rows int
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM local_project_progress").Scan(&rows); err != nil || rows != 0 {
		t.Fatal("failed operations wrote a progress observation", err)
	}
}
