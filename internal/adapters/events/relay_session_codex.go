package events

import (
	"bytes"
	"encoding/json"
)

// These closed structural envelopes come from OpenAI's public rust-v0.162.0
// protocol.rs and models.rs. They identify native records, never tool success,
// task completion, Signal strength or approval to retain their original bodies.
func sessionCodexEnvelope(kind string, p map[string]json.RawMessage, o *nativeObservation) bool {
	stringField := func(key string) bool {
		raw, ok := p[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false
		}
		var value string
		return json.Unmarshal(raw, &value) == nil
	}
	arrayField := func(key string) bool { return sessionJSONArray(p[key]) }
	switch kind {
	case "session_meta":
		o.kind = "native_session_header"
		return str(p, "id") != "" && str(p, "cwd") != ""
	case "turn_context":
		o.kind = "turn_context"
		return str(p, "cwd") != "" && str(p, "model") != "" && len(p["approval_policy"]) > 0 && object(p["sandbox_policy"]) != nil
	case "response_item":
		switch str(p, "type") {
		case "function_call":
			return str(p, "name") != "" && str(p, "call_id") != "" && stringField("arguments")
		case "function_call_output":
			o.kind = "function_call_output"
			return stringField("output") || arrayField("output")
		case "custom_tool_call":
			o.kind = "custom_tool_call"
			return str(p, "name") != "" && str(p, "call_id") != "" && stringField("input")
		case "custom_tool_call_output":
			o.kind = "custom_tool_call_output"
			return str(p, "call_id") != "" && (stringField("output") || arrayField("output"))
		case "message":
			o.kind = "message"
			return str(p, "role") != "" && sessionCodexContent(p["content"])
		case "agent_message":
			o.kind = "agent_message"
			return str(p, "author") != "" && str(p, "recipient") != "" && sessionCodexContent(p["content"])
		case "reasoning":
			o.kind = "reasoning"
			return arrayField("summary")
		}
	case "event_msg":
		switch str(p, "type") {
		case "token_count":
			o.kind = "native_activity"
			if raw, ok := p["info"]; !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				o.tokens = nil
				return true
			}
			info := object(p["info"])
			if info == nil {
				return false
			}
			raw, ok := info["last_token_usage"]
			if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				o.tokens = nil
				return true
			}
			var valid bool
			o.tokens, valid = sessionTokenTotal(object(raw))
			return valid
		case "agent_message", "user_message":
			o.kind = str(p, "type")
			return stringField("message")
		case "agent_reasoning", "agent_reasoning_raw_content":
			o.kind = str(p, "type")
			return stringField("text")
		case "task_started", "turn_started", "task_complete", "turn_complete":
			o.kind = str(p, "type")
			return str(p, "turn_id") != ""
		case "context_compacted":
			o.kind = "context_compacted"
			return true
		case "error", "warning":
			o.kind = str(p, "type")
			return stringField("message")
		}
	}
	return false
}

func sessionJSONArray(raw json.RawMessage) bool {
	var items []json.RawMessage
	return len(bytes.TrimSpace(raw)) > 0 && bytes.TrimSpace(raw)[0] == '[' && json.Unmarshal(raw, &items) == nil
}

func sessionCodexContent(raw json.RawMessage) bool {
	if !sessionJSONArray(raw) {
		return false
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(raw, &items) != nil {
		return false
	}
	for _, item := range items {
		switch str(item, "type") {
		case "input_text", "output_text":
			var value string
			if json.Unmarshal(item["text"], &value) != nil || string(item["text"]) == "null" {
				return false
			}
		default:
			return false // unverified content variants retain an explicit gap
		}
	}
	return true
}
