package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
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

type nativeEnvelope struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	CWD       string          `json:"cwd"`
	Timestamp json.RawMessage `json:"timestamp"`
	Message   *nativeMessage  `json:"message"`
	Payload   *nativePayload  `json:"payload"`
}

// 可选字段缺失保持未知；已提供字段的格式错误由 encoding/json 立即返回。
func validateNativeFormat(raw json.RawMessage) error {
	var record nativeEnvelope
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
	_, err := nativeTimestamp(raw)
	return err
}

func nativeTimestamp(raw json.RawMessage) (int64, error) {
	if absentNativeValue(raw) {
		return 0, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return 0, fmt.Errorf("native timestamp: %w", err)
	}
	switch v := value.(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("native timestamp: %w", err)
		}
		return n, nil
	case string:
		timestamp, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return 0, fmt.Errorf("native timestamp: %w", err)
		}
		return timestamp.UnixMilli(), nil
	default:
		return 0, fmt.Errorf("native timestamp requires integer milliseconds or RFC3339 string")
	}
}

func nativeContent(raw json.RawMessage) (string, []nativeContentBlock, error) {
	if absentNativeValue(raw) {
		return "", nil, nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", nil, err
	}
	if text, ok := value.(string); ok {
		return text, nil, nil
	}
	var blocks []nativeContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return "", nil, err
	}
	var texts []string
	for _, block := range blocks {
		if block.Type == "text" || block.Type == "output_text" || block.Type == "input_text" {
			texts = append(texts, block.Text)
		}
	}
	return strings.Join(texts, "\n"), blocks, nil
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
	if usage.Total == nil && usage.Input != nil && usage.Output != nil && *usage.Input > math.MaxInt-*usage.Output {
		return fmt.Errorf("native token usage exceeds integer range")
	}
	return nil
}

func nativeTokenTotal(usage *nativeUsage) *int {
	if usage == nil {
		return nil
	}
	if usage.Total != nil {
		return usage.Total
	}
	if usage.Input == nil || usage.Output == nil {
		return nil
	}
	total := *usage.Input + *usage.Output
	return &total
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
