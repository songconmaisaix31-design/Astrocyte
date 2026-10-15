package events

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectedFileUsesProductionToolBlockMapping(t *testing.T) {
	root, project := t.TempDir(), t.TempDir()
	header, _ := json.Marshal(map[string]any{"type": "user", "sessionId": "fixture", "cwd": project})
	raw := []byte(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit"},{"type":"tool_use","name":"Grep"}]}}`)
	path := filepath.Join(root, "fixture.jsonl")
	if err := os.WriteFile(path, append(append(header, '\n'), append(raw, '\n')...), 0600); err != nil {
		t.Fatal(err)
	}
	rows, partial, err := ReplayNativeFile(context.Background(), "claude_code", root, path, project, 128, 1)
	if err != nil || partial || len(rows) != 3 {
		t.Fatal("selected file block population changed")
	}
	for i, kind := range []string{"edit_file", "search"} {
		row := rows[i+1]
		direct, err := MapNativeToolRecord("claude_code", row.OriginKey, project, raw, 1, NativeAssociation{Session: "fixture", Scope: project}, i)
		if err != nil || row.Event == nil || direct.Event == nil || row.Event.Kind != kind || row.Event.Kind != direct.Event.Kind || !bytes.Equal(row.RawBytes, raw) || row.OriginKey != direct.OriginKey {
			t.Fatal("selected file/direct population diverged")
		}
	}
	if rows[1].OriginKey == rows[2].OriginKey {
		t.Fatal("tool block identity collapsed")
	}
}

func TestGitAdditionalKeywordMatchesDoNotDuplicatePrimary(t *testing.T) {
	a := ClassifyGitTopics("/Morphogenesis", "Astrocyte MCP agent memory changes")
	if a.PrimaryTopic != "Morphogenesis" || len(a.KeywordTopics) != 3 || a.KeywordTopics[0] != "Astrocyte/Attention Engine" || a.KeywordTopics[1] != "Agent Memory" || a.KeywordTopics[2] != "MCP" {
		t.Fatal("fixed git keyword priority changed")
	}
	b := ClassifyGitTopics("/Astrocyte", "Astrocyte only")
	if b.PrimaryTopic != "Astrocyte/Attention Engine" || len(b.KeywordTopics) != 0 {
		t.Fatal("primary commit topic duplicated")
	}
}

func TestWrapperStaticSemanticsAndRejection(t *testing.T) {
	cases := []struct{ code, kind string }{
		{`const r = await tools.apply_patch("*** Begin Patch\n*** End Patch"); text(r);`, "edit_file"},
		{"await tools.apply_patch(`literal patch`)", "edit_file"},
		{`await tools.exec_command({cmd:"rg term",workdir:"/fixture",yield_time_ms:1000});`, "search"},
		{`const a=await tools.exec_command({cmd:'grep term'});const b=await tools.exec_command({cmd:'rg next'});text(a);`, "search"},
		{`// await tools.apply_patch("fake")
text("await tools.apply_patch('fake')");`, "unknown"},
		{`if(false){await tools.apply_patch("patch")}`, "unknown"},
		{`false && await tools.apply_patch("patch")`, "unknown"},
		{`false ? await tools.apply_patch("patch") : null`, "unknown"},
		{`const tools=other; await tools.apply_patch("patch")`, "unknown"},
		{`await tools.apply_patch(dynamic)`, "unknown"},
		{"await tools.apply_patch(`patch ${secret}`)", "unknown"},
		{`await tools.exec_command({cmd:"git commit -m failed"})`, "unknown"},
		{`await tools.exec_command({cmd:"rg x; destructive"})`, "unknown"},
		{`await tools.apply_patch("patch");await tools.exec_command({cmd:"rg x"})`, "unknown"},
		{`await tools.write_stdin({session_id:1})`, "unknown"},
		{`tools.apply_patch("patch")`, "unknown"},
		{`unknown.tools.apply_patch("patch")`, "unknown"},
	}
	for i, c := range cases {
		if got := wrapperKind(c.code); got != c.kind {
			t.Fatalf("case %d mapped %s", i, got)
		}
	}
}

func TestNativeToolBlockIdentityAndUnknownUsage(t *testing.T) {
	raw := []byte(`{"type":"assistant","sessionId":"fixture","cwd":"/Astrocyte","message":{"usage":{"input_tokens":10,"output_tokens":2},"content":[{"type":"tool_use","name":"Edit","input":{}},{"type":"tool_use","name":"Grep","input":{}},{"type":"thinking","thinking":"pdf"}]}}`)
	for block, kind := range []string{"edit_file", "search"} {
		row, err := MapNativeToolRecord("claude_code", "immutable-origin", "/Astrocyte", raw, 1, NativeAssociation{}, block)
		if err != nil || row.Event == nil || row.Event.Kind != kind || row.OriginKey != "immutable-origin" || !bytes.Equal(row.RawBytes, raw) || !bytes.Equal(row.Event.Raw, raw) || row.Observation.Tokens != nil || row.Observation.Depth != "unknown" || row.Observation.Outcome != "unknown" {
			t.Fatal("block altered raw/identity/semantics")
		}
	}
	if _, err := MapNativeToolRecord("claude_code", "id", "/Astrocyte", raw, 1, NativeAssociation{}, 2); err == nil {
		t.Fatal("thinking counted as tool")
	}
	user, err := nativeRecord("claude_code", []byte(`{"type":"user","message":{"content":[{"type":"tool_use","name":"Edit"}]}}`))
	if err != nil || user.kind != "unknown" {
		t.Fatal("user text/block inflated cognitive events")
	}
}

func TestSignalStrengthTimesCanonicalWeightOnce(t *testing.T) {
	cases := []struct {
		kind, outcome, depth string
		weight               float64
	}{
		{"read_paper", "unknown", "deep", 1}, {"edit_file", "unknown", "unknown", .6}, {"bookmark", "unknown", "unknown", .3}, {"search", "unknown", "unknown", .6}, {"user_approve", "unknown", "unknown", 1.5}, {"user_veto", "unknown", "unknown", -.8}, {"agent_task", "success", "unknown", .8}, {"agent_task", "failure", "unknown", -.1},
	}
	for _, c := range cases {
		s, err := domain.NormalizeSignal("fixture", domain.Event{Kind: c.kind, Topic: "fixture"}, c.outcome, c.depth, 1)
		if err != nil || s == nil || s.Strength != 1 || s.Weight != c.weight || s.Strength*s.Weight != c.weight {
			t.Fatalf("weight/strength contract changed: %s", c.kind)
		}
	}
}
