package privacy_test

import (
	"bufio"
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

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// This is a real_raw_contract_local exposure test, never native eligibility,
// behavioral replay or task_live acceptance. Original frozen bytes are mapped
// independently, persisted unchanged, and attached to an explicitly synthetic
// failure signal so every payload actually reaches production ExplainFocus.
// No externally supplied response/scanner can stand in for MCP.Serve here.
func TestReal100V21ProductionExposure(t *testing.T) {
	if os.Getenv("ASTROCYTE_V21_REAL100_AUDIT") != "1" {
		t.Skip("NOT_RUN: ASTROCYTE_V21_REAL100_AUDIT=1 reads authorized frozen local evidence only")
	}
	const root = "C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012"
	openEvidence := func(relative string) *os.File {
		t.Helper()
		base, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Fatal("authorized evidence root unavailable; details suppressed")
		}
		resolved, err := filepath.EvalSymlinks(filepath.Join(base, relative))
		if err != nil {
			t.Fatal("frozen evidence unavailable; details suppressed")
		}
		rel, err := filepath.Rel(base, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			t.Fatal("frozen evidence path leaves authorized root")
		}
		file, err := os.Open(resolved)
		if err != nil {
			t.Fatal("frozen evidence open failed; details suppressed")
		}
		t.Cleanup(func() { file.Close() })
		return file
	}
	var associations []struct {
		Source  string `json:"source"`
		Ordinal int    `json:"ordinal"`
		events.NativeAssociation
	}
	if json.NewDecoder(openEvidence("diagnosis-w1/associations-p0-local.json")).Decode(&associations) != nil || len(associations) != 100 {
		t.Fatal("frozen observed association decode/count failed; contents suppressed")
	}
	contexts := map[string]events.NativeAssociation{}
	for _, association := range associations {
		key := fmt.Sprintf("%s:%d", association.Source, association.Ordinal)
		if _, exists := contexts[key]; exists {
			t.Fatal("duplicate observed association; contents suppressed")
		}
		contexts[key] = association.NativeAssociation
	}
	reader := bufio.NewScanner(openEvidence("diagnosis-w1/sample-first.jsonl"))
	reader.Buffer(make([]byte, 64*1024), 64<<20)
	seen := map[string]bool{}
	counts := map[string]int{}
	responses := map[string]int{}
	for reader.Scan() {
		var sample struct {
			Source    string `json:"source"`
			Ordinal   int    `json:"ordinal"`
			Project   string `json:"project"`
			RawBase64 string `json:"raw_base64"`
		}
		if json.Unmarshal(reader.Bytes(), &sample) != nil || (sample.Source != "codex" && sample.Source != "claude_code") || sample.Ordinal < 0 || sample.Ordinal >= 50 {
			t.Fatal("frozen sample envelope invalid; contents suppressed")
		}
		key := fmt.Sprintf("%s:%d", sample.Source, sample.Ordinal)
		association, exists := contexts[key]
		if seen[key] || !exists {
			t.Fatal("frozen sample identity/association invalid; contents suppressed")
		}
		seen[key] = true
		raw, err := base64.StdEncoding.DecodeString(sample.RawBase64)
		if err != nil || !json.Valid(raw) {
			t.Fatal("frozen sample original bytes invalid; contents suppressed")
		}
		counts["original_events"]++
		counts["source_"+sample.Source]++
		if association.Session == "" || association.Scope == "" {
			counts["association_incomplete"]++
		}
		native, err := events.MapAssociatedNativeRecord(sample.Source, "privacy-native-probe:"+key, sample.Project, raw, 0, association)
		if err != nil {
			t.Fatal("actual native mapping rejected frozen input; contents suppressed")
		}
		counts["native_mapping_pass"]++
		if native.Event != nil {
			counts["native_kind_mapped"]++
			if native.Event.Topic != "" {
				counts["native_topic_assigned"]++
			}
		} else {
			counts["native_unmapped"]++
		}
		// Numeric names disclose neither sessions/paths nor source bodies.
		t.Run(fmt.Sprintf("event_%03d", counts["original_events"]), func(t *testing.T) {
			f := newFixture(t)
			ctx := context.Background()
			// Retain the actual mapped/pending receipt as observed, without
			// manufacturing topic/outcome or changing native mapping rules.
			var original app.StoredEvent
			if err := f.repo.WithEventTx(ctx, func(tx app.EventTx) error {
				var err error
				original, err = tx.PutEvent(native)
				return err
			}); err != nil || !bytes.Equal(original.RawBytes, raw) {
				t.Fatal("actual native receipt byte preservation failed; contents suppressed")
			}
			// Deliberately separate from the native receipt. The synthetic
			// kind/topic/outcome/time are exposure stress input, not new facts.
			contract := app.StoredEvent{OriginKey: "real_raw_contract_local:" + key, ReceivedAt: ts + 1, RawBytes: raw,
				Observation: app.Observation{Source: sample.Source, SourceKind: "agent_task", Session: association.Session, Scope: scope, TS: ts, Outcome: "failed", Depth: "unknown", Provenance: "real_raw_contract_local"},
				Event:       &domain.Event{Source: sample.Source, Session: association.Session, Kind: "agent_task", Topic: topic, Scope: scope, TS: ts}}
			stored, err := f.events.Record(ctx, []app.StoredEvent{contract})
			if err != nil || len(stored) != 1 || !bytes.Equal(stored[0].RawBytes, raw) {
				t.Fatal("real_raw_contract_local persistence failed; contents suppressed")
			}
			counts["contract_local_signals"]++
			signals, err := f.repo.ListSignals(ctx, ts, ts)
			if err != nil || len(signals) != 2 {
				t.Fatal("isolated contract signal read failed; contents suppressed")
			}
			if _, err = f.engine.ApplySignals(ctx, topic, signals, ts); err != nil {
				t.Fatal("isolated contract dynamics fold failed; contents suppressed")
			}
			candidate, err := f.experiences.CaptureFailure(ctx, stored[0].ID)
			if err != nil || candidate == nil {
				t.Fatal("isolated structured experience capture failed; contents suppressed")
			}
			human := app.Principal{ID: "fixture-human", Kind: "human"}
			if _, err = f.experiences.ApproveExperience(ctx, human, candidate.Experience.ID, app.ApproveExperienceCommand{OwnerProject: scope,
				CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "fixture-realraw-approval", ExpectedVersion: 1, IdempotencyKey: "fixture-realraw-approval"}}); err != nil {
				t.Fatal("trusted human experience approval failed; contents suppressed")
			}
			values := rawUnique(t, raw, derivedEvent(stored[0]))
			counts["raw_unique_strings"] += len(values)
			if len(values) == 0 {
				counts["zero_raw_unique_events"]++
			}
			cause := errors.New(strings.Join(values, " | ")) // Local controlled failure only.
			realArgs := func(name string) any {
				if name == "report_event" {
					return map[string]any{"event": map[string]any{"kind": "agent_task", "topic": topic, "scope": scope, "ts": ts + 20, "raw": json.RawMessage(raw)}}
				}
				if name == "veto" {
					return map[string]any{"topic": topic, "reason": string(raw)}
				}
				return toolArgs(name)
			}
			auditResponse := func(wire []byte, name, path string, failed bool) {
				assertState(t, wire, failed)
				if indexes, _ := rawUniqueLeakIndexes(wire, values); len(indexes) > 0 {
					t.Logf("raw_unique collision event=%d tool=%s path=%s indexes=%v; values suppressed", counts["original_events"], name, path, indexes)
					// Classify against fixed public field names, never print an
					// original value, body, parameter, session or filename.
					for _, index := range indexes {
						for _, field := range []string{"event_id", "kind", "topic", "tokens", "scope", "ts", "evidence", "original", "attention", "momentum", "strength", "weight", "elapsed_days", "decay", "decay_factor", "class", "attention_contribution", "confidence_contribution", "confidence_only", "related", "related_projects", "updated_at", "experience", "pattern", "condition", "verdict", "status", "owner_project", "origin_event_id", "version", "created_at", "accepted", "target", "pending", "capture_pending", "dynamics_pending", "delivery_state", "receipt_ids", "recovery", "content", "isError", "structuredContent"} {
							if values[index] == field {
								t.Logf("raw_unique collision matches public field name: %s; original contents suppressed", field)
							}
						}
					}
				}
				assertRawUniqueAbsent(t, wire, values, raw)
				// All new report/veto raw receipts are also audited individually;
				// original native TS may lie outside the fixture clock range.
				auditV21Rows(t, f, wire)
				assertRawUniqueAbsent(t, wire, rawUnique(t, original.RawBytes, derivedEvent(original)), original.RawBytes)
				responses[path+":"+name]++
			}
			for _, name := range tools {
				b := binding(name == "veto")
				backend := f.backend(t, b, f.events)
				wire := call(t, backend, b, name, realArgs(name))
				auditResponse(wire, name, "success", false)
				wire = call(t, rawErrorBackend{MCPBackend: backend, cause: cause}, b, name, realArgs(name))
				auditResponse(wire, name, "backend_error", true)
				unknownKey := "FIXTURE_V21_UNKNOWN_ARGUMENT_23480"
				if len(values) > 0 {
					unknownKey = values[0]
				}
				wire = call(t, backend, b, name, map[string]any{unknownKey: true})
				auditResponse(wire, name, "decode_error", true)
			}
			for _, name := range []string{"report_event", "veto"} {
				for _, mode := range []string{"capture", "dynamics", "unknown"} {
					b := binding(name == "veto")
					backend := f.backend(t, b, v21ErrorRecorder{f.events, mode, cause})
					wire := call(t, backend, b, name, realArgs(name))
					auditResponse(wire, name, mode+"_error", true)
					payload := toolPayload(t, wire)
					if mode != "unknown" && (payload["delivery_state"] != "committed" || payload[mode+"_pending"] != true || payload["event_id"] == nil) {
						t.Error("real-raw controlled committed error lost delivery facts")
					}
				}
			}
		})
	}
	if reader.Err() != nil || len(seen) != 100 || counts["source_codex"] != 50 || counts["source_claude_code"] != 50 || counts["contract_local_signals"] != 100 {
		t.Fatal("frozen real100 read/coverage incomplete; contents suppressed")
	}
	if counts["raw_unique_strings"] == 0 {
		t.Fatal("real100 audit has no measurable raw_unique strings")
	}
	// Only fixed labels and aggregate counts enter stdout/logs or this report.
	countJSON, _ := json.Marshal(counts)
	coverageJSON, _ := json.Marshal(responses)
	t.Logf("real_raw_contract_local counts=%s production_mcp_responses=%s; native mapping probe is not eligibility/task_live", countJSON, coverageJSON)
}
