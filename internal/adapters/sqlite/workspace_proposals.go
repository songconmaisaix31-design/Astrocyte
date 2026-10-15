package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type WorkspaceRepository struct{ db *DB }

func NewWorkspaceRepository(db *DB) *WorkspaceRepository { return &WorkspaceRepository{db: db} }

type workspaceTransactionKey struct{}
type workspaceTransaction struct {
	ctx  context.Context
	conn *sql.Conn
	db   *DB
}

// WorkspaceTransaction 仅供同一数据库的跨上下文SQLite adapter复用。
func WorkspaceTransaction(ctx context.Context, db *DB) (*sql.Conn, error) {
	t, ok := ctx.Value(workspaceTransactionKey{}).(*workspaceTransaction)
	if !ok {
		return nil, nil
	}
	if t.db != db {
		return nil, errors.New("workspace transaction belongs to another database")
	}
	return t.conn, nil
}

func (r *WorkspaceRepository) WithWorkspaceTx(ctx context.Context, fn func(workspace.WorkspaceTx) error) (err error) {
	if existing, ok := ctx.Value(workspaceTransactionKey{}).(*workspaceTransaction); ok {
		if existing.db != r.db {
			return errors.New("workspace transaction cannot span databases")
		}
		return fn(existing)
	}
	conn, err := r.db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	t := &workspaceTransaction{conn: conn, db: r.db}
	t.ctx = context.WithValue(ctx, workspaceTransactionKey{}, t)
	if err = fn(t); err != nil {
		_, rollbackErr := conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
		return errors.Join(err, rollbackErr)
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		_, rollbackErr := conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
		return errors.Join(err, rollbackErr)
	}
	return nil
}

func (t *workspaceTransaction) Context() context.Context { return t.ctx }

func (t *workspaceTransaction) read(query, resource, id string, out any, args ...any) error {
	var data string
	err := t.conn.QueryRowContext(t.ctx, query, args...).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return apierrors.NewNotFound(resource, id)
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), out)
}

func workspaceCAS(result sql.Result, err error) error {
	if err != nil {
		return storageError(err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return conflict("workspace version changed")
	}
	return nil
}

func (t *workspaceTransaction) LoadProposal(id string, revision int) (p domain.Proposal, err error) {
	if revision == 0 {
		err = t.read("SELECT data FROM workspace_proposals WHERE id=?", "proposal", id, &p, id)
	} else {
		err = t.read("SELECT data FROM workspace_proposal_revisions WHERE id=? AND revision=?", "proposal revision", id, &p, id, revision)
	}
	return
}

func (t *workspaceTransaction) FindAdmission(id string, revision int) (p domain.Proposal, err error) {
	err = t.read("SELECT data FROM workspace_proposals WHERE opportunity_id=? AND opportunity_revision=?", "admission proposal", id, &p, id, revision)
	return
}

func (t *workspaceTransaction) SaveProposal(p domain.Proposal, expected int) error {
	if p.Version != expected+1 {
		return conflict("proposal version must advance once")
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = t.conn.ExecContext(t.ctx, "INSERT INTO workspace_proposals(id,version,revision,opportunity_id,opportunity_revision,data) VALUES(?,?,?,?,?,?) ON CONFLICT DO NOTHING", p.ID, p.Version, p.Revision, nullable(p.OpportunityID), p.OpportunityRevision, string(data))
	} else {
		result, err = t.conn.ExecContext(t.ctx, "UPDATE workspace_proposals SET version=?,revision=?,data=? WHERE id=? AND version=?", p.Version, p.Revision, string(data), p.ID, expected)
	}
	if err = workspaceCAS(result, err); err != nil {
		return err
	}
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO workspace_proposal_revisions(id,revision,data) VALUES(?,?,?) ON CONFLICT(id,revision) DO UPDATE SET data=excluded.data", p.ID, p.Revision, string(data))
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (t *workspaceTransaction) LoadProject(id string) (p domain.LocalProject, err error) {
	err = t.read("SELECT data FROM local_agent_projects WHERE id=?", "project", id, &p, id)
	return
}

func (t *workspaceTransaction) LoadProjectGrant(project, agent string) (g domain.ProjectGrant, err error) {
	err = t.read("SELECT data FROM local_agent_grants WHERE project_id=? AND agent_id=?", "project grant", project, &g, project, agent)
	return
}

func receiptIdentity(caller, command, key string) string {
	b, _ := json.Marshal([]string{caller, command, key})
	digest := sha256.Sum256(b)
	return "workspace:" + hex.EncodeToString(digest[:])
}

func (t *workspaceTransaction) LoadReceipt(caller, command, key string) (v workspace.WorkspaceReceipt, err error) {
	var result string
	err = t.conn.QueryRowContext(t.ctx, "SELECT request_digest,result_json FROM foundation_idempotency WHERE idempotency_key=? AND caller=? AND command_type=?", receiptIdentity(caller, command, key), caller, command).Scan(&v.Digest, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("workspace receipt", key)
	}
	if err != nil {
		return v, err
	}
	if !json.Valid([]byte(result)) {
		return v, errors.New("invalid workspace receipt JSON")
	}
	v.Result = json.RawMessage(result)
	return
}

func (t *workspaceTransaction) SaveReceipt(caller, command, key string, v workspace.WorkspaceReceipt) error {
	if !json.Valid(v.Result) {
		return errors.New("workspace receipt requires a complete result")
	}
	_, err := t.conn.ExecContext(t.ctx, "INSERT INTO foundation_idempotency(idempotency_key,caller,command_type,request_digest,result_code,result_json) VALUES(?,?,?,?,?,?)", receiptIdentity(caller, command, key), caller, command, v.Digest, "ok", string(v.Result))
	return storageError(err)
}

func (t *workspaceTransaction) AppendWorkspaceEvent(e workspace.WorkspaceEvent) error {
	if e.SchemaVersion != 1 || !json.Valid(e.Payload) {
		return errors.New("invalid workspace event schema or JSON")
	}
	_, err := t.conn.ExecContext(t.ctx, "INSERT INTO foundation_outbox(event_id,event_type,schema_version,aggregate_id,aggregate_version,occurred_at,correlation_id,causation_id,payload) VALUES(?,?,?,?,?,?,?,?,?)", e.ID, e.Type, e.SchemaVersion, e.AggregateID, e.AggregateVersion, e.OccurredAt.UTC().Format(time.RFC3339Nano), e.CorrelationID, e.CausationID, string(e.Payload))
	return err
}

func (r *WorkspaceRepository) ListWorkspaceProposals(ctx context.Context, project string) ([]domain.Proposal, error) {
	rows, err := r.db.conn.QueryContext(ctx, "SELECT data FROM workspace_proposals ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Proposal{}
	for rows.Next() {
		var data string
		var p domain.Proposal
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &p); err != nil {
			return nil, err
		}
		if project == "" || p.Content.Scope.ProjectID == project {
			result = append(result, p)
		}
	}
	return result, rows.Err()
}

func (r *WorkspaceRepository) ReadWorkspaceEvents(ctx context.Context, after string, limit int) ([]workspace.WorkspaceEvent, error) {
	if limit < 1 || limit > 1000 {
		return nil, errors.New("workspace event limit must be 1..1000")
	}
	var sequence int64
	if after != "" {
		if err := r.db.conn.QueryRowContext(ctx, "SELECT rowid FROM foundation_outbox WHERE event_id=?", after).Scan(&sequence); err != nil {
			return nil, err
		}
	}
	rows, err := r.db.conn.QueryContext(ctx, "SELECT event_id,event_type,schema_version,aggregate_id,aggregate_version,occurred_at,COALESCE(correlation_id,''),COALESCE(causation_id,''),payload FROM foundation_outbox WHERE rowid>? AND event_type LIKE 'workspace.%' ORDER BY rowid LIMIT ?", sequence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []workspace.WorkspaceEvent{}
	for rows.Next() {
		var e workspace.WorkspaceEvent
		var stamp, payload string
		if err = rows.Scan(&e.ID, &e.Type, &e.SchemaVersion, &e.AggregateID, &e.AggregateVersion, &stamp, &e.CorrelationID, &e.CausationID, &payload); err != nil {
			return nil, err
		}
		e.OccurredAt, err = time.Parse(time.RFC3339Nano, stamp)
		if err != nil {
			return nil, fmt.Errorf("workspace event time: %w", err)
		}
		if !json.Valid([]byte(payload)) {
			return nil, errors.New("invalid stored workspace event")
		}
		e.Payload = json.RawMessage(payload)
		result = append(result, e)
	}
	return result, rows.Err()
}
