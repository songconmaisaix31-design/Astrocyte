package app

import (
	"context"
	"fmt"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// EventTx binds each immutable receipt and its signals in the same transaction.
type EventTx interface {
	PutEvent(StoredEvent) (StoredEvent, error)
	PutSignal(domain.Signal) error
}

type EventRepository interface {
	EventReader
	WithEventTx(context.Context, func(EventTx) error) error
	ListSignals(context.Context, int64, int64) ([]domain.Signal, error)
}

type EventService struct {
	repo           EventRepository
	strength       *float64
	failureCapture []func(context.Context, string) error
}

// EventCapturePendingError reports a known commit with unfinished capture.
// Callers retain these IDs and reconcile them instead of resending new origins.
type EventCapturePendingError struct {
	ReceiptIDs []string
	Cause      error
}

func (e *EventCapturePendingError) Error() string {
	return fmt.Sprintf("receipts committed; failure capture pending: %v", e.Cause)
}
func (e *EventCapturePendingError) Unwrap() error { return e.Cause }

// strength has no product default; callers must provide the selected scale.
func NewEventService(repo EventRepository, strength float64, failureCapture ...func(context.Context, string) error) *EventService {
	return &EventService{repo: repo, strength: &strength, failureCapture: failureCapture}
}

// NewPendingEventService records source observations while the scale remains
// undecided. Mapped events are rejected; no default strength is assigned.
func NewPendingEventService(repo EventRepository) *EventService { return &EventService{repo: repo} }

func (s *EventService) Record(ctx context.Context, records []StoredEvent) ([]StoredEvent, error) {
	result := make([]StoredEvent, 0, len(records))
	err := s.repo.WithEventTx(ctx, func(tx EventTx) error {
		for _, input := range records {
			record, err := tx.PutEvent(input)
			if err != nil {
				return err
			}
			if record.Event != nil {
				if s.strength == nil {
					return fmt.Errorf("mapped event ingestion requires explicitly configured signal strength")
				}
				signal, err := domain.NormalizeSignal(record.ID, *record.Event, record.Observation.Outcome, record.Observation.Depth, *s.strength)
				if err != nil {
					return err
				}
				if signal != nil {
					if err = tx.PutSignal(*signal); err != nil {
						return err
					}
				}
			}
			result = append(result, record)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Capture uses the committed ID and a separately project-bound service.
	// A capture failure cannot roll back or obscure the original receipt.
	for _, record := range result {
		if record.Event == nil || record.Event.Kind != domain.EventKindAgentTask || (record.Observation.Outcome != "failure" && record.Observation.Outcome != "failed") {
			continue
		}
		for _, capture := range s.failureCapture {
			if err := capture(ctx, record.ID); err != nil {
				ids := make([]string, 0, len(result))
				for _, receipt := range result {
					ids = append(ids, receipt.ID)
				}
				return result, &EventCapturePendingError{ReceiptIDs: ids, Cause: err}
			}
		}
	}
	return result, nil
}
