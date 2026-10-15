package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestPendingEventProjectionDoesNotInventMapping(t *testing.T) {
	projected := exposeEvent(app.StoredEvent{ID: "pending-receipt", RawBytes: []byte(`{"session":"PRIVATE_MCP_UNMAPPED_2930"}`)})
	encoded, err := json.Marshal(projected)
	if err != nil {
		t.Fatal("event projection serialization failed")
	}
	var fields map[string]any
	json.Unmarshal(encoded, &fields)
	if len(fields) != 6 || fields["event_id"] != "pending-receipt" || fields["tokens"] != nil || fields["ts"] != nil || fields["topic"] != "" || fields["kind"] != "" || fields["scope"] != "" {
		t.Error("pending mapping invented derived observations")
	}
	if bytes.Contains(encoded, []byte("PRIVATE_MCP_UNMAPPED_2930")) {
		t.Error("pending event projection exposed original metadata")
	}
}

type privateMarshalFailure struct{}

func (privateMarshalFailure) MarshalJSON() ([]byte, error) {
	return nil, errors.New("PRIVATE_MCP_MARSHAL_FAILURE_81923")
}

func TestPublicProtocolErrorsDoNotEchoDetails(t *testing.T) {
	const private = "PRIVATE_MCP_MARSHAL_FAILURE_81923"
	for _, scenario := range []struct {
		name, method, params string
	}{
		{"method", private, `{}`},
		{"tool", "tools/call", `{"name":"` + private + `","arguments":{}}`},
		{"params", "tools/call", `{"name":{"` + private + `":true},"arguments":{}}`},
		{"marshal", "tools/call", `{"name":"marshal","arguments":{}}`},
		{"invalid_backend", "tools/call", `{"name":"invalid","arguments":{}}`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			s := NewServer("fixture", "1")
			s.RegisterTool(Tool{Name: "marshal", Handler: func(context.Context, json.RawMessage) (any, error) { return privateMarshalFailure{}, nil }})
			s.RegisterTool(Tool{Name: "invalid", Handler: func(context.Context, json.RawMessage) (any, error) {
				return nil, &InvalidParamsError{Message: private}
			}})
			lines := serveLines(t, s, reqLine(1, scenario.method, scenario.params))
			if len(lines) != 1 || lines[0]["error"] == nil {
				t.Fatal("expected protocol error")
			}
			wire, err := json.Marshal(lines[0])
			if err != nil || strings.Contains(string(wire), private) {
				t.Error("public protocol error exposed details; contents suppressed")
			}
		})
	}
}
