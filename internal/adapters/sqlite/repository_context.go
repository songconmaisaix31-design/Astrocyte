package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type repositoryContextConnection interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (db *DB) repositoryContextConnection(ctx context.Context) (repositoryContextConnection, error) {
	if ctx.Value(workspaceTransactionKey{}) != nil {
		return WorkspaceTransaction(ctx, db)
	}
	return db.conn, nil
}

func (db *DB) LoadRepositoryProject(ctx context.Context, id string) (domain.LocalProject, error) {
	var result domain.LocalProject
	conn, err := db.repositoryContextConnection(ctx)
	if err != nil {
		return result, err
	}
	var data string
	err = conn.QueryRowContext(ctx, "SELECT data FROM local_agent_projects WHERE id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return result, apierrors.NewNotFound("repository project", id)
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

func (db *DB) LoadRepositoryGrant(ctx context.Context, projectID, agentID string) (domain.ProjectGrant, error) {
	var result domain.ProjectGrant
	conn, err := db.repositoryContextConnection(ctx)
	if err != nil {
		return result, err
	}
	var data string
	err = conn.QueryRowContext(ctx, "SELECT data FROM local_agent_grants WHERE project_id=? AND agent_id=?", projectID, agentID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return result, apierrors.NewNotFound("repository project grant", agentID)
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(data), &result)
	return result, err
}

func (db *DB) LoadRepositoryIndex(ctx context.Context, projectID, worktreeID, backend string) (domain.RepositoryIndex, error) {
	var result domain.RepositoryIndex
	conn, err := db.repositoryContextConnection(ctx)
	if err != nil {
		return result, err
	}
	var data string
	err = conn.QueryRowContext(ctx, "SELECT metadata_json FROM repository_context_indexes WHERE project_id=? AND worktree_id=? AND backend=?", projectID, worktreeID, backend).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return result, apierrors.NewNotFound("repository index", backend)
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, err
	}
	if result.Snapshot.ProjectID != projectID || result.Snapshot.WorktreeID != worktreeID || result.Backend != backend {
		return result, conflict("repository index identity changed")
	}
	return result, nil
}

func (db *DB) SaveRepositoryIndex(ctx context.Context, index domain.RepositoryIndex, expected string) error {
	conn, err := db.repositoryContextConnection(ctx)
	if err != nil {
		return err
	}
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == "" {
		result, err = conn.ExecContext(ctx, "INSERT INTO repository_context_indexes(project_id,worktree_id,backend,generation,metadata_json) VALUES(?,?,?,?,?) ON CONFLICT(project_id,worktree_id,backend) DO NOTHING", index.Snapshot.ProjectID, index.Snapshot.WorktreeID, index.Backend, index.Generation, string(data))
	} else {
		result, err = conn.ExecContext(ctx, "UPDATE repository_context_indexes SET generation=?,metadata_json=? WHERE project_id=? AND worktree_id=? AND backend=? AND generation=?", index.Generation, string(data), index.Snapshot.ProjectID, index.Snapshot.WorktreeID, index.Backend, expected)
	}
	return workspaceCAS(result, err)
}
