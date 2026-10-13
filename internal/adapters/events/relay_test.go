package events

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayEventSinkPersistsDerivedPendingObservationsOnly(t *testing.T) {
	root := filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification", "relay-events", uuid.NewString())
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal("test state directory unavailable")
	}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal("isolated test database unavailable")
	}
	defer db.Close()
	ctx := context.Background()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("isolated test schema unavailable")
	}
	repo := sqlite.NewEventRepository(db)
	sink := NewRelayEventSink(attention.NewPendingEventService(repo), func() time.Time { return time.Unix(10, 0) })
	v := domain.RelayEvent{ID: "fixture-observation", Kind: "commit", Source: "git", Scope: "project-a", Topic: "work-a", TS: time.Unix(5, 0), Agent: "PRIVATE_CANARY", TaskID: "PRIVATE_CANARY", WorktreeID: "PRIVATE_CANARY"}
	if sink.EmitRelayEvent(ctx, v) != nil || sink.EmitRelayEvent(ctx, v) != nil {
		t.Fatal("observation ingest or stable-origin replay failed")
	}
	rows, err := repo.ListEvents(ctx, 0, 20000)
	if err != nil || len(rows) != 1 {
		t.Fatal("event origin not persisted once")
	}
	record := rows[0]
	if record.Event != nil || record.Observation.Source != "git" || record.Observation.SourceKind != "commit" || record.Observation.Outcome != "unknown" || record.Observation.Depth != "unknown" || record.Observation.Tokens != nil {
		t.Fatal("pending source vocabulary or unknown boundary changed")
	}
	var payload map[string]any
	if json.Unmarshal(record.RawBytes, &payload) != nil || len(payload) != 6 || strings.Contains(string(record.RawBytes), "PRIVATE_CANARY") {
		t.Fatal("derived envelope is not closed six-field projection")
	}
	signals, err := repo.ListSignals(ctx, 0, 20000)
	if err != nil || len(signals) != 0 {
		t.Fatal("relay event changed cognitive signal state")
	}
	v.ID = "invalid"
	negative := int64(-1)
	v.Tokens = &negative
	if sink.EmitRelayEvent(ctx, v) == nil {
		t.Fatal("invalid token observation accepted")
	}
	rows, err = repo.ListEvents(ctx, 0, 20000)
	if err != nil || len(rows) != 1 {
		t.Fatal("invalid observation altered event state")
	}
}
