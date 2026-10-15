package s2_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func workspaceMeta(version int) workspace.WorkspaceCommandMeta {
	return workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
}

func requireCode(t *testing.T, err error, code apierrors.Code) {
	t.Helper()
	var e *apierrors.ServiceError
	if !errors.As(err, &e) || e.Code != code {
		t.Fatalf("wanted %s, got %v", code, err)
	}
}

func TestActualSQLiteApprovalRestartAndRollback(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".verification", "s2-first")
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(base, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("local database: %s", filepath.Join(dir, "state.sqlite"))
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: uuid.NewString()}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public S2 implementation validation"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Astrocyte public S2 validation", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	settings.AllowedActions = []string{"read_context"}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: 1, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewWorkspaceRepository(db)
	service := workspace.NewS2Service(repo, agents.ProjectFiles{}, nil)
	goal, err := service.PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "Review the public S2 implementation", Status: "active", SuccessCriteria: []string{"Approval and revocation persist across restart"}}})
	if err != nil {
		t.Fatal(err)
	}
	goalSnapshot, err := service.ReadCurrentGoalSnapshot(ctx, human, project.ID, goal.ID, goal.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if goalSnapshot.ObservedAt.IsZero() || goalSnapshot.ProjectSettingsRevision != project.Settings.Revision || goalSnapshot.Goal.Title != goal.Title {
		t.Fatal("actual goal snapshot omitted its current target or permission revision")
	}
	err = service.WithCurrentGoalSnapshot(ctx, human, goalSnapshot, func(txctx context.Context, current domain.CurrentGoalSnapshot) error {
		conn, err := sqlite.WorkspaceTransaction(txctx, db)
		if err != nil {
			return err
		}
		if conn == nil || current.Goal.ID != goal.ID || current.Goal.Revision != goal.Revision {
			t.Fatal("actual goal snapshot lost its same-database transaction")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	unobserved := goalSnapshot
	unobserved.ObservedAt = time.Time{}
	err = service.WithCurrentGoalSnapshot(ctx, human, unobserved, func(context.Context, domain.CurrentGoalSnapshot) error {
		t.Fatal("unobserved goal callback executed")
		return nil
	})
	requireCode(t, err, apierrors.ValidationFailed)
	calls := int64(2)
	content := domain.ProposalContent{Title: "Verify public workspace approval behavior", Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Scope: domain.ExecutionScope{ProjectID: project.ID, Roots: []string{root}, Actions: []string{"read_context"}, MaxConcurrency: 2}, Budget: domain.ResourceBudget{Calls: &calls}, Stops: domain.StopConditions{Deadline: time.Now().UTC().Add(time.Hour), Conditions: []string{"Stop when checks complete"}, MaxWorkItems: 2, MaxAttempts: 2, MaxNoProgressRounds: 2}, Deliverables: []string{"Actual SQLite approval receipts"}, SuccessCriteria: []string{"No execution is started by approval"}}
	create := workspace.CreateProposalCommand{WorkspaceCommandMeta: workspaceMeta(0), Content: content}
	proposal, err := service.CreateProposal(ctx, human, create)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := service.CreateProposal(ctx, human, create)
	if err != nil || replayed.ID != proposal.ID {
		t.Fatalf("create replay: %v", err)
	}
	newTrace := create
	newTrace.RequestID = uuid.NewString()
	replayed, err = service.CreateProposal(ctx, human, newTrace)
	if err != nil || replayed.ID != proposal.ID {
		t.Fatalf("same command under new request trace: %v", err)
	}
	different := create
	different.Content.Title = "Review revised public approval behavior"
	_, err = service.CreateProposal(ctx, human, different)
	requireCode(t, err, apierrors.VersionConflict)
	proposal, err = service.TransitionProposal(ctx, human, proposal.ID, workspaceMeta(proposal.Version), domain.ProposalInReview)
	if err != nil {
		t.Fatal(err)
	}
	approve := workspace.ApproveProposalCommand{WorkspaceCommandMeta: workspaceMeta(proposal.Version), ExpectedRevision: proposal.Revision, ExpiresAt: content.Stops.Deadline}
	_, err = service.ApproveProposal(ctx, domain.Caller{Kind: "agent", ID: uuid.NewString(), ProjectID: project.ID}, proposal.ID, approve)
	requireCode(t, err, apierrors.ScopeDenied)
	stale := approve
	stale.WorkspaceCommandMeta = workspaceMeta(proposal.Version - 1)
	_, err = service.ApproveProposal(ctx, human, proposal.ID, stale)
	requireCode(t, err, apierrors.VersionConflict)
	grant, err := service.ApproveProposal(ctx, human, proposal.ID, approve)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := service.ApproveProposal(ctx, human, proposal.ID, approve)
	if err != nil || replay.ID != grant.ID {
		t.Fatalf("approve replay: %v", err)
	}
	events, err := repo.ReadWorkspaceEvents(ctx, "", 100)
	if err != nil || len(events) != 1 {
		t.Fatalf("approval outbox: count=%d err=%v", len(events), err)
	}
	redelivered, err := repo.ReadWorkspaceEvents(ctx, "", 100)
	if err != nil || redelivered[0].ID != events[0].ID {
		t.Fatalf("event redelivery: %v", err)
	}
	used := int64(1)
	check := domain.GrantCheck{Caller: human, GrantID: grant.ID, Epoch: grant.Epoch, ProposalRevision: grant.ProposalRevision, ProjectID: project.ID, Root: root, Action: "read_context", Concurrency: 1, Used: domain.ResourceBudget{Calls: &used}}
	if err = service.WithAuthorizedGrant(ctx, check, func(txctx context.Context, g domain.ApprovalGrant) error {
		_, err := sqlite.WorkspaceTransaction(txctx, db)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	otherSpace, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Separate public goal access"})
	if err != nil {
		t.Fatal(err)
	}
	otherProject, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Separate public goal review", Root: filepath.Join(root, "internal", "workspace", "domain"), SpaceID: otherSpace.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	otherMemberID := uuid.NewString()
	if _, err = projects.GrantProjectAgent(ctx, human, otherProject.ID, domain.ProjectGrant{AgentID: otherMemberID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	otherCredential, err := projects.IssueProjectAgentToken(ctx, human, otherProject.ID, otherMemberID)
	if err != nil {
		t.Fatal(err)
	}
	otherMember, err := projects.AuthenticateAgentToken(ctx, otherCredential.Token)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.ReadCurrentGoalSnapshot(ctx, otherMember, project.ID, goal.ID, goal.Revision)
	requireCode(t, err, apierrors.ScopeDenied)
	updatedGoal := goal
	updatedGoal.Title = "Review the current public approval implementation"
	updatedGoal, err = service.PutGoal(ctx, human, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(goal.Revision), Goal: updatedGoal})
	if err != nil {
		t.Fatal(err)
	}
	oldGoal, err := service.ReadGoal(ctx, human, goal.ID, goal.Revision)
	if err != nil || oldGoal.Title != goal.Title || oldGoal.Revision != goal.Revision || updatedGoal.Revision != goal.Revision+1 {
		t.Fatalf("goal revisions changed historical target: %v", err)
	}
	err = service.WithCurrentGoalSnapshot(ctx, human, goalSnapshot, func(context.Context, domain.CurrentGoalSnapshot) error {
		t.Fatal("stale goal snapshot callback executed")
		return nil
	})
	requireCode(t, err, apierrors.ContextStale)
	currentGoalSnapshot, err := service.ReadCurrentGoalSnapshot(ctx, human, project.ID, goal.ID, updatedGoal.Revision)
	if err != nil {
		t.Fatal(err)
	}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: project.Settings})
	if err != nil {
		t.Fatal(err)
	}
	err = service.WithCurrentGoalSnapshot(ctx, human, currentGoalSnapshot, func(context.Context, domain.CurrentGoalSnapshot) error {
		t.Fatal("changed project permission callback executed")
		return nil
	})
	requireCode(t, err, apierrors.ContextStale)
	before := goal
	before.ID = uuid.NewString()
	before.Revision = 1
	err = repo.WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		if err := tx.SaveGoal(before, 0); err != nil {
			return err
		}
		duplicate := grant
		duplicate.ID = uuid.NewString()
		return tx.SaveApproval(duplicate, 0)
	})
	if err == nil {
		t.Fatal("duplicate proposal approval did not fail actual SQLite uniqueness")
	}
	_, err = service.ReadGoal(ctx, human, before.ID, 0)
	requireCode(t, err, apierrors.NotFound)
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo = sqlite.NewWorkspaceRepository(db)
	service = workspace.NewS2Service(repo, agents.ProjectFiles{}, nil)
	persisted, err := service.ReadApproval(ctx, human, grant.ID)
	if err != nil || persisted.Epoch != 1 || persisted.ProposalRevision != 1 {
		t.Fatalf("restart approval: %v", err)
	}
	oldGoal, err = service.ReadGoal(ctx, human, goal.ID, goal.Revision)
	if err != nil || oldGoal.Title != goal.Title || oldGoal.Revision != goal.Revision {
		t.Fatalf("restart historical goal: %v", err)
	}
	_, err = service.ChangeApproval(ctx, human, grant.ID, workspace.ChangeApprovalCommand{WorkspaceCommandMeta: workspaceMeta(grant.Epoch), Reason: "End actual approval validation"}, "revoked")
	if err != nil {
		t.Fatal(err)
	}
	err = service.WithAuthorizedGrant(ctx, check, func(context.Context, domain.ApprovalGrant) error { t.Fatal("revoked callback executed"); return nil })
	requireCode(t, err, apierrors.ApprovalRevoked)
	old, err := service.ReadProposal(ctx, human, proposal.ID, 1)
	if err != nil || old.Status != domain.ProposalApproved {
		t.Fatalf("fixed approved proposal changed: %v", err)
	}
}
