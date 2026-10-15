package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type CurrentGoalReader interface {
	ReadCurrentGoalSnapshot(context.Context, domain.Caller, string, string, int) (domain.CurrentGoalSnapshot, error)
	WithCurrentGoalSnapshot(context.Context, domain.Caller, domain.CurrentGoalSnapshot, func(context.Context, domain.CurrentGoalSnapshot) error) error
}

func currentGoalSnapshot(tx WorkspaceTx, c domain.Caller, projectID, id string, revision int) (snapshot domain.CurrentGoalSnapshot, err error) {
	if projectID == "" || id == "" || revision < 1 {
		return snapshot, domain.WorkspaceError(apierrors.ValidationFailed, "explicit project, goal and positive fixed revision are required")
	}
	project, err := projectAccess(tx, c, projectID)
	if err != nil {
		return snapshot, err
	}
	goal, err := tx.LoadGoal(id, 0)
	if err != nil {
		return snapshot, err
	}
	if goal.ProjectID != project.ID {
		return snapshot, domain.WorkspaceError(apierrors.ScopeDenied, "goal is outside the selected project")
	}
	if goal.Revision != revision {
		return snapshot, domain.WorkspaceError(apierrors.ContextStale, "selected goal revision is no longer current")
	}
	if goal.Status != "active" {
		return snapshot, domain.WorkspaceError(apierrors.ScopeDenied, "selected goal is not active")
	}
	if goal.Title == "" || len(goal.SuccessCriteria) == 0 {
		return snapshot, domain.WorkspaceError(apierrors.EvidenceMissing, "goal has no actual target or success criteria")
	}
	return domain.CurrentGoalSnapshot{Goal: goal, ProjectSettingsRevision: project.Settings.Revision, ObservedAt: time.Now().UTC()}, nil
}

func (s *S2Service) ReadCurrentGoalSnapshot(ctx context.Context, c domain.Caller, projectID, id string, revision int) (snapshot domain.CurrentGoalSnapshot, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	err = s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		var err error
		snapshot, err = currentGoalSnapshot(tx, c, projectID, id, revision)
		return err
	})
	return
}

// 回调只提交同一数据库的判断记录；模型调用在这个事务之外完成。
func (s *S2Service) WithCurrentGoalSnapshot(ctx context.Context, c domain.Caller, expected domain.CurrentGoalSnapshot, fn func(context.Context, domain.CurrentGoalSnapshot) error) (err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if expected.ProjectSettingsRevision < 1 || expected.ObservedAt.IsZero() || fn == nil {
		return domain.WorkspaceError(apierrors.ValidationFailed, "actual goal snapshot and database callback are required")
	}
	return s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		current, err := currentGoalSnapshot(tx, c, expected.Goal.ProjectID, expected.Goal.ID, expected.Goal.Revision)
		if err != nil {
			return err
		}
		if current.ProjectSettingsRevision != expected.ProjectSettingsRevision {
			return domain.WorkspaceError(apierrors.ContextStale, "project permissions changed after goal snapshot")
		}
		return fn(tx.Context(), current)
	})
}
