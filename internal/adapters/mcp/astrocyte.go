package mcp

import (
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// NewAstrocyteServer wires the five cognitive-interface tools (task book §7)
// to a backend. The backend owns all attention/experience data flow; this
// server only decodes arguments, calls the backend and shapes MCP results.
func NewAstrocyteServer(backend app.MCPBackend) *Server {
	s := NewServer("astrocyte", "0.1.0")

	s.RegisterTool(Tool{
		Name:        "astrocyte.get_attention_state",
		Description: "Return the current focus ranking as (topic, attention, momentum), ordered by attention.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Handler: func(ctx context.Context, _ json.RawMessage) (any, error) {
			return backend.GetAttentionState(ctx)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.explain_focus",
		Description: "Explain why a topic holds its current attention: evidence events, weight path and related nodes, each with a timestamp.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"topic": map[string]any{"type": "string"},
			},
			"required": []string{"topic"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.ExplainFocusInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return backend.ExplainFocus(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.get_experiences",
		Description: "Retrieve approved experiences matching a pattern, optionally filtered to a project scope.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string"},
				"scope":   map[string]any{"type": "string"},
			},
			"required": []string{"pattern"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.GetExperiencesInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return backend.GetExperiences(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.report_event",
		Description: "Report an event observed by an external agent into the attention signal stream.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"source":  map[string]any{"type": "string"},
				"session": map[string]any{"type": "string"},
				"ts":      map[string]any{"type": "integer"},
				"kind":    map[string]any{"type": "string"},
				"tokens":  map[string]any{"type": "integer"},
				"scope":   map[string]any{"type": "string"},
				"topic":   map[string]any{"type": "string"},
				"raw":     map[string]any{"type": "object"},
			},
			"required": []string{"source", "session", "ts", "kind", "scope", "topic"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.ReportEventInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return backend.ReportEvent(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.veto",
		Description: "Reflect a user veto of a topic or task back into the attention model.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target": map[string]any{"type": "string"},
				"reason": map[string]any{"type": "string"},
			},
			"required": []string{"target"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.VetoInput
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
			return backend.Veto(ctx, in)
		},
	})

	return s
}
