package s5

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestActualGoalAndRankingSources(t *testing.T) {
	root := actualRunDirectory(t)
	f := newFixture(t, filepath.Join(root, "state.sqlite"), filepath.Join(root, "objects"))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	ref := f.importOfficialReadme(t)
	project := f.project(t, ref, "astro-a")
	other := f.project(t, ref, "astro-b")
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(f.db), agents.ProjectFiles{}, adapter.AttentionReferences{Attention: f.attention})
	f.access.Goals, f.access.Ranking = s2, f.attention
	_, err := f.access.ReadGoal(ctx, f.actor, project.ID, domain.GoalReference{})
	requireServiceCode(t, err, apierrors.ValidationFailed)
	_, err = f.access.ReadRanking(ctx, f.actor, project.ID, "attention", 1)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	meta := func(version int) workspace.WorkspaceCommandMeta {
		return workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
	}
	goal, err := s2.PutGoal(ctx, f.caller, "", workspace.GoalCommand{WorkspaceCommandMeta: meta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "Verify the public Laya CPU export contract", SuccessCriteria: []string{"Read the fixed official export section and its parity conditions"}, Priority: 1, Status: "active"}})
	if err != nil {
		t.Fatal(err)
	}
	goalRef := domain.GoalReference{ID: goal.ID, Revision: goal.Revision}
	actual, err := f.access.ReadGoal(ctx, f.actor, project.ID, goalRef)
	if err != nil || actual.ID != goal.ID || actual.Title != goal.Title {
		t.Fatal("the actual selected Goal was not read", err)
	}
	_, err = f.access.ReadGoal(ctx, f.actor, other.ID, goalRef)
	requireServiceCode(t, err, apierrors.ScopeDenied)
	principal := attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}
	profile, err := f.attention.UpdateRankingProfile(ctx, principal, attention.UpdateRankingProfileCommand{CommandMeta: commandMeta(), Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}})
	if err != nil {
		t.Fatal(err)
	}
	waits := f.db.Conn().Stats().WaitCount
	err = f.db.WithJudgmentTx(ctx, func(ctx context.Context) error {
		if _, err := f.access.ReadGoal(ctx, f.actor, project.ID, goalRef); err != nil {
			return err
		}
		_, err := f.access.ReadRanking(ctx, f.actor, project.ID, profile.Profile.ID, profile.Profile.Version)
		return err
	})
	if err != nil || f.db.Conn().Stats().WaitCount != waits {
		t.Fatal("actual Goal and ranking sources did not reuse the original SQLite transaction", err)
	}
	goal.Title = "Verify the public Laya CPU export and local inference contracts"
	updated, err := s2.PutGoal(ctx, f.caller, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: meta(goal.Revision), Goal: goal})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.access.ReadGoal(ctx, f.actor, project.ID, goalRef)
	requireServiceCode(t, err, apierrors.VersionConflict)
	historical, err := s2.ReadGoal(ctx, f.caller, goal.ID, goal.Revision)
	if err != nil || historical.Title == updated.Title {
		t.Fatal("the original immutable Goal revision was changed", err)
	}
	_, err = f.access.ReadGoal(ctx, f.actor, project.ID, domain.GoalReference{ID: updated.ID, Revision: updated.Revision})
	if err != nil {
		t.Fatal(err)
	}
	command := attention.UpdateRankingProfileCommand{CommandMeta: commandMeta(), Enabled: true, Weights: map[string]float64{"goal_progress": 8, "current_interest": 2, "project_improvement": 2, "originality": 1}}
	command.ExpectedVersion = profile.Profile.Version
	changed, err := f.attention.UpdateRankingProfile(ctx, principal, command)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.access.ReadRanking(ctx, f.actor, project.ID, profile.Profile.ID, profile.Profile.Version)
	requireServiceCode(t, err, apierrors.VersionConflict)
	command.CommandMeta, command.Enabled = commandMeta(), false
	command.ExpectedVersion = changed.Profile.Version
	disabled, err := f.attention.UpdateRankingProfile(ctx, principal, command)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.access.ReadRanking(ctx, f.actor, project.ID, disabled.Profile.ID, disabled.Profile.Version)
	requireServiceCode(t, err, apierrors.ApprovalRequired)
	t.Logf("actual Goal revisions=%d/%d ranking versions=%d/%d/%d; immutable history and same transaction checked", goal.Revision, updated.Revision, profile.Profile.Version, changed.Profile.Version, disabled.Profile.Version)
}
