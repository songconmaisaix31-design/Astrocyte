package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

var _ judgment.Repository = (*DB)(nil)

func (db *DB) FindCached(ctx context.Context, key string) (judgment.Record, bool, error) {
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return judgment.Record{}, false, err
	}
	record, err := readJudgmentRow(conn.QueryRowContext(ctx, `SELECT actor_id,actor_kind,project_id,cache_key,cacheable,request_json,result_json
		FROM judgment_results WHERE cache_key=? AND cacheable=1 ORDER BY created_at DESC,id DESC LIMIT 1`, key))
	if errors.Is(err, sql.ErrNoRows) {
		return judgment.Record{}, false, nil
	}
	return record, err == nil, err
}

func readJudgmentRow(row *sql.Row) (judgment.Record, error) {
	var record judgment.Record
	var projectID, request, result string
	var cacheable int
	if err := row.Scan(&record.Actor.ID, &record.Actor.Kind, &projectID, &record.CacheKey, &cacheable, &request, &result); err != nil {
		return record, err
	}
	if err := json.Unmarshal([]byte(request), &record.Request); err != nil {
		return record, err
	}
	if err := json.Unmarshal([]byte(result), &record.Result); err != nil {
		return record, err
	}
	if record.Request.ProjectID != projectID || record.Result.Version != 1 {
		return record, fmt.Errorf("stored judgment identity or version differs")
	}
	if record.Actor.Kind == "agent" {
		record.Actor.ProjectID = projectID
	}
	record.Cacheable = cacheable == 1
	return record, nil
}

func (db *DB) SaveResult(ctx context.Context, record judgment.Record) error {
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return err
	}
	if err := judgment.ValidateRequest(record.Request); err != nil {
		return err
	}
	if record.Actor.ID == "" || (record.Actor.Kind != "human" && record.Actor.Kind != "agent") || record.Result.ID == "" || record.Result.Version != 1 || record.CacheKey == "" || record.Result.CreatedAt.IsZero() {
		return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "judgment record identity, immutable version, timestamp and cache key are required"}
	}
	request, err := json.Marshal(record.Request)
	if err != nil {
		return err
	}
	result, err := json.Marshal(record.Result)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO judgment_results(id,project_id,actor_id,actor_kind,result_version,cache_key,cacheable,request_json,result_json,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, record.Result.ID, record.Request.ProjectID, record.Actor.ID, record.Actor.Kind, record.Result.Version, record.CacheKey, record.Cacheable, string(request), string(result), record.Result.CreatedAt.Format(time.RFC3339Nano))
	return err
}

func (db *DB) LoadResult(ctx context.Context, id string) (judgment.Record, error) {
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return judgment.Record{}, err
	}
	record, err := readJudgmentRow(conn.QueryRowContext(ctx, `SELECT actor_id,actor_kind,project_id,cache_key,cacheable,request_json,result_json FROM judgment_results WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return record, apierrors.NewNotFound("judgment result", id)
	}
	return record, err
}

func (db *DB) SaveFeedback(ctx context.Context, feedback judgment.Feedback) error {
	if feedback.ID == "" || feedback.ResultID == "" || feedback.ActorKind != "human" || feedback.ActorID == "" || feedback.CreatedAt.IsZero() || feedback.Reason == "" {
		return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "feedback requires a real human and fixed judgment result"}
	}
	return db.WithJudgmentTx(ctx, func(ctx context.Context) error {
		conn, err := db.judgmentExecutor(ctx)
		if err != nil {
			return err
		}
		var version int
		if err := conn.QueryRowContext(ctx, "SELECT result_version FROM judgment_results WHERE id=?", feedback.ResultID).Scan(&version); err != nil {
			return err
		}
		if version != feedback.ResultVersion {
			return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "feedback result version changed"}
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO judgment_feedback(id,result_id,result_version,actor_id,actor_kind,kind,reason,created_at) VALUES(?,?,?,?,?,?,?,?)`, feedback.ID, feedback.ResultID, feedback.ResultVersion, feedback.ActorID, feedback.ActorKind, feedback.Kind, feedback.Reason, feedback.CreatedAt.Format(time.RFC3339Nano))
		return err
	})
}

func (db *DB) ListFeedback(ctx context.Context, id string) ([]judgment.Feedback, error) {
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := conn.QueryContext(ctx, `SELECT id,result_id,result_version,actor_id,actor_kind,kind,reason,created_at FROM judgment_feedback WHERE result_id=? ORDER BY created_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []judgment.Feedback{}
	for rows.Next() {
		var item judgment.Feedback
		var timestamp string
		if err := rows.Scan(&item.ID, &item.ResultID, &item.ResultVersion, &item.ActorID, &item.ActorKind, &item.Kind, &item.Reason, &timestamp); err != nil {
			return nil, err
		}
		item.CreatedAt, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
