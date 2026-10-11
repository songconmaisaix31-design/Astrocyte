package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestBuiltDailySleepCommandAndReplay(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "astrocyte-sleep.exe")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	data := t.TempDir()
	path := filepath.Join(data, "state.sqlite")
	ctx := context.Background()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	day := time.Now().UTC().AddDate(0, 0, -1)
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	var input []app.StoredEvent
	for i := 0; i < 2; i++ {
		ts := day.Add(time.Duration(9+i) * time.Hour).UnixMilli()
		raw := []byte(` { "fixture": "literal-source", "facts": { "method": "literal" } } `)
		input = append(input, app.StoredEvent{OriginKey: day.Format(time.DateOnly) + string(rune('a'+i)), ReceivedAt: ts, RawBytes: raw, Observation: app.Observation{Source: "manual", SourceKind: "collect", Session: "fixture", Scope: "project-a", TS: ts, Outcome: "unknown", Depth: "unknown", Provenance: "contract_local"}, Event: &domain.Event{Source: "manual", Kind: "collect", Scope: "project-a", Session: "fixture", Topic: "fixture-topic", TS: ts, Raw: raw}})
	}
	if _, err = app.NewEventService(sqlite.NewEventRepository(db), .2).Record(ctx, input); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	args := []string{"--data-dir", data, "--day", day.Format(time.DateOnly), "--timezone", "UTC", "--decay", "0.1", "--initial-confidence", "0.5", "--fact-projection", "none"}
	invoke := func() app.SleepResult {
		t.Helper()
		command := exec.Command(binary, args...)
		var out, stderr bytes.Buffer
		command.Stdout, command.Stderr = &out, &stderr
		if err := command.Run(); err != nil {
			t.Fatalf("single day: %v %s", err, stderr.String())
		}
		var result app.SleepResult
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatalf("stdout: %s %v", out.String(), err)
		}
		return result
	}
	first := invoke()
	replayed := invoke()
	if first.Replayed || !replayed.Replayed || first.Run.ID != replayed.Run.ID || first.Run.CompletedAt != replayed.Run.CompletedAt || first.Run.Metrics.RetentionMeasurable || !first.Run.Metrics.StorageMeasurable {
		t.Fatalf("run/replay: %+v %+v", first, replayed)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var runs int
	if err = db.Conn().QueryRow(`SELECT count(*) FROM attention_engine_sleep_runs`).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("daily receipt: %d %v", runs, err)
	}
	nodes, err := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{}).AttentionState(ctx)
	if err != nil || len(nodes) != 1 || nodes[0].Topic != "fixture-topic" {
		t.Fatalf("graph: %+v %v", nodes, err)
	}
	var out, stderr bytes.Buffer
	if err = run(ctx, []string{"--day", day.Format(time.DateOnly)}, &out, &stderr); err == nil {
		t.Fatal("missing explicit policy accepted")
	}
	// The same known source format only becomes measurable through explicit
	// opt-in. Use a separate store so daily replay does not rewrite metrics.
	literalData := t.TempDir()
	literalDB, err := sqlite.Open(filepath.Join(literalData, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = literalDB.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	if _, err = app.NewEventService(sqlite.NewEventRepository(literalDB), .2).Record(ctx, input); err != nil {
		t.Fatal(err)
	}
	if err = literalDB.Close(); err != nil {
		t.Fatal(err)
	}
	args[1] = literalData
	args[len(args)-1] = "facts-map-v1"
	literal := invoke()
	if !literal.Run.Metrics.RetentionMeasurable || literal.Run.Metrics.KeyFactsBefore != 1 || literal.Run.Metrics.KeyFactsRetained != 1 || literal.Run.Metrics.RetentionPrecision != 1 {
		t.Fatalf("literal projection: %+v", literal)
	}
}
