package attention_engine_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// M0 acceptance: existing watcher events must map losslessly into the new Event
// (§4.1). The normalised fields are projections for queryability; the original
// record is preserved verbatim in Event.raw so the evidence chain never loses
// a field. The production adapters are W1's; these fixture mappings pin the
// convention the adapters must honour.

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// assertLossless checks raw preserves the original and the normalised fields
// match the projection.
func assertLossless(t *testing.T, event domain.Event, original any, wantSource, wantSession, wantKind, wantScope string, wantTS int64) {
	t.Helper()
	if string(event.Raw) != mustJSON(t, original) {
		t.Errorf("raw does not preserve original: got %s want %s", event.Raw, mustJSON(t, original))
	}
	if event.Source != wantSource || event.Session != wantSession || event.Kind != wantKind || event.Scope != wantScope || event.TS != wantTS {
		t.Errorf("projection mismatch: %+v", event)
	}
}

func TestOutboxEventMapsLosslessly(t *testing.T) {
	at := time.Date(2026, 10, 11, 1, 2, 3, 456000000, time.UTC)
	original := app.OutboxEvent{
		ID:               "o-1",
		Type:             "material_imported",
		AggregateID:      "m-1",
		AggregateVersion: 3,
		OccurredAt:       at,
		CorrelationID:    "corr-1",
		CausationID:      "cause-0",
		Payload:          json.RawMessage(`{"locator":"arxiv:2501.12948"}`),
	}
	event := domain.Event{
		Source:  domain.EventSourceManual,
		Session: original.CorrelationID,
		TS:      original.OccurredAt.UnixMilli(),
		Kind:    "collect",
		Scope:   original.AggregateID,
		Raw:     json.RawMessage(mustJSON(t, original)),
	}
	assertLossless(t, event, original, "manual", "corr-1", "collect", "m-1", at.UnixMilli())
}

func TestUsageMapsLosslessly(t *testing.T) {
	at := time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC)
	original := app.Usage{
		ID:         "u-1",
		MaterialID: "m-1",
		ActorID:    "human-local",
		ActorKind:  "human",
		Action:     "read",
		OccurredAt: at,
	}
	event := domain.Event{
		Source:  original.ActorKind,
		Session: original.ActorID,
		TS:      original.OccurredAt.UnixMilli(),
		Kind:    original.Action,
		Scope:   original.MaterialID,
		Raw:     json.RawMessage(mustJSON(t, original)),
	}
	assertLossless(t, event, original, "human", "human-local", "read", "m-1", at.UnixMilli())
}

func TestSourceItemMapsLosslessly(t *testing.T) {
	original := app.SourceItem{
		SourceID:   "src-1",
		ExternalID: "video-42",
		Revision:   2,
		Metadata: app.ListingMetadata{
			ExternalID:  "video-42",
			Locator:     "https://example/video/42",
			Title:       "a collected item",
			PublishedAt: 1700000000000,
		},
		Selected: true,
	}
	event := domain.Event{
		Source:  domain.EventSourceManual,
		Session: original.SourceID,
		TS:      original.Metadata.PublishedAt,
		Kind:    domain.EventKindCollect,
		Scope:   original.ExternalID,
		Raw:     json.RawMessage(mustJSON(t, original)),
	}
	assertLossless(t, event, original, "manual", "src-1", "collect", "video-42", 1700000000000)
}

func TestAgentSessionMapsLosslessly(t *testing.T) {
	// local_agent_sessions stores an opaque data JSON. The projection takes the
	// stable session identity and project scope; raw preserves the whole record.
	original := map[string]any{
		"id":         "session-1",
		"project_id": "astrocyte",
		"cli":        "codex",
		"data":       map[string]any{"mode": "native", "stop_confirmed": true},
	}
	event := domain.Event{
		Source:  domain.EventSourceCodex,
		Session: "session-1",
		TS:      0,
		Kind:    domain.EventKindAgentTask,
		Scope:   "astrocyte",
		Raw:     json.RawMessage(mustJSON(t, original)),
	}
	assertLossless(t, event, original, "codex", "session-1", "agent_task", "astrocyte", 0)

	// The raw payload must decode back to the exact original record.
	var back map[string]any
	if err := json.Unmarshal(event.Raw, &back); err != nil {
		t.Fatalf("raw no longer decodes: %v", err)
	}
	if back["id"] != "session-1" || back["project_id"] != "astrocyte" || back["cli"] != "codex" {
		t.Errorf("session raw round-trip lost fields: %v", back)
	}
}
