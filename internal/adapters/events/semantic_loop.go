package events

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type staticLoop struct {
	start, body, end int
	candidates       []map[string]any
}

// closedLoops recognizes data bindings only. It does not execute a loop or
// infer its iteration count as event contribution. Every lexical call has one
// identity and must agree across all closed candidates before mapping.
func closedLoops(t []semanticToken, bindings map[string]any) []staticLoop {
	var loops []staticLoop
	for i := 0; i+5 < len(t); i++ {
		if t[i].literal || t[i].value != "for" || t[i+1].value != "(" || t[i+2].value != "const" {
			continue
		}
		at := i + 3
		var names []string
		if t[at].value == "[" {
			at++
			for at < len(t) && t[at].value != "]" {
				if !staticIdentifier(t[at]) {
					break
				}
				names = append(names, t[at].value)
				at++
				if at < len(t) && t[at].value == "," {
					at++
				} else {
					break
				}
			}
			if at >= len(t) || t[at].value != "]" {
				continue
			}
			at++
		} else if staticIdentifier(t[at]) {
			names = []string{t[at].value}
			at++
		}
		if len(names) == 0 || at >= len(t) || t[at].value != "of" {
			continue
		}
		at++
		value, ok := staticExpression(t, &at, bindings)
		array, isArray := value.([]any)
		if !ok || !isArray || len(array) == 0 || at+1 >= len(t) || t[at].value != ")" {
			continue
		}
		body := at + 1
		bodyData := body
		end := -1
		if t[body].value == "{" {
			end = matchingBrace(t, body)
			bodyData++
		} else {
			end = closedStatementEnd(t, body)
		}
		if end < 0 {
			continue
		}
		// Nested control, early termination, mutation or an arbitrary function
		// body would require execution. Preserve those as unknown.
		if !closedLoopBody(t[bodyData:end], names) {
			continue
		}
		loop := staticLoop{start: i, body: body, end: end}
		for _, element := range array {
			b := make(map[string]any, len(bindings)+len(names))
			for k, v := range bindings {
				b[k] = v
			}
			if len(names) == 1 && t[i+3].value != "[" {
				b[names[0]] = element
			} else {
				tuple, ok := element.([]any)
				if !ok || len(tuple) != len(names) {
					loop.candidates = nil
					break
				}
				for j, name := range names {
					b[name] = tuple[j]
				}
			}
			loop.candidates = append(loop.candidates, b)
		}
		if len(loop.candidates) == len(array) {
			loops = append(loops, loop)
		}
	}
	return loops
}

func closedStatementEnd(t []semanticToken, start int) int {
	depth, calls := 0, 0
	for i := start; i < len(t); i++ {
		if t[i].literal {
			continue
		}
		if t[i].value == "tools" && i+3 < len(t) && t[i+1].value == "." && t[i+3].value == "(" {
			calls++
		}
		switch t[i].value {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
			if depth < 0 {
				return -1
			}
		case ";":
			if depth == 0 && calls == 1 {
				return i
			}
		}
	}
	if depth == 0 && calls == 1 {
		return len(t)
	}
	return -1
}

var staticIdentifierPattern = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

func staticIdentifier(t semanticToken) bool {
	return !t.literal && staticIdentifierPattern.MatchString(t.value)
}

func matchingBrace(t []semanticToken, start int) int {
	depth := 0
	for i := start; i < len(t); i++ {
		if t[i].literal {
			continue
		}
		if t[i].value == "{" {
			depth++
		}
		if t[i].value == "}" {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func closedLoopBody(t []semanticToken, names []string) bool {
	bound := map[string]bool{}
	for _, n := range names {
		bound[n] = true
	}
	for i, token := range t {
		if token.literal {
			continue
		}
		switch token.value {
		case "if", "for", "while", "switch", "function", "break", "continue", "return", "throw", "try", "catch", "delete":
			return false
		case "=":
			if i+1 < len(t) && t[i+1].value == ">" {
				return false
			}
		}
		if !bound[token.value] {
			continue
		}
		end := i + 1
		for end < len(t) && (t[end].value == "." || t[end].value == "[") {
			if t[end].value == "." {
				if end+1 >= len(t) {
					return false
				}
				switch t[end+1].value {
				case "push", "pop", "shift", "unshift", "splice", "sort", "reverse", "fill", "copyWithin":
					return false
				}
				end += 2
			} else {
				end++
				for end < len(t) && t[end].value != "]" {
					end++
				}
				end++
			}
		}
		if end < len(t) && (t[end].value == "=" || (t[end].value == "+" || t[end].value == "-") && end+1 < len(t) && (t[end+1].value == t[end].value || t[end+1].value == "=")) {
			return false
		}
		if i > 0 && (t[i-1].value == "const" || t[i-1].value == "let" || t[i-1].value == "var") {
			return false
		}
	}
	return true
}

func loopSemanticCall(t []semanticToken, at int, loop staticLoop, name, path string, depth int) ([]SemanticTool, int, bool) {
	var selected []SemanticTool
	end := 0
	for _, binding := range loop.candidates {
		argument, stop, ok := staticArgument(t, at, binding)
		if !ok || stop >= len(t) || t[stop].value != ")" {
			return nil, 0, false
		}
		// A nested wrapper can contain a different number of lexical children.
		// Supporting that would change this loop's frozen contribution unit.
		if canonicalTool(name) == "exec" {
			return nil, 0, false
		}
		a, err := json.Marshal(argument)
		if err != nil {
			return nil, 0, false
		}
		children := semanticTool(name, a, path, depth)
		if len(children) != 1 || children[0].Bucket == "unresolved_structure" {
			return nil, 0, false
		}
		if selected != nil && (selected[0].Bucket != children[0].Bucket || toolKind(selected[0].Tool, selected[0].args) != toolKind(children[0].Tool, children[0].args)) {
			return nil, 0, false
		}
		if selected == nil {
			selected = children
		}
		end = stop
	}
	selected[0].Structure = fmt.Sprintf("wrapper/closed_forof_candidates:%d/%s", len(loop.candidates), selected[0].Structure)
	return selected, end, true
}
