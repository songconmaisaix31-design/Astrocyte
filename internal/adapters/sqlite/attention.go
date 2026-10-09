package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// AttentionRepository owns only attention_* tables. BEGIN IMMEDIATE serializes
// read/modify/write callbacks across independent database connections/processes.
type AttentionRepository struct{ db *DB }

func NewAttentionRepository(db *DB) *AttentionRepository { return &AttentionRepository{db: db} }

var _ app.Repository = (*AttentionRepository)(nil)

func (r *AttentionRepository) WithTx(ctx context.Context, fn func(app.AttentionTx) error) error {
	conn, err := r.db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	defer conn.ExecContext(context.Background(), "ROLLBACK")
	if err = fn(&attentionTx{ctx: ctx, conn: conn}); err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, "COMMIT")
	return err
}

type attentionTx struct {
	ctx  context.Context
	conn *sql.Conn
}

var _ app.AttentionTx = (*attentionTx)(nil)

func conflict(message string) error {
	return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: message, RequiredAction: "reload_current_version"}
}
func storageError(err error) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "constraint") {
		return conflict("storage uniqueness or version constraint")
	}
	return err
}

func encode(v any) (string, error) { b, e := json.Marshal(v); return string(b), e }
func equalJSON(a, b any) bool      { x, _ := encode(a); y, _ := encode(b); return x == y }

func (t *attentionTx) load(query, resource, id string, args []any, v any) error {
	var data string
	err := t.conn.QueryRowContext(t.ctx, query, args...).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return apierrors.NewNotFound(resource, id)
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), v)
}

func listJSON[T any](t *attentionTx, query string) ([]T, error) {
	rows, err := t.conn.QueryContext(t.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []T{}
	for rows.Next() {
		var data string
		var v T
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (t *attentionTx) ListMaterials() ([]app.MaterialDetail, error) {
	return listJSON[app.MaterialDetail](t, "SELECT data FROM attention_materials ORDER BY id")
}
func (t *attentionTx) LoadMaterial(id string) (app.MaterialDetail, error) {
	var v app.MaterialDetail
	err := t.load("SELECT data FROM attention_materials WHERE id=?", "material", id, []any{id}, &v)
	return v, err
}
func (t *attentionTx) FindMaterialBySourceKey(key string) (app.MaterialDetail, error) {
	var v app.MaterialDetail
	err := t.load("SELECT data FROM attention_materials WHERE source_key=?", "material", key, []any{key}, &v)
	return v, err
}

// saveHead's table/column parameters are private constants, never client input.
func (t *attentionTx) saveHead(table, id string, expected int, data string, extraCols []string, extras []any) error {
	if expected < 0 {
		return conflict("invalid expected version")
	}
	if expected == 0 {
		cols := append([]string{"id", "version", "data"}, extraCols...)
		args := append([]any{id, 1, data}, extras...)
		_, err := t.conn.ExecContext(t.ctx, "INSERT INTO "+table+" ("+strings.Join(cols, ",")+") VALUES ("+strings.TrimSuffix(strings.Repeat("?,", len(args)), ",")+")", args...)
		return storageError(err)
	}
	sets := []string{"version=?", "data=?"}
	args := []any{expected + 1, data}
	for i, c := range extraCols {
		sets = append(sets, c+"=?")
		args = append(args, extras[i])
	}
	args = append(args, id, expected)
	res, err := t.conn.ExecContext(t.ctx, "UPDATE "+table+" SET "+strings.Join(sets, ",")+" WHERE id=? AND version=?", args...)
	if err != nil {
		return storageError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return conflict("aggregate version changed")
	}
	return nil
}

func requireVersion(version, expected int) error {
	if expected < 0 || version <= 0 || version != expected+1 {
		return conflict("saved aggregate version must advance exactly once")
	}
	return nil
}

func (t *attentionTx) SaveMaterial(v app.MaterialDetail, expected int) error {
	if err := requireVersion(v.Material.Version, expected); err != nil {
		return err
	}
	if len(v.Revisions) == 0 || v.Material.ID == "" {
		return conflict("material requires a source revision")
	}
	key := v.Revisions[0].SourceKey
	if key == "" {
		return conflict("material source key is required")
	}
	if expected > 0 {
		old, err := t.LoadMaterial(v.Material.ID)
		if err != nil {
			return err
		}
		if len(old.Revisions) > len(v.Revisions) || len(old.Uses) > len(v.Uses) {
			return conflict("material history cannot be removed")
		}
		for i, x := range old.Revisions {
			if !equalJSON(x, v.Revisions[i]) {
				return conflict("material revision cannot be overwritten")
			}
		}
		for i, x := range old.Uses {
			if !equalJSON(x, v.Uses[i]) {
				return conflict("usage history cannot be overwritten")
			}
		}
		if old.Revisions[0].SourceKey != key {
			return conflict("material source key cannot change")
		}
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	if err = t.saveHead("attention_materials", v.Material.ID, expected, data, []string{"source_key"}, []any{key}); err != nil {
		return err
	}
	for i, rev := range v.Revisions {
		if rev.MaterialID != v.Material.ID || rev.Revision != i+1 || rev.SourceKey != key || rev.ObjectRef == "" || rev.ContentDigest == "" {
			return conflict("invalid material revision identity")
		}
		data, err = encode(rev)
		if err != nil {
			return err
		}
		_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_material_revisions(material_id,revision,content_digest,data) VALUES(?,?,?,?) ON CONFLICT(material_id,revision) DO NOTHING", v.Material.ID, rev.Revision, rev.ContentDigest, data)
		if err != nil {
			return storageError(err)
		}
		var stored string
		if err = t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_material_revisions WHERE material_id=? AND revision=?", v.Material.ID, rev.Revision).Scan(&stored); err != nil {
			return err
		}
		if stored != data {
			return conflict("material revision cannot be overwritten")
		}
	}
	if v.Material.CurrentRevision < 1 || v.Material.CurrentRevision > len(v.Revisions) {
		return conflict("material current revision is not in immutable history")
	}
	return nil
}

func (t *attentionTx) ListDistillations() ([]app.Distillation, error) {
	return listJSON[app.Distillation](t, "SELECT data FROM attention_distillations ORDER BY id")
}
func (t *attentionTx) FindDistillationByReuseKey(key string) (app.Distillation, error) {
	var v app.Distillation
	err := t.load("SELECT data FROM attention_distillations WHERE reuse_key=?", "distillation", key, []any{key}, &v)
	return v, err
}
func (t *attentionTx) SaveDistillation(v app.Distillation) error {
	data, err := encode(v)
	if err != nil {
		return err
	}
	_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_distillations(id,reuse_key,data) VALUES(?,?,?) ON CONFLICT(id) DO NOTHING", v.ID, v.ReuseKey, data)
	if err != nil {
		return storageError(err)
	}
	var stored string
	if err = t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_distillations WHERE id=?", v.ID).Scan(&stored); err != nil {
		return err
	}
	if stored != data {
		return conflict("distillation cannot be overwritten")
	}
	return nil
}

func (t *attentionTx) ListOpportunities() ([]app.OpportunityDetail, error) {
	return listJSON[app.OpportunityDetail](t, "SELECT data FROM attention_opportunities ORDER BY id")
}
func (t *attentionTx) LoadOpportunity(id string) (app.OpportunityDetail, error) {
	var v app.OpportunityDetail
	err := t.load("SELECT data FROM attention_opportunities WHERE id=?", "opportunity", id, []any{id}, &v)
	return v, err
}
func (t *attentionTx) SaveOpportunity(v app.OpportunityDetail, expected int) error {
	if err := requireVersion(v.Opportunity.Version, expected); err != nil {
		return err
	}
	if len(v.Revisions) == 0 {
		return conflict("opportunity requires revision history")
	}
	if v.Opportunity.Revision != len(v.Revisions) {
		return conflict("opportunity current revision does not match history")
	}
	if expected > 0 {
		old, err := t.LoadOpportunity(v.Opportunity.ID)
		if err != nil {
			return err
		}
		if len(old.Revisions) > len(v.Revisions) || len(old.Reviews) > len(v.Reviews) {
			return conflict("opportunity history cannot be removed")
		}
		for i, x := range old.Revisions {
			if !equalJSON(x, v.Revisions[i]) {
				return conflict("opportunity revision cannot be overwritten")
			}
		}
		for i, x := range old.Reviews {
			if !equalJSON(x, v.Reviews[i]) {
				return conflict("review history cannot be overwritten")
			}
		}
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	if err = t.saveHead("attention_opportunities", v.Opportunity.ID, expected, data, nil, nil); err != nil {
		return err
	}
	for i, rev := range v.Revisions {
		if rev.ID != v.Opportunity.ID || rev.Revision != i+1 {
			return conflict("invalid opportunity revision identity")
		}
		data, err = encode(rev)
		if err != nil {
			return err
		}
		_, err = t.conn.ExecContext(t.ctx, "INSERT INTO attention_opportunity_revisions(opportunity_id,revision,data) VALUES(?,?,?) ON CONFLICT(opportunity_id,revision) DO NOTHING", rev.ID, rev.Revision, data)
		if err != nil {
			return storageError(err)
		}
		var stored string
		if err = t.conn.QueryRowContext(t.ctx, "SELECT data FROM attention_opportunity_revisions WHERE opportunity_id=? AND revision=?", rev.ID, rev.Revision).Scan(&stored); err != nil {
			return err
		}
		if stored != data {
			return conflict("opportunity revision cannot be overwritten")
		}
	}
	return nil
}

func (t *attentionTx) ListJobs() ([]app.Job, error) {
	rows, err := t.conn.QueryContext(t.ctx, "SELECT data,payload,caller FROM attention_jobs ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []app.Job{}
	for rows.Next() {
		var data, payload, caller string
		var v app.Job
		if err = rows.Scan(&data, &payload, &caller); err != nil {
			return nil, err
		}
		if err = decodeJob(data, payload, caller, &v); err != nil {
			return nil, err
		}
		jobs = append(jobs, v)
	}
	return jobs, rows.Err()
}
func decodeJob(data, payload, caller string, v *app.Job) error {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return err
	}
	v.Payload = json.RawMessage(payload)
	return json.Unmarshal([]byte(caller), &v.Caller)
}
func (t *attentionTx) job(query, id string) (app.Job, error) {
	var v app.Job
	var data, payload, caller string
	err := t.conn.QueryRowContext(t.ctx, query, id).Scan(&data, &payload, &caller)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("job", id)
	}
	if err != nil {
		return v, err
	}
	err = decodeJob(data, payload, caller, &v)
	return v, err
}
func (t *attentionTx) LoadJob(id string) (app.Job, error) {
	return t.job("SELECT data,payload,caller FROM attention_jobs WHERE id=?", id)
}
func (t *attentionTx) FindJobByDedupeKey(key string) (app.Job, error) {
	return t.job("SELECT data,payload,caller FROM attention_jobs WHERE dedupe_key=?", key)
}
func (t *attentionTx) SaveJob(v app.Job, expected int) error {
	if v.Version != expected+1 {
		return conflict("saved job version must advance exactly once")
	}
	if expected > 0 {
		old, err := t.LoadJob(v.JobID)
		if err != nil {
			return err
		}
		if old.DedupeKey != v.DedupeKey {
			return conflict("job dedupe identity cannot change")
		}
	}
	data, err := encode(v)
	if err != nil {
		return err
	}
	caller, err := encode(v.Caller)
	if err != nil {
		return err
	}
	payload := string(v.Payload)
	if payload == "" {
		payload = "null"
	}
	if !json.Valid([]byte(payload)) {
		return errors.New("invalid durable job payload")
	}
	return t.saveHead("attention_jobs", v.JobID, expected, data, []string{"dedupe_key", "status", "payload", "caller"}, []any{v.DedupeKey, v.Status, payload, caller})
}

func (t *attentionTx) LoadReceipt(caller, command, key string) (app.Receipt, error) {
	var v app.Receipt
	var result string
	err := t.conn.QueryRowContext(t.ctx, "SELECT digest,result FROM attention_receipts WHERE caller=? AND command=? AND key=?", caller, command, key).Scan(&v.Digest, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return v, apierrors.NewNotFound("receipt", key)
	}
	v.Result = json.RawMessage(result)
	return v, err
}
func (t *attentionTx) SaveReceipt(caller, command, key string, v app.Receipt) error {
	if !json.Valid(v.Result) {
		return errors.New("receipt must contain a complete JSON result")
	}
	_, err := t.conn.ExecContext(t.ctx, "INSERT INTO attention_receipts(caller,command,key,digest,result) VALUES(?,?,?,?,?) ON CONFLICT(caller,command,key) DO NOTHING", caller, command, key, v.Digest, string(v.Result))
	if err != nil {
		return storageError(err)
	}
	old, err := t.LoadReceipt(caller, command, key)
	if err != nil {
		return err
	}
	if old.Digest != v.Digest || string(old.Result) != string(v.Result) {
		return conflict("idempotency key binds a different input or result")
	}
	return nil
}
func (t *attentionTx) AppendEvent(v app.OutboxEvent) error {
	_, err := t.conn.ExecContext(t.ctx, "INSERT INTO attention_outbox(id,type,aggregate_id,aggregate_version,occurred_at,correlation_id,causation_id,payload) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.Type, v.AggregateID, v.AggregateVersion, v.OccurredAt.UTC().Format(time.RFC3339Nano), v.CorrelationID, v.CausationID, string(v.Payload))
	if err != nil {
		return storageError(fmt.Errorf("append Attention event: %w", err))
	}
	return nil
}
