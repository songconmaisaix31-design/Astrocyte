package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// SleepRunTx shares the owner's EngineTx with event, graph and knowledge stores.
// It writes only daily consolidation records; it never opens a nested transaction.
type SleepRunTx struct{ tx EngineTx }

func NewSleepRunTx(tx EngineTx) *SleepRunTx { return &SleepRunTx{tx: tx} }

func (t *SleepRunTx) LoadSleepRun(day string) (app.EngineSleepRun, bool, error) {
	var data string
	err := t.tx.QueryRow("SELECT data FROM attention_engine_sleep_runs WHERE day=?", day).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return app.EngineSleepRun{}, false, nil
	}
	if err != nil {
		return app.EngineSleepRun{}, false, err
	}
	var run app.EngineSleepRun
	err = json.Unmarshal([]byte(data), &run)
	return run, err == nil, err
}

func (t *SleepRunTx) SaveSleepRun(run app.EngineSleepRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}
	if _, err = t.tx.Exec(`INSERT INTO attention_engine_sleep_runs(id,day,started_at,completed_at,data) VALUES(?,?,?,?,?)`,
		run.ID, run.Day, run.StartedAt, run.CompletedAt, string(data)); err != nil {
		return err
	}
	m := run.Metrics
	if _, err = t.tx.Exec(`INSERT INTO attention_engine_sleep_metrics(run_id,key_facts_before,key_facts_retained,bytes_before,bytes_after,retention_precision,storage_reduction) VALUES(?,?,?,?,?,?,?)`,
		run.ID, m.KeyFactsBefore, m.KeyFactsRetained, m.BytesBefore, m.BytesAfter, m.RetentionPrecision, m.StorageReduction); err != nil {
		return err
	}
	var candidates []json.RawMessage
	if len(run.Candidates) > 0 {
		if err = json.Unmarshal(run.Candidates, &candidates); err != nil {
			return err
		}
	}
	for _, candidate := range candidates {
		var identity struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
		}
		if err = json.Unmarshal(candidate, &identity); err != nil {
			return err
		}
		if identity.ID == "" || identity.Kind == "" {
			return fmt.Errorf("sleep candidate requires id and kind")
		}
		if _, err = t.tx.Exec(`INSERT INTO attention_engine_sleep_candidates(run_id,candidate_id,kind,data) VALUES(?,?,?,?)`,
			run.ID, identity.ID, identity.Kind, string(candidate)); err != nil {
			return err
		}
	}
	for _, id := range run.ConflictIDs {
		if _, err = t.tx.Exec(`INSERT INTO attention_engine_sleep_conflicts(run_id,conflict_id) VALUES(?,?)`, run.ID, id); err != nil {
			return err
		}
	}
	return nil
}
