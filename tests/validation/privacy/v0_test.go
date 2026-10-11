package privacy_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// Deliberately RED acceptance probes are opt-in until the schema exception and
// content-versus-authorized-derived policy are decided. Only synthetic data.
const contentCanary = "FIXTURE_PRIVATE_CONTENT_W0_8f392a"
const factCanary = "FIXTURE_PRIVATE_FACT_W0_0c128b"
const scope = "fixture-project-w0"
const topic = "fixture-topic-w0"
const ts int64 = 1791730000123

type fixture struct {
	db           *sqlite.DB
	repo         *sqlite.EventRepository
	events       *app.EventService
	engine       *app.DynamicsEngine
	experiences  *app.ExperienceService
	row          app.StoredEvent
	experienceID string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewEventRepository(db)
	exps := app.NewExperienceService(sqlite.NewExperienceRepository(db), repo, app.ExperienceOptions{ProjectID: scope, Clock: func() time.Time { return time.UnixMilli(ts + 10) }})
	events := app.NewEventService(repo, .1, exps.CaptureCommittedFailure)
	decay, confidence, drop := .05, .5, .9 // Explicit fixture policy, never product defaults.
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence, ConfidenceDropThreshold: &drop})
	human := app.Principal{ID: "fixture-human", Kind: "human"}
	if _, err = engine.SeedTopic(ctx, human, topic, ts-1); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{"message": contentCanary, "facts": map[string]string{"fixture_fact_w0": factCanary}, "kind": "agent_task", "topic": topic, "scope": scope, "ts": ts, "tokens": 876543, "outcome": "failed"})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := events.Record(ctx, []app.StoredEvent{{OriginKey: "fixture:privacy-w0", ReceivedAt: ts + 1, RawBytes: raw, Observation: app.Observation{Source: "codex", SourceKind: "agent_task", Scope: scope, Session: "fixture-session", TS: ts, Tokens: nil, Outcome: "failed", Depth: "unknown", Provenance: "contract_local"}, Event: &domain.Event{Source: "codex", Session: "fixture-session", Scope: scope, Topic: topic, TS: ts, Kind: "agent_task", Tokens: 876543}}})
	if err != nil || len(rows) != 1 {
		t.Fatalf("record failed: %v", err)
	}
	signals, err := repo.ListSignals(ctx, ts, ts)
	if err != nil || len(signals) != 1 {
		t.Fatalf("signals failed: %v", err)
	}
	if _, err = engine.ApplySignals(ctx, topic, signals, ts); err != nil {
		t.Fatal(err)
	}
	candidate, err := exps.CaptureFailure(ctx, rows[0].ID)
	if err != nil || candidate == nil {
		t.Fatalf("capture failed: %v", err)
	}
	if _, err = exps.ApproveExperience(ctx, human, candidate.Experience.ID, app.ApproveExperienceCommand{OwnerProject: scope, CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "fixture-approval", ExpectedVersion: 1, IdempotencyKey: "fixture-approval"}}); err != nil {
		t.Fatal(err)
	}
	return &fixture{db: db, repo: repo, events: events, engine: engine, experiences: exps, row: rows[0], experienceID: candidate.Experience.ID}
}

func binding(human bool) app.MCPBinding {
	principal := app.Principal{ID: "fixture-agent", Kind: "agent"}
	if human {
		principal = app.Principal{ID: "fixture-human", Kind: "human"}
	}
	return app.MCPBinding{Principal: principal, Scope: scope, Session: "fixture-session", CanVeto: human}
}

func (f *fixture) backend(t *testing.T, b app.MCPBinding, recorder app.MCPEventRecorder) app.MCPBackend {
	t.Helper()
	backend, err := app.NewEngineMCPBackend(b, f.engine, f.experiences, recorder, app.MCPBackendOptions{Clock: func() time.Time { return time.UnixMilli(ts + 20) }})
	if err != nil {
		t.Fatal(err)
	}
	return backend
}

func call(t *testing.T, backend app.MCPBackend, b app.MCPBinding, name string, args any) []byte {
	t.Helper()
	server, err := mcp.NewAstrocyteServer(backend, b)
	if err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "astrocyte." + name, "arguments": args}})
	if err != nil {
		t.Fatal(err)
	}
	input := "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2025-11-25\",\"capabilities\":{},\"clientInfo\":{\"name\":\"fixture\",\"version\":\"1\"}}}\n" + string(request) + "\n"
	var output bytes.Buffer
	if err = server.Serve(context.Background(), strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("expected initialize and tool reply, got %d", len(lines))
	}
	return lines[1]
}

func assertState(t *testing.T, wire []byte, errorExpected bool) {
	t.Helper()
	var envelope struct {
		Result struct {
			IsError bool `json:"isError"`
		}
		Error *struct{ Code int }
	}
	if err := json.Unmarshal(wire, &envelope); err != nil {
		t.Fatal(err)
	}
	if (envelope.Error != nil || envelope.Result.IsError) != errorExpected {
		t.Fatalf("unexpected error state, expected=%v", errorExpected)
	}
}

func forbidContent(t *testing.T, wire []byte, raw []byte) {
	t.Helper()
	if bytes.Contains(wire, []byte(contentCanary)) {
		t.Error("V0 raw content canary appears in serialized response")
	}
	if bytes.Contains(wire, []byte(factCanary)) {
		t.Error("V0 raw fact canary appears in serialized response")
	}
	if bytes.Contains(wire, []byte(base64.StdEncoding.EncodeToString(raw))) {
		t.Error("V0 raw bytes appear encoded as base64 in serialized response")
	}
}

func requireAudit(t *testing.T) {
	t.Helper()
	if os.Getenv("ASTROCYTE_V0_PRIVACY_AUDIT") != "1" {
		t.Skip("opt-in RED probe: ASTROCYTE_V0_PRIVACY_AUDIT=1")
	}
}

func TestV0SQLiteReadBoundaries(t *testing.T) {
	requireAudit(t)
	f := newFixture(t)
	ctx := context.Background()
	t.Run("generic_table", func(t *testing.T) {
		var data string
		var raw []byte
		if err := f.db.Conn().QueryRow("SELECT data,raw_bytes FROM attention_engine_events WHERE id=?", f.row.ID).Scan(&data, &raw); err != nil {
			t.Fatal(err)
		}
		if len(raw) != 0 {
			t.Error("V0 raw remains in generic event table")
		}
		forbidContent(t, []byte(data), f.row.RawBytes)
	})
	reads := map[string]func() (app.StoredEvent, error){
		"put_return": func() (app.StoredEvent, error) { return f.row, nil },
		"get_event":  func() (app.StoredEvent, error) { return f.repo.GetEvent(ctx, f.row.ID) },
		"list_events": func() (app.StoredEvent, error) {
			rows, e := f.repo.ListEvents(ctx, ts, ts)
			if e != nil {
				return app.StoredEvent{}, e
			}
			return rows[0], nil
		},
		"dynamics_get": func() (v app.StoredEvent, e error) {
			e = sqlite.NewDynamicsRepository(f.db).ViewDynamics(ctx, func(tx app.DynamicsTx) error { v, e = tx.GetEvent(f.row.ID); return e })
			return
		},
		"sleep_list": func() (v app.StoredEvent, e error) {
			e = sqlite.NewSleepRepository(f.db).WithSleepTx(ctx, func(tx app.SleepTx) error {
				rows, err := tx.ListEvents(ts, ts)
				if err != nil {
					return err
				}
				v = rows[0]
				return nil
			})
			return
		},
		"knowledge_source": func() (v app.StoredEvent, e error) {
			e = sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(f.db)).WithKnowledgeReadTx(ctx, func(tx app.KnowledgeTx) error { v, e = tx.KnowledgeSource(f.row.ID); return e })
			return
		},
	}
	for name, read := range reads {
		t.Run(name, func(t *testing.T) {
			row, err := read()
			if err != nil {
				t.Fatal(err)
			}
			if len(row.RawBytes) > 0 || (row.Event != nil && len(row.Event.Raw) > 0) {
				t.Error("V0 raw reaches non-audit StoredEvent")
			}
		})
	}
}

func toolArgs(name string) any {
	switch name {
	case "get_attention_state":
		return map[string]any{}
	case "explain_focus":
		return map[string]any{"topic": topic}
	case "get_experiences":
		return map[string]any{"pattern": "agent_task"}
	case "report_event":
		return map[string]any{"event": map[string]any{"kind": "search", "ts": ts + 20, "topic": topic, "raw": map[string]any{"message": contentCanary}}}
	case "veto":
		return map[string]any{"topic": topic, "reason": contentCanary}
	}
	panic("unknown fixture tool")
}

var tools = []string{"get_attention_state", "explain_focus", "get_experiences", "report_event", "veto"}

func TestV0MCPSerializationSuccess(t *testing.T) {
	requireAudit(t)
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			b := binding(name == "veto")
			wire := call(t, f.backend(t, b, f.events), b, name, toolArgs(name))
			assertState(t, wire, false)
			forbidContent(t, wire, f.row.RawBytes)
		})
	}
}

// Fault injection loads its error string from actual persisted raw content.
// This demonstrates transport propagation, not a naturally occurring engine error.
type rawErrorBackend struct {
	app.MCPBackend
	cause error
}

func (b rawErrorBackend) GetAttentionState(context.Context) (app.GetAttentionStateResult, error) {
	return app.GetAttentionStateResult{}, b.cause
}
func (b rawErrorBackend) ExplainFocus(context.Context, app.ExplainFocusInput) (app.ExplainFocusResult, error) {
	return app.ExplainFocusResult{}, b.cause
}
func (b rawErrorBackend) GetExperiences(context.Context, app.GetExperiencesInput) (app.GetExperiencesResult, error) {
	return app.GetExperiencesResult{}, b.cause
}
func (b rawErrorBackend) ReportEvent(context.Context, app.ReportEventInput) (app.ReportEventResult, error) {
	return app.ReportEventResult{}, b.cause
}
func (b rawErrorBackend) Veto(context.Context, app.VetoInput) (app.VetoResult, error) {
	return app.VetoResult{}, b.cause
}

func TestV0MCPSerializationBackendErrors(t *testing.T) {
	requireAudit(t)
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			b := binding(name == "veto")
			row, err := f.repo.GetEvent(context.Background(), f.row.ID)
			if err != nil {
				t.Fatal(err)
			}
			var raw struct{ Message string }
			if err = json.Unmarshal(row.RawBytes, &raw); err != nil {
				t.Fatal(err)
			}
			backend := rawErrorBackend{MCPBackend: f.backend(t, b, f.events), cause: fmt.Errorf("fixture injected store failure: %s", raw.Message)}
			wire := call(t, backend, b, name, toolArgs(name))
			assertState(t, wire, true)
			forbidContent(t, wire, f.row.RawBytes)
		})
	}
}

// A real commit followed by a controlled failed callback exercises both text
// and structuredContent acknowledgement serialization without a false rollback.
type pendingRecorder struct{ events *app.EventService }

func (r pendingRecorder) Record(ctx context.Context, rows []app.StoredEvent) ([]app.StoredEvent, error) {
	stored, err := r.events.Record(ctx, rows)
	if err != nil {
		return stored, err
	}
	return stored, &app.MCPDynamicsPendingError{Cause: fmt.Errorf("fixture post-commit failure: %s", contentCanary)}
}

func TestV0MCPSerializationCommittedErrors(t *testing.T) {
	requireAudit(t)
	for _, name := range []string{"report_event", "veto"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			b := binding(name == "veto")
			wire := call(t, f.backend(t, b, pendingRecorder{f.events}), b, name, toolArgs(name))
			assertState(t, wire, true)
			var response struct {
				Result struct {
					Structured map[string]any `json:"structuredContent"`
					Content    []struct{ Text string }
				}
			}
			if err := json.Unmarshal(wire, &response); err != nil {
				t.Fatal(err)
			}
			if response.Result.Structured["delivery_state"] != "committed" || response.Result.Structured["event_id"] == "" {
				t.Fatal("lost committed acknowledgement")
			}
			if response.Result.Structured["dynamics_pending"] != true {
				t.Fatal("lost pending status")
			}
			forbidContent(t, wire, f.row.RawBytes)
		})
	}
}

func TestV0TrustedAuthorization(t *testing.T) {
	f := newFixture(t)
	b := binding(false)
	backend := f.backend(t, b, f.events)
	for _, tc := range []struct {
		name string
		args any
	}{
		{"get_experiences", map[string]any{"pattern": "agent_task", "scope": "foreign"}},
		{"report_event", map[string]any{"event": map[string]any{"kind": "search", "ts": ts, "scope": "foreign"}}},
		{"veto", map[string]any{"topic": topic}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := call(t, backend, b, tc.name, tc.args)
			assertState(t, wire, true)
			forbidContent(t, wire, f.row.RawBytes)
		})
	}
	_, err := f.experiences.ApproveExperience(context.Background(), b.Principal, f.experienceID, app.ApproveExperienceCommand{OwnerProject: scope})
	if err == nil {
		t.Fatal("agent approved experience")
	}
}

func TestV0LiteralSharedValueConflict(t *testing.T) {
	f := newFixture(t)
	b := binding(false)
	wire := call(t, f.backend(t, b, f.events), b, "get_attention_state", toolArgs("get_attention_state"))
	assertState(t, wire, false)
	if !bytes.Contains(f.row.RawBytes, []byte(topic)) || !bytes.Contains(wire, []byte(topic)) {
		t.Fatal("fixture failed to demonstrate shared authorized-derived value")
	}
	forbidContent(t, wire, f.row.RawBytes)
	t.Log("raw.topic equals permitted exposure.topic; universal raw-value substring exclusion conflicts with authorized-derived output")
}

func TestV0MCPSerializationDecodeErrors(t *testing.T) {
	requireAudit(t)
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			b := binding(name == "veto")
			// An untrusted caller uses persisted synthetic content as an unknown
			// argument name. This is the actual decoder path, not backend injection.
			row, err := f.repo.GetEvent(context.Background(), f.row.ID)
			if err != nil {
				t.Fatal(err)
			}
			var raw struct{ Message string }
			if err = json.Unmarshal(row.RawBytes, &raw); err != nil {
				t.Fatal(err)
			}
			args := toolArgs(name).(map[string]any)
			args[raw.Message] = "fixture-unknown-value"
			wire := call(t, f.backend(t, b, f.events), b, name, args)
			assertState(t, wire, true)
			forbidContent(t, wire, row.RawBytes)
		})
	}
}

func TestV0RawFreeCompatibilityBlockers(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	// Data cleanup in a disposable database simulates the proposed adapter-only
	// split. No migration or production file is changed.
	if _, err := f.db.Conn().Exec("UPDATE attention_engine_events SET data=json_remove(data,'$.raw'),raw_bytes=NULL WHERE id=?", f.row.ID); err != nil {
		t.Fatal(err)
	}
	explanation, err := f.engine.ExplainFocus(ctx, topic)
	if err != nil || len(explanation.Evidence) != 1 {
		t.Fatalf("raw-free dynamics read failed: %v", err)
	}
	row := explanation.Evidence[0].Original
	if row.RawBytes != nil || len(row.Event.Raw) != 0 {
		t.Fatal("fixture still contains raw")
	}
	t.Run("failure_capture", func(t *testing.T) {
		_, err := f.experiences.CaptureFailure(ctx, f.row.ID)
		var serviceErr *apierrors.ServiceError
		if !errors.As(err, &serviceErr) || serviceErr.Code != apierrors.EvidenceMissing {
			t.Fatalf("expected raw-free capture blocker, got %v", err)
		}
	})
	t.Run("experience_evidence", func(t *testing.T) {
		_, err := f.experiences.ResolveExperienceEvidence(ctx, app.Principal{ID: "fixture-agent", Kind: "agent"}, f.experienceID, scope)
		var serviceErr *apierrors.ServiceError
		if !errors.As(err, &serviceErr) || serviceErr.Code != apierrors.EvidenceMissing {
			t.Fatalf("expected raw-free evidence blocker, got %v", err)
		}
	})
	t.Run("sleep_literal_projector", func(t *testing.T) {
		facts, err := app.LiteralSleepFactsMap(row.Event.Raw)
		if err != nil || len(facts) != 0 {
			t.Fatalf("expected no facts from absent payload: %v", err)
		}
	})
}

func TestV0DerivedIntegrityBridge(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	derived, err := json.Marshal(map[string]any{"event_id": f.row.ID, "kind": f.row.Event.Kind, "topic": topic, "tokens": f.row.Event.Tokens, "scope": scope, "ts": ts})
	if err != nil {
		t.Fatal(err)
	}
	// Disposable DB experiment only: canonical derived bytes replace the legacy
	// raw slots; the proposed sidecar must retain the original independently.
	if _, err = f.db.Conn().Exec("UPDATE attention_engine_events SET data=json_set(data,'$.raw',json(?)),raw_bytes=? WHERE id=?", string(derived), derived, f.row.ID); err != nil {
		t.Fatal(err)
	}
	row, err := f.repo.GetEvent(ctx, f.row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(row.RawBytes, derived) || !bytes.Equal(row.Event.Raw, derived) {
		t.Fatal("canonical derived bridge did not roundtrip")
	}
	captured, err := f.experiences.CaptureFailure(ctx, f.row.ID)
	if err != nil || captured == nil {
		t.Fatalf("derived integrity bytes failed capture checks: %v", err)
	}
	resolved, err := f.experiences.ResolveExperienceEvidence(ctx, app.Principal{ID: "fixture-agent", Kind: "agent"}, f.experienceID, scope)
	if err != nil || len(resolved) != 1 {
		t.Fatalf("derived integrity bytes failed evidence checks: %v", err)
	}
	explanation, err := f.engine.ExplainFocus(ctx, topic)
	if err != nil || len(explanation.Evidence) != 1 {
		t.Fatalf("derived integrity bytes failed dynamics read: %v", err)
	}
	wire, err := json.Marshal(explanation)
	if err != nil {
		t.Fatal(err)
	}
	forbidContent(t, wire, f.row.RawBytes)
	before, err := app.LiteralSleepFactsMap(f.row.RawBytes)
	if err != nil || len(before) != 1 {
		t.Fatalf("expected original literal fact: %v", err)
	}
	after, err := app.LiteralSleepFactsMap(row.Event.Raw)
	if err != nil || len(after) != 0 {
		t.Fatalf("expected derived bridge to carry no literal facts: %v", err)
	}
	t.Log("canonical derived JSON passes legacy Capture/Resolve checks; original literal facts become unavailable (1 -> 0), raw property names still require MCP projection")
}
