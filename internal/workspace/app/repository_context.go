package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type RepositoryMap interface {
	ReadRepositoryMap(context.Context, domain.Caller, domain.RepositoryQuery) (domain.RepositoryResult, error)
}

type CodeSearch interface {
	SearchRepositoryCode(context.Context, domain.Caller, domain.RepositoryQuery) (domain.RepositoryResult, error)
}

type ImpactQuery interface {
	QueryRepositoryImpact(context.Context, domain.Caller, domain.RepositoryQuery) (domain.RepositoryResult, error)
}

// 读取和显式投影维护使用不同方法，GET不触发索引器。
type RepositoryIndexCommands interface {
	UpdateRepositoryIndex(context.Context, domain.Caller, domain.RepositoryQuery, string) (domain.RepositoryIndex, error)
}

type RepositoryMapCommands interface {
	RefreshRepositoryMap(context.Context, domain.Caller, domain.RepositoryQuery) (domain.RepositoryIndex, error)
}

type RepositoryProjectReader interface {
	LoadRepositoryProject(context.Context, string) (domain.LocalProject, error)
	LoadRepositoryGrant(context.Context, string, string) (domain.ProjectGrant, error)
}

// 仅供平台后台清理生命周期调用，必须复用索引和查询所用的同一个Service。
type RepositoryProjectionCleaner interface {
	CleanupRepositoryContext(context.Context, int) (domain.RepositoryCleanupResult, error)
}

type RepositoryCleanupNotifier interface {
	WakeRepositoryCleanup()
}

type RepositoryCleanupWorker interface {
	RepositoryCleanupNotifier
	RunRepositoryCleanup(context.Context, int) error
}

type RepositoryCleanupState struct {
	RetiredIndexes int
	Retained       []domain.RepositoryIndex
}
