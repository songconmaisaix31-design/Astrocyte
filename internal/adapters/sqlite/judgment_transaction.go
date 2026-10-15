package sqlite

import (
	"context"
	"database/sql"

	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type judgmentExecutor interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (db *DB) judgmentExecutor(ctx context.Context) (judgmentExecutor, error) {
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return nil, err
	}
	if conn != nil {
		return conn, nil
	}
	return db.conn, nil
}

func (db *DB) WithJudgmentTx(ctx context.Context, use func(context.Context) error) error {
	return NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return use(tx.Context())
	})
}
