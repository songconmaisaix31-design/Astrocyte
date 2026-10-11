package events

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

type wrapperToken struct {
	value   string
	literal bool
}

// This is a restrictive lexical recognizer, never a JavaScript interpreter.
// Only unconditional await tools.NAME(literal arguments) contributes. Dynamic
// arguments, interpolation, functions/control flow and mixed semantics remain
// unknown. Quoted text/comments cannot become tool calls.
func wrapperTokens(code string) ([]wrapperToken, bool) {
	var result []wrapperToken
	for i := 0; i < len(code); {
		c := code[i]
		if unicode.IsSpace(rune(c)) {
			i++
			continue
		}
		if i+1 < len(code) && code[i:i+2] == "//" {
			for i < len(code) && code[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(code) && code[i:i+2] == "/*" {
			end := strings.Index(code[i+2:], "*/")
			if end < 0 {
				return nil, false
			}
			i += end + 4
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			quote, start := c, i
			i++
			var b strings.Builder
			for i < len(code) && code[i] != quote {
				if quote == '`' && i+1 < len(code) && code[i:i+2] == "${" {
					return nil, false
				}
				if code[i] == '\\' {
					if i+1 >= len(code) {
						return nil, false
					}
					next := code[i+1]
					switch next {
					case 'n':
						b.WriteByte('\n')
					case 'r':
						b.WriteByte('\r')
					case 't':
						b.WriteByte('\t')
					case '\\', '\'', '"', '`':
						b.WriteByte(next)
					default:
						return nil, false
					}
					i += 2
				} else {
					b.WriteByte(code[i])
					i++
				}
			}
			if i >= len(code) {
				return nil, false
			}
			i++
			value := b.String()
			if quote == '"' {
				if err := json.Unmarshal([]byte(code[start:i]), &value); err != nil {
					return nil, false
				}
			}
			result = append(result, wrapperToken{value, true})
			continue
		}
		if unicode.IsLetter(rune(c)) || c == '_' || c == '$' {
			start := i
			i++
			for i < len(code) && (unicode.IsLetter(rune(code[i])) || unicode.IsDigit(rune(code[i])) || code[i] == '_' || code[i] == '$') {
				i++
			}
			result = append(result, wrapperToken{value: code[start:i]})
			continue
		}
		if unicode.IsDigit(rune(c)) {
			start := i
			i++
			for i < len(code) && (unicode.IsDigit(rune(code[i])) || code[i] == '.') {
				i++
			}
			result = append(result, wrapperToken{value: code[start:i]})
			continue
		}
		result = append(result, wrapperToken{value: string(c)})
		i++
	}
	return result, true
}

func literalArgument(t []wrapperToken, i *int) (any, bool) {
	if *i >= len(t) {
		return nil, false
	}
	v := t[*i]
	*i++
	if v.literal {
		return v.value, true
	}
	switch v.value {
	case "{":
		result := map[string]any{}
		for *i < len(t) && t[*i].value != "}" {
			key := t[*i].value
			*i++
			if *i >= len(t) || t[*i].value != ":" {
				return nil, false
			}
			*i++
			value, ok := literalArgument(t, i)
			if !ok {
				return nil, false
			}
			result[key] = value
			if *i < len(t) && t[*i].value == "," {
				*i++
			} else {
				break
			}
		}
		if *i >= len(t) || t[*i].value != "}" {
			return nil, false
		}
		*i++
		return result, true
	case "[":
		var result []any
		for *i < len(t) && t[*i].value != "]" {
			value, ok := literalArgument(t, i)
			if !ok {
				return nil, false
			}
			result = append(result, value)
			if *i < len(t) && t[*i].value == "," {
				*i++
			} else {
				break
			}
		}
		if *i >= len(t) || t[*i].value != "]" {
			return nil, false
		}
		*i++
		return result, true
	case "true":
		return true, true
	case "false":
		return false, true
	case "null":
		return nil, true
	}
	number, err := strconv.ParseFloat(v.value, 64)
	return number, err == nil
}

func wrapperKind(code string) string {
	kind, _ := wrapperAnalysis(code)
	return kind
}

func wrapperAnalysis(code string) (string, string) {
	tokens, ok := wrapperTokens(code)
	if !ok {
		return "unknown", "dynamic_or_unsupported_literal"
	}
	for i, t := range tokens {
		if t.literal {
			continue
		}
		switch t.value {
		case "if", "else", "for", "while", "do", "switch", "try", "catch", "function", "return", "throw", "eval", "Function", "?", "&", "|":
			return "unknown", "control_flow_or_conditional"
		}
		if t.value == "tools" && i+1 < len(tokens) && tokens[i+1].value == "=" {
			return "unknown", "tools_rebound"
		}
		if t.value == "=" && i+1 < len(tokens) && tokens[i+1].value == ">" {
			return "unknown", "function_or_callback"
		}
	}
	kind := ""
	for i := 0; i+4 < len(tokens); i++ {
		if tokens[i].literal || tokens[i].value != "tools" || tokens[i+1].value != "." || tokens[i+3].value != "(" {
			continue
		}
		if i == 0 || tokens[i-1].literal || tokens[i-1].value != "await" || (i > 1 && tokens[i-2].value == ".") {
			return "unknown", "call_not_unconditional_await"
		}
		end := i + 4
		argument, valid := literalArgument(tokens, &end)
		if !valid || end >= len(tokens) || tokens[end].value != ")" {
			return "unknown", "dynamic_arguments"
		}
		name := tokens[i+2].value
		if name == "exec" {
			return "unknown", "recursive_wrapper"
		}
		raw, err := json.Marshal(argument)
		if err != nil {
			return "unknown", "unsupported_literal"
		}
		mapped := toolKind(name, raw)
		if mapped == "unknown" {
			return "unknown", "nested_tool_unsupported_or_non_cognitive"
		}
		if kind != "" && kind != mapped {
			return "unknown", "mixed_cognitive_semantics"
		}
		kind = mapped
		i = end
	}
	if kind == "" {
		return "unknown", "no_static_tool_call"
	}
	return kind, "recognized"
}
