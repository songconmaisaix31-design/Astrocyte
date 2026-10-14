package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// nativeObservation 保存适配器的原生格式解析状态。
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
	if err := validateNativeFormat(raw); err != nil {
		return nativeObservation{}, err
	}
	var v nativeEnvelope
	if err := json.Unmarshal(raw, &v); err != nil {
		return nativeObservation{}, err
	}
	timestamp, err := nativeTimestamp(v.Timestamp)
	if err != nil {
		return nativeObservation{}, err
	}
	o := nativeObservation{kind: "unknown", outcome: "unknown", depth: "unknown", ts: timestamp, session: v.SessionID, cwd: v.CWD}
	if source == "codex" {
		p := v.Payload
		if p == nil {
			return o, nil
		}
		if v.Type == "session_meta" {
			o.session, o.cwd = p.ID, p.CWD
			if o.ts == 0 {
				o.ts, err = nativeTimestamp(p.Timestamp)
				if err != nil {
					return nativeObservation{}, err
				}
			}
		}
		if v.Type == "response_item" && (p.Type == "function_call" || p.Type == "custom_tool_call") {
			arguments := p.Arguments
			if len(arguments) == 0 {
				arguments = p.Input
			}
			name := canonicalTool(p.Name)
			if name == "Bash" || name == "exec_command" || name == "shell_command" || name == "Read" {
				if err := validateToolObject(arguments); err != nil {
					return nativeObservation{}, err
				}
			}
			o.kind = toolKind(p.Name, arguments)
		}
		if v.Type == "response_item" && p.Type == "message" && p.Role == "assistant" {
			text, _, err := nativeContent(p.Content)
			if err != nil {
				return nativeObservation{}, err
			}
			o.kind = assistantKind(text)
		}
		if v.Type == "event_msg" && p.Type == "agent_message" {
			o.kind = assistantKind(p.Message)
		}
		if v.Type == "event_msg" && p.Type == "token_count" && p.Info != nil {
			o.tokens = nativeTokenTotal(p.Info.Usage)
		}
	} else if source == "claude_code" {
		m := v.Message
		if m == nil {
			return o, nil
		}
		o.tokens = nativeTokenTotal(m.Usage)
		text, blocks, err := nativeContent(m.Content)
		if err != nil {
			return nativeObservation{}, err
		}
		kinds := map[string]bool{}
		for _, b := range blocks {
			if v.Type == "assistant" && b.Type == "tool_use" {
				kinds[toolKind(b.Name, b.Input)] = true
			}
		}
		if len(kinds) == 1 {
			for k := range kinds {
				o.kind = k
			}
		}
		if len(kinds) == 0 && v.Type == "assistant" {
			o.kind = assistantKind(text)
		}
	} else {
		return o, fmt.Errorf("unsupported native source")
	}
	return o, nil
}

// 缺失的 usage 分量保持未知。
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
