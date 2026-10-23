// Package sqlite 保留旧产品的同库组装，项目与空间使用 SWA 唯一仓储。
package sqlite

import (
	"context"
	"database/sql"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlitebase"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/swastore"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type DB struct {
	*swastore.DB
	conn   *sql.DB
	dbPath string
}

// 旧只读来源 adapter 也可直接持有原 SQL 连接，关闭与读取不要求写仓储。
func (db *DB) Close() error  { return db.conn.Close() }
func (db *DB) Conn() *sql.DB { return db.conn }

func Open(path string) (*DB, error) {
	base, err := sqlitebase.Open(path)
	if err != nil {
		return nil, err
	}
	db := &DB{conn: base.Conn(), dbPath: path}
	db.DB, err = swastore.Attach(base, workspaceTransactions{db: db})
	if err != nil {
		_ = base.Close()
		return nil, err
	}
	return db, nil
}

type workspaceTransactions struct{ db *DB }

func (owner workspaceTransactions) Connection(ctx context.Context) (*sql.Conn, error) {
	return WorkspaceTransaction(ctx, owner.db)
}
func (owner workspaceTransactions) WithTransaction(ctx context.Context, fn func(context.Context, *sql.Conn) error) error {
	return NewWorkspaceRepository(owner.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		conn, err := WorkspaceTransaction(tx.Context(), owner.db)
		if err != nil {
			return err
		}
		return fn(tx.Context(), conn)
	})
}
