package s5

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

func TestActualLayaJudgmentServiceSafety(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	config := actualLayaConfig(t, "multilingual", "en", 1)
	model, err := adapter.NewLaya(ctx, config)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer cancel()
		if err := model.Close(); err != nil {
			t.Error(err)
		}
	})
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	service := actualService(t, f, model)
	questions := []judgment.Question{
		{ID: "public-runtime", Type: judgment.Choice, Instructions: "Which runtime is described?", Options: []judgment.Option{{ID: "typescript", Description: "TypeScript inference engine"}, {ID: "music", Description: "Music playback"}}},
		{ID: "public-export", Type: judgment.Truth, Instructions: "Does the export documentation verify torch and ONNX numerical agreement?"},
		{ID: "public-verification-detail", Type: judgment.Score, Instructions: "How explicit is the export verification?", Levels: []string{"No verification", "Verification without tolerance", "Verification with numerical tolerance"}},
	}
	policies := map[judgment.QuestionType]judgment.Policy{}
	for _, question := range questions {
		policy := f.policy(t, project, question.Type, judgment.LayaBackend, config.ModelRevision)
		if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
			t.Fatal(err)
		}
		policies[question.Type] = policy
		request := requestForPolicy(policy, question, ref)
		result, err := service.Evaluate(ctx, f.actor, request)
		if err != nil {
			t.Fatal(err)
		}
		if result.Decision != judgment.Abstain || result.Truncated || result.ProviderModel == nil || *result.ProviderModel != "laya-rl-agent" || result.ModelRevision == nil || *result.ModelRevision != config.ModelRevision || result.Usage == nil || result.Usage.InputTokens <= 0 {
			t.Fatalf("actual model result lost its identity, usage or manual decision: %+v", result)
		}
		if question.Type == judgment.Truth {
			if result.ProviderConfidence != nil || result.ReasonCode != "truth_threshold_unconfigured" {
				t.Fatal("actual Noul confidence or unconfigured policy was changed")
			}
		} else if result.ProviderConfidence == nil || result.ReasonCode != "confidence_threshold_unconfigured" {
			t.Fatal("an unconfigured confidence policy admitted an actual model decision")
		}
		cached, err := service.Evaluate(ctx, f.actor, request)
		if err != nil || !cached.CacheHit || cached.ID != result.ID || cached.Decision != judgment.Abstain {
			t.Fatal("cached model judgment changed the actual abstention", err)
		}
		t.Logf("actual %s: decision=%s, input_tokens=%d", question.Type, result.Decision, result.Usage.InputTokens)
	}

	readme, err := os.ReadFile(filepath.Join(s5Root(t), ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	longRef := f.importPublicText(t, string(readme), "https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md", "Official complete Laya README", []string{"Complete README"})
	space, err := f.attention.GetProjectSpace(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, project.SpaceID)
	if err != nil {
		t.Fatal(err)
	}
	meta := commandMeta()
	meta.ExpectedVersion = space.Space.Version
	if _, err := f.attention.ReferenceMaterial(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, project.SpaceID, attention.ReferenceMaterialCommand{CommandMeta: meta, MaterialID: longRef.MaterialID, Revision: longRef.Revision}); err != nil {
		t.Fatal(err)
	}
	longRequest := requestForPolicy(policies[judgment.Truth], questions[1], longRef)
	long, err := service.Evaluate(ctx, f.actor, longRequest)
	if err != nil || !long.Truncated || long.Decision != judgment.Abstain || long.ReasonCode != "input_truncated" || long.Usage == nil || long.Usage.StateTokensDropped == nil || *long.Usage.StateTokensDropped <= 0 {
		t.Fatalf("actual long official document did not preserve truncation and abstain: %+v, %v", long, err)
	}
	space, err = f.attention.GetProjectSpace(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, project.SpaceID)
	if err != nil {
		t.Fatal(err)
	}
	meta = commandMeta()
	meta.ExpectedVersion = space.Space.Version
	if _, err := f.attention.ReferenceMaterial(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, project.SpaceID, attention.ReferenceMaterialCommand{CommandMeta: meta, MaterialID: ref.MaterialID, Revision: ref.Revision}); err != nil {
		t.Fatal(err)
	}
	optionsJSON, err := exec.CommandContext(ctx, config.Node, filepath.Join(s5Root(t), "tests", "s5", "public-options.mjs")).Output()
	if err != nil {
		t.Fatal(err)
	}
	var options []judgment.Option
	if err := json.Unmarshal(optionsJSON, &options); err != nil {
		t.Fatal(err)
	}
	optionQuestion := judgment.Question{ID: "public-documentation-passages", Type: judgment.Choice, Instructions: "Which section documents CPU Node inference?", Options: options}
	optionResult, err := service.Evaluate(ctx, f.actor, requestForPolicy(policies[judgment.Choice], optionQuestion, ref))
	coverage, found := optionResult.Options[optionQuestion.ID]
	if err != nil || !found || coverage.Total != len(options) || coverage.Distinct >= coverage.Total || optionResult.Decision != judgment.Abstain {
		t.Fatalf("actual collapsed documentation options did not abstain: %+v, %v", optionResult, err)
	}
	opportunity := actualOpportunity(t, f, ref, "Verify public CPU export documentation", "How does the documented export verify the local runtime graphs?")
	delegation, err := judgment.NewDelegationService(f.db, service, f.access, adapter.AttentionOpportunities{Attention: f.attention})
	if err != nil {
		t.Fatal(err)
	}
	admissionOption := questions[0].Options[0].ID
	rule := judgment.DelegationRule{ID: uuid.NewString(), ProjectID: project.ID, WorkflowID: uuid.NewString(), OwnerID: f.actor.ID, Revision: 1, State: "active", Purpose: judgment.AdmissionPurpose, Question: questions[0], PolicyRevision: policies[judgment.Choice].Revision, Language: "en", Materials: []judgment.EvidenceRef{ref}, AdmissionChoice: &admissionOption, ReviewConditions: []string{"fixed_inputs"}}
	if err := delegation.SetRule(ctx, f.actor, rule, 0); err != nil {
		t.Fatal(err)
	}
	bound, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, opportunity.ID, opportunity.Revision)
	if err != nil || bound.ProviderConfidence == nil || *bound.ProviderConfidence >= 1 || bound.Decision != judgment.Abstain {
		t.Fatal("the actual model did not establish a finite uncertain bound judgment", err)
	}
	lowPolicy := policies[judgment.Choice]
	lowPolicy.Revision++
	explicitReviewThreshold := (1 + *bound.ProviderConfidence) / 2
	lowPolicy.MinConfidence = &explicitReviewThreshold
	if err := f.access.SetPolicy(ctx, f.actor, lowPolicy, lowPolicy.Revision-1); err != nil {
		t.Fatal(err)
	}
	policies[judgment.Choice] = lowPolicy
	rule.Revision++
	rule.PolicyRevision = lowPolicy.Revision
	if err := delegation.SetRule(ctx, f.actor, rule, rule.Revision-1); err != nil {
		t.Fatal(err)
	}
	low, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, opportunity.ID, opportunity.Revision)
	if err != nil || low.ReasonCode != "low_provider_confidence" || low.Decision != judgment.Abstain {
		t.Fatal("the actual model confidence did not enforce the explicitly submitted review threshold", err)
	}
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, low.ID, project.ID, opportunity.ID, opportunity.Revision)
	requireServiceCode(t, err, apierrors.EvidenceMissing)
	rule.Revision++
	rule.Question = optionQuestion
	rule.AdmissionChoice = &options[0].ID
	if err := delegation.SetRule(ctx, f.actor, rule, rule.Revision-1); err != nil {
		t.Fatal(err)
	}
	collapsed, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, opportunity.ID, opportunity.Revision)
	if err != nil || collapsed.Decision != judgment.Abstain || len(collapsed.Options) == 0 {
		t.Fatal("actual collapsed model options did not preserve an unusable bound result", err)
	}
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, collapsed.ID, project.ID, opportunity.ID, opportunity.Revision)
	requireServiceCode(t, err, apierrors.EvidenceMissing)
	missingRequest := requestForPolicy(policies[judgment.Truth], questions[1])
	missing, err := service.Evaluate(ctx, f.actor, missingRequest)
	if err != nil || missing.Decision != judgment.NeedEvidence || missing.Usage != nil || missing.Backend != judgment.RuleBackend {
		t.Fatal("absent actual evidence reached a model", err)
	}
	cloud := f.policy(t, project, judgment.Truth, judgment.JevBackend, adapter.JevModel)
	cloud.Revision = policies[judgment.Truth].Revision + 1
	cloud.ExternalScope = "fixed-public-documentation"
	if err := f.access.SetPolicy(ctx, f.actor, cloud, policies[judgment.Truth].Revision); err != nil {
		t.Fatal(err)
	}
	_, err = service.Evaluate(ctx, f.actor, requestForPolicy(cloud, questions[1], ref))
	requireServiceCode(t, err, apierrors.ScopeDenied)
	policy := policies[judgment.Choice]
	if err := model.Close(); err != nil {
		t.Fatal(err)
	}
	failureRequest := requestForPolicy(policy, questions[0], ref)
	failureRequest.Question.ID = "closed-real-worker"
	failed, err := service.Evaluate(ctx, f.actor, failureRequest)
	if err == nil || failed.Decision != judgment.Abstain || failed.ReasonCode != "backend_call_failed" {
		t.Fatal("an actual closed worker produced a successful decision")
	}
	stored, readErr := service.Read(ctx, f.actor, failed.ID)
	if readErr != nil || stored.ReasonCode != "backend_call_failed" || stored.Decision != judgment.Abstain {
		t.Fatal("actual backend failure was not retained", readErr)
	}
	again, err := service.Evaluate(ctx, f.actor, failureRequest)
	if err == nil || again.CacheHit || again.ID == failed.ID {
		t.Fatal("an actual failed model request was cached or automatically restarted")
	}
	t.Logf("actual model=%s, revision=%s, input_tokens=%d, long_dropped=%d, options=%d/%d", *long.ProviderModel, *long.ModelRevision, long.Usage.InputTokens, *long.Usage.StateTokensDropped, coverage.Distinct, coverage.Total)
}
