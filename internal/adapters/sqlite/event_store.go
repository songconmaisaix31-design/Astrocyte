package sqlite

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type EventRepository struct{ base *EngineRepository }

func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{base: NewEngineRepository(db)}
}

var _ app.EventRepository = (*EventRepository)(nil)

func (r *EventRepository) WithEventTx(ctx context.Context, fn func(app.EventTx) error) error {
	return r.base.WithTx(ctx, func(tx EngineTx) error { return fn(&eventTx{tx: tx}) })
}

type eventTx struct{ tx EngineTx }

// NewEventTx shares the caller's EngineTx for atomic daily consolidation.
func NewEventTx(tx EngineTx) *eventTx { return &eventTx{tx: tx} }

func (t *eventTx) ListEvents(from, until int64) ([]app.StoredEvent, error) {
	rows, err := t.tx.Query("SELECT "+eventColumns+" FROM attention_engine_events WHERE ts>=? AND ts<=? ORDER BY ts,id", from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []app.StoredEvent{}
	for rows.Next() {
		v, err := scanStoredEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (t *eventTx) ListSignals(from, until int64) ([]domain.Signal, error) {
	rows, err := t.tx.Query(`SELECT data FROM attention_engine_signals WHERE ts>=? AND ts<=? ORDER BY ts,event_id,topic,class`, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Signal{}
	for rows.Next() {
		var data string
		var v domain.Signal
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

const eventColumns = "id,origin_key,received_at,source,ts,kind,data,raw_bytes,observation,mapping_status"

type eventScanner interface{ Scan(...any) error }

func scanStoredEvent(row eventScanner) (app.StoredEvent, error) {
	var v app.StoredEvent
	var origin, observation sql.NullString
	var received sql.NullInt64
	var source, kind, data, status string
	var ts int64
	err := row.Scan(&v.ID, &origin, &received, &source, &ts, &kind, &data, &v.RawBytes, &observation, &status)
	if err != nil {
		return v, err
	}
	v.OriginKey, v.ReceivedAt = origin.String, received.Int64
	if observation.Valid {
		if err = json.Unmarshal([]byte(observation.String), &v.Observation); err != nil {
			return v, err
		}
	} else {
		v.Observation = app.Observation{Source: source, SourceKind: kind, TS: ts, Outcome: "unknown", Depth: "unknown", Provenance: "legacy:unknown"}
	}
	if status == "mapped" {
		if err = json.Unmarshal([]byte(data), &v.Event); err != nil {
			return v, err
		}
		if v.Event != nil && v.RawBytes != nil {
			v.Event.Raw = append(json.RawMessage(nil), v.RawBytes...)
		}
	}
	return v, nil
}

func (t *eventTx) PutEvent(v app.StoredEvent) (app.StoredEvent, error) {
	if v.OriginKey == "" || v.Observation.Source == "" || v.Observation.SourceKind == "" || v.Observation.Provenance == "" || len(v.RawBytes) == 0 {
		return v, fmt.Errorf("source receipt identity, provenance and raw bytes required")
	}
	old, err := scanStoredEvent(t.tx.QueryRow("SELECT "+eventColumns+" FROM attention_engine_events WHERE origin_key=?", v.OriginKey))
	if err == nil {
		// Receipt replay cannot overwrite observed source data or its original ID.
		a, _ := json.Marshal(old.Observation)
		b, _ := json.Marshal(v.Observation)
		if !bytes.Equal(old.RawBytes, v.RawBytes) || !bytes.Equal(a, b) {
			return v, fmt.Errorf("origin key already binds different source bytes or observations")
		}
		return old, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return v, err
	}
	v.ID = rand.Text()
	status, kind, topic := "pending", v.Observation.SourceKind, ""
	if v.Event != nil {
		status, kind, topic = "mapped", v.Event.Kind, v.Event.Topic
		v.Event.Raw = append(json.RawMessage(nil), v.RawBytes...)
	}
	data, err := json.Marshal(v.Event)
	if err != nil {
		return v, err
	}
	obs, err := json.Marshal(v.Observation)
	if err != nil {
		return v, err
	}
	_, err = t.tx.Exec(`INSERT INTO attention_engine_events(id,source,ts,kind,topic,data,origin_key,received_at,raw_bytes,observation,mapping_status) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.Observation.Source, v.Observation.TS, kind, topic, string(data), v.OriginKey, v.ReceivedAt, v.RawBytes, string(obs), status)
	return v, err
}

func (t *eventTx) PutSignal(v domain.Signal) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var prior string
	err = t.tx.QueryRow(`SELECT data FROM attention_engine_signals WHERE event_id=? AND topic=? AND class=?`, v.EventID, v.Topic, v.Class).Scan(&prior)
	if err == nil {
		var old domain.Signal
		if err = json.Unmarshal([]byte(prior), &old); err != nil {
			return err
		}
		if old != v {
			return fmt.Errorf("replayed signal differs from persisted strength or weight")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = t.tx.Exec(`INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES(?,?,?,?,?,?)`, v.EventID, v.Topic, v.Class, v.Weight, v.TS, string(data))
	return err
}

func (r *EventRepository) GetEvent(ctx context.Context, id string) (app.StoredEvent, error) {
	return scanStoredEvent(r.base.Conn().QueryRowContext(ctx, "SELECT "+eventColumns+" FROM attention_engine_events WHERE id=?", id))
}
func (r *EventRepository) ListEvents(ctx context.Context, from, until int64) ([]app.StoredEvent, error) {
	rows, err := r.base.Conn().QueryContext(ctx, "SELECT "+eventColumns+" FROM attention_engine_events WHERE ts>=? AND ts<=? ORDER BY ts,id", from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []app.StoredEvent{}
	for rows.Next() {
		v, e := scanStoredEvent(rows)
		if e != nil {
			return nil, e
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (r *EventRepository) ListSignals(ctx context.Context, from, until int64) ([]domain.Signal, error) {
	rows, err := r.base.Conn().QueryContext(ctx, `SELECT data FROM attention_engine_signals WHERE ts>=? AND ts<=? ORDER BY ts,event_id,topic,class`, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Signal{}
	for rows.Next() {
		var data string
		var v domain.Signal
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
