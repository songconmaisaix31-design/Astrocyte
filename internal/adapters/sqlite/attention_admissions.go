package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type AdmissionRepository struct{ db *DB }

func NewAdmissionRepository(db *DB) *AdmissionRepository { return &AdmissionRepository{db: db} }

var _ attention.AdmissionRepository = (*AdmissionRepository)(nil)

// 来源、准入、回执、事件和草稿共用已经存在的同数据库事务。
func (r *AdmissionRepository) WithAdmissionTx(ctx context.Context, fn func(attention.AdmissionTx) error) error {
	return NewWorkspaceRepository(r.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		conn, err := WorkspaceTransaction(tx.Context(), r.db)
		if err != nil {
			return err
		}
		if conn == nil {
			return errors.New("admission requires the current owned SQLite transaction")
		}
		return fn(&admissionTx{attentionTx: &attentionTx{ctx: tx.Context(), conn: conn}})
	})
}

type admissionTx struct{ *attentionTx }

func (t *admissionTx) Context() context.Context { return t.ctx }

func (t *admissionTx) LoadAdmission(id string) (domain.Admission, error) {
	var admission domain.Admission
	if err := t.load("SELECT data FROM attention_admissions WHERE id=?", "admission", id, []any{id}, &admission); err != nil {
		return admission, err
	}
	if admission.ID != id {
		return admission, errors.New("admission identity differs from its stored key")
	}
	return admission, nil
}

func (t *admissionTx) FindAdmission(id string, revision int) (domain.Admission, bool, error) {
	var data string
	err := t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_admissions WHERE opportunity_id=? AND opportunity_revision=?", id, revision).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Admission{}, false, nil
	}
	if err != nil {
		return domain.Admission{}, false, err
	}
	var admission domain.Admission
	if err := json.Unmarshal([]byte(data), &admission); err != nil {
		return admission, false, err
	}
	if admission.OpportunityID != id || admission.OpportunityRevision != revision {
		return admission, false, errors.New("admission differs from its stored opportunity revision")
	}
	return admission, true, nil
}

func (t *admissionTx) SaveAdmission(admission domain.Admission, expected int) error {
	if admission.Version != expected+1 || admission.ID == "" || admission.AdmittedEventID == "" || admission.OwnerNamespace == "" {
		return conflict("admission must advance its actual version exactly once")
	}
	data, err := json.Marshal(admission)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_admissions(id,opportunity_id,opportunity_revision,version,admitted_event_id,owner_namespace,data) VALUES(?,?,?,?,?,?,?)", admission.ID, admission.OpportunityID, admission.OpportunityRevision, admission.Version, admission.AdmittedEventID, admission.OwnerNamespace, string(data))
	} else {
		result, err = t.conn.ExecContext(t.ctx, "UPDATE attention_admissions SET version=?,data=? WHERE id=? AND opportunity_id=? AND opportunity_revision=? AND admitted_event_id=? AND owner_namespace=? AND version=?", admission.Version, string(data), admission.ID, admission.OpportunityID, admission.OpportunityRevision, admission.AdmittedEventID, admission.OwnerNamespace, expected)
	}
	return workspaceCAS(result, err)
}

func (t *admissionTx) LoadAdmissionEvent(id string) (attention.OutboxEvent, error) {
	var result attention.OutboxEvent
	var timestamp, payload string
	err := t.conn.QueryRowContext(t.ctx, "SELECT id,type,aggregate_id,aggregate_version,occurred_at,correlation_id,causation_id,payload FROM attention_outbox WHERE id=?", id).Scan(&result.ID, &result.Type, &result.AggregateID, &result.AggregateVersion, &timestamp, &result.CorrelationID, &result.CausationID, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return result, apierrors.NewNotFound("admission event", id)
	}
	if err != nil {
		return result, err
	}
	result.OccurredAt, err = time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return result, err
	}
	if !json.Valid([]byte(payload)) {
		return result, errors.New("admission event payload is not valid JSON")
	}
	result.Payload = json.RawMessage(payload)
	return result, nil
}
