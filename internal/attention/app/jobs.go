package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *Service) ImportMaterial(ctx context.Context, p Principal, c ImportMaterialCommand) (ImportJobResult, error) {
	result, err := command(s, ctx, p, c.CommandMeta, "ImportMaterial", c, func(tx AttentionTx) (ImportJobResult, error) {
		if c.ExpectedVersion != 1 || strings.TrimSpace(c.SourceLocator) == "" || !slices.Contains([]string{"paper", "video", "text", "file"}, c.Kind) {
			return ImportJobResult{}, domain.ErrInvalid
		}
		// A supplied export is a fixed snapshot. An unversioned remote URL or
		// local file can change, so a new explicit command must reread it. Its
		// receipt still prevents duplicate delivery of the same command; the
		// canonical source key and actual digest prevent duplicate revisions.
		refresh := ""
		if c.ExportText == "" {
			refresh = c.IdempotencyKey
		}
		identity := struct{ Source, Key, Kind, Adapter, Text, File, Digest, Refresh string }{c.SourceLocator, c.SourceKey, c.Kind, c.Adapter, c.ExportText, c.LocalFileRef, c.ContentDigest, refresh}
		bytes, _ := json.Marshal(identity)
		dedupeKey := digestBytes(bytes)
		old, err := tx.FindJobByDedupeKey(dedupeKey)
		if err == nil {
			return ImportJobResult{SchemaVersion: 1, JobID: old.JobID, Status: old.Status}, nil
		}
		if !isMissing(err) {
			return ImportJobResult{}, err
		}
		payload, err := json.Marshal(c)
		if err != nil {
			return ImportJobResult{}, err
		}
		now := s.options.Clock()
		job := Job{SchemaVersion: 1, JobID: rand.Text(), Status: "queued", Kind: "import", Version: 1, DedupeKey: dedupeKey, OperationID: rand.Text(), MaxAttempts: s.options.MaxAttempts, CreatedAt: now, UpdatedAt: now, DeadlineAt: now.Add(s.options.JobTimeout), Payload: payload, Caller: p}
		if err = tx.SaveJob(job, 0); err != nil {
			return ImportJobResult{}, err
		}
		if err = s.event(tx, "job_queued", job.JobID, job.Version, c.CommandMeta, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID}); err != nil {
			return ImportJobResult{}, err
		}
		return ImportJobResult{SchemaVersion: 1, JobID: job.JobID, Status: job.Status}, nil
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

func (s *Service) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Service) ListJobs(ctx context.Context, p Principal) (apierrors.ListResult, error) {
	if err := authorize(p, false); err != nil {
		return apierrors.EmptyList(), err
	}
	var jobs []Job
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error { var err error; jobs, err = tx.ListJobs(); return err })
	return listResult(jobs), mapError(err, "")
}

func (s *Service) GetJob(ctx context.Context, p Principal, id string) (Job, error) {
	if err := authorize(p, false); err != nil {
		return Job{}, err
	}
	var job Job
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error { var err error; job, err = tx.LoadJob(id); return err })
	return job, mapError(err, "")
}

func jobRules(j Job) domain.JobRules {
	// SourceReader only downloads sources or parses existing exports. Ordinary
	// imports have no paid/effectful action to reconcile. Other job kinds retain
	// the conservative external-start recovery rule for future model adapters.
	return domain.JobRules{State: j.Status, Attempts: j.Attempts, MaxAttempts: j.MaxAttempts, Deadline: j.DeadlineAt, ExternalStarted: j.ExternalStarted && j.Kind != "import", DeliveryUnknown: j.DeliveryUnknown}
}

func (s *Service) RetryJob(ctx context.Context, p Principal, id string, m CommandMeta) (Job, error) {
	input := struct {
		ID string
		CommandMeta
	}{id, m}
	result, err := command(s, ctx, p, m, "RetryJob", input, func(tx AttentionTx) (Job, error) {
		job, err := tx.LoadJob(id)
		if err != nil {
			return Job{}, err
		}
		if job.Version != m.ExpectedVersion {
			return Job{}, domain.ErrVersion
		}
		rules, err := jobRules(job).Retry(s.options.Clock())
		if err != nil {
			return Job{}, err
		}
		job.Status = rules.State
		job.Version++
		job.ExternalStarted = false
		job.CancelRequested = false
		job.Error = nil
		job.UpdatedAt = s.options.Clock()
		// Preserve the original deadline, attempts, payload and operation identity.
		if err = tx.SaveJob(job, m.ExpectedVersion); err != nil {
			return Job{}, err
		}
		if err = s.event(tx, "job_retried", id, job.Version, m, map[string]any{"job_id": id, "operation_id": job.OperationID}); err != nil {
			return Job{}, err
		}
		return job, nil
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

func (s *Service) CancelJob(ctx context.Context, p Principal, id string, m CommandMeta) (Job, error) {
	input := struct {
		ID string
		CommandMeta
	}{id, m}
	changed := false
	result, err := command(s, ctx, p, m, "CancelJob", input, func(tx AttentionTx) (Job, error) {
		job, err := tx.LoadJob(id)
		if err != nil {
			return Job{}, err
		}
		if job.Version != m.ExpectedVersion {
			return Job{}, domain.ErrVersion
		}
		rules, err := jobRules(job).Cancel()
		if err != nil {
			return Job{}, err
		}
		job.Status = rules.State
		job.DeliveryUnknown = rules.DeliveryUnknown
		job.CancelRequested = true
		job.Version++
		job.UpdatedAt = s.options.Clock()
		if job.DeliveryUnknown {
			job.Error = mapError(domain.ErrUnknown, m.RequestID).(*apierrors.ServiceError)
		}
		if err = tx.SaveJob(job, m.ExpectedVersion); err != nil {
			return Job{}, err
		}
		if err = s.event(tx, "job_cancelled", id, job.Version, m, map[string]any{"job_id": id, "operation_id": job.OperationID, "delivery_unknown": job.DeliveryUnknown}); err != nil {
			return Job{}, err
		}
		changed = true
		return job, nil
	})
	if err == nil && changed {
		s.activeMu.Lock()
		work := s.activeJobs[id]
		s.activeMu.Unlock()
		if work != nil {
			work.cancel()
		}
	}
	return result, err
}

// RecoverJobs is called once at service startup before claiming. Running jobs
// are interrupted records, never silently requeued. Unknown external effects
// retain the original operation ID and require reconciliation.
func (s *Service) RecoverJobs(ctx context.Context) error {
	return mapError(s.repo.WithTx(ctx, func(tx AttentionTx) error {
		jobs, err := tx.ListJobs()
		if err != nil {
			return err
		}
		for _, job := range jobs {
			if job.Status != "running" {
				continue
			}
			rules := jobRules(job).Recover()
			old := job.Version
			job.Status = rules.State
			job.DeliveryUnknown = rules.DeliveryUnknown
			job.Version++
			job.UpdatedAt = s.options.Clock()
			if job.DeliveryUnknown {
				job.Error = mapError(domain.ErrUnknown, "").(*apierrors.ServiceError)
			} else {
				job.Error = serviceError(apierrors.ProviderUnavailable, "Local worker was interrupted", "retry_job")
			}
			if err = tx.SaveJob(job, old); err != nil {
				return err
			}
			if err = s.event(tx, "job_recovered", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID, "delivery_unknown": job.DeliveryUnknown}); err != nil {
				return err
			}
		}
		return nil
	}), "")
}

// Run uses a fixed Go worker pool and the durable database queue. No paid model
// backend is constructed here. It stops claims on cancellation and waits for
// cooperative adapters before returning; adapters must honor their context.
func (s *Service) Run(ctx context.Context) error {
	s.runMu.Lock()
	if s.running {
		s.runMu.Unlock()
		return serviceError(apierrors.VersionConflict, "Attention workers already started", "reuse_running_service")
	}
	s.running = true
	s.runMu.Unlock()
	defer func() { s.runMu.Lock(); s.running = false; s.runMu.Unlock() }()
	if err := s.RecoverJobs(ctx); err != nil {
		return err
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errorsFound := make(chan error, s.options.WorkerConcurrency)
	var workers sync.WaitGroup
	for i := 0; i < s.options.WorkerConcurrency; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			for {
				if workerCtx.Err() != nil {
					return
				}
				worked, err := s.ProcessNextJob(workerCtx)
				if err != nil {
					if workerCtx.Err() == nil {
						errorsFound <- err
						cancel()
					}
					return
				}
				if worked {
					continue
				}
				select {
				case <-workerCtx.Done():
					return
				case <-s.wake:
				case <-ticker.C:
				}
			}
		}()
	}
	workers.Wait()
	select {
	case err := <-errorsFound:
		return err
	default:
		return nil
	}
}

// ProcessNextJob is also useful for an owned, deterministic batch entrypoint.
// Claim CAS and recording external start precede the source read.
func (s *Service) ProcessNextJob(ctx context.Context) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	var claimed Job
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		jobs, err := tx.ListJobs()
		if err != nil {
			return err
		}
		for _, job := range jobs {
			if job.Status != "queued" {
				continue
			}
			rules, err := jobRules(job).Start(s.options.Clock())
			old := job.Version
			if err != nil {
				job.Status = "failed"
				job.Error = serviceError(apierrors.ValidationFailed, "Job attempt limit or total deadline reached", "create_new_work_after_review")
				job.Version++
				job.UpdatedAt = s.options.Clock()
				if err = tx.SaveJob(job, old); err != nil {
					return err
				}
				if err = s.event(tx, "job_failed", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID}); err != nil {
					return err
				}
				continue
			}
			job.Status = rules.State
			job.Attempts = rules.Attempts
			job.Version++
			job.UpdatedAt = s.options.Clock()
			if err = tx.SaveJob(job, old); err != nil {
				return err
			}
			claimed = job
			break
		}
		return nil
	})
	if err != nil || claimed.JobID == "" {
		return false, mapError(err, "")
	}
	var c ImportMaterialCommand
	if err = json.Unmarshal(claimed.Payload, &c); err != nil {
		return true, s.failJob(claimed, domain.ErrInvalid, false)
	}
	if claimed.Kind != "import" {
		return true, s.failJob(claimed, domain.ErrInvalid, false)
	}
	if s.sources == nil {
		return true, s.failJob(claimed, serviceError(apierrors.ProviderUnavailable, "No source reader is configured", "configure_import_adapter"), false)
	}
	if s.objects == nil {
		return true, s.failJob(claimed, serviceError(apierrors.ProviderUnavailable, "No object store is configured", "configure_object_store"), false)
	}
	remaining := claimed.DeadlineAt.Sub(s.options.Clock())
	workCtx, cancel := context.WithTimeout(ctx, remaining)
	work := &activeWork{cancel: cancel}
	s.activeMu.Lock()
	s.activeJobs[claimed.JobID] = work
	s.activeMu.Unlock()
	defer func() {
		cancel()
		s.activeMu.Lock()
		if s.activeJobs[claimed.JobID] == work {
			delete(s.activeJobs, claimed.JobID)
		}
		s.activeMu.Unlock()
	}()
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		current, err := tx.LoadJob(claimed.JobID)
		if err != nil {
			return err
		}
		if current.Status != "running" || current.Version != claimed.Version {
			return domain.ErrVersion
		}
		current.ExternalStarted = true
		current.Version++
		current.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(current, claimed.Version); err != nil {
			return err
		}
		claimed = current
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return true, nil
		}
		return true, s.failJob(claimed, err, false)
	}
	source, err := s.sources.ReadSource(workCtx, c)
	if err != nil {
		return true, s.failJob(claimed, sourceError(err), externalOutcomeUnknown(err))
	}
	if workCtx.Err() != nil {
		return true, s.failJob(claimed, sourceError(workCtx.Err()), false)
	}
	if strings.TrimSpace(source.SourceKey) == "" || strings.TrimSpace(source.SourceLocator) == "" || source.Kind != c.Kind || strings.TrimSpace(source.Text) == "" {
		return true, s.failJob(claimed, domain.ErrInvalid, false)
	}
	digest := digestBytes([]byte(source.Text))
	if c.ContentDigest != "" && c.ContentDigest != digest {
		return true, s.failJob(claimed, serviceError(apierrors.ValidationFailed, "Content digest does not match actual source bytes", "correct_content_digest"), false)
	}
	objectRef, err := s.objects.Publish(workCtx, []byte(source.Text))
	if err != nil {
		return true, s.failJob(claimed, err, false)
	}
	attachments := make([]AttachmentRef, 0, len(source.Attachments))
	for _, attachment := range source.Attachments {
		ref, err := s.objects.Publish(workCtx, attachment.Data)
		if err != nil {
			return true, s.failJob(claimed, err, false)
		}
		attachments = append(attachments, AttachmentRef{Name: attachment.Name, MediaType: attachment.MediaType, ObjectRef: ref, SourceLocator: attachment.SourceLocator})
	}
	if err = s.finishImport(workCtx, claimed, c, source, digest, objectRef, attachments); err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return true, nil
		}
		return true, s.failJob(claimed, err, false)
	}
	return true, nil
}

func externalOutcomeUnknown(err error) bool {
	if errors.Is(err, domain.ErrUnknown) {
		return true
	}
	var e *apierrors.ServiceError
	if errors.As(err, &e) {
		return e.Code == apierrors.DeliveryUnknown
	}
	return false
}

func sourceError(err error) error {
	var e *apierrors.ServiceError
	if errors.As(err, &e) || errors.Is(err, domain.ErrUnknown) || errors.Is(err, domain.ErrInvalid) {
		return err
	}
	e = serviceError(apierrors.ProviderUnavailable, "The source read was interrupted or unavailable", "retry_import_job")
	e.Retryable = true
	return e
}

func (s *Service) failJob(claim Job, cause error, unknown bool) error {
	// Result persistence must survive cancellation of the request or worker.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mapError(s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" {
			return nil
		}
		old := job.Version
		job.Status = "failed"
		job.Version++
		job.UpdatedAt = s.options.Clock()
		job.DeliveryUnknown = unknown
		if unknown {
			job.Error = mapError(domain.ErrUnknown, "").(*apierrors.ServiceError)
		} else {
			job.Error = mapError(cause, "").(*apierrors.ServiceError)
		}
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		return s.event(tx, "job_failed", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID, "delivery_unknown": job.DeliveryUnknown})
	}), "")
}

func (s *Service) finishImport(ctx context.Context, claim Job, c ImportMaterialCommand, source ImportedSource, digest, objectRef string, attachments []AttachmentRef) error {
	return s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Status != "running" || job.Version != claim.Version || job.CancelRequested {
			return domain.ErrVersion
		}
		row, err := tx.FindMaterialBySourceKey(source.SourceKey)
		if err != nil && !isMissing(err) {
			return err
		}
		now := s.options.Clock()
		oldVersion := row.Material.Version
		if isMissing(err) {
			row = MaterialDetail{SchemaVersion: 1, Material: Material{ID: rand.Text(), Version: 0, Lifecycle: "active", Title: source.Title, Kind: source.Kind, SourceLocator: source.SourceLocator, CollectionReason: c.CollectionReason, SourceSpans: nonNil(source.SourceSpans), CreatedAt: now}, Revisions: []MaterialRevision{}, Distillations: []Distillation{}, Uses: []Usage{}}
		}
		if row.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Source material was withdrawn", "review_material_lifecycle")
		}
		digests := make([]string, 0, len(row.Revisions))
		for _, r := range row.Revisions {
			digests = append(digests, r.ContentDigest)
		}
		revision, reused, err := domain.NextMaterialRevision(digests, digest)
		if err != nil {
			return err
		}
		if !reused {
			row.Revisions = append(row.Revisions, MaterialRevision{MaterialID: row.Material.ID, Revision: revision, SourceKey: source.SourceKey, SourceLocator: source.SourceLocator, ContentDigest: digest, ObjectRef: objectRef, SourceSpans: nonNil(source.SourceSpans), Provenance: source.Provenance, CreatedAt: now, Summary: source.Summary, Attachments: attachments})
			row.Material.Version++
			row.Material.CurrentRevision = revision
			row.Material.ImportStatus = "succeeded"
			row.Material.Title = source.Title
			row.Material.SourceLocator = source.SourceLocator
			row.Material.SourceSpans = nonNil(source.SourceSpans)
			if err = tx.SaveMaterial(row, oldVersion); err != nil {
				return err
			}
			if err = s.event(tx, "material_imported", row.Material.ID, row.Material.Version, c.CommandMeta, map[string]any{"material_id": row.Material.ID, "revision": revision, "job_id": job.JobID}); err != nil {
				return err
			}
		}
		oldJobVersion := job.Version
		job.Status = "succeeded"
		job.Version++
		job.UpdatedAt = now
		job.Error = nil
		job.DeliveryUnknown = false
		job.MaterialID = &row.Material.ID
		job.MaterialRevision = &revision
		if err = tx.SaveJob(job, oldJobVersion); err != nil {
			return err
		}
		return s.event(tx, "job_succeeded", job.JobID, job.Version, c.CommandMeta, map[string]any{"job_id": job.JobID, "material_id": row.Material.ID, "revision": revision, "reused": reused})
	})
}
