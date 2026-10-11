package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func call(id int, name, args string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":{"name":"astrocyte.%s","arguments":%s}}`+"\n", id, name, args)
}

func decodeTool[T any](t *testing.T, response json.RawMessage) T {
	t.Helper()
	var wire struct {
		Error  json.RawMessage `json:"error"`
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(response, &wire); err != nil || len(wire.Error) != 0 || wire.Result.IsError || len(wire.Result.Content) != 1 {
		t.Fatalf("tool failed: %s (%v)", response, err)
	}
	var value T
	if err := json.Unmarshal([]byte(wire.Result.Content[0].Text), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// Builds and runs the shipped command, rather than substituting a test server.
// All events, identities and numerical policies below are contract-local fixtures.
func TestBuiltCommandStdioPersistence(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "astrocyte-mcp.exe")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	data := t.TempDir()
	args := []string{"--data-dir", data, "--project", "fixture-project", "--session", "fixture-session", "--principal-id", "fixture-agent", "--principal-kind", "agent", "--signal-strength", "0.1", "--decay", "0.02", "--initial-confidence", "0.5", "--success-confidence-delta", "0.05", "--confidence-drop-threshold", "0.2"}
	invoke := func(script string) []json.RawMessage {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, args...)
		command.Stdin = bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"fixture","version":"1"}}}` + "\n" + script)
		var out, stderr bytes.Buffer
		command.Stdout, command.Stderr = &out, &stderr
		if err := command.Run(); err != nil {
			t.Fatalf("stdio: %v; stderr=%s; stdout=%s", err, stderr.String(), out.String())
		}
		var replies []json.RawMessage
		scanner := bufio.NewScanner(&out)
		scanner.Buffer(make([]byte, 64*1024), 16<<20)
		for scanner.Scan() {
			if !json.Valid(scanner.Bytes()) {
				t.Fatalf("nonprotocol stdout: %s", scanner.Text())
			}
			replies = append(replies, append(json.RawMessage(nil), scanner.Bytes()...))
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		return replies
	}
	ts := time.Now().UnixMilli()
	// Birth is a human/Sleep use case, not an authority granted to an MCP agent.
	// Seed this test topic through the owner's human gate before agent reporting.
	db, err := sqlite.Open(filepath.Join(data, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	decay, confidence := .02, .5
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence})
	if _, err = engine.SeedTopic(context.Background(), app.Principal{Kind: "human", ID: "fixture-human"}, "fixture-topic", ts); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	raw := fmt.Sprintf(`{ "ts": %d, "kind": "edit_file", "topic": "fixture-topic", "source": "manual", "session": "forged", "raw": { "literal": "\\n", "outcome": "success" } }`, ts)
	pendingRaw := fmt.Sprintf(`{ "ts": %d, "kind": "search" }`, ts)
	responses := invoke(call(2, "report_event", `{"event":`+raw+`}`) + call(3, "report_event", `{"event":`+pendingRaw+`}`) + call(4, "get_attention_state", `{}`) + call(5, "explain_focus", `{"topic":"fixture-topic"}`) + call(6, "veto", `{"topic":"fixture-topic"}`) + call(7, "report_event", fmt.Sprintf(`{"event":{"ts":%d,"kind":"commit","topic":"fixture-topic","scope":"foreign"}}`, ts)))
	if len(responses) != 7 {
		t.Fatalf("responses: %d", len(responses))
	}
	reported := decodeTool[app.ReportEventResult](t, responses[1])
	pending := decodeTool[app.ReportEventResult](t, responses[2])
	if !reported.Accepted || reported.EventID == "" || !pending.Accepted || pending.EventID == "" {
		t.Fatalf("receipts: %+v %+v", reported, pending)
	}
	state := decodeTool[app.GetAttentionStateResult](t, responses[3])
	if len(state.Topics) != 1 || state.Topics[0].Topic != "fixture-topic" || state.Topics[0].Attention <= 0 {
		t.Fatalf("state: %+v", state)
	}
	explanation := decodeTool[app.FocusExplanation](t, responses[4])
	if len(explanation.Evidence) != 1 || explanation.Evidence[0].EventID != reported.EventID || explanation.Evidence[0].Strength != 0.1 || !bytes.Equal(explanation.Evidence[0].Original.RawBytes, []byte(raw)) {
		t.Fatalf("explanation: %+v", explanation)
	}
	for _, rejected := range responses[5:] {
		var response map[string]any
		_ = json.Unmarshal(rejected, &response)
		result, _ := response["result"].(map[string]any)
		if response["error"] == nil && result["isError"] != true {
			t.Fatalf("authority accepted: %s", rejected)
		}
	}
	checkStore := func() {
		db, err := sqlite.Open(filepath.Join(data, "state.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		repo := sqlite.NewEventRepository(db)
		rows, err := repo.ListEvents(context.Background(), ts, ts)
		if err != nil || len(rows) != 2 {
			t.Fatalf("events: %d %v", len(rows), err)
		}
		for _, row := range rows {
			if row.Observation.Source != "mcp" || row.Observation.Session != "fixture-session" || row.Observation.Scope != "fixture-project" || row.Observation.Outcome != "unknown" || row.Observation.Tokens != nil {
				t.Fatalf("binding: %+v", row)
			}
			if row.ID == reported.EventID && !bytes.Equal(row.RawBytes, []byte(raw)) {
				t.Fatalf("raw: %q", row.RawBytes)
			}
			if row.ID == pending.EventID && (row.Event != nil || !bytes.Equal(row.RawBytes, []byte(pendingRaw))) {
				t.Fatalf("pending: %+v", row)
			}
		}
		signals, err := repo.ListSignals(context.Background(), ts, ts)
		if err != nil || len(signals) != 1 {
			t.Fatalf("signals: %d %v", len(signals), err)
		}
	}
	checkStore()
	restarted := invoke(call(2, "get_attention_state", `{}`) + call(3, "explain_focus", `{"topic":"fixture-topic"}`))
	if len(decodeTool[app.GetAttentionStateResult](t, restarted[1]).Topics) != 1 || decodeTool[app.FocusExplanation](t, restarted[2]).Evidence[0].EventID != reported.EventID {
		t.Fatal("restart lost original state")
	}
	checkStore()
	// A new topic does not acquire a node through an ordinary agent report.
	// Its committed receipt must still be recoverable when folding is pending.
	unseededRaw := fmt.Sprintf(`{ "ts": %d, "kind": "edit_file", "topic": "unseeded" }`, ts+1)
	unseeded := invoke(call(2, "report_event", `{"event":`+unseededRaw+`}`))
	var pendingWire struct {
		Result struct {
			IsError bool `json:"isError"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal(unseeded[1], &pendingWire); err != nil {
		t.Fatal(err)
	}
	var delivery app.ReportEventResult
	if len(pendingWire.Result.Content) != 1 || json.Unmarshal([]byte(pendingWire.Result.Content[0].Text), &delivery) != nil || !pendingWire.Result.IsError || !delivery.Accepted || delivery.EventID == "" || delivery.CapturePending || !delivery.DynamicsPending || delivery.DeliveryState != "committed" {
		t.Fatalf("lost pending receipt: %s", unseeded[1])
	}
	db, err = sqlite.Open(filepath.Join(data, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	row, err := sqlite.NewEventRepository(db).GetEvent(context.Background(), delivery.EventID)
	if err != nil || !bytes.Equal(row.RawBytes, []byte(unseededRaw)) {
		t.Fatalf("pending bytes: %+v %v", row, err)
	}
	nodes, err := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{}).AttentionState(context.Background())
	if err != nil || len(nodes) != 1 {
		t.Fatalf("agent created topic: %+v %v", nodes, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, badArgs := range [][]string{args[:len(args)-10], append(append([]string(nil), args...), "--human-veto")} {
		command := exec.Command(binary, badArgs...)
		if output, err := command.CombinedOutput(); err == nil {
			t.Fatalf("invalid startup accepted: %s", output)
		}
	}
}
