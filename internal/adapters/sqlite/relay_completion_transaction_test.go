package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 实际空库读取应返回 NotFound，当前事务不能再次占用自身的单一连接。
func TestRelayCompletionProjectPermissionUsesActualCurrentTransaction(t *testing.T) {
	for _, name := range []string{"project", "grant"} {
		t.Run(name, func(t *testing.T) {
			db := openAttentionDB(t, filepath.Join(acceptedExperienceDirectory(t), "state.sqlite"))
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := NewAcceptedExperienceRepository(db).WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
				var err error
				if name == "project" {
					_, err = db.LoadProject(tx.Context(), "absent-completion-project")
				} else {
					_, err = db.LoadGrant(tx.Context(), "absent-completion-project", "absent-completion-agent")
				}
				var missing *apierrors.ServiceError
				if errors.As(err, &missing) && missing.Code == apierrors.NotFound {
					return nil
				}
				if err == nil {
					return errors.New("empty actual database returned a project permission record")
				}
				return err
			})
			if err != nil {
				t.Fatalf("current transaction project permission read did not return actual NotFound: %v", err)
			}
		})
	}
}
