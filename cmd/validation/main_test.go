package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func synthetic(source, raw string) sample {
	return sample{Source: source, HeaderComplete: true, Project: "/approved", RawBase64: base64.StdEncoding.EncodeToString([]byte(raw))}
}

func TestDiagnosticStageSeparation(t *testing.T) {
	d := &diagnosis{Counts: map[string]int{}, Kinds: map[string]int{}}
	diagnose(synthetic("claude_code", `{"type":"assistant","cwd":"/approved","sessionId":"fixture","timestamp":"2026-10-01T00:00:00Z","message":{"content":[{"type":"tool_use","name":"Edit","input":{"body":"LOCAL_ONLY_SENTINEL"}}]}}`), d)
	diagnose(synthetic("codex", `{"type":"event_msg","timestamp":"2026-10-01T00:00:00Z","payload":{"type":"agent_message","message":"LOCAL_ONLY_SENTINEL"}}`), d)
	for key, want := range map[string]int{"sample_input": 2, "adapter_pass": 2, "mapped_event": 1, "unknown_kind": 1, "empty_topic": 1, "normalizable_with_probe_strength": 1, "topic_gate_blocked": 1, "nil_strength_blocks_pending_service": 1, "unknown_depth": 2, "unknown_outcome": 2, "eligible_current_configuration": 0} {
		if d.Counts[key] != want {
			t.Fatalf("%s count got %d want %d", key, d.Counts[key], want)
		}
	}
	b, _ := json.Marshal(d)
	if strings.Contains(string(b), "LOCAL_ONLY_SENTINEL") || strings.Contains(string(b), "/approved") {
		t.Fatal("derived diagnostic exposed native values")
	}
}

func TestDiagnosticParseAndScopeRejections(t *testing.T) {
	d := &diagnosis{Counts: map[string]int{}, Kinds: map[string]int{}}
	diagnose(synthetic("codex", "invalid"), d)
	diagnose(synthetic("claude_code", `{"cwd":"/outside"}`), d)
	if d.Counts["input_parse_rejected"] != 1 || d.Counts["adapter_project_or_source_rejected"] != 1 || d.Counts["mapped_event"] != 0 {
		t.Fatal("parse and project gates not separated")
	}
}

func TestNormalizationUnknownDepthAndOutcomeRetainsNoSignal(t *testing.T) {
	for _, kind := range []string{domain.EventKindReadPaper, domain.EventKindAgentTask} {
		s, err := domain.NormalizeSignal("fixture", domain.Event{Kind: kind}, "unknown", "unknown", 1)
		if err != nil || s != nil {
			t.Fatal("unknown depth/outcome became signal")
		}
	}
}

type pendingRepo struct {
	app.EventRepository
	records []app.StoredEvent
}
type pendingTx struct{ records []app.StoredEvent }

func (tx *pendingTx) PutEvent(v app.StoredEvent) (app.StoredEvent, error) {
	v.ID = "fixture"
	tx.records = append(tx.records, v)
	return v, nil
}
func (tx *pendingTx) PutSignal(domain.Signal) error { return nil }
func (r *pendingRepo) WithEventTx(_ context.Context, fn func(app.EventTx) error) error {
	tx := &pendingTx{}
	if err := fn(tx); err != nil {
		return err
	}
	r.records = append(r.records, tx.records...)
	return nil
}

func TestActualPendingConfigurationRejectsMappedButRetainsUnknown(t *testing.T) {
	r := &pendingRepo{}
	s := app.NewPendingEventService(r)
	unknown, _ := events.MapNativeRecord("codex", "unknown", "/approved", json.RawMessage(`{"type":"event_msg"}`), 0)
	if _, err := s.Record(context.Background(), []app.StoredEvent{unknown}); err != nil || len(r.records) != 1 {
		t.Fatal("unknown receipt not retained")
	}
	mapped, _ := events.MapNativeRecord("codex", "mapped", "/approved", json.RawMessage(`{"type":"response_item","payload":{"type":"function_call","name":"apply_patch","arguments":"fixture"}}`), 0)
	if _, err := s.Record(context.Background(), []app.StoredEvent{mapped}); err == nil || !strings.Contains(err.Error(), "explicitly configured signal strength") || len(r.records) != 1 {
		t.Fatal("nil strength did not reject and roll back mapped receipt")
	}
}

type nilPolicyStore struct{ app.DynamicsStore }

func (nilPolicyStore) WithDynamicsTx(_ context.Context, fn func(app.DynamicsTx) error) error {
	return fn(nil)
}

func TestActualNilDynamicsBlocksSeedBeforeStoreMutation(t *testing.T) {
	for _, options := range []app.DynamicsOptions{{}, {Decay: new(float64)}} {
		e := app.NewDynamicsEngine(nilPolicyStore{}, options)
		_, err := e.SeedTopic(context.Background(), app.Principal{ID: "human", Kind: "human"}, "fixture", 1)
		if err == nil || !strings.Contains(err.Error(), "remains undecided") {
			t.Fatal("unconfigured engine accepted seed")
		}
	}
}

func TestIncompleteSnapshotCannotProduceSuccessReport(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "sample.jsonl")
	out := filepath.Join(dir, "counts.json")
	b, _ := json.Marshal(synthetic("claude_code", `{"type":"user"}`))
	if err := os.WriteFile(in, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(in, out); err == nil {
		t.Fatal("partial sample accepted")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("partial report created")
	}
}

func TestCompleteSnapshotDerivedOutputIsPrivateAndImmutable(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "sample.jsonl")
	output := filepath.Join(dir, "counts.json")
	var lines []byte
	for _, source := range []string{"claude_code", "codex"} {
		for i := 0; i < 50; i++ {
			v := synthetic(source, `{"type":"LOCAL_ONLY_SENTINEL","message":{"content":"LOCAL_ONLY_SENTINEL"}}`)
			v.Ordinal = i
			b, _ := json.Marshal(v)
			lines = append(lines, append(b, '\n')...)
		}
	}
	if err := os.WriteFile(input, lines, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(input, output); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(output)
	if strings.Contains(string(b), "LOCAL_ONLY_SENTINEL") || strings.Contains(string(b), "raw_base64") || strings.Contains(string(b), "/approved") {
		t.Fatal("native fields exposed")
	}
	if err := run(input, output); err == nil {
		t.Fatal("existing evidence overwritten")
	}
	after, _ := os.ReadFile(output)
	if string(after) != string(b) {
		t.Fatal("first report changed")
	}
	original, _ := os.ReadFile(input)
	if string(original) != string(lines) {
		t.Fatal("sample mutated")
	}
}
