package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ app.ProjectProgressRepository = (*DB)(nil)

// 没有进度记录时返回零值，查询保持原有观察状态。
func (db *DB) LoadProjectProgress(ctx context.Context, id string) (domain.ProjectProgress, error) {
	var progress domain.ProjectProgress
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_project_progress WHERE project_id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return progress, nil
	}
	if err != nil {
		return progress, err
	}
	err = json.Unmarshal([]byte(data), &progress)
	return progress, err
}

// 使用原 revision 的 CAS 约束，新记录的 expected 为零。
func (db *DB) SaveProjectProgress(ctx context.Context, progress domain.ProjectProgress, expected int) error {
	if progress.Revision != expected+1 || expected < 0 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "invalid project progress revision"}
	}
	data, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO local_project_progress(project_id,data,revision) VALUES(?,?,?) ON CONFLICT(project_id) DO NOTHING", progress.ProjectID, string(data), progress.Revision)
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE local_project_progress SET data=?,revision=? WHERE project_id=? AND revision=?", string(data), progress.Revision, progress.ProjectID, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "project progress changed; reload before saving", RequiredAction: "reload_project_progress"}
	}
	return nil
}
