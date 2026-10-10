package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func normalizeGitHubRepository(item domain.GitHubRepository) domain.GitHubRepository {
	item.Metadata = domain.NormalizeGitHubMetadata(item.Metadata)
	return item
}

func (db *DB) ListGitHubRepositories(ctx context.Context) ([]domain.GitHubRepository, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT data FROM local_github_repositories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.GitHubRepository{}
	for rows.Next() {
		var data string
		var item domain.GitHubRepository
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(data), &item); err != nil {
			return nil, err
		}
		result = append(result, normalizeGitHubRepository(item))
	}
	return result, rows.Err()
}

func (db *DB) LoadGitHubRepository(ctx context.Context, id string) (domain.GitHubRepository, error) {
	var item domain.GitHubRepository
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_github_repositories WHERE id=?", id).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return item, apierrors.NewNotFound("GitHub repository", id)
	}
	if err != nil {
		return item, err
	}
	err = json.Unmarshal([]byte(data), &item)
	return normalizeGitHubRepository(item), err
}

func (db *DB) SaveGitHubRepository(ctx context.Context, item domain.GitHubRepository, expected int) error {
	item = normalizeGitHubRepository(item)
	if item.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "repository revision differs"}
	}
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	var result sql.Result
	if expected == 0 {
		result, err = db.conn.ExecContext(ctx, "INSERT INTO local_github_repositories(id,revision,data) VALUES(?,?,?) ON CONFLICT(id) DO NOTHING", item.ID, item.Revision, string(data))
	} else {
		result, err = db.conn.ExecContext(ctx, "UPDATE local_github_repositories SET revision=?,data=? WHERE id=? AND revision=?", item.Revision, string(data), item.ID, expected)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "repository changed concurrently", RequiredAction: "reload_repository"}
	}
	return nil
}
