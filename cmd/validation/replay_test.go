package main

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func replayFixture(id, topic string, ts int64) replayInput {
	raw := []byte(`{"synthetic":"PRIVATE_RAW_SENTINEL_must_never_leave_sqlite"}`)
	return replayInput{StableID: id, Row: app.StoredEvent{OriginKey: "fixture:" + id, RawBytes: raw,
		Observation: app.Observation{Source: "codex", SourceKind: "function_call", Scope: "fixture-scope", TS: ts, Provenance: "fixture", Outcome: "unknown", Depth: "unknown"},
		Event:       &domain.Event{Source: "codex", Kind: "edit_file", Topic: topic, Scope: "fixture-scope", TS: ts}}}
}

func readReplayJSON(t *testing.T, path string, target any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "PRIVATE_RAW_SENTINEL") || strings.Contains(string(b), "origin_key") || strings.Contains(string(b), "session") {
		t.Fatal("private source fields escaped derived projection")
	}
	if err := json.Unmarshal(b, target); err != nil {
		t.Fatal(err)
	}
}

func TestV1ProductionSingleEventEmptyTopicClockAndThirtySleeps(t *testing.T) {
	out := t.TempDir()
	start := v1Window().UnixMilli()
	// Deliberately unsorted: two distinct calls at identical ts must each fold
	// independently (.65 then 1), preserving momentum between saturation calls.
	inputs := []replayInput{replayFixture("c", "", start+1000), replayFixture("b", "Morphogenesis", start+1000), replayFixture("a", "Morphogenesis", start+1000), replayFixture("d", "Astrocyte/Attention Engine", start+29*86400000+1234)}
	if err := replayV1Records(inputs, out, map[string]int{}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	var sleeps []replaySleep
	readReplayJSON(t, filepath.Join(out, "sleep-runs-safe.json"), &sleeps)
	if len(sleeps) != 30 || sleeps[0].InputEvents != 3 || sleeps[0].Pending != 1 || sleeps[29].InputEvents != 1 {
		t.Fatalf("daily real inputs or unassigned pending mismatch: %+v", sleeps)
	}
	for i, run := range sleeps {
		end := start + int64(i+1)*86400000 - 1
		if run.StartedAt != end || run.CompletedAt != end || run.Replayed || run.InputSignals != run.InputEvents {
			t.Fatalf("day %d was not a real simulated Sleep", i)
		}
	}
	db, err := sqlite.Open(filepath.Join(out, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.Conn().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM attention_engine_events WHERE topic='' AND kind='edit_file' AND typeof(raw_bytes)='blob'").Scan(&n); err != nil || n != 1 {
		t.Fatalf("actual mapped unassigned BLOB missing: %d %v", n, err)
	}
	var series map[string][]replayPoint
	readReplayJSON(t, filepath.Join(out, "topic-timeseries-safe.json"), &series)
	p := series["Morphogenesis"]
	firstAttention := .05*math.Exp(-.1*1000/86400000) + .6
	if p[1].ReplayID != "a" || p[2].ReplayID != "b" || math.Abs(p[1].Attention-firstAttention) > 1e-12 || p[2].Attention != 1 || p[1].Momentum == p[2].Momentum {
		t.Fatalf("equal-ts events were grouped or reordered: %+v", p[:3])
	}
	var sample map[string]any
	readReplayJSON(t, filepath.Join(out, "weights-50-safe.json"), &sample)
	// End-to-end independent recurrence over the actual production trace.
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("independent Python reference unavailable; production checks above passed")
	}
	command := exec.Command(python, "../../scripts/validation/verify_v1.py", "--trace", filepath.Join(out, "operations-safe.jsonl"), "--output", filepath.Join(out, "independent.json"), "--errors", filepath.Join(out, "errors.json"), "--expected-population", "4")
	command.Env = append(os.Environ(), "PYTHONDONTWRITEBYTECODE=1")
	if b, err := command.CombinedOutput(); err != nil {
		t.Fatalf("independent production-trace reference failed: %v %s", err, b)
	}
}

func TestV1RejectsDuplicateOutOfWindowAndOverwritingEvidence(t *testing.T) {
	start := v1Window().UnixMilli()
	for _, inputs := range [][]replayInput{
		{replayFixture("a", "", start), replayFixture("a", "", start)},
		{replayFixture("a", "", start-1)},
		{replayFixture("a", "", start+30*86400000)},
		{{StableID: "nil-event"}},
	} {
		if err := replayV1Records(inputs, t.TempDir(), nil, time.Now()); err == nil {
			t.Fatal("invalid frozen replay input accepted")
		}
	}
	path := filepath.Join(t.TempDir(), "first-red.json")
	if err := writeExclusive(path, map[string]string{"result": "RED"}); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(path, map[string]string{"result": "PASS"}); err == nil {
		t.Fatal("first evidence was overwritten")
	}
}
