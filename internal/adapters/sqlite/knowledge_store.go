package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeRepository struct{ base *EngineRepository }

func NewKnowledgeRepository(base *EngineRepository) *KnowledgeRepository {
	return &KnowledgeRepository{base: base}
}

var _ app.KnowledgeRepository = (*KnowledgeRepository)(nil)

func (r *KnowledgeRepository) WithKnowledgeTx(ctx context.Context, fn func(app.KnowledgeTx) error) error {
	return r.base.WithTx(ctx, func(tx EngineTx) error { return fn(BindKnowledgeTx(tx)) })
}

func BindKnowledgeTx(tx EngineTx) app.KnowledgeTx { return &knowledgeTx{tx: tx} }

type knowledgeTx struct{ tx EngineTx }

func (t *knowledgeTx) LoadKnowledge(id string) (app.StoredKnowledge, error) {
	v := app.StoredKnowledge{CrystallizedVersions: []app.KnowledgeVersion{}, Shares: []app.ShareAudit{}, Actions: []app.ActionAudit{}}
	var data string
	err := t.tx.QueryRow(`SELECT data,owner_project,version,created_at,updated_at FROM attention_engine_knowledge WHERE id=?`, id).
		Scan(&data, &v.OwnerProject, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("knowledge", id)
	}
	if err != nil {
		return v, err
	}
	if err = json.Unmarshal([]byte(data), &v.Entry); err != nil {
		return v, err
	}
	if bytes.Equal(v.Entry.Crystallized, []byte("null")) {
		v.Entry.Crystallized = nil
	}
	rows, err := t.tx.Query(`SELECT version,content,recorded_at FROM attention_engine_knowledge_versions WHERE entry_id=? ORDER BY version`, id)
	if err != nil {
		return v, err
	}
	for rows.Next() {
		var version app.KnowledgeVersion
		var content string
		if err = rows.Scan(&version.Version, &content, &version.RecordedAt); err != nil {
			rows.Close()
			return v, err
		}
		version.Content = json.RawMessage(content)
		v.CrystallizedVersions = append(v.CrystallizedVersions, version)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return v, err
	}
	rows, err = t.tx.Query(`SELECT owner_project,target_project,actor_id,actor_kind,ts,request_id FROM attention_engine_knowledge_shares WHERE entry_id=? ORDER BY target_project`, id)
	if err != nil {
		return v, err
	}
	for rows.Next() {
		var share app.ShareAudit
		if err = rows.Scan(&share.OwnerProject, &share.TargetProject, &share.ActorID, &share.ActorKind, &share.TS, &share.RequestID); err != nil {
			rows.Close()
			return v, err
		}
		v.Shares = append(v.Shares, share)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return v, err
	}
	rows, err = t.tx.Query(`SELECT action,actor_id,actor_kind,ts,request_id,version FROM attention_engine_knowledge_actions WHERE entry_id=? ORDER BY ts,request_id`, id)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	for rows.Next() {
		var action app.ActionAudit
		if err = rows.Scan(&action.Action, &action.ActorID, &action.ActorKind, &action.TS, &action.RequestID, &action.Version); err != nil {
			return v, err
		}
		v.Actions = append(v.Actions, action)
	}
	return v, rows.Err()
}

func (t *knowledgeTx) ListKnowledge(project string, confirmedOnly bool) ([]app.StoredKnowledge, error) {
	query := `SELECT k.id FROM attention_engine_knowledge k WHERE k.owner_project!='' AND (k.owner_project=? OR EXISTS(SELECT 1 FROM attention_engine_knowledge_shares s WHERE s.entry_id=k.id AND s.target_project=?))`
	if confirmedOnly {
		query += ` AND k.status='confirmed'`
	}
	rows, err := t.tx.Query(query+` ORDER BY k.id`, project, project)
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
	result := make([]app.StoredKnowledge, 0, len(ids))
	for _, id := range ids {
		v, err := t.LoadKnowledge(id)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (t *knowledgeTx) SaveKnowledge(v app.StoredKnowledge, expected int) error {
	if v.OwnerProject == "" || v.Entry.ID == "" || v.Version != expected+1 {
		return domain.ErrInvalid
	}
	previous := app.StoredKnowledge{}
	if expected > 0 {
		var err error
		previous, err = t.LoadKnowledge(v.Entry.ID)
		if err != nil {
			return err
		}
		if previous.Version != expected || previous.OwnerProject != v.OwnerProject || previous.CreatedAt != v.CreatedAt {
			return domain.ErrVersion
		}
	}
	if len(v.CrystallizedVersions) < len(previous.CrystallizedVersions) {
		return domain.ErrVersion
	}
	for i, version := range v.CrystallizedVersions {
		if version.Version != i+1 || !json.Valid(version.Content) || version.RecordedAt <= 0 {
			return domain.ErrInvalid
		}
		if i > 0 && version.RecordedAt <= v.CrystallizedVersions[i-1].RecordedAt {
			return domain.ErrVersion
		}
		if i < len(previous.CrystallizedVersions) {
			old := previous.CrystallizedVersions[i]
			if old.RecordedAt != version.RecordedAt || !bytes.Equal(old.Content, version.Content) {
				return domain.ErrVersion
			}
		}
	}
	if len(v.CrystallizedVersions) == 0 {
		if len(v.Entry.Crystallized) != 0 {
			return domain.ErrVersion
		}
	} else if !bytes.Equal(v.Entry.Crystallized, v.CrystallizedVersions[len(v.CrystallizedVersions)-1].Content) {
		return domain.ErrVersion
	}
	data, err := json.Marshal(v.Entry)
	if err != nil {
		return err
	}
	if expected == 0 {
		_, err = t.tx.Exec(`INSERT INTO attention_engine_knowledge(id,topic,status,data,owner_project,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`,
			v.Entry.ID, v.Entry.Topic, v.Entry.Status, string(data), v.OwnerProject, v.Version, v.CreatedAt, v.UpdatedAt)
	} else {
		var result sql.Result
		result, err = t.tx.Exec(`UPDATE attention_engine_knowledge SET topic=?,status=?,data=?,version=?,updated_at=? WHERE id=? AND version=?`,
			v.Entry.Topic, v.Entry.Status, string(data), v.Version, v.UpdatedAt, v.Entry.ID, expected)
		if err == nil {
			var count int64
			count, err = result.RowsAffected()
			if err == nil && count != 1 {
				return domain.ErrVersion
			}
		}
	}
	if err != nil {
		return storageError(err)
	}
	for _, version := range v.CrystallizedVersions[len(previous.CrystallizedVersions):] {
		if _, err = t.tx.Exec(`INSERT INTO attention_engine_knowledge_versions(entry_id,version,content,recorded_at) VALUES(?,?,?,?)`, v.Entry.ID, version.Version, string(version.Content), version.RecordedAt); err != nil {
			return err
		}
	}
	return nil
}

func (t *knowledgeTx) AddKnowledgeShare(id string, v app.ShareAudit) error {
	_, err := t.tx.Exec(`INSERT INTO attention_engine_knowledge_shares(entry_id,owner_project,target_project,actor_id,actor_kind,ts,request_id) VALUES(?,?,?,?,?,?,?)`, id, v.OwnerProject, v.TargetProject, v.ActorID, v.ActorKind, v.TS, v.RequestID)
	return storageError(err)
}

func (t *knowledgeTx) AddKnowledgeAction(id string, v app.ActionAudit) error {
	_, err := t.tx.Exec(`INSERT INTO attention_engine_knowledge_actions(entry_id,action,actor_id,actor_kind,ts,request_id,version) VALUES(?,?,?,?,?,?,?)`, id, v.Action, v.ActorID, v.ActorKind, v.TS, v.RequestID, v.Version)
	return storageError(err)
}

func (t *knowledgeTx) KnowledgeAttention(topic string) (domain.AttentionNode, error) {
	var v domain.AttentionNode
	var data string
	err := t.tx.QueryRow(`SELECT data FROM attention_engine_nodes WHERE topic=?`, topic).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("attention_node", topic)
	}
	if err == nil {
		err = json.Unmarshal([]byte(data), &v)
	}
	return v, err
}

func (t *knowledgeTx) KnowledgeSource(id string) (app.StoredEvent, error) {
	v := app.StoredEvent{ID: id}
	var data, observation, status string
	var origin sql.NullString
	var received sql.NullInt64
	err := t.tx.QueryRow(`SELECT data,observation,origin_key,received_at,raw_bytes,mapping_status FROM attention_engine_events WHERE id=?`, id).
		Scan(&data, &observation, &origin, &received, &v.RawBytes, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("event", id)
	}
	if err != nil {
		return v, err
	}
	if err = json.Unmarshal([]byte(observation), &v.Observation); err != nil {
		return v, err
	}
	v.OriginKey, v.ReceivedAt = origin.String, received.Int64
	if status == "mapped" {
		v.Event = &domain.Event{}
		err = json.Unmarshal([]byte(data), v.Event)
	}
	return v, err
}

func (t *knowledgeTx) LoadKnowledgeReceipt(caller, command, key string) (app.Receipt, error) {
	var v app.Receipt
	var result string
	err := t.tx.QueryRow(`SELECT digest,result FROM attention_receipts WHERE caller=? AND command=? AND key=?`, caller, command, key).Scan(&v.Digest, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("receipt", key)
	}
	v.Result = json.RawMessage(result)
	return v, err
}

func (t *knowledgeTx) SaveKnowledgeReceipt(caller, command, key string, v app.Receipt) error {
	_, err := t.tx.Exec(`INSERT INTO attention_receipts(caller,command,key,digest,result) VALUES(?,?,?,?,?)`, caller, command, key, v.Digest, string(v.Result))
	return storageError(err)
}
