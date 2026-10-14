package app_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestReadOnlyFailureExplicitRetryCeilingAndStableOperation(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	reader := importers.NewReader(nil)
	reader.Arxiv.Client = &http.Client{Timeout: time.Nanosecond}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{MaxAttempts: 2})
	c := ImportMaterialCommand{CommandMeta: meta("failure", 1), Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v1"}
	response, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("queue absent-source read")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process absent-source read")
	}
	job, err := s.GetJob(ctx, human, response.JobID)
	if err != nil || job.Status != "failed" || job.DeliveryUnknown || job.Attempts != 1 {
		t.Fatal("actual absent-source read failure collapsed delivery states")
	}
	op, deadline := job.OperationID, job.DeadlineAt
	job, err = s.RetryJob(ctx, human, job.JobID, meta("retry", job.Version))
	if err != nil || job.Status != "queued" || job.OperationID != op || !job.DeadlineAt.Equal(deadline) {
		t.Fatal("explicit retry reset durable operation or budget")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("process explicit absent-source retry")
	}
	job, err = s.GetJob(ctx, human, job.JobID)
	if err != nil {
		t.Fatal("load actual read failure")
	}
	_, err = s.RetryJob(ctx, human, job.JobID, meta("exhausted", job.Version))
	errorCode(t, err, apierrors.ValidationFailed)
	if job.Attempts != 2 {
		t.Fatal("bounded attempt count lost")
	}
	if job.Error == nil || job.Error.Retryable || job.Error.RequiredAction != "review_source_or_configuration_before_new_work" {
		t.Fatal("exhausted read still invited invisible retry")
	}
}

func TestRunningCancelStopsOwnedContextAndDoesNotCommitMaterial(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	started, stopped := make(chan struct{}), make(chan struct{})
	reader := &observedSource{Reader: importers.NewReader(nil), beforeReadContext: func(ctx context.Context) { close(started); <-ctx.Done(); close(stopped) }}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
	response, err := s.ImportMaterial(ctx, human, importCommand("cancel", "original manual input"))
	if err != nil {
		t.Fatal("queue manual import for cancellation")
	}
	done := make(chan error, 1)
	go func() { _, err := s.ProcessNextJob(ctx); done <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not enter production reader observation boundary")
	}
	job, err := s.GetJob(ctx, human, response.JobID)
	if err != nil {
		t.Fatal("load running read")
	}
	cancelled, err := s.CancelJob(ctx, human, job.JobID, meta("cancel-operation", job.Version))
	if err != nil || cancelled.Status != "cancelled" || cancelled.DeliveryUnknown {
		t.Fatal("readonly cancellation failed")
	}
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("owned context cancellation did not reach reader")
	}
	if err = <-done; err != nil {
		t.Fatal("cancelled read worker failed")
	}
	if countRows(t, db, "attention_materials") != 0 || countRows(t, db, "attention_jobs") != 1 {
		t.Fatal("cancellation discarded job or committed material")
	}
}

func TestCancellationReceiptCannotCancelLaterRetry(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	reader := &observedSource{Reader: importers.NewReader(nil), beforeReadContext: func(ctx context.Context) {
		close(started)
		select {
		case <-ctx.Done():
		case <-release:
		}
	}}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{})
	response, err := s.ImportMaterial(ctx, human, importCommand("import", "original manual input"))
	if err != nil {
		t.Fatal("queue manual import")
	}
	job, err := s.GetJob(ctx, human, response.JobID)
	if err != nil {
		t.Fatal("load queued read")
	}
	m := meta("cancel", job.Version)
	job, err = s.CancelJob(ctx, human, job.JobID, m)
	if err != nil {
		t.Fatal("cancel queued read")
	}
	job, err = s.RetryJob(ctx, human, job.JobID, meta("retry", job.Version))
	if err != nil {
		t.Fatal("explicitly retry known cancellation")
	}
	done := make(chan error, 1)
	go func() { _, err := s.ProcessNextJob(ctx); done <- err }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("retry worker did not enter reader boundary")
	}
	if _, err = s.CancelJob(ctx, human, job.JobID, m); err != nil {
		t.Fatal("old cancellation receipt did not replay")
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal("production read retry failed")
	}
	job, err = s.GetJob(ctx, human, job.JobID)
	if err != nil || job.Status != "succeeded" {
		t.Fatal("replayed cancellation interrupted later real work")
	}
}

func TestWorkerPoolFiniteConcurrencyAndGracefulStop(t *testing.T) {
	_, db, objects := fixture(t)
	var active, maxActive atomic.Int32
	started := make(chan struct{}, 8)
	release := make(chan struct{})
	reader := &observedSource{Reader: importers.NewReader(nil), beforeReadContext: func(ctx context.Context) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := maxActive.Load(); n > old && !maxActive.CompareAndSwap(old, n); old = maxActive.Load() {
		}
		started <- struct{}{}
		select {
		case <-ctx.Done():
		case <-release:
		}
	}}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{WorkerConcurrency: 2})
	for i := 0; i < 8; i++ {
		c := importCommand(fmt.Sprint(i), fmt.Sprint(i))
		c.SourceLocator += "/blob/main/" + fmt.Sprint(i)
		if _, err := s.ImportMaterial(context.Background(), human, c); err != nil {
			t.Fatal("queue finite manual imports")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("pool did not reach two reader boundaries")
		}
	}
	if active.Load() != 2 || maxActive.Load() != 2 {
		t.Fatal("worker concurrency exceeded bound")
	}
	close(release)
	deadline := time.Now().Add(3 * time.Second)
	for {
		jobs, err := s.ListJobs(context.Background(), human)
		if err != nil {
			t.Fatal("read actual finite queue")
		}
		finished := 0
		for _, item := range jobs.Items {
			if item.(Job).Status == "succeeded" {
				finished++
			}
		}
		if finished == 8 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("pool did not drain eight actual imports")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal("worker pool did not stop gracefully")
	}
	if maxActive.Load() > 2 || reader.calls.Load() != 8 {
		t.Fatal("duplicate claim or unbounded production execution")
	}
}

func TestCanonicalContractKindsAndMutableSourceRefresh(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	for _, kind := range []string{"paper", "video", "text", "file"} {
		c := importCommand("kind-"+kind, "manual bytes-"+kind)
		c.Kind = kind
		if _, err := s.ImportMaterial(ctx, human, c); err != nil {
			t.Fatalf("legal contract kind %s rejected", kind)
		}
	}
	for _, kind := range []string{"note", "article"} {
		c := importCommand("bad-"+kind, "manual bytes")
		c.Kind = kind
		_, err := s.ImportMaterial(ctx, human, c)
		errorCode(t, err, apierrors.ValidationFailed)
	}
	_, db, objects := fixture(t)
	root, err := filepath.Abs("../../../.verification")
	if err != nil {
		t.Fatal("resolve owned export root")
	}
	dir, err := os.MkdirTemp(root, "refresh-export-")
	if err != nil {
		t.Fatal("create owned export directory")
	}
	file := filepath.Join(dir, "export.txt")
	if err = os.WriteFile(file, []byte("first manual content"), 0600); err != nil {
		t.Fatal("write selected manual content")
	}
	s2 := NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader([]string{dir}), objects, ServiceOptions{})
	c := importCommand("manual-first", "")
	c.LocalFileRef = file
	response, err := s2.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal("queue selected mutable export")
	}
	if _, err = s2.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual first export")
	}
	first, err := s2.GetJob(ctx, human, response.JobID)
	if err != nil || first.MaterialID == nil {
		t.Fatal("actual first export not saved")
	}
	if err = os.WriteFile(file, []byte("changed manual content"), 0600); err != nil {
		t.Fatal("update selected mutable export")
	}
	c.CommandMeta, c.Refresh = meta("manual-again", 1), true
	response, err = s2.ImportMaterial(ctx, human, c)
	if err != nil || response.JobID == first.JobID {
		t.Fatal("explicit refresh did not queue a new read")
	}
	if _, err = s2.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual changed export")
	}
	row, err := s2.GetMaterial(ctx, human, *first.MaterialID)
	if err != nil || len(row.Revisions) != 2 {
		t.Fatal("mutable file refresh did not append immutable revision")
	}
}
