package mcp

import (
	"context"
	"encoding/json"
)

// 工具调用接收原 JSON 参数并返回可序列化结果。
type Handler func(ctx context.Context, arguments json.RawMessage) (any, error)

type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) Error() string { return e.Message }

// 工具定义与其处理函数使用同一注册项。
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     Handler
}
