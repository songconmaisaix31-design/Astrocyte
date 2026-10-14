package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type nativeUsage struct {
	Total  *int `json:"total_tokens"`
	Input  *int `json:"input_tokens"`
	Output *int `json:"output_tokens"`
}

type nativeContentBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Text  string          `json:"text"`
	Input json.RawMessage `json:"input"`
}

type nativeMessage struct {
	Content json.RawMessage `json:"content"`
	Usage   *nativeUsage    `json:"usage"`
}

type nativePayload struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	CWD       string          `json:"cwd"`
	Name      string          `json:"name"`
	Role      string          `json:"role"`
	Message   string          `json:"message"`
	Timestamp json.RawMessage `json:"timestamp"`
	Content   json.RawMessage `json:"content"`
	Arguments json.RawMessage `json:"arguments"`
	Input     json.RawMessage `json:"input"`
	Info      *struct {
		Usage *nativeUsage `json:"last_token_usage"`
	} `json:"info"`
}

// 可选字段缺失保持未知；已提供字段的格式错误由 encoding/json 立即返回。
func validateNativeFormat(raw json.RawMessage) error {
	var record struct {
		Type      string          `json:"type"`
		SessionID string          `json:"sessionId"`
		CWD       string          `json:"cwd"`
		Timestamp json.RawMessage `json:"timestamp"`
		Message   *nativeMessage  `json:"message"`
		Payload   *nativePayload  `json:"payload"`
	}
	if err := json.Unmarshal(raw, &record); err != nil {
		return fmt.Errorf("native record format: %w", err)
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("native record requires a JSON object")
	}
	if err := validateNativeTimestamp(record.Timestamp); err != nil {
		return err
	}
	if record.Message != nil {
		if err := validateNativeContent(record.Message.Content); err != nil {
			return err
		}
		if err := validateNativeUsage(record.Message.Usage); err != nil {
			return err
		}
	}
	if record.Payload != nil {
		p := record.Payload
		if err := validateNativeTimestamp(p.Timestamp); err != nil {
			return err
		}
		if err := validateNativeContent(p.Content); err != nil {
			return err
		}
		if p.Info != nil {
			if err := validateNativeUsage(p.Info.Usage); err != nil {
				return err
			}
		}
	}
	return nil
}

func absentNativeValue(raw json.RawMessage) bool {
	return len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func validateNativeTimestamp(raw json.RawMessage) error {
	if absentNativeValue(raw) {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("native timestamp: %w", err)
	}
	switch v := value.(type) {
	case json.Number:
		if _, err := v.Int64(); err != nil {
			return fmt.Errorf("native timestamp: %w", err)
		}
	case string:
		if _, err := time.Parse(time.RFC3339Nano, v); err != nil {
			return fmt.Errorf("native timestamp: %w", err)
		}
	default:
		return fmt.Errorf("native timestamp requires integer milliseconds or RFC3339 string")
	}
	return nil
}

func validateNativeContent(raw json.RawMessage) error {
	if absentNativeValue(raw) {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("native content: %w", err)
	}
	if _, ok := value.(string); ok {
		return nil
	}
	var blocks []nativeContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return fmt.Errorf("native content blocks: %w", err)
	}
	for _, block := range blocks {
		if canonicalTool(block.Name) == "Bash" || canonicalTool(block.Name) == "Read" {
			if err := validateToolObject(block.Input); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateNativeUsage(usage *nativeUsage) error {
	if usage == nil {
		return nil
	}
	for _, value := range []*int{usage.Total, usage.Input, usage.Output} {
		if value != nil && *value < 0 {
			return fmt.Errorf("native token usage must be nonnegative")
		}
	}
	return nil
}

func validateToolObject(raw json.RawMessage) error {
	if absentNativeValue(raw) {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("native tool arguments: %w", err)
	}
	if text, ok := value.(string); ok {
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			return fmt.Errorf("native encoded tool arguments: %w", err)
		}
	}
	if _, ok := value.(map[string]any); !ok {
		return fmt.Errorf("native tool arguments require a JSON object")
	}
	return nil
}
