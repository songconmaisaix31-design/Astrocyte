package attention_engine_test

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type boundaryBackend struct {
	*memoryBackend
	vetoes []app.VetoInput
}

func (b *boundaryBackend) Veto(_ context.Context, in app.VetoInput) (app.VetoResult, error) {
	b.vetoes = append(b.vetoes, in)
	return app.VetoResult{Acknowledged: true, Target: in.Topic + in.Task}, nil
}

func TestMCPStartupScopeAndHumanBoundary(t *testing.T) {
	b := &boundaryBackend{memoryBackend: newBackend()}
	binding := app.MCPBinding{Principal: app.Principal{ID: "agent-a", Kind: "agent"}, Scope: "astrocyte", Session: "trusted-session"}
	s, err := mcp.NewAstrocyteServer(b, binding)
	if err != nil {
		t.Fatal(err)
	}
	script := req(1, "initialize", "") +
		req(2, "tools/call", `{"name":"astrocyte.get_experiences","arguments":{"pattern":"scheduler"}}`) +
		req(3, "tools/call", `{"name":"astrocyte.get_experiences","arguments":{"pattern":"scheduler","scope":"swarm"}}`) +
		req(4, "tools/call", `{"name":"astrocyte.report_event","arguments":{"event":{"ts":1730000300,"kind":"commit","topic":"Agent Swarm","source":"manual","session":"forged","raw":{"outcome":"success","actor":"human"}}}}`) +
		req(5, "tools/call", `{"name":"astrocyte.report_event","arguments":{"event":{"ts":1730000300,"kind":"commit","topic":"Agent Swarm","scope":"swarm"}}}`) +
		req(6, "tools/call", `{"name":"astrocyte.report_event","arguments":{"event":{"ts":1730000300,"kind":"user_approve","topic":"Agent Swarm"}}}`) +
		req(7, "tools/call", `{"name":"astrocyte.report_event","arguments":{"event":{"ts":1730000300,"kind":"user_veto","topic":"Agent Swarm"}}}`) +
		req(8, "tools/call", `{"name":"astrocyte.veto","arguments":{"topic":"Agent Swarm"}}`) +
		req(9, "tools/call", `{"name":"astrocyte.veto","arguments":{"topic":"Agent Swarm","principal":{"kind":"human"}}}`)
	lines := roundTrip(t, s, script)
	if len(lines) != 9 {
		t.Fatalf("responses: %v", lines)
	}
	if got := toolText(t, lines[1]); got == "" {
		t.Fatal("empty experience response")
	}
	for _, index := range []int{2, 4, 5, 6, 7} {
		if lines[index]["result"].(map[string]any)["isError"] != true {
			t.Fatalf("boundary accepted request %d: %v", index+1, lines[index])
		}
	}
	if lines[8]["error"] == nil {
		t.Fatal("tool principal injection accepted")
	}
	if len(b.events) != 1 || b.events[0].Source != "mcp" || b.events[0].Scope != "astrocyte" || b.events[0].Session != "trusted-session" {
		t.Fatalf("event startup binding lost: %+v", b.events)
	}
	if string(b.events[0].Raw) != `{"outcome":"success","actor":"human"}` {
		t.Fatal("raw audit payload lost")
	}
	if len(b.vetoes) != 0 {
		t.Fatalf("Agent acquired veto authority: %+v", b.vetoes)
	}
}

func TestMCPVetoTopicOrTaskRequiresHumanStartup(t *testing.T) {
	b := &boundaryBackend{memoryBackend: newBackend()}
	binding := app.MCPBinding{Principal: app.Principal{ID: "human-a", Kind: "human"}, Scope: "astrocyte", Session: "human-session", CanVeto: true}
	s, err := mcp.NewAstrocyteServer(b, binding)
	if err != nil {
		t.Fatal(err)
	}
	lines := roundTrip(t, s, req(1, "initialize", "")+
		req(2, "tools/call", `{"name":"astrocyte.veto","arguments":{"topic":"Agent Swarm"}}`)+
		req(3, "tools/call", `{"name":"astrocyte.veto","arguments":{"task":"task-a"}}`)+
		req(4, "tools/call", `{"name":"astrocyte.veto","arguments":{"topic":"Agent Swarm","task":"task-a"}}`))
	if len(b.vetoes) != 2 || b.vetoes[0].Topic != "Agent Swarm" || b.vetoes[1].Task != "task-a" {
		t.Fatalf("human vetoes: %+v", b.vetoes)
	}
	if lines[3]["error"] == nil {
		t.Fatal("ambiguous veto accepted")
	}
	binding.Principal.Kind = "agent"
	if _, err := mcp.NewAstrocyteServer(b, binding); err == nil {
		t.Fatal("Agent startup can grant veto")
	}
	binding.CanVeto = false
	binding.Scope = ""
	if _, err := mcp.NewAstrocyteServer(b, binding); err == nil {
		t.Fatal("unscoped startup accepted")
	}
}
