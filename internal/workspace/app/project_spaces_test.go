package app_test

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func actualAttentionSpaces(db *sqlite.DB) projectspacebridge.Attention {
	return projectspacebridge.Attention{Spaces: workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})}
}
