package s5

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func TestActualLayaCompleteSourceInput(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	config := actualLayaConfig(t, "multilingual", "en", 1)
	question := judgment.Question{ID: "official-onnx-verification", Type: judgment.Truth, Instructions: "Does the official export documentation verify torch and ONNX numerical agreement?"}
	policy, goal, _ := actualSourcePolicy(t, f, f.policy(t, project, judgment.Truth, judgment.LayaBackend, config.ModelRevision), question, ref, "1")
	if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
		t.Fatal(err)
	}
	model, err := adapter.NewLaya(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := model.Close(); err != nil {
			t.Error(err)
		}
	})
	core := actualService(t, f, model)
	request := requestForPolicy(policy, question, ref)
	result, err := core.Evaluate(ctx, f.actor, request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Backend != judgment.LayaBackend || result.ProviderModel == nil || *result.ProviderModel != "laya-rl-agent" || result.ModelRevision == nil || *result.ModelRevision != config.ModelRevision || result.Value == nil || result.ProviderConfidence != nil || result.Usage == nil || result.Usage.InputTokens <= 0 || result.Usage.StateTokens == nil || *result.Usage.StateTokens <= 0 || result.Truncated || result.Decision != judgment.Abstain || result.ReasonCode != "truth_threshold_unconfigured" || !reflect.DeepEqual(result.Sources, policy.Sources) {
		t.Fatal("actual Laya did not preserve complete source input, pinned identity, real usage or the explicit manual decision")
	}
	cached, err := core.Evaluate(ctx, f.actor, request)
	if err != nil || !cached.CacheHit || cached.ID != result.ID {
		t.Fatal("actual model source result was not reused from the authorized SQLite cache", err)
	}
	goal.Title = "Verify the documented local CPU parity and loading contract"
	if _, err := f.access.Goals.(*workspace.S2Service).PutGoal(ctx, f.caller, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: goal.Revision}, Goal: goal}); err != nil {
		t.Fatal(err)
	}
	_, err = core.Evaluate(ctx, f.actor, request)
	requireServiceCode(t, err, apierrors.ContextStale)
	t.Logf("one actual CPU inference with complete Goal/taste/example input: input_tokens=%d state_tokens=%d truncated=%v confidence=null decision=%s; same source cache and actual Goal change checked", result.Usage.InputTokens, *result.Usage.StateTokens, result.Truncated, result.Decision)
}
