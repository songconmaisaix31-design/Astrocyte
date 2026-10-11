package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLifecycleAndVersionNegotiation(t *testing.T) {
	for _, version := range []string{"2024-11-05", "2025-03-26", "2025-06-18", "2025-11-25", "unknown"} {
		t.Run(version, func(t *testing.T) {
			params := `{"protocolVersion":"` + version + `","capabilities":{},"clientInfo":{"name":"client","version":"1"}}`
			lines := serveLines(t, NewServer("test", "1"), reqLine(1, "initialize", params))
			want := version
			if version == "unknown" {
				want = protocolVersion
			}
			if got := lines[0]["result"].(map[string]any)["protocolVersion"]; got != want {
				t.Fatalf("version %v, want %v", got, want)
			}
		})
	}
	var output strings.Builder
	err := NewAstrocyteServer(stubBackend{}).Serve(context.Background(), strings.NewReader(reqLine(1, "tools/call", `{"name":"astrocyte.get_attention_state","arguments":{}}`)), &output)
	if err != nil {
		t.Fatal(err)
	}
	var got response
	if err := json.Unmarshal([]byte(output.String()), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error == nil || got.Error.Code != codeInvalidRequest {
		t.Fatalf("uninitialized request: %s", output.String())
	}
}

func TestInvalidToolParameters(t *testing.T) {
	for _, args := range []string{`{}`, `{"topic":""}`, `{"topic":"valid","scope":"foreign"}`, `{"topic":1}`, `null`} {
		lines := serveLines(t, NewAstrocyteServer(stubBackend{}), reqLine(1, "tools/call", `{"name":"astrocyte.explain_focus","arguments":`+args+`}`))
		if lines[0]["error"] == nil || lines[0]["error"].(map[string]any)["code"] != float64(codeInvalidParams) {
			t.Fatalf("invalid arguments accepted: %s: %v", args, lines)
		}
	}
}

func TestInvalidRPCMessages(t *testing.T) {
	for _, message := range []string{`{`, `[]`, `null`, `{"jsonrpc":"1.0","id":1,"method":"ping"}`, `{"jsonrpc":"2.0","id":null,"method":"ping"}`} {
		var output strings.Builder
		if err := NewServer("test", "1").Serve(context.Background(), strings.NewReader(message+"\n"), &output); err != nil {
			t.Fatal(err)
		}
		var got response
		if err := json.Unmarshal([]byte(output.String()), &got); err != nil {
			t.Fatal(err)
		}
		if got.Error == nil {
			t.Fatalf("invalid RPC accepted: %s", message)
		}
	}
}
