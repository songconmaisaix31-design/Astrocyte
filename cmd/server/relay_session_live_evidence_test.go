package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
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
	if err != nil || evidence != expected {
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
	if err != nil || len(state.Tasks) != 1 {
		t.Fatal("实际任务身份不可核实")
	}
	var task domain.RelayTask
	for _, v := range state.Tasks {
		task = v
	}
	owned, ok := state.OwnedSessions[task.ID]
	if !ok || owned.TaskID != task.ID || owned.SourceStatus != "available" {
		t.Fatal("实际 owned archive 未选择")
	}
	binding := relayCaptureBinding(owned)
	batch, err := events.ReadRelaySession(ctx, binding, events.RelaySessionCursor{})
	if err != nil || !batch.HeaderVerified || !batch.SnapshotComplete || batch.Next.Offset != batch.SnapshotBytes || batch.TailBytes != 0 || batch.Next.Counts.Lines < 2 {
		t.Fatal("实际归档当前全部完整行未覆盖")
	}
	if batch.SessionComplete {
		t.Fatal("实际归档缺封存及前缀证据，不能宣称完整 Session")
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
	var before, after int
	if err := readonly.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&before); err != nil {
		t.Fatal("实际事件计数不可用")
	}
	client := newActualRelayHTTP(t)
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	var view workbench.SessionEvents
	client.request(t, "GET", "/api/v1/relay/session-events?task_id="+task.ID+"&page_size=100", nil, &view)
	if view.Total < int(batch.Next.Counts.Projected) || len(view.Events) != view.Total || len(view.Coverage) != 1 || view.Coverage[0].Lines != batch.Next.Counts.Lines {
		t.Fatal("实际完整文件与 HTTP 覆盖数量不符")
	}
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
	writeActualJSON(t, filepath.Join(evidence, "committed-public.json"), map[string]any{"task_id": task.ID, "project_id": task.ProjectID, "agent_name": task.Agent, "registered_at": registered.RegisteredAt, "dispatched_at": dispatched.UTC().Format(time.RFC3339Nano), "events": committed, "native_session_event_ids": nativeIDs, "archive_event_ids": nativeIDs, "forbidden_canaries": []string{"L1_PUBLIC_CANARY_976E"}, "coverage": view.Coverage})
	writeActualJSON(t, filepath.Join(evidence, "actual-evidence-derived.json"), map[string]any{"lines": batch.Next.Counts.Lines, "projected": batch.Next.Counts.Projected, "unknown": batch.Next.Counts.Unknown, "errors": batch.Next.Counts.Errors, "snapshot_bytes": batch.SnapshotBytes, "snapshot_complete": batch.SnapshotComplete, "session_complete": batch.SessionComplete, "get_event_matches": len(committed), "readonly_get_event_count_unchanged": before == after, "negative_file_checks": 5, "archive_format": binding.Format, "has_private_paths_in_public": strings.Contains(fmt.Sprint(committed), private.AppRoot)})
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
