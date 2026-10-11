package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"
)

// serveLines runs Serve against a real os.Pipe pair, feeds newline-delimited
// input and returns every decoded response line.
func serveLines(t *testing.T, s *Server, input string) []map[string]any {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	or, ow, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		err := s.Serve(context.Background(), pr, ow)
		ow.Close()
		done <- err
	}()
	if _, err := pw.WriteString(input); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	if err := <-done; err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var lines []map[string]any
	scanner := bufio.NewScanner(or)
	for scanner.Scan() {
		var m map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			t.Fatalf("bad response line %q: %v", scanner.Text(), err)
		}
		lines = append(lines, m)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	or.Close()
	return lines
}

func reqLine(id int, method string, params string) string {
	if params == "" {
		params = "{}"
	}
	return "{\"jsonrpc\":\"2.0\",\"id\":" + jsonInt(id) + ",\"method\":\"" + method + "\",\"params\":" + params + "}\n"
}

func jsonInt(id int) string {
	b, _ := json.Marshal(id)
	return string(b)
}

func TestInitializeAndToolsList(t *testing.T) {
	s := NewServer("test", "0.0.0")
	s.RegisterTool(Tool{Name: "t", Description: "d", InputSchema: map[string]any{"type": "object"}, Handler: func(ctx context.Context, a json.RawMessage) (any, error) { return nil, nil }})

	input := reqLine(1, "initialize", "") + reqLine(2, "tools/list", "")
	lines := serveLines(t, s, input)

	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(lines))
	}
	init := lines[0]["result"].(map[string]any)
	if init["protocolVersion"] == nil {
		t.Fatalf("initialize missing protocolVersion: %v", init)
	}
	ti := init["capabilities"].(map[string]any)["tools"]
	if ti == nil {
		t.Fatalf("initialize missing capabilities.tools: %v", init)
	}
	if init["serverInfo"].(map[string]any)["name"] != "test" {
		t.Fatalf("wrong server name: %v", init["serverInfo"])
	}

	list := lines[1]["result"].(map[string]any)
	tools, ok := list["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %v", list["tools"])
	}
	if tools[0].(map[string]any)["name"] != "t" {
		t.Fatalf("wrong tool name: %v", tools[0])
	}
}

func TestToolCallSuccess(t *testing.T) {
	s := NewServer("test", "0.0.0")
	s.RegisterTool(Tool{
		Name:        "echo",
		Description: "echo",
		InputSchema: map[string]any{"type": "object"},
		Handler: func(ctx context.Context, a json.RawMessage) (any, error) {
			return map[string]any{"got": string(a)}, nil
		},
	})

	input := reqLine(1, "tools/call", `{"name":"echo","arguments":{"x":1}}`)
	lines := serveLines(t, s, input)
	if len(lines) != 1 {
		t.Fatalf("expected 1 response, got %d", len(lines))
	}
	result := lines[0]["result"].(map[string]any)
	content := result["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("expected 1 content item: %v", content)
	}
	item := content[0].(map[string]any)
	if item["type"] != "text" {
		t.Fatalf("expected text content: %v", item)
	}
	text := item["text"].(string)
	var got map[string]any
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("result not JSON: %s", text)
	}
	if got["got"] != `{"x":1}` {
		t.Fatalf("arguments not round-tripped: %s", text)
	}
	if result["isError"] != false {
		t.Fatalf("expected isError false: %v", result)
	}
}

func TestToolCallHandlerError(t *testing.T) {
	s := NewServer("test", "0.0.0")
	s.RegisterTool(Tool{
		Name:        "fail",
		Description: "fail",
		InputSchema: map[string]any{"type": "object"},
		Handler: func(ctx context.Context, a json.RawMessage) (any, error) {
			return nil, context.DeadlineExceeded
		},
	})

	input := reqLine(1, "tools/call", `{"name":"fail","arguments":{}}`)
	lines := serveLines(t, s, input)
	result := lines[0]["result"].(map[string]any)
	if result["isError"] != true {
		t.Fatalf("expected isError true: %v", result)
	}
	content := result["content"].([]any)
	if content[0].(map[string]any)["text"] == "" {
		t.Fatalf("expected error text: %v", content)
	}
}

func TestUnknownTool(t *testing.T) {
	s := NewServer("test", "0.0.0")
	input := reqLine(1, "tools/call", `{"name":"nope","arguments":{}}`)
	lines := serveLines(t, s, input)
	if lines[0]["error"] == nil {
		t.Fatalf("expected error: %v", lines[0])
	}
	errObj := lines[0]["error"].(map[string]any)
	if errObj["code"] != float64(codeInvalidParams) {
		t.Fatalf("expected invalid params code, got %v", errObj["code"])
	}
}

func TestUnknownMethod(t *testing.T) {
	s := NewServer("test", "0.0.0")
	input := reqLine(1, "nope", "")
	lines := serveLines(t, s, input)
	errObj := lines[0]["error"].(map[string]any)
	if errObj["code"] != float64(codeMethodNotFound) {
		t.Fatalf("expected method-not-found code, got %v", errObj["code"])
	}
}

func TestNotificationGetsNoResponse(t *testing.T) {
	s := NewServer("test", "0.0.0")
	input := "{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n" + reqLine(1, "ping", "")
	lines := serveLines(t, s, input)
	if len(lines) != 1 {
		t.Fatalf("expected only the ping response, got %d", len(lines))
	}
	if lines[0]["result"] == nil {
		t.Fatalf("expected ping result: %v", lines[0])
	}
}
