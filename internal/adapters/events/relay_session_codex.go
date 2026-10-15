package events

import (
	"bytes"
	"encoding/json"
)

// 仅识别公开协议结构，不推断工具结果、任务成功或认知信号。
func sessionCodexEnvelope(kind string, p map[string]json.RawMessage, o *nativeObservation) (bool, error) {
	var fields struct {
		Type      string `json:"type"`
		ID        string `json:"id"`
		CWD       string `json:"cwd"`
		Model     string `json:"model"`
		Name      string `json:"name"`
		CallID    string `json:"call_id"`
		Role      string `json:"role"`
		Author    string `json:"author"`
		Recipient string `json:"recipient"`
		TurnID    string `json:"turn_id"`
	}
	data, err := json.Marshal(p)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return false, err
	}
	stringField := func(key string) (bool, error) {
		raw, ok := p[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false, nil
		}
		_, err := sessionString(p, key)
		return err == nil, err
	}
	outputField := func() (bool, error) {
		if len(p["output"]) == 0 {
			return false, nil
		}
		var output any
		if err := json.Unmarshal(p["output"], &output); err != nil {
			return false, err
		}
		switch output.(type) {
		case string, []any:
			return true, nil
		}
		return false, sessionFailure("output_type_invalid")
	}
	switch kind {
	case "session_meta":
		o.kind = "native_session_header"
		return fields.ID != "" && fields.CWD != "", nil
	case "turn_context":
		o.kind = "turn_context"
		policy, err := sessionObject(p["sandbox_policy"])
		if err != nil {
			return false, err
		}
		return fields.CWD != "" && fields.Model != "" && len(p["approval_policy"]) > 0 && policy != nil, nil
	case "response_item":
		switch fields.Type {
		case "function_call":
			o.kind = sessionToolKind(fields.Name)
			valid, err := stringField("arguments")
			return valid && fields.Name != "" && fields.CallID != "", err
		case "function_call_output":
			o.kind = "function_call_output"
			return outputField()
		case "custom_tool_call":
			o.kind = "custom_tool_call"
			valid, err := stringField("input")
			return valid && fields.Name != "" && fields.CallID != "", err
		case "custom_tool_call_output":
			o.kind = "custom_tool_call_output"
			valid, err := outputField()
			return valid && fields.CallID != "", err
		case "message", "agent_message":
			o.kind = fields.Type
			valid, err := sessionCodexContent(p["content"])
			if fields.Type == "message" {
				return valid && fields.Role != "", err
			}
			return valid && fields.Author != "" && fields.Recipient != "", err
		case "reasoning":
			o.kind = "reasoning"
			return sessionJSONArray(p["summary"])
		}
	case "event_msg":
		switch fields.Type {
		case "token_count":
			o.kind = "native_activity"
			info, err := sessionObject(p["info"])
			if err != nil {
				return false, err
			}
			if info == nil {
				return true, nil
			}
			usage, err := sessionObject(info["last_token_usage"])
			if err != nil {
				return false, err
			}
			if usage == nil {
				return true, nil
			}
			var valid bool
			o.tokens, valid, err = sessionTokenTotal(usage)
			return valid, err
		case "agent_message", "user_message":
			o.kind = fields.Type
			return stringField("message")
		case "agent_reasoning", "agent_reasoning_raw_content":
			o.kind = fields.Type
			return stringField("text")
		case "task_started", "turn_started", "task_complete", "turn_complete":
			o.kind = fields.Type
			return fields.TurnID != "", nil
		case "context_compacted":
			o.kind = "context_compacted"
			return true, nil
		case "error", "warning":
			o.kind = fields.Type
			return stringField("message")
		}
	}
	return false, nil
}

func sessionToolKind(name string) string {
	switch name {
	case "Edit", "Write", "MultiEdit", "apply_patch":
		return "edit_file"
	case "Grep", "Glob", "WebSearch", "WebFetch", "web.run":
		return "search"
	case "ReadPaper":
		return "read_paper"
	}
	return "native_activity"
}

func sessionJSONArray(raw json.RawMessage) (bool, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return false, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return false, err
	}
	return true, nil
}

func sessionCodexContent(raw json.RawMessage) (bool, error) {
	valid, err := sessionJSONArray(raw)
	if err != nil || !valid {
		return false, err
	}
	var items []struct {
		Type string  `json:"type"`
		Text *string `json:"text"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return false, err
	}
	for _, item := range items {
		switch item.Type {
		case "input_text", "output_text":
			if item.Text == nil {
				return false, sessionFailure("content_text_missing")
			}
		default:
			return false, nil
		}
	}
	return true, nil
}
