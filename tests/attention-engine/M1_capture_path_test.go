package attention_engine_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1CommittedFailureCaptureUsesOriginalReceipt(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "capture.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewEventRepository(db)
	experiences := app.NewExperienceService(sqlite.NewExperienceRepository(db), repo, app.ExperienceOptions{ProjectID: "trusted-startup-project"})
	var captured *app.StoredExperience
	svc := app.NewEventService(repo, 1, func(ctx context.Context, id string) error {
		var err error
		captured, err = experiences.CaptureFailure(ctx, id)
		return err
	})
	raw := []byte(`{ "type":"agent_task", "outcome":"failed", "source_fact":"contract_local explicit failure" }`)
	input := app.StoredEvent{OriginKey: "fixture:closed-task", RawBytes: raw, ReceivedAt: 1791590400500,
		Observation: app.Observation{Source: "codex", SourceKind: "agent_task", Scope: "trusted-startup-project", TS: 1791590400123, Outcome: "failed", Depth: "unknown", Provenance: "contract_local:explicit_failure"},
		Event:       &domain.Event{Source: "codex", Kind: "agent_task", Scope: "trusted-startup-project", TS: 1791590400123, Raw: raw}}
	result, err := svc.Record(ctx, []app.StoredEvent{input})
	if err != nil || len(result) != 1 || captured == nil || captured.OriginEventID != result[0].ID || captured.Experience.Status != "quarantine" || captured.Experience.Evidence[0].EventID != result[0].ID {
		t.Fatalf("committed capture path: %v", err)
	}
	candidateID := captured.Experience.ID
	result, err = svc.Record(ctx, []app.StoredEvent{input})
	if err != nil || captured == nil || captured.Experience.ID != candidateID {
		t.Fatal("failure replay duplicated candidate")
	}
	// Post-commit errors retain the durable receipt for reconciliation.
	input.OriginKey = "fixture:capture-unavailable"
	result, err = app.NewEventService(repo, 1, func(context.Context, string) error { return errors.New("capture unavailable") }).Record(ctx, []app.StoredEvent{input})
	if err == nil || len(result) != 1 {
		t.Fatal("post-commit capture error obscured committed receipt")
	}
	var pending *app.EventCapturePendingError
	if !errors.As(err, &pending) || len(pending.ReceiptIDs) != 1 || pending.ReceiptIDs[0] != result[0].ID {
		t.Fatal("capture-pending outcome lost typed receipt identity")
	}
	if _, err = repo.GetEvent(ctx, result[0].ID); err != nil {
		t.Fatal("post-commit failure rolled back original evidence")
	}
}
