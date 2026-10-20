package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	orcaadapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/orca"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func composeOrcaExecution(db *sqlite.DB, projects workspaceapp.LocalProjects) (workspaceapp.OrcaExecution, error) {
	root := os.Getenv("SWA_INSTALL_ROOT")
	if root == "" {
		if os.Getenv("SWA_DATA_DIR") != "" {
			return nil, errors.New("SWA_DATA_DIR requires the explicit installed runtime root")
		}
		return nil, nil
	}
	runtime, err := orcaadapter.New(orcaadapter.Config{
		InstallRoot: root, DataRoot: os.Getenv("SWA_DATA_DIR"),
		NodeExecutable: filepath.Join(root, "node", "node.exe"),
		ClientEntry:    filepath.Join(root, "runtime", "client.cjs"),
		PiEntry:        filepath.Join(root, "runtime", "pi-entry.mjs"),
		CodexAuthFile:  os.Getenv("SWA_CODEX_AUTH_FILE"),
		ProtectedRoots: filepath.SplitList(os.Getenv("SWA_PROTECTED_ROOTS")),
	})
	if err != nil {
		return nil, err
	}
	return workspaceapp.NewOrcaExecutionService(db, agents.ProjectFiles{}, projects, runtime, db)
}
