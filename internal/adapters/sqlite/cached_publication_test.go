package sqlite

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type localReturnedResult struct {
	t                 *testing.T
	objectRoot, saved string
	calls             int
}

func (d *localReturnedResult) ConfigurationID(context.Context) (string, error) {
	return "contract-local-result", nil
}
func (d *localReturnedResult) Distill(context.Context, app.DistillationInput) (app.DistillationOutput, error) {
	d.calls++
	// Both paths are constructed underneath this test's owned TempDir. The
	// original objects survive; only the expected directory becomes a file.
	if err := os.Rename(d.objectRoot, d.saved); err != nil {
		d.t.Fatal(err)
	}
	if err := os.WriteFile(d.objectRoot, []byte("owned test publication obstruction"), 0600); err != nil {
		d.t.Fatal(err)
	}
	return app.DistillationOutput{OutputText: "Contract-local returned result", Provenance: app.Provenance{Processor: "local-test", Version: "1", Mode: "contract_local"}}, nil
}

func TestReturnedResultRealOSPublicationFailureCanRetryAfterRestartWithoutProcessor(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath, objectRoot := filepath.Join(dir, "state.sqlite"), filepath.Join(dir, "objects")
	db := openAttentionDB(t, dbPath)
	store, err := objects.New(objectRoot)
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	reader.Arxiv.Client = &http.Client{Transport: paperExportTransport{}}
	d := &localReturnedResult{t: t, objectRoot: objectRoot, saved: filepath.Join(dir, "preserved-objects")}
	options := app.ServiceOptions{Distiller: d, AllowedProcessingSourceKeys: []string{"arxiv:2504.16054"}}
	s := app.NewAttentionService(NewAttentionRepository(db), reader, store, options)
	human := app.Principal{ID: "cached-result-test", Kind: "human"}
	meta := func(key string, version int) app.CommandMeta {
		return app.CommandMeta{SchemaVersion: 1, RequestID: key, IdempotencyKey: key, ExpectedVersion: version}
	}
	imported, err := s.ImportMaterial(ctx, human, app.ImportMaterialCommand{CommandMeta: meta("source", 1), Adapter: "arxiv", Kind: "paper", SourceLocator: "2504.16054v1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	source, _ := s.GetJob(ctx, human, imported.JobID)
	if source.MaterialID == nil {
		t.Fatal("source import failed", source)
	}
	queued, err := s.RequestDistillation(ctx, human, app.RequestDistillationCommand{CommandMeta: meta("process", 1), Stage: "content", ProcessingConfig: "contract-local", InputRefs: []app.SourceRef{{MaterialID: *source.MaterialID, Revision: 1, Locator: "https://arxiv.org/abs/2504.16054v1"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	failed, _ := s.GetJob(ctx, human, queued.JobID)
	var payload struct{ Result *app.DistillationOutput }
	if err = NewAttentionRepository(db).WithTx(ctx, func(tx app.AttentionTx) error {
		job, err := tx.LoadJob(failed.JobID)
		if err != nil {
			return err
		}
		return json.Unmarshal(job.Payload, &payload)
	}); err != nil {
		t.Fatal(err)
	}
	if failed.Status != "failed" || failed.DeliveryUnknown || failed.Error == nil || !failed.Error.Retryable || failed.Error.RequiredAction != "repair_storage_then_retry_cached_result" || payload.Result == nil || d.calls != 1 {
		t.Fatal("real OS failure hid safely cached result or disabled repair/retry", failed, payload, d.calls)
	}
	var failureEvent string
	if err = db.Conn().QueryRowContext(ctx, "SELECT payload FROM attention_outbox WHERE aggregate_id=? AND type='attention.job_failed' ORDER BY occurred_at DESC LIMIT 1", failed.JobID).Scan(&failureEvent); err != nil || !strings.Contains(failureEvent, "cause_detail") || !strings.Contains(failureEvent, "objects") {
		t.Fatal("actual local publication cause was not preserved", failureEvent, err)
	}
	operation, deadline := failed.OperationID, failed.DeadlineAt
	if err = os.Remove(objectRoot); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(d.saved, objectRoot); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, dbPath)
	options.Distiller = nil // Result must survive a real restart without native availability.
	s = app.NewAttentionService(NewAttentionRepository(db), reader, store, options)
	if _, err = s.RetryJob(ctx, human, failed.JobID, meta("retry-after-repair", failed.Version)); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	done, _ := s.GetJob(ctx, human, failed.JobID)
	if done.Status != "succeeded" || done.DistillationID == nil || done.OperationID != operation || !done.DeadlineAt.Equal(deadline) || done.Attempts != 2 || d.calls != 1 {
		t.Fatal("repair lost result/budget/operation or resent processor", done, d.calls)
	}
}
