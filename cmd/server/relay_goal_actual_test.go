package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 仅在受信文件编辑已应用本次原生输出后显式运行，不创建或操作原生控制器。
func recordRelayGoalActualOwnedOutputAndHostChecks(t *testing.T) {
	t.Helper()
	if os.Getenv("ASTROCYTE_L1_GOAL_RECORD") != "1" {
		t.Fatal("NOT_RUN: explicit trusted application and actual goal checks required")
	}
	ctx := context.Background()
	evidenceDir := os.Getenv("ASTROCYTE_L1_EVIDENCE_DIR")
	taskID := os.Getenv("ASTROCYTE_L1_TASK_ID")
	operation := os.Getenv("ASTROCYTE_L1_INPUT_OPERATION_ID")
	if !filepath.IsAbs(evidenceDir) || taskID == "" || operation == "" {
		t.Fatal("必须明确本次原任务、已接受输入和私有服务记录")
	}
	var private struct {
		AppRoot, DBPath, SourceRoot string
	}
	var locator struct {
		AppRoot    string `json:"app_root"`
		DBPath     string `json:"db_path"`
		SourceRoot string `json:"source_root"`
	}
	readActualJSON(t, filepath.Join(evidenceDir, "service-private.json"), &locator)
	private.AppRoot, private.DBPath, private.SourceRoot = locator.AppRoot, locator.DBPath, locator.SourceRoot
	root, err := relayRoot()
	if err != nil || private.AppRoot != root || private.DBPath != filepath.Join(root, "app-state", "state.sqlite") || !strings.EqualFold(filepath.Clean(private.SourceRoot), filepath.Clean(os.Getenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT"))) {
		t.Fatal("实际服务记录与本次明确配置不符")
	}
	db, err := sqlite.Open(private.DBPath)
	if err != nil {
		t.Fatal("实际 SQLite 不可用")
	}
	defer db.Close()
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		t.Fatal("实际 Relay store 不可用")
	}
	defer store.Close()
	state, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal("原任务快照不可用")
	}
	task, exists := state.Tasks[taskID]
	if !exists {
		t.Fatal("明确原任务不存在")
	}
	accepted := false
	for _, receipt := range state.Receipts {
		if receipt.TaskID == taskID && receipt.OperationID == operation && receipt.Status == "accepted" && (receipt.Action == "dispatch" || receipt.Action == "send") {
			accepted = true
		}
	}
	if !accepted {
		t.Fatal("原输入缺少已接受收据，禁止重发或倒造")
	}
	controllers := append([]domain.RelayControllerEvidence{}, state.ControllerHistory[taskID]...)
	controllers = append(controllers, state.ControllerEvidence[taskID])
	var identity events.RelayGoalIdentity
	for _, controller := range controllers {
		for _, turn := range controller.Turns {
			if turn.OperationID != operation {
				continue
			}
			if identity.InputOperationID != "" {
				t.Fatal("原输入存在多个未区分的控制器关系")
			}
			identity = events.RelayGoalIdentity{TaskID: taskID, ProjectID: task.ProjectID, WorktreeID: task.WorktreeID, TerminalID: task.TerminalID, ControllerID: controller.ControllerID, StartupOperationID: controller.StartupOperationID, SessionID: controller.SessionID, NativeID: controller.NativeID, InputOperationID: operation, TurnID: turn.TurnID, AcceptedAt: turn.AcceptedAt}
		}
	}
	bound, exists := state.OwnedSessions[taskID]
	if !exists || identity.InputOperationID == "" || bound.SourceStatus != "available" || bound.ControllerID != identity.ControllerID || bound.StartupOperationID != identity.StartupOperationID || bound.SessionID != identity.SessionID || bound.NativeID != identity.NativeID || bound.Ownership != "owned" {
		t.Fatal("原输入与正选来源关系未确认")
	}
	output, err := events.ReadRelayGoalOutput(ctx, relayCaptureBinding(bound), identity.TurnID, identity.AcceptedAt)
	if err != nil {
		t.Fatal("本次原生输出来源拒绝；原件和首错误保留")
	}
	outputID, err := sqlite.NewAccessSessionReader(db).ReadCommittedRelayGoalOutput(ctx, output.Origin, identity.SessionID, identity.ProjectID, relayGoalProjection(output.Projection))
	if err != nil {
		t.Fatal("本次输出尚无逐项核实的实际 SQL event_id")
	}
	registry := agents.NewRegistry(filepath.Join(root, "app-state", "native-sessions"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	readers, err := relayPublicLaunchContexts(projects)
	if err != nil || len(readers) != 1 {
		t.Fatal("原批准固定上下文配置不完整")
	}
	public, ok := readers[0].(*relayPublicContext)
	if !ok || public.commit != "da3bdddb085808863090a05f60cc1c64415039cf" {
		t.Fatal("实际目标不是原批准 Mean 和 Median baseline")
	}
	project, err := db.LoadProject(ctx, task.ProjectID)
	if err != nil || public.verify(ctx, project) != nil {
		t.Fatal("实际应用目录或 baseline 身份改变")
	}
	objectStore, err := objects.New(filepath.Join(root, "app-state", "objects"))
	if err != nil {
		t.Fatal("原对象存储不可用")
	}
	repo := sqlite.NewEventRepository(db)
	goalEvidence := events.NewRelayGoalEvidenceStore(objectStore, attention.NewPendingEventService(repo), repo)
	files := map[string][]byte{}
	for _, name := range []string{"stats.go", "stats_test.go"} {
		files[name], err = os.ReadFile(filepath.Join(public.root, name))
		if err != nil {
			t.Fatal("已应用目标文件不可用")
		}
	}
	for _, name := range []string{"README.md", "go.mod"} {
		original, err := public.git(ctx, "show", public.commit+":"+name)
		if err != nil {
			t.Fatal("原只读文件不可用")
		}
		current, err := os.ReadFile(filepath.Join(public.root, name))
		if err != nil || !bytes.Equal(original, current) {
			t.Fatal("原只读文件改变")
		}
	}
	overlay := os.Getenv("ASTROCYTE_L1_GOAL_OVERLAY")
	contract := readRelayActualGoalContract(t, public.root, overlay)
	checks := []events.RelayGoalHostCheck{}
	for _, check := range []struct {
		name, executable string
		args             []string
	}{{"go_test", "go", []string{"test", "./..."}}, {"go_build", "go", []string{"build", "./..."}}, {"git_diff_check", "git", []string{"diff", "--check"}}} {
		command := exec.CommandContext(ctx, check.executable, check.args...)
		command.Dir = public.root
		command.Env = append(os.Environ(), "GOMAXPROCS=2", "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOFLAGS=-p=1 -overlay="+overlay)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		receipt := events.RelayGoalHostCheck{Name: check.name, StartedAt: time.Now().UTC()}
		runErr := command.Run()
		receipt.FinishedAt, receipt.Stdout, receipt.Stderr = time.Now().UTC(), stdout.Bytes(), stderr.Bytes()
		if check.name == "go_test" {
			receipt.Contract = contract
		}
		if runErr != nil {
			receipt.ExitCode = -1
			var exit *exec.ExitError
			if errors.As(runErr, &exit) {
				receipt.ExitCode = exit.ExitCode()
			}
		}
		checks = append(checks, receipt)
		if runErr != nil {
			id, recordErr := goalEvidence.Record(ctx, identity, output, outputID, files, checks)
			writeActualJSON(t, filepath.Join(evidenceDir, "goal-check-failure-derived.json"), map[string]any{"evidence_id": id, "recorded": recordErr == nil, "passed": false})
			t.Fatal("固定目标实际检查失败，真实结果已保留")
		}
	}
	for name, original := range files {
		current, err := os.ReadFile(filepath.Join(public.root, name))
		if err != nil || !bytes.Equal(original, current) {
			t.Fatal("检查期间实际目标文件改变")
		}
	}
	if public.verify(ctx, project) != nil {
		t.Fatal("检查期间批准目录或 baseline 身份改变")
	}
	id, err := goalEvidence.Record(ctx, identity, output, outputID, files, checks)
	if err != nil {
		t.Fatal("实际目标检查关系保存失败，不重发模型输入")
	}
	writeActualJSON(t, filepath.Join(evidenceDir, "goal-check-derived.json"), map[string]any{"evidence_id": id, "task_id": taskID, "input_operation_id": operation, "output_event_ids": []string{outputID}, "check_evidence_ids": []string{id}, "passed": true})
}

func readRelayActualGoalContract(t *testing.T, root, overlay string) []byte {
	t.Helper()
	if !filepath.IsAbs(overlay) || strings.ContainsAny(overlay, "\r\n\t ") {
		t.Fatal("必须明确首输入前固定的宿主目标检查")
	}
	raw, err := os.ReadFile(overlay)
	if err != nil {
		t.Fatal("固定检查配置不可用")
	}
	var config struct{ Replace map[string]string }
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&config) != nil || decoder.Decode(new(any)) != io.EOF || len(config.Replace) != 1 {
		t.Fatal("宿主目标检查只允许添加一个独立测试文件")
	}
	virtual := filepath.Join(root, "relay_host_goal_test.go")
	var contract []byte
	for selected, source := range config.Replace {
		if !strings.EqualFold(filepath.Clean(selected), filepath.Clean(virtual)) || !filepath.IsAbs(source) {
			t.Fatal("宿主检查不得替换目标原文件")
		}
		contract, err = os.ReadFile(source)
		if err != nil || len(contract) == 0 {
			t.Fatal("首输入前固定的宿主测试不可用")
		}
	}
	return contract
}
