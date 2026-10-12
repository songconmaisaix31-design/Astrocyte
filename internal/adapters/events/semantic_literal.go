package events

import (
	"encoding/json"
	"strconv"
	"strings"
)

type staticDataFunction struct {
	parameter string
	body      []semanticToken
}
type staticQuotePattern struct {
	text string
	all  bool
}

// This is a finite data-expression grammar, not JS evaluation. Only immutable
// literal const bindings, string concatenation, literal array join and JSON
// serialization are supported. Every other expression stays unknown.
func staticExpression(t []semanticToken, i *int, bindings map[string]any) (any, bool) {
	value, ok := staticPrimary(t, i, bindings)
	if !ok {
		return nil, false
	}
	for *i < len(t) {
		if *i+1 < len(t) && t[*i].value == "." && (*i+2 >= len(t) || t[*i+2].value != "(") {
			key := t[*i+1].value
			*i += 2
			switch object := value.(type) {
			case map[string]any:
				next, ok := object[key]
				if !ok {
					return nil, false
				}
				value = next
			case []any:
				if key != "length" {
					return nil, false
				}
				value = float64(len(object))
			case string:
				if key != "length" {
					return nil, false
				}
				value = float64(len(object))
			default:
				return nil, false
			}
			continue
		}
		if t[*i].value == "[" {
			*i++
			index, ok := staticExpression(t, i, bindings)
			if !ok || *i >= len(t) || t[*i].value != "]" {
				return nil, false
			}
			*i++
			switch object := value.(type) {
			case map[string]any:
				key, ok := index.(string)
				if !ok {
					return nil, false
				}
				next, ok := object[key]
				if !ok {
					return nil, false
				}
				value = next
			case []any:
				n, ok := index.(float64)
				if !ok || n != float64(int(n)) || n < 0 || n >= float64(len(object)) {
					return nil, false
				}
				value = object[int(n)]
			default:
				return nil, false
			}
			continue
		}
		if *i+3 < len(t) && t[*i].value == "." && (t[*i+1].value == "replace" || t[*i+1].value == "replaceAll") && t[*i+2].value == "(" {
			method := t[*i+1].value
			*i += 3
			pattern, ok := staticExpression(t, i, bindings)
			if !ok || *i >= len(t) || t[*i].value != "," {
				return nil, false
			}
			*i++
			replacement, ok := staticExpression(t, i, bindings)
			if !ok || *i >= len(t) || t[*i].value != ")" {
				return nil, false
			}
			*i++
			subject, sok := value.(string)
			with, wok := replacement.(string)
			if !sok || !wok || strings.Contains(with, "$") {
				return nil, false
			}
			count := 1
			if method == "replaceAll" {
				count = -1
			}
			var target string
			switch p := pattern.(type) {
			case string:
				target = p
			case staticQuotePattern:
				target = p.text
				if p.all {
					count = -1
				}
			default:
				return nil, false
			}
			value = strings.Replace(subject, target, with, count)
			continue
		}
		if t[*i].value == "+" {
			*i++
			right, ok := staticExpression(t, i, bindings)
			if !ok {
				return nil, false
			}
			a, aok := value.(string)
			b, bok := right.(string)
			if !aok || !bok {
				return nil, false
			}
			value = a + b
			continue
		}
		if *i+3 < len(t) && t[*i].value == "." && t[*i+1].value == "join" && t[*i+2].value == "(" {
			*i += 3
			separator, ok := staticExpression(t, i, bindings)
			if !ok || *i >= len(t) || t[*i].value != ")" {
				return nil, false
			}
			*i++
			sep, ok := separator.(string)
			array, aok := value.([]any)
			if !ok || !aok {
				return nil, false
			}
			var parts []string
			for _, item := range array {
				switch v := item.(type) {
				case string:
					parts = append(parts, v)
				case float64:
					parts = append(parts, strconv.FormatFloat(v, 'g', -1, 64))
				default:
					return nil, false
				}
			}
			value = strings.Join(parts, sep)
			continue
		}
		break
	}
	return value, true
}

func staticPrimary(t []semanticToken, i *int, bindings map[string]any) (any, bool) {
	if *i >= len(t) {
		return nil, false
	}
	token := t[*i]
	*i++
	if token.literal {
		if !token.dynamic {
			return token.value, true
		}
		// Only scalar literal/const substitutions can resolve templates.
		var b strings.Builder
		s := token.value
		for {
			start := strings.Index(s, "${")
			if start < 0 {
				b.WriteString(s)
				return b.String(), true
			}
			b.WriteString(s[:start])
			end, ok := skipJSInterpolation(s, start+2)
			if !ok {
				return nil, false
			}
			expr, valid := semanticTokens(s[start+2 : end-1])
			if !valid {
				return nil, false
			}
			at := 0
			v, ok := staticExpression(expr, &at, bindings)
			if !ok || at != len(expr) {
				return nil, false
			}
			switch v := v.(type) {
			case string:
				b.WriteString(v)
			case float64:
				b.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
			case bool:
				b.WriteString(strconv.FormatBool(v))
			default:
				return nil, false
			}
			s = s[end:]
		}
	}
	if token.dynamic {
		if strings.HasPrefix(token.value, "/") {
			end := strings.LastIndex(token.value, "/")
			if end > 0 {
				pattern, flags := token.value[1:end], token.value[end+1:]
				if flags != "" && flags != "g" {
					return nil, false
				}
				switch pattern {
				case "'", "\"", "`":
					return staticQuotePattern{pattern, flags == "g"}, true
				case `\\`:
					return staticQuotePattern{"\\", flags == "g"}, true
				case `\n`:
					return staticQuotePattern{"\n", flags == "g"}, true
				case `\r`:
					return staticQuotePattern{"\r", flags == "g"}, true
				case `\r\n`:
					return staticQuotePattern{"\r\n", flags == "g"}, true
				}
			}
		}
		return nil, false
	}
	switch token.value {
	case "{":
		result := map[string]any{}
		for *i < len(t) && t[*i].value != "}" {
			key := t[*i].value
			*i++
			if *i < len(t) && t[*i].value == ":" {
				*i++
				v, ok := staticExpression(t, i, bindings)
				if !ok {
					return nil, false
				}
				result[key] = v
			} else {
				v, ok := bindings[key]
				if !ok {
					return nil, false
				}
				result[key] = v
			}
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
			v, ok := staticExpression(t, i, bindings)
			if !ok {
				return nil, false
			}
			result = append(result, v)
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
	case "(":
		v, ok := staticExpression(t, i, bindings)
		if !ok || *i >= len(t) || t[*i].value != ")" {
			return nil, false
		}
		*i++
		return v, true
	case "true":
		return true, true
	case "false":
		return false, true
	case "null":
		return nil, true
	case "JSON":
		if *i+2 >= len(t) || t[*i].value != "." || t[*i+1].value != "stringify" || t[*i+2].value != "(" {
			return nil, false
		}
		*i += 3
		v, ok := staticExpression(t, i, bindings)
		if !ok || *i >= len(t) || t[*i].value != ")" {
			return nil, false
		}
		*i++
		b, err := json.Marshal(v)
		return string(b), err == nil
	}
	if v, ok := bindings[token.value]; ok {
		if function, ok := v.(staticDataFunction); ok {
			if *i >= len(t) || t[*i].value != "(" {
				return nil, false
			}
			*i++
			arg, ok := staticExpression(t, i, bindings)
			if !ok || *i >= len(t) || t[*i].value != ")" {
				return nil, false
			}
			*i++
			nested := map[string]any{}
			for k, v := range bindings {
				if _, function := v.(staticDataFunction); !function {
					nested[k] = v
				}
			}
			nested[function.parameter] = arg
			at := 0
			value, ok := staticExpression(function.body, &at, nested)
			return value, ok && at == len(function.body)
		}
		return v, true
	}
	n, err := strconv.ParseFloat(token.value, 64)
	return n, err == nil
}

func staticBindings(tokens []semanticToken) map[string]any {
	bindings := map[string]any{}
	seen := map[string]bool{}
	for j := 0; j+3 < len(tokens); j++ {
		if tokens[j].value == "function" && j+7 < len(tokens) && tokens[j+2].value == "(" && tokens[j+4].value == ")" && tokens[j+5].value == "{" && tokens[j+6].value == "return" {
			end := j + 7
			for end < len(tokens) && tokens[end].value != ";" && tokens[end].value != "}" {
				end++
			}
			if end < len(tokens) {
				bindings[tokens[j+1].value] = staticDataFunction{tokens[j+3].value, tokens[j+7 : end]}
			}
		}
		if tokens[j].literal || tokens[j].value != "const" || tokens[j+1].literal || tokens[j+2].value != "=" {
			continue
		}
		name := tokens[j+1].value
		if seen[name] {
			delete(bindings, name)
			continue
		}
		seen[name] = true
		// Compile only a single-parameter, expression-only data helper. No
		// statement/callback/loop or arbitrary historical function is evaluated.
		arrow := j + 3
		parameter := ""
		bodyStart := 0
		if arrow+2 < len(tokens) && tokens[arrow+1].value == "=" && tokens[arrow+2].value == ">" {
			parameter = tokens[arrow].value
			bodyStart = arrow + 3
		}
		if arrow+4 < len(tokens) && tokens[arrow].value == "(" && tokens[arrow+2].value == ")" && tokens[arrow+3].value == "=" && tokens[arrow+4].value == ">" {
			parameter = tokens[arrow+1].value
			bodyStart = arrow + 5
		}
		if parameter != "" && bodyStart < len(tokens) && tokens[bodyStart].value != "{" {
			end := bodyStart
			for end < len(tokens) && tokens[end].value != ";" {
				end++
			}
			if end < len(tokens) {
				bindings[name] = staticDataFunction{parameter, tokens[bodyStart:end]}
				continue
			}
		}
		at := j + 3
		v, ok := staticExpression(tokens, &at, bindings)
		if ok && at < len(tokens) && (tokens[at].value == ";" || tokens[at].value == "," || tokens[at].value == ")") {
			bindings[name] = v
		}
	}
	// Const references to objects/arrays can still be mutated. Reject observed
	// writes and mutating methods instead of freezing the earlier literal value.
	for j, token := range tokens {
		if token.literal {
			continue
		}
		value, exists := bindings[token.value]
		if !exists {
			continue
		}
		switch value.(type) {
		case map[string]any, []any:
		default:
			continue
		}
		end := j + 1
		for end < len(tokens) && (tokens[end].value == "." || tokens[end].value == "[") {
			if tokens[end].value == "." {
				if end+1 >= len(tokens) {
					break
				}
				method := tokens[end+1].value
				switch method {
				case "push", "pop", "shift", "unshift", "splice", "sort", "reverse", "fill", "copyWithin":
					delete(bindings, token.value)
				}
				end += 2
			} else {
				end++
				for end < len(tokens) && tokens[end].value != "]" {
					end++
				}
				end++
			}
		}
		if end > j+1 && end < len(tokens) && (tokens[end].value == "=" || tokens[end].value == "+" || tokens[end].value == "-") {
			delete(bindings, token.value)
		}
	}
	return bindings
}

func staticArgument(tokens []semanticToken, start int, bindings map[string]any) (any, int, bool) {
	i := start
	v, ok := staticExpression(tokens, &i, bindings)
	return v, i, ok
}
