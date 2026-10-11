package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// CaptureCommittedFailure is W1's post-commit callback. Application setup binds
// it to the trusted project; the returned event ID is reread from persistence.
func (s *ExperienceService) CaptureCommittedFailure(ctx context.Context, eventID string) error {
	_, err := s.CaptureFailure(ctx, eventID)
	return err
}

// RecordEvents checks the trusted project before W1's receipt/signal commit.
// W0/MCP may use it for project-bound ingestion. The post-commit capture is
// idempotent whether the supplied EventService already has the W1 callback.
func (s *ExperienceService) RecordEvents(ctx context.Context, events *EventService, records []StoredEvent) ([]StoredEvent, error) {
	if err := s.project(s.options.ProjectID); err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Observation.Scope != s.options.ProjectID || (record.Event != nil && record.Event.Scope != s.options.ProjectID) {
			return nil, serviceError(apierrors.ScopeDenied, "Event receipt differs from trusted project binding", "use_bound_project")
		}
	}
	stored, err := events.Record(ctx, records)
	if err != nil {
		return stored, err
	}
	for _, record := range stored {
		if _, err := s.CaptureFailure(ctx, record.ID); err != nil {
			// Original event IDs remain available if the post-commit capture failed.
			return stored, err
		}
	}
	return stored, nil
}
