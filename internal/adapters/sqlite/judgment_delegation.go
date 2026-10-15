package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

var _ judgment.DelegationRepository = (*DB)(nil)

func (db *DB) LoadDelegationRule(ctx context.Context, id string) (judgment.DelegationRule, error) {
	var rule judgment.DelegationRule
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return rule, err
	}
	var data, project, workflow, owner, state string
	var revision int
	err = conn.QueryRowContext(ctx, "SELECT data,project_id,workflow_id,owner_id,revision,state FROM judgment_delegation_rules WHERE id=?", id).Scan(&data, &project, &workflow, &owner, &revision, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return rule, apierrors.NewNotFound("delegation rule", id)
	}
	if err != nil {
		return rule, err
	}
	if err := json.Unmarshal([]byte(data), &rule); err != nil {
		return rule, err
	}
	if rule.ID != id || rule.ProjectID != project || rule.WorkflowID != workflow || rule.OwnerID != owner || rule.Revision != revision || rule.State != state {
		return rule, conflict("persisted delegation identity differs")
	}
	return rule, judgment.ValidateDelegationRule(rule)
}

func (db *DB) ListDelegationRules(ctx context.Context, project string) ([]judgment.DelegationRule, error) {
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := conn.QueryContext(ctx, "SELECT data FROM judgment_delegation_rules WHERE project_id=? ORDER BY id", project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := []judgment.DelegationRule{}
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var rule judgment.DelegationRule
		if err := json.Unmarshal([]byte(data), &rule); err != nil {
			return nil, err
		}
		if rule.ProjectID != project {
			return nil, conflict("persisted delegation project differs")
		}
		if err := judgment.ValidateDelegationRule(rule); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (db *DB) SaveDelegationRule(ctx context.Context, rule judgment.DelegationRule, expected int) error {
	if err := judgment.ValidateDelegationRule(rule); err != nil {
		return err
	}
	if expected < 0 || rule.Revision != expected+1 {
		return conflict("delegation requires its next exact revision")
	}
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return err
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = conn.ExecContext(ctx, "INSERT INTO judgment_delegation_rules(id,project_id,workflow_id,owner_id,revision,state,data) VALUES(?,?,?,?,?,?,?) ON CONFLICT DO NOTHING", rule.ID, rule.ProjectID, rule.WorkflowID, rule.OwnerID, rule.Revision, rule.State, string(data))
	} else {
		result, err = conn.ExecContext(ctx, "UPDATE judgment_delegation_rules SET revision=?,state=?,data=? WHERE id=? AND project_id=? AND workflow_id=? AND owner_id=? AND revision=?", rule.Revision, rule.State, string(data), rule.ID, rule.ProjectID, rule.WorkflowID, rule.OwnerID, expected)
	}
	return workspaceCAS(result, err)
}

func (db *DB) LoadAdmissionBinding(ctx context.Context, id string) (judgment.AdmissionBinding, error) {
	var binding judgment.AdmissionBinding
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return binding, err
	}
	var data, project, workflow, owner, opportunity, rule string
	var revision, ruleRevision int
	err = conn.QueryRowContext(ctx, "SELECT data,project_id,workflow_id,owner_id,opportunity_id,opportunity_revision,rule_id,rule_revision FROM judgment_admission_bindings WHERE result_id=?", id).Scan(&data, &project, &workflow, &owner, &opportunity, &revision, &rule, &ruleRevision)
	if errors.Is(err, sql.ErrNoRows) {
		return binding, apierrors.NewNotFound("admission judgment binding", id)
	}
	if err != nil {
		return binding, err
	}
	if err := json.Unmarshal([]byte(data), &binding); err != nil {
		return binding, err
	}
	c := binding.Context
	if binding.ResultID != id || binding.ResultVersion != 1 || binding.ProjectID != project || c.RuleID != rule || c.RuleVersion != ruleRevision || c.WorkflowID != workflow || c.OwnerID != owner || c.Opportunity.ID != opportunity || c.Opportunity.Revision != revision {
		return binding, conflict("persisted admission binding identity differs")
	}
	return binding, nil
}

func (db *DB) SaveAdmissionBinding(ctx context.Context, binding judgment.AdmissionBinding) error {
	c := binding.Context
	if binding.ResultID == "" || binding.ResultVersion != 1 || binding.ProjectID == "" || binding.Actor.ID == "" || binding.Actor.Kind != "human" || c.Purpose != judgment.AdmissionPurpose || c.RuleID == "" || c.RuleVersion < 1 || c.OwnerID != binding.Actor.ID || c.WorkflowID == "" || c.Opportunity.ID == "" || c.Opportunity.Revision < 1 || binding.CreatedAt.IsZero() {
		return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "admission judgment binding requires actual fixed result, rule, opportunity and human workflow"}
	}
	conn, err := db.judgmentExecutor(ctx)
	if err != nil {
		return err
	}
	data, err := json.Marshal(binding)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, "INSERT INTO judgment_admission_bindings(result_id,rule_id,rule_revision,project_id,workflow_id,owner_id,opportunity_id,opportunity_revision,data,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)", binding.ResultID, c.RuleID, c.RuleVersion, binding.ProjectID, c.WorkflowID, c.OwnerID, c.Opportunity.ID, c.Opportunity.Revision, string(data), binding.CreatedAt.Format(time.RFC3339Nano))
	return err
}
