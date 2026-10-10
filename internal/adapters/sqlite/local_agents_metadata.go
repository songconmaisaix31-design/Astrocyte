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

var _ app.RegisteredProjectMetadataRepository = (*DB)(nil)

func (db *DB) LoadProjectMetadata(ctx context.Context, id string) (domain.ProjectHumanMetadata, error) {
	var data string
	var metadata domain.ProjectHumanMetadata
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_project_metadata WHERE project_id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return metadata, nil
	}
	if err != nil {
		return metadata, err
	}
	err = json.Unmarshal([]byte(data), &metadata)
	return metadata, err
}

func (db *DB) SaveProjectMetadata(ctx context.Context, id string, metadata domain.ProjectHumanMetadata, expected int) error {
	if metadata.Revision != expected+1 || expected < 0 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "invalid human metadata revision"}
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO local_project_metadata(project_id,data,revision) VALUES(?,?,?) ON CONFLICT(project_id) DO NOTHING", id, string(data), metadata.Revision)
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE local_project_metadata SET data=?,revision=? WHERE project_id=? AND revision=?", string(data), metadata.Revision, id, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "human project metadata changed; reload before saving", RequiredAction: "reload_project_metadata"}
	}
	return nil
}
