package attention_engine_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM3DailyRunStoreRollbackAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "m3.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	run := app.EngineSleepRun{
		ID: "sleep:2026-10-11", Day: "2026-10-11", StartedAt: 1, CompletedAt: 2,
		Candidates:  json.RawMessage(`[{"id":"concept:reading","kind":"concept","topic":"reading"}]`),
		Metrics:     app.EngineSleepMetrics{KeyFactsBefore: 2, KeyFactsRetained: 1, BytesBefore: 100, BytesAfter: 200, RetentionPrecision: .5, StorageReduction: -1},
		ConflictIDs: []string{"missing-conflict"},
	}
	repo := sqlite.NewEngineRepository(db)
	if err = repo.WithTx(ctx, func(tx sqlite.EngineTx) error { return sqlite.NewSleepRunTx(tx).SaveSleepRun(run) }); err == nil {
		t.Fatal("missing conflict accepted")
	}
	for _, table := range []string{"attention_engine_sleep_runs", "attention_engine_sleep_metrics", "attention_engine_sleep_candidates", "attention_engine_sleep_conflicts"} {
		var count int
		if err = db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("partial %s: %d %v", table, count, err)
		}
	}
	run.ConflictIDs = []string{}
	if err = repo.WithTx(ctx, func(tx sqlite.EngineTx) error { return sqlite.NewSleepRunTx(tx).SaveSleepRun(run) }); err != nil {
		t.Fatal(err)
	}
	if err = repo.WithTx(ctx, func(tx sqlite.EngineTx) error { return sqlite.NewSleepRunTx(tx).SaveSleepRun(run) }); err == nil {
		t.Fatal("duplicate daily record accepted")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	repo = sqlite.NewEngineRepository(db)
	if err = repo.WithTx(ctx, func(tx sqlite.EngineTx) error {
		got, found, err := sqlite.NewSleepRunTx(tx).LoadSleepRun(run.Day)
		if err != nil {
			return err
		}
		if !found || !reflect.DeepEqual(got, run) {
			t.Fatalf("restart changed run: found=%v got=%+v", found, got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var reduction float64
	if err = db.Conn().QueryRowContext(ctx, "SELECT storage_reduction FROM attention_engine_sleep_metrics WHERE run_id=?", run.ID).Scan(&reduction); err != nil || reduction != -1 {
		t.Fatalf("growth hidden: %f %v", reduction, err)
	}
}
