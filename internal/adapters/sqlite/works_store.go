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

var _ app.WorksRepository = (*DB)(nil)

const (
	worksAlbumKey = "works_album"
)

// LoadWorksSnapshot returns an empty album for a fresh database so a cache GET
// never invents a work. It performs no scan or filesystem access.
func (db *DB) LoadWorksSnapshot(ctx context.Context) (domain.WorksSnapshot, error) {
	snapshot := domain.WorksSnapshot{Works: []domain.Work{}, Failures: []domain.WorkDiscoveryFailure{}}
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM works_album WHERE key=?", worksAlbumKey).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return snapshot, nil
	}
	if err != nil {
		return snapshot, err
	}
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		return snapshot, err
	}
	if snapshot.Works == nil {
		snapshot.Works = []domain.Work{}
	}
	if snapshot.Failures == nil {
		snapshot.Failures = []domain.WorkDiscoveryFailure{}
	}
	return snapshot, nil
}

// SaveWorksSnapshot replaces the replaceable album observation. Human metadata
// lives in a separate table and is never part of this write.
func (db *DB) SaveWorksSnapshot(ctx context.Context, snapshot domain.WorksSnapshot) error {
	if snapshot.Works == nil {
		snapshot.Works = []domain.Work{}
	}
	if snapshot.Failures == nil {
		snapshot.Failures = []domain.WorkDiscoveryFailure{}
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	_, err = db.conn.ExecContext(ctx,
		"INSERT INTO works_album(key,data) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET data=excluded.data",
		worksAlbumKey, string(data))
	return err
}

// LoadWorkMetadata returns the zero record for an absent work so a cache GET
// never invents an attitude or handoff.
func (db *DB) LoadWorkMetadata(ctx context.Context, id string) (domain.WorkHumanMetadata, error) {
	var metadata domain.WorkHumanMetadata
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM works_metadata WHERE work_id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return metadata, nil
	}
	if err != nil {
		return metadata, err
	}
	err = json.Unmarshal([]byte(data), &metadata)
	return metadata, err
}

// SaveWorkMetadata uses the same revision CAS discipline as project metadata.
// expected is the revision this write is based on; a new record passes 0.
func (db *DB) SaveWorkMetadata(ctx context.Context, id string, metadata domain.WorkHumanMetadata, expected int) error {
	if metadata.Revision != expected+1 || expected < 0 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "invalid work metadata revision"}
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO works_metadata(work_id,data,revision) VALUES(?,?,?) ON CONFLICT(work_id) DO NOTHING", id, string(data), metadata.Revision)
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE works_metadata SET data=?,revision=? WHERE work_id=? AND revision=?", string(data), metadata.Revision, id, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "work metadata changed; reload before saving", RequiredAction: "reload_work_metadata"}
	}
	return nil
}
