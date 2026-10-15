package events_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayGoalEvidenceActualStoresRefuseUnrecordedOutput(t *testing.T) {
	root := t.TempDir()
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal("实际 SQLite 打开失败")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("既有 SQLite 迁移失败")
	}
	objectRoot := filepath.Join(root, "objects")
	store, err := objects.New(objectRoot)
	if err != nil {
		t.Fatal("实际对象目录打开失败")
	}
	repo := sqlite.NewEventRepository(db)
	evidence := events.NewRelayGoalEvidenceStore(store, attention.NewPendingEventService(repo), repo)
	if got, err := evidence.Read(ctx, "unrecorded-original-event"); err == nil || got.EvidenceID != "" {
		t.Fatal("缺少原事件时返回了目标通过")
	}
	if id, err := evidence.Record(ctx, events.RelayGoalIdentity{}, events.RelayGoalOutput{}, "", nil, nil); err == nil || id != "" {
		t.Fatal("缺少原始输出时写入了目标通过")
	}
	var count int
	if db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM attention_engine_events").Scan(&count) != nil || count != 0 {
		t.Fatal("拒绝检查产生了事件")
	}
	files, err := os.ReadDir(objectRoot)
	if err != nil || len(files) != 0 {
		t.Fatal("拒绝检查产生了对象")
	}
	data, err := json.Marshal(events.RelayGoalEvidence{Identity: events.RelayGoalIdentity{NativeID: "private-native"}, Output: []byte("private-output")})
	if err != nil || string(data) != "{}" {
		t.Fatal("私有目标证据进入公共序列化")
	}
}
