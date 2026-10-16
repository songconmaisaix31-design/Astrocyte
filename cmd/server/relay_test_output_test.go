package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func relayActualTestDirectory(t *testing.T) string {
	t.Helper()
	parent := filepath.Join("..", "..", "dist", "relay-actual")
	explicit := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if explicit != "" {
		if !filepath.IsAbs(explicit) || strings.TrimSpace(explicit) != explicit {
			t.Fatal("实际测试输出根必须为明确的绝对目录")
		}
		parent = filepath.Join(explicit, "relay-actual")
	}
	parent, err := filepath.Abs(parent)
	if err != nil {
		t.Fatalf("实际测试父目录无法解析，错误类型=%T", err)
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatalf("实际测试父目录不可用，错误类型=%T", err)
	}
	selected, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatalf("实际测试独占目录创建失败，错误类型=%T", err)
	}
	physical, err := filepath.EvalSymlinks(selected)
	if err != nil {
		t.Fatalf("实际测试物理目录无法解析，错误类型=%T", err)
	}
	selectedInfo, err := os.Stat(selected)
	if err != nil {
		t.Fatalf("实际测试目录身份不可用，错误类型=%T", err)
	}
	physicalInfo, err := os.Stat(physical)
	if err != nil || !physicalInfo.IsDir() || !os.SameFile(selectedInfo, physicalInfo) {
		t.Fatalf("实际测试目录身份不一致，读取失败=%t，错误类型=%T", err != nil, err)
	}
	// 目录原值只保存在本次自有原件，测试输出仅使用安全类型和布尔结果。
	evidence, err := json.Marshal(struct {
		Selected string `json:"selected"`
		Physical string `json:"physical"`
		Explicit bool   `json:"explicit_output_root"`
		SameFile bool   `json:"same_file"`
	}{selected, physical, explicit != "", true})
	if err != nil {
		t.Fatalf("实际目录证据编码失败，错误类型=%T", err)
	}
	if err := os.WriteFile(filepath.Join(physical, "directory.json"), evidence, 0600); err != nil {
		t.Fatalf("实际目录证据保存失败，错误类型=%T", err)
	}
	return physical
}

func relayActualTestFailure(t *testing.T, root, message string, failure error, metadata map[string]string) {
	t.Helper()
	detail := ""
	if failure != nil {
		detail = failure.Error()
	}
	evidence, err := json.Marshal(struct {
		Message  string            `json:"message"`
		Type     string            `json:"error_type"`
		Error    string            `json:"error"`
		Metadata map[string]string `json:"metadata"`
	}{message, fmt.Sprintf("%T", failure), detail, metadata})
	if err != nil {
		t.Fatalf("实际失败诊断编码失败，错误类型=%T", err)
	}
	if err := os.WriteFile(filepath.Join(root, "failure.json"), evidence, 0600); err != nil {
		t.Fatalf("实际失败诊断保存失败，原错误类型=%T，保存错误类型=%T", failure, err)
	}
	t.Fatalf("%s，读取失败=%t，错误类型=%T", message, failure != nil, failure)
}
