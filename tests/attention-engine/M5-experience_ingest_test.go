package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM5ExperienceAutomaticCaptureThroughProductionEventStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "experience.sqlite")
	db := experienceDB(t, path)
	defer func() { db.Close() }()
	repo := sqlite.NewEventRepository(db)
	svc := experienceService(db, "A")
	// Explicit reported agent-task receipt is a contract_local input. The
	// production event+signal transaction and post-commit capture are exercised.
	raw := []byte(" \n{\"kind\":\"agent_task\",\"outcome\":\"failure\",\"error\":\"task failed\",\"unknown_field\":true}\r\n")
	failure := app.StoredEvent{
		OriginKey: "contract-local:explicit-task-failure", ReceivedAt: 1800000000001, RawBytes: raw,
		Observation: app.Observation{Source: "mcp", SourceKind: "agent_task", Session: "explicit-task", Scope: "A", TS: 1800000000000, Outcome: "failure", Provenance: "contract_local_reported_task"},
		Event:       &domain.Event{Source: "mcp", Session: "explicit-task", Scope: "A", TS: 1800000000000, Kind: "agent_task", Topic: "task", Raw: raw},
	}
	source := app.NewEventService(repo, 1, svc.CaptureCommittedFailure) // Fixture scale; no product default.
	stored, err := svc.RecordEvents(ctx, source, []app.StoredEvent{failure})
	if err != nil || len(stored) != 1 {
		t.Fatalf("production ingest: %+v %v", stored, err)
	}
	id := stored[0].ID
	var candidateID, state string
	if err = db.Conn().QueryRowContext(ctx, `SELECT id,status FROM attention_engine_experiences WHERE origin_event_id=?`, id).Scan(&candidateID, &state); err != nil || state != "quarantine" {
		t.Fatalf("automatic capture missing: %s %s %v", candidateID, state, err)
	}
	signals, err := repo.ListSignals(ctx, 0, 1900000000000)
	if err != nil || len(signals) != 1 || signals[0].EventID != id || !signals[0].ConfidenceOnly() || signals[0].Weight != -.1 {
		t.Fatalf("original failure signal: %+v %v", signals, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = experienceDB(t, path)
	repo, svc = sqlite.NewEventRepository(db), experienceService(db, "A")
	source = app.NewEventService(repo, 1, svc.CaptureCommittedFailure)
	stored, err = svc.RecordEvents(ctx, source, []app.StoredEvent{failure})
	if err != nil || len(stored) != 1 || stored[0].ID != id {
		t.Fatalf("restart original receipt replay: %+v %v", stored, err)
	}
	var count int
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_experiences`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("automatic capture duplicated: %d %v", count, err)
	}
	record, err := svc.GetExperience(ctx, experienceHuman, candidateID, "A")
	if err != nil || record.Experience.Condition != "outcome=failure" || record.Experience.Evidence[0].EventID != id {
		t.Fatalf("source spelling/evidence: %+v %v", record, err)
	}
	evidence, err := svc.ResolveExperienceEvidence(ctx, experienceHuman, candidateID, "A")
	if err != nil || len(evidence) != 1 || !bytes.Equal(evidence[0].RawBytes, raw) || !bytes.Equal(evidence[0].Event.Raw, raw) {
		t.Fatalf("original bytes: %+v %v", evidence, err)
	}
	// A real S1 outbox shape maps to pending import/distillation evidence, not
	// agent_task. Its failure must not create a second experience candidate.
	outbox := json.RawMessage(`{"id":"import-failure","type":"attention.job_failed","aggregate_id":"job","aggregate_version":1,"occurred_at":"2026-10-11T12:00:00Z","payload":{"project_id":"A","operation_id":"import-op","error":"import failed"}}`)
	job, err := events.MapS1Outbox(outbox, 1800000000001)
	if err != nil {
		t.Fatal(err)
	}
	// Application's trusted project binding is supplied explicitly here; the
	// raw outbox payload remains byte-for-byte original evidence.
	job.Observation.Scope = "A"
	if _, err = svc.RecordEvents(ctx, source, []app.StoredEvent{job}); err != nil {
		t.Fatal(err)
	}
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_experiences`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("S1 job failure captured: %d %v", count, err)
	}
	forged := failure
	forged.OriginKey = "other-project"
	forged.Observation.Scope = "B"
	_, err = svc.RecordEvents(ctx, source, []app.StoredEvent{forged})
	experienceError(t, err, apierrors.ScopeDenied)
}

func TestM5ExperiencePostCommitCaptureFailureRetainsKnownReceiptForReconciliation(t *testing.T) {
	ctx := context.Background()
	db := experienceDB(t, filepath.Join(t.TempDir(), "experience.sqlite"))
	defer db.Close()
	repo := sqlite.NewEventRepository(db)
	svc := experienceService(db, "A")
	raw := []byte(`{"kind":"agent_task","outcome":"failed"}`)
	record := app.StoredEvent{OriginKey: "contract-local:capture-interrupted", ReceivedAt: 1800000000001, RawBytes: raw,
		Observation: app.Observation{Source: "mcp", SourceKind: "agent_task", Scope: "A", TS: 1800000000000, Outcome: "failed", Provenance: "contract_local_capture_failure"},
		Event:       &domain.Event{Source: "mcp", Kind: "agent_task", Scope: "A", Topic: "task", TS: 1800000000000, Raw: raw}}
	interrupted := app.NewEventService(repo, 1, func(context.Context, string) error { return errors.New("capture unavailable") })
	committed, err := svc.RecordEvents(ctx, interrupted, []app.StoredEvent{record})
	if err == nil || len(committed) != 1 || committed[0].ID == "" {
		t.Fatalf("committed receipt hidden by capture failure: %+v %v", committed, err)
	}
	originalID := committed[0].ID
	var count int
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_experiences`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unaccepted capture claimed complete: %d %v", count, err)
	}
	// Explicit reconciliation reuses the committed source ID; the service does
	// not automatically rerun an unknown or failed external agent task.
	ready := app.NewEventService(repo, 1, svc.CaptureCommittedFailure)
	committed, err = svc.RecordEvents(ctx, ready, []app.StoredEvent{record})
	if err != nil || len(committed) != 1 || committed[0].ID != originalID {
		t.Fatalf("reconciliation lost original ID: %+v %v", committed, err)
	}
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_experiences WHERE origin_event_id=?`, originalID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("reconciled candidate: %d %v", count, err)
	}
}
