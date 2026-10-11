package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1ExistingS1ActionOutputPath(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := store.Publish(ctx, []byte("selected test material"))
	if err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewAttentionRepository(db)
	if err = repo.WithTx(ctx, func(tx app.AttentionTx) error {
		return tx.SaveMaterial(app.MaterialDetail{SchemaVersion: 1,
			Material:  app.Material{ID: "m", Version: 1, CurrentRevision: 1, Kind: "paper", Lifecycle: "active"},
			Revisions: []app.MaterialRevision{{MaterialID: "m", Revision: 1, SourceKey: "selected-test-source", ContentDigest: ref, ObjectRef: ref}},
			Uses:      []app.Usage{}, Distillations: []app.Distillation{}}, 0)
	}); err != nil {
		t.Fatal(err)
	}
	now := time.UnixMilli(1791590400123)
	svc := app.NewAttentionService(repo, nil, store, app.ServiceOptions{Clock: func() time.Time { return now }})
	_, err = svc.UpdateMaterial(ctx, app.Principal{ID: "human", Kind: "human"}, "m", app.UpdateMaterialCommand{
		CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "pin-request", IdempotencyKey: "pin-key", ExpectedVersion: 1},
		Lifecycle:   "active", Pinned: true})
	if err != nil {
		t.Fatal(err)
	}
	records, partial, err := events.ReadS1Outbox(ctx, db.Conn(), "", 128)
	if err != nil || partial || len(records) != 1 {
		t.Fatalf("real S1 output path records=%d partial=%v error=%v", len(records), partial, err)
	}
	var envelope struct {
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		Payload    json.RawMessage `json:"payload"`
		OccurredAt time.Time       `json:"occurred_at"`
	}
	if err = json.Unmarshal(records[0], &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.ID == "" || envelope.Type != "attention.material_updated" || envelope.OccurredAt.UnixMilli() != now.UnixMilli() || !bytes.Contains(envelope.Payload, []byte(`"pinned":true`)) {
		t.Fatal("S1 action receipt lost")
	}
	mapped, err := events.MapS1Outbox(records[0], now.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	receipts, err := app.NewEventService(sqlite.NewEventRepository(db), 1).Record(ctx, []app.StoredEvent{mapped})
	if err != nil || len(receipts) != 1 || receipts[0].Event == nil || receipts[0].Event.Kind != "user_approve" || receipts[0].Observation.Tokens != nil || receipts[0].Event.Topic != "" {
		t.Fatalf("real S1 normalized receipt path: %v", err)
	}
	stableID := receipts[0].ID
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	replayed, partial, err := events.ReadS1Outbox(ctx, db.Conn(), "", 128)
	if err != nil || partial || len(replayed) != 1 || !bytes.Equal(records[0], replayed[0]) {
		t.Fatal("S1 source bytes changed after restart")
	}
	receipts, err = app.NewEventService(sqlite.NewEventRepository(db), 1).Record(ctx, []app.StoredEvent{mapped})
	if err != nil || len(receipts) != 1 || receipts[0].ID != stableID || !bytes.Equal(receipts[0].RawBytes, records[0]) {
		t.Fatal("S1 stable event receipt replay changed")
	}
	signals, err := sqlite.NewEventRepository(db).ListSignals(ctx, 0, math.MaxInt64)
	if err != nil || len(signals) != 1 || signals[0].EventID != stableID || signals[0].Weight != 1.5 || signals[0].Strength != 1 || signals[0].Topic != "" {
		t.Fatal("S1 action was not consumed through real Event/Signal path")
	}
}
