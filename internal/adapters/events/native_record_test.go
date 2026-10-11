package events

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestNativeObservationPreservesUnknowns(t *testing.T) {
	o, err := nativeRecord("claude_code", []byte(`{"type":"assistant","sessionId":"s","timestamp":"2026-10-10T00:00:00Z","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"paper.pdf"}}]}}`))
	if err != nil || o.kind != "unknown" || o.tokens != nil || o.outcome != "unknown" || o.depth != "unknown" || o.ts != 1791590400000 {
		t.Fatalf("observation: %+v, %v", o, err)
	}
	o, err = nativeRecord("codex", []byte(`{"type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"git commit -m done\"}"}}`))
	if err != nil || o.kind != "unknown" {
		t.Fatalf("commit command is not committed history: %+v, %v", o, err)
	}
	o, err = nativeRecord("claude_code", []byte(`{"type":"assistant","message":{"usage":{"input_tokens":3,"output_tokens":4},"content":[{"type":"tool_use","name":"Edit"},{"type":"tool_use","name":"Grep"}]}}`))
	if err != nil || o.kind != "unknown" || o.tokens == nil || *o.tokens != 7 {
		t.Fatalf("mixed observations: %+v, %v", o, err)
	}
}

func TestBoundedNativeRecordAccounting(t *testing.T) {
	header := `{"type":"session_meta", "payload":{"id":"actual-selected-session", "cwd":"project"}}`
	record := `{"type":"response_item", "payload":{"type":"function_call", "name":"apply_patch", "arguments":"sensitive input"}}`
	lines, partial, err := readNative(context.Background(), "codex", "project", strings.NewReader(header+"\n"+record+"\n"+record+"\n"), 2, false)
	if err != nil || !partial || len(lines) != 2 || !bytes.Equal(lines[1].raw, []byte(record)) || lines[1].observation.kind != "edit_file" {
		t.Fatalf("bounded replay: %+v %v %v", lines, partial, err)
	}
	_, _, err = readNative(context.Background(), "codex", "other-project", strings.NewReader(header+"\n"), 2, false)
	if err == nil {
		t.Fatal("mismatched project accepted")
	}
}

func TestS1JobsAndDeferredRemainUnknown(t *testing.T) {
	for _, v := range []app.OutboxEvent{
		{Type: "attention.job_succeeded"}, {Type: "attention.job_failed"},
		{Type: "attention.distillation_recorded"},
		{Type: "attention.opportunity_reviewed", Payload: []byte(`{"feedback":"later"}`)},
		{Type: "attention.material_used", Payload: []byte(`{"actor_kind":"agent","action":"read"}`)},
	} {
		o := s1Record(v)
		if o.kind != "unknown" || o.outcome != "unknown" || o.depth != "unknown" || o.tokens != nil {
			t.Fatalf("invented observation: %+v", o)
		}
	}
}
