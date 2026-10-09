package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type testDistiller struct {
	config      string
	configError error
	calls       int
	call        func(context.Context, DistillationInput) (DistillationOutput, error)
}

func (d *testDistiller) ConfigurationID(context.Context) (string, error) {
	return d.config, d.configError
}
func (d *testDistiller) Distill(ctx context.Context, i DistillationInput) (DistillationOutput, error) {
	d.calls++
	return d.call(ctx, i)
}
func automaticFixture(t *testing.T) (*Service, *memoryRepo, *memoryObjects, MaterialDetail, *testDistiller) {
	s, r, o := fixture(t)
	m := importFixture(t, s, "auto-source", "Actual fixed source A")
	row := r.state.Materials[m.Material.ID]
	row.Revisions[0].SourceKey = "arxiv:public"
	row.Revisions[0].Provenance = Provenance{Processor: "arxiv", Version: "v1", Mode: "official_full_text", Source: row.Material.SourceLocator}
	r.state.Materials[m.Material.ID] = row
	d := &testDistiller{config: "actual-model:test;isolation:text-only"}
	d.call = func(_ context.Context, i DistillationInput) (DistillationOutput, error) {
		if !r.mu.TryLock() {
			t.Fatal("model I/O under transaction")
		}
		r.mu.Unlock()
		return DistillationOutput{OutputText: "Actual fake model output (contract test only)", Provenance: Provenance{Processor: "test-distiller", Version: "1", Mode: "selected-text-only", Model: "test-model"}}, nil
	}
	s.options.Distiller = d
	s.options.AllowedProcessingSourceKeys = []string{"arxiv:public"}
	return s, r, o, m, d
}
func autoCommand(m MaterialDetail, key string) RequestDistillationCommand {
	return RequestDistillationCommand{CommandMeta: meta(key, 1), InputRefs: sourceRefs(m), Stage: "content", ProcessingConfig: "user-selected-process-v1", Question: "Actual selected question"}
}
func runAutomatic(t *testing.T, s *Service, c RequestDistillationCommand) Job {
	t.Helper()
	ctx := context.Background()
	queued, err := s.RequestDistillation(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ProcessNextJob(ctx)
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.GetJob(ctx, human, queued.JobID)
	if err != nil {
		t.Fatal(err)
	}
	return job
}
func TestAutomaticFixedSnapshotsReuseAndContinuation(t *testing.T) {
	s, r, o, m, d := automaticFixture(t)
	ctx := context.Background()
	c := autoCommand(m, "auto-first")
	queued, err := s.RequestDistillation(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	// Head changes after queueing do not change the explicitly selected snapshot.
	row := r.state.Materials[m.Material.ID]
	ref, _ := o.Publish(ctx, []byte("Actual fixed source B"))
	row.Revisions = append(row.Revisions, MaterialRevision{MaterialID: m.Material.ID, Revision: 2, SourceKey: "arxiv:public", SourceLocator: row.Material.SourceLocator, ObjectRef: ref, ContentDigest: digestBytes([]byte("Actual fixed source B")), Provenance: row.Revisions[0].Provenance})
	row.Material.CurrentRevision = 2
	row.Material.Version++
	r.state.Materials[m.Material.ID] = row
	original := d.call
	d.call = func(ctx context.Context, i DistillationInput) (DistillationOutput, error) {
		if i.Inputs[0].Text != "Actual fixed source A" || i.Inputs[0].Ref.Revision != 1 || i.OperationID == "" {
			t.Fatalf("not frozen: %+v", i)
		}
		return original(ctx, i)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(ctx, human, queued.JobID)
	if job.Status != "succeeded" || job.DistillationID == nil || d.calls != 1 {
		t.Fatalf("job: %+v", job)
	}
	saved := r.state.Distillations[*job.DistillationID]
	if saved.Provenance.Model != "test-model" || !strings.Contains(saved.ProcessingConfig, d.config) {
		t.Fatal("lost actual config/provenance")
	}
	c.CommandMeta = meta("auto-newkey-samework", 1)
	repeat, err := s.RequestDistillation(ctx, human, c)
	if err != nil || repeat.JobID != job.JobID {
		t.Fatalf("reuse: %+v %v", repeat, err)
	}
	s.ProcessNextJob(ctx)
	if d.calls != 1 {
		t.Fatal("repeated model work")
	}
	c.CommandMeta = meta("auto-continue", 1)
	c.PriorDistillationIDs = []string{saved.ID}
	c.Question = "New concrete unresolved question"
	d.call = func(ctx context.Context, i DistillationInput) (DistillationOutput, error) {
		if len(i.PriorDistillations) != 1 || i.PriorDistillations[0].ID != saved.ID || i.Question != c.Question {
			t.Fatal("prior round not selected")
		}
		return original(ctx, i)
	}
	next := runAutomatic(t, s, c)
	if next.Status != "succeeded" || next.JobID == job.JobID || d.calls != 2 {
		t.Fatal("new round reused prior work")
	}
	material, _ := s.GetMaterial(ctx, human, m.Material.ID)
	if material.Material.AttentionScore != 0 || material.Material.AgentUsageCount != 2 || material.Material.LongTermValue != 0 {
		t.Fatalf("processing heat: %+v", material.Material)
	}
	if len(r.state.Opportunities) != 0 {
		t.Fatal("no suggestion invented a candidate")
	}
}
func TestAutomaticRealSuggestionAndLocalFailureRecovery(t *testing.T) {
	s, r, _, m, d := automaticFixture(t)
	d.call = func(_ context.Context, i DistillationInput) (DistillationOutput, error) {
		return DistillationOutput{OutputText: "Actual suggestion", PendingQuestions: []string{"Missing related source"}, CandidateSuggestion: &CandidateSuggestion{Title: "Model suggested candidate", EvidenceRefs: i.InputsRefsForTest(), Purpose: "Improve selected material", NextStep: "Check baseline"}, Provenance: Provenance{Processor: "test-distiller", Version: "1", Mode: "model", Model: "test"}}, nil
	}
	c := autoCommand(m, "suggestion")
	c.Stage = "topic"
	job := runAutomatic(t, s, c)
	if job.Status != "succeeded" || len(r.state.Opportunities) != 0 || r.state.Distillations[*job.DistillationID].CandidateSuggestion == nil {
		t.Fatalf("suggestion job %+v", job)
	}
	// Outbox failure must roll back both records and completion, keeping inputs.
	s, r, _, m, d = automaticFixture(t)
	c = autoCommand(m, "atomic")
	result, err := s.RequestDistillation(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	original := d.call
	d.call = func(ctx context.Context, i DistillationInput) (DistillationOutput, error) {
		out, err := original(ctx, i)
		r.failEvent = true
		return out, err
	}
	_, err = s.ProcessNextJob(context.Background())
	if err == nil {
		t.Fatal("outbox failure hidden")
	}
	r.failEvent = false
	if len(r.state.Distillations) != 0 || len(r.state.Materials) != 1 || r.state.Jobs[result.JobID].Status != "running" {
		t.Fatal("partial completion escaped tx")
	}
	if err = s.RecoverJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	job = r.state.Jobs[result.JobID]
	if job.DeliveryUnknown {
		t.Fatal("durable delivered result treated as lost")
	}
	_, err = s.RetryJob(context.Background(), human, job.JobID, meta("retry-local-publish", job.Version))
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(context.Background())
	job = r.state.Jobs[job.JobID]
	if job.Status != "succeeded" || d.calls != 1 || len(r.state.Distillations) != 1 {
		t.Fatalf("retry resent paid work: %+v calls=%d", job, d.calls)
	}
}

// Test-only projection, not an application source fetch.
func (i DistillationInput) InputsRefsForTest() []SourceRef {
	refs := []SourceRef{}
	for _, input := range i.Inputs {
		refs = append(refs, input.Ref)
	}
	return refs
}
func TestAutomaticScopeConfigAndHonestFailures(t *testing.T) {
	ctx := context.Background()
	s, r, _, m, d := automaticFixture(t)
	c := autoCommand(m, "deny-agent")
	_, err := s.RequestDistillation(ctx, agent, c)
	errorCode(t, err, apierrors.ScopeDenied)
	s.options.AllowedProcessingSourceKeys = nil
	c.CommandMeta = meta("deny-scope", 1)
	_, err = s.RequestDistillation(ctx, human, c)
	errorCode(t, err, apierrors.ScopeDenied)
	if d.calls != 0 {
		t.Fatal("denied call")
	}
	s.options.AllowedProcessingSourceKeys = []string{"arxiv:public"}
	s.options.Distiller = nil
	c.CommandMeta = meta("notconfigured", 1)
	job := runAutomatic(t, s, c)
	if job.Status != "failed" || job.Error == nil || job.ExternalStarted || len(r.state.Distillations) != 0 {
		t.Fatalf("unconfigured claimed summary: %+v", job)
	}
	s.options.Distiller = d
	c.CommandMeta = meta("config-fixed", 1)
	queued, err := s.RequestDistillation(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	d.config = "changed model configuration"
	s.ProcessNextJob(ctx)
	job = r.state.Jobs[queued.JobID]
	if job.Status != "failed" || job.Error.Code != apierrors.ContextStale || job.ExternalStarted || d.calls != 0 {
		t.Fatalf("changed processor silently used: %+v", job)
	}
	s, r, _, m, d = automaticFixture(t)
	c = autoCommand(m, "invented-ref")
	d.call = func(_ context.Context, _ DistillationInput) (DistillationOutput, error) {
		return DistillationOutput{OutputText: "invented", RelatedRefs: []SourceRef{{MaterialID: "private", Revision: 1, Locator: "private"}}, Provenance: Provenance{Processor: "test", Version: "1", Mode: "model"}}, nil
	}
	job = runAutomatic(t, s, c)
	if job.Status != "failed" || job.DeliveryUnknown || len(r.state.Distillations) != 0 {
		t.Fatalf("hallucinated refs: %+v", job)
	}
}
func TestAutomaticUnknownRecoveryAndExplicitBoundedRetry(t *testing.T) {
	s, r, _, m, d := automaticFixture(t)
	ctx := context.Background()
	c := autoCommand(m, "unknown")
	d.call = func(context.Context, DistillationInput) (DistillationOutput, error) {
		return DistillationOutput{}, errors.New("connection lost after processor entry")
	}
	job := runAutomatic(t, s, c)
	if job.Status != "failed" || !job.DeliveryUnknown || job.Error.Code != apierrors.DeliveryUnknown {
		t.Fatalf("unknown lost: %+v", job)
	}
	_, err := s.RetryJob(ctx, human, job.JobID, meta("retry-unknown", job.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
	if err = s.RecoverJobs(ctx); err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	if d.calls != 1 || len(r.state.Materials) != 1 {
		t.Fatal("unknown retried or source lost")
	}
	s, r, _, m, d = automaticFixture(t)
	c = autoCommand(m, "known-failure")
	s.options.MaxAttempts = 2
	d.call = func(context.Context, DistillationInput) (DistillationOutput, error) {
		return DistillationOutput{}, serviceError(apierrors.ProviderUnavailable, "native process did not start", "retry_job")
	}
	job = runAutomatic(t, s, c)
	if job.DeliveryUnknown {
		t.Fatal("known start failure unknown")
	}
	retried, err := s.RetryJob(ctx, human, job.JobID, meta("explicit-retry", job.Version))
	if err != nil || retried.OperationID != job.OperationID || !retried.DeadlineAt.Equal(job.DeadlineAt) {
		t.Fatalf("retry %+v %v", retried, err)
	}
	s.ProcessNextJob(ctx)
	job, _ = s.GetJob(ctx, human, job.JobID)
	_, err = s.RetryJob(ctx, human, job.JobID, meta("retry-limit", job.Version))
	errorCode(t, err, apierrors.ValidationFailed)
}
