package sqlite

import (
	"context"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/swastore"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 尚保留的旧原生输入组装复用同一项目仓储连接和存储投影。
func (db *DB) localProjectConnection(ctx context.Context) (swastore.SQL, error) {
	return db.DB.ProjectConnection(ctx)
}
func localNativeSessionForStorage(session domain.NativeSession) domain.NativeSession {
	return swastore.NativeSessionForStorage(session)
}
