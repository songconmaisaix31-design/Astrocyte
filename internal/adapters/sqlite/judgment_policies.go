package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

var _ judgment.PolicyRepository = (*DB)(nil)

func (db *DB) LoadJudgmentPolicy(ctx context.Context, project string, questionType judgment.QuestionType) (judgment.Policy, error) {
	var policy judgment.Policy
	var data string
	var revision int
	err := db.conn.QueryRowContext(ctx, "SELECT version,data FROM judgment_policies WHERE project_id=? AND question_type=?", project, questionType).Scan(&revision, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return policy, apierrors.NewNotFound("judgment policy", project+":"+string(questionType))
	}
	if err != nil {
		return policy, err
	}
	if err = json.Unmarshal([]byte(data), &policy); err != nil {
		return policy, err
	}
	if policy.Revision != revision || policy.ProjectID != project || policy.QuestionType != questionType {
		return policy, &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "stored judgment policy identity differs"}
	}
	return policy, judgment.ValidatePolicy(policy)
}

func (db *DB) SaveJudgmentPolicy(ctx context.Context, policy judgment.Policy, expected int) error {
	if policy.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment policy requires next revision"}
	}
	if err := judgment.ValidatePolicy(policy); err != nil {
		return err
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO judgment_policies(project_id,question_type,version,data) VALUES(?,?,?,?) ON CONFLICT(project_id,question_type) DO NOTHING", policy.ProjectID, policy.QuestionType, policy.Revision, string(data))
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE judgment_policies SET version=?,data=? WHERE project_id=? AND question_type=? AND version=?", policy.Revision, string(data), policy.ProjectID, policy.QuestionType, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "judgment policy was changed concurrently"}
	}
	return nil
}
