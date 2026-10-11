package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type ExperienceRepository struct{ base *EngineRepository }

func NewExperienceRepository(db *DB) *ExperienceRepository {
	return &ExperienceRepository{base: NewEngineRepository(db)}
}

var _ app.ExperienceRepository = (*ExperienceRepository)(nil)

func (r *ExperienceRepository) WithExperienceTx(ctx context.Context, fn func(app.ExperienceTx) error) error {
	return r.base.WithTx(ctx, func(tx EngineTx) error { return fn(&experienceTx{EngineTx: tx}) })
}

type experienceTx struct{ EngineTx }

func (t *experienceTx) FindExperience(project, eventID string) (app.StoredExperience, bool, error) {
	var id string
	err := t.QueryRow(`SELECT id FROM attention_engine_experiences WHERE owner_project=? AND origin_event_id=?`, project, eventID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return app.StoredExperience{}, false, nil
	}
	if err != nil {
		return app.StoredExperience{}, false, err
	}
	record, err := t.LoadExperience(id)
	return record, err == nil, err
}

func (t *experienceTx) LoadExperience(id string) (app.StoredExperience, error) {
	var record app.StoredExperience
	var data string
	var origin sql.NullString
	err := t.QueryRow(`SELECT owner_project,origin_event_id,version,created_at,updated_at,data FROM attention_engine_experiences WHERE id=?`, id).
		Scan(&record.OwnerProject, &origin, &record.Version, &record.CreatedAt, &record.UpdatedAt, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return record, apierrors.NewNotFound("experience", id)
	}
	if err != nil {
		return record, err
	}
	record.OriginEventID = origin.String
	if err = json.Unmarshal([]byte(data), &record.Experience); err != nil {
		return record, err
	}
	record.Shares = []app.ShareAudit{}
	record.Actions = []app.ActionAudit{}
	rows, err := t.Query(`SELECT owner_project,target_project,actor_id,actor_kind,ts,request_id FROM attention_engine_experience_shares WHERE experience_id=? ORDER BY target_project`, id)
	if err != nil {
		return record, err
	}
	for rows.Next() {
		var audit app.ShareAudit
		if err = rows.Scan(&audit.OwnerProject, &audit.TargetProject, &audit.ActorID, &audit.ActorKind, &audit.TS, &audit.RequestID); err != nil {
			rows.Close()
			return record, err
		}
		record.Shares = append(record.Shares, audit)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return record, err
	}
	rows, err = t.Query(`SELECT action,actor_id,actor_kind,ts,request_id,version FROM attention_engine_experience_actions WHERE experience_id=? ORDER BY version`, id)
	if err != nil {
		return record, err
	}
	defer rows.Close()
	for rows.Next() {
		var audit app.ActionAudit
		if err = rows.Scan(&audit.Action, &audit.ActorID, &audit.ActorKind, &audit.TS, &audit.RequestID, &audit.Version); err != nil {
			return record, err
		}
		record.Actions = append(record.Actions, audit)
	}
	return record, rows.Err()
}

func (t *experienceTx) ListApprovedExperiences(project, pattern string) ([]app.StoredExperience, error) {
	// instr treats a supplied pattern literally; '%' and '_' do not widen scope.
	rows, err := t.Query(`SELECT e.id FROM attention_engine_experiences e
		WHERE e.status='approved' AND instr(e.pattern,?)>0
		AND (e.owner_project=? OR EXISTS(SELECT 1 FROM attention_engine_experience_shares s
		 WHERE s.experience_id=e.id AND s.owner_project=e.owner_project AND s.target_project=? AND s.actor_kind='human'))
		ORDER BY e.id`, pattern, project, project)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := make([]app.StoredExperience, 0, len(ids))
	for _, id := range ids {
		record, err := t.LoadExperience(id)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (t *experienceTx) InsertExperience(record app.StoredExperience) error {
	data, err := json.Marshal(record.Experience)
	if err != nil {
		return err
	}
	_, err = t.Exec(`INSERT INTO attention_engine_experiences
		(id,pattern,verdict,status,data,owner_project,origin_event_id,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, record.Experience.ID, record.Experience.Pattern, record.Experience.Verdict,
		record.Experience.Status, string(data), record.OwnerProject, record.OriginEventID, record.Version, record.CreatedAt, record.UpdatedAt)
	return storageError(err)
}

func (t *experienceTx) SaveExperience(record app.StoredExperience, expected int) error {
	data, err := json.Marshal(record.Experience)
	if err != nil {
		return err
	}
	result, err := t.Exec(`UPDATE attention_engine_experiences SET pattern=?,verdict=?,status=?,data=?,version=?,updated_at=?
		WHERE id=? AND owner_project=? AND origin_event_id=? AND version=?`, record.Experience.Pattern, record.Experience.Verdict,
		record.Experience.Status, string(data), record.Version, record.UpdatedAt, record.Experience.ID, record.OwnerProject, record.OriginEventID, expected)
	if err != nil {
		return storageError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return conflict("experience version changed")
	}
	return nil
}

func (t *experienceTx) AppendExperienceShare(id string, audit app.ShareAudit) error {
	_, err := t.Exec(`INSERT INTO attention_engine_experience_shares
		(experience_id,owner_project,target_project,actor_id,actor_kind,ts,request_id) VALUES(?,?,?,?,?,?,?)`,
		id, audit.OwnerProject, audit.TargetProject, audit.ActorID, audit.ActorKind, audit.TS, audit.RequestID)
	return storageError(err)
}

func (t *experienceTx) AppendExperienceAction(id string, audit app.ActionAudit) error {
	_, err := t.Exec(`INSERT INTO attention_engine_experience_actions
		(experience_id,action,actor_id,actor_kind,ts,request_id,version) VALUES(?,?,?,?,?,?,?)`,
		id, audit.Action, audit.ActorID, audit.ActorKind, audit.TS, audit.RequestID, audit.Version)
	return storageError(err)
}

// Experience commands reuse the existing receipt table in the same EngineTx
// as their head and audit writes. Operation names isolate them from S1 commands.
func (t *experienceTx) LoadExperienceReceipt(caller, command, key string) (app.Receipt, error) {
	var receipt app.Receipt
	var result string
	err := t.QueryRow(`SELECT digest,result FROM attention_receipts WHERE caller=? AND command=? AND key=?`, caller, command, key).
		Scan(&receipt.Digest, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return receipt, apierrors.NewNotFound("receipt", key)
	}
	receipt.Result = json.RawMessage(result)
	return receipt, err
}

func (t *experienceTx) SaveExperienceReceipt(caller, command, key string, receipt app.Receipt) error {
	if !json.Valid(receipt.Result) {
		return errors.New("experience receipt must contain a complete JSON result")
	}
	_, err := t.Exec(`INSERT INTO attention_receipts(caller,command,key,digest,result) VALUES(?,?,?,?,?)`,
		caller, command, key, receipt.Digest, string(receipt.Result))
	return storageError(err)
}
