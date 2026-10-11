package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1ReceiptSignalRestartAndIdempotentReplay(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "events.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	input := []app.StoredEvent{}
	for i, kind := range []string{"read_paper", "edit_file", "collect", "user_approve", "user_veto", "agent_task", "agent_task"} {
		raw := []byte(fmt.Sprintf(" { \"kind\" : %q, \"selected_test_fact\" : \"preserve bytes\" } \n", kind))
		outcome, depth := "unknown", "unknown"
		if i == 0 {
			depth = "deep"
		}
		if i == 5 {
			outcome = "success"
		}
		if i == 6 {
			outcome = "failure"
		}
		input = append(input, app.StoredEvent{OriginKey: fmt.Sprintf("fixture:%d", i), ReceivedAt: 1791590400500, RawBytes: raw,
			Observation: app.Observation{Source: "manual", SourceKind: kind, Scope: "selected-project", TS: 1791590400123, Outcome: outcome, Depth: depth, Provenance: "contract_local:explicit_weight_sample"},
			Event:       &domain.Event{Source: "manual", Kind: kind, Scope: "selected-project", Topic: "selected-topic", TS: 1791590400123, Raw: raw}})
	}
	job, err := events.MapS1Outbox(json.RawMessage(`{"id":"import-failure","type":"attention.job_failed","aggregate_id":"job","aggregate_version":1,"occurred_at":"2026-10-10T00:00:00.123Z","correlation_id":"request","causation_id":"cause","payload":{"cause":{"code":"source_unavailable"}}}`), 1791590400500)
	if err != nil {
		t.Fatal(err)
	}
	if job.Event != nil || job.Observation.SourceKind != "attention.job_failed" || job.Observation.Outcome != "failure" {
		t.Fatal("S1 failed import became agent_task")
	}
	input = append(input, job)
	repo := sqlite.NewEventRepository(db)
	svc := app.NewEventService(repo, 1)
	first, err := svc.Record(ctx, input)
	if err != nil || len(first) != len(input) {
		t.Fatalf("receipt accounting: %d/%d %v", len(first), len(input), err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	repo = sqlite.NewEventRepository(db)
	svc = app.NewEventService(repo, 1)
	replayed, err := svc.Record(ctx, input)
	if err != nil || len(replayed) != len(first) {
		t.Fatalf("restart replay: %v", err)
	}
	for i, v := range replayed {
		if v.ID != first[i].ID || !bytes.Equal(v.RawBytes, input[i].RawBytes) || v.Observation.Tokens != nil {
			t.Fatal("receipt ID/raw bytes/unknown token volume changed")
		}
		if v.Event != nil && !bytes.Equal(v.Event.Raw, input[i].RawBytes) {
			t.Fatal("Event.Raw differs from original BLOB")
		}
		var storage string
		if err = db.Conn().QueryRowContext(ctx, "SELECT typeof(raw_bytes) FROM attention_engine_events WHERE id=?", v.ID).Scan(&storage); err != nil || storage != "blob" {
			t.Fatal("original raw bytes are not BLOB")
		}
	}
	all, err := repo.ListEvents(ctx, 0, math.MaxInt64)
	if err != nil || len(all) != len(input) {
		t.Fatal("idempotent replay duplicated events")
	}
	signals, err := repo.ListSignals(ctx, 0, math.MaxInt64)
	if err != nil || len(signals) != 7 {
		t.Fatalf("stored signals: count=%d err=%v", len(signals), err)
	}
	if err = sqlite.NewEngineRepository(db).WithTx(ctx, func(tx sqlite.EngineTx) error {
		eventTx := sqlite.NewEventTx(tx)
		rows, e := eventTx.ListEvents(0, math.MaxInt64)
		if e != nil {
			return e
		}
		ss, e := eventTx.ListSignals(0, math.MaxInt64)
		if e != nil {
			return e
		}
		if len(rows) != len(input) || len(ss) != 7 {
			return fmt.Errorf("shared transaction read differs")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Replay cannot change calibration or mutate original source evidence.
	if _, err = app.NewEventService(repo, 2).Record(ctx, input); err == nil {
		t.Fatal("replay changed strength")
	}
	candidate := input[0]
	candidate.OriginKey = "must-rollback"
	candidate.Event = nil
	changed := input[1]
	changed.RawBytes = []byte(`{"different":true}`)
	if _, err = svc.Record(ctx, []app.StoredEvent{candidate, changed}); err == nil {
		t.Fatal("source rewrite accepted")
	}
	all, err = repo.ListEvents(ctx, 0, math.MaxInt64)
	if err != nil || len(all) != len(input) {
		t.Fatal("failed batch did not roll back new receipt")
	}
	if err = repo.WithEventTx(ctx, func(tx app.EventTx) error {
		return tx.PutSignal(domain.Signal{EventID: first[len(first)-1].ID, Class: domain.SignalFeedbackFailure, Weight: -.1, Strength: 1})
	}); err == nil {
		t.Fatal("pending observation produced a signal")
	}
}
