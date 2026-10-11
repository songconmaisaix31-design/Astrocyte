// Package attentionruntime connects the original attention use cases at process
// entrypoints. Domain rules and persistence remain in their owner packages.
package attentionruntime

import (
	"context"
	"fmt"
	"sort"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type EventRecorder struct {
	Events     *app.EventService
	Repository app.EventRepository
	Engine     app.Engine
}

// Record retains committed receipts if the subsequent state update fails.
// Replaying the same source identity invokes the owner's idempotent folding.
func (r *EventRecorder) Record(ctx context.Context, records []app.StoredEvent) ([]app.StoredEvent, error) {
	rows, err := r.Events.Record(ctx, records)
	if err != nil || len(rows) == 0 {
		return rows, err
	}
	ids := map[string]bool{}
	from, until := rows[0].Observation.TS, rows[0].Observation.TS
	for _, row := range rows {
		ids[row.ID] = true
		if row.Observation.TS < from {
			from = row.Observation.TS
		}
		if row.Observation.TS > until {
			until = row.Observation.TS
		}
	}
	signals, err := r.Repository.ListSignals(ctx, from, until)
	if err != nil {
		return rows, &app.MCPDynamicsPendingError{Cause: fmt.Errorf("signal update pending: %w", err)}
	}
	byTopic := map[string][]domain.Signal{}
	for _, signal := range signals {
		if ids[signal.EventID] && signal.Topic != "" {
			byTopic[signal.Topic] = append(byTopic[signal.Topic], signal)
		}
	}
	topics := make([]string, 0, len(byTopic))
	for topic := range byTopic {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	for _, topic := range topics {
		batch := byTopic[topic]
		now := batch[len(batch)-1].TS
		if _, err = r.Engine.ApplySignals(ctx, topic, batch, now); err != nil {
			return rows, &app.MCPDynamicsPendingError{Cause: fmt.Errorf("attention update pending: %w", err)}
		}
	}
	return rows, nil
}
