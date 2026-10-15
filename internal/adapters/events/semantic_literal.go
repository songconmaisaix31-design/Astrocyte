package events

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/token"
)

type staticDataFunction struct {
	parameter string
	body      ast.Expression
}

// 仅计算成熟 AST 中已经确认的常量数据，函数和外部调用不会执行。
func staticValue(expression ast.Expression, bindings map[string]any, depth int) (any, bool) {
	if expression == nil || depth > 64 {
		return nil, false
	}
	next := func(value ast.Expression) (any, bool) { return staticValue(value, bindings, depth+1) }
	switch value := expression.(type) {
	case *ast.StringLiteral:
		return value.Value.String(), true
	case *ast.NumberLiteral:
		switch number := value.Value.(type) {
		case int64:
			return float64(number), true
		case float64:
			return number, true
		}
	case *ast.BooleanLiteral:
		return value.Value, true
	case *ast.NullLiteral:
		return nil, true
	case *ast.Identifier:
		result, ok := bindings[value.Name.String()]
		return result, ok
	case *ast.ArrayLiteral:
		result := make([]any, 0, len(value.Value))
		for _, element := range value.Value {
			parsed, ok := next(element)
			if !ok {
				return nil, false
			}
			result = append(result, parsed)
		}
		return result, true
	case *ast.ObjectLiteral:
		result := map[string]any{}
		for _, element := range value.Value {
			var key string
			var input ast.Expression
			switch property := element.(type) {
			case *ast.PropertyShort:
				if property.Initializer != nil {
					return nil, false
				}
				key, input = property.Name.Name.String(), &property.Name
			case *ast.PropertyKeyed:
				if property.Computed || property.Kind != ast.PropertyKindValue {
					return nil, false
				}
				switch name := property.Key.(type) {
				case *ast.StringLiteral:
					key = name.Value.String()
				case *ast.Identifier:
					key = name.Name.String()
				default:
					return nil, false
				}
				input = property.Value
			default:
				return nil, false
			}
			parsed, ok := next(input)
			if !ok {
				return nil, false
			}
			result[key] = parsed
		}
		return result, true
	case *ast.BinaryExpression:
		if value.Operator != token.PLUS {
			return nil, false
		}
		left, a := next(value.Left)
		right, b := next(value.Right)
		ls, lok := left.(string)
		rs, rok := right.(string)
		if a && b && lok && rok {
			return ls + rs, true
		}
	case *ast.UnaryExpression:
		if value.Operator == token.MINUS && !value.Postfix {
			parsed, ok := next(value.Operand)
			if number, valid := parsed.(float64); ok && valid {
				return -number, true
			}
		}
	case *ast.TemplateLiteral:
		if value.Tag != nil || len(value.Elements) != len(value.Expressions)+1 {
			return nil, false
		}
		var text strings.Builder
		for i, element := range value.Elements {
			if !element.Valid {
				return nil, false
			}
			text.WriteString(element.Parsed.String())
			if i < len(value.Expressions) {
				parsed, ok := next(value.Expressions[i])
				if !ok {
					return nil, false
				}
				part, ok := scalarText(parsed)
				if !ok {
					return nil, false
				}
				text.WriteString(part)
			}
		}
		return text.String(), true
	case *ast.DotExpression:
		parsed, ok := next(value.Left)
		if !ok {
			return nil, false
		}
		key := value.Identifier.Name.String()
		switch object := parsed.(type) {
		case map[string]any:
			item, found := object[key]
			return item, found
		case []any:
			if key == "length" {
				return float64(len(object)), true
			}
		case string:
			if key == "length" {
				return float64(len(object)), true
			}
		}
	case *ast.BracketExpression:
		parsed, ok := next(value.Left)
		key, valid := next(value.Member)
		if !ok || !valid {
			return nil, false
		}
		switch object := parsed.(type) {
		case map[string]any:
			name, ok := key.(string)
			item, found := object[name]
			return item, ok && found
		case []any:
			index, ok := key.(float64)
			if ok && index >= 0 && index == float64(int(index)) && index < float64(len(object)) {
				return object[int(index)], true
			}
		}
	case *ast.ArrowFunctionLiteral:
		if value.Async || value.ParameterList.Rest != nil || len(value.ParameterList.List) != 1 {
			return nil, false
		}
		parameter := value.ParameterList.List[0]
		name, ok := parameter.Target.(*ast.Identifier)
		body, valid := value.Body.(*ast.ExpressionBody)
		if ok && valid && parameter.Initializer == nil {
			return staticDataFunction{name.Name.String(), body.Expression}, true
		}
	case *ast.CallExpression:
		if name, ok := value.Callee.(*ast.Identifier); ok && len(value.ArgumentList) == 1 {
			function, valid := bindings[name.Name.String()].(staticDataFunction)
			argument, found := next(value.ArgumentList[0])
			if valid && found {
				return staticValue(function.body, map[string]any{function.parameter: argument}, depth+1)
			}
			return nil, false
		}
		member, ok := value.Callee.(*ast.DotExpression)
		if !ok {
			return nil, false
		}
		method := member.Identifier.Name.String()
		if owner, ok := member.Left.(*ast.Identifier); ok && owner.Name.String() == "JSON" && method == "stringify" && len(value.ArgumentList) == 1 {
			argument, ok := next(value.ArgumentList[0])
			if !ok {
				return nil, false
			}
			encoded, err := json.Marshal(argument)
			return string(encoded), err == nil
		}
		subject, ok := next(member.Left)
		if !ok {
			return nil, false
		}
		if method == "join" && len(value.ArgumentList) == 1 {
			separator, valid := next(value.ArgumentList[0])
			sep, isString := separator.(string)
			array, isArray := subject.([]any)
			if !valid || !isString || !isArray {
				return nil, false
			}
			parts := make([]string, 0, len(array))
			for _, element := range array {
				part, ok := scalarText(element)
				if !ok {
					return nil, false
				}
				parts = append(parts, part)
			}
			return strings.Join(parts, sep), true
		}
		if (method == "replace" || method == "replaceAll") && len(value.ArgumentList) == 2 {
			text, ok := subject.(string)
			replacement, valid := next(value.ArgumentList[1])
			with, isString := replacement.(string)
			if !ok || !valid || !isString || strings.Contains(with, "$") {
				return nil, false
			}
			count := 1
			if method == "replaceAll" {
				count = -1
			}
			var pattern string
			if regex, ok := value.ArgumentList[0].(*ast.RegExpLiteral); ok {
				if regex.Flags != "" && regex.Flags != "g" {
					return nil, false
				}
				if regex.Flags == "g" {
					count = -1
				}
				switch regex.Pattern {
				case "'", "\"", "`":
					pattern = regex.Pattern
				case `\\`:
					pattern = "\\"
				case `\n`:
					pattern = "\n"
				case `\r`:
					pattern = "\r"
				case `\r\n`:
					pattern = "\r\n"
				default:
					return nil, false
				}
			} else {
				parsed, ok := next(value.ArgumentList[0])
				pattern, valid = parsed.(string)
				if !ok || !valid {
					return nil, false
				}
			}
			return strings.Replace(text, pattern, with, count), true
		}
	}
	return nil, false
}

func scalarText(value any) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), true
	case bool:
		return strconv.FormatBool(value), true
	}
	return "", false
}
