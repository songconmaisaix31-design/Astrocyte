package events

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRelayGoalOutputActualFileRefusesAliasesAndNonArchive(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	approved := filepath.Join(root, "approved")
	if os.Mkdir(workspace, 0700) != nil || os.Mkdir(approved, 0700) != nil {
		t.Fatal("实际独占目录创建失败")
	}
	path := filepath.Join(approved, "public.txt")
	if os.WriteFile(path, []byte("Public text file for source rejection checks.\n"), 0600) != nil {
		t.Fatal("实际公开文件创建失败")
	}
	binding := RelaySessionBinding{TaskID: uuid.NewString(), TerminalID: uuid.NewString(), SessionID: uuid.NewString(), WorktreeID: uuid.NewString(), ProjectID: uuid.NewString(), NativeID: uuid.NewString(), Format: "codex", Path: path, WorkspaceRoot: workspace, ApprovedRoots: []string{approved}, Owned: true, Authorized: true}
	accepted := time.Now().UTC()
	if output, err := ReadRelayGoalOutput(context.Background(), binding, "", accepted); err == nil || len(output.Text) != 0 {
		t.Fatal("未确认 turn 没有立即拒绝")
	}
	if output, err := ReadRelayGoalOutput(context.Background(), binding, uuid.NewString(), accepted); err == nil || len(output.Text) != 0 {
		t.Fatal("普通文件被误认作原生输出")
	}
	if os.Link(path, filepath.Join(approved, "public-alias.txt")) != nil {
		t.Fatal("实际文件链接创建失败")
	}
	if output, err := ReadRelayGoalOutput(context.Background(), binding, uuid.NewString(), accepted); err == nil || err.Error() != "relay session capture: source_identity_unverifiable_or_aliased" || len(output.Text) != 0 || output.Origin != "" {
		t.Fatal("多名称来源未按单链接保护拒绝")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if output, err := ReadRelayGoalOutput(cancelled, binding, uuid.NewString(), accepted); err == nil || len(output.Text) != 0 {
		t.Fatal("已取消读取返回了输出")
	}
}

func TestRelayGoalOutputNeverSerializesOriginal(t *testing.T) {
	data, err := json.Marshal(RelayGoalOutput{Text: []byte("private output"), Origin: "private origin", Projection: RelaySessionProjection{ID: "private reference"}})
	if err != nil || string(data) != "{}" {
		t.Fatal("本机输出进入了公共序列化")
	}
}
