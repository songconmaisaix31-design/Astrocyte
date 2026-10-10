package agents

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type contextRPCWriter struct {
	p     *nativeProcess
	reply func(map[string]json.RawMessage) map[string]json.RawMessage
}

func (w contextRPCWriter) Close() error { return nil }
func (w contextRPCWriter) Write(data []byte) (int, error) {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(data, &request); err != nil {
		return 0, err
	}
	var id string
	_ = json.Unmarshal(request["id"], &id)
	response := w.reply(request)
	w.p.mu.Lock()
	ch := w.p.pending[id]
	w.p.mu.Unlock()
	ch <- response
	return len(data), nil
}
func contextTestProcess(t *testing.T, cli string, reply func(map[string]json.RawMessage) map[string]json.RawMessage) (*Native, domain.NativeSession) {
	t.Helper()
	n := NewRegistry(t.TempDir()).adapters[cli]
	p := &nativeProcess{cli: cli, nativeID: "original", cmd: &exec.Cmd{Dir: t.TempDir()}, done: make(chan struct{}), pending: map[string]chan map[string]json.RawMessage{}, status: "idle"}
	p.stdin = contextRPCWriter{p: p, reply: reply}
	n.processes["owned"] = p
	return n, domain.NativeSession{ID: "owned", NativeID: "original", Ownership: "owned"}
}

func TestCodexNativeContextUsesOwnedIdentityAndAllBoundedPages(t *testing.T) {
	calls := []string{}
	var directory string
	n, session := contextTestProcess(t, "codex", func(r map[string]json.RawMessage) map[string]json.RawMessage {
		var method string
		_ = json.Unmarshal(r["method"], &method)
		calls = append(calls, method)
		var params map[string]any
		_ = json.Unmarshal(r["params"], &params)
		if params["threadId"] != "original" {
			t.Fatal("context request escaped owned native ID")
		}
		if method == "thread/read" {
			result, _ := json.Marshal(map[string]any{"thread": map[string]any{"id": "original", "cwd": directory}})
			return map[string]json.RawMessage{"result": result}
		}
		if method != "thread/items/list" || params["limit"] != float64(32) || params["sortDirection"] != "asc" {
			t.Fatal("unbounded/unknown native history call")
		}
		if params["cursor"] == nil {
			return map[string]json.RawMessage{"result": json.RawMessage(`{"data":[{"item":{"type":"userMessage","content":[{"type":"text","text":"prior context"}]}},{"item":{"type":"commandExecution","text":"do not expose tool output"}}],"nextCursor":"next"}`)}
		}
		return map[string]json.RawMessage{"result": json.RawMessage(`{"data":[{"item":{"type":"agentMessage","text":"prior reply"}}],"nextCursor":null}`)}
	})
	directory = n.processes["owned"].cmd.Dir
	events, err := n.ReadContext(context.Background(), session)
	if err != nil || len(events) != 2 || events[0].Kind != "user_text" || events[0].Text != "prior context" || events[1].Text != "prior reply" || len(calls) != 3 {
		t.Fatalf("original native context was incomplete: %+v %v", events, err)
	}
}

func TestNativeContextRejectsForeignMetadataAndDoesNotExposePartialOverflow(t *testing.T) {
	n, s := contextTestProcess(t, "codex", func(map[string]json.RawMessage) map[string]json.RawMessage {
		return map[string]json.RawMessage{"result": json.RawMessage(`{"thread":{"id":"foreign","cwd":"foreign"}}`)}
	})
	if events, err := n.ReadContext(context.Background(), s); err == nil || len(events) != 0 {
		t.Fatal("foreign native history was exposed")
	}
	n, s = contextTestProcess(t, "pi", func(r map[string]json.RawMessage) map[string]json.RawMessage {
		if string(r["type"]) != `"get_messages"` {
			t.Fatal("Pi context read sent an inference command")
		}
		data, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": "prior native input"}, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": strings.Repeat("x", 128*1024)}}}}})
		return map[string]json.RawMessage{"data": data, "success": json.RawMessage("true")}
	})
	if events, err := n.ReadContext(context.Background(), s); err == nil || len(events) != 0 {
		t.Fatal("partial native context was misrepresented as complete after overflow")
	}
	if n.Capabilities()["read_context"].Status != "unknown" {
		t.Fatal("failed projection fabricated native read support")
	}
}
