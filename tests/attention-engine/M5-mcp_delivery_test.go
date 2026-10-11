package attention_engine_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Test-only injection at the real W5 transaction boundary. The original W1
// receipt transaction succeeds; W5's actual callback then fails to capture.
type mcpUnavailableExperienceRepository struct{}

func (mcpUnavailableExperienceRepository) WithExperienceTx(context.Context, func(app.ExperienceTx) error) error {
	return errors.New("contract_local injected experience storage unavailable")
}

type mcpDeliveryFixtureRecorder struct {
	service *app.EventService
	mode    string
}

func (r mcpDeliveryFixtureRecorder) Record(ctx context.Context, records []app.StoredEvent) ([]app.StoredEvent, error) {
	if r.mode == "capture" {
		// Public MCP observations remain unverified/unknown. Append a separate,
		// explicit trusted synthetic task failure to exercise the actual batch
		// post-commit callback, never trust a caller's raw outcome assertion.
		input := records[0]
		raw := []byte(`{"fixture":"M5 contract_local trusted failure"}`)
		records = append(records, app.StoredEvent{OriginKey: input.OriginKey + ":contract_local_failure", ReceivedAt: input.ReceivedAt, RawBytes: raw,
			Observation: app.Observation{Source: domain.EventSourceCodex, SourceKind: domain.EventKindAgentTask, Scope: "project-a", Session: "m5-trusted-fixture", TS: mcpFixtureNow, Outcome: "failed", Depth: "unknown", Provenance: "M5 contract_local synthetic trusted failure"},
			Event:       &domain.Event{Source: domain.EventSourceCodex, Kind: domain.EventKindAgentTask, Scope: "project-a", Session: "m5-trusted-fixture", TS: mcpFixtureNow, Topic: "delivery-fixture", Raw: raw}})
	}
	rows, err := r.service.Record(ctx, records)
	if r.mode == "unknown" && err == nil {
		// Simulate losing the acknowledgement after an actual SQLite commit.
		return nil, errors.New("contract_local injected acknowledgement lost")
	}
	return rows, err
}

func mcpProcessRecorder(events *sqlite.EventRepository, experiences *app.ExperienceService, mode string) app.MCPEventRecorder {
	callback := experiences.CaptureCommittedFailure
	if mode == "capture" {
		failing := app.NewExperienceService(mcpUnavailableExperienceRepository{}, events, app.ExperienceOptions{ProjectID: "project-a", Clock: mcpFixtureClock})
		callback = failing.CaptureCommittedFailure
	}
	service := app.NewEventService(events, .01, callback)
	if mode != "" {
		return mcpDeliveryFixtureRecorder{service: service, mode: mode}
	}
	return service
}

func TestMCPRealStdioPostcommitDeliveryAndRestart(t *testing.T) {
	for _, operation := range []string{"report_event", "veto"} {
		for _, mode := range []string{"capture", "unknown"} {
			t.Run(operation+"/"+mode, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "delivery.sqlite")
				db := mcpOpen(t, path)
				db.Close()
				args := `{"event":{"ts":1770681600000,"kind":"agent_task","topic":"delivery-fixture","raw":{"outcome":"failed"}}}`
				if operation == "veto" {
					args = `{"task":"unresolved-task","reason":"contract_local veto"}`
				}
				responses := mcpSubprocess(t, path, "human", mcpCall(2, operation, args), "ASTROCYTE_M5_MCP_DELIVERY_FIXTURE="+mode)
				result := responses[1]["result"].(map[string]any)
				if result["isError"] != true {
					t.Fatalf("postcommit failure must remain a tool error: %v", result)
				}
				// Inspect only text, as an independent client that discards
				// structuredContent for errors would. Both must preserve IDs.
				var text map[string]any
				if err := json.Unmarshal([]byte(toolText(t, responses[1])), &text); err != nil {
					t.Fatalf("error text lacks receipt data: %v; response=%v", err, result)
				}
				if !reflect.DeepEqual(text, result["structuredContent"]) || text["error"] == "" || !strings.Contains(text["recovery"].(string), "do not resubmit") {
					t.Fatalf("client-visible recovery instructions missing: %v", result)
				}
				origin := text["origin_key"].(string)
				db = mcpOpen(t, path)
				repo := sqlite.NewEventRepository(db)
				rows, err := repo.ListEvents(context.Background(), 0, mcpFixtureNow+1)
				if err != nil {
					t.Fatal(err)
				}
				var original, failure app.StoredEvent
				for _, row := range rows {
					if row.OriginKey == origin {
						original = row
					} else {
						failure = row
					}
				}
				if original.ID == "" {
					t.Fatal("original receipt absent after error")
				}
				accepted := "accepted"
				if operation == "veto" {
					accepted = "acknowledged"
					if text["pending"] != true || original.Event != nil {
						t.Fatal("task veto gained a fake topic")
					}
				} else if original.Observation.Outcome != "unknown" {
					t.Fatal("unverified raw outcome became trusted")
				}
				if mode == "capture" {
					if len(rows) != 2 || text[accepted] != true || text["event_id"] != original.ID || text["delivery_state"] != "committed" || text["capture_pending"] != true {
						t.Fatalf("known commit discarded: %v; rows=%v", text, rows)
					}
					ids := text["receipt_ids"].([]any)
					if len(ids) != 2 || !mcpContainsID(ids, original.ID) || !mcpContainsID(ids, failure.ID) {
						t.Fatalf("typed receipt IDs lost: %v", text)
					}
				} else {
					if len(rows) != 1 || text["delivery_state"] != "delivery_unknown" || text["capture_pending"] != false {
						t.Fatalf("unknown delivery overstated: %v", text)
					}
					if _, ok := text[accepted]; ok {
						t.Fatalf("unknown write incorrectly acknowledged/rejected: %v", text)
					}
					if _, ok := text["event_id"]; ok {
						t.Fatalf("unknown delivery fabricated ID: %v", text)
					}
				}
				var candidates int
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_experiences").Scan(&candidates); err != nil {
					t.Fatal(err)
				}
				if candidates != 0 {
					t.Fatal("failed capture already created a candidate")
				}
				db.Close()
				// Restart into actual W5 catchup, without re-Record/replay/new
				// origin. Read tool remains approved-only after recovery.
				restarted := mcpSubprocess(t, path, "human", mcpCall(2, "get_experiences", `{"pattern":"agent"}`), "ASTROCYTE_M5_MCP_RECOVER=1")
				if got := mcpDecoded[app.GetExperiencesResult](t, restarted[1]); len(got.Experiences) != 0 {
					t.Fatal("quarantine candidate leaked to MCP")
				}
				db = mcpOpen(t, path)
				defer db.Close()
				repo = sqlite.NewEventRepository(db)
				recoveredRows, err := repo.ListEvents(context.Background(), 0, mcpFixtureNow+1)
				if err != nil || !reflect.DeepEqual(rows, recoveredRows) {
					t.Fatalf("restart replayed/changed original receipts: %v", err)
				}
				experiences := app.NewExperienceService(sqlite.NewExperienceRepository(db), repo, app.ExperienceOptions{ProjectID: "project-a", Clock: mcpFixtureClock})
				recovered, err := experiences.CaptureFailures(context.Background(), 0, mcpFixtureNow+1)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "capture" {
					if len(recovered) != 1 || recovered[0].OriginEventID != failure.ID || recovered[0].Experience.Status != domain.StatusQuarantine || len(recovered[0].Experience.Evidence) != 1 || recovered[0].Experience.Evidence[0].EventID != failure.ID {
						t.Fatalf("recovery lost original failure evidence: %v", recovered)
					}
				} else if len(recovered) != 0 {
					t.Fatal("unknown report generated a failure candidate")
				}
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_experiences").Scan(&candidates); err != nil {
					t.Fatal(err)
				}
				if candidates != len(recovered) {
					t.Fatal("restart catchup duplicated candidate")
				}
			})
		}
	}
}

func mcpContainsID(ids []any, id string) bool {
	for _, value := range ids {
		if value == id {
			return true
		}
	}
	return false
}
