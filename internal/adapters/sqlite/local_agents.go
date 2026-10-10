package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ app.LocalProjectRepository = (*DB)(nil)

func (db *DB) ListProjects(ctx context.Context) ([]domain.LocalProject, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT data FROM local_agent_projects ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.LocalProject{}
	for rows.Next() {
		var data string
		var p domain.LocalProject
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func (db *DB) LoadProject(ctx context.Context, id string) (domain.LocalProject, error) {
	var p domain.LocalProject
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_agent_projects WHERE id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return p, apierrors.NewNotFound("local project", id)
	}
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(data), &p)
	return p, err
}
func (db *DB) SaveProject(ctx context.Context, p domain.LocalProject, expected int) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if p.Settings.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "project settings revision differs", RequiredAction: "reload_project_settings"}
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?) ON CONFLICT(id) DO NOTHING", p.ID, p.Settings.Revision, string(data))
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE local_agent_projects SET version=?,data=? WHERE id=? AND version=?", p.Settings.Revision, string(data), p.ID, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "project settings changed", RequiredAction: "reload_project_settings"}
	}
	return nil
}

type storedLocalGrant struct {
	domain.ProjectGrant
	PrivateTokenDigest string `json:"private_token_digest"`
}

func (db *DB) LoadGrant(ctx context.Context, projectID, agentID string) (domain.ProjectGrant, error) {
	var g storedLocalGrant
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_agent_grants WHERE project_id=? AND agent_id=?", projectID, agentID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return g.ProjectGrant, apierrors.NewNotFound("project Agent grant", agentID)
	}
	if err != nil {
		return g.ProjectGrant, err
	}
	err = json.Unmarshal([]byte(data), &g)
	g.TokenDigest = g.PrivateTokenDigest
	return g.ProjectGrant, err
}
func (db *DB) SaveGrant(ctx context.Context, g domain.ProjectGrant) error {
	data, err := json.Marshal(storedLocalGrant{ProjectGrant: g, PrivateTokenDigest: g.TokenDigest})
	if err != nil {
		return err
	}
	_, err = db.conn.ExecContext(ctx, "INSERT INTO local_agent_grants(project_id,agent_id,data) VALUES(?,?,?) ON CONFLICT(project_id,agent_id) DO UPDATE SET data=excluded.data", g.ProjectID, g.AgentID, string(data))
	return err
}
func (db *DB) ListSessions(ctx context.Context, projectID string) ([]domain.NativeSession, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT data FROM local_agent_sessions WHERE project_id=? ORDER BY id", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.NativeSession{}
	for rows.Next() {
		var data string
		var s domain.NativeSession
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(data), &s); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
func (db *DB) LoadSession(ctx context.Context, id string) (domain.NativeSession, error) {
	var s domain.NativeSession
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_agent_sessions WHERE id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return s, apierrors.NewNotFound("native session", id)
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal([]byte(data), &s)
	return s, err
}
func (db *DB) SaveSession(ctx context.Context, s domain.NativeSession) error {
	// Persist fixed refs and file snapshot metadata, not duplicate original bodies.
	s.ContextPacket.Materials = append([]domain.ContextMaterial{}, s.ContextPacket.Materials...)
	for i := range s.ContextPacket.Materials {
		s.ContextPacket.Materials[i].Text = ""
	}
	s.ContextPacket.Files = append([]domain.ContextFile{}, s.ContextPacket.Files...)
	for i := range s.ContextPacket.Files {
		s.ContextPacket.Files[i].Text = ""
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	result, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_sessions(id,project_id,data) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET data=excluded.data WHERE local_agent_sessions.project_id=excluded.project_id", s.ID, s.ProjectID, string(data))
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "session belongs to another project", RequiredAction: "check_project_session"}
	}
	return nil
}
