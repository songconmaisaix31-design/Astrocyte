package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	modernsqlite "modernc.org/sqlite"
)

type SwarmSQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type SwarmRepository struct {
	db *DB
}

type swarmTransaction struct {
	ctx  context.Context
	exec SwarmSQLExecutor
	db   *DB
}
type swarmTransactionKey struct{}

var _ swarmapp.Repository = (*SwarmRepository)(nil)

func NewSwarmRepository(db *DB) (*SwarmRepository, error) {
	if db == nil {
		return nil, domain.Failure(apierrors.ValidationFailed, "SQLite database required")
	}
	return &SwarmRepository{db: db}, nil
}

func (r *SwarmRepository) Within(ctx context.Context, fn func(swarmapp.Transaction) error) (err error) {
	if tx, ok := ctx.Value(swarmTransactionKey{}).(*swarmTransaction); ok {
		if tx.db != r.db {
			return domain.Failure(apierrors.ScopeDenied, "Swarm transaction belongs to another database")
		}
		return fn(tx)
	}
	conn, err := WorkspaceTransaction(ctx, r.db)
	if err != nil {
		return err
	}
	if conn != nil {
		tx := &swarmTransaction{ctx: ctx, exec: conn, db: r.db}
		tx.ctx = context.WithValue(ctx, swarmTransactionKey{}, tx)
		return fn(tx)
	}
	conn, err = r.db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, rollbackErr := conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
			err = errors.Join(err, rollbackErr)
		}
	}()
	tx := &swarmTransaction{ctx: ctx, exec: conn, db: r.db}
	tx.ctx = context.WithValue(ctx, swarmTransactionKey{}, tx)
	if err = fn(tx); err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}

func swarmDecode[T any](row *sql.Row, kind, id string) (T, error) {
	var out T
	var data string
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, apierrors.NewNotFound(kind, id)
		}
		return out, err
	}
	err := json.Unmarshal([]byte(data), &out)
	return out, err
}
func swarmRows[T any](rows *sql.Rows) ([]T, error) {
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		var data string
		var v T
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(data), &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func swarmConstraint(err error) error {
	var e *modernsqlite.Error
	if errors.As(err, &e) && e.Code()&255 == 19 {
		return domain.Failure(apierrors.VersionConflict, "concurrent Swarm identity or lease conflict")
	}
	return err
}
func swarmCAS(result sql.Result, err error) error {
	if err != nil {
		return swarmConstraint(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return domain.Failure(apierrors.VersionConflict, "Swarm compare-and-swap failed")
	}
	return nil
}

func (r *SwarmRepository) List(ctx context.Context) ([]domain.Mission, error) {
	rows, err := r.db.conn.QueryContext(ctx, "SELECT data FROM swarm_missions ORDER BY id")
	if err != nil {
		return nil, err
	}
	return swarmRows[domain.Mission](rows)
}
func (r *SwarmRepository) Load(ctx context.Context, id string) (domain.Mission, error) {
	return swarmDecode[domain.Mission](r.db.conn.QueryRowContext(ctx, "SELECT data FROM swarm_missions WHERE id=?", id), "mission", id)
}

func (r *SwarmRepository) Timeline(ctx context.Context, id string) ([]swarmapp.TimelineEvent, error) {
	var exec SwarmSQLExecutor = r.db.conn
	if tx, ok := ctx.Value(swarmTransactionKey{}).(*swarmTransaction); ok {
		if tx.db != r.db {
			return nil, domain.Failure(apierrors.ScopeDenied, "timeline transaction belongs to another database")
		}
		exec = tx.exec
	}
	rows, err := exec.QueryContext(ctx, "SELECT event_id,event_type,aggregate_id,aggregate_version,occurred_at FROM foundation_outbox WHERE event_type LIKE 'swarm.%' AND (aggregate_id=? OR aggregate_id IN (SELECT id FROM swarm_work_items WHERE mission_id=?) OR aggregate_id IN (SELECT a.id FROM swarm_artifacts a JOIN swarm_work_items w ON a.work_item_id=w.id WHERE w.mission_id=?)) ORDER BY occurred_at,event_id", id, id, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []swarmapp.TimelineEvent{}
	for rows.Next() {
		var e swarmapp.TimelineEvent
		var at string
		if err = rows.Scan(&e.ID, &e.Type, &e.AggregateID, &e.Version, &at); err != nil {
			return nil, err
		}
		e.OccurredAt, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (t *swarmTransaction) Mission(id string) (domain.Mission, error) {
	return swarmDecode[domain.Mission](t.exec.QueryRowContext(t.ctx, "SELECT data FROM swarm_missions WHERE id=?", id), "mission", id)
}
func (t *swarmTransaction) MissionByGrant(id string) (domain.Mission, error) {
	return swarmDecode[domain.Mission](t.exec.QueryRowContext(t.ctx, "SELECT data FROM swarm_missions WHERE grant_id=?", id), "mission grant", id)
}
func (t *swarmTransaction) SaveMission(m domain.Mission, expected int) error {
	if m.Version != expected+1 {
		return domain.Failure(apierrors.VersionConflict, "mission version must increase by one")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if expected == 0 {
		return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO swarm_missions(id,grant_id,project_id,version,data) VALUES(?,?,?,?,?)", m.ID, m.Grant.ID, m.Grant.ProjectID, m.Version, string(b)))
	}
	return swarmCAS(t.exec.ExecContext(t.ctx, "UPDATE swarm_missions SET version=?,data=? WHERE id=? AND version=? AND grant_id=? AND project_id=?", m.Version, string(b), m.ID, expected, m.Grant.ID, m.Grant.ProjectID))
}
func (t *swarmTransaction) WorkItem(id string) (domain.WorkItem, error) {
	return swarmDecode[domain.WorkItem](t.exec.QueryRowContext(t.ctx, "SELECT data FROM swarm_work_items WHERE id=?", id), "work item", id)
}
func (t *swarmTransaction) WorkItems(id string) ([]domain.WorkItem, error) {
	rows, err := t.exec.QueryContext(t.ctx, "SELECT data FROM swarm_work_items WHERE mission_id=? ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	return swarmRows[domain.WorkItem](rows)
}
func (t *swarmTransaction) SaveWorkItem(w domain.WorkItem, expected int) error {
	if w.Version != expected+1 {
		return domain.Failure(apierrors.VersionConflict, "work item version must increase by one")
	}
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	if expected == 0 {
		return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO swarm_work_items(id,mission_id,version,fence,data) VALUES(?,?,?,?,?)", w.ID, w.MissionID, w.Version, w.Fence, string(b)))
	}
	return swarmCAS(t.exec.ExecContext(t.ctx, "UPDATE swarm_work_items SET version=?,fence=?,data=? WHERE id=? AND version=? AND mission_id=? AND fence<=?", w.Version, w.Fence, string(b), w.ID, expected, w.MissionID, w.Fence))
}
func (t *swarmTransaction) Lease(id string) (domain.Lease, error) {
	return swarmDecode[domain.Lease](t.exec.QueryRowContext(t.ctx, "SELECT data FROM swarm_leases WHERE work_item_id=?", id), "lease", id)
}
func (t *swarmTransaction) ActiveLeases(project string) ([]domain.Lease, error) {
	rows, err := t.exec.QueryContext(t.ctx, "SELECT data FROM swarm_leases WHERE project_id=? ORDER BY work_item_id", project)
	if err != nil {
		return nil, err
	}
	return swarmRows[domain.Lease](rows)
}
func (t *swarmTransaction) SaveLease(l domain.Lease) error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	// 过期或未知执行者仍保留占用，只有原生停止确认后才删除。
	return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO swarm_leases(work_item_id,mission_id,project_id,holder_id,fence,operation_id,data) VALUES(?,?,?,?,?,?,?)", l.WorkItemID, l.MissionID, l.ProjectID, l.HolderID, l.Fence, l.OperationID, string(b)))
}
func (t *swarmTransaction) DeleteLease(id string, fence int64) error {
	return swarmCAS(t.exec.ExecContext(t.ctx, "DELETE FROM swarm_leases WHERE work_item_id=? AND fence=?", id, fence))
}
func (t *swarmTransaction) Artifact(id string) (domain.Artifact, error) {
	return swarmDecode[domain.Artifact](t.exec.QueryRowContext(t.ctx, "SELECT data FROM swarm_artifacts WHERE id=?", id), "artifact", id)
}

func (t *swarmTransaction) ArtifactVersion(id string, version int) (domain.Artifact, error) {
	var out domain.Artifact
	rows, err := t.exec.QueryContext(t.ctx, "SELECT o.payload FROM foundation_outbox o JOIN swarm_artifacts a ON a.id=o.aggregate_id WHERE o.aggregate_id=? AND o.aggregate_version=? AND o.event_type IN ('swarm.artifact_submitted','swarm.artifact_accepted','swarm.artifact_revision_requested','swarm.artifact_experience_linked')", id, version)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		if found {
			return out, domain.Failure(apierrors.VersionConflict, "artifact version has conflicting durable snapshots")
		}
		var data string
		if err = rows.Scan(&data); err != nil {
			return out, err
		}
		if err = json.Unmarshal([]byte(data), &out); err != nil {
			return out, err
		}
		found = true
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	if !found {
		return out, apierrors.NewNotFound("artifact version", id)
	}
	if out.ID != id || out.Version != version {
		return out, domain.Failure(apierrors.ContextStale, "artifact snapshot identity does not match its event")
	}
	return out, nil
}
func (t *swarmTransaction) SaveArtifact(a domain.Artifact, expected int) error {
	if a.Version != expected+1 {
		return domain.Failure(apierrors.VersionConflict, "artifact version must increase by one")
	}
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if expected == 0 {
		return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO swarm_artifacts(id,work_item_id,version,data) VALUES(?,?,?,?)", a.ID, a.WorkItemID, a.Version, string(b)))
	}
	return swarmCAS(t.exec.ExecContext(t.ctx, "UPDATE swarm_artifacts SET version=?,data=? WHERE id=? AND version=? AND work_item_id=?", a.Version, string(b), a.ID, expected, a.WorkItemID))
}
func (t *swarmTransaction) Receipt(caller, command, key string) (swarmapp.Receipt, bool, error) {
	var r swarmapp.Receipt
	var result string
	err := t.exec.QueryRowContext(t.ctx, "SELECT request_digest,result_code FROM foundation_idempotency WHERE caller=? AND command_type=? AND idempotency_key=?", caller, "swarm."+command, key).Scan(&r.Digest, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	r.Result = []byte(result)
	return r, true, nil
}
func (t *swarmTransaction) SaveReceipt(caller, command, key string, r swarmapp.Receipt) error {
	if !json.Valid(r.Result) {
		return fmt.Errorf("Swarm receipt must contain valid JSON")
	}
	return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO foundation_idempotency(caller,command_type,idempotency_key,request_digest,result_code) VALUES(?,?,?,?,?)", caller, "swarm."+command, key, r.Digest, string(r.Result)))
}
func (t *swarmTransaction) Event(e swarmapp.Event) error {
	b, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}
	return swarmCAS(t.exec.ExecContext(t.ctx, "INSERT INTO foundation_outbox(event_id,event_type,schema_version,aggregate_id,aggregate_version,occurred_at,correlation_id,payload) VALUES(?,?,1,?,?,?,?,?)", e.ID, e.Type, e.AggregateID, e.Version, e.At.UTC().Format(time.RFC3339Nano), e.RequestID, string(b)))
}
func (t *swarmTransaction) Consume(consumer, event string) (bool, error) {
	r, err := t.exec.ExecContext(t.ctx, "INSERT INTO swarm_consumed_events(consumer_id,event_id) VALUES(?,?) ON CONFLICT(consumer_id,event_id) DO NOTHING", consumer, event)
	if err != nil {
		return false, err
	}
	n, err := r.RowsAffected()
	return n == 1, err
}
