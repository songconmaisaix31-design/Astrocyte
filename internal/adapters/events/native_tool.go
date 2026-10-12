package events

import (
	"encoding/json"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func NativeToolBlocks(source string, raw json.RawMessage) []int {
	v := object(raw)
	if source == "codex" {
		p := object(v["payload"])
		if str(v, "type") == "response_item" && (str(p, "type") == "function_call" || str(p, "type") == "custom_tool_call") {
			return []int{0}
		}
		return nil
	}
	if source != "claude_code" || str(v, "type") != "assistant" {
		return nil
	}
	var blocks []map[string]json.RawMessage
	_ = json.Unmarshal(object(v["message"])["content"], &blocks)
	var result []int
	for i, b := range blocks {
		if str(b, "type") == "tool_use" {
			result = append(result, i)
		}
	}
	return result
}

// MapNativeToolRecord maps the preselected structural unit using the same
// canonical production rules. A Claude block is not a whole message; raw and
// the caller's immutable origin stay unchanged. Aggregated message usage is
// unknown per block when multiple tool blocks share that usage.
func MapNativeToolRecord(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) (app.StoredEvent, error) {
	row, err := MapAssociatedNativeRecord(source, origin, project, raw, receivedAt, association)
	if err != nil {
		return row, err
	}
	v := object(raw)
	if source == "codex" {
		p := object(v["payload"])
		if block != 0 || str(v, "type") != "response_item" || (str(p, "type") != "function_call" && str(p, "type") != "custom_tool_call") {
			return app.StoredEvent{}, fmt.Errorf("not a structural tool record")
		}
		return row, nil
	}
	if source != "claude_code" || str(v, "type") != "assistant" {
		return app.StoredEvent{}, fmt.Errorf("not a structural tool block")
	}
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(object(v["message"])["content"], &blocks) != nil || block < 0 || block >= len(blocks) || str(blocks[block], "type") != "tool_use" {
		return app.StoredEvent{}, fmt.Errorf("not a structural tool block")
	}
	o, err := nativeRecord(source, raw)
	if err != nil {
		return row, err
	}
	o.kind = toolKind(str(blocks[block], "name"), blocks[block]["input"])
	o.session, o.cwd = row.Observation.Session, row.Observation.Scope
	count := 0
	for _, b := range blocks {
		if str(b, "type") == "tool_use" {
			count++
		}
	}
	if count > 1 {
		o.tokens = nil
	}
	return nativeReceipt(origin, source, "assistant", "native:selected_tool_block", raw, o, receivedAt), nil
}

// NativeToolForm exposes only closed vocabulary, never a private tool name or
// arguments. It separates opaque wrappers from unsupported direct calls.
func NativeToolForm(source string, raw json.RawMessage, block int) string {
	v := object(raw)
	name := ""
	if source == "codex" {
		name = str(object(v["payload"]), "name")
	} else {
		var blocks []map[string]json.RawMessage
		_ = json.Unmarshal(object(v["message"])["content"], &blocks)
		if block >= 0 && block < len(blocks) {
			name = str(blocks[block], "name")
		}
	}
	name = canonicalTool(name)
	for _, rule := range nativeKindRules {
		if rule.value == name && rule.selector != "assistant_marker" && rule.selector != "git_record" {
			return rule.selector + ":" + rule.value
		}
	}
	return "unsupported_direct_tool"
}

func NativeToolUnknownReason(source string, raw json.RawMessage, block int) string {
	form := NativeToolForm(source, raw, block)
	if form == "wrapper:exec" {
		p := object(object(raw)["payload"])
		argument := p["input"]
		if len(argument) == 0 {
			argument = p["arguments"]
		}
		var code string
		if json.Unmarshal(argument, &code) != nil {
			return "wrapper_invalid_input"
		}
		_, reason := wrapperAnalysis(code)
		return "wrapper_" + reason
	}
	if form == "shell_search:Bash" || form == "shell_search:exec_command" || form == "shell_search:shell_command" {
		return "command_not_single_recognized_search"
	}
	return "unsupported_direct_tool"
}
