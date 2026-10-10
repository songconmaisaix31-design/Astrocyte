package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func trackingJobSource(job Job) string {
	if job.Kind != "source_sync" && job.Kind != "source_recommendation" {
		return ""
	}
	var payload struct{ SourceID string }
	_ = json.Unmarshal(job.Payload, &payload)
	return payload.SourceID
}

func (s *Service) recoverTrackingState(tx AttentionTx, job Job) error {
	id := trackingJobSource(job)
	if id == "" {
		return nil
	}
	port, err := trackingPort(tx)
	if err != nil {
		return err
	}
	source, err := port.LoadTrackingSource(id)
	if err != nil {
		return err
	}
	old := source.Version
	source.Version++
	source.Status = "failed"
	source.LastError = job.Error
	items, err := port.ListSourceItems(id)
	if err != nil {
		return err
	}
	for _, item := range items {
		if job.Kind == "source_sync" {
			item.Stale = true
		} else if item.Recommendation != nil && item.Recommendation.Status == "pending" {
			item.Recommendation.Status = "failed"
			item.Recommendation.Error = job.Error
		}
		if err = port.SaveSourceItem(item); err != nil {
			return err
		}
	}
	return port.SaveTrackingSource(source, old)
}

// Automatic retry only requeues explicitly retryable, known-safe failures.
// Budget/deadline/operation/payload survive. Delay derives from persisted time;
// no separate hidden retry state or provider loop is introduced.
func (s *Service) retrySafeJobs(ctx context.Context) error {
	return mapError(s.repo.WithTx(ctx, func(tx AttentionTx) error {
		jobs, err := tx.ListJobs()
		if err != nil {
			return err
		}
		now := s.options.Clock()
		for _, job := range jobs {
			if job.Status != "failed" || job.DeliveryUnknown || job.Error == nil || !job.Error.Retryable || job.CancelRequested {
				continue
			}
			// The cached result is safe to publish again, but local storage must
			// be repaired first. Keep this failed record actionable for explicit
			// human RetryJob rather than exhausting attempts on the obstruction.
			if job.Error.RequiredAction == "repair_storage_then_retry_cached_result" {
				continue
			}
			if job.Kind != "import" && job.Kind != "source_sync" && job.Kind != "distillation" && job.Kind != "source_recommendation" {
				continue
			}
			// Only a classified known failure can reach here. A lost native
			// outcome has DeliveryUnknown and is excluded above.
			if job.Attempts >= job.MaxAttempts || !now.Before(job.DeadlineAt) {
				continue
			}
			if now.Sub(job.UpdatedAt) < time.Duration(max(1, job.Attempts))*time.Second {
				continue
			}
			sourceID := trackingJobSource(job)
			if sourceID != "" {
				superseded := false
				for _, other := range jobs {
					if other.JobID != job.JobID && trackingJobSource(other) == sourceID && (other.Status == "queued" || other.Status == "running" || other.CreatedAt.After(job.CreatedAt)) {
						superseded = true
						break
					}
				}
				if superseded {
					continue
				}
			}
			old := job.Version
			job.Version++
			job.Status = "queued"
			job.ExternalStarted = false
			job.UpdatedAt = now
			// Retain the last real error while retrying; final success clears it.
			if err = tx.SaveJob(job, old); err != nil {
				return err
			}
			if sourceID != "" {
				port, err := trackingPort(tx)
				if err != nil {
					return err
				}
				source, err := port.LoadTrackingSource(sourceID)
				if err != nil {
					return err
				}
				old = source.Version
				source.Version++
				source.Status = "syncing"
				if job.Kind == "source_recommendation" {
					source.Status = "recommending"
				}
				if err = port.SaveTrackingSource(source, old); err != nil {
					return err
				}
			}
			if err = s.event(tx, "job_retried", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID, "attempts": job.Attempts, "error": job.Error, "automatic": true}); err != nil {
				return err
			}
		}
		return nil
	}), "")
}

func retryableInterruption() *apierrors.ServiceError {
	err := serviceError(apierrors.ProviderUnavailable, "Local worker was interrupted", "retry_job")
	err.Retryable = true
	return err
}
