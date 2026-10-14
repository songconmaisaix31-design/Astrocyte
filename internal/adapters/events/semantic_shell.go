package events

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

func shellLiteral(word *syntax.Word) (string, bool) {
	var text strings.Builder
	var appendPart func(syntax.WordPart) bool
	appendPart = func(part syntax.WordPart) bool {
		switch value := part.(type) {
		case *syntax.Lit:
			text.WriteString(value.Value)
		case *syntax.SglQuoted:
			text.WriteString(value.Value)
		case *syntax.DblQuoted:
			for _, child := range value.Parts {
				if !appendPart(child) {
					return false
				}
			}
		default:
			return false
		}
		return true
	}
	for _, part := range word.Parts {
		if !appendPart(part) {
			return "", false
		}
	}
	return text.String(), true
}

// Bash 和 here-document 语法由 mvdan AST 区分命令与正文，解析失败立即保留未知状态。
func parseBashCommands(code string) shellASTResult {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(code), "")
	if err != nil {
		return shellASTResult{ParseErrors: 1, ErrorIDs: []string{"bash_syntax_error"}}
	}
	result := shellASTResult{}
	syntax.Walk(file, func(node syntax.Node) bool {
		switch value := node.(type) {
		case *syntax.CallExpr:
			if len(value.Args) > 0 {
				head, literal := shellLiteral(value.Args[0])
				result.Commands = append(result.Commands, shellASTCommand{Head: head, Dynamic: !literal})
			}
		case *syntax.IfClause, *syntax.ForClause, *syntax.WhileClause, *syntax.CaseClause, *syntax.FuncDecl:
			result.Commands = append(result.Commands, shellASTCommand{Dynamic: true})
		}
		return true
	})
	return result
}

func semanticCommands(name, cmd, path string) []SemanticTool {
	var parsed shellASTResult
	// 来源为 Bash 或明确包含 here-document 的输入使用 Bash 语法，其余沿宿主原生格式。
	if name == "Bash" || runtime.GOOS != "windows" || strings.Contains(cmd, "<<") {
		parsed = parseBashCommands(cmd)
	} else {
		var available bool
		parsed, available = parseShellAST(cmd)
		if !available {
			parsed = shellASTResult{ParseErrors: 1, ErrorIDs: []string{"native_shell_parser_unavailable"}}
		}
	}
	s := SemanticTool{Path: path, Tool: name, Bucket: "noncognitive_shell", Reason: "nonsearch_command", Structure: "shell_invocation", ParseErrors: parsed.ErrorIDs}
	if parsed.ParseErrors > 0 {
		s.Bucket, s.Reason = "unresolved_structure", "dynamic_shell_structure"
		s.ShellHeads = []string{"parse_error"}
		return []SemanticTool{s}
	}
	search, unresolved := false, false
	var searchArgs json.RawMessage
	for _, command := range parsed.Commands {
		head := strings.ToLower(strings.TrimSuffix(filepath.Base(strings.ReplaceAll(command.Head, "\\", "/")), ".exe"))
		safeHead := "other"
		if command.Dynamic || head == "invoke-expression" || head == "iex" || head == "eval" {
			safeHead, unresolved = "dynamic", true
		} else if head == "rg" || head == "grep" || head == "find" {
			safeHead, search = head, true
			if searchArgs == nil {
				searchArgs, _ = json.Marshal(map[string]string{"cmd": head})
			}
		} else {
			s.NonCognitiveSegments++
			switch head {
			case "git", "ls", "tree", "dir", "get-childitem", "get-content", "cat", "pwd", "cd", "set-location", "echo", "write-output", "go", "python", "python3", "npm", "pnpm", "pip", "pytest", "gofmt", "cargo", "node", "cmake", "make", "curl", "sleep", "timeout", "powershell", "pwsh", "cmd":
				safeHead = head
			}
		}
		s.ShellHeads = append(s.ShellHeads, safeHead)
	}
	if len(parsed.Commands) == 0 {
		s.ShellHeads, s.NonCognitiveSegments = []string{"expression"}, 1
	}
	if search {
		s.Bucket, s.Reason = "cognitive_search", "one_search_invocation"
	}
	if unresolved {
		s.Bucket, s.Reason = "unresolved_structure", "dynamic_shell_structure"
	}
	s.args, _ = json.Marshal(map[string]string{"cmd": cmd})
	if search && !unresolved {
		s.args = searchArgs
	}
	return []SemanticTool{s}
}
