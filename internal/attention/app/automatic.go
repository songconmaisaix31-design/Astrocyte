package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

var _ AutomaticDistillationService = (*Service)(nil)

// Payload freezes the caller-selected round and actual processor configuration.
// It is persisted before external calls, alongside the original operation ID.
type automaticPayload struct {
	Command         RequestDistillationCommand
	Snapshots       []SourceSnapshot
	ProcessorConfig string
	EffectiveConfig string
	ReuseKey        string
	Result          *DistillationOutput
	SpaceID         string
}

func (s *Service) GetDistillerStatus(ctx context.Context, p Principal) (DistillerStatus, error) {
	status := DistillerStatus{SchemaVersion: 1, Processor: "injected-distiller", AllowedSourceKeys: nonNil(slices.Clone(s.options.AllowedProcessingSourceKeys))}
	if err := authorize(p, false); err != nil {
		return status, err
	}
	if s.options.Distiller == nil {
		if s.options.ProjectDistillers != nil {
			status.Processor = "selected-project-cli"
			status.Available = true
			status.Reason = "Select a project and CLI with explicit model-processing permission; availability is checked for that selection"
			status.RequiredAction = "select_permitted_project_cli"
			return status, nil
		}
		status.Reason = "Automatic processor is not configured"
		status.RequiredAction = "configure_distiller"
		return status, nil
	}
	if provider, ok := s.options.Distiller.(DistillerStatusProvider); ok {
		actual, err := provider.Status(ctx)
		if err != nil {
			return status, mapError(err, "")
		}
		actual.SchemaVersion = 1
		actual.AllowedSourceKeys = nonNil(slices.Clone(s.options.AllowedProcessingSourceKeys))
		if len(actual.AllowedSourceKeys) == 0 {
			actual.Available = false
			actual.Reason = "No public source has been explicitly authorized"
			actual.RequiredAction = "configure_processing_source_keys"
		}
		return actual, nil
	}
	id, err := s.options.Distiller.ConfigurationID(ctx)
	if err != nil || strings.TrimSpace(id) == "" {
		status.Reason = "Automatic processor configuration or native isolation is unavailable"
		status.RequiredAction = "configure_distiller_isolation"
		return status, nil
	}
	status.ConfigurationID = &id
	status.Available = len(status.AllowedSourceKeys) > 0
	if !status.Available {
		status.Reason = "No public source has been explicitly authorized"
		status.RequiredAction = "configure_processing_source_keys"
	} else {
		status.Reason = "Processor configured; actual model provenance is recorded with each completed result"
	}
	// The port supplies configuration identity, not an independently verified
	// model name. Unknown model stays null instead of guessing from CLI defaults.
	return status, nil
}

func (s *Service) RequestDistillation(ctx context.Context, p Principal, c RequestDistillationCommand) (ImportJobResult, error) {
	c.InputRefs = normalizeRefs(c.InputRefs)
	c.PriorDistillationIDs = slices.Clone(c.PriorDistillationIDs)
	slices.Sort(c.PriorDistillationIDs)
	c.PriorDistillationIDs = slices.Compact(c.PriorDistillationIDs)
	c.Question = strings.TrimSpace(c.Question)
	prior, found, err := replay[ImportJobResult](s, ctx, p, c.CommandMeta, "RequestDistillation", c)
	if err != nil || found {
		return prior, err
	}
	if c.ExpectedVersion != 1 || len(c.InputRefs) == 0 || !slices.Contains([]string{"content", "topic", "project"}, c.Stage) || (strings.TrimSpace(c.ProcessingConfig) == "" && c.ProjectID == "") {
		return ImportJobResult{}, mapError(domain.ErrInvalid, c.RequestID)
	}
	// Configuration discovery performs no paid call and holds no SQL transaction.
	config := "unconfigured"
	var unavailable error
	processor, spaceID, resolveError := s.resolveDistiller(ctx, p, c)
	if resolveError != nil {
		return ImportJobResult{}, mapError(resolveError, c.RequestID)
	}
	if processor == nil {
		unavailable = serviceError(apierrors.ProviderUnavailable, "Automatic processor is not configured", "configure_distiller")
	} else {
		config, unavailable = processor.ConfigurationID(ctx)
		if unavailable == nil && strings.TrimSpace(config) == "" {
			unavailable = serviceError(apierrors.ProviderUnavailable, "Processor configuration is unavailable", "configure_distiller")
		}
		if unavailable != nil {
			config = "unavailable"
		}
	}
	effectiveBytes, _ := json.Marshal(struct{ CallerConfig, ProcessorConfig string }{c.ProcessingConfig, config})
	if c.ProjectID != "" {
		effectiveBytes, _ = json.Marshal(struct{ CallerConfig, ProcessorConfig, ProjectID, CLI, SpaceID string }{c.ProcessingConfig, config, c.ProjectID, c.CLI, spaceID})
	}
	effective := "automatic:" + string(effectiveBytes)
	identity, _ := json.Marshal(struct {
		Inputs                  []SourceRef
		Prior                   []string
		Stage, Config, Question string
	}{c.InputRefs, c.PriorDistillationIDs, c.Stage, effective, c.Question})
	reuseKey := "automatic:" + digestBytes(identity)
	result, err := command(s, ctx, p, c.CommandMeta, "RequestDistillation", c, func(tx AttentionTx) (ImportJobResult, error) {
		snapshots, _, err := s.automaticInputs(tx, c, spaceID)
		if err != nil {
			return ImportJobResult{}, err
		}
		old, err := tx.FindJobByDedupeKey(reuseKey)
		if err == nil {
			return ImportJobResult{SchemaVersion: 1, JobID: old.JobID, Status: old.Status}, nil
		}
		if !isMissing(err) {
			return ImportJobResult{}, err
		}
		payload, err := json.Marshal(automaticPayload{Command: c, Snapshots: snapshots, ProcessorConfig: config, EffectiveConfig: effective, ReuseKey: reuseKey, SpaceID: spaceID})
		if err != nil {
			return ImportJobResult{}, err
		}
		now := s.options.Clock()
		job := Job{SchemaVersion: 1, JobID: rand.Text(), Status: "queued", Kind: "distillation", Version: 1, DedupeKey: reuseKey, OperationID: rand.Text(), MaxAttempts: s.options.MaxAttempts, CreatedAt: now, UpdatedAt: now, DeadlineAt: now.Add(s.options.JobTimeout), Payload: payload, Caller: p}
		if len(c.InputRefs) == 1 {
			// An unambiguous source pointer supports navigation even while the
			// job is pending/failed. Multiple inputs remain on the fixed record.
			id, revision := c.InputRefs[0].MaterialID, c.InputRefs[0].Revision
			job.MaterialID, job.MaterialRevision = &id, &revision
		}
		if unavailable != nil {
			job.Status = "failed"
			job.Error = mapError(unavailable, c.RequestID).(*apierrors.ServiceError)
		}
		if err = tx.SaveJob(job, 0); err != nil {
			return ImportJobResult{}, err
		}
		if err = s.event(tx, "job_"+job.Status, job.JobID, 1, c.CommandMeta, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID, "input_refs": c.InputRefs, "prior_distillation_ids": c.PriorDistillationIDs}); err != nil {
			return ImportJobResult{}, err
		}
		return ImportJobResult{SchemaVersion: 1, JobID: job.JobID, Status: job.Status}, nil
	})
	if err == nil {
		s.signal()
	}
	return result, err
}

// Source keys are explicit composition settings for the two public examples;
// neither domain membership, a URL nor an Agent bearer grants this permission.
func (s *Service) automaticInputs(tx AttentionTx, c RequestDistillationCommand, spaceID string) ([]SourceSnapshot, []Distillation, error) {
	if err := validateSourceRefs(tx, c.InputRefs); err != nil {
		return nil, nil, err
	}
	snapshots := make([]SourceSnapshot, 0, len(c.InputRefs))
	for _, ref := range c.InputRefs {
		row, err := tx.LoadMaterial(ref.MaterialID)
		if err != nil {
			return nil, nil, err
		}
		var version MaterialRevision
		for _, v := range row.Revisions {
			if v.Revision == ref.Revision {
				version = v
				break
			}
		}
		if c.ProjectID != "" {
			if spaceID == "" {
				return nil, nil, serviceError(apierrors.ScopeDenied, "Selected project has no Attention space", "include_fixed_project_materials")
			}
			if err := projectReferenceAllowed(tx, spaceID, ref, false); err != nil {
				return nil, nil, err
			}
		} else if !slices.Contains(s.options.AllowedProcessingSourceKeys, version.SourceKey) || !slices.Contains([]string{"arxiv", "summarize", "arxiv+summarize"}, version.Provenance.Processor) {
			return nil, nil, serviceError(apierrors.ScopeDenied, "Source version was not authorized for public automatic processing", "choose_authorized_public_source")
		}
		snapshots = append(snapshots, SourceSnapshot{Ref: ref, SourceKey: version.SourceKey, Title: row.Material.Title, Summary: version.Summary, ContentDigest: version.ContentDigest, Provenance: version.Provenance})
	}
	prior := []Distillation{}
	if len(c.PriorDistillationIDs) == 0 {
		return snapshots, prior, nil
	}
	all, err := tx.ListDistillations()
	if err != nil {
		return nil, nil, err
	}
	for _, id := range c.PriorDistillationIDs {
		found := false
		for _, d := range all {
			if d.ID != id {
				continue
			}
			found = true
			refs := append(slices.Clone(d.InputRefs), d.RelatedRefs...)
			if d.CandidateSuggestion != nil {
				refs = append(refs, d.CandidateSuggestion.EvidenceRefs...)
			}
			if d.Status != "succeeded" || !refsWithin(refs, c.InputRefs) {
				return nil, nil, serviceError(apierrors.ScopeDenied, "Prior round contains unselected source versions", "select_all_prior_round_sources")
			}
			if err = validateSourceRefs(tx, refs); err != nil {
				return nil, nil, err
			}
			prior = append(prior, d)
			break
		}
		if !found {
			return nil, nil, apierrors.NewNotFound("distillation", id)
		}
	}
	return snapshots, prior, nil
}

func refsWithin(refs, selected []SourceRef) bool {
	for _, ref := range refs {
		found := false
		for _, input := range selected {
			if ref.MaterialID == input.MaterialID && ref.Revision == input.Revision && ref.Locator == input.Locator && (input.Span == nil || (ref.Span != nil && *ref.Span == *input.Span)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (s *Service) processAutomatic(ctx context.Context, claim Job) error {
	var payload automaticPayload
	if err := json.Unmarshal(claim.Payload, &payload); err != nil {
		return s.failJob(claim, domain.ErrInvalid, false)
	}
	if s.objects == nil {
		return s.failJob(claim, serviceError(apierrors.ProviderUnavailable, "Automatic processor or object store is not configured", "configure_distiller"), false)
	}
	workCtx, cancel := context.WithTimeout(ctx, claim.DeadlineAt.Sub(s.options.Clock()))
	work := &activeWork{cancel: cancel}
	s.activeMu.Lock()
	s.activeJobs[claim.JobID] = work
	s.activeMu.Unlock()
	defer func() {
		cancel()
		s.activeMu.Lock()
		if s.activeJobs[claim.JobID] == work {
			delete(s.activeJobs, claim.JobID)
		}
		s.activeMu.Unlock()
	}()
	if payload.Result != nil {
		if payload.Command.ProjectID != "" {
			_, spaceID, err := s.resolveDistiller(workCtx, claim.Caller, payload.Command)
			if err != nil {
				return s.failJob(claim, err, false)
			}
			if spaceID != payload.SpaceID {
				return s.failJob(claim, serviceError(apierrors.ScopeDenied, "Selected project's Attention space changed", "request_current_project_distillation"), false)
			}
		}
		// Explicit retry after a local publication/storage failure reuses the
		// actual delivered result, never resending the paid processor call.
		return s.completeAutomatic(workCtx, claim, payload, *payload.Result)
	}
	processor, spaceID, err := s.resolveDistiller(workCtx, claim.Caller, payload.Command)
	if err != nil {
		return s.failJob(claim, err, false)
	}
	if processor == nil {
		return s.failJob(claim, serviceError(apierrors.ProviderUnavailable, "Selected processor is unavailable", "configure_project_cli"), false)
	}
	if spaceID != payload.SpaceID {
		return s.failJob(claim, serviceError(apierrors.ScopeDenied, "Selected project's Attention space changed", "request_current_project_distillation"), false)
	}
	currentConfig, err := processor.ConfigurationID(workCtx)
	if err != nil {
		return s.failJob(claim, err, false)
	}
	if currentConfig != payload.ProcessorConfig {
		return s.failJob(claim, serviceError(apierrors.ContextStale, "Processor configuration changed after the fixed request", "request_new_distillation"), false)
	}
	var snapshots []SourceSnapshot
	var prior []Distillation
	refs := []string{}
	err = s.repo.WithTx(workCtx, func(tx AttentionTx) error {
		var err error
		snapshots, prior, err = s.automaticInputs(tx, payload.Command, payload.SpaceID)
		if err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			row, err := tx.LoadMaterial(snapshot.Ref.MaterialID)
			if err != nil {
				return err
			}
			for _, r := range row.Revisions {
				if r.Revision == snapshot.Ref.Revision {
					refs = append(refs, r.ObjectRef)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return s.failJob(claim, err, false)
	}
	if len(payload.Snapshots) == len(snapshots) {
		// Mutable material titles must not introduce later metadata into a
		// queued round. Immutable revisions and the original selection remain
		// the authority; source access was rechecked immediately above.
		for i := range snapshots {
			snapshots[i].Title = payload.Snapshots[i].Title
		}
	}
	for i, ref := range refs {
		data, err := s.objects.Read(workCtx, ref)
		if err != nil {
			return s.failJob(claim, err, false)
		}
		// The object port verifies the immutable object on read. The processor
		// digest describes precisely the transmitted UTF-8 text, independently
		// of any revision-level bundle identity (PDF/HTML/export attachments).
		snapshots[i].ContentDigest = digestBytes(data)
		snapshots[i].Text = string(data)
	}
	// Paid/native processor entry is fenced by persisted CAS. No external I/O
	// occurs in this transaction; a lost worker after this point stays unknown.
	err = s.repo.WithTx(workCtx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" || job.CancelRequested {
			return domain.ErrVersion
		}
		if _, _, err = s.automaticInputs(tx, payload.Command, payload.SpaceID); err != nil {
			return err
		}
		used := map[string]bool{}
		for _, ref := range payload.Command.InputRefs {
			if used[ref.MaterialID] {
				continue
			}
			used[ref.MaterialID] = true
			row, err := tx.LoadMaterial(ref.MaterialID)
			if err != nil {
				return err
			}
			if err = s.machineRead(tx, &row, Principal{ID: "processor:" + claim.OperationID, Kind: "agent"}); err != nil {
				return err
			}
		}
		job.ExternalStarted = true
		job.Version++
		job.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(job, claim.Version); err != nil {
			return err
		}
		claim = job
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return nil
		}
		return s.failJob(claim, err, false)
	}
	output, err := processor.Distill(workCtx, DistillationInput{JobID: claim.JobID, OperationID: claim.OperationID, Inputs: snapshots, PriorDistillations: prior, Stage: payload.Command.Stage, ProcessingConfig: payload.EffectiveConfig, Question: payload.Command.Question})
	if err != nil {
		return s.failJob(claim, err, automaticOutcomeUnknown(err))
	}
	// Persist the delivered result before local publication, so an explicit
	// retry after a storage failure cannot resend the paid call. A failure in
	// this narrow persistence gap stays conservatively unknown.
	if err = validateAutomaticOutput(payload.Command, output); err != nil {
		return s.failJob(claim, err, false)
	}
	if candidate := output.CandidateSuggestion; candidate != nil {
		candidate.EvidenceRefs = nonNil(normalizeRefs(candidate.EvidenceRefs))
		candidate.GoalRefs = nonNil(candidate.GoalRefs)
		candidate.MissingEvidence = nonNil(candidate.MissingEvidence)
	}
	payload.Result = &output
	resultCtx, resultCancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = s.repo.WithTx(resultCtx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Version != claim.Version || job.Status != "running" || job.CancelRequested {
			return domain.ErrVersion
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		job.Payload = data
		job.Version++
		job.UpdatedAt = s.options.Clock()
		if err = tx.SaveJob(job, claim.Version); err != nil {
			return err
		}
		claim = job
		return nil
	})
	resultCancel()
	if err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return nil
		}
		return s.failJob(claim, err, true)
	}
	return s.completeAutomatic(workCtx, claim, payload, output)
}

func (s *Service) completeAutomatic(ctx context.Context, claim Job, payload automaticPayload, output DistillationOutput) error {
	if err := validateAutomaticOutput(payload.Command, output); err != nil {
		return s.failJob(claim, err, false)
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		_, _, err := s.automaticInputs(tx, payload.Command, payload.SpaceID)
		return err
	})
	if err != nil {
		return s.failJob(claim, err, false)
	}
	outputRef, err := s.objects.Publish(ctx, []byte(output.OutputText))
	if err != nil {
		return s.failJob(claim, err, false)
	}
	if err = s.finishAutomatic(ctx, claim, payload, output, outputRef); err != nil {
		if errors.Is(err, domain.ErrVersion) {
			return nil
		}
		return s.failJob(claim, err, false)
	}
	return nil
}

func automaticOutcomeUnknown(err error) bool {
	if externalOutcomeUnknown(err) {
		return true
	}
	var e *apierrors.ServiceError
	// The trusted adapter can report a known start failure or a fully delivered
	// malformed result. Unclassified errors/cancellation after entry are unknown.
	if errors.As(err, &e) {
		return !slices.Contains([]apierrors.Code{apierrors.ProviderUnavailable, apierrors.ValidationFailed, apierrors.EvidenceMissing, apierrors.ScopeDenied, apierrors.UnsupportedCapability, apierrors.ContextStale}, e.Code)
	}
	return true
}

func validateAutomaticOutput(c RequestDistillationCommand, o DistillationOutput) error {
	if strings.TrimSpace(o.OutputText) == "" || o.Provenance.Processor == "" || o.Provenance.Version == "" || o.Provenance.Mode == "" || o.Provenance.Mode == "manual" || o.Provenance.Processor == "human" {
		return domain.ErrEvidence
	}
	if !refsWithin(o.RelatedRefs, c.InputRefs) {
		return domain.ErrEvidence
	}
	out := domain.DistillationOutput{SourcePreserved: len(c.InputRefs) > 0, Summary: o.OutputText, Related: domainRefs(o.RelatedRefs), ExistingView: strings.Join(o.RelatedIdeas, "\n"), Conflict: strings.Join(o.Conflicts, "\n"), OpenQuestions: o.PendingQuestions, Goal: strings.Join(o.GoalRefs, "\n"), ExistingAsset: strings.Join(o.ExistingAssets, "\n"), ExpectedImprovement: o.ExpectedImprovement, MinimumOutcome: o.MinimumArtifact, MissingEvidence: o.MissingEvidence}
	if err := domain.ValidateDistillation(c.Stage, domainRefs(c.InputRefs), out); err != nil {
		return err
	}
	if suggestion := o.CandidateSuggestion; suggestion != nil {
		if strings.TrimSpace(suggestion.Title) == "" || !refsWithin(suggestion.EvidenceRefs, c.InputRefs) {
			return domain.ErrEvidence
		}
		if err := domain.ValidateDimensions(dimensions(suggestion.Dimensions)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) finishAutomatic(ctx context.Context, claim Job, p automaticPayload, o DistillationOutput, outputRef string) error {
	return s.repo.WithTx(ctx, func(tx AttentionTx) error {
		job, err := tx.LoadJob(claim.JobID)
		if err != nil {
			return err
		}
		if job.Status != "running" || job.Version != claim.Version || job.CancelRequested {
			return domain.ErrVersion
		}
		if _, _, err = s.automaticInputs(tx, p.Command, p.SpaceID); err != nil {
			return err
		}
		if err = validateSourceRefs(tx, o.RelatedRefs); err != nil {
			return err
		}
		if o.CandidateSuggestion != nil {
			if err = validateSourceRefs(tx, o.CandidateSuggestion.EvidenceRefs); err != nil {
				return err
			}
		}
		d, err := tx.FindDistillationByReuseKey(p.ReuseKey)
		if err != nil && !isMissing(err) {
			return err
		}
		if isMissing(err) {
			d = Distillation{ID: rand.Text(), PriorDistillationIDs: nonNil(p.Command.PriorDistillationIDs), InputRefs: nonNil(p.Command.InputRefs), Stage: p.Command.Stage, OutputRef: outputRef, OutputText: o.OutputText, Status: "succeeded", NextQuestion: o.NextQuestion, Question: p.Command.Question, ProcessingConfig: p.EffectiveConfig, RelatedRefs: nonNil(o.RelatedRefs), RelatedIdeas: nonNil(o.RelatedIdeas), Conflicts: nonNil(o.Conflicts), PendingQuestions: nonNil(o.PendingQuestions), GoalRefs: nonNil(o.GoalRefs), ExistingAssets: nonNil(o.ExistingAssets), ExpectedImprovement: o.ExpectedImprovement, MinimumArtifact: o.MinimumArtifact, MissingEvidence: nonNil(o.MissingEvidence), CandidateSuggestion: o.CandidateSuggestion, Provenance: o.Provenance, ReuseKey: p.ReuseKey, CreatedAt: s.options.Clock()}
			if err = tx.SaveDistillation(d); err != nil {
				return err
			}
			if err = s.event(tx, "distillation_recorded", d.ID, 1, CommandMeta{}, map[string]any{"distillation_id": d.ID, "job_id": job.JobID, "input_refs": d.InputRefs, "prior_distillation_ids": p.Command.PriorDistillationIDs}); err != nil {
				return err
			}
		}
		old := job.Version
		job.Status = "succeeded"
		job.Version++
		job.UpdatedAt = s.options.Clock()
		job.Error = nil
		job.DeliveryUnknown = false
		job.DistillationID = &d.ID
		if err = tx.SaveJob(job, old); err != nil {
			return err
		}
		return s.event(tx, "job_succeeded", job.JobID, job.Version, CommandMeta{}, map[string]any{"job_id": job.JobID, "operation_id": job.OperationID, "distillation_id": d.ID})
	})
}
