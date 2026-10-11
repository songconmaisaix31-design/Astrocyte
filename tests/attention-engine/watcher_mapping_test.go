package attention_engine_test

import (
	"encoding/json"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"testing"
)

// Receipt contract fixtures. Production source mapping and replay belong to W1;
// a session stop or import job does not establish an Agent task outcome.
func TestWatcherUnknownSemanticsRemainPending(t *testing.T) {
	for _, source := range []struct{ kind, raw string }{
		{"local_agent_session", `{"id":"session-1","project_id":"astrocyte","cli":"codex","data":{"stop_confirmed":true}}`},
		{"attention.job_failed", `{"job_id":"import-1","error":"timeout"}`},
		{"material_usage", `{"action":"read","material_id":"m-1"}`},
		{"attention_source_item", `{"selected":false,"published_at":1700000000000}`},
	} {
		r := app.StoredEvent{ID: source.kind, OriginKey: "fixture:" + source.kind, RawBytes: []byte(source.raw), Observation: app.Observation{SourceKind: source.kind, Outcome: "unknown", Depth: "unknown"}}
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		var back app.StoredEvent
		if err = json.Unmarshal(b, &back); err != nil {
			t.Fatal(err)
		}
		if back.Event != nil || back.Observation.Tokens != nil || back.Observation.Outcome != "unknown" || string(back.RawBytes) != source.raw {
			t.Fatalf("pending receipt changed: %+v", back)
		}
	}
}
