package s5

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/admissionbridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func actualOpportunity(t *testing.T, f *fixture, ref judgment.EvidenceRef, title, question string) attention.Opportunity {
	t.Helper()
	ctx := context.Background()
	principal := attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}
	content, err := f.attention.GetContent(ctx, principal, ref.MaterialID, ref.Revision)
	if err != nil {
		t.Fatal(err)
	}
	locator := "https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md"
	refs := []attention.SourceRef{{MaterialID: ref.MaterialID, Revision: ref.Revision, Locator: locator}}
	distillation, err := f.attention.RecordDistillation(ctx, principal, attention.RecordDistillationCommand{CommandMeta: commandMeta(), InputRefs: refs, Stage: "topic", OutputText: content.Text, Question: question, PendingQuestions: []string{question}, ProcessingConfig: "manual-public-document-review"})
	if err != nil {
		t.Fatal(err)
	}
	opportunity, err := f.attention.CreateOpportunity(ctx, principal, attention.OpportunityCommand{CommandMeta: commandMeta(), Title: title, EvidenceRefs: refs, NextStep: "Read the fixed official CPU export section and verify the documented behavior", Purpose: question, DistillationIDs: []string{distillation.Distillation.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if opportunity.Opportunity.State != "ready_for_review" {
		t.Fatal("the actual human public documentation record did not establish a ready opportunity")
	}
	return opportunity.Opportunity
}

func TestActualDelegationBindingsAndAdmission(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	policy := f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1")
	question := judgment.Question{ID: "fixed-source-availability", Type: judgment.Truth, Instructions: "Are all selected fixed official source sections present?", Rule: "all_evidence_present"}
	policy, _, _ = actualSourcePolicy(t, f, policy, question, ref, "1")
	if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
		t.Fatal(err)
	}
	core := actualService(t, f)
	delegation, err := judgment.NewDelegationService(f.db, core, f.access, adapter.AttentionOpportunities{Attention: f.attention})
	if err != nil {
		t.Fatal(err)
	}
	first := actualOpportunity(t, f, ref, "Verify official ONNX export parity", "Which numerical tolerance does the official export verify?")
	second := actualOpportunity(t, f, ref, "Verify official local CPU inference", "Which local runtime loads the exported CPU graphs?")
	accepted := 1.0
	rule := judgment.DelegationRule{ID: uuid.NewString(), ProjectID: project.ID, WorkflowID: uuid.NewString(), OwnerID: f.actor.ID, Revision: 1, State: "disabled", Purpose: judgment.AdmissionPurpose, Question: question, PolicyRevision: policy.Revision, Language: "en", Materials: []judgment.EvidenceRef{ref}, AdmissionValueMin: &accepted, AdmissionValueMax: &accepted, ReviewConditions: []string{"fixed_inputs"}}
	if err := delegation.SetRule(ctx, f.actor, rule, 0); err != nil {
		t.Fatal(err)
	}
	_, err = delegation.Evaluate(ctx, f.actor, rule.ID, 1, first.ID, first.Revision)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	rule, err = delegation.SetState(ctx, f.actor, rule.ID, 1, "active")
	if err != nil {
		t.Fatal(err)
	}
	result, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, first.ID, first.Revision)
	if err != nil || result.Context == nil || result.Context.Opportunity.ID != first.ID || result.Decision != judgment.Recommend {
		t.Fatal("actual configured delegation did not preserve its fixed opportunity", err)
	}
	cached, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, first.ID, first.Revision)
	if err != nil || !cached.CacheHit || cached.ID != result.ID {
		t.Fatal("the same fixed workflow did not reuse its authorized actual cache", err)
	}
	other, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, second.ID, second.Revision)
	if err != nil || other.ID == result.ID || other.CacheHit {
		t.Fatal("two actual opportunities shared one immutable binding", err)
	}
	meta := commandMeta()
	meta.ExpectedVersion = second.Version
	if _, err := f.attention.ReviseOpportunity(ctx, attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, second.ID, attention.OpportunityCommand{CommandMeta: meta, Title: "Verify official local CPU checkpoint loading", EvidenceRefs: second.EvidenceRefs, NextStep: second.NextStep, Purpose: second.Purpose, DistillationIDs: second.DistillationIDs}); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, other.ID, project.ID, second.ID, second.Revision)
	requireServiceCode(t, err, apierrors.VersionConflict)
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, result.ID, project.ID, second.ID, second.Revision)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	secondRule := rule
	secondRule.ID, secondRule.WorkflowID, secondRule.Revision = uuid.NewString(), uuid.NewString(), 1
	if err := delegation.SetRule(ctx, f.actor, secondRule, 0); err != nil {
		t.Fatal(err)
	}
	otherWorkflow, err := delegation.Evaluate(ctx, f.actor, secondRule.ID, secondRule.Revision, first.ID, first.Revision)
	if err != nil || otherWorkflow.ID == result.ID || otherWorkflow.CacheHit {
		t.Fatal("two actual workflows reused the wrong immutable binding", err)
	}
	reviewRule := secondRule
	reviewRule.ID, reviewRule.WorkflowID = uuid.NewString(), uuid.NewString()
	reviewRule.ReviewConditions = []string{"fixed_inputs", "human_adopt"}
	if err := delegation.SetRule(ctx, f.actor, reviewRule, 0); err != nil {
		t.Fatal(err)
	}
	reviewed, err := delegation.Evaluate(ctx, f.actor, reviewRule.ID, reviewRule.Revision, first.ID, first.Revision)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = delegation.Authorize(ctx, reviewRule.WorkflowID, reviewed.ID, project.ID, first.ID, first.Revision)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	if _, err := core.Feedback(ctx, f.actor, reviewed.ID, reviewed.Version, "adopt", "Adopt this fixed public evidence check for the configured review condition"); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := delegation.Authorize(ctx, reviewRule.WorkflowID, reviewed.ID, project.ID, first.ID, first.Revision); err != nil {
		t.Fatal("the actual fixed-version human adoption was not used", err)
	}
	if _, err := core.Feedback(ctx, f.actor, reviewed.ID, reviewed.Version, "reject", "Withdraw the fixed-version adoption during the actual review check"); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = delegation.Authorize(ctx, reviewRule.WorkflowID, reviewed.ID, project.ID, first.ID, first.Revision)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	rule, err = delegation.SetState(ctx, f.actor, rule.ID, rule.Revision, "paused")
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = delegation.Authorize(ctx, rule.WorkflowID, result.ID, project.ID, first.ID, first.Revision)
	requireServiceCode(t, err, apierrors.ApprovalRevoked)
	rule, err = delegation.SetState(ctx, f.actor, rule.ID, rule.Revision, "active")
	if err != nil {
		t.Fatal(err)
	}
	current, err := delegation.Evaluate(ctx, f.actor, rule.ID, rule.Revision, first.ID, first.Revision)
	if err != nil || current.ID == result.ID || current.CacheHit {
		t.Fatal("a resumed new rule version reused its previous binding", err)
	}
	proposals := workspace.NewS2Service(sqlite.NewWorkspaceRepository(f.db), agents.ProjectFiles{}, adapter.AttentionReferences{Attention: f.attention})
	admissions, err := admissionbridge.Connect(f.attention, sqlite.NewAdmissionRepository(f.db), proposals, attention.AdmissionOptions{Projects: adapter.AdmissionProjects{Authorization: f.access, Rules: f.db}, Judgment: adapter.AdmissionAuthorizer{Delegation: delegation}})
	if err != nil {
		t.Fatal(err)
	}
	workflow := adapter.AdmissionWorkflow{Delegation: delegation, Admissions: admissions}
	command := judgment.DelegatedAdmissionCommand{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedOpportunityVersion: first.Version}
	receipt, err := workflow.AdmitJudgment(ctx, f.actor, current.ID, command)
	if err != nil || receipt.AdmissionID == "" || receipt.ProposalID == "" {
		t.Fatal("actual S1 transaction and S2 bridge did not persist the delegated admission and draft", err)
	}
	replayed, err := workflow.AdmitJudgment(ctx, f.actor, current.ID, command)
	if err != nil || replayed != receipt {
		t.Fatal("actual repeated admission lost its original idempotent receipt", err)
	}
	var authority, decision, purpose, questionType string
	if err := f.db.Conn().QueryRowContext(ctx, "SELECT json_extract(data,'$.authority'),json_extract(data,'$.decision_id'),json_extract(data,'$.purpose'),json_extract(data,'$.question_type') FROM attention_admissions WHERE id=?", receipt.AdmissionID).Scan(&authority, &decision, &purpose, &questionType); err != nil {
		t.Fatal(err)
	}
	if authority != "delegated_judgment" || decision != current.ID || purpose != judgment.AdmissionPurpose || questionType != "truth" {
		t.Fatal("actual persisted authority differs from the configured fixed judgment")
	}
	policy.Revision++
	policy.Versions.Policy += "/changed"
	if err := f.access.SetPolicy(ctx, f.actor, policy, policy.Revision-1); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = delegation.Authorize(ctx, secondRule.WorkflowID, otherWorkflow.ID, project.ID, first.ID, first.Revision)
	requireServiceCode(t, err, apierrors.VersionConflict)
	rule, err = delegation.SetState(ctx, f.actor, rule.ID, rule.Revision, "paused")
	if err != nil {
		t.Fatal("pause was blocked by an outdated policy", err)
	}
	rule, err = delegation.SetState(ctx, f.actor, rule.ID, rule.Revision, "revoked")
	if err != nil {
		t.Fatal(err)
	}
	_, err = delegation.SetState(ctx, f.actor, rule.ID, rule.Revision, "active")
	requireServiceCode(t, err, apierrors.VersionConflict)
	t.Logf("actual delegated admission=%s, draft=%s, fixed decision=%s", receipt.AdmissionID, receipt.ProposalID, current.ID)
}
