package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ app.RegisteredProjectRepository = (*DB)(nil)

func normalizeProjectDiscovery(s domain.ProjectDiscoverySnapshot) domain.ProjectDiscoverySnapshot {
	if s.Status == "" {
		s.Status = "unknown"
	}
	if s.Projects == nil {
		s.Projects = []domain.RegisteredProject{}
	}
	if s.Failures == nil {
		s.Failures = []domain.ProjectDiscoveryFailure{}
	}
	for i := range s.Projects {
		if s.Projects[i].Limitations == nil {
			s.Projects[i].Limitations = []string{}
		}
		if s.Projects[i].Contributors == nil {
			s.Projects[i].Contributors = []domain.ProjectContributor{}
		}
		for j := range s.Projects[i].Contributors {
			c := &s.Projects[i].Contributors[j]
			// Earlier board caches projected header creation as activity. Repair
			// only that explicitly identified source in memory on GET, no writes.
			if c.Source == "native_session_header" {
				if c.CreatedAt == nil {
					c.CreatedAt = c.ActivityAt
				}
				c.ActivityAt = nil
			}
		}
	}
	if s.Board == nil {
		s.Board = []domain.ProjectSummary{}
	}
	if s.Sources == nil {
		s.Sources = []domain.ProjectSourceObservation{}
	}
	return s
}

func (db *DB) LoadRegisteredProjectDiscovery(ctx context.Context) (domain.ProjectDiscoverySnapshot, error) {
	s := normalizeProjectDiscovery(domain.ProjectDiscoverySnapshot{})
	var data string
	err := db.conn.QueryRowContext(ctx, "SELECT data FROM local_project_discovery WHERE id='registered'").Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal([]byte(data), &s)
	return normalizeProjectDiscovery(s), err
}

func (db *DB) SaveRegisteredProjectDiscovery(ctx context.Context, s domain.ProjectDiscoverySnapshot) error {
	data, err := json.Marshal(normalizeProjectDiscovery(s))
	if err != nil {
		return err
	}
	_, err = db.conn.ExecContext(ctx, "INSERT INTO local_project_discovery(id,data) VALUES('registered',?) ON CONFLICT(id) DO UPDATE SET data=excluded.data", string(data))
	return err
}
