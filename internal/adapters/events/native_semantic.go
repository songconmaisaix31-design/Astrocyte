package events

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

const SemanticPolicyVersion = "v22.7"

// SemanticTool contains safe structural fields only. Arguments never leave this
// adapter. Category is decided from tool semantics before kind mapping.
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

func NativeToolName(source string, raw json.RawMessage, block int) string {
	name, _ := nativeToolInput(source, raw, block)
	return name
}

func nativeToolInput(source string, raw json.RawMessage, block int) (string, json.RawMessage) {
	v := object(raw)
	if source == "codex" {
		p := object(v["payload"])
		a := p["arguments"]
		if len(a) == 0 {
			a = p["input"]
		}
		return str(p, "name"), a
	}
	var blocks []map[string]json.RawMessage
	_ = json.Unmarshal(object(v["message"])["content"], &blocks)
	if block < 0 || block >= len(blocks) {
		return "", nil
	}
	return str(blocks[block], "name"), blocks[block]["input"]
}

func SemanticTools(source string, raw json.RawMessage, block int) []SemanticTool {
	name, args := nativeToolInput(source, raw, block)
	return semanticTool(name, args, "0", 0)
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
	// MCP tool names retain explicit search/paper semantics, not arbitrary text.
	if strings.Contains(name, "arxiv") {
		return one("cognitive_paper", "arxiv_tool")
	}
	return []SemanticTool{s}
}

// Shell scanning is purely lexical: quoted payloads cannot invent child commands.
func semanticCommandSegments(name, cmd, path string) []SemanticTool {
	var valid bool
	cmd, valid = stripShellData(cmd)
	if !valid {
		return []SemanticTool{{Path: path, Tool: name, Bucket: "unresolved_structure", Reason: "unclosed_shell_data", Structure: "shell"}}
	}
	var parts []string
	start, quote := 0, byte(0)
	escaped := false
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' || c == '`' {
			escaped = true
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if strings.ContainsRune(";|&\n\r", rune(c)) {
			parts = append(parts, cmd[start:i])
			start = i + 1
		}
	}
	if quote != 0 {
		return []SemanticTool{{Path: path, Tool: name, Bucket: "unresolved_structure", Reason: "unclosed_shell_quote", Structure: "shell"}}
	}
	parts = append(parts, cmd[start:])
	var result []SemanticTool
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		head := strings.ToLower(strings.Trim(fields[0], "\"'"))
		head = strings.TrimSuffix(filepath.Base(strings.ReplaceAll(head, "\\", "/")), ".exe")
		s := SemanticTool{Path: fmt.Sprintf("%s.%d", path, len(result)), Tool: name, Bucket: "noncognitive_shell", Reason: "nonsearch_command", Structure: "shell_segment"}
		// Heads are an allowlisted vocabulary; never expose variable/path/script text.
		switch head {
		case "rg", "grep", "find":
			s.Head = head
			s.Bucket = "cognitive_search"
			s.Reason = "explicit_search_command"
		case "git", "ls", "tree", "dir", "get-childitem", "get-content", "cat", "pwd", "cd", "set-location", "echo", "write-output", "go", "python", "python3", "npm", "pnpm", "pip", "pytest", "gofmt", "cargo", "node", "cmake", "make", "curl", "sleep", "timeout", "powershell", "pwsh", "cmd":
			s.Head = head
		default:
			s.Head = "other"
		}
		assignment := regexp.MustCompile(`^\$?(?:env:)?[A-Za-z_][A-Za-z0-9_:]*\s*=`).MatchString(strings.TrimSpace(part)) && !strings.Contains(part, "$(")
		if assignment {
			s.Head = "assignment"
			s.Reason = "literal_shell_setup"
		}
		if !assignment && (strings.HasPrefix(fields[0], "$") || strings.ContainsAny(fields[0], "({}") || head == "if" || head == "for" || head == "foreach" || head == "while" || head == "eval" || head == "iex" || head == "invoke-expression") {
			s.Bucket = "unresolved_structure"
			s.Reason = "dynamic_shell_structure"
		}
		// The argument contains only this isolated literal command, never parent code.
		s.args, _ = json.Marshal(map[string]string{"cmd": strings.TrimSpace(part)})
		result = append(result, s)
	}
	if len(result) == 0 {
		return []SemanticTool{{Path: path, Tool: name, Bucket: "noncognitive_shell", Reason: "empty_command", Structure: "shell"}}
	}
	return result
}

// Here-document and here-string bodies are literal data consumed by the outer
// command, not additional shell invocations. Replace data without executing it.
func stripShellData(cmd string) (string, bool) {
	var result strings.Builder
	for i := 0; i < len(cmd); {
		if i+2 < len(cmd) && cmd[i] == '@' && (cmd[i+1] == '\'' || cmd[i+1] == '"') && (cmd[i+2] == '\n' || cmd[i+2] == '\r') {
			close := "\n" + string(cmd[i+1]) + "@"
			end := strings.Index(cmd[i+2:], close)
			if end < 0 {
				return "", false
			}
			result.WriteString("''")
			i += 2 + end + len(close)
			continue
		}
		if strings.HasPrefix(cmd[i:], "<<") {
			marker := regexp.MustCompile(`^<<-?\s*['"]?([A-Za-z_][A-Za-z0-9_]*)['"]?`).FindStringSubmatch(cmd[i:])
			if len(marker) > 1 {
				lineEnd := strings.IndexByte(cmd[i:], '\n')
				if lineEnd < 0 {
					return "", false
				}
				body := i + lineEnd + 1
				delimiter := marker[1]
				end := strings.Index("\n"+cmd[body:], "\n"+delimiter)
				if end < 0 {
					return "", false
				}
				finish := body + end + len(delimiter)
				if finish < len(cmd) && cmd[finish] != '\r' && cmd[finish] != '\n' {
					return "", false
				}
				result.WriteString("''")
				result.WriteString(cmd[i+len(marker[0]) : i+lineEnd])
				i = finish
				continue
			}
		}
		// Copy regular quoted strings whole so embedded << or @' is just data.
		if cmd[i] == '\'' || cmd[i] == '"' {
			quote := cmd[i]
			start := i
			i++
			for i < len(cmd) {
				if cmd[i] == '\\' || cmd[i] == '`' {
					i += 2
					continue
				}
				if cmd[i] == quote {
					i++
					break
				}
				i++
			}
			if i > len(cmd) {
				return "", false
			}
			result.WriteString(cmd[start:i])
			continue
		}
		result.WriteByte(cmd[i])
		i++
	}
	return result.String(), true
}

func semanticCommands(name, cmd, path string) []SemanticTool {
	segments := semanticCommandSegments(name, cmd, path)
	var parseErrors []string
	// Windows shell input uses the host's existing syntax parser. Quoted JS,
	// strings/here-data, assignments and result scriptblocks cannot invent calls.
	if clean, ok := stripShellData(cmd); ok {
		if ast, available := parseShellAST(clean); available {
			if ast.ParseErrors > 0 && clean != cmd {
				if original, _ := parseShellAST(cmd); original.ParseErrors == 0 {
					ast = original
				}
			}
			segments = nil
			if ast.ParseErrors > 0 {
				parseErrors = ast.ErrorIDs
				segments = append(segments, SemanticTool{Bucket: "unresolved_structure", Head: "parse_error"})
			}
			for _, command := range ast.Commands {
				head := strings.ToLower(strings.TrimSuffix(filepath.Base(strings.ReplaceAll(command.Head, "\\", "/")), ".exe"))
				part := SemanticTool{Bucket: "noncognitive_shell", Head: "other"}
				if command.Dynamic {
					part.Bucket = "unresolved_structure"
					part.Head = "dynamic"
				} else if head == "rg" || head == "grep" || head == "find" {
					part.Bucket = "cognitive_search"
					part.Head = head
					part.args, _ = json.Marshal(map[string]string{"cmd": head})
				} else {
					switch head {
					case "git", "ls", "tree", "dir", "get-childitem", "get-content", "cat", "pwd", "cd", "set-location", "echo", "write-output", "go", "python", "python3", "npm", "pnpm", "pip", "pytest", "gofmt", "cargo", "node", "cmake", "make", "curl", "sleep", "timeout", "powershell", "pwsh", "cmd":
						part.Head = head
					}
					if head == "invoke-expression" || head == "iex" || head == "eval" {
						part.Bucket = "unresolved_structure"
						part.Head = "dynamic"
					}
				}
				segments = append(segments, part)
			}
			if len(segments) == 0 {
				segments = []SemanticTool{{Bucket: "noncognitive_shell", Head: "expression"}}
			}
		}
	}
	s := SemanticTool{Path: path, Tool: name, Bucket: "noncognitive_shell", Reason: "nonsearch_command", Structure: "shell_invocation"}
	s.ParseErrors = parseErrors
	search, unresolved := false, false
	var searchArgs json.RawMessage
	for _, part := range segments {
		s.ShellHeads = append(s.ShellHeads, part.Head)
		if part.Bucket == "cognitive_search" {
			search = true
			if searchArgs == nil {
				searchArgs = part.args
			}
		}
		if part.Bucket == "unresolved_structure" {
			unresolved = true
		}
		if strings.HasPrefix(part.Bucket, "noncognitive_") {
			s.NonCognitiveSegments++
		}
	}
	if search {
		s.Bucket = "cognitive_search"
		s.Reason = "one_search_invocation"
	}
	if unresolved {
		s.Bucket = "unresolved_structure"
		s.Reason = "dynamic_shell_structure"
	}
	s.args, _ = json.Marshal(map[string]string{"cmd": cmd})
	if search && !unresolved {
		s.args = searchArgs
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

type semanticToken struct {
	value            string
	literal, dynamic bool
}

func semanticTokens(code string) ([]semanticToken, bool) {
	var result []semanticToken
	for i := 0; i < len(code); {
		c := code[i]
		if unicode.IsSpace(rune(c)) {
			i++
			continue
		}
		if strings.HasPrefix(code[i:], "//") {
			for i < len(code) && code[i] != '\n' {
				i++
			}
			continue
		}
		if strings.HasPrefix(code[i:], "/*") {
			end := strings.Index(code[i+2:], "*/")
			if end < 0 {
				return result, false
			}
			i += end + 4
			continue
		}
		// Regex literals used in output processing may contain quotes or fake
		// tools calls. Their contents are never scanned as executable tokens.
		if c == '/' && len(result) > 0 && !result[len(result)-1].literal && strings.Contains("(,=:[!", result[len(result)-1].value) {
			start := i
			i++
			inClass, closed := false, false
			for i < len(code) {
				if code[i] == '\\' {
					i += 2
					continue
				}
				if code[i] == '[' {
					inClass = true
				}
				if code[i] == ']' {
					inClass = false
				}
				if code[i] == '/' && !inClass {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return result, false
			}
			for i < len(code) && unicode.IsLetter(rune(code[i])) {
				i++
			}
			result = append(result, semanticToken{value: code[start:i], dynamic: true})
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			quote := c
			i++
			var b strings.Builder
			dynamic := false
			closed := false
			for i < len(code) {
				if code[i] == quote {
					closed = true
					i++
					break
				}
				if quote == '`' && strings.HasPrefix(code[i:], "${") {
					dynamic = true
					end, ok := skipJSInterpolation(code, i+2)
					if !ok {
						return result, false
					}
					b.WriteString(code[i:end])
					i = end
					continue
				}
				if code[i] == '\\' {
					v, _, tail, e := strconv.UnquoteChar(code[i:], quote)
					if e != nil {
						if i+1 >= len(code) {
							return result, false
						}
						// JS identity escapes and escaped newlines are literal; no
						// interpolation or historical evaluation is involved.
						if code[i+1] != '\n' && code[i+1] != '\r' {
							b.WriteByte(code[i+1])
						}
						i += 2
						continue
					}
					b.WriteRune(v)
					i = len(code) - len(tail)
				} else {
					b.WriteByte(code[i])
					i++
				}
			}
			if !closed {
				return result, false
			}
			result = append(result, semanticToken{value: b.String(), literal: true, dynamic: dynamic})
			continue
		}
		if unicode.IsDigit(rune(c)) || c == '-' && i+1 < len(code) && unicode.IsDigit(rune(code[i+1])) {
			start := i
			i++
			for i < len(code) && (unicode.IsDigit(rune(code[i])) || strings.ContainsRune(".eE+-", rune(code[i]))) {
				i++
			}
			result = append(result, semanticToken{value: code[start:i]})
			continue
		}
		if unicode.IsLetter(rune(c)) || c == '_' || c == '$' {
			start := i
			i++
			for i < len(code) && (unicode.IsLetter(rune(code[i])) || unicode.IsDigit(rune(code[i])) || code[i] == '_' || code[i] == '$') {
				i++
			}
			result = append(result, semanticToken{value: code[start:i]})
			continue
		}
		result = append(result, semanticToken{value: string(c)})
		i++
	}
	return result, true
}

func skipJSInterpolation(code string, start int) (int, bool) {
	depth := 1
	for i := start; i < len(code); i++ {
		if code[i] == '\'' || code[i] == '"' || code[i] == '`' {
			quote := code[i]
			i++
			for i < len(code) {
				if code[i] == '\\' {
					i += 2
					continue
				}
				if quote == '`' && strings.HasPrefix(code[i:], "${") {
					end, ok := skipJSInterpolation(code, i+2)
					if !ok {
						return 0, false
					}
					i = end
					continue
				}
				if code[i] == quote {
					break
				}
				i++
			}
			if i >= len(code) {
				return 0, false
			}
			continue
		}
		if code[i] == '{' {
			depth++
		}
		if code[i] == '}' {
			depth--
			if depth == 0 {
				return i + 1, true
			}
		}
	}
	return 0, false
}

func semanticLiteral(tokens []semanticToken, start int) (any, int, bool) {
	// Reuse the canonical literal grammar, with dynamic templates explicitly barred.
	var t []wrapperToken
	for _, v := range tokens[start:] {
		if v.dynamic {
			t = append(t, wrapperToken{value: "__dynamic_literal__"})
		} else {
			t = append(t, wrapperToken{value: v.value, literal: v.literal})
		}
	}
	i := 0
	v, ok := literalArgument(t, &i)
	return v, start + i, ok
}

func semanticWrapper(code, path string, depth int) []SemanticTool {
	tokens, valid := semanticTokens(code)
	if !valid {
		return []SemanticTool{{Path: path, Tool: "exec", Bucket: "unresolved_structure", Reason: "unsupported_js_structure", Structure: "wrapper"}}
	}
	bindings := staticBindings(tokens)
	var result []SemanticTool
	unresolved := func(p, reason string) {
		result = append(result, SemanticTool{Path: p, Tool: "exec", Bucket: "unresolved_structure", Reason: reason, Structure: "wrapper"})
	}
	// A tools rebinding invalidates identity; do not guess through aliases.
	for i, t := range tokens {
		if !t.literal && t.value == "tools" && i+1 < len(tokens) && tokens[i+1].value == "=" {
			unresolved(path, "tools_rebound")
			return result
		}
	}
	// Braced control/function bodies are recorded rather than executed. Ordinary
	// result callbacks are harmless unless a tool invocation is inside them.
	var contexts []bool
	pendingUnsafe := false
	ordinal := 0
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t.literal {
			continue
		}
		switch t.value {
		case "if", "for", "foreach", "while", "switch", "function":
			pendingUnsafe = true
		case "=":
			if i+1 < len(tokens) && tokens[i+1].value == ">" {
				pendingUnsafe = true
			}
		case "{":
			contexts = append(contexts, pendingUnsafe)
			pendingUnsafe = false
		case "}":
			if len(contexts) > 0 {
				contexts = contexts[:len(contexts)-1]
			}
			pendingUnsafe = false
		case ";":
			pendingUnsafe = false
		}
		if t.value != "tools" || i+1 >= len(tokens) {
			continue
		}
		if tokens[i+1].value == "[" {
			unresolved(fmt.Sprintf("%s.%d", path, ordinal), "computed_tool_name")
			ordinal++
			continue
		}
		if i+3 >= len(tokens) || tokens[i+1].value != "." || tokens[i+3].value != "(" || tokens[i+2].literal {
			continue
		}
		p := fmt.Sprintf("%s.%d", path, ordinal)
		ordinal++
		if i > 0 && tokens[i-1].value == "." {
			unresolved(p, "untrusted_tools_receiver")
			continue
		}
		unsafe := pendingUnsafe
		for _, v := range contexts {
			unsafe = unsafe || v
		}
		// Conditional operators before this call in the same statement are unsafe.
		for j := i - 1; j >= 0 && tokens[j].value != ";" && tokens[j].value != "{" && tokens[j].value != "}"; j-- {
			if !tokens[j].literal && (tokens[j].value == "?" || tokens[j].value == "&" || tokens[j].value == "|") {
				unsafe = true
			}
		}
		argument, end, ok := staticArgument(tokens, i+4, bindings)
		if !ok || end >= len(tokens) || tokens[end].value != ")" {
			unresolved(p, "dynamic_arguments")
			result[len(result)-1].Tool = tokens[i+2].value
			continue
		}
		if unsafe {
			unresolved(p, "conditional_or_function_call")
			result[len(result)-1].Tool = tokens[i+2].value
			i = end
			continue
		}
		a, _ := json.Marshal(argument)
		children := semanticTool(tokens[i+2].value, a, p, depth)
		for j := range children {
			children[j].Structure = "wrapper/" + children[j].Structure
		}
		result = append(result, children...)
		i = end
	}
	if len(result) == 0 {
		// No tool invocation is control only when the complete literal program was
		// understood lexically and has no dynamic eval or runtime-dispatched tooling.
		unsafe := false
		for _, t := range tokens {
			if !t.literal && (t.value == "eval" || t.value == "Function" || t.value == "tools") {
				unsafe = true
			}
		}
		if unsafe {
			unresolved(path, "opaque_dispatch")
		} else {
			result = append(result, SemanticTool{Path: path, Tool: "exec", Bucket: "noncognitive_control", Reason: "no_tool_invocation", Structure: "wrapper"})
		}
	}
	return result
}

// MapNativeSemanticTools is the production child mapping boundary. Every child
// retains the immutable parent's raw and observed association. Unknown usage
// is not duplicated across siblings. Unresolved/noncognitive rows remain local
// observations and do not fabricate Events.
func MapNativeSemanticTools(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) ([]app.StoredEvent, []SemanticTool, error) {
	return NewNativeSemanticMapper().MapTool(source, origin, project, raw, receivedAt, association, block)
}

func mapNativeSemanticTools(source, origin, project string, raw json.RawMessage, receivedAt int64, association NativeAssociation, block int) ([]app.StoredEvent, []SemanticTool, error) {
	parent, err := MapNativeToolRecord(source, origin, project, raw, receivedAt, association, block)
	if err != nil {
		return nil, nil, err
	}
	children := SemanticTools(source, raw, block)
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
