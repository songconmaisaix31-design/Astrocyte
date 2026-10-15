package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workdomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestCurrentGoalGuardSharesAttentionCommitRollbackAndReopen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	core, db, objects := fixture(t)
	space, err := core.CreateProjectSpace(ctx, human, ProjectSpaceCommand{CommandMeta: meta("goal-transaction-space", 1), Title: "Public Goal and Attention transaction validation"})
	if err != nil {
		t.Fatal("create actual project space")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal("resolve approved public project root")
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal("resolve actual public Git worktree identity")
	}
	caller := workdomain.Caller{ID: human.ID, Kind: human.Kind}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, caller, workdomain.RegisterProjectCommand{Name: "Public Goal transaction validation", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal("register actual approved public Git project")
	}
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	goalMeta := func(key string, version int) workspace.WorkspaceCommandMeta {
		return workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: key, IdempotencyKey: key, ExpectedVersion: version}
	}
	goal, err := s2.PutGoal(ctx, caller, "", workspace.GoalCommand{WorkspaceCommandMeta: goalMeta("goal-transaction-initial", 0), Goal: workdomain.Goal{ProjectID: project.ID, Title: "Check fixed Goal authorization before committing an Attention update", Status: "active", SuccessCriteria: []string{"The original Goal and the Attention update use one actual database transaction"}}})
	if err != nil {
		t.Fatal("save explicit human Goal")
	}
	snapshot, err := s2.ReadCurrentGoalSnapshot(ctx, caller, project.ID, goal.ID, goal.Revision)
	if err != nil {
		t.Fatal("read actual current Goal snapshot")
	}
	command := UpdateRankingProfileCommand{CommandMeta: meta("goal-guarded-ranking", 1), Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}}
	beforeEvents, beforeReceipts := countRows(t, db, "attention_outbox"), countRows(t, db, "attention_receipts")
	waits := db.Conn().Stats().WaitCount
	var saved RankingProfileDetail
	err = s2.WithCurrentGoalSnapshot(ctx, caller, snapshot, func(txctx context.Context, current workdomain.CurrentGoalSnapshot) error {
		conn, err := sqlite.WorkspaceTransaction(txctx, db)
		if err != nil {
			return err
		}
		if conn == nil || current.Goal.ID != goal.ID || current.Goal.Revision != goal.Revision {
			t.Fatal("Goal guard did not provide its actual transaction and fixed Goal")
		}
		saved, err = core.UpdateRankingProfile(txctx, human, command)
		if err != nil {
			return err
		}
		// 在同一真实事务内再次插入刚保存的版本，触发 SQLite 唯一约束。
		_, err = conn.ExecContext(txctx, "INSERT INTO attention_ranking_profile_revisions(profile_id,version,data) SELECT profile_id,version,data FROM attention_ranking_profile_revisions WHERE profile_id='attention'")
		return err
	})
	if err == nil || saved.Profile == nil || db.Conn().Stats().WaitCount != waits {
		t.Fatal("actual post-write SQL failure or original shared connection was not observed")
	}
	if countRows(t, db, "attention_ranking_profiles") != 0 || countRows(t, db, "attention_ranking_profile_revisions") != 0 || countRows(t, db, "attention_outbox") != beforeEvents || countRows(t, db, "attention_receipts") != beforeReceipts {
		t.Fatal("failed outer Goal transaction retained an Attention head, revision, event or receipt")
	}
	err = s2.WithCurrentGoalSnapshot(ctx, caller, snapshot, func(txctx context.Context, _ workdomain.CurrentGoalSnapshot) error {
		var err error
		saved, err = core.UpdateRankingProfile(txctx, human, command)
		return err
	})
	if err != nil || saved.Profile == nil || saved.Profile.Version != 1 || countRows(t, db, "attention_outbox") != beforeEvents+1 || countRows(t, db, "attention_receipts") != beforeReceipts+1 {
		t.Fatal("original rolled-back command could not commit once under the current Goal")
	}
	changed := goal
	changed.Title = "Check the current Goal and retained Attention transaction history"
	changed, err = s2.PutGoal(ctx, caller, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: goalMeta("goal-transaction-next", goal.Revision), Goal: changed})
	if err != nil {
		t.Fatal("save actual next human Goal revision")
	}
	entered := false
	callback := func(context.Context, workdomain.CurrentGoalSnapshot) error { entered = true; return nil }
	err = s2.WithCurrentGoalSnapshot(ctx, caller, snapshot, callback)
	errorCode(t, err, apierrors.ContextStale)
	if entered {
		t.Fatal("superseded Goal permitted a new Attention operation")
	}
	if _, err = projects.GrantProjectAgent(ctx, caller, project.ID, workdomain.ProjectGrant{AgentID: agent.ID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal("grant actual project read membership")
	}
	credential, err := projects.IssueProjectAgentToken(ctx, caller, project.ID, agent.ID)
	if err != nil {
		t.Fatal("issue actual scoped project credential")
	}
	member, err := projects.AuthenticateAgentToken(ctx, credential.Token)
	if err != nil {
		t.Fatal("authenticate actual scoped project credential")
	}
	memberSnapshot, err := s2.ReadCurrentGoalSnapshot(ctx, member, project.ID, changed.ID, changed.Revision)
	if err != nil {
		t.Fatal("read current Goal through actual project membership")
	}
	if _, err = projects.RevokeProjectAgent(ctx, caller, project.ID, agent.ID); err != nil {
		t.Fatal("revoke actual project membership")
	}
	err = s2.WithCurrentGoalSnapshot(ctx, member, memberSnapshot, callback)
	errorCode(t, err, apierrors.ApprovalRevoked)
	if entered || countRows(t, db, "attention_outbox") != beforeEvents+1 || countRows(t, db, "attention_receipts") != beforeReceipts+1 {
		t.Fatal("stale or revoked Goal access ran the consumer or changed Attention state")
	}
	var file string
	if err = db.Conn().QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&file); err != nil {
		t.Fatal("resolve Goal and Attention database for reopen")
	}
	if err = db.Close(); err != nil {
		t.Fatal("close original Goal and Attention database")
	}
	reopenedDB, err := sqlite.Open(file)
	if err != nil {
		t.Fatal("reopen original Goal and Attention database")
	}
	t.Cleanup(func() {
		if err := reopenedDB.Close(); err != nil {
			t.Error("close reopened Goal and Attention database")
		}
	})
	reopened := NewAttentionService(sqlite.NewAttentionRepository(reopenedDB), importers.NewReader(nil), objects, ServiceOptions{})
	s2 = workspace.NewS2Service(sqlite.NewWorkspaceRepository(reopenedDB), agents.ProjectFiles{}, nil)
	historical, err := s2.ReadGoal(ctx, caller, goal.ID, goal.Revision)
	if err != nil || !sameAdmissionJSON(t, historical, goal) {
		t.Fatal("fixed Goal history changed after Attention transaction and database reopen")
	}
	replay, err := reopened.UpdateRankingProfile(ctx, human, command)
	if err != nil || !sameAdmissionJSON(t, replay, saved) {
		t.Fatal("successful guarded Attention receipt changed after reopen")
	}
	loaded, err := reopened.GetRankingProfile(ctx, human)
	if err != nil || !sameAdmissionJSON(t, loaded, saved) {
		t.Fatal("guarded Attention head or history changed after reopen")
	}
	err = s2.WithCurrentGoalSnapshot(ctx, member, memberSnapshot, callback)
	errorCode(t, err, apierrors.ApprovalRevoked)
	if entered || countRows(t, reopenedDB, "attention_ranking_profile_revisions") != 1 || countRows(t, reopenedDB, "attention_outbox") != beforeEvents+1 || countRows(t, reopenedDB, "attention_receipts") != beforeReceipts+1 {
		t.Fatal("reopen or revoked snapshot changed original Attention history")
	}
	assertNoAdmissionExecution(t, reopenedDB)
}
