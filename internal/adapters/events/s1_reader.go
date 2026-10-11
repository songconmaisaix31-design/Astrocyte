package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// ReadS1Outbox reads the canonical S1 output without acknowledging or changing
// its receipts. The caller can replay the same stable IDs after a restart.
func ReadS1Outbox(ctx context.Context, db *sql.DB, afterID string, limit int) ([]json.RawMessage, bool, error) {
	if limit < 1 || limit > maxReplayRecords {
		return nil, false, fmt.Errorf("outbox limit must be 1..128")
	}
	rows, err := db.QueryContext(ctx, `SELECT id,type,aggregate_id,aggregate_version,occurred_at,correlation_id,causation_id,payload FROM attention_outbox WHERE id>? ORDER BY id LIMIT ?`, afterID, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	result := []json.RawMessage{}
	for rows.Next() {
		if len(result) == limit {
			return result, true, nil
		}
		var v struct {
			ID               string `json:"id"`
			Type             string `json:"type"`
			AggregateID      string `json:"aggregate_id"`
			AggregateVersion int    `json:"aggregate_version"`
			OccurredAt       string `json:"occurred_at"`
			CorrelationID    string `json:"correlation_id"`
			CausationID      string `json:"causation_id"`
		}
		var payload string
		if err = rows.Scan(&v.ID, &v.Type, &v.AggregateID, &v.AggregateVersion, &v.OccurredAt, &v.CorrelationID, &v.CausationID, &payload); err != nil {
			return nil, false, err
		}
		if !json.Valid([]byte(payload)) {
			return nil, false, fmt.Errorf("invalid S1 outbox payload")
		}
		data, err := json.Marshal(v)
		if err != nil {
			return nil, false, err
		}
		// Keep payload bytes, including original whitespace. Marshal of a
		// RawMessage would compact those bytes before the event store sees them.
		data = append(data[:len(data)-1], []byte(`,"payload":`)...)
		data = append(data, []byte(payload)...)
		data = append(data, '}')
		result = append(result, data)
	}
	return result, false, rows.Err()
}

func outboxRecord(raw json.RawMessage) (app.OutboxEvent, error) {
	v := object(raw)
	var version int
	if v == nil || str(v, "id") == "" || json.Unmarshal(v["aggregate_version"], &version) != nil {
		return app.OutboxEvent{}, fmt.Errorf("invalid S1 outbox envelope")
	}
	t, err := time.Parse(time.RFC3339Nano, str(v, "occurred_at"))
	if err != nil {
		return app.OutboxEvent{}, fmt.Errorf("invalid S1 event timestamp")
	}
	return app.OutboxEvent{ID: str(v, "id"), Type: str(v, "type"), AggregateID: str(v, "aggregate_id"), AggregateVersion: version, OccurredAt: t, CorrelationID: str(v, "correlation_id"), CausationID: str(v, "causation_id"), Payload: v["payload"]}, nil
}
