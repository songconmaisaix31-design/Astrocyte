package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRunCLIHelpIsLocal(t *testing.T) {
	t.Setenv("SWA_PROJECT_TOKEN", "")
	t.Setenv("SWA_ENDPOINT", "invalid-endpoint")
	for _, verb := range []string{"run", "wait", "export"} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), []string{verb, "--help"}, &stdout, &stderr)
		if exit != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "swa run") || !strings.Contains(stdout.String(), "8=等待超时") || !strings.Contains(stdout.String(), "9=需要用户回答") || strings.Contains(stdout.String(), "invalid-endpoint") {
			t.Fatalf("完整任务帮助依赖服务或遗漏退出状态：exit=%d", exit)
		}
	}
}

func TestRunCLIRejectsInvalidOptionsBeforeTransport(t *testing.T) {
	t.Setenv("SWA_ENDPOINT", "invalid-endpoint")
	t.Setenv("SWA_PROJECT_TOKEN", "")
	for _, args := range [][]string{
		{"run", "--project", "unresolved-project", "--request-id", "invalid", "--json"},
		{"run", "--project", "unresolved-project", "--timeout", "0s", "--json"},
		{"run", "--project", "unresolved-project", "--on-input", "stop", "--json"},
		{"wait", "--project", "unresolved-project", "--controller", "unresolved-controller", "--task", "unresolved-root", "--json"},
		{"wait", "--project", "unresolved-project", "--request-id", uuid.NewString(), "--json"},
		{"export", "--project", "unresolved-project", "--strict", "--json"},
		{"run", "--from", "untrusted-identity", "--json"},
		{"run", "--unknown", "--help", "--json"},
	} {
		var stdout, stderr bytes.Buffer
		exit := run(context.Background(), args, &stdout, &stderr)
		requireLocalRefusal(t, exit, 2, "validation_failed", "", &stdout)
	}
}

func TestRunCLIRequiresExistingGrant(t *testing.T) {
	t.Setenv("SWA_PROJECT_TOKEN", "")
	var stdout, stderr bytes.Buffer
	exit := run(context.Background(), []string{"wait", "--project", "unresolved-project", "--controller", "unresolved-controller", "--task", "unresolved-root", "--run", "unresolved-run", "--json"}, &stdout, &stderr)
	if exit != 3 || !strings.Contains(stdout.String(), `"code":"scope_denied"`) {
		t.Fatalf("只读重连绕过当前项目授权：exit=%d", exit)
	}
}

func TestRunExportPreflightUsesRealFilesystem(t *testing.T) {
	parent := t.TempDir()
	path := filepath.Join(parent, "new-export")
	if actual, err := prepareRunExport(path); err != nil || actual != path {
		t.Fatalf("新绝对目录前置失败：%v", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("前置验证擅自创建了目录")
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareRunExport(path); err == nil {
		t.Fatal("已存在目录被接受")
	}
	if _, err := prepareRunExport("relative-output"); err == nil {
		t.Fatal("相对导出路径被接受")
	}
	if _, err := prepareRunExport(filepath.Join(parent, "missing", "output")); err == nil {
		t.Fatal("不存在父目录被接受")
	}
	link := filepath.Join(parent, "linked-parent")
	if err := os.Symlink(parent, link); err != nil {
		t.Logf("当前环境不允许创建实际符号链接：%v", err)
		return
	}
	if _, err := prepareRunExport(filepath.Join(link, "output")); err == nil {
		t.Fatal("符号链接父路径被接受")
	}
}
