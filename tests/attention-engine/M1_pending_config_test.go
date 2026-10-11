package attention_engine_test

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1PendingRecorderDoesNotChooseSignalScale(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "pending.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewEventRepository(db)
	svc := app.NewPendingEventService(repo)
	input := app.StoredEvent{OriginKey: "fixture:pending", RawBytes: []byte(`{"source_kind":"work_group_set"}`), Observation: app.Observation{Source: "manual", SourceKind: "work_group_set", Outcome: "unknown", Depth: "unknown", Provenance: "contract_local:pending"}}
	rows, err := svc.Record(ctx, []app.StoredEvent{input})
	if err != nil || len(rows) != 1 || rows[0].Event != nil {
		t.Fatal("pending-only source required a guessed scale")
	}
	input.OriginKey = "fixture:requires-scale"
	input.Event = &domain.Event{Source: "manual", Kind: "collect"}
	if _, err = svc.Record(ctx, []app.StoredEvent{input}); err == nil {
		t.Fatal("unset scale accepted mapped signal input")
	}
	all, err := repo.ListEvents(ctx, 0, math.MaxInt64)
	if err != nil || len(all) != 1 {
		t.Fatal("unset scale rejection did not roll back mapped input")
	}
	ss, err := repo.ListSignals(ctx, 0, math.MaxInt64)
	if err != nil || len(ss) != 0 {
		t.Fatal("pending-only recorder generated signals")
	}
}
