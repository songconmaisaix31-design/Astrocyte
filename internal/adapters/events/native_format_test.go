package events

import "testing"

// 字面量定义格式输入，验证错误传播和正常未知状态。
func TestMalformedNativeFieldsFailBeforeMapping(t *testing.T) {
	for _, raw := range []string{
		`null`, `[]`, `{"payload":"broken"}`, `{"cwd":12}`,
		`{"timestamp":"broken"}`, `{"timestamp":1.5}`,
		`{"message":{"content":{}}}`, `{"message":{"content":[{"type":12}]}}`,
		`{"message":{"usage":{"input_tokens":"3"}}}`,
		`{"payload":{"info":{"last_token_usage":{"total_tokens":-1}}}}`,
		`{"type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"broken"}}`,
	} {
		if _, err := MapNativeRecord("codex", "input", "/approved", []byte(raw), 1); err == nil {
			t.Fatalf("malformed input accepted: %s", raw)
		}
	}
	if _, err := SemanticTools("claude_code", []byte(`{"message":{"content":"text only"}}`), 0); err == nil {
		t.Fatal("text-only message acquired a tool block")
	}
	for _, raw := range []string{`{}`, `{"payload":{"type":"unknown"}}`, `{"message":{"content":"unclassified","usage":{"input_tokens":null}}}`} {
		row, err := MapNativeRecord("codex", "input", "/approved", []byte(raw), 1)
		if err != nil || row.Event != nil || row.Observation.Tokens != nil || row.Observation.Outcome != "unknown" {
			t.Fatalf("optional or unclassified input changed: %v", err)
		}
	}
}
