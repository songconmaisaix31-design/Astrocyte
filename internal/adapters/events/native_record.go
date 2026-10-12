package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// nativeObservation is adapter-local parsing state, not an engine contract.
type nativeObservation struct {
	session, cwd, kind, outcome, depth string
	ts                                 int64
	tokens                             *int
}

func object(raw json.RawMessage) map[string]json.RawMessage {
	var v map[string]json.RawMessage
	_ = json.Unmarshal(raw, &v)
	return v
}
func str(v map[string]json.RawMessage, key string) string {
	var s string
	_ = json.Unmarshal(v[key], &s)
	return s
}
func millis(raw json.RawMessage) int64 {
	var n int64
	if json.Unmarshal(raw, &n) == nil {
		return n
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func nativeRecord(source string, raw []byte) (nativeObservation, error) {
	if !json.Valid(raw) {
		return nativeObservation{}, fmt.Errorf("invalid native JSON record")
	}
	v := object(raw)
	o := nativeObservation{kind: "unknown", outcome: "unknown", depth: "unknown", ts: millis(v["timestamp"])}
	o.session, o.cwd = str(v, "sessionId"), str(v, "cwd")
	if source == "codex" {
		p := object(v["payload"])
		if str(v, "type") == "session_meta" {
			o.session, o.cwd = str(p, "id"), str(p, "cwd")
			if o.ts == 0 {
				o.ts = millis(p["timestamp"])
			}
		}
		if str(v, "type") == "response_item" && (str(p, "type") == "function_call" || str(p, "type") == "custom_tool_call") {
			arguments := p["arguments"]
			if len(arguments) == 0 {
				arguments = p["input"]
			}
			o.kind = toolKind(str(p, "name"), arguments)
		}
		if str(v, "type") == "response_item" && str(p, "type") == "message" && str(p, "role") == "assistant" {
			o.kind = assistantKind(messageText(p["content"]))
		}
		if str(v, "type") == "event_msg" && str(p, "type") == "agent_message" {
			o.kind = assistantKind(str(p, "message"))
		}
		if str(v, "type") == "event_msg" && str(p, "type") == "token_count" {
			u := object(object(p["info"])["last_token_usage"])
			o.tokens = tokenTotal(u)
		}
	} else if source == "claude_code" {
		m := object(v["message"])
		o.tokens = tokenTotal(object(m["usage"]))
		var blocks []map[string]json.RawMessage
		_ = json.Unmarshal(m["content"], &blocks)
		kinds := map[string]bool{}
		for _, b := range blocks {
			if str(v, "type") == "assistant" && str(b, "type") == "tool_use" {
				kinds[toolKind(str(b, "name"), b["input"])] = true
			}
		}
		if len(kinds) == 1 {
			for k := range kinds {
				o.kind = k
			}
		}
		if len(kinds) == 0 && str(v, "type") == "assistant" {
			o.kind = assistantKind(messageText(m["content"]))
		}
	} else {
		return o, fmt.Errorf("unsupported native source")
	}
	return o, nil
}

// A missing usage component remains unknown; cache and cumulative totals are
// not substituted for per-record input/output usage.
func tokenTotal(u map[string]json.RawMessage) *int {
	var total int
	if string(u["total_tokens"]) != "null" && json.Unmarshal(u["total_tokens"], &total) == nil && total >= 0 {
		return &total
	}
	var input, output int
	if string(u["input_tokens"]) == "null" || string(u["output_tokens"]) == "null" || json.Unmarshal(u["input_tokens"], &input) != nil || json.Unmarshal(u["output_tokens"], &output) != nil || input < 0 || output < 0 {
		return nil
	}
	total = input + output
	return &total
}
