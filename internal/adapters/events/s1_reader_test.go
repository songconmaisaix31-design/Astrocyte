package events

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func TestS1OutboxPayloadAndReceiptAccounting(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`CREATE TABLE attention_outbox(id TEXT,type TEXT,aggregate_id TEXT,aggregate_version INTEGER,occurred_at TEXT,correlation_id TEXT,causation_id TEXT,payload TEXT)`); err != nil {
		t.Fatal(err)
	}
	payload := ` { "material_id" : "m", "private_field" : "kept local" } `
	for _, id := range []string{"one", "two"} {
		if _, err = db.Exec(`INSERT INTO attention_outbox VALUES(?, 'attention.material_imported', 'm', 2, '2026-10-10T00:00:00Z', 'request', 'cause', ?)`, id, payload); err != nil {
			t.Fatal(err)
		}
	}
	records, partial, err := ReadS1Outbox(context.Background(), db, "", 1)
	if err != nil || !partial || len(records) != 1 || !bytes.Contains(records[0], []byte(payload)) {
		t.Fatalf("outbox preservation: count=%d partial=%v err=%v", len(records), partial, err)
	}
	v, err := outboxRecord(records[0])
	if err != nil || v.ID != "one" || v.AggregateVersion != 2 || v.CorrelationID != "request" || v.CausationID != "cause" || v.OccurredAt.UnixMilli() != 1791590400000 {
		t.Fatal("source receipt or millisecond time changed")
	}
	records, partial, err = ReadS1Outbox(context.Background(), db, v.ID, 1)
	if err != nil || partial || len(records) != 1 {
		t.Fatal("outbox paging dropped records")
	}
	v, err = outboxRecord(records[0])
	if err != nil || v.ID != "two" {
		t.Fatal("outbox cursor lost identity")
	}
}
