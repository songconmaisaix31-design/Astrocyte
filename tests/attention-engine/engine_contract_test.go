package attention_engine_test

import (
	"encoding/json"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// M0 schema tests: the §4 contract types must serialise losslessly and the §5.2
// weight table must be exactly the hardcoded first version. No engine dynamics
// are exercised here; this only proves the contract locks the shapes.

func TestSignalWeightTable(t *testing.T) {
	cases := map[domain.SignalClass]float64{
		domain.SignalImplicitDepth:    1.0,
		domain.SignalImplicitRegular:  0.6,
		domain.SignalImplicitLight:    0.3,
		domain.SignalExplicitPositive: 1.5,
		domain.SignalExplicitNegative: -0.8,
		domain.SignalFeedbackSuccess:  0.8,
		domain.SignalFeedbackFailure:  -0.1,
	}
	for class, want := range cases {
		if got := domain.SignalWeight(class); got != want {
			t.Errorf("SignalWeight(%s) = %v, want %v", class, got, want)
		}
	}
	if got := domain.SignalWeight(domain.SignalClass("unknown")); got != 0 {
		t.Errorf("unknown signal class must carry zero weight, got %v", got)
	}
}

// roundTrip marshals v to JSON and back, returning the decoded map so every
// field can be asserted by name.
func roundTrip(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestEventJSONShape(t *testing.T) {
	e := domain.Event{
		Source:  domain.EventSourceCodex,
		Session: "sess-1",
		TS:      1700000000123,
		Kind:    domain.EventKindEditFile,
		Tokens:  512,
		Scope:   "attention-engine-w0-core",
		Topic:   "cognitive-core",
		Raw:     json.RawMessage(`{"file":"engine_contract.go","lines":120}`),
	}
	m := roundTrip(t, e)
	for k, want := range map[string]any{
		"source":  "codex",
		"session": "sess-1",
		"ts":      float64(1700000000123),
		"kind":    "edit_file",
		"tokens":  float64(512),
		"scope":   "attention-engine-w0-core",
		"topic":   "cognitive-core",
	} {
		if got, ok := m[k]; !ok || got != want {
			t.Errorf("Event.%s = %v, want %v", k, got, want)
		}
	}
	// raw must survive verbatim as the original JSON object.
	raw, ok := m["raw"].(map[string]any)
	if !ok || raw["file"] != "engine_contract.go" {
		t.Errorf("Event.raw lost original payload: %v", m["raw"])
	}
}

func TestAttentionNodeJSONShape(t *testing.T) {
	n := domain.AttentionNode{
		ID:              "n-1",
		Topic:           "agent-swarm",
		Attention:       0.73,
		Momentum:        0.18,
		Confidence:      0.6,
		Decay:           0.01,
		Evidence:        []domain.EvidenceRef{{EventID: "e-1", WeightAtUse: 0.6, TS: 1}},
		RelatedProjects: []string{"astrocyte"},
		RelatedAgents:   []string{"codex"},
		RelatedTopics:   map[string]float64{"morphogenesis": 0.8},
		LastActive:      1700000000000,
		EmotionalWeight: 1.5,
	}
	m := roundTrip(t, n)
	if m["attention"] != float64(0.73) || m["momentum"] != float64(0.18) {
		t.Errorf("AttentionNode numeric fields corrupted: %v", m)
	}
	if m["related_topics"].(map[string]any)["morphogenesis"] != float64(0.8) {
		t.Error("AttentionNode related_topics lost")
	}
	if ev := m["evidence"].([]any)[0].(map[string]any); ev["event_id"] != "e-1" || ev["weight_at_use"] != float64(0.6) {
		t.Errorf("AttentionNode evidence lost: %v", ev)
	}
}

func TestExperienceJSONShape(t *testing.T) {
	e := domain.Experience{
		ID:        "x-1",
		Pattern:   "centralized scheduler",
		Verdict:   domain.ExperienceVerdictAvoid,
		Condition: ">100 agents",
		Evidence:  []domain.EventRef{{EventID: "e-9", TS: 9}},
		Status:    domain.ExperienceStatusQuarantine,
	}
	m := roundTrip(t, e)
	if m["verdict"] != "avoid" || m["status"] != "quarantine" {
		t.Errorf("Experience verdict/status lost: %v", m)
	}
}

func TestKnowledgeEntryJSONShape(t *testing.T) {
	k := domain.KnowledgeEntry{
		ID:           "k-1",
		Topic:        "attention-engine",
		Fluid:        "evolving summary",
		Crystallized: json.RawMessage(`{"fact":"true","at":"2026-10-11"}`),
		Sources:      []domain.EventRef{{EventID: "e-1", TS: 1}},
		Status:       domain.KnowledgeInbox,
	}
	m := roundTrip(t, k)
	if m["status"] != "inbox" || m["fluid"] != "evolving summary" {
		t.Errorf("KnowledgeEntry fields lost: %v", m)
	}
}

func TestConflictJSONShape(t *testing.T) {
	c := domain.Conflict{
		ID:       "c-1",
		Topic:    "scheduler",
		Kind:     domain.ConflictKindKnowledge,
		LeftRef:  "k-1",
		RightRef: "k-2",
		Detail:   "contradictory capacity facts",
		Status:   domain.ConflictStatusPending,
	}
	m := roundTrip(t, c)
	if m["status"] != "pending" || m["kind"] != "knowledge_entry" {
		t.Errorf("Conflict fields lost: %v", m)
	}
}

// The EvidenceRef contract (§4.5) is a view, not an independent object: it must
// decode from the exact triple the engine writes and reads.
func TestEvidenceRefDecode(t *testing.T) {
	var ref domain.EvidenceRef
	if err := json.Unmarshal([]byte(`{"event_id":"e-1","weight_at_use":1.0,"ts":42}`), &ref); err != nil {
		t.Fatalf("decode EvidenceRef: %v", err)
	}
	if ref.EventID != "e-1" || ref.WeightAtUse != 1.0 || ref.TS != 42 {
		t.Errorf("EvidenceRef decode mismatch: %+v", ref)
	}
}
