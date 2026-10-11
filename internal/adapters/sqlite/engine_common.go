package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// EngineRepository is the shared base for every attention-engine store (events,
// engine, sleep, knowledge, experience). It owns no domain table; each track
// composes its richer tx on top of EngineTx. BEGIN IMMEDIATE serializes
// read/modify/write callbacks across independent connections and processes,
// the same contract as the S1 Attention repository.
type EngineRepository struct{ db *DB }

// NewEngineRepository returns a repository scoped to the attention_engine_* tables.
func NewEngineRepository(db *DB) *EngineRepository { return &EngineRepository{db: db} }

// Conn exposes the shared connection for read-only cross-table joins.
func (r *EngineRepository) Conn() *sql.DB { return r.db.conn }

// WithTx runs fn in one serialized transaction. Do not call external providers
// or objects inside the callback.
func (r *EngineRepository) WithTx(ctx context.Context, fn func(EngineTx) error) error {
	return r.withTransaction(ctx, "BEGIN IMMEDIATE", fn)
}

// WithReadTx takes a consistent deferred snapshot without claiming a writer
// lock, including when SQLite query_only is enabled for retrieval acceptance.
func (r *EngineRepository) WithReadTx(ctx context.Context, fn func(EngineTx) error) error {
	return r.withTransaction(ctx, "BEGIN", fn)
}

func (r *EngineRepository) withTransaction(ctx context.Context, begin string, fn func(EngineTx) error) error {
	conn, err := r.db.conn.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, begin); err != nil {
		return err
	}
	defer conn.ExecContext(context.Background(), "ROLLBACK")
	if err = fn(&engineTx{ctx: ctx, conn: conn, readOnly: begin == "BEGIN"}); err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, "COMMIT")
	return err
}

// engineTx is the transaction handle passed to WithTx callbacks.
type engineTx struct {
	ctx      context.Context
	conn     *sql.Conn
	readOnly bool
}

// EngineTx is the shared transaction contract every engine store embeds. The
// query primitives are context-bound so a store never holds an unmanaged
// connection; domain-specific methods live in each store's richer tx.
type EngineTx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func (t *engineTx) Exec(query string, args ...any) (sql.Result, error) {
	if t.readOnly {
		return nil, errors.New("engine read transaction cannot execute mutations")
	}
	return t.conn.ExecContext(t.ctx, query, args...)
}
func (t *engineTx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.conn.QueryContext(t.ctx, query, args...)
}
func (t *engineTx) QueryRow(query string, args ...any) *sql.Row {
	return t.conn.QueryRowContext(t.ctx, query, args...)
}

// engineLoadJSON decodes one engine row's data column into v, mapping a missing
// row to a NotFound error. Shared by every engine store.
func engineLoadJSON(t *engineTx, query, resource, id string, args []any, v any) error {
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

// engineListJSON decodes every row's data column in id order. Shared by every
// engine store.
func engineListJSON[T any](t *engineTx, query string) ([]T, error) {
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
