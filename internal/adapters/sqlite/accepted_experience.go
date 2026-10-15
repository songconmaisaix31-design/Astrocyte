package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type AcceptedExperienceRepository struct{ db *DB }

func NewAcceptedExperienceRepository(db *DB) *AcceptedExperienceRepository {
	return &AcceptedExperienceRepository{db: db}
}

var _ attention.AcceptedExperienceRepository = (*AcceptedExperienceRepository)(nil)

// 复用既有经验存储和真实业务事务，不修改冻结的引擎算法或旧迁移。
func (r *AcceptedExperienceRepository) WithAcceptedExperienceTx(ctx context.Context, fn func(attention.AcceptedExperienceTx) error) error {
	if r.db == nil {
		return errors.New("actual experience database is required")
	}
	return NewWorkspaceRepository(r.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		conn, err := WorkspaceTransaction(tx.Context(), r.db)
		if err != nil {
			return err
		}
		if conn == nil {
			return errors.New("experience source requires the current owned SQLite transaction")
		}
		return fn(&acceptedExperienceTx{experienceTx: &experienceTx{EngineTx: &engineTx{ctx: tx.Context(), conn: conn}}, ctx: tx.Context()})
	})
}

func (r *AcceptedExperienceRepository) WithExperienceTx(ctx context.Context, fn func(attention.ExperienceTx) error) error {
	return r.WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error { return fn(tx) })
}

func (r *AcceptedExperienceRepository) ReadAcceptedExperienceSource(ctx context.Context, id string) (attention.AcceptedExperienceSource, error) {
	var result attention.AcceptedExperienceSource
	err := r.WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
		var err error
		result, err = tx.LoadAcceptedExperienceSource(id)
		return err
	})
	return result, err
}

type acceptedExperienceTx struct {
	*experienceTx
	ctx context.Context
}

func (t *acceptedExperienceTx) Context() context.Context { return t.ctx }

func (t *acceptedExperienceTx) LoadAcceptedExperienceSource(id string) (attention.AcceptedExperienceSource, error) {
	var source attention.AcceptedExperienceSource
	var kind, eventID, project, localProject, data string
	err := t.QueryRow("SELECT source_kind,event_id,owner_project,local_project_id,data FROM experience_completion_sources WHERE experience_id=?", id).Scan(&kind, &eventID, &project, &localProject, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return source, apierrors.NewNotFound("experience completion source", id)
	}
	if err != nil {
		return source, err
	}
	if err := json.Unmarshal([]byte(data), &source); err != nil {
		return source, err
	}
	if source.Completion.SourceKind != kind || source.Completion.EventID != eventID || source.Completion.OwnerProject != project || source.Completion.LocalProjectID != localProject {
		return source, errors.New("experience source differs from its stored immutable identity")
	}
	return source, nil
}

func (t *acceptedExperienceTx) LoadExperience(id string) (attention.StoredExperience, error) {
	record, err := t.experienceTx.LoadExperience(id)
	if err != nil {
		return record, err
	}
	source, err := t.LoadAcceptedExperienceSource(id)
	var missing *apierrors.ServiceError
	if errors.As(err, &missing) && missing.Code == apierrors.NotFound {
		if record.OriginEventID == "" {
			return record, errors.New("experience has neither an original event nor a completion source")
		}
		return record, nil
	}
	if err != nil {
		return record, err
	}
	if record.OwnerProject != source.Completion.OwnerProject || record.OriginEventID != "" || len(record.Experience.Evidence) != 1 || record.Experience.Evidence[0].EventID != source.Completion.EventID || record.Experience.Evidence[0].TS != source.Completion.OccurredAt {
		return record, errors.New("experience candidate differs from its fixed completion source")
	}
	record.SourceKind = source.Completion.SourceKind
	record.OriginEventID = source.Completion.EventID
	return record, nil
}

func (t *acceptedExperienceTx) FindExperience(project, eventID string) (attention.StoredExperience, bool, error) {
	rows, err := t.Query(`SELECT e.id FROM attention_engine_experiences e WHERE e.owner_project=?
		AND (e.origin_event_id=? OR EXISTS(SELECT 1 FROM experience_completion_sources s
		WHERE s.experience_id=e.id AND s.owner_project=e.owner_project AND s.event_id=?))`, project, eventID, eventID)
	if err != nil {
		return attention.StoredExperience{}, false, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return attention.StoredExperience{}, false, errors.Join(err, rows.Close())
		}
		ids = append(ids, id)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return attention.StoredExperience{}, false, err
	}
	if len(ids) == 0 {
		return attention.StoredExperience{}, false, nil
	}
	if len(ids) != 1 {
		return attention.StoredExperience{}, false, errors.New("original receipt resolves to multiple experience candidates")
	}
	row, err := t.LoadExperience(ids[0])
	return row, err == nil, err
}

func (t *acceptedExperienceTx) InsertExperience(record attention.StoredExperience) error {
	if record.SourceKind == "" {
		return t.experienceTx.InsertExperience(record)
	}
	if record.SourceKind != attention.ExperienceSourceAcceptedArtifact && record.SourceKind != attention.ExperienceSourceRelayCompletion {
		return errors.New("unsupported actual completion source kind")
	}
	data, err := json.Marshal(record.Experience)
	if err != nil {
		return err
	}
	// 业务收据保持自己的来源身份；冻结的原生事件外键留空。
	_, err = t.Exec(`INSERT INTO attention_engine_experiences
		(id,pattern,verdict,status,data,owner_project,origin_event_id,version,created_at,updated_at)
		VALUES(?,?,?,?,?,?,NULL,?,?,?)`, record.Experience.ID, record.Experience.Pattern, record.Experience.Verdict,
		record.Experience.Status, string(data), record.OwnerProject, record.Version, record.CreatedAt, record.UpdatedAt)
	return storageError(err)
}

func (t *acceptedExperienceTx) SaveExperience(record attention.StoredExperience, expected int) error {
	if record.SourceKind == "" {
		return t.experienceTx.SaveExperience(record, expected)
	}
	stored, err := t.LoadExperience(record.Experience.ID)
	if err != nil {
		return err
	}
	if stored.SourceKind != record.SourceKind || stored.OriginEventID != record.OriginEventID || stored.OwnerProject != record.OwnerProject {
		return errors.New("experience governance changed the immutable completion identity")
	}
	if len(record.Experience.Evidence) != 1 || len(stored.Experience.Evidence) != 1 || record.Experience.Evidence[0] != stored.Experience.Evidence[0] {
		return errors.New("experience governance changed the fixed completion evidence")
	}
	data, err := json.Marshal(record.Experience)
	if err != nil {
		return err
	}
	result, err := t.Exec(`UPDATE attention_engine_experiences SET pattern=?,verdict=?,status=?,data=?,version=?,updated_at=?
		WHERE id=? AND owner_project=? AND origin_event_id IS NULL AND version=?
		AND EXISTS(SELECT 1 FROM experience_completion_sources s WHERE s.experience_id=attention_engine_experiences.id
		AND s.source_kind=? AND s.event_id=? AND s.owner_project=attention_engine_experiences.owner_project)`,
		record.Experience.Pattern, record.Experience.Verdict, record.Experience.Status, string(data), record.Version,
		record.UpdatedAt, record.Experience.ID, record.OwnerProject, expected, record.SourceKind, record.OriginEventID)
	return workspaceCAS(result, err)
}

func (t *acceptedExperienceTx) ListApprovedExperiences(project, pattern string) ([]attention.StoredExperience, error) {
	rows, err := t.experienceTx.ListApprovedExperiences(project, pattern)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i], err = t.LoadExperience(rows[i].Experience.ID)
		if err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (t *acceptedExperienceTx) SaveAcceptedExperienceSource(id string, source attention.AcceptedExperienceSource) error {
	record, err := t.experienceTx.LoadExperience(id)
	if err != nil {
		return err
	}
	completion := source.Completion
	if record.Experience.Status != domain.StatusQuarantine || record.Version != 1 || record.OriginEventID != "" || completion.EventID == "" || completion.OwnerProject != record.OwnerProject || completion.LocalProjectID == "" || completion.Version < 1 || completion.OccurredAt < 1 || len(record.Experience.Evidence) != 1 || record.Experience.Evidence[0].EventID != completion.EventID || record.Experience.Evidence[0].TS != completion.OccurredAt {
		return errors.New("completion source must bind the actual new quarantined candidate")
	}
	if completion.SourceKind != attention.ExperienceSourceAcceptedArtifact && completion.SourceKind != attention.ExperienceSourceRelayCompletion {
		return errors.New("unsupported actual completion source kind")
	}
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	_, err = t.Exec("INSERT INTO experience_completion_sources(experience_id,source_kind,event_id,owner_project,local_project_id,data) VALUES(?,?,?,?,?,?)", id, completion.SourceKind, completion.EventID, completion.OwnerProject, completion.LocalProjectID, string(data))
	return storageError(err)
}
