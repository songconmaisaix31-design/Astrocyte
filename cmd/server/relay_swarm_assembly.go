package main

import (
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 使用同一数据库、批准服务、Registry和原件存储；未知执行事实由领域拒绝。
func assembleSwarmLedger(services *httpapi.Services, db *sqlite.DB, registry *agents.Registry, originals workspace.ContextObjectStore) error {
	if services == nil || services.WorkspaceS2 == nil || db == nil || registry == nil || originals == nil {
		return errors.New("Swarm requires the original database, approval service, Registry and object store")
	}
	repository, err := sqlite.NewSwarmRepository(db)
	if err != nil {
		return err
	}
	grants, err := sqlite.NewSwarmApprovalBridge(db, services.WorkspaceS2)
	if err != nil {
		return err
	}
	prepared, ok := services.LocalProjects.(workspacedomain.RelayPreparedInputReader)
	if !ok {
		return errors.New("Swarm requires the actual Workspace prepared input reader")
	}
	members, err := agents.NewRelayPreparedMemberEvidenceSource(registry, prepared)
	if err != nil {
		return err
	}
	native, err := sqlite.NewSwarmNativeBridge(repository, members, members)
	if err != nil {
		return err
	}
	evidence, err := sqlite.NewSwarmGitEvidence(originals)
	if err != nil {
		return err
	}
	service, err := swarm.NewSwarmService(repository, grants, native, evidence, native)
	if err != nil {
		return err
	}
	services.SwarmS3 = service
	return nil
}
