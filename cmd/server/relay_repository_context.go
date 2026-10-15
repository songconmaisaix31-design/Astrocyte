package main

import (
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 配置只由可信启动入口提供，HTTP不能选择程序、部署身份或投影目录。
func assembleRepositoryContext(services *httpapi.Services, db *sqlite.DB, options repositorycontext.Options) (*repositorycontext.Service, error) {
	if services == nil || db == nil {
		return nil, errors.New("repository context requires the original services and database")
	}
	provider, err := repositorycontext.New(db, db, options)
	if err != nil {
		return nil, err
	}
	services.RepositoryContext = &httpapi.RepositoryContextServices{Map: provider, Search: provider, Impact: provider, Indexes: provider, MapCommands: provider}
	return provider, nil
}

func assembleConfiguredRepositoryContext(services *httpapi.Services, db *sqlite.DB, objects workspace.ContextObjectStore, config repositoryContextConfig) error {
	if config.EnvironmentID == "" {
		return nil
	}
	if services == nil || services.WorkspaceS2 == nil || objects == nil {
		return errors.New("fixed Workspace context requires the original service and object store")
	}
	provider, err := assembleRepositoryContext(services, db, repositorycontext.Options{
		Git:   repositorycontext.Git{EnvironmentID: config.EnvironmentID},
		Graph: repositorycontext.CodeGraph{Node: config.Node, SDK: config.SDK, Script: config.Script},
		AOCI:  config.AOCI, CacheRoot: config.CacheRoot,
		SnapshotBudget: domain.RepositoryBudget{MaxFiles: config.MaxFiles, MaxBytes: config.MaxBytes, MaxResults: config.MaxResults, MaxDepth: config.MaxDepth},
	})
	if err != nil {
		return err
	}
	// 当前原生绑定端口缺失时，BindSession沿既有服务明确返回unsupported_capability。
	services.WorkspaceS2.ConfigureContext(provider, objects, nil)
	return nil
}
