package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

var _ TrackingService = (*Service)(nil)
var publicDecimalID = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)
var publicDouyinID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,200}$`)

func trackingPort(tx AttentionTx) (TrackingTx, error) {
	port, ok := tx.(TrackingTx)
	if !ok {
		return nil, serviceError(apierrors.UnsupportedCapability, "Tracking storage is not configured", "configure_tracking_storage")
	}
	return port, nil
}
func trackingResult(tx TrackingTx, source TrackingSource) (TrackingSourceResult, error) {
	items, err := tx.ListSourceItems(source.ID)
	if err != nil {
		return TrackingSourceResult{}, err
	}
	// Job state is authoritative; old metadata links do not invent body success.
	for i := range items {
		if items[i].ImportJobID != nil {
			job, err := tx.LoadJob(*items[i].ImportJobID)
			if err != nil {
				return TrackingSourceResult{}, err
			}
			items[i].MaterialID = job.MaterialID
		}
	}
	return TrackingSourceResult{SchemaVersion: 1, Source: source, Items: nonNil(items), Warnings: nonNil(source.Warnings), NextCursor: source.NextCursor, HasMore: source.HasMore, Jobs: []ImportJobResult{}}, nil
}

// Source identity comes from a validated official public locator/UID, never an
// arbitrary caller-chosen key. Multiple accounts and explicit folders coexist.
func normalizeTrackingSource(c BindTrackingSourceCommand) (TrackingSource, error) {
	s := TrackingSource{Platform: c.Platform, SourceKind: c.SourceKind, OwnerID: strings.TrimSpace(c.OwnerID), ExternalID: strings.TrimSpace(c.ExternalID), Title: strings.TrimSpace(c.Title), Status: "pending", Warnings: []string{}}
	if c.SourceKind != "uploads" && c.SourceKind != "favorites" {
		return s, domain.ErrInvalid
	}
	locator := strings.TrimSpace(c.Locator)
	if s.Platform == "bilibili" && s.SourceKind == "uploads" && publicDecimalID.MatchString(locator) {
		locator = "https://space.bilibili.com/" + locator + "/video"
	}
	if locator == "" && s.Platform == "bilibili" && s.SourceKind == "uploads" {
		uid := s.ExternalID
		if uid == "" {
			uid = s.OwnerID
		}
		locator = "https://space.bilibili.com/" + uid + "/video"
	}
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return s, serviceError(apierrors.ValidationFailed, "An official public HTTPS source is required", "provide_public_profile_or_folder")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	switch s.Platform {
	case "bilibili":
		owner, id := "", ""
		if u.Hostname() == "space.bilibili.com" && len(parts) >= 1 && publicDecimalID.MatchString(parts[0]) {
			owner = parts[0]
			if s.SourceKind == "uploads" && (len(parts) == 1 || (len(parts) == 2 && parts[1] == "video") || (len(parts) == 3 && parts[1] == "upload" && parts[2] == "video")) {
				id = owner
				s.Locator = "https://space.bilibili.com/" + owner + "/video"
			}
			if s.SourceKind == "favorites" && len(parts) == 2 && parts[1] == "favlist" {
				id = u.Query().Get("fid")
				s.Locator = "https://space.bilibili.com/" + owner + "/favlist?fid=" + id
			}
		} else if s.SourceKind == "favorites" && (u.Hostname() == "www.bilibili.com" || u.Hostname() == "bilibili.com") && strings.HasPrefix(u.Path, "/medialist/detail/ml") {
			id = strings.TrimPrefix(u.Path, "/medialist/detail/ml")
			owner = s.OwnerID
			s.Locator = "https://www.bilibili.com/medialist/detail/ml" + id
		}
		if !publicDecimalID.MatchString(id) {
			return s, serviceError(apierrors.ValidationFailed, "Choose a public UID for uploads or an explicit public favorite folder", "choose_public_folder_or_profile")
		}
		if (s.ExternalID != "" && s.ExternalID != id) || (s.OwnerID != "" && owner != "" && s.OwnerID != owner) {
			return s, serviceError(apierrors.ValidationFailed, "Source identity does not match its public URL", "correct_public_source_identity")
		}
		s.ExternalID, s.OwnerID = id, owner
	case "douyin":
		if u.Hostname() != "www.douyin.com" || len(parts) != 2 || parts[0] != "user" || parts[1] == "self" {
			return s, serviceError(apierrors.ScopeDenied, "Douyin /user/self requires login; public-only tracking needs a real public profile", "provide_public_douyin_profile")
		}
		if !publicDouyinID.MatchString(parts[1]) || (s.ExternalID != "" && s.ExternalID != parts[1]) {
			return s, domain.ErrInvalid
		}
		if s.SourceKind == "favorites" {
			return s, serviceError(apierrors.UnsupportedCapability, "Anonymous Douyin favorite collections are not supported by the upstream adapter", "provide_supported_public_collection")
		}
		s.ExternalID, s.OwnerID = parts[1], parts[1]
		s.Locator = "https://www.douyin.com/user/" + parts[1]
	default:
		return s, serviceError(apierrors.UnsupportedCapability, "This platform has no public listing adapter", "choose_supported_public_platform")
	}
	return s, nil
}

func (s *Service) ListTrackingSources(ctx context.Context, p Principal) (apierrors.ListResult, error) {
	if err := authorize(p, true); err != nil {
		return apierrors.EmptyList(), err
	}
	var rows []TrackingSource
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		rows, err = port.ListTrackingSources()
		return err
	})
	return listResult(rows), mapError(err, "")
}
func (s *Service) GetTrackingSource(ctx context.Context, p Principal, id string) (TrackingSourceResult, error) {
	if err := authorize(p, true); err != nil {
		return TrackingSourceResult{}, err
	}
	var result TrackingSourceResult
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		source, err := port.LoadTrackingSource(id)
		if err != nil {
			return err
		}
		result, err = trackingResult(port, source)
		return err
	})
	return result, mapError(err, "")
}
func (s *Service) BindTrackingSource(ctx context.Context, p Principal, c BindTrackingSourceCommand) (TrackingSourceResult, error) {
	return command(s, ctx, p, c.CommandMeta, "BindTrackingSource", c, func(tx AttentionTx) (TrackingSourceResult, error) {
		if c.ExpectedVersion != 1 {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		source, err := normalizeTrackingSource(c)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		port, err := trackingPort(tx)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		rows, err := port.ListTrackingSources()
		if err != nil {
			return TrackingSourceResult{}, err
		}
		for _, row := range rows {
			if row.Platform == source.Platform && row.SourceKind == source.SourceKind && row.ExternalID == source.ExternalID {
				return trackingResult(port, row)
			}
		}
		source.ID, source.Version = rand.Text(), 1
		if err = port.SaveTrackingSource(source, 0); err != nil {
			return TrackingSourceResult{}, err
		}
		if err = s.event(tx, "tracking_source_bound", source.ID, source.Version, c.CommandMeta, map[string]any{"platform": source.Platform, "source_kind": source.SourceKind, "external_id": source.ExternalID}); err != nil {
			return TrackingSourceResult{}, err
		}
		return trackingResult(port, source)
	})
}

type syncListingPayload struct {
	SourceID      string
	Limit         int
	Cursor        string
	InitialCursor string
	Seen          []string
	Count         int
	Warnings      []string
	Done          bool
	HasMore       bool
	NextCursor    *string
}
type recommendationPayload struct {
	SourceID      string
	Configuration string
	Input         ListingRecommendationInput
	Result        map[string]SourceRecommendation
}

// Only fixed input and trusted processing scope identify paid work. Transient
// recommendation/selection/job state and request keys cannot justify a new call.
func recommendationInputKey(payload recommendationPayload) string {
	items := make([]SourceItem, 0, len(payload.Input.Items))
	for _, item := range payload.Input.Items {
		items = append(items, SourceItem{SourceID: item.SourceID, ExternalID: item.ExternalID, Revision: item.Revision, Metadata: item.Metadata})
	}
	slices.SortFunc(items, func(a, b SourceItem) int { return strings.Compare(a.ExternalID, b.ExternalID) })
	fixed := recommendationPayload{SourceID: payload.SourceID, Configuration: payload.Configuration, Input: ListingRecommendationInput{Caller: payload.Input.Caller, ProjectID: payload.Input.ProjectID, CLI: payload.Input.CLI, Items: items}}
	data, _ := json.Marshal(fixed)
	return digestBytes(data)
}

func findRecommendationWork(tx AttentionTx, wanted recommendationPayload) (*Job, *recommendationPayload, error) {
	jobs, err := tx.ListJobs()
	if err != nil {
		return nil, nil, err
	}
	key := recommendationInputKey(wanted)
	var found *Job
	var cached *recommendationPayload
	for _, job := range jobs {
		if job.Kind != "source_recommendation" || job.Caller != wanted.Input.Caller {
			continue
		}
		var payload recommendationPayload
		if json.Unmarshal(job.Payload, &payload) != nil {
			continue
		}
		payload.Input.Caller = job.Caller
		if recommendationInputKey(payload) != key {
			continue
		}
		// Scan persisted payloads as well as current keys so existing pre-policy
		// jobs (including old UNKNOWN) remain authoritative after an upgrade.
		if job.DeliveryUnknown || (job.Error != nil && job.Error.Code == apierrors.DeliveryUnknown) {
			return nil, nil, domain.ErrUnknown
		}
		if found == nil || job.CreatedAt.After(found.CreatedAt) {
			copy := job
			found, cached = &copy, &payload
		}
	}
	return found, cached, nil
}

func (s *Service) queueTrackingJob(tx AttentionTx, p Principal, m CommandMeta, kind string, payload any) (Job, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return Job{}, err
	}
	now := s.options.Clock()
	job := Job{SchemaVersion: 1, JobID: rand.Text(), OperationID: rand.Text(), Kind: kind, Status: "queued", Version: 1, DedupeKey: digestBytes([]byte(kind + ":" + p.ID + ":" + m.IdempotencyKey)), Payload: encoded, Caller: p, MaxAttempts: s.options.MaxAttempts, CreatedAt: now, UpdatedAt: now, DeadlineAt: now.Add(s.options.JobTimeout)}
	if err = tx.SaveJob(job, 0); err != nil {
		return Job{}, err
	}
	err = s.event(tx, "job_queued", job.JobID, 1, m, map[string]any{"job_id": job.JobID, "kind": kind, "operation_id": job.OperationID})
	return job, err
}

func (s *Service) SyncTrackingSource(ctx context.Context, p Principal, id string, c SyncTrackingSourceCommand) (TrackingSourceResult, error) {
	input := struct {
		ID string
		SyncTrackingSourceCommand
	}{id, c}
	result, err := command(s, ctx, p, c.CommandMeta, "SyncTrackingSource", input, func(tx AttentionTx) (TrackingSourceResult, error) {
		port, err := trackingPort(tx)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		source, err := port.LoadTrackingSource(id)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		if source.Version != c.ExpectedVersion {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		if source.Status == "syncing" || source.Status == "recommending" {
			return TrackingSourceResult{}, serviceError(apierrors.VersionConflict, "Source work is still active", "wait_for_source_job")
		}
		limit := c.Limit
		if limit == 0 {
			limit = 100
		}
		if limit < 1 || limit > 100 {
			return TrackingSourceResult{}, domain.ErrInvalid
		}
		job, err := s.queueTrackingJob(tx, p, c.CommandMeta, "source_sync", syncListingPayload{SourceID: id, Limit: limit, Cursor: c.Cursor, InitialCursor: c.Cursor, Seen: []string{}, Warnings: []string{}})
		if err != nil {
			return TrackingSourceResult{}, err
		}
		old := source.Version
		source.Version++
		source.Status = "syncing"
		if err = port.SaveTrackingSource(source, old); err != nil {
			return TrackingSourceResult{}, err
		}
		result, err := trackingResult(port, source)
		result.Jobs = []ImportJobResult{{SchemaVersion: 1, JobID: job.JobID, Status: job.Status}}
		return result, err
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

func selectedItems(items []SourceItem, selections []SourceItemSelection) ([]SourceItem, error) {
	result := []SourceItem{}
	seen := map[string]bool{}
	for _, selection := range selections {
		if seen[selection.ExternalID] {
			continue
		}
		seen[selection.ExternalID] = true
		found := false
		for _, item := range items {
			if item.ExternalID != selection.ExternalID {
				continue
			}
			found = true
			if item.Revision != selection.Revision {
				return nil, domain.ErrVersion
			}
			if item.Stale || item.Metadata.UnavailableReason != "" || item.Metadata.Locator == "" {
				return nil, serviceError(apierrors.EvidenceMissing, "This listing item is stale or unavailable", "refresh_listing_or_choose_available_item")
			}
			result = append(result, item)
			break
		}
		if !found {
			return nil, apierrors.NewNotFound("source_item", selection.ExternalID)
		}
	}
	if len(result) == 0 || len(result) > 100 {
		return nil, domain.ErrInvalid
	}
	return result, nil
}

func (s *Service) RecommendSourceItems(ctx context.Context, p Principal, id string, c RecommendSourceItemsCommand) (TrackingSourceResult, error) {
	if err := authorize(p, true); err != nil {
		return TrackingSourceResult{}, err
	}
	if s.options.ListingRecommender == nil {
		return TrackingSourceResult{}, serviceError(apierrors.ProviderUnavailable, "No selected CLI recommendation processor is configured", "select_configured_project_cli")
	}
	if c.ProjectID == "" || c.CLI == "" {
		return TrackingSourceResult{}, serviceError(apierrors.ScopeDenied, "Select a project with explicit model-processing permission and a configured CLI", "select_permitted_project_cli")
	}
	configuration, err := s.options.ListingRecommender.ConfigurationID(ctx, p, c.ProjectID, c.CLI)
	if err != nil {
		return TrackingSourceResult{}, mapError(err, c.RequestID)
	}
	input := struct {
		ID string
		RecommendSourceItemsCommand
	}{id, c}
	result, err := command(s, ctx, p, c.CommandMeta, "RecommendSourceItems", input, func(tx AttentionTx) (TrackingSourceResult, error) {
		port, err := trackingPort(tx)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		source, err := port.LoadTrackingSource(id)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		if source.Version != c.ExpectedVersion {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		if source.Status == "syncing" {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		items, err := port.ListSourceItems(id)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		selections := c.Items
		if len(selections) == 0 {
			for _, item := range items {
				if !item.Stale && item.Metadata.UnavailableReason == "" && item.Metadata.Locator != "" {
					selections = append(selections, SourceItemSelection{item.ExternalID, item.Revision})
				}
			}
		}
		chosen, err := selectedItems(items, selections)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		payload := recommendationPayload{SourceID: id, Configuration: configuration, Input: ListingRecommendationInput{Caller: p, ProjectID: c.ProjectID, CLI: c.CLI, Items: chosen}}
		oldJob, cached, err := findRecommendationWork(tx, payload)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		if oldJob != nil {
			result, err := trackingResult(port, source)
			result.Jobs = []ImportJobResult{{SchemaVersion: 1, JobID: oldJob.JobID, Status: oldJob.Status}}
			if oldJob.Status == "succeeded" {
				for i := range result.Items {
					if recommendation, ok := cached.Result[result.Items[i].ExternalID]; ok && recommendation.MetadataRevision == result.Items[i].Revision {
						result.Items[i].Recommendation = &recommendation
					}
				}
			}
			return result, err
		}
		if source.Status == "recommending" {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		job, err := s.queueTrackingJob(tx, p, c.CommandMeta, "source_recommendation", payload)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		for _, item := range chosen {
			item.Recommendation = &SourceRecommendation{Status: "pending", MetadataRevision: item.Revision, ConfigurationID: configuration}
			if err = port.SaveSourceItem(item); err != nil {
				return TrackingSourceResult{}, err
			}
		}
		old := source.Version
		source.Version++
		source.Status = "recommending"
		if err = port.SaveTrackingSource(source, old); err != nil {
			return TrackingSourceResult{}, err
		}
		result, err := trackingResult(port, source)
		result.Jobs = []ImportJobResult{{SchemaVersion: 1, JobID: job.JobID, Status: job.Status}}
		return result, err
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

func (s *Service) SelectSourceItems(ctx context.Context, p Principal, id string, c SelectSourceItemsCommand) (TrackingSourceResult, error) {
	input := struct {
		ID string
		SelectSourceItemsCommand
	}{id, c}
	result, err := command(s, ctx, p, c.CommandMeta, "SelectSourceItems", input, func(tx AttentionTx) (TrackingSourceResult, error) {
		port, err := trackingPort(tx)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		source, err := port.LoadTrackingSource(id)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		if source.Version != c.ExpectedVersion {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		if source.Status == "syncing" || source.Status == "recommending" {
			return TrackingSourceResult{}, domain.ErrVersion
		}
		items, err := port.ListSourceItems(id)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		chosen, err := selectedItems(items, c.Items)
		if err != nil {
			return TrackingSourceResult{}, err
		}
		jobs := []ImportJobResult{}
		for _, item := range chosen {
			if item.Recommendation == nil || item.Recommendation.Status != "succeeded" || item.Recommendation.MetadataRevision != item.Revision {
				return TrackingSourceResult{}, serviceError(apierrors.EvidenceMissing, "Complete CLI metadata recommendation before selecting body imports", "recommend_current_metadata_first")
			}
			job, err := s.enqueueImport(tx, p, ImportMaterialCommand{CommandMeta: CommandMeta{SchemaVersion: 1, ExpectedVersion: 1, RequestID: c.RequestID, IdempotencyKey: c.IdempotencyKey + ":" + item.ExternalID}, SourceLocator: item.Metadata.Locator, Kind: "video", Adapter: "summarize_url", Title: item.Metadata.Title, CollectionReason: c.CollectionReason})
			if err != nil {
				return TrackingSourceResult{}, err
			}
			jobs = append(jobs, job)
			item.Selected = true
			item.ImportJobID = &job.JobID
			if err = port.SaveSourceItem(item); err != nil {
				return TrackingSourceResult{}, err
			}
		}
		old := source.Version
		source.Version++
		if err = port.SaveTrackingSource(source, old); err != nil {
			return TrackingSourceResult{}, err
		}
		result, err := trackingResult(port, source)
		result.Jobs = jobs
		if err != nil {
			return result, err
		}
		if err = s.event(tx, "source_items_selected", id, source.Version, c.CommandMeta, map[string]any{"items": chosen, "jobs": jobs}); err != nil {
			return result, err
		}
		return result, nil
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

// SyncSourcesOnce is an owned startup hook, not a timer or automatic body import.
// The durable worker queue handles progress, recovery and bounded retries.
func (s *Service) SyncSourcesOnce(ctx context.Context) error {
	var sources []TrackingSource
	var jobs []Job
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		port, err := trackingPort(tx)
		if err != nil {
			return err
		}
		sources, err = port.ListTrackingSources()
		if err != nil {
			return err
		}
		jobs, err = tx.ListJobs()
		return err
	})
	if err != nil {
		return mapError(err, "")
	}
	for _, source := range sources {
		if source.Status == "syncing" || source.Status == "recommending" {
			continue
		}
		recoverable := false
		for _, job := range jobs {
			if trackingJobSource(job) == source.ID && job.Kind == "source_sync" && job.Status == "failed" && !job.DeliveryUnknown && job.Error != nil && job.Error.Retryable && job.Attempts < job.MaxAttempts && s.options.Clock().Before(job.DeadlineAt) {
				superseded := false
				for _, other := range jobs {
					if other.JobID != job.JobID && trackingJobSource(other) == source.ID && other.CreatedAt.After(job.CreatedAt) {
						superseded = true
						break
					}
				}
				if !superseded {
					recoverable = true
					break
				}
			}
		}
		if recoverable {
			continue
		}
		key := rand.Text()
		err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
			port, err := trackingPort(tx)
			if err != nil {
				return err
			}
			current, err := port.LoadTrackingSource(source.ID)
			if err != nil {
				return err
			}
			if current.Version != source.Version {
				return domain.ErrVersion
			}
			_, err = s.queueTrackingJob(tx, Principal{ID: "server-startup", Kind: "system"}, CommandMeta{RequestID: key, IdempotencyKey: key}, "source_sync", syncListingPayload{SourceID: source.ID, Limit: 100, Seen: []string{}, Warnings: []string{}})
			if err != nil {
				return err
			}
			old := current.Version
			current.Version++
			current.Status = "syncing"
			return port.SaveTrackingSource(current, old)
		})
		if err != nil {
			return mapError(err, "")
		}
	}
	s.signal()
	return nil
}

func (s *Service) ListSourceCollections(ctx context.Context, p Principal, platform, ownerID string) (apierrors.ListResult, error) {
	if err := authorize(p, true); err != nil {
		return apierrors.EmptyList(), err
	}
	if s.options.CollectionReader == nil {
		return apierrors.EmptyList(), serviceError(apierrors.ProviderUnavailable, "No public collection reader is configured", "configure_public_listing_reader")
	}
	rows, err := s.options.CollectionReader.ListCollections(ctx, platform, ownerID)
	return listResult(rows), mapError(err, "")
}
