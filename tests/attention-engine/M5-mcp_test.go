package attention_engine_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// memoryBackend is a test double standing in for the W2/W5 read-side use cases.
// It exists only to exercise the MCP transport and the §7 DTO contract until M0
// publishes the real engine and experience read-side; real filtering and
// evidence assembly live behind those use cases, not here.
type memoryBackend struct {
	topics      []app.FocusItem
	focus       map[string]app.ExplainFocusResult
	experiences []app.ExperienceItem
	events      []app.ReportEventInput
}

func (b *memoryBackend) GetAttentionState(context.Context) (app.GetAttentionStateResult, error) {
	return app.GetAttentionStateResult{Topics: append([]app.FocusItem(nil), b.topics...)}, nil
}

func (b *memoryBackend) ExplainFocus(_ context.Context, in app.ExplainFocusInput) (app.ExplainFocusResult, error) {
	r, ok := b.focus[in.Topic]
	if !ok {
		return app.ExplainFocusResult{}, fmt.Errorf("no focus for topic %q", in.Topic)
	}
	return r, nil
}

func (b *memoryBackend) GetExperiences(_ context.Context, in app.GetExperiencesInput) (app.GetExperiencesResult, error) {
	var out []app.ExperienceItem
	for _, e := range b.experiences {
		if in.Scope != "" && e.Scope != in.Scope {
			continue
		}
		if !strings.Contains(e.Pattern, in.Pattern) {
			continue
		}
		out = append(out, e)
	}
	return app.GetExperiencesResult{Experiences: out}, nil
}

func (b *memoryBackend) ReportEvent(_ context.Context, in app.ReportEventInput) (app.ReportEventResult, error) {
	b.events = append(b.events, in)
	return app.ReportEventResult{Accepted: true, EventID: fmt.Sprintf("evt-%d", len(b.events))}, nil
}

func (b *memoryBackend) Veto(_ context.Context, in app.VetoInput) (app.VetoResult, error) {
	return app.VetoResult{Acknowledged: true, Target: in.Target}, nil
}

func roundTrip(t *testing.T, s *mcp.Server, script string) []map[string]any {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	or, ow, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		err := s.Serve(context.Background(), pr, ow)
		ow.Close()
		done <- err
	}()
	if _, err := pw.WriteString(script); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	if err := <-done; err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var lines []map[string]any
	scanner := bufio.NewScanner(or)
	for scanner.Scan() {
		var m map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			t.Fatalf("bad response %q: %v", scanner.Text(), err)
		}
		lines = append(lines, m)
	}
	or.Close()
	return lines
}

func req(id int, method, params string) string {
	if params == "" {
		params = "{}"
		if method == "initialize" {
			params = `{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}`
		}
	}
	return fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"id\":%d,\"method\":%q,\"params\":%s}\n", id, method, params)
}

func toolText(t *testing.T, resp map[string]any) string {
	t.Helper()
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result: %v", resp)
	}
	content := result["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func newBackend() *memoryBackend {
	return &memoryBackend{
		topics: []app.FocusItem{
			{Topic: "Agent Swarm", Attention: 0.9, Momentum: 0.2},
			{Topic: "sqlite-vec", Attention: 0.6, Momentum: -0.1},
			{Topic: "sleep consolidation", Attention: 0.3, Momentum: 0.4},
		},
		focus: map[string]app.ExplainFocusResult{
			"Agent Swarm": {
				Topic:     "Agent Swarm",
				Attention: 0.9,
				Momentum:  0.2,
				Evidence: []app.EvidenceItem{
					{EventID: "evt-1", Kind: "read_paper", Source: "arxiv", Weight: 1.0, TS: 1730000001},
					{EventID: "evt-2", Kind: "edit_file", Source: "claude_code", Weight: 0.6, TS: 1730000100},
					{EventID: "evt-3", Kind: "user_approve", Source: "manual", Weight: 1.5, TS: 1730000200},
				},
				Related: []app.RelatedTopic{
					{Topic: "Morphogenesis", Strength: 0.7},
					{Topic: "MCP", Strength: 0.5},
				},
				Projects: []string{"astrocyte", "swarm"},
			},
		},
		experiences: []app.ExperienceItem{
			{ID: "exp-1", Pattern: "centralized scheduler", Verdict: "avoid", Condition: ">100 agents", Scope: "astrocyte", Evidence: []string{"evt-9"}},
			{ID: "exp-2", Pattern: "streaming scheduler", Verdict: "prefer", Condition: ">100 agents", Scope: "swarm", Evidence: []string{"evt-10"}},
		},
	}
}

func TestMCPFiveOperationsOverStdio(t *testing.T) {
	b := newBackend()
	s := mcp.NewAstrocyteServer(b)

	script := req(1, "initialize", "") +
		req(2, "tools/list", "") +
		req(3, "tools/call", `{"name":"astrocyte.get_attention_state","arguments":{}}`) +
		req(4, "tools/call", `{"name":"astrocyte.explain_focus","arguments":{"topic":"Agent Swarm"}}`) +
		req(5, "tools/call", `{"name":"astrocyte.get_experiences","arguments":{"pattern":"scheduler","scope":"astrocyte"}}`) +
		req(6, "tools/call", `{"name":"astrocyte.get_experiences","arguments":{"pattern":"scheduler"}}`) +
		req(7, "tools/call", `{"name":"astrocyte.report_event","arguments":{"source":"codex","session":"s1","ts":1730000300,"kind":"commit","tokens":40,"scope":"astrocyte","topic":"Agent Swarm"}}`) +
		req(8, "tools/call", `{"name":"astrocyte.veto","arguments":{"target":"Agent Swarm","reason":"out of scope"}}`)

	lines := roundTrip(t, s, script)
	if len(lines) != 8 {
		t.Fatalf("expected 8 responses, got %d", len(lines))
	}

	// initialize
	if lines[0]["result"].(map[string]any)["protocolVersion"] == nil {
		t.Fatalf("missing protocolVersion: %v", lines[0])
	}

	// tools/list exposes exactly the five §7 tools.
	tools := lines[1]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}

	// get_attention_state returns focus ordered by attention.
	var state app.GetAttentionStateResult
	if err := json.Unmarshal([]byte(toolText(t, lines[2])), &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Topics) != 3 || state.Topics[0].Topic != "Agent Swarm" {
		t.Fatalf("wrong focus ranking: %v", state.Topics)
	}

	// explain_focus returns a complete evidence chain with timestamps.
	var explain app.ExplainFocusResult
	if err := json.Unmarshal([]byte(toolText(t, lines[3])), &explain); err != nil {
		t.Fatal(err)
	}
	if len(explain.Evidence) != 3 {
		t.Fatalf("expected 3 evidence items, got %d", len(explain.Evidence))
	}
	for _, e := range explain.Evidence {
		if e.EventID == "" || e.Kind == "" || e.TS == 0 {
			t.Fatalf("evidence item missing id/kind/timestamp: %+v", e)
		}
	}
	if len(explain.Related) != 2 || len(explain.Projects) != 2 {
		t.Fatalf("related nodes or projects missing: %+v", explain)
	}

	// get_experiences with a scope returns only that scope.
	var scoped app.GetExperiencesResult
	if err := json.Unmarshal([]byte(toolText(t, lines[4])), &scoped); err != nil {
		t.Fatal(err)
	}
	if len(scoped.Experiences) != 1 || scoped.Experiences[0].Scope != "astrocyte" {
		t.Fatalf("scope filter wrong: %+v", scoped.Experiences)
	}

	// get_experiences without a scope returns every matching pattern.
	var all app.GetExperiencesResult
	if err := json.Unmarshal([]byte(toolText(t, lines[5])), &all); err != nil {
		t.Fatal(err)
	}
	if len(all.Experiences) != 2 {
		t.Fatalf("expected 2 experiences across scopes, got %d", len(all.Experiences))
	}

	// report_event is accepted with an id.
	var reported app.ReportEventResult
	if err := json.Unmarshal([]byte(toolText(t, lines[6])), &reported); err != nil {
		t.Fatal(err)
	}
	if !reported.Accepted || reported.EventID != "evt-1" {
		t.Fatalf("report_event not accepted: %+v", reported)
	}

	// veto is acknowledged.
	var veto app.VetoResult
	if err := json.Unmarshal([]byte(toolText(t, lines[7])), &veto); err != nil {
		t.Fatal(err)
	}
	if !veto.Acknowledged || veto.Target != "Agent Swarm" {
		t.Fatalf("veto not acknowledged: %+v", veto)
	}
}
