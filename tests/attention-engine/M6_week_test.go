package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/songconmaisaix31-design/Astrocyte/cmd/attentionruntime"
	"io"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type m6Day struct {
	Day   string                 `json:"day"`
	State []domain.AttentionNode `json:"state"`
	Sleep app.EngineSleepRun     `json:"sleep"`
}
type m6Snapshot struct {
	State           []domain.AttentionNode `json:"state"`
	Recommendations []app.KnowledgeMatch   `json:"recommendations"`
	Explanation     app.FocusExplanation   `json:"explanation"`
	Experiences     []app.StoredExperience `json:"experiences"`
	Conflicts       []app.DynamicsConflict `json:"conflicts"`
	Events          []app.StoredEvent      `json:"events"`
}

func m6Knowledge(db *sqlite.DB, now time.Time) *app.KnowledgeService {
	return app.NewKnowledgeService(sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db)), app.KnowledgeServiceOptions{Clock: func() time.Time { return now }})
}
func m6Engine(db *sqlite.DB, now time.Time) *app.DynamicsEngine {
	decay, confidence, success, drop := .1, .5, .05, .2
	return app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence, SuccessConfidenceDelta: &success, ConfidenceDropThreshold: &drop, Clock: func() time.Time { return now }})
}
func m6Read(t *testing.T, db *sqlite.DB, now time.Time) m6Snapshot {
	t.Helper()
	ctx := context.Background()
	engine := m6Engine(db, now)
	knowledge := m6Knowledge(db, now)
	human := app.Principal{ID: "local-human", Kind: "human"}
	lambda := .00000001
	var result m6Snapshot
	var err error
	result.State, err = engine.AttentionState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	result.Recommendations, err = knowledge.RetrieveKnowledge(ctx, human, "project-a", app.KnowledgeQuery{Now: now.UnixMilli(), LambdaR: &lambda, AgeBasis: app.KnowledgeAgeCreated, ReferenceBoost: .1})
	if err != nil {
		t.Fatal(err)
	}
	result.Explanation, err = engine.ExplainFocus(ctx, "implementation")
	if err != nil {
		t.Fatal(err)
	}
	result.Experiences, err = app.NewExperienceService(sqlite.NewExperienceRepository(db), sqlite.NewEventRepository(db), app.ExperienceOptions{ProjectID: "project-a"}).GetExperiences(ctx, human, "project-a", "")
	if err != nil {
		t.Fatal(err)
	}
	result.Conflicts, err = engine.Conflicts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	result.Events, err = sqlite.NewEventRepository(db).ListEvents(ctx, 0, now.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestM6ColdProcess(t *testing.T) {
	path := os.Getenv("ASTROCYTE_M6_COLD_DB")
	if path == "" {
		t.Skip("cold process helper")
	}
	now, err := time.Parse(time.RFC3339, os.Getenv("ASTROCYTE_M6_COLD_NOW"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	result := m6Read(t, db, now)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("M6-COLD:" + string(data))
	os.Exit(0)
}

// M6 joins the original W1/W2/W3/W4/W5/W7 implementations. All seven days,
// identities, raw observations and policy values are contract-local fixtures.
// No model, real native transcript, personal root or scheduler participates.
func TestM6MixedWeekRecommendationExplanationAndRestart(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	now := start
	path := filepath.Join(t.TempDir(), "m6.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	for _, topic := range []string{"research", "implementation"} {
		if _, err = m6Engine(db, now).SeedTopic(ctx, app.Principal{ID: "fixture-human", Kind: "human"}, topic, start.UnixMilli()); err != nil {
			t.Fatal(err)
		}
	}
	repo := sqlite.NewEventRepository(db)
	experiences := app.NewExperienceService(sqlite.NewExperienceRepository(db), repo, app.ExperienceOptions{ProjectID: "project-a", Clock: func() time.Time { return now }})
	recorder := app.NewEventService(repo, .2, experiences.CaptureCommittedFailure)
	originals := map[string][]byte{}
	var days []m6Day
	var failedID string
	for dayIndex := 0; dayIndex < 7; dayIndex++ {
		day := start.AddDate(0, 0, dayIndex)
		now = day.Add(23 * time.Hour)
		topic, n := "research", 2
		if dayIndex >= 3 {
			topic, n = "implementation", 4
		}
		var batch []app.StoredEvent
		for i := 0; i < n; i++ {
			source := []string{"manual", "codex", "claude_code", "mcp"}[i%4]
			kind := []string{"collect", "edit_file", "commit", "search"}[i%4]
			value := "literal-v1"
			if dayIndex == 5 {
				value = "literal-v2"
			}
			raw := []byte(fmt.Sprintf(` { "facts": { "fixture-method": %q }, "padding": %q } `, value, strings.Repeat("contract_local synthetic fact; ", 50)))
			ts := day.Add(time.Duration(9+i) * time.Hour).UnixMilli()
			key := fmt.Sprintf("m6:%d:%d", dayIndex, i)
			batch = append(batch, app.StoredEvent{OriginKey: key, ReceivedAt: ts, RawBytes: raw, Observation: app.Observation{Source: source, SourceKind: kind, Scope: "project-a", Session: "fixture-week", TS: ts, Outcome: "unknown", Depth: "unknown", Provenance: "contract_local:M6"}, Event: &domain.Event{Source: source, Kind: kind, Scope: "project-a", Session: "fixture-week", TS: ts, Topic: topic, Raw: raw}})
		}
		if dayIndex == 1 {
			ts := day.Add(14 * time.Hour).UnixMilli()
			raw := []byte(` { "outcome": "failed", "facts": { "fixture-failure": "rule-captured" } } `)
			batch = append(batch, app.StoredEvent{OriginKey: "m6:trusted-failed-task", ReceivedAt: ts, RawBytes: raw, Observation: app.Observation{Source: "codex", SourceKind: "agent_task", Scope: "project-a", Session: "fixture-week", TS: ts, Outcome: "failed", Depth: "unknown", Provenance: "contract_local:trusted_task_completion"}, Event: &domain.Event{Source: "codex", Kind: "agent_task", Scope: "project-a", TS: ts, Topic: "research", Raw: raw}})
		}
		rows, err := recorder.Record(ctx, batch)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			originals[row.ID] = row.RawBytes
			if row.OriginKey == "m6:trusted-failed-task" {
				failedID = row.ID
			}
		}
		// Machine Works refresh remains a pending source observation, not a topic
		// or successful agent task. It goes through the actual original sink.
		work := workspacedomain.WorkEvent{WorkID: "fixture-work", Kind: "works.refreshed", Source: "machine", Scope: "project-a", At: day.Add(15 * time.Hour), Payload: json.RawMessage(` { "revision": 1, "counts": null } `)}
		if err = events.NewWorksEventSink(app.NewPendingEventService(repo), func() time.Time { return now }).EmitWorkEvent(ctx, work); err != nil {
			t.Fatal(err)
		}
		engine := m6Engine(db, now)
		decay, confidence, success, drop := .1, .5, .05, .2
		sleep := attentionruntime.NewSleepService(db, app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence, SuccessConfidenceDelta: &success, ConfidenceDropThreshold: &drop, Clock: func() time.Time { return now }}, app.SleepOptions{Clock: func() time.Time { return now }, ProjectFacts: app.LiteralSleepFactsMap})
		result, err := sleep.Run(ctx, day)
		if err != nil {
			t.Fatalf("day%d Sleep: %v", dayIndex+1, err)
		}
		if result.Replayed || !result.Run.Metrics.RetentionMeasurable || !result.Run.Metrics.StorageMeasurable || result.Run.Metrics.Representation == "" {
			t.Fatalf("daily metrics: %+v", result)
		}
		state, err := engine.AttentionState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		days = append(days, m6Day{Day: day.Format(time.DateOnly), State: state, Sleep: result.Run})
		replay, err := sleep.Run(ctx, day)
		if err != nil || !replay.Replayed || !reflect.DeepEqual(replay.Run, result.Run) {
			t.Fatalf("daily replay: %+v %v", replay, err)
		}
	}
	if days[0].State[0].Topic != "research" || days[6].State[0].Topic != "implementation" {
		t.Fatalf("focus did not migrate: %+v", days)
	}
	if visible, err := experiences.GetExperiences(ctx, app.Principal{ID: "fixture-agent", Kind: "agent"}, "project-a", ""); err != nil || len(visible) != 0 {
		t.Fatalf("quarantine visible: %+v %v", visible, err)
	}
	candidate, err := experiences.CaptureFailure(ctx, failedID)
	if err != nil || candidate == nil {
		t.Fatalf("failure: %+v %v", candidate, err)
	}
	knowledge := app.NewKnowledgeService(sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db)), app.KnowledgeServiceOptions{Clock: func() time.Time { return now }})
	engine := m6Engine(db, now)
	api := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{KnowledgeGovernance: knowledge, ExperienceGovernance: experiences, ConflictGate: engine}, AgentToken: "fixture-agent", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	// Use the actual session/CSRF middleware and gate routes, with a fresh
	// in-process HTTP connection; the business services are never stubbed.
	httpHarness := newM6HTTP(t, api.Handler())
	defer httpHarness.Close()
	ids := map[string]string{}
	rows, err := db.Conn().Query(`SELECT id,topic FROM attention_engine_knowledge WHERE owner_project='project-a' ORDER BY topic`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id, topic string
		if err = rows.Scan(&id, &topic); err != nil {
			t.Fatal(err)
		}
		ids[topic] = id
	}
	if err = rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("knowledge: %+v", ids)
	}
	human := app.Principal{ID: "local-human", Kind: "human"}
	for topic, id := range ids {
		entry, err := knowledge.GetKnowledge(ctx, human, id, "project-a")
		if err != nil {
			t.Fatal(err)
		}
		command := app.ConfirmKnowledgeCommand{CommandMeta: m6Meta("confirm:"+topic, entry.Version), OwnerProject: "project-a"}
		httpHarness.Command("/api/v1/knowledge/"+id+"/confirm", command, "agent", 403)
		httpHarness.Command("/api/v1/knowledge/"+id+"/confirm", command, "human", 200)
	}
	httpHarness.Command("/api/v1/experiences/"+candidate.Experience.ID+"/approve", app.ApproveExperienceCommand{CommandMeta: m6Meta("approve-experience", candidate.Version), OwnerProject: "project-a"}, "human", 200)
	entry, err := knowledge.GetKnowledge(ctx, human, ids["implementation"], "project-a")
	if err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 2; version++ {
		now = now.Add(time.Millisecond)
		entry, err = knowledge.AppendCrystallized(ctx, app.KnowledgeCrystallizedUpdate{ID: entry.Entry.ID, OwnerProject: "project-a", ExpectedVersion: entry.Version, TS: now.UnixMilli(), Policy: app.KnowledgeKeepConfirmation, Content: json.RawMessage(fmt.Sprintf(`{"fixture-version":%d}`, version))})
		if err != nil {
			t.Fatal(err)
		}
	}
	httpHarness.Command("/api/v1/knowledge/"+entry.Entry.ID+"/shares", app.ShareMemoryCommand{CommandMeta: m6Meta("share-knowledge", entry.Version), OwnerProject: "project-a", TargetProject: "project-b"}, "human", 200)
	entry, err = knowledge.GetKnowledge(ctx, human, entry.Entry.ID, "project-b")
	if err != nil || len(entry.CrystallizedVersions) != 2 {
		t.Fatalf("share/version: %+v %v", entry, err)
	}
	httpHarness.Command("/api/v1/knowledge/"+entry.Entry.ID+"/references", app.ReferenceKnowledgeCommand{CommandMeta: m6Meta("reference", entry.Version), ProjectID: "project-b"}, "human", 200)
	if _, err = knowledge.GetKnowledge(ctx, human, ids["research"], "project-b"); err == nil {
		t.Fatal("unshared memory visible")
	}
	conflicts, err := engine.Conflicts(ctx)
	if err != nil || len(conflicts) == 0 {
		t.Fatalf("literal contradictory facts missing: %+v %v", conflicts, err)
	}
	httpHarness.Command("/api/v1/conflicts/"+conflicts[0].Conflict.ID+"/adjudications", app.AdjudicateConflictCommand{CommandMeta: m6Meta("adjudicate", conflicts[0].Version), Decision: app.ConflictDecisionKeepLeft}, "agent", 403)
	httpHarness.Command("/api/v1/conflicts/"+conflicts[0].Conflict.ID+"/adjudications", app.AdjudicateConflictCommand{CommandMeta: m6Meta("adjudicate", conflicts[0].Version), Decision: app.ConflictDecisionKeepLeft}, "human", 200)
	before := m6Read(t, db, now)
	if len(before.Recommendations) != 2 || before.Recommendations[0].Knowledge.Entry.Topic != "implementation" || len(before.Experiences) != 1 {
		t.Fatalf("recommendation: %+v", before)
	}
	attentionByTopic := map[string]float64{}
	for _, node := range before.State {
		attentionByTopic[node.Topic] = node.Attention
	}
	for _, match := range before.Recommendations {
		attention, exists := attentionByTopic[match.Knowledge.Entry.Topic]
		if !exists || attention <= 0 || math.Abs(match.Inputs.Attention-attention) > 1e-12 || match.Score <= 0 {
			t.Fatalf("recommendation did not consume original node: topic=%s storedA=%.12f inputA=%.12f score=%.12f", match.Knowledge.Entry.Topic, attention, match.Inputs.Attention, match.Score)
		}
		age := float64(now.UnixMilli()-match.Knowledge.CreatedAt) / 1000
		boost := 1.0
		for _, action := range match.Knowledge.Actions {
			if action.Action == "reference" && action.ActorKind == "human" {
				boost += .1
			}
		}
		expected := attention * boost * math.Exp(-.00000001*age)
		if match.Inputs.Relevance != 1 || math.Abs(match.Inputs.AccessBoost-boost) > 1e-12 || math.Abs(match.Inputs.AgeSeconds-age) > 1e-12 || math.Abs(match.Score-expected) > 1e-12 {
			t.Fatalf("independent recommendation formula: got=%+v want=%.12f", match, expected)
		}
	}
	for _, item := range before.Explanation.Evidence {
		if item.EventID != item.Original.ID || !bytes.Equal(item.Original.RawBytes, originals[item.EventID]) || item.Strength != .2 {
			t.Fatalf("evidence: %+v", item)
		}
	}
	failureExplanation, err := engine.ExplainFocus(ctx, "research")
	if err != nil {
		t.Fatal(err)
	}
	foundFailure := false
	for _, item := range failureExplanation.Evidence {
		if item.EventID == failedID {
			foundFailure = true
			if !item.ConfidenceOnly || item.AttentionContribution != 0 || item.ConfidenceContribution >= 0 {
				t.Fatalf("failure attention: %+v", item)
			}
		}
	}
	if !foundFailure {
		t.Fatal("failure chain lost")
	}
	var pendingSignals, sleepCount int
	if err = db.Conn().QueryRow(`SELECT count(*) FROM attention_engine_signals s JOIN attention_engine_events e ON e.id=s.event_id WHERE e.mapping_status='pending'`).Scan(&pendingSignals); err != nil || pendingSignals != 0 {
		t.Fatalf("pending signals: %d %v", pendingSignals, err)
	}
	if err = db.Conn().QueryRow(`SELECT count(*) FROM attention_engine_sleep_runs`).Scan(&sleepCount); err != nil || sleepCount != 7 {
		t.Fatalf("daily runs: %d %v", sleepCount, err)
	}
	if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	lambda := .00000001
	readOnly, err := knowledge.RetrieveKnowledge(ctx, human, "project-a", app.KnowledgeQuery{Now: now.UnixMilli(), LambdaR: &lambda, AgeBasis: app.KnowledgeAgeCreated, ReferenceBoost: .1})
	if err != nil || !reflect.DeepEqual(readOnly, before.Recommendations) {
		t.Fatalf("read-only recommendation: %+v %v", readOnly, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestM6ColdProcess$")
	child.Env = append(os.Environ(), "ASTROCYTE_M6_COLD_DB="+path, "ASTROCYTE_M6_COLD_NOW="+now.Format(time.RFC3339Nano))
	output, err := child.CombinedOutput()
	if err != nil {
		t.Fatalf("cold process: %v %s", err, output)
	}
	var cold m6Snapshot
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "M6-COLD:") {
			if err = json.Unmarshal([]byte(strings.TrimPrefix(line, "M6-COLD:")), &cold); err != nil {
				t.Fatal(err)
			}
		}
	}
	// JSON-RPC/report serialization compacts nested RawMessage whitespace;
	// compare the public JSON shape and independently require exact BLOB bytes.
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	coldJSON, err := json.Marshal(cold)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeJSON, coldJSON) {
		t.Fatal("cold process changed persisted public state")
	}
	if len(cold.Events) != len(before.Events) {
		t.Fatal("cold process lost receipts")
	}
	for i, event := range cold.Events {
		if event.ID != before.Events[i].ID || !bytes.Equal(event.RawBytes, before.Events[i].RawBytes) {
			t.Fatalf("cold BLOB mismatch: %s", event.ID)
		}
	}
	report := struct {
		EvidenceClass string               `json:"evidence_class"`
		Days          []m6Day              `json:"days"`
		Final         m6Snapshot           `json:"final"`
		Failure       app.FocusExplanation `json:"failure_explanation"`
	}{"contract_local synthetic week; original production services + HTTP gates + cold child process; no task_live claim", days, before, failureExplanation}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if reportDir := os.Getenv("ASTROCYTE_M6_REPORT_DIR"); reportDir != "" {
		if err = os.MkdirAll(reportDir, 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(reportDir, "M6-week-result.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, day := range days {
		t.Logf("%s focus=%s A=%.9f retention=%.6f reduction=%.6f", day.Day, day.State[0].Topic, day.State[0].Attention, day.Sleep.Metrics.RetentionPrecision, day.Sleep.Metrics.StorageReduction)
	}
}
