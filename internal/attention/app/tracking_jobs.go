package app

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *Service) trackingWork(ctx context.Context, claim *Job) (context.Context, func(), error) {
	workCtx, cancel := context.WithTimeout(ctx, claim.DeadlineAt.Sub(s.options.Clock()))
	work := &activeWork{cancel: cancel}
	s.activeMu.Lock()
	s.activeJobs[claim.JobID] = work
	s.activeMu.Unlock()
	cleanup := func() {
		cancel()
		s.activeMu.Lock()
		if s.activeJobs[claim.JobID] == work {
			delete(s.activeJobs, claim.JobID)
		}
		s.activeMu.Unlock()
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" {
			return domain.ErrVersion
		}
		old := job.Version
		job.Version++
		job.ExternalStarted = true
		job.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		*claim = job
		return nil
	})
	if err != nil {
		cleanup()
	}
	return workCtx, cleanup, err
}

func (s *Service) processSourceSync(ctx context.Context, claim Job) error {
	var payload syncListingPayload
	if json.Unmarshal(claim.Payload, &payload) != nil {
		return s.failJob(claim, domain.ErrInvalid, false)
	}
	if s.options.ListingReader == nil {
		return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.ProviderUnavailable, "No public listing reader is configured", "configure_public_listing_reader"), false)
	}
	workCtx, cleanup, err := s.trackingWork(ctx, &claim)
	if err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return nil
		}
		return err
	}
	defer cleanup()
	if payload.Done {
		return s.finishTrackingSync(workCtx, claim, payload, payload.HasMore, payload.NextCursor)
	}
	var source TrackingSource
	err = s.repo.WithTx(workCtx, func(tx AttentionTx) error {
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		source, err = port.LoadTrackingSource(payload.SourceID)
		return err
	})
	if err != nil {
		return s.failTrackingJob(claim, payload.SourceID, err, false)
	}
	// Payload cursor/count are committed after every page. A restart resumes a
	// known safe read, preserves its total cap, and never drops cached pages.
	for payload.Count < payload.Limit {
		page, err := s.options.ListingReader.ReadPage(workCtx, source, payload.Cursor, payload.Limit-payload.Count)
		if err != nil {
			return s.failTrackingJob(claim, payload.SourceID, sourceError(err), externalOutcomeUnknown(err))
		}
		if workCtx.Err() != nil {
			return s.failTrackingJob(claim, payload.SourceID, sourceError(workCtx.Err()), false)
		}
		if len(page.Items) > payload.Limit-payload.Count || (page.HasMore && (page.NextCursor == nil || *page.NextCursor == "" || *page.NextCursor == payload.Cursor || len(page.Items) == 0)) {
			return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.ProviderUnavailable, "Public listing pagination did not advance within its cap", "inspect_public_listing"), false)
		}
		err = s.persistListingPage(workCtx, &claim, &payload, page)
		if err != nil {
			if errors.Is(err, domain.ErrVersion) {
				return nil
			}
			return s.failTrackingJob(claim, payload.SourceID, err, false)
		}
		if !page.HasMore || payload.Count >= payload.Limit {
			return s.finishTrackingSync(workCtx, claim, payload, page.HasMore, page.NextCursor)
		}
	}
	return nil
}

func (s *Service) persistListingPage(ctx context.Context, claim *Job, payload *syncListingPayload, page ListingPage) error {
	return s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" || job.CancelRequested {
			return domain.ErrVersion
		}
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		items, err := port.ListSourceItems(payload.SourceID)
		if err != nil {
			return err
		}
		indices := map[string]SourceItem{}
		for _, item := range items {
			indices[item.ExternalID] = item
		}
		next := *payload
		next.Seen = slices.Clone(payload.Seen)
		next.Warnings = append(slices.Clone(payload.Warnings), page.Warnings...)
		for _, metadata := range page.Items {
			if strings.TrimSpace(metadata.ExternalID) == "" || (metadata.Locator == "" && metadata.UnavailableReason == "") {
				return domain.ErrInvalid
			}
			item, exists := indices[metadata.ExternalID]
			if slices.Contains(next.Seen, metadata.ExternalID) {
				if !exists || !reflect.DeepEqual(item.Metadata, metadata) {
					return serviceError(apierrors.ProviderUnavailable, "Public listing changed while paging", "sync_public_listing_again")
				}
				continue
			}
			next.Seen = append(next.Seen, metadata.ExternalID)
			if !exists {
				item = SourceItem{SourceID: payload.SourceID, ExternalID: metadata.ExternalID, Revision: 1, Metadata: metadata}
			} else if !reflect.DeepEqual(item.Metadata, metadata) {
				item.Revision++
				item.Metadata = metadata
				item.Recommendation = nil
			}
			item.Stale = false
			if err = port.SaveSourceItem(item); err != nil {
				return err
			}
			indices[item.ExternalID] = item
		}
		// Count provider rows, including overlap, to bound actual observations.
		next.Count += len(page.Items)
		next.Done = !page.HasMore || next.Count >= next.Limit
		next.HasMore, next.NextCursor = page.HasMore, page.NextCursor
		if page.NextCursor != nil {
			next.Cursor = *page.NextCursor
		}
		job.Payload, err = json.Marshal(next)
		if err != nil {
			return err
		}
		old := job.Version
		job.Version++
		job.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		if err = s.event(tx, "source_sync_progress", payload.SourceID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "observed": next.Count, "limit": next.Limit, "next_cursor": page.NextCursor, "has_more": page.HasMore}); err != nil {
			return err
		}
		*claim = job
		*payload = next
		return nil
	})
}

func (s *Service) finishTrackingSync(ctx context.Context, claim Job, payload syncListingPayload, hasMore bool, cursor *string) error {
	return s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" || job.CancelRequested {
			return domain.ErrVersion
		}
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		source, err := port.LoadTrackingSource(payload.SourceID)
		if err != nil {
			return err
		}
		items, err := port.ListSourceItems(source.ID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if slices.Contains(payload.Seen, item.ExternalID) {
				item.Stale = false
			} else if payload.InitialCursor == "" {
				item.Stale = true
			} else {
				continue
			}
			if err = port.SaveSourceItem(item); err != nil {
				return err
			}
		}
		now := s.options.Clock()
		old := source.Version
		source.Version++
		source.Status = "succeeded"
		source.LastSuccessAt = &now
		source.LastError = nil
		source.HasMore = hasMore
		source.NextCursor = cursor
		source.Warnings = nonNil(payload.Warnings)
		if err = port.SaveTrackingSource(source, old); err != nil {
			return err
		}
		old = job.Version
		job.Version++
		job.Status = "succeeded"
		job.Error = nil
		job.UpdatedAt = now
		job.ExternalStarted = false
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		return s.event(tx, "job_succeeded", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "source_id": source.ID, "observed": payload.Count, "has_more": hasMore})
	})
}

func (s *Service) processSourceRecommendation(ctx context.Context, claim Job) error {
	var payload recommendationPayload
	if json.Unmarshal(claim.Payload, &payload) != nil {
		return s.failJob(claim, domain.ErrInvalid, false)
	}
	if payload.Result == nil {
		if s.options.ListingRecommender == nil {
			return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.ProviderUnavailable, "Selected CLI recommendation is unavailable", "configure_project_cli"), false)
		}
		workCtx, cleanup, err := s.trackingWork(ctx, &claim)
		if err != nil {
			if errors.Is(err, domain.ErrVersion) {
				return nil
			}
			return err
		}
		defer cleanup()
		config, err := s.options.ListingRecommender.ConfigurationID(workCtx, claim.Caller, payload.Input.ProjectID, payload.Input.CLI)
		if err != nil {
			return s.failTrackingJob(claim, payload.SourceID, err, false)
		}
		if config != payload.Configuration {
			return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.VersionConflict, "Selected CLI configuration changed", "recommend_with_current_cli_configuration"), false)
		}
		payload.Input.JobID, payload.Input.OperationID = claim.JobID, claim.OperationID
		payload.Input.Caller = claim.Caller
		output, err := s.options.ListingRecommender.Recommend(workCtx, payload.Input)
		if err != nil {
			return s.failTrackingJob(claim, payload.SourceID, sourceError(err), externalOutcomeUnknown(err) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
		}
		if len(output) != len(payload.Input.Items) {
			return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.EvidenceMissing, "CLI recommendation did not cover the selected metadata", "inspect_cli_recommendation_output"), false)
		}
		for _, item := range payload.Input.Items {
			r, ok := output[item.ExternalID]
			if !ok || strings.TrimSpace(r.Text) == "" || strings.TrimSpace(r.Reason) == "" || r.Provenance.Processor == "" || r.Provenance.Mode == "" {
				return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.EvidenceMissing, "CLI recommendation lacks text, reason or actual provenance", "inspect_cli_recommendation_output"), false)
			}
			r.MetadataRevision = item.Revision
			r.Status = "succeeded"
			r.ConfigurationID = payload.Configuration
			r.Error = nil
			output[item.ExternalID] = r
		}
		payload.Result = output
		// Durable returned model result precedes final cache publication. If this
		// write fails, mark UNKNOWN; do not repeat a paid call after restart.
		err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
			job, err := tx.LoadJob(claim.JobID)
			if err != nil {
				return err
			}
			if job.Version != claim.Version || job.Status != "running" {
				return domain.ErrVersion
			}
			job.Payload, err = json.Marshal(payload)
			if err != nil {
				return err
			}
			old := job.Version
			job.Version++
			job.UpdatedAt = s.options.Clock()
			if err = tx.SaveJob(job, old); err != nil {
				return err
			}
			claim = job
			return nil
		})
		if err != nil {
			if errors.Is(err, domain.ErrVersion) {
				return nil
			}
			return s.failTrackingJob(claim, payload.SourceID, err, true)
		}
	}
	return s.finishRecommendations(ctx, claim, payload)
}

func (s *Service) finishRecommendations(ctx context.Context, claim Job, payload recommendationPayload) error {
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" || job.CancelRequested {
			return domain.ErrVersion
		}
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		source, err := port.LoadTrackingSource(payload.SourceID)
		if err != nil {
			return err
		}
		items, err := port.ListSourceItems(source.ID)
		if err != nil {
			return err
		}
		for _, snapshot := range payload.Input.Items {
			found := false
			for _, item := range items {
				if item.ExternalID != snapshot.ExternalID {
					continue
				}
				found = true
				if item.Revision != snapshot.Revision || item.Stale || !reflect.DeepEqual(item.Metadata, snapshot.Metadata) {
					return domain.ErrVersion
				}
				recommendation := payload.Result[item.ExternalID]
				item.Recommendation = &recommendation
				if err = port.SaveSourceItem(item); err != nil {
					return err
				}
			}
			if !found {
				return domain.ErrVersion
			}
		}
		old := source.Version
		source.Version++
		source.Status = "succeeded"
		source.LastError = nil
		if err = port.SaveTrackingSource(source, old); err != nil {
			return err
		}
		old = job.Version
		job.Version++
		job.Status = "succeeded"
		job.Error = nil
		job.ExternalStarted = false
		job.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		return s.event(tx, "job_succeeded", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "source_id": source.ID, "recommendation_count": len(payload.Result)})
	})
	if err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return s.failTrackingJob(claim, payload.SourceID, serviceError(apierrors.VersionConflict, "Metadata changed before recommendation publication", "recommend_current_metadata"), false)
		}
		return s.failTrackingJob(claim, payload.SourceID, err, false)
	}
	return nil
}

func (s *Service) failTrackingJob(claim Job, sourceID string, cause error, unknown bool) error {
	if err := s.failJob(claim, cause, unknown); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mapError(s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		source, err := port.LoadTrackingSource(sourceID)
		if err != nil {
			return err
		}
		old := source.Version
		source.Version++
		source.Status = "failed"
		source.LastError = job.Error
		items, err := port.ListSourceItems(sourceID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if claim.Kind == "source_sync" {
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
	}), "")
}
