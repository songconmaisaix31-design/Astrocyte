package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayWorkbenchExplicitSourceActualSQLiteProtection(t *testing.T) {
	parent := filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification")
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal("验证目录不可用")
	}
	root, err := os.MkdirTemp(parent, "h-source-guard-")
	if err != nil {
		t.Fatal("本次独立验证目录不可用")
	}
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", "")
	paths, err := workbenchProtectedSources()
	if err != nil || len(paths) != 0 {
		t.Fatal("默认配置不应选择私人来源")
	}
	sourceDir := filepath.Join(root, "owned-public-source")
	if err := os.MkdirAll(sourceDir, 0700); err != nil {
		t.Fatal("本次公开来源目录不可用")
	}
	source := filepath.Join(sourceDir, "state.sqlite")
	db, err := sqlite.Open(source)
	if err != nil {
		t.Fatal("本次真实 SQLite 初始化失败")
	}
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal("本次真实 SQLite 迁移失败")
	}
	if err := db.Close(); err != nil {
		t.Fatal("本次真实 SQLite 关闭失败")
	}
	encoded, err := json.Marshal([]string{source})
	if err != nil {
		t.Fatal("显式来源配置无法编码")
	}
	t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", string(encoded))
	if _, err := guardWorkbenchExplicitDataDir(sourceDir); err == nil {
		t.Fatal("显式真实来源目录必须拒绝应用写入")
	}
	appDir := filepath.Join(root, "owned-app")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		t.Fatal("本次独立应用目录不可用")
	}
	if err := os.Link(source, filepath.Join(appDir, "state.sqlite")); err != nil {
		t.Fatal("本机真实 hardlink 不可用")
	}
	if _, err := guardWorkbenchExplicitDataDir(appDir); err == nil {
		t.Fatal("显式来源的实际 hardlink 必须拒绝")
	}
	for _, invalid := range []string{`["relative.sqlite"]`, `null`, `[] {}`} {
		t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", invalid)
		if _, err := guardWorkbenchExplicitDataDir(filepath.Join(root, "new-app")); err == nil {
			t.Fatal("无效显式保护配置必须立即失败")
		}
	}
}
