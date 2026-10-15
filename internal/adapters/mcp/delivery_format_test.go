package mcp

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// 以下仅验证格式函数的明确参数，不创建业务收据、来源事件或投递事实。
func TestDeliveryErrorFormatRulesPreserveNumbers(t *testing.T) {
	const number = "9007199254740993"
	input := map[string]any{"number": json.Number(number), "label": "explicit rule input"}
	result, err := deliveryErrorResult(input)
	if err != nil || !result.IsError || len(result.Content) != 1 || result.Content[0].Type != "text" {
		t.Fatal("explicit formatting input failed", result, err)
	}
	if result.StructuredContent["number"] != json.Number(number) || result.StructuredContent["label"] != input["label"] || result.StructuredContent["error"] != "Tool operation failed" {
		t.Fatal("structured formatting changed the explicit input")
	}
	decoder := json.NewDecoder(strings.NewReader(result.Content[0].Text))
	decoder.UseNumber()
	var text map[string]any
	if err := decoder.Decode(&text); err != nil || text["number"] != json.Number(number) || text["label"] != input["label"] || text["error"] != result.StructuredContent["error"] {
		t.Fatal("text formatting changed the explicit input", err)
	}
}

func TestDeliveryErrorFormatRulesRejectInvalidObjects(t *testing.T) {
	for _, input := range []any{nil, []string{"rule input"}, "rule input", map[string]any{"number": math.Inf(1)}} {
		result, err := deliveryErrorResult(input)
		if err == nil || result.StructuredContent != nil || len(result.Content) != 0 {
			t.Fatal("invalid explicit formatting input did not fail immediately", result, err)
		}
	}
}
