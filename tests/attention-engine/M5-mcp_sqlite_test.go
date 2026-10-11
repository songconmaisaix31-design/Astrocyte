package attention_engine_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

const mcpFixtureStart int64 = 1767225600000
const mcpFixtureNow int64 = mcpFixtureStart + 40*86400000

func mcpFixtureClock() time.Time { return time.UnixMilli(mcpFixtureNow) }
func mcpPtr[T any](v T) *T       { return &v }

// This helper is a separate process with real stdin/stdout and SQLite services.
// It exits before the Go test runner can print non-protocol text to stdout.
func TestMCPServerProcess(t *testing.T) {
	if os.Getenv("ASTROCYTE_M5_MCP_HELPER") != "1" {
		t.Skip("stdio subprocess helper")
	}
	err := func() error {
		db, err := sqlite.Open(os.Getenv("ASTROCYTE_M5_MCP_DB"))
		if err != nil {
			return err
		}
		defer db.Close()
		kind := os.Getenv("ASTROCYTE_M5_MCP_KIND")
		binding := app.MCPBinding{Principal: app.Principal{ID: "m5-" + kind, Kind: kind}, Scope: "project-a", Session: "m5-stdio", CanVeto: kind == "human"}
		events := sqlite.NewEventRepository(db)
		experiences := app.NewExperienceService(sqlite.NewExperienceRepository(db), events, app.ExperienceOptions{ProjectID: binding.Scope, Clock: mcpFixtureClock})
		recorder := mcpProcessRecorder(events, experiences, os.Getenv("ASTROCYTE_M5_MCP_DELIVERY_FIXTURE"))
		if os.Getenv("ASTROCYTE_M5_MCP_RECOVER") == "1" {
			if _, err := experiences.CaptureFailures(context.Background(), 0, mcpFixtureNow+1); err != nil {
				return err
			}
		}
		backend, err := app.NewEngineMCPBackend(binding,
			app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{}),
			experiences, recorder, app.MCPBackendOptions{Clock: mcpFixtureClock})
		if err != nil {
			return err
		}
		server, err := mcp.NewAstrocyteServer(backend, binding)
		if err != nil {
			return err
		}
		return server.Serve(context.Background(), os.Stdin, os.Stdout)
	}()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func mcpSubprocess(t *testing.T, path, kind, script string, fixtureEnv ...string) []map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMCPServerProcess$")
	command.Env = append(os.Environ(), "ASTROCYTE_M5_MCP_HELPER=1", "ASTROCYTE_M5_MCP_DB="+path, "ASTROCYTE_M5_MCP_KIND="+kind)
	command.Env = append(command.Env, fixtureEnv...)
	command.Stdin = bytes.NewBufferString(req(1, "initialize", "") + "{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n" + script)
	var output, stderr bytes.Buffer
	command.Stdout, command.Stderr = &output, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("MCP subprocess: %v; stderr=%s; stdout=%s", err, stderr.String(), output.String())
	}
	var responses []map[string]any
	scanner := bufio.NewScanner(&output)
	scanner.Buffer(make([]byte, 64*1024), 16<<20)
	for scanner.Scan() {
		var response map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			t.Fatalf("stdout is not MCP JSON: %q: %v", scanner.Text(), err)
		}
		responses = append(responses, response)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return responses
}

func mcpCall(id int, name, args string) string {
	return req(id, "tools/call", `{"name":"astrocyte.`+name+`","arguments":`+args+`}`)
}

func mcpDecoded[T any](t *testing.T, response map[string]any) T {
	t.Helper()
	result, ok := response["result"].(map[string]any)
	if !ok || result["isError"] == true {
		t.Fatalf("MCP call failed: %v", response)
	}
	var value T
	if err := json.Unmarshal([]byte(toolText(t, response)), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func mcpSeedReceipt(t *testing.T, service *app.EventService, key, project, topic, kind, source, outcome, depth string, ts int64) app.StoredEvent {
	t.Helper()
	raw := []byte(fmt.Sprintf(`{ "fixture": %q, "kind": %q, "outcome": %q, "depth": %q }`, key, kind, outcome, depth))
	rows, err := service.Record(context.Background(), []app.StoredEvent{{OriginKey: "M5-fixture:" + key, ReceivedAt: ts, RawBytes: raw,
		Observation: app.Observation{Source: source, SourceKind: kind, Session: "m5-fixture", Scope: project, TS: ts, Outcome: outcome, Depth: depth, Provenance: "M5 contract_local synthetic input"},
		Event:       &domain.Event{Source: source, Session: "m5-fixture", Scope: project, TS: ts, Topic: topic, Kind: kind, Raw: raw}}})
	if err != nil {
		t.Fatal(err)
	}
	return rows[0]
}

func mcpOpen(t *testing.T, path string) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db
}

func mcpReadSnapshot(t *testing.T, db *sqlite.DB) string {
	t.Helper()
	var rows []string
	for _, query := range []string{
		"SELECT data||lifecycle||version||updated_at FROM attention_engine_nodes ORDER BY id",
		"SELECT data||version||updated_at FROM attention_engine_experiences ORDER BY id",
		"SELECT id FROM attention_engine_events ORDER BY id",
	} {
		result, err := db.Conn().Query(query)
		if err != nil {
			t.Fatal(err)
		}
		for result.Next() {
			var row string
			if err := result.Scan(&row); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, row)
		}
		if err := result.Err(); err != nil {
			t.Fatal(err)
		}
		result.Close()
	}
	data, _ := json.Marshal(rows)
	return string(data)
}

func TestMCPRealStdioSQLiteFiveOperationsAndRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m5.sqlite")
	db := mcpOpen(t, path)
	defer func() { db.Close() }()
	eventRepo := sqlite.NewEventRepository(db)
	eventService := app.NewEventService(eventRepo, .01) // fixture scale, not a product default
	dynamicsRepo := sqlite.NewDynamicsRepository(db)
	engine := app.NewDynamicsEngine(dynamicsRepo, app.DynamicsOptions{Decay: mcpPtr(.1), InitialConfidence: mcpPtr(.5), ConfidenceDropThreshold: mcpPtr(.05)})
	human := app.Principal{ID: "fixture-human", Kind: "human"}
	for _, topic := range []string{"Agent Swarm", "SQLite"} {
		if _, err := engine.SeedTopic(ctx, human, topic, mcpFixtureStart-86400000); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 31; i++ {
		kind, source, depth := domain.EventKindReadPaper, domain.EventSourceClaudeCode, "deep"
		if i >= 18 {
			kind, source, depth = domain.EventKindEditFile, domain.EventSourceCodex, "unknown"
		}
		if i == 30 {
			kind, source = domain.EventKindUserApprove, domain.EventSourceManual
		}
		project := "project-a"
		if i%2 == 1 {
			project = "project-b"
		}
		mcpSeedReceipt(t, eventService, fmt.Sprintf("chain-%02d", i), project, "Agent Swarm", kind, source, "unknown", depth, mcpFixtureStart+int64(i)*86400000)
	}
	failure := mcpSeedReceipt(t, eventService, "failure-a", "project-a", "Agent Swarm", domain.EventKindAgentTask, domain.EventSourceCodex, "failed", "unknown", mcpFixtureStart+31*86400000)
	signals, err := eventRepo.ListSignals(ctx, mcpFixtureStart, mcpFixtureStart+31*86400000)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 32 {
		t.Fatalf("seed evidence signals: %d", len(signals))
	}
	if _, err := engine.ApplySignals(ctx, "Agent Swarm", signals, mcpFixtureStart+31*86400000); err != nil {
		t.Fatal(err)
	}
	// Seed one persisted association; the lifecycle algorithm remains W2's test.
	if err := dynamicsRepo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		n, err := tx.GetNode("Agent Swarm")
		if err != nil {
			return err
		}
		var history app.DynamicsHistory
		if err := json.Unmarshal(n.History, &history); err != nil {
			return err
		}
		n.Node.RelatedTopics["SQLite"] = .7
		history.RelatedAt = map[string]int64{"SQLite": mcpFixtureStart + 31*86400000}
		n.History, err = json.Marshal(history)
		if err != nil {
			return err
		}
		n.Version++
		return tx.SaveNode(n)
	}); err != nil {
		t.Fatal(err)
	}
	newExperience := func(project string) *app.ExperienceService {
		return app.NewExperienceService(sqlite.NewExperienceRepository(db), eventRepo, app.ExperienceOptions{ProjectID: project, Clock: mcpFixtureClock})
	}
	approve := func(service *app.ExperienceService, row *app.StoredExperience) {
		t.Helper()
		if row == nil {
			t.Fatal("failure produced no candidate")
		}
		if _, err := service.ApproveExperience(ctx, human, row.Experience.ID, app.ApproveExperienceCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "approve-" + row.Experience.ID, IdempotencyKey: "approve-" + row.Experience.ID, ExpectedVersion: 1}, OwnerProject: row.OwnerProject}); err != nil {
			t.Fatal(err)
		}
	}
	experienceA, experienceB := newExperience("project-a"), newExperience("project-b")
	candidateA, err := experienceA.CaptureFailure(ctx, failure.ID)
	if err != nil {
		t.Fatal(err)
	}
	approve(experienceA, candidateA)
	for i := 0; i < 3; i++ {
		project := "project-b"
		if i == 2 {
			project = "project-a"
		}
		raw := mcpSeedReceipt(t, eventService, fmt.Sprintf("extra-failure-%d", i), project, "SQLite", domain.EventKindAgentTask, domain.EventSourceCodex, "failed", "unknown", mcpFixtureStart+int64(32+i)*86400000)
		service := experienceB
		if project == "project-a" {
			service = experienceA
		}
		candidate, err := service.CaptureFailure(ctx, raw.ID)
		if err != nil {
			t.Fatal(err)
		}
		if i == 2 {
			continue
		} // quarantine must never appear via MCP
		approve(service, candidate)
		if i == 1 {
			if _, err := service.ShareExperience(ctx, human, candidate.Experience.ID, app.ShareMemoryCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "share-b-a", IdempotencyKey: "share-b-a", ExpectedVersion: 2}, OwnerProject: "project-b", TargetProject: "project-a"}); err != nil {
				t.Fatal(err)
			}
		}
	}
	before := mcpReadSnapshot(t, db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	readScript := req(2, "tools/list", "") + mcpCall(3, "get_attention_state", `{}`) + mcpCall(4, "explain_focus", `{"topic":"Agent Swarm"}`) +
		mcpCall(5, "get_experiences", `{"pattern":"agent_task"}`) + mcpCall(6, "get_experiences", `{"pattern":"agent_task","scope":"project-b"}`) + mcpCall(7, "explain_focus", `{"topic":"Unobserved"}`)
	first := mcpSubprocess(t, path, "human", readScript)
	if len(first) != 7 {
		t.Fatalf("read responses: %d", len(first))
	}
	if got := first[1]["result"].(map[string]any)["tools"].([]any); len(got) != 5 {
		t.Fatalf("tools: %v", got)
	}
	state := mcpDecoded[app.GetAttentionStateResult](t, first[2])
	if len(state.Topics) != 2 || state.Topics[0].Topic != "Agent Swarm" || state.Topics[0].Attention < state.Topics[1].Attention {
		t.Fatalf("focus ordering: %+v", state)
	}
	explanation := mcpDecoded[app.FocusExplanation](t, first[3])
	if len(explanation.Evidence) != 32 || len(explanation.RelatedProjects) != 2 || len(explanation.RelatedAgents) != 2 || len(explanation.Related) != 1 || explanation.Related[0].TS == 0 {
		t.Fatalf("incomplete explanation: %+v", explanation)
	}
	kinds := map[string]int{}
	for _, item := range explanation.Evidence {
		kinds[item.Kind]++
		if item.EventID == "" || item.Original.ID != item.EventID || item.TS == 0 || item.Original.Event == nil || item.Original.Event.TS != item.TS || len(item.Original.RawBytes) == 0 || item.Strength != .01 || item.Decay != .1 || math.Abs(item.DecayFactor-math.Exp(-item.Decay*item.ElapsedDays)) > 1e-12 {
			t.Fatalf("evidence path: %+v", item)
		}
		if item.Kind == domain.EventKindAgentTask && (!item.ConfidenceOnly || item.AttentionContribution != 0 || item.ConfidenceContribution >= 0) {
			t.Fatalf("failure path: %+v", item)
		}
	}
	if kinds[domain.EventKindReadPaper] != 18 || kinds[domain.EventKindEditFile] != 12 || kinds[domain.EventKindUserApprove] != 1 {
		t.Fatalf("explanation references: %v", kinds)
	}
	experiences := mcpDecoded[app.GetExperiencesResult](t, first[4])
	if len(experiences.Experiences) != 2 {
		t.Fatalf("approved owner/share scope: %+v", experiences)
	}
	owners := map[string]int{}
	for _, row := range experiences.Experiences {
		owners[row.OwnerProject]++
		if row.Experience.Status != domain.StatusApproved || len(row.Experience.Evidence) != 1 || row.Experience.Evidence[0].TS == 0 {
			t.Fatalf("experience evidence/status: %+v", row)
		}
	}
	if owners["project-a"] != 1 || owners["project-b"] != 1 {
		t.Fatalf("owners: %v", owners)
	}
	for _, index := range []int{5, 6} {
		if first[index]["result"].(map[string]any)["isError"] != true {
			t.Fatalf("scope/unknown focus accepted: %v", first[index])
		}
	}
	db = mcpOpen(t, path)
	if after := mcpReadSnapshot(t, db); after != before {
		t.Fatal("read tools mutated persisted state")
	}
	db.Close()
	second := mcpSubprocess(t, path, "human", readScript)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("cold stdio restart changed read results")
	}

	reportedJSON := `{ "source":"manual", "session":"forged", "ts":` + fmt.Sprint(mcpFixtureNow) + `, "kind":"commit", "topic":"Agent Swarm", "raw":{"outcome":"success","actor_kind":"human"} }`
	writeScript := mcpCall(2, "report_event", `{"event":`+reportedJSON+`}`) +
		mcpCall(3, "report_event", `{"event":{"ts":`+fmt.Sprint(mcpFixtureNow)+`,"kind":"search"}}`) +
		mcpCall(4, "report_event", `{"event":{"ts":`+fmt.Sprint(mcpFixtureNow)+`,"kind":"agent_task","topic":"Agent Swarm","raw":{"outcome":"success"}}}`) +
		mcpCall(5, "veto", `{"topic":"Agent Swarm","reason":"user rejected focus"}`) +
		mcpCall(6, "veto", `{"task":"unresolved-task","reason":"stop this task"}`)
	writes := mcpSubprocess(t, path, "human", writeScript)
	if len(writes) != 6 {
		t.Fatalf("write responses: %d", len(writes))
	}
	reported := mcpDecoded[app.ReportEventResult](t, writes[1])
	pending := mcpDecoded[app.ReportEventResult](t, writes[2])
	feedback := mcpDecoded[app.ReportEventResult](t, writes[3])
	veto := mcpDecoded[app.VetoResult](t, writes[4])
	task := mcpDecoded[app.VetoResult](t, writes[5])
	if !reported.Accepted || !pending.Accepted || !feedback.Accepted || !veto.Acknowledged || veto.Pending || !task.Acknowledged || !task.Pending {
		t.Fatalf("write receipts: %v", writes)
	}
	db = mcpOpen(t, path)
	eventRepo = sqlite.NewEventRepository(db)
	for _, id := range []string{reported.EventID, pending.EventID, feedback.EventID, veto.EventID, task.EventID} {
		row, err := eventRepo.GetEvent(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if row.Observation.Scope != "project-a" || row.Observation.Session != "m5-stdio" {
			t.Fatalf("startup binding: %+v", row)
		}
		var signals int
		if err := db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_signals WHERE event_id=?", id).Scan(&signals); err != nil {
			t.Fatal(err)
		}
		if id == pending.EventID || id == task.EventID {
			if row.Event != nil || signals != 0 {
				t.Fatalf("pending receipt became a signal: %+v", row)
			}
		}
		if id == reported.EventID && (!bytes.Equal(row.RawBytes, []byte(reportedJSON)) || row.Event.Source != "mcp" || row.Observation.Tokens != nil || row.Observation.Outcome != "unknown" || signals != 1) {
			t.Fatalf("report provenance/raw: %+v", row)
		}
		if id == feedback.EventID && (row.Observation.Outcome != "unknown" || signals != 0) {
			t.Fatalf("forged feedback influenced signal: %+v", row)
		}
		if id == veto.EventID && (row.Event.Source != "manual" || row.Event.Kind != "user_veto" || signals != 1) {
			t.Fatalf("veto did not enter event stream: %+v", row)
		}
	}
	beforeAgent := mcpReadSnapshot(t, db)
	db.Close()
	agentResponses := mcpSubprocess(t, path, "agent", mcpCall(2, "veto", `{"topic":"Agent Swarm"}`)+
		mcpCall(3, "report_event", `{"event":{"ts":`+fmt.Sprint(mcpFixtureNow)+`,"kind":"user_approve","topic":"Agent Swarm"}}`)+
		mcpCall(4, "report_event", `{"event":{"ts":`+fmt.Sprint(mcpFixtureNow)+`,"kind":"commit","topic":"Agent Swarm","scope":"project-b"}}`))
	for _, response := range agentResponses[1:] {
		if response["result"].(map[string]any)["isError"] != true {
			t.Fatalf("agent bypass: %v", response)
		}
	}
	db = mcpOpen(t, path)
	if after := mcpReadSnapshot(t, db); after != beforeAgent {
		t.Fatal("denied Agent tools modified storage")
	}
	// Apply the persisted human signal through W2 explicitly, outside MCP.
	eventRepo = sqlite.NewEventRepository(db)
	postSignals, err := eventRepo.ListSignals(ctx, mcpFixtureNow, mcpFixtureNow)
	if err != nil {
		t.Fatal(err)
	}
	engine = app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{})
	if _, err := engine.ApplySignals(ctx, "Agent Swarm", postSignals, mcpFixtureNow); err != nil {
		t.Fatal(err)
	}
	db.Close()
	afterApply := mcpSubprocess(t, path, "human", mcpCall(2, "explain_focus", `{"topic":"Agent Swarm"}`))
	final := mcpDecoded[app.FocusExplanation](t, afterApply[1])
	if len(final.Evidence) != 34 {
		t.Fatalf("cold reread lost applied report/veto chain: %d", len(final.Evidence))
	}
	var vetoFound bool
	for _, item := range final.Evidence {
		if item.EventID == veto.EventID {
			vetoFound = item.Weight == -.8 && item.TS == mcpFixtureNow && item.Original.ID == veto.EventID
		}
	}
	if !vetoFound {
		t.Fatal("user veto weight/evidence missing after engine replay and process restart")
	}
	t.Logf("stdio process+SQLite: 5 operations, topic/task veto, 32→34 explanation receipts, 2 approved scoped experiences, exact raw BLOB and cold restart")
}
