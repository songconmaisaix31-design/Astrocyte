package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
	"github.com/dop251/goja/token"
)

// Goja 仅构建 AST；禁用 source maps，防止解析输入引发文件读取。
func parseWrapper(code string) (*ast.BlockStatement, error) {
	program, err := parser.ParseFile(nil, "", "async function __astrocyte__(){\n"+code+"\n}", 0, parser.WithDisableSourceMaps)
	if err != nil {
		return nil, err
	}
	if len(program.Body) != 1 {
		return nil, fmt.Errorf("wrapper must contain one function body")
	}
	function, ok := program.Body[0].(*ast.FunctionDeclaration)
	if !ok {
		return nil, fmt.Errorf("wrapper function body unavailable")
	}
	return function.Function.Body, nil
}

// 只遍历成熟 AST 的节点字段，跳过重复的声明索引和来源附件。
func syntaxChildren(node ast.Node) []ast.Node {
	value := reflect.ValueOf(node)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil
	}
	var result []ast.Node
	var appendNode func(reflect.Value)
	appendNode = func(field reflect.Value) {
		if !field.IsValid() || !field.CanInterface() {
			return
		}
		if (field.Kind() == reflect.Pointer || field.Kind() == reflect.Interface) && field.IsNil() {
			return
		}
		if child, ok := field.Interface().(ast.Node); ok {
			result = append(result, child)
			return
		}
		if field.Kind() == reflect.Struct && field.CanAddr() {
			if child, ok := field.Addr().Interface().(ast.Node); ok {
				result = append(result, child)
				return
			}
		}
		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				appendNode(field.Index(i))
			}
		}
	}
	for i := 0; i < value.NumField(); i++ {
		name := value.Type().Field(i).Name
		if name != "DeclarationList" && name != "File" && name != "Source" {
			appendNode(value.Field(i))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Idx0() < result[j].Idx0() })
	return result
}

func walkSyntax(node ast.Node, visit func(ast.Node)) {
	visit(node)
	for _, child := range syntaxChildren(node) {
		walkSyntax(child, visit)
	}
}

func rootIdentifier(expression ast.Expression) string {
	switch value := expression.(type) {
	case *ast.Identifier:
		return value.Name.String()
	case *ast.DotExpression:
		return rootIdentifier(value.Left)
	case *ast.BracketExpression:
		return rootIdentifier(value.Left)
	}
	return ""
}

func mutatedNames(body ast.Node) map[string]bool {
	result := map[string]bool{}
	walkSyntax(body, func(node ast.Node) {
		switch value := node.(type) {
		case *ast.AssignExpression:
			result[rootIdentifier(value.Left)] = true
		case *ast.UnaryExpression:
			if value.Operator == token.INCREMENT || value.Operator == token.DECREMENT || value.Operator == token.DELETE {
				result[rootIdentifier(value.Operand)] = true
			}
		case *ast.CallExpression:
			if member, ok := value.Callee.(*ast.DotExpression); ok {
				switch member.Identifier.Name.String() {
				case "push", "pop", "shift", "unshift", "splice", "sort", "reverse", "fill", "copyWithin":
					result[rootIdentifier(member.Left)] = true
				}
			}
		}
	})
	return result
}

func immutableBindings(body *ast.BlockStatement, mutable map[string]bool) map[string]any {
	result := map[string]any{}
	for _, statement := range body.List {
		declaration, ok := statement.(*ast.LexicalDeclaration)
		if !ok || declaration.Token != token.CONST {
			continue
		}
		for _, binding := range declaration.List {
			name, ok := binding.Target.(*ast.Identifier)
			if !ok || mutable[name.Name.String()] {
				continue
			}
			if value, ok := staticValue(binding.Initializer, result, 0); ok {
				result[name.Name.String()] = value
			}
		}
	}
	return result
}

func runtimeControl(node ast.Node) bool {
	switch value := node.(type) {
	case *ast.IfStatement, *ast.WhileStatement, *ast.DoWhileStatement, *ast.ForStatement, *ast.ForInStatement,
		*ast.SwitchStatement, *ast.TryStatement, *ast.CatchStatement, *ast.FunctionDeclaration, *ast.FunctionLiteral,
		*ast.ArrowFunctionLiteral, *ast.ReturnStatement, *ast.ThrowStatement, *ast.BranchStatement, *ast.ConditionalExpression:
		return true
	case *ast.BinaryExpression:
		return value.Operator == token.LOGICAL_AND || value.Operator == token.LOGICAL_OR || value.Operator == token.COALESCE
	}
	return false
}

func closedLoopBindings(loop *ast.ForOfStatement, bindings map[string]any, mutable map[string]bool) ([]map[string]any, bool) {
	declaration, ok := loop.Into.(*ast.ForDeclaration)
	if !ok || !declaration.IsConst || loop.Await {
		return nil, false
	}
	var names []string
	switch target := declaration.Target.(type) {
	case *ast.Identifier:
		names = append(names, target.Name.String())
	case *ast.ArrayPattern:
		if target.Rest != nil {
			return nil, false
		}
		for _, element := range target.Elements {
			name, ok := element.(*ast.Identifier)
			if !ok {
				return nil, false
			}
			names = append(names, name.Name.String())
		}
	default:
		return nil, false
	}
	for _, name := range names {
		if mutable[name] {
			return nil, false
		}
	}
	closed := true
	walkSyntax(loop.Body, func(node ast.Node) {
		if runtimeControl(node) {
			closed = false
		}
		if _, nested := node.(*ast.ForOfStatement); nested {
			closed = false
		}
	})
	if !closed {
		return nil, false
	}
	source, ok := staticValue(loop.Source, bindings, 0)
	array, valid := source.([]any)
	if !ok || !valid || len(array) == 0 {
		return nil, false
	}
	result := make([]map[string]any, 0, len(array))
	for _, element := range array {
		candidate := make(map[string]any, len(bindings)+len(names))
		for key, value := range bindings {
			candidate[key] = value
		}
		if _, single := declaration.Target.(*ast.Identifier); single {
			candidate[names[0]] = element
		} else {
			tuple, ok := element.([]any)
			if !ok || len(tuple) != len(names) {
				return nil, false
			}
			for i, name := range names {
				candidate[name] = tuple[i]
			}
		}
		result = append(result, candidate)
	}
	return result, true
}

func semanticWrapper(code, path string, depth int) []SemanticTool {
	body, err := parseWrapper(code)
	if err != nil {
		return []SemanticTool{{Path: path, Tool: "exec", Bucket: "unresolved_structure", Reason: "unsupported_js_structure", Structure: "wrapper"}}
	}
	mutable := mutatedNames(body)
	rebound, opaque := mutable["tools"], false
	walkSyntax(body, func(node ast.Node) {
		if binding, ok := node.(*ast.Binding); ok && rootIdentifier(binding.Target) == "tools" {
			rebound = true
		}
		if identifier, ok := node.(*ast.Identifier); ok && (identifier.Name.String() == "eval" || identifier.Name.String() == "Function" || identifier.Name.String() == "tools") {
			opaque = true
		}
	})
	if rebound {
		return []SemanticTool{{Path: path, Tool: "exec", Bucket: "unresolved_structure", Reason: "tools_rebound", Structure: "wrapper"}}
	}
	bindings := immutableBindings(body, mutable)
	var result []SemanticTool
	ordinal := 0
	var visit func(ast.Node, bool, []map[string]any, bool)
	visit = func(node ast.Node, unsafe bool, candidates []map[string]any, inLoop bool) {
		unsafe = unsafe || runtimeControl(node)
		if loop, ok := node.(*ast.ForOfStatement); ok {
			values, closed := closedLoopBindings(loop, bindings, mutable)
			visit(loop.Body, unsafe || !closed, values, true)
			return
		}
		if call, ok := node.(*ast.CallExpression); ok {
			name, direct, computed := toolCallName(call)
			if direct || computed {
				p := fmt.Sprintf("%s.%d", path, ordinal)
				ordinal++
				failed := SemanticTool{Path: p, Tool: name, Bucket: "unresolved_structure", Reason: "dynamic_arguments", Structure: "wrapper"}
				if computed {
					failed.Tool, failed.Reason = "exec", "computed_tool_name"
				}
				if unsafe {
					failed.Reason = "conditional_or_function_call"
				}
				if unsafe || computed || len(call.ArgumentList) != 1 || (inLoop && canonicalTool(name) == "exec") {
					result = append(result, failed)
					return
				}
				if !inLoop {
					candidates = []map[string]any{bindings}
				}
				var selected []SemanticTool
				for _, candidate := range candidates {
					value, ok := staticValue(call.ArgumentList[0], candidate, 0)
					if !ok {
						selected = nil
						break
					}
					encoded, err := json.Marshal(value)
					if err != nil {
						selected = nil
						break
					}
					children := semanticTool(name, encoded, p, depth)
					if inLoop && (len(children) != 1 || children[0].Bucket == "unresolved_structure" || selected != nil && (selected[0].Bucket != children[0].Bucket || toolKind(selected[0].Tool, selected[0].args) != toolKind(children[0].Tool, children[0].args))) {
						selected = nil
						break
					}
					if selected == nil {
						selected = children
					}
				}
				if selected == nil {
					if inLoop {
						failed.Reason = "closed_loop_candidate_unresolved_or_disagreement"
					}
					result = append(result, failed)
					return
				}
				for i := range selected {
					if inLoop {
						selected[i].Structure = fmt.Sprintf("wrapper/closed_forof_candidates:%d/%s", len(candidates), selected[i].Structure)
					} else {
						selected[i].Structure = "wrapper/" + selected[i].Structure
					}
				}
				result = append(result, selected...)
				return
			}
		}
		for _, child := range syntaxChildren(node) {
			visit(child, unsafe, candidates, inLoop)
		}
	}
	visit(body, false, nil, false)
	if len(result) == 0 {
		bucket, reason := "noncognitive_control", "no_tool_invocation"
		if opaque {
			bucket, reason = "unresolved_structure", "opaque_dispatch"
		}
		result = append(result, SemanticTool{Path: path, Tool: "exec", Bucket: bucket, Reason: reason, Structure: "wrapper"})
	}
	return result
}

func toolCallName(call *ast.CallExpression) (string, bool, bool) {
	if member, ok := call.Callee.(*ast.DotExpression); ok {
		if owner, ok := member.Left.(*ast.Identifier); ok && owner.Name.String() == "tools" {
			return member.Identifier.Name.String(), true, false
		}
	}
	if member, ok := call.Callee.(*ast.BracketExpression); ok {
		if owner, ok := member.Left.(*ast.Identifier); ok && owner.Name.String() == "tools" {
			return "", false, true
		}
	}
	return "", false, false
}
