package mcp

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// NewAstrocyteServer wires the five cognitive-interface tools (task book §7)
// to a backend. The backend owns all attention/experience data flow; this
// server only decodes arguments, calls the backend and shapes MCP results.
func NewAstrocyteServer(backend app.MCPBackend, binding app.MCPBinding) (*Server, error) {
	backend, err := app.BindMCPBackend(backend, binding)
	if err != nil {
		return nil, err
	}
	s := NewServer("astrocyte", "0.1.0")

	s.RegisterTool(Tool{
		Name:        "astrocyte.get_attention_state",
		Description: "Return the current focus ranking as (topic, attention, momentum), ordered by attention.",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			if err := decodeArguments(args, &struct{}{}); err != nil {
				return nil, err
			}
			return backend.GetAttentionState(ctx)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.explain_focus",
		Description: "Explain why a topic holds its current attention: evidence events, weight path and related nodes, each with a timestamp.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"topic": map[string]any{"type": "string"},
			},
			"required": []string{"topic"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.ExplainFocusInput
			if err := decodeArguments(args, &in); err != nil {
				return nil, err
			}
			if in.Topic == "" {
				return nil, &InvalidParamsError{"topic is required"}
			}
			return backend.ExplainFocus(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.get_experiences",
		Description: "Retrieve approved experiences matching a pattern within the startup-bound project scope.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string"},
				"scope":   map[string]any{"type": "string"},
			},
			"required": []string{"pattern"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.GetExperiencesInput
			if err := decodeArguments(args, &in); err != nil {
				return nil, err
			}
			if in.Pattern == "" {
				return nil, &InvalidParamsError{"pattern is required"}
			}
			return backend.GetExperiences(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.report_event",
		Description: "Report an event observed by an external agent into the attention signal stream.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{"event": map[string]any{
				"type": "object", "additionalProperties": false,
				"properties": map[string]any{
					"source":  map[string]any{"type": "string"},
					"session": map[string]any{"type": "string"},
					"ts":      map[string]any{"type": "integer"},
					"kind":    map[string]any{"type": "string"},
					"tokens":  map[string]any{"type": "integer"},
					"scope":   map[string]any{"type": "string"},
					"topic":   map[string]any{"type": "string"},
					"raw":     map[string]any{},
				},
				"required": []string{"ts", "kind"},
			}},
			"required": []string{"event"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var request struct {
				Event json.RawMessage `json:"event"`
			}
			if err := decodeArguments(args, &request); err != nil {
				return nil, err
			}
			var in app.ReportEventInput
			if len(bytes.TrimSpace(request.Event)) == 0 || bytes.TrimSpace(request.Event)[0] != '{' {
				return nil, &InvalidParamsError{"event must be an object"}
			}
			if err := decodeArguments(request.Event, &in); err != nil {
				return nil, err
			}
			in.OriginalPayload = append(json.RawMessage(nil), request.Event...)
			if in.Kind == "" || in.TS <= 0 || (in.Tokens != nil && *in.Tokens < 0) {
				return nil, &InvalidParamsError{"event requires positive ts, kind and nonnegative tokens"}
			}
			return backend.ReportEvent(ctx, in)
		},
	})

	s.RegisterTool(Tool{
		Name:        "astrocyte.veto",
		Description: "Reflect a user veto of a topic or task back into the attention model.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"topic":  map[string]any{"type": "string"},
				"task":   map[string]any{"type": "string"},
				"reason": map[string]any{"type": "string"},
			},
			"oneOf": []any{map[string]any{"required": []string{"topic"}, "not": map[string]any{"required": []string{"task"}}}, map[string]any{"required": []string{"task"}, "not": map[string]any{"required": []string{"topic"}}}},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var in app.VetoInput
			if err := decodeArguments(args, &in); err != nil {
				return nil, err
			}
			if (in.Topic == "") == (in.Task == "") {
				return nil, &InvalidParamsError{"exactly one topic or task is required"}
			}
			return backend.Veto(ctx, in)
		},
	})

	return s, nil
}

func decodeArguments(args json.RawMessage, value any) error {
	d := json.NewDecoder(bytes.NewReader(args))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return &InvalidParamsError{err.Error()}
	}
	return nil
}
