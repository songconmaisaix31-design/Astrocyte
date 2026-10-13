package events

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestRelaySessionCorrelationIDsDoNotCollapseRecords(t *testing.T) {
	b := sessionFixture(t)
	call := sessionCall("shared-call")
	output := sessionJSON(map[string]any{"type": "response_item", "timestamp": "2026-10-14T01:00:02Z", "payload": map[string]any{"type": "function_call_output", "call_id": "shared-call", "output": "PRIVATE_CANARY"}})
	appendSession(t, b, call+output+call)
	r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if e != nil || len(r.Items) != 4 || r.Next.Counts.Unknown != 0 || r.Next.Counts.Duplicates != 0 || r.Next.Counts.Lines != 4 {
		t.Fatal("call correlation collapsed separate records")
	}
	ids := map[string]bool{}
	for _, v := range r.Items {
		if ids[v.Origin] {
			t.Fatal("record origins collided")
		}
		ids[v.Origin] = true
	}
	claude := sessionFixture(t)
	claude.Format = "claude_code"
	part := func(kind string) string {
		return sessionJSON(map[string]any{"type": "assistant", "sessionId": claude.NativeID, "cwd": claude.WorkspaceRoot, "timestamp": "2026-10-14T01:00:01Z", "message": map[string]any{"id": "shared-message", "content": []any{map[string]any{"type": "tool_use", "name": kind, "input": "PRIVATE_CANARY"}}, "usage": map[string]any{"input_tokens": 2, "output_tokens": 3}}})
	}
	writeSession(t, claude, part("Edit")+part("Grep"))
	c, e := ReadRelaySession(context.Background(), claude, RelaySessionCursor{})
	if e != nil || len(c.Items) != 2 || c.Next.Counts.Unknown != 0 || c.Items[0].Origin == c.Items[1].Origin || c.Items[0].Projection.Tokens == nil || *c.Items[0].Projection.Tokens != 5 {
		t.Fatal("message correlation collapsed separate parts")
	}
}

func TestRelaySessionTokenCoverageRejectsOverflowAndNegativeComponents(t *testing.T) {
	for _, usage := range []string{`{"input_tokens":-1,"output_tokens":1,"total_tokens":2}`, `{"input_tokens":` + strconv.Itoa(math.MaxInt) + `,"output_tokens":1}`, `{"input_tokens":null,"output_tokens":1}`, `{"total_tokens":1.5}`, `{"input_tokens":1}`} {
		b := sessionFixture(t)
		appendSession(t, b, sessionJSON(map[string]any{"type": "event_msg", "timestamp": "2026-10-14T01:00:01Z", "payload": map[string]any{"type": "token_count", "info": map[string]any{"last_token_usage": json.RawMessage(usage)}}}))
		r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
		if e != nil || r.Next.Counts.Lines != 2 || r.Next.Counts.Unknown != 1 || r.Items[1].Projection.Tokens != nil || r.SessionComplete {
			t.Fatal("invalid usage invented token total or dropped coverage")
		}
	}
}

func TestRelaySessionPublicCodexEnvelopesKeepUnknownSemantics(t *testing.T) {
	b := sessionFixture(t)
	response := func(payload map[string]any) string {
		return sessionJSON(map[string]any{"type": "response_item", "timestamp": "2026-10-14T01:00:01Z", "payload": payload})
	}
	event := func(payload map[string]any) string {
		return sessionJSON(map[string]any{"type": "event_msg", "timestamp": "2026-10-14T01:00:02Z", "payload": payload})
	}
	appendSession(t, b, response(map[string]any{"type": "message", "role": "user", "content": []any{map[string]any{"type": "input_text", "text": "PRIVATE_CANARY user prompt"}}})+
		response(map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "PRIVATE_CANARY assistant body"}}})+
		response(map[string]any{"type": "reasoning", "summary": []any{map[string]any{"type": "summary_text", "text": "PRIVATE_CANARY thinking"}}})+
		response(map[string]any{"type": "function_call_output", "call_id": "shared", "output": []any{map[string]any{"type": "input_text", "text": "PRIVATE_CANARY tool output"}}})+
		sessionJSON(map[string]any{"type": "turn_context", "timestamp": "2026-10-14T01:00:02Z", "payload": map[string]any{"cwd": b.WorkspaceRoot, "model": "public-model", "approval_policy": "never", "sandbox_policy": map[string]any{"type": "read-only"}}})+
		event(map[string]any{"type": "agent_message", "message": "PRIVATE_CANARY answer"})+
		event(map[string]any{"type": "user_message", "message": "PRIVATE_CANARY input"})+
		event(map[string]any{"type": "task_started", "turn_id": "PRIVATE_CANARY turn"})+
		event(map[string]any{"type": "task_complete", "turn_id": "PRIVATE_CANARY turn", "last_agent_message": "PRIVATE_CANARY", "error": map[string]any{"message": "PRIVATE_CANARY failed turn"}})+
		event(map[string]any{"type": "token_count", "info": nil}))
	r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if e != nil || r.Next.Counts.Lines != 11 || r.Next.Counts.Unknown != 0 || len(r.Items) != 11 || r.SessionComplete || !r.SnapshotComplete {
		t.Fatal("normal native envelopes confused structural coverage or completion")
	}
	seen := map[string]bool{}
	for _, item := range r.Items {
		seen[item.Projection.Kind] = true
		data, _ := json.Marshal(item.Projection)
		var fields map[string]any
		_ = json.Unmarshal(data, &fields)
		if len(fields) != 6 || strings.Contains(string(data), "PRIVATE_CANARY") {
			t.Fatal("normal envelope exported body or private metadata")
		}
	}
	if !seen["message"] || !seen["task_complete"] || !seen["turn_context"] || !seen["function_call_output"] || !seen["reasoning"] {
		t.Fatal("native envelope identity lost")
	}
	appendSession(t, b, response(map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "future_content_PRIVATE_CANARY", "body": "PRIVATE_CANARY"}}})+event(map[string]any{"type": "never_seen_PRIVATE_CANARY", "message": "PRIVATE_CANARY"}))
	unknown, e := ReadRelaySession(context.Background(), b, r.Next)
	if e != nil || unknown.Next.Counts.Unknown != 2 || unknown.Next.Counts.Lines != 13 || unknown.SessionComplete {
		t.Fatal("future nested or event schema silently accepted")
	}
}

func TestRelaySessionCancelledRead(t *testing.T) {
	b := sessionFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, e := ReadRelaySession(ctx, b, RelaySessionCursor{})
	if e == nil || r.Next.Offset != 0 || r.SessionComplete || strings.Contains(e.Error(), "PRIVATE_CANARY") {
		t.Fatal("cancelled source read advanced")
	}
}

func TestRelaySessionRedirectedRead(t *testing.T) {
	b := sessionFixture(t)
	alias := b.Path + ".symlink"
	if os.Symlink(b.Path, alias) != nil {
		t.Skip("NOT_RUN: symlink privilege unavailable")
	}
	b.Path = alias
	r, e := ReadRelaySession(context.Background(), b, RelaySessionCursor{})
	if e == nil || r.Next.Offset != 0 || r.SessionComplete {
		t.Fatal("redirected archive accepted")
	}
}
