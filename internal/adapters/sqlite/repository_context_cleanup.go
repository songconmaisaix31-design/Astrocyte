package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 先在既有Workspace事务中撤下失效元数据，提交后才能清理磁盘投影。
func (db *DB) PrepareRepositoryContextCleanup(ctx context.Context, inspect func(context.Context, domain.LocalProject, domain.RepositoryIndex) (bool, error)) (app.RepositoryCleanupState, error) {
	var state app.RepositoryCleanupState
	if ctx.Value(workspaceTransactionKey{}) != nil || inspect == nil {
		return state, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "repository cleanup must run outside a request transaction"}
	}
	err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		conn, err := WorkspaceTransaction(tx.Context(), db)
		if err != nil {
			return err
		}
		rows, err := conn.QueryContext(ctx, "SELECT i.project_id,i.worktree_id,i.backend,i.generation,i.metadata_json,p.data FROM repository_context_indexes i LEFT JOIN local_agent_projects p ON p.id=i.project_id ORDER BY i.project_id,i.worktree_id,i.backend")
		if err != nil {
			return err
		}
		defer rows.Close()
		retired := []domain.RepositoryIndex{}
		for rows.Next() {
			var projectID, worktreeID, backend, generation, metadata string
			var projectJSON sql.NullString
			if err := rows.Scan(&projectID, &worktreeID, &backend, &generation, &metadata, &projectJSON); err != nil {
				return err
			}
			var index domain.RepositoryIndex
			if err := json.Unmarshal([]byte(metadata), &index); err != nil {
				return err
			}
			if index.Snapshot.ProjectID != projectID || index.Snapshot.WorktreeID != worktreeID || index.Backend != backend || index.Generation != generation {
				return conflict("repository cleanup metadata identity changed")
			}
			var project domain.LocalProject
			if projectJSON.Valid {
				if err := json.Unmarshal([]byte(projectJSON.String), &project); err != nil {
					return err
				}
				if project.ID != projectID {
					return conflict("repository cleanup project identity changed")
				}
			}
			valid := projectJSON.Valid && project.Settings.AllowDirectory && len(project.Settings.AllowedSubdirs) > 0 && project.Settings.Revision == index.Snapshot.SettingsRevision
			if valid {
				valid, err = inspect(ctx, project, index)
				if err != nil {
					return err
				}
			}
			if valid {
				state.Retained = append(state.Retained, index)
			} else {
				retired = append(retired, index)
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		for _, index := range retired {
			result, err := conn.ExecContext(ctx, "DELETE FROM repository_context_indexes WHERE project_id=? AND worktree_id=? AND backend=? AND generation=?", index.Snapshot.ProjectID, index.Snapshot.WorktreeID, index.Backend, index.Generation)
			if err := workspaceCAS(result, err); err != nil {
				return err
			}
			state.RetiredIndexes++
		}
		return nil
	})
	if err != nil {
		return app.RepositoryCleanupState{}, err
	}
	return state, nil
}
