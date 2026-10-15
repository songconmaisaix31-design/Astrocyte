package events

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRelaySourceIdentityActualFileAndAliasRejection(t *testing.T) {
	parent := filepath.Join("..", "..", "..", "dist", "relay-source")
	if outputRoot := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT"); outputRoot != "" {
		if !filepath.IsAbs(outputRoot) {
			t.Fatal("明确的测试输出根必须为绝对目录")
		}
		parent = filepath.Join(outputRoot, "relay-source")
	}
	parent, err := filepath.Abs(parent)
	if err != nil || os.MkdirAll(parent, 0700) != nil {
		t.Fatal("独占来源检查目录不可用")
	}
	root, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("独占来源目录创建失败")
	}
	workspace := filepath.Join(root, "workspace")
	approved := filepath.Join(root, "approved")
	if os.Mkdir(workspace, 0700) != nil || os.Mkdir(approved, 0700) != nil {
		t.Fatal("实际目录创建失败")
	}
	path := filepath.Join(approved, "public-source.txt")
	content := []byte("Public file for filesystem identity checks.\n")
	if os.WriteFile(path, content, 0600) != nil {
		t.Fatal("实际公开文件创建失败")
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal("实际文件状态不可用")
	}
	identity, err := VerifyRelaySessionSource(context.Background(), path, workspace, []string{approved})
	if err != nil || identity == "" {
		t.Fatalf("合法单链接文件检查失败: %v", err)
	}
	again, err := VerifyRelaySessionSource(context.Background(), path, workspace, []string{approved})
	if err != nil || again != identity {
		t.Fatal("同一文件的身份检查不稳定")
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("来源检查改变了实际文件")
	}
	if got, err := VerifyRelaySessionSource(context.Background(), path, workspace, []string{workspace}); err == nil || got != "" {
		t.Fatal("未授权来源没有被拒绝")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := VerifyRelaySessionSource(cancelled, path, workspace, []string{approved}); err == nil || got != "" {
		t.Fatal("已取消检查返回了可用身份")
	}
	if os.Link(path, filepath.Join(approved, "owned-alias.txt")) != nil {
		t.Fatal("实际文件硬链接创建失败")
	}
	if got, err := VerifyRelaySessionSource(context.Background(), path, workspace, []string{approved}); err == nil || got != "" || err.Error() != "relay session capture: source_identity_unverifiable_or_aliased" {
		t.Fatal("多个名称的来源没有按原单链接保护拒绝")
	}
}
