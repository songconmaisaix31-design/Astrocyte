package privacy_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/mcp"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// v2.1 audits exposure, not the superseded V0 raw-table isolation proposal.
// All synthetic probes run by default; no opt-in skip counts as acceptance.
const privateSession = "FIXTURE_V21_SESSION_PRIVATE_41d982"

func newV21Fixture(t *testing.T) (*fixture, app.MCPBinding) {
	t.Helper()
	f := newFixture(t)
	var raw map[string]any
	if json.Unmarshal(f.row.RawBytes, &raw) != nil {
		t.Fatal("fixture raw decode failed")
	}
	raw["body"] = "FIXTURE_V21_FULL_BODY_PRIVATE_b30217"
	raw["parameters"] = `{"command":"FIXTURE_V21_FULL_PARAMETERS_PRIVATE_40a328","args":["private"]}`
	raw["attachment"] = base64.StdEncoding.EncodeToString([]byte("FIXTURE_V21_ATTACHMENT_PRIVATE_38290e"))
	raw["session"] = privateSession
	raw["session_meta"] = map[string]any{"cwd": "FIXTURE_V21_CWD_PRIVATE_a17e02", "title": "FIXTURE_V21_TITLE_PRIVATE_0291ac", "actor": "fixture-human", "request": "fixture-approval"}
	raw["event_id"] = f.row.ID // Exact derived-value reuse is permitted.
	encoded, err := json.Marshal(raw)
	if err != nil {
		t.Fatal("fixture raw encode failed")
	}
	// Update only a disposable actual SQLite database, never the production store.
	if _, err = f.db.Conn().Exec("UPDATE attention_engine_events SET data=json_set(data,'$.raw',json(?)),raw_bytes=? WHERE id=?", string(encoded), encoded, f.row.ID); err != nil {
		t.Fatal("fixture storage update failed")
	}
	f.row, err = f.repo.GetEvent(context.Background(), f.row.ID)
	if err != nil {
		t.Fatal("fixture storage read failed")
	}
	b := binding(false)
	b.Session = privateSession
	return f, b
}

// rawUnique excludes exact values of the six permitted event fields. It does
// not whitelist message substrings, protocol scalar 0/1/true, or arbitrary keys.
func rawUnique(t *testing.T, raw json.RawMessage, derived map[string]any) []string {
	t.Helper()
	allowed := map[string]bool{}
	for _, key := range []string{"event_id", "kind", "topic", "tokens", "scope", "ts"} {
		if value, ok := derived[key]; ok && value != nil {
			allowed[fmt.Sprint(value)] = true
		}
	}
	var payload any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if d.Decode(&payload) != nil {
		t.Fatal("audit raw JSON decode failed; contents suppressed")
	}
	values := map[string]bool{}
	var visit func(any)
	visit = func(value any) {
		switch value := value.(type) {
		case string:
			if len(value) >= 8 && !allowed[value] {
				values[value] = true
			}
		case map[string]any:
			for _, child := range value {
				visit(child)
			}
		case []any:
			for _, child := range value {
				visit(child)
			}
		}
	}
	visit(payload)
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func derivedEvent(row app.StoredEvent) map[string]any {
	derived := map[string]any{"event_id": row.ID}
	if row.Event != nil {
		derived["kind"], derived["topic"], derived["tokens"], derived["scope"], derived["ts"] = row.Event.Kind, row.Event.Topic, row.Event.Tokens, row.Event.Scope, row.Event.TS
	}
	return derived
}

func rawUniqueLeakIndexes(wire []byte, values []string) ([]int, error) {
	texts := []string{string(wire)}
	// MCP text contains another JSON document. Check decoded string values too,
	// so escaped full parameter JSON/unicode cannot evade the string assertion.
	var visit func(any)
	visit = func(value any) {
		switch value := value.(type) {
		case string:
			texts = append(texts, value)
			var nested any
			if json.Unmarshal([]byte(value), &nested) == nil {
				visit(nested)
			}
		case map[string]any:
			for key, child := range value {
				texts = append(texts, key)
				visit(child)
			}
		case []any:
			for _, child := range value {
				visit(child)
			}
		}
	}
	var envelope any
	if json.Unmarshal(wire, &envelope) != nil {
		return nil, errors.New("audit response JSON decode failed; contents suppressed")
	}
	visit(envelope)
	var indexes []int
	for index, value := range values {
		for _, text := range texts {
			if strings.Contains(text, value) {
				indexes = append(indexes, index)
				break
			}
		}
	}
	return indexes, nil
}

func assertRawUniqueAbsent(t *testing.T, wire []byte, values []string, raw []byte) {
	t.Helper()
	indexes, err := rawUniqueLeakIndexes(wire, values)
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range indexes {
		t.Errorf("raw_unique value index=%d appears in exposure; contents suppressed", index)
	}
	if len(raw) > 0 && bytes.Contains(wire, []byte(base64.StdEncoding.EncodeToString(raw))) {
		t.Error("complete original raw bytes appear as base64; contents suppressed")
	}
}

func auditV21Rows(t *testing.T, f *fixture, wire []byte) {
	t.Helper()
	rows, err := f.repo.ListEvents(context.Background(), ts-1, ts+100)
	if err != nil || len(rows) == 0 {
		t.Fatal("audit persisted receipt enumeration failed")
	}
	for _, row := range rows {
		values := rawUnique(t, row.RawBytes, derivedEvent(row))
		if len(values) == 0 {
			t.Fatal("audit fixture unexpectedly has no raw_unique values")
		}
		assertRawUniqueAbsent(t, wire, values, row.RawBytes)
	}
}

func v21Args(t *testing.T, f *fixture, name string) any {
	t.Helper()
	if name == "report_event" {
		var raw map[string]any
		if json.Unmarshal(f.row.RawBytes, &raw) != nil {
			t.Fatal("fixture request decode failed")
		}
		// A new report has its own derived identity/kind. Copy only the
		// distinct private fixture payload, not another event's derived IDs.
		for _, key := range []string{"event_id", "kind", "topic", "tokens", "scope", "ts"} {
			delete(raw, key)
		}
		return map[string]any{"event": map[string]any{"kind": "search", "topic": topic, "scope": scope, "ts": ts + 20, "source": "FIXTURE_V21_SOURCE_PRIVATE_ad4812", "session": privateSession, "raw": raw}}
	}
	return toolArgs(name)
}

func toolPayload(t *testing.T, wire []byte) map[string]any {
	t.Helper()
	var envelope struct {
		Result struct {
			Content []struct{ Text string }
		}
	}
	if json.Unmarshal(wire, &envelope) != nil || len(envelope.Result.Content) != 1 {
		t.Fatal("missing MCP text payload; contents suppressed")
	}
	var payload map[string]any
	if json.Unmarshal([]byte(envelope.Result.Content[0].Text), &payload) != nil {
		t.Fatal("invalid MCP text payload JSON; contents suppressed")
	}
	return payload
}

func TestV21ExposureSuccess(t *testing.T) {
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			f, b := newV21Fixture(t)
			b = v21Human(b, name)
			wire := call(t, f.backend(t, b, f.events), b, name, v21Args(t, f, name))
			assertState(t, wire, false)
			auditV21Rows(t, f, wire)
			if name == "explain_focus" {
				payload := toolPayload(t, wire)
				evidence, ok := payload["evidence"].([]any)
				if !ok || len(evidence) != 1 {
					t.Fatal("lost explanation evidence")
				}
				signal := evidence[0].(map[string]any)
				original := signal["event"].(map[string]any)
				if _, exists := signal["original"]; exists {
					t.Error("legacy original wrapper exposed instead of derived event")
				}
				if len(original) != 6 {
					t.Error("event projection must contain exactly six permitted fields")
				}
				for key, value := range derivedEvent(f.row) {
					actual, _ := json.Marshal(original[key])
					expected, _ := json.Marshal(value)
					if !bytes.Equal(actual, expected) {
						t.Errorf("lost derived event field %s", key)
					}
				}
				before, err := f.engine.ExplainFocus(context.Background(), topic)
				if err != nil {
					t.Fatal("core explanation read failed")
				}
				beforeJSON, _ := json.Marshal(before)
				var metrics map[string]any
				json.Unmarshal(beforeJSON, &metrics)
				beforeSignal := metrics["evidence"].([]any)[0].(map[string]any)
				for key, value := range beforeSignal {
					if key != "original" && fmt.Sprint(signal[key]) != fmt.Sprint(value) {
						t.Errorf("explanation metric changed: %s", key)
					}
				}
				for key, value := range metrics {
					if key != "evidence" && key != "related_agents" && fmt.Sprint(payload[key]) != fmt.Sprint(value) {
						t.Errorf("explanation state changed: %s", key)
					}
				}
				if len(before.Evidence[0].Original.RawBytes) == 0 {
					t.Error("MCP projection mutated core original receipt")
				}
				if _, exists := payload["related_agents"]; exists {
					t.Error("explanation exposed source-agent metadata")
				}
			}
			if name == "get_experiences" {
				payload := toolPayload(t, wire)
				rows, ok := payload["experiences"].([]any)
				if !ok || len(rows) != 1 {
					t.Fatal("lost approved structured experience")
				}
				row := rows[0].(map[string]any)
				if _, present := row["actions"]; present {
					t.Error("internal action audit metadata serialized")
				}
				if _, present := row["shares"]; present {
					t.Error("internal share audit metadata serialized")
				}
				before, err := f.experiences.GetExperiences(context.Background(), b.Principal, b.Scope, "agent_task")
				if err != nil || len(before) != 1 {
					t.Fatal("approved experience source read failed")
				}
				encoded, _ := json.Marshal(before[0].Experience)
				var expected map[string]any
				json.Unmarshal(encoded, &expected)
				if fmt.Sprint(row["experience"]) != fmt.Sprint(expected) || row["origin_event_id"] != f.row.ID {
					t.Error("approved experience capability or evidence reference changed")
				}
			}
		})
	}
}

func v21Human(b app.MCPBinding, name string) app.MCPBinding {
	if name == "veto" {
		b.Principal, b.CanVeto = app.Principal{ID: "fixture-human", Kind: "human"}, true
	}
	return b
}

func TestV21ExposureBackendErrors(t *testing.T) {
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			f, b := newV21Fixture(t)
			b = v21Human(b, name)
			cause := errors.New(strings.Join(rawUnique(t, f.row.RawBytes, derivedEvent(f.row)), " | "))
			backend := rawErrorBackend{MCPBackend: f.backend(t, b, f.events), cause: cause}
			wire := call(t, backend, b, name, v21Args(t, f, name))
			assertState(t, wire, true)
			auditV21Rows(t, f, wire)
		})
	}
}

type v21ErrorRecorder struct {
	events *app.EventService
	mode   string
	cause  error
}

func (r v21ErrorRecorder) Record(ctx context.Context, rows []app.StoredEvent) ([]app.StoredEvent, error) {
	if r.mode == "unknown" {
		return nil, r.cause
	}
	stored, err := r.events.Record(ctx, rows)
	if err != nil {
		return stored, err
	}
	if r.mode == "capture" {
		return stored, &app.EventCapturePendingError{ReceiptIDs: []string{stored[0].ID}, Cause: r.cause}
	}
	return stored, &app.MCPDynamicsPendingError{Cause: r.cause}
}

func TestV21ExposureDeliveryErrors(t *testing.T) {
	for _, name := range []string{"report_event", "veto"} {
		for _, mode := range []string{"capture", "dynamics", "wrapped_invalid", "unknown"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				f, b := newV21Fixture(t)
				b = v21Human(b, name)
				var cause error = errors.New(strings.Join(rawUnique(t, f.row.RawBytes, derivedEvent(f.row)), " | "))
				if mode == "wrapped_invalid" {
					cause = &mcp.InvalidParamsError{Message: privateSession}
				}
				recorder := v21ErrorRecorder{f.events, mode, cause}
				wire := call(t, f.backend(t, b, recorder), b, name, v21Args(t, f, name))
				assertState(t, wire, true)
				auditV21Rows(t, f, wire)
				payload := toolPayload(t, wire)
				var envelope struct {
					Result struct {
						Structured map[string]any `json:"structuredContent"`
					}
				}
				json.Unmarshal(wire, &envelope)
				if fmt.Sprint(payload) != fmt.Sprint(envelope.Result.Structured) {
					t.Error("text and structured delivery facts differ")
				}
				if mode == "unknown" {
					if payload["delivery_state"] != "delivery_unknown" || payload["event_id"] != nil {
						t.Error("unknown delivery incorrectly asserts a commit")
					}
				} else {
					id, ok := payload["event_id"].(string)
					if !ok || id == "" || payload["delivery_state"] != "committed" {
						t.Fatal("committed receipt lost")
					}
					if _, err := f.repo.GetEvent(context.Background(), id); err != nil {
						t.Error("reported committed event_id does not resolve")
					}
					flag := mode
					if flag == "wrapped_invalid" {
						flag = "dynamics"
					}
					if payload[flag+"_pending"] != true {
						t.Errorf("lost %s pending flag", mode)
					}
					if ids, ok := payload["receipt_ids"].([]any); !ok || len(ids) != 1 || ids[0] != id {
						t.Error("lost original receipt_ids")
					}
				}
				if recovery, ok := payload["recovery"].(string); !ok || !strings.Contains(recovery, "reconcile") || !strings.Contains(recovery, "do not resubmit") {
					t.Error("lost safe reconciliation guidance")
				}
			})
		}
	}
}

func TestV21TaskVetoExposure(t *testing.T) {
	for _, mode := range []string{"success", "dynamics", "unknown"} {
		t.Run(mode, func(t *testing.T) {
			f, b := newV21Fixture(t)
			b = v21Human(b, "veto")
			var recorder app.MCPEventRecorder = f.events
			if mode != "success" {
				recorder = v21ErrorRecorder{f.events, mode, errors.New(privateSession)}
			}
			wire := call(t, f.backend(t, b, recorder), b, "veto", map[string]any{"task": "FIXTURE_V21_TASK_PRIVATE_29183a", "reason": contentCanary})
			assertState(t, wire, mode != "success")
			auditV21Rows(t, f, wire)
			payload := toolPayload(t, wire)
			if payload["target"] != "" || payload["pending"] != true {
				t.Error("task-only veto must retain pending state without exposing raw target")
			}
			if bytes.Contains(wire, []byte("FIXTURE_V21_TASK_PRIVATE_29183a")) {
				t.Error("raw task appears in exposure; contents suppressed")
			}
		})
	}
}

func TestV21ExposureDecodeErrors(t *testing.T) {
	for _, name := range tools {
		for _, mode := range []string{"unknown_field", "wrong_type"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				f, b := newV21Fixture(t)
				b = v21Human(b, name)
				args := map[string]any{privateSession: strings.Join(rawUnique(t, f.row.RawBytes, derivedEvent(f.row)), " | ")}
				if mode == "wrong_type" {
					switch name {
					case "get_attention_state":
						args = map[string]any{privateSession: 1}
					case "explain_focus", "veto":
						args = map[string]any{"topic": map[string]any{privateSession: privateSession}}
					case "get_experiences":
						args = map[string]any{"pattern": map[string]any{privateSession: privateSession}}
					case "report_event":
						args = map[string]any{"event": map[string]any{"kind": map[string]any{privateSession: privateSession}, "ts": ts}}
					}
				}
				wire := call(t, f.backend(t, b, f.events), b, name, args)
				assertState(t, wire, true)
				auditV21Rows(t, f, wire)
			})
		}
	}
}

func TestV21RawUniqueAuditSemantics(t *testing.T) {
	f, _ := newV21Fixture(t)
	values := rawUnique(t, f.row.RawBytes, derivedEvent(f.row))
	if len(values) != 10 {
		t.Fatalf("unexpected raw_unique set size: %d", len(values))
	}
	for _, value := range values {
		if value == topic || value == scope || value == f.row.ID || value == "agent_task" {
			t.Error("authorized derived reuse incorrectly counted as raw_unique")
		}
	}
	// Prove a quoted JSON parameter value is caught after decoding MCP text.
	parameter := `{"arg":"PRIVATE_PARAMETER_AUDIT_84922"}`
	text, _ := json.Marshal(map[string]any{"leak": parameter})
	wire, _ := json.Marshal(map[string]any{"result": map[string]any{"content": []any{map[string]any{"type": "text", "text": string(text)}}}})
	indexes, err := rawUniqueLeakIndexes(wire, []string{parameter})
	if err != nil || len(indexes) != 1 {
		t.Fatal("audit failed to detect escaped full parameter JSON")
	}
	shortAndShared := rawUnique(t, json.RawMessage(`{"short":"abcdefg","integer":0,"boolean":true,"shared":"fixture-topic-w0"}`), map[string]any{"topic": topic})
	if len(shortAndShared) != 0 {
		t.Error("audit incorrectly forbids short strings, generic scalars, or shared derived values")
	}
}

func TestV21NativeSourceMetadata(t *testing.T) {
	for _, sourcePath := range []string{"native_source", "reported_source"} {
		t.Run(sourcePath, func(t *testing.T) {
			f, b := newV21Fixture(t)
			backend := f.backend(t, b, f.events)
			if sourcePath == "native_source" {
				// This is actual native-style EventService/SQLite ingestion, not
				// a fabricated backend explanation or test-only source override.
				var raw map[string]any
				json.Unmarshal(f.row.RawBytes, &raw)
				delete(raw, "event_id")
				raw["source"], raw["ts"] = "claude_code", ts+21
				encoded, _ := json.Marshal(raw)
				rows, err := f.events.Record(context.Background(), []app.StoredEvent{{OriginKey: "fixture:v21-native-source", ReceivedAt: ts + 21,
					RawBytes: encoded, Observation: app.Observation{Source: "claude_code", SourceKind: "agent_task", Session: privateSession, Scope: scope, TS: ts + 21, Outcome: "failed", Depth: "unknown", Provenance: "contract_local"},
					Event: &domain.Event{Source: "claude_code", Session: privateSession, TS: ts + 21, Kind: "agent_task", Topic: topic, Scope: scope, Tokens: 876543}}})
				if err != nil || len(rows) != 1 {
					t.Fatal("native-style source ingestion failed")
				}
			} else {
				args := v21Args(t, f, "report_event").(map[string]any)
				args["event"].(map[string]any)["source"] = "claude_code"
				wire := call(t, backend, b, "report_event", args)
				assertState(t, wire, false)
				auditV21Rows(t, f, wire)
				id := toolPayload(t, wire)["event_id"].(string)
				row, err := f.repo.GetEvent(context.Background(), id)
				if err != nil || row.Event == nil || row.Event.Source != "mcp" || row.Observation.Source != "mcp" {
					t.Fatal("startup binding failed to replace reported source")
				}
			}
			signals, err := f.repo.ListSignals(context.Background(), ts+20, ts+21)
			if err != nil || len(signals) != 1 {
				t.Fatal("source probe signal read failed")
			}
			if _, err = f.engine.ApplySignals(context.Background(), topic, signals, ts+21); err != nil {
				t.Fatal("source probe dynamics fold failed")
			}
			before, err := f.engine.ExplainFocus(context.Background(), topic)
			if err != nil {
				t.Fatal("native source core explanation failed")
			}
			if sourcePath == "native_source" && !strings.Contains(strings.Join(before.RelatedAgents, " "), "claude_code") {
				t.Fatal("native source reproduction did not reach core related_agents")
			}
			for _, name := range tools {
				t.Run(name, func(t *testing.T) {
					bound := v21Human(b, name)
					wire := call(t, f.backend(t, bound, f.events), bound, name, v21Args(t, f, name))
					assertState(t, wire, false)
					auditV21Rows(t, f, wire)
				})
			}
		})
	}
}
