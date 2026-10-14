package main

import (
	"bytes"
	"context"
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
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 读取当前实际控制器已提交的派生通知，验证来源选择、持久游标和收据身份。
func TestRelaySessionActualNativeNotificationEvidence(t *testing.T) {
	evidence := os.Getenv("ASTROCYTE_L1_NATIVE_EVIDENCE_DIR")
	if evidence == "" {
		t.Skip("NOT_RUN: explicit actual native notification evidence directory required")
	}
	parent, err := filepath.Abs(filepath.Join("..", "..", "dist", "l1-live"))
	if err != nil || !relayInside(parent, evidence) {
		t.Fatal("实际通知验证目录超出本轮范围")
	}
	var private struct {
		AppRoot    string `json:"app_root"`
		DBPath     string `json:"db_path"`
		SourceRoot string `json:"source_root"`
	}
	readActualJSON(t, filepath.Join(evidence, "service-private.json"), &private)
	if !relayInside(filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification"), private.AppRoot) || private.DBPath != filepath.Join(private.AppRoot, "app-state", "state.sqlite") {
		t.Fatal("实际通知数据库目录超出本轮范围")
	}
	t.Setenv("LOCALAPPDATA", filepath.Dir(filepath.Dir(private.AppRoot)))
	store, err := sqlite.OpenRelayStore(private.AppRoot)
	if err != nil {
		t.Fatal("实际通知运行时数据库不可用")
	}
	defer store.Close()
	ctx := context.Background()
	state, err := store.Snapshot(ctx)
	if err != nil || len(state.Tasks) != 1 {
		t.Fatal("实际通知任务身份不可核实")
	}
	var task domain.RelayTask
	for _, item := range state.Tasks {
		task = item
	}
	owned := state.OwnedSessions[task.ID]
	if owned.NativeCapture == nil || owned.NativeCapture.SourceStatus != "available" || owned.NativeCapture.PrefixComplete != "confirmed" {
		t.Fatal("实际通知来源前缀未确认")
	}
	source := domain.RelayNativeSessionBatch{OwnedSession: owned, RelayNativeSessionCapture: *owned.NativeCapture}
	db, err := sqlite.Open(private.DBPath)
	if err != nil {
		t.Fatal("实际通知事件仓储不可用")
	}
	defer db.Close()
	repo := sqlite.NewEventRepository(db)
	loaded, err := (relaySessionFileState{root: private.AppRoot}).Load()
	if err != nil || loaded[task.ID].Native == nil {
		t.Fatal("实际通知游标未持久化")
	}
	checkpoint := *loaded[task.ID].Native
	if checkpoint.BindingID != owned.BindingID || checkpoint.Sequence != source.LastSequence {
		t.Fatal("实际通知序号与持久位置不符")
	}
	capture := events.NewRelaySessionCapture(attention.NewPendingEventService(repo), nil)
	restarted, err := capture.CaptureNative(ctx, relayNativeCaptureBatch(source), checkpoint)
	if err != nil || !reflect.DeepEqual(restarted, checkpoint) {
		t.Fatal("实际通知重读改写了位置或计数")
	}
	invalid := checkpoint
	invalid.Sequence++
	failure, err := capture.CaptureNative(ctx, relayNativeCaptureBatch(source), invalid)
	if err == nil || failure.Sequence != invalid.Sequence || failure.Blocked != "native_sequence_invalid" {
		t.Fatal("实际通知游标负例未立即失败")
	}
	client := newActualRelayHTTP(t, actualRelayAPIURL(t, evidence))
	client.request(t, "GET", "/api/v1/auth/session", nil, nil)
	var view workbench.SessionEvents
	client.request(t, "GET", "/api/v1/relay/session-events?task_id="+task.ID+"&page_size=100", nil, &view)
	if view.Total == 0 || len(view.Coverage) != 1 || view.Coverage[0].Lines != source.LastSequence {
		t.Fatal("实际通知 HTTP 覆盖不可核实")
	}
	all := append([]workbench.Event{}, view.Events...)
	for page := 2; len(all) < view.Total; page++ {
		var next workbench.SessionEvents
		client.request(t, "GET", "/api/v1/relay/session-events?task_id="+task.ID+"&page_size=100&page="+fmt.Sprint(page), nil, &next)
		if next.Total != view.Total || len(next.Events) == 0 {
			t.Fatal("实际通知分页范围改变")
		}
		all = append(all, next.Events...)
	}
	if len(all) != view.Total {
		t.Fatal("实际通知分页数量不符")
	}
	nativeIDs := []string{}
	for _, event := range all {
		row, err := repo.GetEvent(ctx, event.ID)
		if err != nil || row.ID != event.ID || row.Event != nil || row.Observation.Outcome != "unknown" || row.Observation.Depth != "unknown" {
			t.Fatal("实际通知提交 ID 未独立解析")
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(row.RawBytes, &fields); err != nil || len(fields) != 6 {
			t.Fatal("实际通知不满足六字段边界")
		}
		if bytes.Contains(row.RawBytes, []byte("L1_PUBLIC_CANARY_976E")) || bytes.Contains(row.RawBytes, []byte(private.AppRoot)) {
			t.Fatal("实际原文或私有路径泄露到通知投影")
		}
		var projection workbench.Event
		decoder := json.NewDecoder(bytes.NewReader(row.RawBytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&projection); err != nil {
			t.Fatal("实际通知投影不可解码")
		}
		projection.ID = row.ID
		if !reflect.DeepEqual(projection, event) {
			t.Fatal("实际通知 HTTP 六字段与 GetEvent 不符")
		}
		if strings.HasPrefix(row.OriginKey, "relay-session:") || row.Observation.Source == "native" {
			t.Fatal("实际通知选择混入原档或任务级原生重复事件")
		}
		if strings.HasPrefix(row.OriginKey, "relay-native-session:") {
			nativeIDs = append(nativeIDs, row.ID)
		}
	}
	if len(nativeIDs) != int(source.LastSequence) {
		t.Fatal("实际通知全部序号未对应提交记录")
	}
	registration := struct {
		RegisteredAt string `json:"registered_at"`
	}{}
	readActualJSON(t, filepath.Join(evidence, "registration-public.json"), &registration)
	dispatched := time.Time{}
	for _, receipt := range state.Receipts {
		if receipt.TaskID == task.ID && receipt.Action == "dispatch" && receipt.Status == "accepted" {
			dispatched = receipt.CreatedAt
		}
	}
	writeActualJSON(t, filepath.Join(evidence, "committed-public.json"), map[string]any{"task_id": task.ID, "project_id": task.ProjectID, "agent_name": task.Agent, "registered_at": registration.RegisteredAt, "dispatched_at": dispatched.UTC().Format(time.RFC3339Nano), "events": all, "native_session_event_ids": nativeIDs, "archive_event_ids": []string{}, "forbidden_canaries": []string{"L1_PUBLIC_CANARY_976E"}, "coverage": view.Coverage, "test_liveness_window_seconds": 10, "source": "owned_app_server_notifications"})
	writeActualJSON(t, filepath.Join(evidence, "native-evidence-derived.json"), map[string]any{"last_sequence": source.LastSequence, "prefix_complete": source.PrefixComplete, "finalized": source.Finalized, "unknown": source.Unknown, "errors": source.Errors, "first_gap": source.FirstGap, "get_event_matches": len(nativeIDs), "mixed_archive_or_native_summary": false, "checkpoint_restart_unchanged": true})
}
