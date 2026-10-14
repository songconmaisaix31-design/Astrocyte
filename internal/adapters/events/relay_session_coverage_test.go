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
	if e != nil || len(r.Items) != 4 || r.Next.Counts.Unknown != 1 || r.Next.Counts.Duplicates != 0 || r.Next.Counts.Lines != 4 {
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
