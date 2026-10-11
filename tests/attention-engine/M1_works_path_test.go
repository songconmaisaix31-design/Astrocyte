package attention_engine_test

import (
	"bytes"
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1WorksServiceUsesPendingReceiptsWithoutInventedAttention(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "works-events.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	now := time.UnixMilli(1791590400123)
	sink := events.NewWorksEventSink(attentionapp.NewPendingEventService(sqlite.NewEventRepository(db)), func() time.Time { return now })
	work := workspacedomain.Work{ID: "selected-fixture-work", Platform: "codex", Root: t.TempDir(), Name: "selected fixture", Kind: "folder"}
	if err = db.SaveWorksSnapshot(ctx, workspacedomain.WorksSnapshot{Works: []workspacedomain.Work{work}}); err != nil {
		t.Fatal(err)
	}
	works := workspaceapp.NewWorksService(db, nil, nil, sink, workspaceapp.WorksServiceOptions{Clock: func() time.Time { return now }})
	_, err = works.SetWorkGroup(ctx, workspacedomain.Caller{ID: "human", Kind: "human"}, work.ID, workspaceapp.WorkGroupCommand{ExpectedRevision: 0, Group: "selected group"})
	if err != nil {
		t.Fatal(err)
	}
	machine := workspacedomain.WorkEvent{WorkID: work.ID, Kind: workspacedomain.WorkEventSynced, Source: workspacedomain.WorkEventSourceMachine, Scope: work.Root, At: now, Payload: []byte(` { "revision" : 1, "source_metadata" : "kept bytes" } `)}
	if err = sink.EmitWorkEvent(ctx, machine); err != nil {
		t.Fatal(err)
	}
	if err = sink.EmitWorkEvent(ctx, machine); err != nil {
		t.Fatal(err)
	}
	rows, err := sqlite.NewEventRepository(db).ListEvents(ctx, 0, math.MaxInt64)
	if err != nil || len(rows) != 2 {
		t.Fatalf("works receipt accounting: %d %v", len(rows), err)
	}
	for _, v := range rows {
		if v.Event != nil || v.Observation.Tokens != nil || v.Observation.Outcome != "unknown" || v.Observation.Depth != "unknown" || v.Observation.TS != now.UnixMilli() {
			t.Fatal("works metadata invented human attention or agent task")
		}
	}
	var machineID string
	for _, v := range rows {
		if v.Observation.Source == "machine" {
			machineID = v.ID
			if !bytes.Contains(v.RawBytes, machine.Payload) {
				t.Fatal("work payload bytes compacted")
			}
		}
	}
	ss, err := sqlite.NewEventRepository(db).ListSignals(ctx, 0, math.MaxInt64)
	if err != nil || len(ss) != 0 {
		t.Fatal("pending works produced signals")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := sqlite.NewEventRepository(db).GetEvent(ctx, machineID)
	if err != nil || !bytes.Contains(stored.RawBytes, machine.Payload) || stored.Event != nil {
		t.Fatal("works raw/pending state changed after restart")
	}
}

func TestM1WorksSavedButSinkFailureIsExplicit(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "works.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	eventDB, err := sqlite.Open(filepath.Join(t.TempDir(), "events.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = eventDB.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	sink := events.NewWorksEventSink(attentionapp.NewEventService(sqlite.NewEventRepository(eventDB), 1), nil)
	if err = eventDB.Close(); err != nil {
		t.Fatal(err)
	}
	work := workspacedomain.Work{ID: "selected-fixture-work", Platform: "codex", Root: t.TempDir(), Name: "selected fixture", Kind: "folder"}
	if err = db.SaveWorksSnapshot(ctx, workspacedomain.WorksSnapshot{Works: []workspacedomain.Work{work}}); err != nil {
		t.Fatal(err)
	}
	works := workspaceapp.NewWorksService(db, nil, nil, sink, workspaceapp.WorksServiceOptions{})
	_, err = works.SetWorkGroup(ctx, workspacedomain.Caller{ID: "human", Kind: "human"}, work.ID, workspaceapp.WorkGroupCommand{ExpectedRevision: 0, Group: "selected group"})
	if err == nil {
		t.Fatal("source hid EventStore failure after saving work metadata")
	}
	metadata, err := db.LoadWorkMetadata(ctx, work.ID)
	if err != nil || metadata.Revision != 1 || metadata.Group != "selected group" {
		t.Fatal("source failure lost known saved metadata")
	}
}
