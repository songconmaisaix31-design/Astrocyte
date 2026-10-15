package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayGoalOutputActualSQLiteRequiresOriginalCommittedEvent(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal("实际 SQLite 打开失败")
	}
	defer db.Close()
	ctx := context.Background()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("既有 SQLite 迁移失败")
	}
	reader := NewAccessSessionReader(db)
	projection := workbench.Event{ID: uuid.NewString(), Kind: "task_complete", Scope: uuid.NewString(), TS: time.Now().UTC().Format(time.RFC3339Nano)}
	origin := "relay-session:" + uuid.NewString() + ":" + projection.ID
	if id, err := reader.ReadCommittedRelayGoalOutput(ctx, origin, uuid.NewString(), projection.Scope, projection); err == nil || id != "" {
		t.Fatal("未提交位置返回了事件引用")
	}
	if id, err := reader.ReadCommittedRelayGoalOutput(ctx, origin, uuid.NewString(), uuid.NewString(), projection); err == nil || id != "" {
		t.Fatal("跨项目查询没有拒绝")
	}
	var count int
	if db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM attention_engine_events").Scan(&count) != nil || count != 0 {
		t.Fatal("只读查询产生了事件")
	}
}
