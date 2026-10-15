package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayWorkbenchActualLinkedGenerationRejected(t *testing.T) {
	parent, err := filepath.Abs(filepath.Join("..", "..", "dist", "workbench-links"))
	if err != nil || os.MkdirAll(parent, 0700) != nil {
		t.Fatal("本次链接验证目录不可用")
	}
	for _, kind := range []string{"symlink", "junction"} {
		t.Run(kind, func(t *testing.T) {
			root, err := os.MkdirTemp(parent, "owned-"+kind+"-")
			if err != nil {
				t.Fatal("本次链接验证独占目录不可用")
			}
			t.Setenv("LOCALAPPDATA", root)
			t.Setenv("ASTROCYTE_WORKBENCH_PROTECTED_SOURCES", "")
			target := filepath.Join(root, "owned-public-generation")
			appDir := filepath.Join(root, "owned-app")
			snapshotRoot := filepath.Join(workbenchRoot(), "snapshots")
			for _, dir := range []string{target, appDir, snapshotRoot} {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal("本次实际目录创建失败")
				}
			}
			snapshot := filepath.Join(target, "snapshot.sqlite")
			db, err := sqlite.Open(snapshot)
			if err != nil {
				t.Fatal("实际 SQLite 打开失败")
			}
			if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
				t.Fatal("实际 SQLite 迁移失败")
			}
			if err := db.Close(); err != nil {
				t.Fatal("实际 SQLite 关闭失败")
			}
			if err := os.Link(snapshot, filepath.Join(appDir, "state.sqlite")); err != nil {
				t.Fatal("实际 SQLite hardlink 创建失败")
			}
			linked := filepath.Join(snapshotRoot, "owned-linked-generation")
			if kind == "symlink" || runtime.GOOS != "windows" {
				if err := os.Symlink(target, linked); err != nil {
					var code syscall.Errno
					if errors.As(err, &code) {
						t.Fatalf("实际目录 symlink 创建失败，OS error=%d", uint64(code))
					}
					t.Fatal("实际目录 symlink 创建失败，OS error 未提供")
				}
			} else {
				command := exec.Command("cmd.exe", "/c", "mklink", "/J", linked, target)
				command.Stdout, command.Stderr = io.Discard, io.Discard
				if err := command.Run(); err != nil {
					t.Fatal("实际目录 junction 创建失败")
				}
			}
			if _, err := guardWorkbenchExplicitDataDir(appDir); err == nil {
				t.Fatal("实际链接出版目录被跳过并允许应用库写入")
			}
		})
	}
}
