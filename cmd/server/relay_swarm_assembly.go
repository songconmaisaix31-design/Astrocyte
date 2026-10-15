package main

import (
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
)

// 账本只使用同一数据库及当前S2批准服务，缺少原生端口的动作明确拒绝。
func assembleSwarmLedger(services *httpapi.Services, db *sqlite.DB) error {
	if services == nil || services.WorkspaceS2 == nil || db == nil {
		return errors.New("Swarm ledger requires the original database and Workspace approval service")
	}
	repository, err := sqlite.NewSwarmRepository(db)
	if err != nil {
		return err
	}
	grants, err := sqlite.NewSwarmApprovalBridge(db, services.WorkspaceS2)
	if err != nil {
		return err
	}
	service, err := swarm.NewSwarmLedger(repository, grants)
	if err != nil {
		return err
	}
	services.SwarmS3 = service
	return nil
}
