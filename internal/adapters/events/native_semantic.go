package events

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

const SemanticPolicyVersion = "v22.9"

// SemanticTool 仅公开安全的结构字段，分类先于 kind 映射，参数留在适配器内。
type SemanticTool struct {
	Path                 string   `json:"path"`
	Tool                 string   `json:"tool"`
	Bucket               string   `json:"bucket"`
	Reason               string   `json:"reason"`
	Head                 string   `json:"command_head,omitempty"`
	Structure            string   `json:"structure"`
	ShellHeads           []string `json:"shell_heads,omitempty"`
	NonCognitiveSegments int      `json:"noncognitive_segments,omitempty"`
	ParseErrors          []string `json:"parse_errors,omitempty"`
	args                 json.RawMessage
}

func NativeToolName(source string, raw json.RawMessage, block int) (string, error) {
	name, _, err := nativeToolInput(source, raw, block)
	return name, err
}

func nativeToolInput(source string, raw json.RawMessage, block int) (string, json.RawMessage, error) {
	if _, err := nativeRecord(source, raw); err != nil {
		return "", nil, err
	}
	v := object(raw)
	if source == "codex" {
		p := object(v["payload"])
		a := p["arguments"]
		if len(a) == 0 {
			a = p["input"]
		}
		return str(p, "name"), a, nil
	}
	var blocks []map[string]json.RawMessage
	if err := json.Unmarshal(object(v["message"])["content"], &blocks); err != nil {
		return "", nil, fmt.Errorf("native tool content: %w", err)
	}
	if block < 0 || block >= len(blocks) {
		return "", nil, fmt.Errorf("native tool block index outside observed content")
	}
	return str(blocks[block], "name"), blocks[block]["input"], nil
}

func SemanticTools(source string, raw json.RawMessage, block int) ([]SemanticTool, error) {
	name, args, err := nativeToolInput(source, raw, block)
	if err != nil {
		return nil, err
	}
	return semanticTool(name, args, "0", 0), nil
}

func decodedArgument(raw json.RawMessage) json.RawMessage {
	var s string
	if json.Unmarshal(raw, &s) == nil && json.Valid([]byte(s)) {
		return []byte(s)
	}
	return raw
}

func semanticTool(name string, args json.RawMessage, path string, depth int) []SemanticTool {
	name = canonicalTool(name)
	s := SemanticTool{Path: path, Tool: name, Bucket: "unresolved_structure", Reason: "unknown_tool_semantics", Structure: "direct", args: args}
	one := func(bucket, reason string) []SemanticTool {
		s.Bucket, s.Reason = bucket, reason
		return []SemanticTool{s}
	}
	switch name {
	case "exec":
		var code string
		if depth >= 8 || json.Unmarshal(args, &code) != nil {
			return one("unresolved_structure", "wrapper_input_or_depth")
		}
		return semanticWrapper(code, path, depth+1)
	case "Edit", "Write", "MultiEdit", "apply_patch":
		return one("cognitive_edit", "editing_tool")
	case "Grep", "Glob", "WebSearch", "WebFetch", "web.run", "web__run", "search_evidence", "mcp__morph_research__search_evidence":
		return one("cognitive_search", "search_tool")
	case "ReadPaper", "arxiv", "read_paper":
		return one("cognitive_paper", "paper_reader_tool")
	case "Bash", "exec_command", "shell_command":
		v := object(decodedArgument(args))
		cmd := str(v, "cmd")
		if cmd == "" {
			cmd = str(v, "command")
		}
		if cmd == "" {
			return one("unresolved_structure", "missing_literal_command")
		}
		return semanticCommands(name, cmd, path)
	case "Read":
		if paperTarget(args) {
			return one("cognitive_paper", "literal_paper_target")
		}
		return one("noncognitive_read", "ordinary_file_read")
	case "js":
		return one("noncognitive_browser", "browser_js_control")
	case "SendMessage", "TaskCreate", "TaskUpdate", "TaskList", "TaskGet", "TodoWrite", "TodoRead", "request_user_input_async", "request_user_input", "write_stdin", "wait", "sleep", "clock__sleep", "clock__curr_time", "view_image", "apply_candidate", "approve_candidate", "complete_research_task", "discover_tasks", "inherit_experience", "lease_task", "project_context", "research_candidate", "research_experiment", "verify_research":
		return one("noncognitive_control", "task_process_or_callback_control")
	}
	if name == "create_goal" || name == "get_goal" || name == "update_goal" || name == "mcp__codex_app__list_threads" || name == "mcp__codex_app__read_thread" || name == "mcp__codex_apps__gmail_batch_read_email" {
		return one("noncognitive_control", "task_or_message_control")
	}
	if name == "mcp__codex_apps__gmail_search_emails" {
		return one("cognitive_search", "search_tool")
	}
	if strings.HasPrefix(name, "mcp__morph_research__") {
		return semanticTool(strings.TrimPrefix(name, "mcp__morph_research__"), args, path, depth)
	}
	if strings.Contains(name, "arxiv") {
		return one("cognitive_paper", "arxiv_tool")
	}
	return []SemanticTool{s}
}

func paperTarget(args json.RawMessage) bool {
	v := object(decodedArgument(args))
	target := str(v, "file_path")
	if target == "" {
		target = str(v, "path")
	}
	target = strings.ToLower(target)
	return strings.HasSuffix(target, ".pdf") || strings.Contains(target, "arxiv.org/pdf/") || strings.Contains(target, "arxiv.org/abs/")
}

// 每个语义单位保留父事件 bytes 和已观察关联；未知 token 使用量不复制给多个子项。
func MapNativeSemanticTools(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) ([]app.StoredEvent, []SemanticTool, error) {
	return NewNativeSemanticMapper().MapTool(source, origin, project, raw, receivedAt, association, block)
}

func mapNativeSemanticTools(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) ([]app.StoredEvent, []SemanticTool, error) {
	parent, err := MapNativeToolRecord(source, origin, project, raw, receivedAt, association, block)
	if err != nil {
		return nil, nil, err
	}
	name, args, err := nativeToolInput(source, raw, block)
	if err != nil {
		return nil, nil, err
	}
	children := semanticTool(name, args, "0", 0)
	var rows []app.StoredEvent
	for _, s := range children {
		o, err := nativeRecord(source, raw)
		if err != nil {
			return nil, nil, err
		}
		o.session, o.cwd = parent.Observation.Session, parent.Observation.Scope
		o.kind = "unknown"
		if strings.HasPrefix(s.Bucket, "cognitive_") {
			o.kind = toolKind(s.Tool, s.args)
		}
		if len(children) > 1 {
			o.tokens = nil
		}
		rows = append(rows, nativeReceipt(origin+"/semantic/"+s.Path, source, parent.Observation.SourceKind, "native:semantic:"+SemanticPolicyVersion, raw, o, receivedAt))
	}
	return rows, children, nil
}
