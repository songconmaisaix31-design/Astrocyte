package sqlite

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type SleepRepository struct{ base *EngineRepository }

func NewSleepRepository(db *DB) *SleepRepository {
	return &SleepRepository{base: NewEngineRepository(db)}
}

var _ app.SleepRepository = (*SleepRepository)(nil)

func (r *SleepRepository) WithSleepTx(ctx context.Context, fn func(app.SleepTx) error) error {
	return r.base.WithTx(ctx, func(tx EngineTx) error {
		return fn(&sleepTx{
			DynamicsTx: NewDynamicsTx(tx), KnowledgeTx: BindKnowledgeTx(tx),
			SleepRunTx: NewSleepRunTx(tx), events: NewEventTx(tx), tx: tx,
		})
	})
}

type sleepTx struct {
	app.DynamicsTx
	app.KnowledgeTx
	*SleepRunTx
	events *eventTx
	tx     EngineTx
}

// SQLite length(TEXT) counts characters; CAST AS BLOB measures persisted UTF-8.
func (t *sleepTx) SleepKnowledgeBytes(owner, topic string) (int64, error) {
	var size int64
	err := t.tx.QueryRow(`SELECT COALESCE(SUM(length(CAST(data AS BLOB))),0) FROM attention_engine_knowledge WHERE owner_project=? AND topic=?`, owner, topic).Scan(&size)
	return size, err
}

func (t *sleepTx) ListEvents(from, until int64) ([]app.StoredEvent, error) {
	return t.events.ListEvents(from, until)
}

func (t *sleepTx) ListSignals(from, until int64) ([]domain.Signal, error) {
	return t.events.ListSignals(from, until)
}
