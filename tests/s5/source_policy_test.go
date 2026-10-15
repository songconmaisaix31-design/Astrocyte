package s5

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func actualSourcePolicy(t *testing.T, f *fixture, policy judgment.Policy, question judgment.Question, ref judgment.EvidenceRef, expected string) (judgment.Policy, domain.Goal, attention.TasteProfile) {
	t.Helper()
	ctx := context.Background()
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(f.db), agents.ProjectFiles{}, adapter.AttentionReferences{Attention: f.attention})
	f.access.Goals, f.access.Ranking, f.access.Taste, f.access.Rules = s2, f.attention, f.attention, f.db
	goal, err := s2.PutGoal(ctx, f.caller, "", workspace.GoalCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString()}, Goal: domain.Goal{ProjectID: policy.ProjectID, Title: "Verify the fixed official Laya CPU documentation", SuccessCriteria: []string{"Read the official export verification and local runtime contract"}, Priority: 1, Status: "active"}})
	if err != nil {
		t.Fatal(err)
	}
	rubric := attention.TasteRubric{ID: question.ID, QuestionType: string(question.Type), Criteria: question.Instructions, Options: []attention.TasteOption{}, Levels: append([]string{}, question.Levels...)}
	for _, option := range question.Options {
		rubric.Options = append(rubric.Options, attention.TasteOption{ID: option.ID, Description: option.Description})
	}
	profile, err := f.attention.UpdateTasteProfile(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, attention.UpdateTasteProfileCommand{CommandMeta: commandMeta(), Enabled: true, Rubrics: []attention.TasteRubric{rubric}, Examples: []attention.TasteExample{{ID: "official-public-source-review", RubricID: question.ID, MaterialID: ref.MaterialID, Revision: ref.Revision, ContentDigest: ref.Digest, Expected: expected}}, Delegations: []attention.TasteDelegation{}})
	if err != nil || profile.Profile == nil {
		t.Fatal("actual explicit complete taste configuration failed", err)
	}
	policy.Sources = &judgment.SourceReferences{Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Taste: judgment.ProfileReference{ID: profile.Profile.ID, Version: profile.Profile.Version}}
	policy.Versions.Goal, policy.Versions.Taste = fmt.Sprintf("%s@%d", goal.ID, goal.Revision), fmt.Sprintf("%s@%d", profile.Profile.ID, profile.Profile.Version)
	return policy, goal, *profile.Profile
}

func TestActualCompleteSourcesInvalidateCacheAndAdmission(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	question := judgment.Question{ID: "fixed-source-availability", Type: judgment.Truth, Instructions: "Are all selected fixed official source sections present?", Rule: "all_evidence_present"}
	policy, goal, taste := actualSourcePolicy(t, f, f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1"), question, ref, "1")
	principal := attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}
	ranking, err := f.attention.UpdateRankingProfile(ctx, principal, attention.UpdateRankingProfileCommand{CommandMeta: commandMeta(), Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}})
	if err != nil {
		t.Fatal(err)
	}
	policy.Sources.Ranking = &judgment.ProfileReference{ID: ranking.Profile.ID, Version: ranking.Profile.Version}
	if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
		t.Fatal(err)
	}
	core := actualService(t, f)
	delegation, err := judgment.NewDelegationService(f.db, core, f.access, adapter.AttentionOpportunities{Attention: f.attention})
	if err != nil {
		t.Fatal(err)
	}
	opportunity := actualOpportunity(t, f, ref, "Verify official CPU export numerical agreement", question.Instructions)
	accepted := 1.0
	rule := judgment.DelegationRule{ID: uuid.NewString(), ProjectID: project.ID, WorkflowID: uuid.NewString(), OwnerID: f.actor.ID, Revision: 1, State: "active", Purpose: judgment.AdmissionPurpose, Question: question, PolicyRevision: policy.Revision, Language: "en", Materials: []judgment.EvidenceRef{ref}, AdmissionValueMin: &accepted, AdmissionValueMax: &accepted, ReviewConditions: []string{"fixed_inputs"}}
	if err := delegation.SetRule(ctx, f.actor, rule, 0); err != nil {
		t.Fatal(err)
	}
	var previous judgment.JudgmentResult
	var request judgment.JudgmentRequest
	checkCurrent := func() {
		t.Helper()
		request = requestForPolicy(policy, question, ref)
		current, err := core.Evaluate(ctx, f.actor, request)
		if err != nil || current.Decision != judgment.Recommend || current.CacheHit || current.ID == previous.ID || previous.ID != "" && current.InputDigest == previous.InputDigest {
			t.Fatal("actual changed source contents were not included in a fresh judgment input", err)
		}
		if previous.ID == "" {
			cached, err := core.Evaluate(ctx, f.actor, request)
			if err != nil || !cached.CacheHit || cached.ID != current.ID {
				t.Fatal("the same complete actual sources did not reuse their authorized cache", err)
			}
		}
		previous = current
	}
	var bound judgment.JudgmentResult
	bindCurrent := func() {
		t.Helper()
		bound, err = delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, opportunity.ID, opportunity.Revision)
		if err != nil {
			t.Fatal(err)
		}
		waits := f.db.Conn().Stats().WaitCount
		err = f.db.WithJudgmentTx(ctx, func(ctx context.Context) error {
			_, _, _, err := delegation.Authorize(ctx, rule.WorkflowID, bound.ID, project.ID, opportunity.ID, opportunity.Revision)
			return err
		})
		if err != nil || waits != f.db.Conn().Stats().WaitCount {
			t.Fatal("actual full source admission did not recheck within the original SQLite transaction", err)
		}
	}
	checkStale := func(code apierrors.Code) {
		t.Helper()
		_, err := core.Evaluate(ctx, f.actor, request)
		requireServiceCode(t, err, code)
		_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, bound.ID, project.ID, opportunity.ID, opportunity.Revision)
		requireServiceCode(t, err, code)
		if _, err := core.Read(ctx, f.actor, previous.ID); err != nil {
			t.Fatal("historical authorized result disappeared after a source revision", err)
		}
	}
	refresh := func() {
		t.Helper()
		policy.Revision++
		policy.Versions.Goal = fmt.Sprintf("%s@%d", policy.Sources.Goal.ID, policy.Sources.Goal.Revision)
		policy.Versions.Taste = fmt.Sprintf("%s@%d", policy.Sources.Taste.ID, policy.Sources.Taste.Version)
		if err := f.access.SetPolicy(ctx, f.actor, policy, policy.Revision-1); err != nil {
			t.Fatal(err)
		}
		rule.Revision++
		rule.PolicyRevision = policy.Revision
		if err := delegation.SetRule(ctx, f.actor, rule, rule.Revision-1); err != nil {
			t.Fatal(err)
		}
		checkCurrent()
		bindCurrent()
	}
	updateTaste := func() {
		t.Helper()
		meta := commandMeta()
		meta.ExpectedVersion = taste.Version
		updated, err := f.attention.UpdateTasteProfile(ctx, principal, attention.UpdateTasteProfileCommand{CommandMeta: meta, Enabled: true, Rubrics: taste.Rubrics, Examples: taste.Examples, Delegations: taste.Delegations})
		if err != nil || updated.Profile == nil {
			t.Fatal(err)
		}
		taste = *updated.Profile
	}
	checkCurrent()
	bindCurrent()
	goal.Title = "Verify the fixed official export parity and local CPU contract"
	goal, err = f.access.Goals.(*workspace.S2Service).PutGoal(ctx, f.caller, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: goal.Revision}, Goal: goal})
	if err != nil {
		t.Fatal(err)
	}
	checkStale(apierrors.ContextStale)
	t.Log("actual Goal change rejected the old cache and admission before policy changed")
	policy.Sources.Goal.Revision = goal.Revision
	refresh()
	taste.Rubrics[0].Criteria = "Check the presence of each actual fixed official export section before recommending review"
	updateTaste()
	checkStale(apierrors.VersionConflict)
	t.Log("actual rubric change rejected the old cache and admission before policy changed")
	policy.Sources.Taste.Version = taste.Version
	refresh()
	taste.Rubrics = append(taste.Rubrics,
		attention.TasteRubric{ID: "official-export-format", QuestionType: "choice", Criteria: "Which model format does the official export section produce?", Options: []attention.TasteOption{{ID: "onnx", Description: "Split encoder and head ONNX files"}, {ID: "safetensors", Description: "Only the source safetensors checkpoint"}}, Levels: []string{}},
		attention.TasteRubric{ID: "official-export-verification", QuestionType: "score", Criteria: "How explicitly does the official export section describe numerical verification?", Options: []attention.TasteOption{}, Levels: []string{"No verification described", "Verification mentioned", "Torch and ONNX agreement and tolerance described"}})
	taste.Examples = append(taste.Examples,
		attention.TasteExample{ID: "official-format-review", RubricID: "official-export-format", MaterialID: ref.MaterialID, Revision: ref.Revision, ContentDigest: ref.Digest, Expected: "onnx"},
		attention.TasteExample{ID: "official-verification-review", RubricID: "official-export-verification", MaterialID: ref.MaterialID, Revision: ref.Revision, ContentDigest: ref.Digest, Expected: "2"})
	updateTaste()
	checkStale(apierrors.VersionConflict)
	policy.Sources.Taste.Version = taste.Version
	refresh()
	taste.Rubrics[1].Options[0].Description = "The documented encoder.onnx and head.onnx output files"
	updateTaste()
	checkStale(apierrors.VersionConflict)
	t.Log("actual finite choice option change rejected the old cache and admission before policy changed")
	policy.Sources.Taste.Version = taste.Version
	refresh()
	taste.Rubrics[2].Levels[2] = "Torch and ONNX agreement verified within the documented 1e-4 tolerance"
	updateTaste()
	checkStale(apierrors.VersionConflict)
	t.Log("actual ordered score level change rejected the old cache and admission before policy changed")
	policy.Sources.Taste.Version = taste.Version
	refresh()
	apiText, err := os.ReadFile(filepath.Join(s5Root(t), ".verification", "s5", "upstream", "jev-api.md"))
	if err != nil {
		t.Fatal(err)
	}
	example := f.importPublicText(t, string(apiText), "https://docs.typesafe.ai/api.md", "Official Jev API reference", []string{"Complete typed evaluation API reference"})
	space, err := f.attention.GetProjectSpace(ctx, principal, project.SpaceID)
	if err != nil {
		t.Fatal(err)
	}
	meta := commandMeta()
	meta.ExpectedVersion = space.Space.Version
	if _, err := f.attention.ReferenceMaterial(ctx, principal, project.SpaceID, attention.ReferenceMaterialCommand{CommandMeta: meta, MaterialID: example.MaterialID, Revision: example.Revision}); err != nil {
		t.Fatal(err)
	}
	taste.Examples[0].MaterialID, taste.Examples[0].Revision, taste.Examples[0].ContentDigest = example.MaterialID, example.Revision, example.Digest
	updateTaste()
	checkStale(apierrors.VersionConflict)
	t.Log("actual fixed example change rejected the old cache and admission before policy changed")
	policy.Sources.Taste.Version = taste.Version
	refresh()
	sourceRule := rule
	sourceRule.ID, sourceRule.WorkflowID, sourceRule.Revision, sourceRule.State = uuid.NewString(), uuid.NewString(), 1, "disabled"
	if err := delegation.SetRule(ctx, f.actor, sourceRule, 0); err != nil {
		t.Fatal(err)
	}
	taste.Delegations = []attention.TasteDelegation{{RuleID: sourceRule.ID, WorkflowID: sourceRule.WorkflowID, Revision: sourceRule.Revision}}
	updateTaste()
	checkStale(apierrors.VersionConflict)
	t.Log("actual taste delegation selection change rejected the old cache and admission before policy changed")
	policy.Sources.Taste.Version = taste.Version
	refresh()
	sourceRule, err = delegation.SetState(ctx, f.actor, sourceRule.ID, sourceRule.Revision, "paused")
	if err != nil {
		t.Fatal(err)
	}
	checkStale(apierrors.VersionConflict)
	t.Log("actual referenced delegation state/revision change rejected the old cache and admission before policy changed")
	taste.Delegations[0].Revision = sourceRule.Revision
	updateTaste()
	policy.Sources.Taste.Version = taste.Version
	refresh()
	meta = commandMeta()
	meta.ExpectedVersion = ranking.Profile.Version
	ranking, err = f.attention.UpdateRankingProfile(ctx, principal, attention.UpdateRankingProfileCommand{CommandMeta: meta, Enabled: true, Weights: map[string]float64{"goal_progress": 8, "current_interest": 2, "project_improvement": 2, "originality": 1}})
	if err != nil {
		t.Fatal(err)
	}
	checkStale(apierrors.VersionConflict)
	t.Log("actual ranking weight change rejected the old cache and admission before policy changed")
	policy.Sources.Ranking.Version = ranking.Profile.Version
	refresh()
	material, err := f.attention.GetMaterial(ctx, principal, example.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	meta = commandMeta()
	meta.ExpectedVersion = material.Material.Version
	if _, err := f.attention.UpdateMaterial(ctx, principal, example.MaterialID, attention.UpdateMaterialCommand{CommandMeta: meta, Lifecycle: "withdrawn"}); err != nil {
		t.Fatal(err)
	}
	_, err = core.Evaluate(ctx, f.actor, request)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, bound.ID, project.ID, opportunity.ID, opportunity.Revision)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	_, err = core.Read(ctx, f.actor, previous.ID)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	_, err = core.Feedback(ctx, f.actor, previous.ID, previous.Version, "adopt", "Verify current access to the actual example source")
	requireServiceCode(t, err, apierrors.ScopeDenied)
	t.Logf("actual Goal revision=%d, complete Taste versions=%d, ranking version=%d, source delegation revision=%d; every independent source change rejected old cache/proof before policy changed", goal.Revision, taste.Version, ranking.Profile.Version, sourceRule.Revision)
}

func TestActualAutomaticSourcesAndExternalPermission(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	question := judgment.Question{ID: "official-export-availability", Type: judgment.Truth, Instructions: "Is the fixed official export documentation present?", Rule: "all_evidence_present"}
	policy := f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1")
	if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
		t.Fatal(err)
	}
	delegation, err := judgment.NewDelegationService(f.db, actualService(t, f), f.access, adapter.AttentionOpportunities{Attention: f.attention})
	if err != nil {
		t.Fatal(err)
	}
	accepted := 1.0
	rule := judgment.DelegationRule{ID: uuid.NewString(), ProjectID: project.ID, WorkflowID: uuid.NewString(), OwnerID: f.actor.ID, Revision: 1, State: "active", Purpose: judgment.AdmissionPurpose, Question: question, PolicyRevision: policy.Revision, Language: "en", Materials: []judgment.EvidenceRef{ref}, AdmissionValueMin: &accepted, AdmissionValueMax: &accepted, ReviewConditions: []string{"fixed_inputs"}}
	err = delegation.SetRule(ctx, f.actor, rule, 0)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	policy, _, _ = actualSourcePolicy(t, f, policy, question, ref, "1")
	policy.Revision++
	policy.Backend, policy.Versions.Model, policy.ExternalScope = judgment.JevBackend, adapter.JevModel, "fixed-public-official-documents-only"
	err = f.access.SetPolicy(ctx, f.actor, policy, 1)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	policy.ExternalSources = policy.Sources
	err = f.access.SetPolicy(ctx, f.actor, policy, 1)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	policy.ExternalEvidence = []judgment.EvidenceRef{ref}
	if err := f.access.SetPolicy(ctx, f.actor, policy, 1); err != nil {
		t.Fatal("the exact public Goal/taste/example external permission could not be saved", err)
	}
	t.Log("actual automatic use without Goal/taste denied; cloud source and fixed example permissions independently required; no provider/model was invoked")
}
