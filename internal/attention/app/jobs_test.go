package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestReadOnlyFailureExplicitRetryCeilingAndStableOperation(t *testing.T) {
	s, _, _ := fixture(t)
	s.options.MaxAttempts = 2
	s.sources = sourceFunc(func(context.Context, ImportMaterialCommand) (ImportedSource, error) {
		return ImportedSource{}, errors.New("a safe read failed")
	})
	response, err := s.ImportMaterial(context.Background(), human, importCommand("failure", "bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(context.Background(), human, response.JobID)
	if job.Status != "failed" || job.DeliveryUnknown || job.Attempts != 1 {
		t.Fatalf("read failure %+v", job)
	}
	op := job.OperationID
	deadline := job.DeadlineAt
	job, err = s.RetryJob(context.Background(), human, job.JobID, meta("retry", job.Version))
	if err != nil || job.Status != "queued" || job.OperationID != op || !job.DeadlineAt.Equal(deadline) {
		t.Fatal("retry reset durable operation/time budget")
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	job, _ = s.GetJob(context.Background(), human, job.JobID)
	_, err = s.RetryJob(context.Background(), human, job.JobID, meta("exhausted", job.Version))
	errorCode(t, err, apierrors.ValidationFailed)
	if job.Attempts != 2 {
		t.Fatal("attempt count lost")
	}
}

func TestRecoveryReadOnlyVsUnknownModelAndDeadline(t *testing.T) {
	s, r, _ := fixture(t)
	now := s.options.Clock()
	for _, kind := range []string{"import", "distillation"} {
		r.state.Jobs[kind] = Job{SchemaVersion: 1, JobID: kind, Status: "running", Kind: kind, Version: 2, Attempts: 1, MaxAttempts: 3, DeadlineAt: now.Add(time.Hour), ExternalStarted: true, OperationID: "original-" + kind, Payload: []byte(`{"durable":true}`), Caller: human}
	}
	if err := s.RecoverJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	read, _ := s.GetJob(context.Background(), human, "import")
	model, _ := s.GetJob(context.Background(), human, "distillation")
	if read.Status != "failed" || read.DeliveryUnknown || model.Status != "failed" || !model.DeliveryUnknown || model.Error.Code != apierrors.DeliveryUnknown {
		t.Fatal("recovery collapsed different external effects")
	}
	if string(model.Payload) != `{"durable":true}` || model.Caller.ID != human.ID {
		t.Fatal("internal payload or caller lost on persistence")
	}
	_, err := s.RetryJob(context.Background(), human, model.JobID, meta("blind-resend", model.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
	if r.state.Jobs[model.JobID].OperationID != "original-distillation" {
		t.Fatal("unknown operation replaced")
	}
	s.options.Clock = func() time.Time { return now.Add(time.Hour) }
	_, err = s.RetryJob(context.Background(), human, read.JobID, meta("expired", read.Version))
	errorCode(t, err, apierrors.ValidationFailed)
}

func TestRunningCancelStopsOwnedContextAndDoesNotCommitMaterial(t *testing.T) {
	s, r, _ := fixture(t)
	started := make(chan struct{})
	stopped := make(chan struct{})
	s.sources = sourceFunc(func(ctx context.Context, _ ImportMaterialCommand) (ImportedSource, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		return ImportedSource{}, ctx.Err()
	})
	response, err := s.ImportMaterial(context.Background(), human, importCommand("cancel", "original input"))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := s.ProcessNextJob(context.Background()); done <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not start")
	}
	job, _ := s.GetJob(context.Background(), human, response.JobID)
	cancelled, err := s.CancelJob(context.Background(), human, job.JobID, meta("cancel-operation", job.Version))
	if err != nil || cancelled.Status != "cancelled" || cancelled.DeliveryUnknown {
		t.Fatal("readonly cancel failed")
	}
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not propagate to adapter")
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(r.state.Materials) != 0 || len(r.state.Jobs) != 1 {
		t.Fatal("cancel discarded job or committed a material")
	}
}

func TestCancellationReceiptCannotCancelLaterRetry(t *testing.T) {
	s, _, _ := fixture(t)
	response, err := s.ImportMaterial(context.Background(), human, importCommand("import", "original"))
	if err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(context.Background(), human, response.JobID)
	m := meta("cancel", job.Version)
	job, err = s.CancelJob(context.Background(), human, job.JobID, m)
	if err != nil {
		t.Fatal(err)
	}
	job, err = s.RetryJob(context.Background(), human, job.JobID, meta("retry", job.Version))
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	base := s.sources
	s.sources = sourceFunc(func(ctx context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		close(started)
		select {
		case <-ctx.Done():
			return ImportedSource{}, ctx.Err()
		case <-release:
			return base.ReadSource(ctx, c)
		}
	})
	done := make(chan error, 1)
	go func() { _, err := s.ProcessNextJob(context.Background()); done <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker not started")
	}
	if _, err = s.CancelJob(context.Background(), human, job.JobID, m); err != nil {
		t.Fatal("old cancel receipt did not replay")
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	job, _ = s.GetJob(context.Background(), human, job.JobID)
	if job.Status != "succeeded" {
		t.Fatalf("replayed old cancellation interrupted new work: %+v", job)
	}
}

func TestWorkerPoolFiniteConcurrencyAndGracefulStop(t *testing.T) {
	s, _, _ := fixture(t)
	s.options.WorkerConcurrency = 2
	var active, maxActive, calls atomic.Int32
	started := make(chan struct{}, 8)
	release := make(chan struct{})
	base := s.sources
	s.sources = sourceFunc(func(ctx context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		n := active.Add(1)
		defer active.Add(-1)
		calls.Add(1)
		for old := maxActive.Load(); n > old && !maxActive.CompareAndSwap(old, n); old = maxActive.Load() {
		}
		started <- struct{}{}
		select {
		case <-ctx.Done():
			return ImportedSource{}, ctx.Err()
		case <-release:
			return base.ReadSource(ctx, c)
		}
	})
	for i := 0; i < 8; i++ {
		c := importCommand(fmt.Sprint(i), fmt.Sprint(i))
		c.SourceKey = fmt.Sprint(i)
		if _, err := s.ImportMaterial(context.Background(), human, c); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			cancel()
			t.Fatal("pool did not start")
		}
	}
	if calls.Load() != 2 || maxActive.Load() != 2 {
		cancel()
		t.Fatal("worker concurrency exceeded bound")
	}
	close(release)
	// Waiting on durable statuses rather than sleeping checks that every claim
	// finishes. This is a bounded test wait, not application retry machinery.
	deadline := time.Now().Add(3 * time.Second)
	for {
		jobs, err := s.ListJobs(context.Background(), human)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		finished := 0
		for _, j := range jobs.Items {
			if j.(Job).Status == "succeeded" {
				finished++
			}
		}
		if finished == 8 {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("pool failed to drain finite queue")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if maxActive.Load() > 2 || calls.Load() != 8 {
		t.Fatal("duplicate claim or unbounded execution")
	}
}

func TestCanonicalContractKindsAndMutableSourceRefresh(t *testing.T) {
	s, _, _ := fixture(t)
	for _, kind := range []string{"paper", "video", "text", "file"} {
		c := importCommand("kind-"+kind, "bytes-"+kind)
		c.Kind = kind
		if _, err := s.ImportMaterial(context.Background(), human, c); err != nil {
			t.Fatalf("legal contract kind %s rejected: %v", kind, err)
		}
	}
	for _, kind := range []string{"note", "article"} {
		c := importCommand("bad-"+kind, "bytes")
		c.Kind = kind
		_, err := s.ImportMaterial(context.Background(), human, c)
		errorCode(t, err, apierrors.ValidationFailed)
	}
	s2, _, _ := fixture(t)
	var mu sync.Mutex
	text := "first remote content"
	s2.sources = sourceFunc(func(context.Context, ImportMaterialCommand) (ImportedSource, error) {
		mu.Lock()
		defer mu.Unlock()
		return ImportedSource{SourceKey: "canonical-source", SourceLocator: "https://example.test/source", Kind: "text", Text: text}, nil
	})
	c := importCommand("remote-first", "")
	response, err := s2.ImportMaterial(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s2.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	first, _ := s2.GetJob(context.Background(), human, response.JobID)
	mu.Lock()
	text = "changed remote content"
	mu.Unlock()
	c.CommandMeta = meta("remote-again", 1)
	response, err = s2.ImportMaterial(context.Background(), human, c)
	if err != nil || response.JobID == first.JobID {
		t.Fatal("mutable remote was never reread")
	}
	if _, err = s2.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	row, _ := s2.GetMaterial(context.Background(), human, *first.MaterialID)
	if len(row.Revisions) != 2 {
		t.Fatal("remote update did not append immutable revision")
	}
}

func TestExplicitUnknownReadResultIsPreservedEvenForImport(t *testing.T) {
	s, _, _ := fixture(t)
	s.sources = sourceFunc(func(context.Context, ImportMaterialCommand) (ImportedSource, error) {
		return ImportedSource{}, serviceError(apierrors.DeliveryUnknown, "Unknown external adapter outcome", "reconcile")
	})
	response, err := s.ImportMaterial(context.Background(), human, importCommand("unknown", "source"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(context.Background(), human, response.JobID)
	if !job.DeliveryUnknown {
		t.Fatal("explicit unknown was replaced by ordinary read failure")
	}
	_, err = s.RetryJob(context.Background(), human, job.JobID, meta("retry", job.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
}
