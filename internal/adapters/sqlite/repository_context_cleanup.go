package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type repositoryCleanupCandidate struct {
	projectID, worktreeID, backend, generation, metadata string
	projectJSON                                          sql.NullString
	index                                                domain.RepositoryIndex
	valid                                                bool
}

// 外部观察全部在写事务之外完成；提交前再次核实项目和索引的准确版本。
func (db *DB) PrepareRepositoryContextCleanup(ctx context.Context, inspect func(context.Context, domain.LocalProject, domain.RepositoryIndex) (bool, error)) (app.RepositoryCleanupState, error) {
	var state app.RepositoryCleanupState
	if ctx.Value(workspaceTransactionKey{}) != nil || inspect == nil {
		return state, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: "repository cleanup must run outside a request transaction"}
	}
	rows, err := db.conn.QueryContext(ctx, "SELECT i.project_id,i.worktree_id,i.backend,i.generation,i.metadata_json,p.data FROM repository_context_indexes i LEFT JOIN local_agent_projects p ON p.id=i.project_id ORDER BY i.project_id,i.worktree_id,i.backend")
	if err != nil {
		return state, err
	}
	defer rows.Close()
	candidates := []repositoryCleanupCandidate{}
	for rows.Next() {
		var candidate repositoryCleanupCandidate
		if err := rows.Scan(&candidate.projectID, &candidate.worktreeID, &candidate.backend, &candidate.generation, &candidate.metadata, &candidate.projectJSON); err != nil {
			return state, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return state, err
	}
	if err := rows.Close(); err != nil {
		return state, err
	}
	for i := range candidates {
		candidate := &candidates[i]
		if err := json.Unmarshal([]byte(candidate.metadata), &candidate.index); err != nil {
			return state, err
		}
		index := candidate.index
		if index.Snapshot.ProjectID != candidate.projectID || index.Snapshot.WorktreeID != candidate.worktreeID || index.Backend != candidate.backend || index.Generation != candidate.generation {
			return state, conflict("repository cleanup metadata identity changed")
		}
		var project domain.LocalProject
		if candidate.projectJSON.Valid {
			if err := json.Unmarshal([]byte(candidate.projectJSON.String), &project); err != nil {
				return state, err
			}
			if project.ID != candidate.projectID {
				return state, conflict("repository cleanup project identity changed")
			}
		}
		candidate.valid = candidate.projectJSON.Valid && project.Settings.AllowDirectory && len(project.Settings.AllowedSubdirs) > 0 && project.Settings.Revision == index.Snapshot.SettingsRevision
		if candidate.valid {
			candidate.valid, err = inspect(ctx, project, index)
			if err != nil {
				return state, err
			}
		}
	}
	err = NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		conn, err := WorkspaceTransaction(tx.Context(), db)
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			var metadata string
			var projectJSON sql.NullString
			err := conn.QueryRowContext(ctx, "SELECT i.metadata_json,p.data FROM repository_context_indexes i LEFT JOIN local_agent_projects p ON p.id=i.project_id WHERE i.project_id=? AND i.worktree_id=? AND i.backend=? AND i.generation=?", candidate.projectID, candidate.worktreeID, candidate.backend, candidate.generation).Scan(&metadata, &projectJSON)
			if err == sql.ErrNoRows {
				return conflict("repository cleanup index changed during observation")
			}
			if err != nil {
				return err
			}
			if metadata != candidate.metadata || projectJSON != candidate.projectJSON {
				return conflict("repository cleanup authorization or index changed during observation")
			}
		}
		for _, candidate := range candidates {
			if candidate.valid {
				state.Retained = append(state.Retained, candidate.index)
				continue
			}
			result, err := conn.ExecContext(ctx, "DELETE FROM repository_context_indexes WHERE project_id=? AND worktree_id=? AND backend=? AND generation=? AND metadata_json=?", candidate.projectID, candidate.worktreeID, candidate.backend, candidate.generation, candidate.metadata)
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
