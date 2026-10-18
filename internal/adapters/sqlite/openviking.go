package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (db *DB) RecoverOpenViking(ctx context.Context) error {
	rows, err := db.ListOpenViking(ctx, "")
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status != "submitting" {
			continue
		}
		expected := row.Version
		row.Version++
		row.Status = "unknown"
		row.Error = "进程在外部提交结束前终止；原操作结果需要核实"
		row.UpdatedAt = time.Now().UTC()
		if err := db.SaveOpenViking(ctx, row, expected); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) ReserveOpenViking(ctx context.Context, row domain.OpenVikingSubmission) (domain.OpenVikingSubmission, bool, error) {
	data, err := json.Marshal(row)
	if err != nil {
		return row, false, err
	}
	result, err := db.conn.ExecContext(ctx, `INSERT INTO attention_openviking_submissions(id,version,request_id,space_id,material_id,revision,base_url,account,user_id,space_version,destination,status,data) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`, row.ID, row.Version, row.RequestID, row.SpaceID, row.MaterialID, row.Revision, row.BaseURL, row.Account, row.User, row.SpaceVersion, row.Destination, row.Status, string(data))
	if err != nil {
		return row, false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return row, false, err
	}
	if n == 1 {
		return row, true, nil
	}
	var raw string
	err = db.conn.QueryRowContext(ctx, `SELECT data FROM attention_openviking_submissions WHERE request_id=?`, row.RequestID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		err = db.conn.QueryRowContext(ctx, `SELECT data FROM attention_openviking_submissions WHERE space_id=? AND material_id=? AND revision=? AND base_url=? AND account=? AND user_id=? AND destination=? AND status<>'failed'`, row.SpaceID, row.MaterialID, row.Revision, row.BaseURL, row.Account, row.User, row.Destination).Scan(&raw)
		if err != nil {
			return row, false, err
		}
		var original domain.OpenVikingSubmission
		if err := json.Unmarshal([]byte(raw), &original); err != nil {
			return row, false, err
		}
		if original.Status == "completed" {
			return original, false, nil
		}
		return original, false, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "该资料版本已有提交记录，请核实原操作", RequiredAction: "inspect_original_submission"}
	}
	if err != nil {
		return row, false, err
	}
	var prior domain.OpenVikingSubmission
	if err := json.Unmarshal([]byte(raw), &prior); err != nil {
		return row, false, err
	}
	if prior.SpaceID != row.SpaceID || prior.SpaceVersion != row.SpaceVersion || prior.MaterialID != row.MaterialID || prior.Revision != row.Revision || prior.Destination != row.Destination || prior.BaseURL != row.BaseURL || prior.ActorID != row.ActorID || prior.Account != row.Account || prior.User != row.User {
		return row, false, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "request_id 已用于其他提交内容", RequiredAction: "inspect_original_submission"}
	}
	return prior, false, nil
}

func (db *DB) SaveOpenViking(ctx context.Context, row domain.OpenVikingSubmission, expected int) error {
	if row.Version != expected+1 {
		return fmt.Errorf("invalid OpenViking record version")
	}
	data, err := json.Marshal(row)
	if err != nil {
		return err
	}
	result, err := db.conn.ExecContext(ctx, `UPDATE attention_openviking_submissions SET version=?,status=?,data=? WHERE id=? AND version=? AND request_id=? AND space_id=? AND material_id=? AND revision=? AND base_url=? AND account=? AND user_id=? AND space_version=? AND destination=? AND json_extract(data,'$.actor_id')=? AND json_extract(data,'$.approved_at')=? AND json_extract(data,'$.locator')=? AND json_extract(data,'$.title')=?`, row.Version, row.Status, string(data), row.ID, expected, row.RequestID, row.SpaceID, row.MaterialID, row.Revision, row.BaseURL, row.Account, row.User, row.SpaceVersion, row.Destination, row.ActorID, row.ApprovedAt.Format(time.RFC3339Nano), row.Locator, row.Title)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "提交记录已变化", RequiredAction: "reload_submission"}
	}
	return nil
}

func (db *DB) GetOpenViking(ctx context.Context, id string) (domain.OpenVikingSubmission, error) {
	var row domain.OpenVikingSubmission
	var raw string
	err := db.conn.QueryRowContext(ctx, `SELECT data FROM attention_openviking_submissions WHERE id=?`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return row, apierrors.NewNotFound("OpenViking submission", id)
	}
	if err != nil {
		return row, err
	}
	err = json.Unmarshal([]byte(raw), &row)
	return row, err
}

func (db *DB) ListOpenViking(ctx context.Context, spaceID string) ([]domain.OpenVikingSubmission, error) {
	query := `SELECT data FROM attention_openviking_submissions`
	args := []any{}
	if spaceID != "" {
		query += ` WHERE space_id=?`
		args = append(args, spaceID)
	}
	query += ` ORDER BY id`
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.OpenVikingSubmission{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var row domain.OpenVikingSubmission
		if err := json.Unmarshal([]byte(raw), &row); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}
