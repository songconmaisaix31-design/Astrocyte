package events

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestCanonicalNativeKindsAndUnknownSemantics(t *testing.T) {
	cases := []struct{ source, raw, kind string }{
		{"claude_code", `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Write"}]}}`, "edit_file"},
		{"claude_code", `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"WebSearch"}]}}`, "search"},
		{"claude_code", `{"type":"assistant","message":{"content":[{"type":"text","text":"arxiv.org/abs/1234.5678"}]}}`, "read_paper"},
		{"claude_code", `{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"pdf"}]}}`, "unknown"},
		{"claude_code", `{"type":"user","message":{"content":"pdf"}}`, "unknown"},
		{"codex", `{"type":"response_item","payload":{"type":"custom_tool_call","name":"functions.apply_patch","input":"*** Begin Patch"}}`, "edit_file"},
		{"codex", `{"type":"response_item","payload":{"type":"function_call","name":"tools.exec_command","arguments":"{\"cmd\":\"rg paper\"}"}}`, "search"},
		{"codex", `{"type":"response_item","payload":{"type":"function_call","name":"functions.exec_command","arguments":"{\"cmd\":\"echo search edit\"}"}}`, "unknown"},
		{"codex", `{"type":"response_item","payload":{"type":"function_call","name":"functions.exec_command","arguments":"{\"cmd\":\"git commit -m failed\"}"}}`, "unknown"},
		{"codex", `{"type":"response_item","payload":{"type":"function_call","name":"unknown.apply_patch"}}`, "unknown"},
		{"codex", `{"type":"response_item","payload":{"type":"function_call","name":"functions.exec","arguments":"arbitrary javascript"}}`, "unknown"},
		{"codex", `{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"paper.pdf"}]}}`, "read_paper"},
		{"codex", `{"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"pdf"}]}}`, "unknown"},
		{"codex", `{"type":"event_msg","payload":{"type":"agent_message","message":"PDF"}}`, "read_paper"},
		{"codex", `{"type":"event_msg","payload":{"type":"task_complete","message":"success pdf"}}`, "unknown"},
		{"codex", `{"type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"total_tokens":10}}}}`, "unknown"},
	}
	for i, c := range cases {
		o, err := nativeRecord(c.source, []byte(c.raw))
		if err != nil || o.kind != c.kind || o.depth != "unknown" || o.outcome != "unknown" {
			t.Fatalf("case %d changed semantics", i)
		}
	}
}

func TestSeedMatchingPriorityAndUnassigned(t *testing.T) {
	labels := SeedTopics()
	if len(labels) != 6 || labels[5] != "McPet" {
		t.Fatal("fixed seed count changed")
	}
	labels[0] = "mutated"
	if SeedTopics()[0] != "Morphogenesis" {
		t.Fatal("caller mutated canonical seeds")
	}
	cases := []struct{ scope, text, topic string }{
		{"C:/repo/Morphogenesis/astrocyte", "MCP", "Morphogenesis"},
		{"C:/repo/Astrocyte", "Morphogenesis", "Astrocyte/Attention Engine"},
		{"C:/Users/fixture/.codex/sessions", "", ""},
		{"C:/Users/fixture/.claude/projects", "", ""},
		{"/unrelated", "Agent Memory and Agent Swarm", "Agent Swarm"},
		{"/unrelated", "MCP transport", "MCP"},
		{"/unrelated", "attention_engine", "Astrocyte/Attention Engine"},
		{"/unrelated", "memory pdf general chat", ""},
	}
	for _, c := range cases {
		match := MatchSeedTopic(c.scope, c.text)
		if match.Topic != c.topic || (match.Bucket == "unassigned") != (c.topic == "") {
			t.Fatal("seed matching or priority changed")
		}
	}
}

func TestAssociationAndRawOriginPreservedAcrossNativeLayouts(t *testing.T) {
	for _, layout := range []string{"claude-main", "claude-subagent", "codex-legacy", "codex-dated"} {
		t.Run(layout, func(t *testing.T) {
			root := t.TempDir()
			project := filepath.Join(t.TempDir(), "Astrocyte")
			if err := os.Mkdir(project, 0700); err != nil {
				t.Fatal(err)
			}
			source := "codex"
			relative := "session.jsonl"
			header := map[string]any{"type": "session_meta", "payload": map[string]string{"id": "fixture-session", "cwd": project}}
			raw := []byte(`{"type":"response_item","timestamp":"2026-10-01T00:00:00Z","payload":{"type":"custom_tool_call","name":"tools.apply_patch","input":"LOCAL_RAW_SENTINEL"}}`)
			if strings.HasPrefix(layout, "claude") {
				source = "claude_code"
				header = map[string]any{"type": "user", "sessionId": "fixture-session", "cwd": project}
				raw = []byte(`{"type":"assistant","timestamp":"2026-10-01T00:00:00Z","message":{"content":[{"type":"tool_use","name":"Edit","input":{"content":"LOCAL_RAW_SENTINEL"}}]}}`)
			}
			if layout == "claude-subagent" {
				relative = "project/session/subagents/agent.jsonl"
			}
			if layout == "codex-dated" {
				relative = "2026/10/01/session.jsonl"
			}
			path := filepath.Join(root, relative)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			h, _ := json.Marshal(header)
			if err := os.WriteFile(path, append(append(h, '\n'), append(raw, '\n')...), 0600); err != nil {
				t.Fatal(err)
			}
			rows, partial, err := ReplayNativeFile(context.Background(), source, root, path, project, 128, 1)
			if err != nil || partial || len(rows) != 2 || rows[1].Event == nil {
				t.Fatal("layout read failed")
			}
			a, err := ObserveNativeAssociation(source, project, NativeAssociation{}, h)
			if err != nil {
				t.Fatal(err)
			}
			direct, err := MapAssociatedNativeRecord(source, rows[1].OriginKey, project, raw, 1, a)
			if err != nil || direct.OriginKey != rows[1].OriginKey || !bytes.Equal(direct.RawBytes, raw) || !bytes.Equal(direct.Event.Raw, raw) || direct.Event.Topic != "Astrocyte/Attention Engine" || direct.Observation.Session != "fixture-session" || direct.Observation.Tokens != nil || direct.Event.Kind != rows[1].Event.Kind || direct.Event.Topic != rows[1].Event.Topic {
				t.Fatal("direct/file mappings diverged or source mutated")
			}
		})
	}
}

func TestAssociationRejectsChangedSessionAndProjectWithoutInventingMissing(t *testing.T) {
	previous := NativeAssociation{Session: "observed", Scope: "/approved"}
	for _, raw := range []string{`{"sessionId":"different"}`, `{"cwd":"/outside"}`} {
		if _, err := ObserveNativeAssociation("claude_code", "/approved", previous, []byte(raw)); err == nil {
			t.Fatal("foreign association accepted")
		}
	}
	row, err := MapNativeRecord("codex", "unchanged", "/approved", []byte(`{"type":"response_item","payload":{"type":"function_call","name":"apply_patch"}}`), 0)
	if err != nil || row.Observation.Scope != "" || row.Observation.Session != "" || row.Event.Topic != "" || ClassifyNativeReceipt(row).Bucket != "unassigned" {
		t.Fatal("missing observation fabricated")
	}
	_, _, err = readNative(context.Background(), "claude_code", "/approved", strings.NewReader("{\"sessionId\":\"observed\",\"cwd\":\"/approved\"}\n{\"cwd\":\"/outside\"}\n"), 128, false)
	if err == nil {
		t.Fatal("later foreign cwd accepted")
	}
}

func TestGitLogRecordUsesActualCommitIdentityAndUnknownOutcome(t *testing.T) {
	raw := json.RawMessage(`{"commit":"0123456789012345678901234567890123456789","ts":1790784000000,"subject":"fixture subject"}`)
	row, err := MapGitLogRecord("/approved/Morphogenesis", raw, 0)
	if err != nil || row.Event == nil || row.Event.Kind != domain.EventKindCommit || row.Event.Source != domain.EventSourceManual || row.Event.Topic != "Morphogenesis" || row.OriginKey != "git:0123456789012345678901234567890123456789" || !bytes.Equal(row.RawBytes, raw) || row.Observation.Outcome != "unknown" || row.Observation.Tokens != nil {
		t.Fatal("git record identity/unknown semantics changed")
	}
	for _, invalid := range []string{`{"commit":"commit message","ts":1}`, `{"commit":"0123456789012345678901234567890123456789","ts":0}`} {
		if _, err := MapGitLogRecord("/approved", []byte(invalid), 0); err == nil {
			t.Fatal("message or missing ts accepted as commit")
		}
	}
}
