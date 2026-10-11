package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type DynamicsRepository struct{ base *EngineRepository }

func NewDynamicsRepository(db *DB) *DynamicsRepository {
	return &DynamicsRepository{base: NewEngineRepository(db)}
}

var _ app.DynamicsStore = (*DynamicsRepository)(nil)

func (r *DynamicsRepository) WithDynamicsTx(ctx context.Context, fn func(app.DynamicsTx) error) error {
	return r.base.WithTx(ctx, func(t EngineTx) error { return fn(NewDynamicsTx(t)) })
}

// NewDynamicsTx composes W2 operations into an existing W0 transaction; W3's
// daily commit can therefore include attention, conflicts and its own rows.
func NewDynamicsTx(t EngineTx) app.DynamicsTx { return &dynamicsTx{t: t} }

func (r *DynamicsRepository) ViewDynamics(ctx context.Context, fn func(app.DynamicsTx) error) error {
	t, err := r.base.Conn().BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer t.Rollback()
	if err = fn(NewDynamicsTx(&dynamicsReadTx{ctx: ctx, tx: t})); err != nil {
		return err
	}
	return t.Commit()
}

type dynamicsReadTx struct {
	ctx context.Context
	tx  *sql.Tx
}

func (t *dynamicsReadTx) Exec(string, ...any) (sql.Result, error) {
	return nil, fmt.Errorf("dynamics view is read-only")
}
func (t *dynamicsReadTx) Query(q string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, q, args...)
}
func (t *dynamicsReadTx) QueryRow(q string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, q, args...)
}

type dynamicsTx struct{ t EngineTx }

func dynamicsMissing(err error, kind, id string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return apierrors.NewNotFound(kind, id)
	}
	return err
}

func (t *dynamicsTx) GetNode(topic string) (app.StoredAttentionNode, error) {
	var n app.StoredAttentionNode
	var data string
	err := t.t.QueryRow("SELECT data,updated_at,lifecycle,version FROM attention_engine_nodes WHERE topic=?", topic).Scan(&data, &n.UpdatedAt, &n.Lifecycle, &n.Version)
	if err != nil {
		return n, dynamicsMissing(err, "attention_topic", topic)
	}
	var saved app.StoredAttentionNode
	if err = json.Unmarshal([]byte(data), &saved); err != nil {
		return n, err
	}
	n.Node, n.History = saved.Node, saved.History
	if n.Node.ID == "" {
		err = json.Unmarshal([]byte(data), &n.Node)
	}
	return n, err
}

func (t *dynamicsTx) ListNodes() ([]app.StoredAttentionNode, error) {
	rows, err := t.t.Query("SELECT topic FROM attention_engine_nodes ORDER BY topic")
	if err != nil {
		return nil, err
	}
	topics := []string{}
	for rows.Next() {
		var topic string
		if err = rows.Scan(&topic); err != nil {
			rows.Close()
			return nil, err
		}
		topics = append(topics, topic)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	nodes := []app.StoredAttentionNode{}
	for _, topic := range topics {
		n, err := t.GetNode(topic)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (t *dynamicsTx) SaveNode(n app.StoredAttentionNode) error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_, err = t.t.Exec(`INSERT INTO attention_engine_nodes(id,topic,attention,momentum,confidence,last_active,data,updated_at,lifecycle,version)
	 VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(topic) DO UPDATE SET attention=excluded.attention,momentum=excluded.momentum,
	 confidence=excluded.confidence,last_active=excluded.last_active,data=excluded.data,updated_at=excluded.updated_at,lifecycle=excluded.lifecycle,version=excluded.version`,
		n.Node.ID, n.Node.Topic, n.Node.Attention, n.Node.Momentum, n.Node.Confidence, n.Node.LastActive, string(data), n.UpdatedAt, n.Lifecycle, n.Version)
	return err
}

func (t *dynamicsTx) GetEvent(id string) (app.StoredEvent, error) {
	var e app.StoredEvent
	var data, status string
	var origin sql.NullString
	var received sql.NullInt64
	var observation sql.NullString
	err := t.t.QueryRow("SELECT id,data,origin_key,received_at,raw_bytes,observation,mapping_status FROM attention_engine_events WHERE id=?", id).Scan(&e.ID, &data, &origin, &received, &e.RawBytes, &observation, &status)
	if err != nil {
		return e, dynamicsMissing(err, "event", id)
	}
	e.OriginKey, e.ReceivedAt = origin.String, received.Int64
	if observation.Valid {
		if err = json.Unmarshal([]byte(observation.String), &e.Observation); err != nil {
			return e, err
		}
	}
	if status == "mapped" {
		var core domain.Event
		if err = json.Unmarshal([]byte(data), &core); err != nil {
			return e, err
		}
		e.Event = &core
	}
	return e, nil
}

func (t *dynamicsTx) GetSignal(s domain.Signal) (domain.Signal, error) {
	var data string
	var saved domain.Signal
	err := t.t.QueryRow("SELECT data FROM attention_engine_signals WHERE event_id=? AND topic=? AND class=?", s.EventID, s.Topic, s.Class).Scan(&data)
	if err != nil {
		return saved, dynamicsMissing(err, "signal", s.EventID)
	}
	err = json.Unmarshal([]byte(data), &saved)
	return saved, err
}

func (t *dynamicsTx) GetConflict(id string) (app.DynamicsConflict, error) {
	var c app.DynamicsConflict
	var data string
	err := t.t.QueryRow("SELECT data,created_at,version FROM attention_engine_conflicts WHERE id=?", id).Scan(&data, &c.CreatedAt, &c.Version)
	if err != nil {
		return c, dynamicsMissing(err, "conflict", id)
	}
	if err = json.Unmarshal([]byte(data), &c.Conflict); err != nil {
		return c, err
	}
	rows, err := t.t.Query("SELECT action,actor_id,actor_kind,ts,request_id,version FROM attention_engine_conflict_actions WHERE conflict_id=? ORDER BY version", id)
	if err != nil {
		return c, err
	}
	defer rows.Close()
	c.Actions = []app.ActionAudit{}
	for rows.Next() {
		var a app.ActionAudit
		if err = rows.Scan(&a.Action, &a.ActorID, &a.ActorKind, &a.TS, &a.RequestID, &a.Version); err != nil {
			return c, err
		}
		c.Actions = append(c.Actions, a)
	}
	return c, rows.Err()
}

func (t *dynamicsTx) ListConflicts() ([]app.DynamicsConflict, error) {
	rows, err := t.t.Query("SELECT id FROM attention_engine_conflicts ORDER BY created_at,id")
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
	list := []app.DynamicsConflict{}
	for _, id := range ids {
		c, err := t.GetConflict(id)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (t *dynamicsTx) SaveConflict(c app.DynamicsConflict) error {
	data, err := json.Marshal(c.Conflict)
	if err != nil {
		return err
	}
	_, err = t.t.Exec(`INSERT INTO attention_engine_conflicts(id,topic,kind,status,data,created_at,version) VALUES(?,?,?,?,?,?,?)
	ON CONFLICT(id) DO UPDATE SET status=excluded.status,data=excluded.data,version=excluded.version`,
		c.Conflict.ID, c.Conflict.Topic, c.Conflict.Kind, c.Conflict.Status, string(data), c.CreatedAt, c.Version)
	return err
}

func (t *dynamicsTx) AddConflictAction(id string, a app.ActionAudit) error {
	_, err := t.t.Exec("INSERT INTO attention_engine_conflict_actions(conflict_id,action,actor_id,actor_kind,ts,request_id,version) VALUES(?,?,?,?,?,?,?)", id, a.Action, a.ActorID, a.ActorKind, a.TS, a.RequestID, a.Version)
	return err
}
