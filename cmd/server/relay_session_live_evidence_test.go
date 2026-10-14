package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/olivere/ndjson"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只检查当前实际服务持有的会话和提交记录，不生成或回放原生协议。
func TestRelaySessionActualCommittedEvidence(t *testing.T) {
	evidence := os.Getenv("ASTROCYTE_L1_EVIDENCE_DIR")
	if evidence == "" {
		t.Skip("NOT_RUN: explicit actual-service evidence directory required")
	}
	expected, err := filepath.Abs(filepath.Join("..", "..", "dist", "l1-live"))
	if err != nil || !relayInside(expected, evidence) {
		t.Fatal("实际验证目录超出当前 H 工作目录")
	}
	var private struct {
		AppRoot    string `json:"app_root"`
		DBPath     string `json:"db_path"`
		SourceRoot string `json:"source_root"`
	}
	readActualJSON(t, filepath.Join(evidence, "service-private.json"), &private)
	if !relayInside(filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification"), private.AppRoot) || private.DBPath != filepath.Join(private.AppRoot, "app-state", "state.sqlite") {
		t.Fatal("实际数据库目录超出本轮范围")
	}
	t.Setenv("LOCALAPPDATA", filepath.Dir(filepath.Dir(private.AppRoot)))
	store, err := sqlite.OpenRelayStore(private.AppRoot)
	if err != nil {
		t.Fatal("实际运行时数据库不可用")
	}
	defer store.Close()
	ctx := context.Background()
	state, err := store.Snapshot(ctx)
	if err != nil || len(state.Tasks) == 0 {
		t.Fatal("实际任务身份不可核实")
	}
	var task domain.RelayTask
	if selected := os.Getenv("ASTROCYTE_L1_TASK_ID"); selected != "" {
		var found bool
		task, found = state.Tasks[selected]
		if !found {
			t.Fatal("本次明确选定任务不存在")
		}
	} else {
		if len(state.Tasks) != 1 {
			t.Fatal("实际多任务验证必须明确选定 task_id")
		}
		for _, v := range state.Tasks {
			task = v
		}
	}
	owned, ok := state.OwnedSessions[task.ID]
	if !ok || owned.TaskID != task.ID || owned.SourceStatus != "available" {
		t.Fatal("实际 owned archive 未选择")
	}
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", private.SourceRoot)
	provider, err := relayWorkspaceProvider(private.AppRoot)
	if err != nil {
		t.Fatal("实际公开项目配置未被生产组装接受")
	}
	work := domain.Work{Root: private.SourceRoot, Kind: domain.WorkKindFolder}
	prepared := state.Workspaces[task.WorktreeID]
	kept, err := provider.Prepare(ctx, work, prepared)
	if err != nil || !reflect.DeepEqual(kept, prepared) {
		t.Fatal("实际已持有公开工作树无法保持原身份")
	}
	work.Root = private.AppRoot
	_, err = provider.Prepare(ctx, work, prepared)
	var denied *apierrors.ServiceError
	if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
		t.Fatal("实际其他目录未被公开项目配置拒绝")
	}
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", "")
	provider, err = relayWorkspaceProvider(private.AppRoot)
	if err != nil {
		t.Fatal("生产默认配置不可用")
	}
	prepared.Prepared = false
	work.Root = private.SourceRoot
	_, err = provider.Prepare(ctx, work, prepared)
	if !errors.As(err, &denied) || denied.Code != apierrors.UnsupportedCapability {
		t.Fatal("生产默认公开项目启动权限未保持关闭")
	}
	binding := relayCaptureBinding(owned)
	batch, err := events.ReadRelaySession(ctx, binding, events.RelaySessionCursor{})
	if err != nil || !batch.HeaderVerified || !batch.SnapshotComplete || batch.Next.Offset != batch.SnapshotBytes || batch.TailBytes != 0 || batch.Next.Counts.Lines < 2 {
		t.Fatal("实际归档当前全部完整行未覆盖")
	}
	if batch.SessionComplete {
		t.Fatal("实际归档缺封存及前缀证据，不能宣称完整 Session")
	}
	file, err := os.Open(binding.Path)
	if err != nil {
		t.Fatal("实际原生模型元数据不可用")
	}
	defer file.Close()
	models := map[string]bool{}
	toolCalls := 0
	reader := ndjson.NewReaderSize(file, int(batch.SnapshotBytes)+1)
	for reader.Next() {
		var record struct {
			Type    string `json:"type"`
			Payload struct {
				Type  string `json:"type"`
				Model string `json:"model"`
			} `json:"payload"`
		}
		if reader.Decode(&record) != nil {
			t.Fatal("实际原生模型元数据格式错误")
		}
		if record.Type == "turn_context" && record.Payload.Model != "" {
			models[record.Payload.Model] = true
		}
		if record.Type == "response_item" && (record.Payload.Type == "function_call" || record.Payload.Type == "custom_tool_call") {
			toolCalls++
		}
	}
	if reader.Err() != nil || len(models) == 0 || toolCalls != 0 {
		t.Fatal("实际原生会话缺少模型元数据或发生工具调用")
	}
	restart, err := events.ReadRelaySession(ctx, binding, batch.Next)
	if err != nil || len(restart.Items) != 0 || restart.Next.Offset != batch.Next.Offset || restart.Next.Counts != batch.Next.Counts {
		t.Fatal("实际静默重读改变了位置或计数")
	}
	for _, negative := range []string{"unowned", "permission", "native", "cwd", "cursor"} {
		rejectedBinding, cursor := binding, batch.Next
		switch negative {
		case "unowned":
			rejectedBinding.Owned = false
		case "permission":
			rejectedBinding.Authorized = false
		case "native":
			rejectedBinding.NativeID += "-different"
		case "cwd":
			rejectedBinding.WorkspaceRoot = private.SourceRoot
		case "cursor":
			cursor.Offset--
		}
		failure, err := events.ReadRelaySession(ctx, rejectedBinding, cursor)
		if err == nil || failure.SessionComplete || failure.Next.FirstError == "" {
			t.Fatalf("实际文件负例没有立即失败: %s", negative)
		}
	}
	db, err := sqlite.Open(private.DBPath)
	if err != nil {
		t.Fatal("实际事件仓储不可用")
	}
	defer db.Close()
	repo := sqlite.NewEventRepository(db)
	readonly, err := sql.Open("sqlite", "file:"+filepath.ToSlash(private.DBPath)+"?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal("独立只读 SQLite 不可用")
	}
	defer readonly.Close()
	var signals, nodes int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_signals").Scan(&signals); err != nil {
		t.Fatal("实际认知信号计数不可用")
	}
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_nodes").Scan(&nodes); err != nil {
		t.Fatal("实际认知节点计数不可用")
	}
	for _, item := range batch.Items {
		var id string
		if err := readonly.QueryRowContext(ctx, "SELECT id FROM attention_engine_events WHERE origin_key = ?", item.Origin).Scan(&id); err != nil {
			t.Fatal("实际文件记录未逐条对应唯一提交 ID")
		}
		row, err := repo.GetEvent(ctx, id)
		if err != nil || row.OriginKey != item.Origin {
			t.Fatal("实际文件记录身份未在事件仓储定位")
		}
	}
	var before, after int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&before); err != nil {
		t.Fatal("实际事件计数不可用")
	}
	client := newActualRelayHTTP(t)
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	path := "/api/v1/relay/session-events?task_id=" + task.ID + "&page_size=100"
	for _, denied := range []string{"no_session", "cross_origin", "agent", "csrf"} {
		method := "GET"
		if denied == "csrf" {
			method = "POST"
		}
		req, err := http.NewRequest(method, "http://127.0.0.1:17342"+path, nil)
		if err != nil {
			t.Fatal("实际权限请求不可用")
		}
		req.Header.Set("Origin", "http://127.0.0.1:17342")
		if denied == "cross_origin" {
			req.Header.Set("Origin", "https://untrusted.example")
		}
		if denied == "agent" {
			req.Header.Set("Authorization", "Bearer ungranted-public-check")
		}
		transport := client.client
		if denied == "no_session" {
			transport = &http.Client{Timeout: 30 * time.Second}
		}
		response, err := transport.Do(req)
		if err != nil {
			t.Fatal("实际权限请求连接失败")
		}
		response.Body.Close()
		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("实际权限拒绝失败: %s status=%d", denied, response.StatusCode)
		}
	}
	var view workbench.SessionEvents
	client.request(t, "GET", path, nil, &view)
	if view.Total < int(batch.Next.Counts.Projected) || len(view.Events) == 0 || len(view.Coverage) != 1 || view.Coverage[0].Lines != batch.Next.Counts.Lines || view.Coverage[0].Source != "archive" {
		t.Fatal("实际完整文件与 HTTP 覆盖数量不符")
	}
	all := append([]workbench.Event{}, view.Events...)
	for page := 2; len(all) < view.Total; page++ {
		var next workbench.SessionEvents
		client.request(t, "GET", path+"&page="+fmt.Sprint(page), nil, &next)
		if next.Total != view.Total || len(next.Events) == 0 {
			t.Fatal("实际文件 HTTP 分页范围改变")
		}
		all = append(all, next.Events...)
	}
	if len(all) != view.Total {
		t.Fatal("实际文件 HTTP 全部事件范围不符")
	}
	view.Events = all
	committed := make([]workbench.Event, 0, len(view.Events))
	nativeIDs := []string{}
	for _, event := range view.Events {
		row, err := repo.GetEvent(ctx, event.ID)
		if err != nil || row.ID != event.ID || row.Event != nil || row.Observation.Depth != "unknown" || row.Observation.Outcome != "unknown" {
			t.Fatal("实际提交 ID 未在既有仓储解析")
		}
		if strings.HasPrefix(row.OriginKey, "relay-session:") {
			nativeIDs = append(nativeIDs, row.ID)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(row.RawBytes, &fields); err != nil || len(fields) != 6 {
			t.Fatal("实际仓储不满足六字段边界")
		}
		for _, raw := range fields {
			if bytes.Contains(raw, []byte("L1_PUBLIC_CANARY_976E")) {
				t.Fatal("公开输入标记泄露到派生字段")
			}
		}
		var projection workbench.Event
		decoder := json.NewDecoder(bytes.NewReader(row.RawBytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&projection); err != nil {
			t.Fatal("实际提交投影不可解码")
		}
		projection.ID = row.ID
		if !reflect.DeepEqual(projection, event) {
			t.Fatal("HTTP 六字段与独立 GetEvent 不符")
		}
		committed = append(committed, projection)
	}
	for page := 1; page <= len(view.Events); page++ {
		var single workbench.SessionEvents
		client.request(t, "GET", fmt.Sprintf("/api/v1/relay/session-events?task_id=%s&page_size=1&page=%d", task.ID, page), nil, &single)
		if len(single.Events) != 1 || !reflect.DeepEqual(single.Events[0], committed[page-1]) {
			t.Fatal("实际分页改变了提交事件顺序")
		}
	}
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&after); err != nil || after != before {
		t.Fatal("实际静默 GET 期间事件库发生写入")
	}
	registered := struct {
		RegisteredAt string `json:"registered_at"`
	}{}
	readActualJSON(t, filepath.Join(evidence, "registration-public.json"), &registered)
	dispatched := time.Time{}
	for _, receipt := range state.Receipts {
		if receipt.TaskID == task.ID && receipt.Action == "dispatch" && receipt.Status == "accepted" {
			dispatched = receipt.CreatedAt
		}
	}
	if dispatched.IsZero() {
		t.Fatal("实际派发接受收据缺失")
	}
	if len(nativeIDs) != int(batch.Next.Counts.Projected) {
		t.Fatal("实际 Session 提交 ID 未覆盖全部投影")
	}
	checkpoint, err := (relaySessionFileState{root: private.AppRoot}).Load()
	if err != nil || !reflect.DeepEqual(checkpoint[task.ID].Cursor, batch.Next) {
		t.Fatal("实际服务持久游标未覆盖当前原生文件")
	}
	if prior := os.Getenv("ASTROCYTE_L1_PRIOR_STAGE"); prior != "" {
		if prior != "initial" && prior != "append" {
			t.Fatal("实际前阶段必须属于本次验证")
		}
		var previous events.RelaySessionCursor
		readActualJSON(t, filepath.Join(evidence, prior+"-cursor.json"), &previous)
		incremental, err := events.ReadRelaySession(ctx, binding, previous)
		if err != nil || !incremental.SnapshotComplete || incremental.Next.Offset != batch.Next.Offset || incremental.Next.Counts != batch.Next.Counts {
			t.Fatal("实际追加或停止后的增量范围与全文件不符")
		}
		if prior == "initial" && (incremental.Next.Offset <= previous.Offset || len(incremental.Items) == 0) {
			t.Fatal("实际追加没有新原生记录")
		}
		if prior == "append" && owned.ExitedAt == nil {
			t.Fatal("实际停止缺少进程退出证据")
		}
	}
	writeActualJSON(t, filepath.Join(evidence, "current-cursor.json"), batch.Next)
	writeActualJSON(t, filepath.Join(evidence, "actual-model-derived.json"), map[string]any{"models": models, "tool_calls": toolCalls, "signals": signals, "nodes": nodes, "production_assembly": true, "source": "controller_selected_archive"})
	writeActualJSON(t, filepath.Join(evidence, "committed-public.json"), map[string]any{"task_id": task.ID, "project_id": task.ProjectID, "agent_name": task.Agent, "registered_at": registered.RegisteredAt, "dispatched_at": dispatched.UTC().Format(time.RFC3339Nano), "events": committed, "native_session_event_ids": nativeIDs, "archive_event_ids": nativeIDs, "forbidden_canaries": []string{"L1_PUBLIC_CANARY_976E"}, "coverage": view.Coverage})
	writeActualJSON(t, filepath.Join(evidence, "actual-evidence-derived.json"), map[string]any{"lines": batch.Next.Counts.Lines, "projected": batch.Next.Counts.Projected, "unknown": batch.Next.Counts.Unknown, "errors": batch.Next.Counts.Errors, "snapshot_bytes": batch.SnapshotBytes, "snapshot_complete": batch.SnapshotComplete, "session_complete": batch.SessionComplete, "get_event_matches": len(committed), "readonly_get_event_count_unchanged": before == after, "negative_file_checks": 5, "archive_format": binding.Format, "has_private_paths_in_public": strings.Contains(fmt.Sprint(committed), private.AppRoot), "exit_confirmed": owned.ExitedAt != nil, "archive_finalized": owned.ArchiveFinalized, "archive_append_only": owned.ArchiveAppendOnly, "native_version": state.Sessions[task.TerminalID].Version})
}

func readActualJSON(t *testing.T, path string, value any) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal("当前实际验证文件不可用")
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		t.Fatal("当前实际验证文件不可解码")
	}
}
