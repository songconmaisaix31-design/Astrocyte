package s1_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// Error injection only: never returns model output or claims real inference.
type lostContractLocalProcessor struct {
	calls int
	input app.DistillationInput
}

func (*lostContractLocalProcessor) ConfigurationID(context.Context) (string, error) {
	return "contract_local:error-only-config-v1", nil
}
func (p *lostContractLocalProcessor) Distill(_ context.Context, input app.DistillationInput) (app.DistillationOutput, error) {
	p.calls++
	p.input = input
	return app.DistillationOutput{}, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "contract_local injected lost result", RequiredAction: "reconcile_original_operation"}
}
func localCommand(version int) app.CommandMeta {
	return app.CommandMeta{SchemaVersion: 1, RequestID: rand.Text(), IdempotencyKey: rand.Text(), ExpectedVersion: version}
}

// Real application API + SQLite/objects + database reopen, with local Atom/PDF
// source responses and one injected processor error. This is contract_local,
// not the separate native CLI timeout or a public-network/model success.
func TestContractLocalAutomaticUnknownSurvivesSQLiteReopenWithoutRetry(t *testing.T) {
	ctx := context.Background()
	p := app.Principal{ID: "contract_local-human", Kind: "human"}
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/query" {
			id := r.URL.Query().Get("id_list")
			w.Header().Set("Content-Type", "application/atom+xml")
			fmt.Fprintf(w, `<feed><entry><id>https://arxiv.org/abs/%s</id><title>contract_local paper %s</title><summary>contract_local metadata %s</summary></entry></feed>`, id, id, id)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/pdf/") {
			w.Header().Set("Content-Type", "application/pdf")
			fmt.Fprintf(w, "%%PDF-1.7\ncontract_local representative %s\n", r.URL.Path)
			return
		}
		http.NotFound(w, r)
	}))
	defer fixture.Close()
	target, err := url.Parse(fixture.URL)
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	reader.Arxiv.Client = &http.Client{Transport: localArxivTransport{target: target}}
	dir := t.TempDir()
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	processor := &lostContractLocalProcessor{}
	open := func() (*sqlite.DB, *app.Service) {
		db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		if err = db.RunMigrations(ctx, migrations.FS); err != nil {
			db.Close()
			t.Fatal(err)
		}
		service := app.NewAttentionService(sqlite.NewAttentionRepository(db), reader, store, app.ServiceOptions{Distiller: processor, AllowedProcessingSourceKeys: []string{"arxiv:2501.12948"}})
		return db, service
	}
	db, service := open()
	defer func() { db.Close() }()
	var material app.MaterialDetail
	for _, revision := range []string{"v1", "v2"} {
		receipt, err := service.ImportMaterial(ctx, p, app.ImportMaterialCommand{CommandMeta: localCommand(1), Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948" + revision})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = service.ProcessNextJob(ctx); err != nil {
			t.Fatal(err)
		}
		job, err := service.GetJob(ctx, p, receipt.JobID)
		if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
			t.Fatalf("import: %+v %v", job, err)
		}
		material, err = service.GetMaterial(ctx, p, *job.MaterialID)
		if err != nil {
			t.Fatal(err)
		}
	}
	if material.Material.CurrentRevision != 2 {
		t.Fatal("second source revision missing")
	}
	oldText, err := service.GetContent(ctx, p, material.Material.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	request := app.RequestDistillationCommand{CommandMeta: localCommand(1), InputRefs: []app.SourceRef{{MaterialID: material.Material.ID, Revision: 1, Locator: material.Revisions[0].SourceLocator}}, Stage: "content", ProcessingConfig: "contract_local-selected-v1", Question: "contract_local old revision failure case"}
	original, err := service.RequestDistillation(ctx, p, request)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := request
	duplicate.CommandMeta = localCommand(1)
	reused, err := service.RequestDistillation(ctx, p, duplicate)
	if err != nil || reused.JobID != original.JobID {
		t.Fatalf("dedupe: %+v %v", reused, err)
	}
	if _, err = service.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	job, err := service.GetJob(ctx, p, original.JobID)
	if err != nil || job.Status != "failed" || !job.DeliveryUnknown || job.Error == nil || job.Error.Code != apierrors.DeliveryUnknown {
		t.Fatalf("lost result: %+v %v", job, err)
	}
	if processor.calls != 1 || len(processor.input.Inputs) != 1 || processor.input.Inputs[0].Ref.Revision != 1 || processor.input.Inputs[0].Text != oldText.Text {
		t.Fatal("processor input was not fixed to the selected old revision")
	}
	var persistedPayload, persistedCaller string
	if err = db.Conn().QueryRowContext(ctx, "SELECT payload,caller FROM attention_jobs WHERE id=?", job.JobID).Scan(&persistedPayload, &persistedCaller); err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err = json.Unmarshal([]byte(persistedPayload), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload["Result"]) > 0 && string(payload["Result"]) != "null" {
		t.Fatal("unknown result was fabricated")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, service = open()
	if err = service.RecoverJobs(ctx); err != nil {
		t.Fatal(err)
	}
	restarted, err := service.GetJob(ctx, p, job.JobID)
	if err != nil || !restarted.DeliveryUnknown || restarted.OperationID != job.OperationID {
		t.Fatalf("restart: %+v %v", restarted, err)
	}
	_, err = service.RetryJob(ctx, p, job.JobID, localCommand(restarted.Version))
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.DeliveryUnknown {
		t.Fatalf("unknown retry allowed: %v", err)
	}
	service.ProcessNextJob(ctx)
	if processor.calls != 1 {
		t.Fatal("reopen or retry resent unknown work")
	}
	var afterPayload, afterCaller string
	if err = db.Conn().QueryRowContext(ctx, "SELECT payload,caller FROM attention_jobs WHERE id=?", job.JobID).Scan(&afterPayload, &afterCaller); err != nil {
		t.Fatal(err)
	}
	if afterPayload != persistedPayload || afterCaller != persistedCaller {
		t.Fatal("unknown request/caller changed after restart")
	}
	var outputs, candidates int
	if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_distillations").Scan(&outputs); err != nil {
		t.Fatal(err)
	}
	if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_opportunities").Scan(&candidates); err != nil {
		t.Fatal(err)
	}
	if outputs != 0 || candidates != 0 {
		t.Fatal("lost result generated a record or candidate")
	}
	observed, err := service.GetMaterial(ctx, p, material.Material.ID)
	if err != nil || observed.Material.HumanUsageCount != material.Material.HumanUsageCount || observed.Material.AgentUsageCount != material.Material.AgentUsageCount+1 {
		t.Fatal("internal processor retrieval became a human signal or was counted twice")
	}
	if content, err := service.GetContent(ctx, p, material.Material.ID, 1); err != nil || content.Text != oldText.Text {
		t.Fatal("unknown processing lost the original material")
	}
}
