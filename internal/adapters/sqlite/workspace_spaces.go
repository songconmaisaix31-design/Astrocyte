package sqlite

import "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/swastore"

type ProjectSpaceRepository = swastore.ProjectSpaceRepository

func NewProjectSpaceRepository(db *DB) *ProjectSpaceRepository {
	return swastore.NewProjectSpaceRepository(db.DB)
}
