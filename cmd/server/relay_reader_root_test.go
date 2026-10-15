package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelayReaderRootActualDirectoriesAndDatabaseAliases(t *testing.T) {
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "local"))
	t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", "[]")
	reader := filepath.Join(root, "cognition")
	if os.MkdirAll(reader, 0700) != nil {
		t.Fatal("实际独立认知目录创建失败")
	}
	t.Setenv("ASTROCYTE_WORKBENCH_READER_ROOT", reader)
	if selected, err := workbenchReaderRoot(); err != nil || selected != reader {
		t.Fatal("独立认知根未保留明确配置")
	}
	app := filepath.Join(root, "app")
	if selected, err := guardWorkbenchExplicitDataDir(app); err != nil || selected != app {
		t.Fatal("独立应用目录未通过实际身份检查")
	}
	for _, snapshotRoot := range []string{workbenchRoot(), reader} {
		generation := filepath.Join(snapshotRoot, "snapshots", "actual-generation")
		if os.MkdirAll(generation, 0700) != nil {
			t.Fatal("实际快照目录创建失败")
		}
		if _, err := guardWorkbenchExplicitDataDir(generation); err == nil {
			t.Fatal("快照目录被用作应用目录")
		}
		appDir, err := os.MkdirTemp(root, "app-")
		if err != nil {
			t.Fatal("实际应用目录创建失败")
		}
		appDB := filepath.Join(appDir, "state.sqlite")
		if os.WriteFile(appDB, nil, 0600) != nil || os.Link(appDB, filepath.Join(generation, "snapshot.sqlite")) != nil {
			t.Fatal("实际同文件双名称创建失败")
		}
		if _, err := guardWorkbenchExplicitDataDir(appDir); err == nil {
			t.Fatal("应用数据库与快照同文件未被拒绝")
		}
	}
	for _, invalid := range []string{"relative", reader + " ", filepath.Join(root, "missing")} {
		t.Setenv("ASTROCYTE_WORKBENCH_READER_ROOT", invalid)
		if _, err := workbenchReaderRoot(); err == nil {
			t.Fatal("不可用认知根未立即拒绝")
		}
	}
}
