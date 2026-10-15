package sqlite

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccessSessionUnrelatedPlaintextAndForeignScopeDoNotBlockOwnedReceipt(t *testing.T) {
	base := os.Getenv("LOCALAPPDATA")
	if !filepath.IsAbs(base) {
		base = t.TempDir()
	}
	root := filepath.Join(base, "astrocyte-workbench", "v0.2-access", "verification", "access-session-reader", uuid.NewString())
	if os.MkdirAll(root, 0700) != nil {
		t.Fatal("isolated fixture directory unavailable")
	}
	db, e := Open(filepath.Join(root, "app.sqlite"))
	if e != nil {
		t.Fatal("isolated app DB unavailable")
	}
	defer db.Close()
	ctx := context.Background()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("isolated schema unavailable")
	}
	task, terminal, session, binding := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	stamp := time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC)
	projection, _ := json.Marshal(workbench.Event{ID: uuid.NewString(), Kind: "native_activity", Topic: "relay:" + task, Scope: "public-project", TS: stamp.Format(time.RFC3339Nano)})
	service := attention.NewPendingEventService(NewEventRepository(db))
	good, e := service.Record(ctx, []attention.StoredEvent{{OriginKey: "relay-session:" + binding + ":" + uuid.NewString(), RawBytes: projection, Observation: attention.Observation{Source: "codex", SourceKind: "native_activity", Session: session, Scope: "public-project", TS: stamp.UnixMilli(), Outcome: "unknown", Depth: "unknown", Provenance: "public-fixture"}}})
	if e != nil {
		t.Fatal("public owned receipt unavailable")
	}
	_, e = service.Record(ctx, []attention.StoredEvent{{OriginKey: "unrelated-public:" + uuid.NewString(), RawBytes: []byte("PRIVATE_CANARY non JSON unrelated"), Observation: attention.Observation{Source: "public-unrelated", SourceKind: "unknown", Session: uuid.NewString(), Scope: "foreign-project", TS: stamp.UnixMilli(), Outcome: "unknown", Depth: "unknown", Provenance: "public-fixture"}}, {OriginKey: "relay-session:" + uuid.NewString() + ":" + uuid.NewString(), RawBytes: []byte("PRIVATE_CANARY foreign non JSON"), Observation: attention.Observation{Source: "public-foreign", SourceKind: "unknown", Session: uuid.NewString(), Scope: "foreign-project", TS: stamp.UnixMilli(), Outcome: "unknown", Depth: "unknown", Provenance: "public-fixture"}}})
	if e != nil {
		t.Fatal("unrelated public fixture unavailable")
	}
	reader := NewAccessSessionReader(db)
	scopes := []workbench.SessionAdmissionScope{{TaskID: task, ProjectID: "public-project", TerminalID: terminal, NativeSessionID: session, NativeBindingID: binding}}
	rows, total, e := reader.ReadCommittedSessionEvents(ctx, scopes, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e != nil || total != 1 || len(rows) != 1 || rows[0].ID != good[0].ID {
		t.Fatal("unrelated pending plaintext blocked owned committed receipt")
	}
	_, e = service.Record(ctx, []attention.StoredEvent{{OriginKey: "relay-session:" + binding + ":" + uuid.NewString(), RawBytes: []byte("PRIVATE_CANARY malformed selected projection"), Observation: attention.Observation{Source: "codex", SourceKind: "unknown", Session: session, Scope: "public-project", TS: stamp.UnixMilli(), Outcome: "unknown", Depth: "unknown", Provenance: "public-fixture"}}})
	if e != nil {
		t.Fatal("selected malformed fixture unavailable")
	}
	_, _, e = reader.ReadCommittedSessionEvents(ctx, scopes, workbench.SessionEventsQuery{Page: 1, PageSize: 50})
	if e == nil || e.Error() != "session admission projection invalid" {
		t.Fatal("selected malformed projection silently skipped or exposed private error")
	}
}

func TestAccessSessionSubMillisecondSourceOrderBeforePagination(t *testing.T) {
	base := os.Getenv("LOCALAPPDATA")
	if !filepath.IsAbs(base) {
		base = t.TempDir()
	}
	root := filepath.Join(base, "astrocyte-workbench", "v0.2-access", "verification", "access-session-reader", uuid.NewString())
	if os.MkdirAll(root, 0700) != nil {
		t.Fatal("public precision fixture unavailable")
	}
	db, e := Open(filepath.Join(root, "app.sqlite"))
	if e != nil {
		t.Fatal("public precision DB unavailable")
	}
	defer db.Close()
	ctx := context.Background()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("public precision schema unavailable")
	}
	task, terminal, session, binding := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	stamp := time.Date(2026, 10, 14, 1, 0, 0, 0, time.UTC)
	service := attention.NewPendingEventService(NewEventRepository(db))
	newestID := ""
	for i := 0; i < 2; i++ {
		ts := stamp.Add(time.Duration(i+1) * time.Nanosecond)
		projection, _ := json.Marshal(workbench.Event{ID: uuid.NewString(), Kind: "native_activity", Topic: "relay:" + task, Scope: "public-project", TS: ts.Format(time.RFC3339Nano)})
		committed, err := service.Record(ctx, []attention.StoredEvent{{OriginKey: "relay-session:" + binding + ":" + uuid.NewString(), RawBytes: projection, Observation: attention.Observation{Source: "codex", SourceKind: "native_activity", Session: session, Scope: "public-project", TS: ts.UnixMilli(), Outcome: "unknown", Depth: "unknown", Provenance: "public-fixture"}}})
		if err != nil || len(committed) != 1 {
			t.Fatal("precision receipt unavailable")
		}
		newestID = committed[0].ID
	}
	rows, total, e := NewAccessSessionReader(db).ReadCommittedSessionEvents(ctx, []workbench.SessionAdmissionScope{{TaskID: task, ProjectID: "public-project", TerminalID: terminal, NativeSessionID: session, NativeBindingID: binding}}, workbench.SessionEventsQuery{Page: 1, PageSize: 1})
	if e != nil || total != 2 || len(rows) != 1 || rows[0].ID != newestID {
		t.Fatal("receipt tie ordering overrode finer source timestamp")
	}
}
