package s5

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	modernsqlite "modernc.org/sqlite"
	sqlitecode "modernc.org/sqlite/lib"
)

func actualRunDirectory(t *testing.T) string {
	t.Helper()
	liveEnabled(t)
	base := os.Getenv("ASTROCYTE_S5_RUN_ROOT")
	if !filepath.IsAbs(base) {
		t.Fatal("ASTROCYTE_S5_RUN_ROOT must be an explicit absolute owned evidence directory")
	}
	root, err := os.MkdirTemp(base, "service-")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func actualService(t *testing.T, f *fixture, models ...judgment.Model) *judgment.Service {
	t.Helper()
	service, err := judgment.NewService(f.db, f.access, models...)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func requestForPolicy(policy judgment.Policy, question judgment.Question, refs ...judgment.EvidenceRef) judgment.JudgmentRequest {
	return judgment.JudgmentRequest{SchemaVersion: "1.0", ProjectID: policy.ProjectID, WorktreeID: policy.WorktreeID, SnapshotID: policy.SnapshotID, Question: question, EvidenceRefs: refs, Language: policy.Languages[0], Backend: policy.Backend, ExternalScope: policy.ExternalScope, Versions: policy.Versions, Budget: policy.Budget, Sources: policy.Sources}
}

func requireServiceCode(t *testing.T, err error, codes ...apierrors.Code) {
	t.Helper()
	var serviceError *apierrors.ServiceError
	if !errors.As(err, &serviceError) {
		t.Fatalf("expected a production structured error, got %v", err)
	}
	for _, code := range codes {
		if serviceError.Code == code {
			return
		}
	}
	t.Fatalf("actual error code %s differs from %v", serviceError.Code, codes)
}

func TestActualRulesVersionsAndFeedbackRestart(t *testing.T) {
	root := actualRunDirectory(t)
	path, objects := filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	var original judgment.JudgmentResult
	var request judgment.JudgmentRequest
	var human judgment.Actor
	var feedbackIDs []string
	if !t.Run("write_and_versions", func(t *testing.T) {
		f := newFixture(t, path, objects)
		human = f.actor
		ref := f.importOfficialReadme(t)
		project := f.project(t, ref, "astro-a")
		service := actualService(t, f)
		policy := f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1")
		if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
			t.Fatal(err)
		}
		question := judgment.Question{ID: "public-export-evidence", Type: judgment.Truth, Instructions: "Is the selected official Laya export documentation available?", Rule: "all_evidence_present"}
		request = requestForPolicy(policy, question, ref)
		var err error
		original, err = service.Evaluate(ctx, f.actor, request)
		if err != nil {
			t.Fatal(err)
		}
		if original.Decision != judgment.Recommend || original.Value == nil || *original.Value != 1 || original.ProviderConfidence != nil || original.Usage != nil || original.Backend != judgment.RuleBackend || original.CacheHit {
			t.Fatalf("actual deterministic judgment is invalid: %+v", original)
		}
		cached, err := service.Evaluate(ctx, f.actor, request)
		if err != nil || cached.ID != original.ID || !cached.CacheHit {
			t.Fatal("actual SQLite cache was not reused", err)
		}
		for _, kind := range []string{"adopt", "later", "reject", "revise", "already_solved"} {
			feedback, err := service.Feedback(ctx, f.actor, original.ID, original.Version, kind, "S5 public documentation feedback verification: "+kind)
			if err != nil {
				t.Fatal(err)
			}
			if feedback.ActorID != f.actor.ID || feedback.ActorKind != "human" || feedback.ResultVersion != original.Version || feedback.Kind != kind {
				t.Fatal("feedback differs from the actual authenticated human and fixed result")
			}
			feedbackIDs = append(feedbackIDs, feedback.ID)
		}
		_, err = service.Feedback(ctx, f.actor, original.ID, original.Version+1, "adopt", "Verify exact persisted result version")
		requireServiceCode(t, err, apierrors.VersionConflict)
		for _, change := range []struct {
			name string
			set  func(*judgment.Versions)
		}{
			{"goal", func(v *judgment.Versions) { v.Goal += "/updated" }},
			{"taste", func(v *judgment.Versions) { v.Taste += "/updated" }},
			{"policy", func(v *judgment.Versions) { v.Policy += "/updated" }},
			{"model", func(v *judgment.Versions) { v.Model += "/updated" }},
			{"access", func(v *judgment.Versions) { v.AccessEpoch += "/updated" }},
		} {
			t.Run(change.name, func(t *testing.T) {
				previous := policy.Revision
				policy.Revision++
				change.set(&policy.Versions)
				if err := f.access.SetPolicy(ctx, f.actor, policy, previous); err != nil {
					t.Fatal(err)
				}
				changed, err := service.Evaluate(ctx, f.actor, requestForPolicy(policy, question, ref))
				if err != nil || changed.ID == original.ID || changed.CacheHit {
					t.Fatal("changed actual policy reused an old judgment", err)
				}
				_, err = service.Evaluate(ctx, f.actor, request)
				requireServiceCode(t, err, apierrors.VersionConflict)
			})
		}
	}) {
		return
	}
	t.Run("reopen_actual_sqlite", func(t *testing.T) {
		f := newFixture(t, path, objects)
		service := actualService(t, f)
		stored, err := service.Read(ctx, f.actor, original.ID)
		if err != nil || stored.ID != original.ID || stored.InputDigest != original.InputDigest || stored.Value == nil || *stored.Value != 1 {
			t.Fatal("actual SQLite restart lost the fixed judgment", err)
		}
		history, err := service.FeedbackHistory(ctx, f.actor, original.ID)
		if err != nil || len(history) != len(feedbackIDs) {
			t.Fatal("actual SQLite restart lost feedback", err)
		}
		for i, row := range history {
			if row.ID != feedbackIDs[i] || row.ActorID != human.ID || row.ActorKind != "human" || row.ResultVersion != original.Version {
				t.Fatal("reopened feedback changed its original actor or fixed version")
			}
		}
		_, err = service.Evaluate(ctx, f.actor, request)
		requireServiceCode(t, err, apierrors.VersionConflict)
	})
}

func TestActualGrantRevocationAndAtomicRollback(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	policy := f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1")
	if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
		t.Fatal(err)
	}
	service := actualService(t, f)
	request := requestForPolicy(policy, judgment.Question{ID: "grant-bound-evidence", Type: judgment.Truth, Instructions: "Is the fixed public export document available?", Rule: "all_evidence_present"}, ref)
	grant, err := f.projects.GrantProjectAgent(ctx, f.caller, project.ID, domain.ProjectGrant{AgentID: "s5-public-verifier", Actions: []string{"read_context"}})
	if err != nil {
		t.Fatal(err)
	}
	token, err := f.projects.IssueProjectAgentToken(ctx, f.caller, project.ID, grant.AgentID)
	if err != nil {
		t.Fatal(err)
	}
	caller, err := f.projects.(workspace.ScopedAgentTokens).AuthenticateAgentToken(ctx, token.Token)
	if err != nil {
		t.Fatal(err)
	}
	actor := judgment.Actor{ID: caller.ID, Kind: caller.Kind, ProjectID: caller.ProjectID}
	result, err := service.Evaluate(ctx, actor, request)
	if err != nil {
		t.Fatal(err)
	}
	cached, err := service.Evaluate(ctx, actor, request)
	if err != nil || !cached.CacheHit || cached.ID != result.ID {
		t.Fatal("actual granted Agent cache was not reused", err)
	}
	_, err = service.Feedback(ctx, actor, result.ID, result.Version, "adopt", "Agent cannot provide human feedback")
	requireServiceCode(t, err, apierrors.ScopeDenied)
	if _, err := f.projects.RevokeProjectAgent(ctx, f.caller, project.ID, caller.ID); err != nil {
		t.Fatal(err)
	}
	_, err = service.Evaluate(ctx, actor, request)
	requireServiceCode(t, err, apierrors.ScopeDenied, apierrors.ApprovalRevoked)
	_, err = service.Read(ctx, actor, result.ID)
	requireServiceCode(t, err, apierrors.ScopeDenied, apierrors.ApprovalRevoked)
	_, err = f.projects.(workspace.ScopedAgentTokens).AuthenticateAgentToken(ctx, token.Token)
	requireServiceCode(t, err, apierrors.ApprovalRevoked)

	request.Question.ID = "atomic-public-evidence"
	before := f.db.Conn().Stats().WaitCount
	var rolledBack judgment.JudgmentResult
	err = sqlite.NewWorkspaceRepository(f.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		var err error
		rolledBack, err = service.Evaluate(tx.Context(), f.actor, request)
		if err != nil {
			return err
		}
		record, err := f.db.LoadResult(tx.Context(), rolledBack.ID)
		if err != nil {
			return err
		}
		return f.db.SaveResult(tx.Context(), record)
	})
	var constraint *modernsqlite.Error
	if !errors.As(err, &constraint) || constraint.Code() != sqlitecode.SQLITE_CONSTRAINT_PRIMARYKEY {
		t.Fatal("the real duplicate INSERT did not fail with its SQLite primary key constraint", err)
	}
	if f.db.Conn().Stats().WaitCount != before {
		t.Fatal("nested judgment borrowed another database connection")
	}
	_, err = service.Read(ctx, f.actor, rolledBack.ID)
	requireServiceCode(t, err, apierrors.NotFound)
}

func TestActualPublicWorktreeIsolationAndUncommittedChange(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	ref := f.importOfficialReadme(t)
	service := actualService(t, f)
	var requests []judgment.JudgmentRequest
	var results []judgment.JudgmentResult
	for _, checkout := range []string{"astro-a", "astro-b"} {
		project := f.project(t, ref, checkout)
		policy := f.policy(t, project, judgment.Truth, judgment.RuleBackend, "rule:all_evidence_present/v1")
		if err := f.access.SetPolicy(ctx, f.actor, policy, 0); err != nil {
			t.Fatal(err)
		}
		request := requestForPolicy(policy, judgment.Question{ID: "worktree-public-evidence", Type: judgment.Truth, Instructions: "Is the selected official export documentation present?", Rule: "all_evidence_present"}, ref)
		result, err := service.Evaluate(ctx, f.actor, request)
		if err != nil {
			t.Fatal(err)
		}
		requests, results = append(requests, request), append(results, result)
	}
	if requests[0].WorktreeID == requests[1].WorktreeID || results[0].ID == results[1].ID || results[1].CacheHit {
		t.Fatal("actual public Git worktrees shared a snapshot or cached judgment")
	}
	otherProject := f.actor
	otherProject.ProjectID = requests[1].ProjectID
	_, err := service.Read(ctx, otherProject, results[0].ID)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	readme := filepath.Join(s5Root(t), ".verification", "s5", "projects", "astro-a", "README.md")
	original, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	publicDocumentation, err := os.ReadFile(filepath.Join(s5Root(t), ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(readme, original, 0600); err != nil {
			t.Error(err)
		}
	})
	if err := os.WriteFile(readme, append(append([]byte{}, original...), publicDocumentation...), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = service.Evaluate(ctx, f.actor, requests[0])
	requireServiceCode(t, err, apierrors.ContextStale)
	unchanged, err := service.Evaluate(ctx, f.actor, requests[1])
	if err != nil || !unchanged.CacheHit || unchanged.ID != results[1].ID {
		t.Fatal("an unrelated public worktree lost its unchanged actual cache", err)
	}
}
