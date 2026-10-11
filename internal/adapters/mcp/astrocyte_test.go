package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type stubBackend struct{}

func (stubBackend) GetAttentionState(context.Context) (app.GetAttentionStateResult, error) {
	return app.GetAttentionStateResult{Topics: []app.FocusItem{{Topic: "Agent Swarm", Attention: 0.8, Momentum: 0.1}}}, nil
}
func (stubBackend) ExplainFocus(context.Context, app.ExplainFocusInput) (app.ExplainFocusResult, error) {
	return app.ExplainFocusResult{}, nil
}
func (stubBackend) GetExperiences(context.Context, app.GetExperiencesInput) (app.GetExperiencesResult, error) {
	return app.GetExperiencesResult{}, nil
}
func (stubBackend) ReportEvent(context.Context, app.ReportEventInput) (app.ReportEventResult, error) {
	return app.ReportEventResult{}, nil
}
func (stubBackend) Veto(context.Context, app.VetoInput) (app.VetoResult, error) {
	return app.VetoResult{}, nil
}

func TestAstrocyteRegistersFiveTools(t *testing.T) {
	s := NewAstrocyteServer(stubBackend{})
	input := reqLine(1, "tools/list", "")
	lines := serveLines(t, s, input)
	tools := lines[0]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{
		"astrocyte.get_attention_state",
		"astrocyte.explain_focus",
		"astrocyte.get_experiences",
		"astrocyte.report_event",
		"astrocyte.veto",
	} {
		if !names[want] {
			t.Fatalf("missing tool %q in %v", want, names)
		}
	}
}

func TestAstrocyteDispatchGetAttentionState(t *testing.T) {
	s := NewAstrocyteServer(stubBackend{})
	input := reqLine(1, "tools/call", `{"name":"astrocyte.get_attention_state","arguments":{}}`)
	lines := serveLines(t, s, input)
	result := lines[0]["result"].(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	var got app.GetAttentionStateResult
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("result not a GetAttentionStateResult: %v", err)
	}
	if len(got.Topics) != 1 || got.Topics[0].Topic != "Agent Swarm" {
		t.Fatalf("unexpected topics: %v", got.Topics)
	}
}
