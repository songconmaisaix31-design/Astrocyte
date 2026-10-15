package distillers

import "testing"

// 仅核实明确 JSONL 格式输入，不生成实际 CLI、模型输出或来源事实。
func TestProcessorEventJSONLRulesFailAtMalformedRecord(t *testing.T) {
	events, err := decodeProcessorEvents([]byte("{\"type\":\"rule_input\",\"thread_id\":\"explicit-rule-value\"}\n{\"type\":\"second_rule_input\"}\n"))
	if err != nil || len(events) != 2 || events[0].Type != "rule_input" || events[0].ThreadID != "explicit-rule-value" || events[1].Type != "second_rule_input" {
		t.Fatalf("explicit JSONL format rules rejected: %+v %v", events, err)
	}
	for _, raw := range []string{
		"{\"type\":\"rule_input\"}\n{invalid}\n{\"type\":\"later_rule_input\"}\n",
		"{\"type\":\"rule_input\"}\n{\"type\":",
		"{\"type\":false}\n",
		"null\n",
		"{}\n",
	} {
		if events, err := decodeProcessorEvents([]byte(raw)); err == nil || events != nil {
			t.Fatal("malformed or ill-typed JSONL published partial events", events, err)
		}
	}
}
