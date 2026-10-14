package events

import (
	"encoding/json"
	"strings"

	"github.com/dop251/goja/ast"
)

func wrapperKind(code string) string {
	kind, _ := wrapperAnalysis(code)
	return kind
}

// 旧父事件映射保留单一 kind 和直接 await 限界，语法由同一成熟 AST 提供。
func wrapperAnalysis(code string) (string, string) {
	body, err := parseWrapper(code)
	if err != nil {
		return "unknown", "dynamic_or_unsupported_literal"
	}
	valid, kind := true, ""
	var visit func(ast.Node, ast.Node)
	visit = func(node, parent ast.Node) {
		if runtimeControl(node) {
			valid = false
		}
		if _, loop := node.(*ast.ForOfStatement); loop {
			valid = false
		}
		if binding, ok := node.(*ast.Binding); ok && rootIdentifier(binding.Target) == "tools" {
			valid = false
		}
		if call, ok := node.(*ast.CallExpression); ok {
			name, direct, computed := toolCallName(call)
			if direct || computed {
				await, awaited := parent.(*ast.AwaitExpression)
				if !direct || !awaited || await.Argument != call || len(call.ArgumentList) != 1 || name == "exec" {
					valid = false
					return
				}
				value, ok := staticValue(call.ArgumentList[0], nil, 0)
				if !ok {
					valid = false
					return
				}
				encoded, err := json.Marshal(value)
				if err != nil {
					valid = false
					return
				}
				if name == "exec_command" || name == "Bash" || name == "shell_command" {
					argument := object(encoded)
					command := str(argument, "cmd")
					if command == "" {
						command = str(argument, "command")
					}
					if strings.ContainsAny(command, ";|&\n\r") {
						valid = false
					}
				}
				mapped := toolKind(name, encoded)
				if mapped == "unknown" || kind != "" && kind != mapped {
					valid = false
				}
				kind = mapped
				return
			}
		}
		for _, child := range syntaxChildren(node) {
			visit(child, node)
		}
	}
	visit(body, nil)
	if mutatedNames(body)["tools"] || !valid || kind == "" {
		return "unknown", "unsupported_literal_or_control"
	}
	return kind, "recognized"
}
