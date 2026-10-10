package distillers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workdomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type selectedProcessorStub struct {
	t      *testing.T
	output string
	calls  int
}

// Cross-adapter test bridge only: application consent is fixed locally, but
// engineering request bounds are checked by the actual W2 Registry. A cancelled
// context stops before configuration observation, directories or native calls.
type nativeBoundaryProcessor struct{ registry *agents.Registry }

func (p nativeBoundaryProcessor) ConfigurationID(context.Context, app.Principal, string, string) (string, error) {
	return "contract-local-permitted-config", nil
}
func (p nativeBoundaryProcessor) ProjectSpaceID(context.Context, app.Principal, string, string) (string, error) {
	return "space", nil
}
func (p nativeBoundaryProcessor) ProcessSelectedText(ctx context.Context, _ app.Principal, request app.SelectedTextRequest) (app.SelectedTextResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err := p.registry.ProcessSelectedText(ctx, workdomain.TextRequest{CLI: request.CLI, Prompt: request.Prompt, DeadlineSeconds: request.DeadlineSeconds, OutputLimit: request.OutputLimit})
	return app.SelectedTextResult{}, err
}

func TestSelectedTextRequestsPassActualNativeBoundsWithoutLaunching(t *testing.T) {
	ctx := context.Background()
	processor := nativeBoundaryProcessor{registry: agents.NewRegistry(t.TempDir())}
	caller := app.Principal{ID: "real-caller", Kind: "human"}
	d, err := NewSelectedTextFactory(processor, 30*time.Minute).Resolve(ctx, caller, "project", "codex")
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Distill(ctx, app.DistillationInput{Stage: "summary", Inputs: []app.SourceSnapshot{{Ref: app.SourceRef{MaterialID: "material", Revision: 1}, Text: "fixed source text"}}})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("distillation did not pass actual native bounds before cancellation", err)
	}
	_, err = NewListingRecommender(processor, 30*time.Minute).Recommend(ctx, app.ListingRecommendationInput{Caller: caller, ProjectID: "project", CLI: "codex", Items: []app.SourceItem{{SourceID: "source", ExternalID: "id", Revision: 1, Metadata: app.ListingMetadata{Title: "fixed public metadata"}}}})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("recommendation did not pass actual native bounds before cancellation", err)
	}
}

func (s *selectedProcessorStub) ConfigurationID(_ context.Context, p app.Principal, project, cli string) (string, error) {
	if p.Kind != "human" || p.ID != "real-caller" || project != "project" || cli != "selected-cli" {
		s.t.Fatal("caller/project/CLI identity changed")
	}
	return "actual-config", nil
}
func (s *selectedProcessorStub) ProjectSpaceID(context.Context, app.Principal, string, string) (string, error) {
	return "space", nil
}
func (s *selectedProcessorStub) ProcessSelectedText(ctx context.Context, p app.Principal, request app.SelectedTextRequest) (app.SelectedTextResult, error) {
	s.calls++
	if p.ID != "real-caller" || request.CLI != "selected-cli" || request.ProjectID != "project" || request.DeadlineSeconds < 1 || request.DeadlineSeconds > 1800 || !strings.Contains(request.Prompt, "metadata only title") {
		s.t.Fatal("selected context or engineering bounds changed")
	}
	return app.SelectedTextResult{Text: s.output, Version: "contract-local-version", Model: nil}, nil
}

func TestSelectedCLIRecommendationHasFixedMetadataAndHonestProvenance(t *testing.T) {
	p := &selectedProcessorStub{t: t, output: `{"recommendations":[{"external_id":"2:123","text":"建议先读简介","reason":"只有标题简介，尚未阅读正文"}]}`}
	r := NewListingRecommender(p, 30*time.Minute)
	input := app.ListingRecommendationInput{Caller: app.Principal{ID: "real-caller", Kind: "human"}, ProjectID: "project", CLI: "selected-cli", JobID: "job", OperationID: "operation", Items: []app.SourceItem{{SourceID: "source", ExternalID: "2:123", Revision: 7, Metadata: app.ListingMetadata{ExternalID: "2:123", Title: "metadata only title", Description: "description without body"}}}}
	if _, err := r.ConfigurationID(context.Background(), input.Caller, input.ProjectID, input.CLI); err != nil {
		t.Fatal(err)
	}
	out, err := r.Recommend(context.Background(), input)
	if err != nil || out["2:123"].MetadataRevision != 7 || out["2:123"].Provenance.Processor != "selected-cli" || out["2:123"].Provenance.Model != "" || p.calls != 1 {
		t.Fatal("fixed revision or unknown model changed", err, out)
	}
	for _, bad := range []string{`{"recommendations":[{"external_id":"invented","text":"x","reason":"y"}]}`, `{"recommendations":[{"external_id":"2:123","text":"x","reason":"y","score":10}]}`, p.output + " trailing"} {
		p.output = bad
		if _, err = r.Recommend(context.Background(), input); err == nil {
			t.Fatal("invented/scored/trailing output accepted", bad)
		}
	}
}
