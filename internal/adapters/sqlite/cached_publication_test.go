package sqlite

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
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

type cachedProjectText struct {
	result             *localReturnedResult
	spaceID            string
	unknown, revoked   bool
	configurationCalls int
}

func (p *cachedProjectText) ProjectSpaceID(context.Context, app.Principal, string, string) (string, error) {
	if p.revoked {
		return "", &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "contract-local project model permission revoked"}
	}
	return p.spaceID, nil
}
func (p *cachedProjectText) ConfigurationID(context.Context, app.Principal, string, string) (string, error) {
	p.configurationCalls++
	if p.unknown {
		return "", &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "native configuration unknown after restart"}
	}
	return "contract-local-selected-config", nil
}
func (p *cachedProjectText) ProcessSelectedText(ctx context.Context, _ app.Principal, _ app.SelectedTextRequest) (app.SelectedTextResult, error) {
	output, err := p.result.Distill(ctx, app.DistillationInput{})
	if err != nil {
		return app.SelectedTextResult{}, err
	}
	raw, err := json.Marshal(map[string]any{"output_text": output.OutputText, "next_question": nil, "related_refs": []app.SourceRef{}, "related_ideas": []string{}, "conflicts": []string{}, "pending_questions": []string{}, "goal_refs": []string{}, "existing_assets": []string{}, "expected_improvement": "", "minimum_artifact": "", "missing_evidence": []string{}, "candidate_suggestion": nil})
	return app.SelectedTextResult{Text: string(raw), Version: "contract-local-selected-version"}, err
}

func TestReturnedResultRealOSPublicationFailureCanRetryAfterRestartWithoutProcessor(t *testing.T) {
	for _, mode := range []string{"legacy", "native_unknown", "revoked", "space_changed", "reference_removed"} {
		t.Run(mode, func(t *testing.T) { returnedResultOSRecovery(t, mode) })
	}
}

func returnedResultOSRecovery(t *testing.T, mode string) {
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
	selected := &cachedProjectText{result: d}
	if mode != "legacy" {
		options.Distiller = nil
		options.ProjectDistillers = distillers.NewSelectedTextFactory(selected, time.Minute)
	}
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
	command := app.RequestDistillationCommand{CommandMeta: meta("process", 1), Stage: "content", ProcessingConfig: "contract-local", InputRefs: []app.SourceRef{{MaterialID: *source.MaterialID, Revision: 1, Locator: "https://arxiv.org/abs/2504.16054v1"}}}
	spaceVersion := 0
	if mode != "legacy" {
		space, err := s.CreateProjectSpace(ctx, human, app.ProjectSpaceCommand{CommandMeta: meta("project-space", 1), Title: "Contract-local selected project"})
		if err != nil {
			t.Fatal(err)
		}
		space, err = s.ReferenceMaterial(ctx, human, space.Space.ID, app.ReferenceMaterialCommand{CommandMeta: meta("project-source", space.Space.Version), MaterialID: *source.MaterialID, Revision: 1})
		if err != nil {
			t.Fatal(err)
		}
		selected.spaceID, spaceVersion = space.Space.ID, space.Space.Version
		command.ProjectID, command.CLI = "selected-project", "selected-cli"
		command.ProcessingConfig = ""
	}
	queued, err := s.RequestDistillation(ctx, human, command)
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
	selected.unknown, selected.configurationCalls = true, 0
	if mode == "revoked" {
		selected.revoked = true
	}
	if mode == "space_changed" {
		selected.spaceID = "changed-project-space"
	}
	s = app.NewAttentionService(NewAttentionRepository(db), reader, store, options)
	if mode == "reference_removed" {
		if _, err = s.RemoveMaterialReference(ctx, human, selected.spaceID, app.RemoveMaterialReferenceCommand{CommandMeta: meta("remove-scope", spaceVersion), MaterialID: *source.MaterialID}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.RetryJob(ctx, human, failed.JobID, meta("retry-after-repair", failed.Version)); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	done, _ := s.GetJob(ctx, human, failed.JobID)
	if selected.configurationCalls != 0 || d.calls != 1 {
		t.Fatal("cached selected result observed native config or resent processor", selected.configurationCalls, d.calls)
	}
	if mode == "revoked" || mode == "space_changed" || mode == "reference_removed" {
		if done.Status != "failed" || done.Error == nil || done.Error.Code != apierrors.ScopeDenied || done.DistillationID != nil || done.DeliveryUnknown {
			t.Fatal("revoked/changed fixed project scope allowed cached publication", done)
		}
		return
	}
	if done.Status != "succeeded" || done.DistillationID == nil || done.OperationID != operation || !done.DeadlineAt.Equal(deadline) || done.Attempts != 2 || d.calls != 1 {
		t.Fatal("repair lost result/budget/operation or resent processor", done, d.calls)
	}
	if mode == "native_unknown" {
		command.CommandMeta = meta("fresh-model-after-restart", 1)
		command.Question = "New processing after cold restart"
		fresh, err := s.RequestDistillation(ctx, human, command)
		if err != nil {
			t.Fatal(err)
		}
		freshJob, _ := s.GetJob(ctx, human, fresh.JobID)
		if freshJob.Status != "failed" || freshJob.Error == nil || freshJob.Error.Code != apierrors.EvidenceMissing || d.calls != 1 || selected.configurationCalls != 1 {
			t.Fatal("fresh processing bypassed native configuration check", freshJob, d.calls, selected.configurationCalls)
		}
	}
}
