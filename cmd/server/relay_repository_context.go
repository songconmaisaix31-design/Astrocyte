package main

import (
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
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
