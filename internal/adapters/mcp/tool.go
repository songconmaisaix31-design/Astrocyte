package mcp

import (
	"context"
	"encoding/json"
)

// Handler executes one tool call. arguments is the raw JSON object passed to
// tools/call; it returns any JSON-serializable result.
type Handler func(ctx context.Context, arguments json.RawMessage) (any, error)

type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) Error() string { return e.Message }

// Tool is a single MCP tool definition plus its handler.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     Handler
}
