package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func testAstrocyteServer(t *testing.T) *Server {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	decay, confidence := .1, .5
	dynamics := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence})
	// 人工创建主题走实际生产接口，不补写来源事件或关注分值。
	if _, err = dynamics.SeedTopic(ctx, app.Principal{ID: "local-human", Kind: "human"}, "Agent Swarm", time.Now().UnixMilli()); err != nil {
		t.Fatal(err)
	}
	events := sqlite.NewEventRepository(db)
	binding := app.MCPBinding{Principal: app.Principal{ID: "test-agent", Kind: "agent"}, Scope: "astrocyte", Session: "test"}
	backend, err := app.NewEngineMCPBackend(binding, dynamics,
		app.NewExperienceService(sqlite.NewExperienceRepository(db), events, app.ExperienceOptions{ProjectID: "astrocyte"}),
		app.NewEventService(events, .01), app.MCPBackendOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewAstrocyteServer(backend, binding)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAstrocyteRegistersFiveTools(t *testing.T) {
	s := testAstrocyteServer(t)
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
	s := testAstrocyteServer(t)
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
