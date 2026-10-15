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

type RepositoryProjectReader interface {
	LoadRepositoryProject(context.Context, string) (domain.LocalProject, error)
	LoadRepositoryGrant(context.Context, string, string) (domain.ProjectGrant, error)
}

// 提交事务内只核保存的快照和当前权限；真实Git观察由事务外调用方完成。
type ProjectSnapshotReferenceValidator interface {
	ValidateProjectSnapshotReference(context.Context, domain.Caller, domain.ProjectContextSnapshot) error
}
