package app

import (
	"context"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// SleepTx composes the domain owners inside one daily transaction. Sleep never
// opens a second store transaction while this callback is active.
type SleepTx interface {
	DynamicsTx
	KnowledgeTx
	ListEvents(int64, int64) ([]StoredEvent, error)
	ListSignals(int64, int64) ([]domain.Signal, error)
	SleepKnowledgeBytes(string, string) (int64, error)
	LoadSleepRun(string) (EngineSleepRun, bool, error)
	SaveSleepRun(EngineSleepRun) error
}

type SleepRepository interface {
	WithSleepTx(context.Context, func(SleepTx) error) error
}
