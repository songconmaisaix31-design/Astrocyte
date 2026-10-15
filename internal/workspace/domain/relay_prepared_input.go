package domain

import (
	"context"
	"time"
)

// 只读原私有操作的准备记录；不包含输入正文或 ObjectStore 引用。
type RelayPreparedInput struct {
	Session            NativeSession `json:"-"`
	OperationID        string        `json:"-"`
	ActorID            string        `json:"-"`
	ActorKind          string        `json:"-"`
	ActorProjectID     string        `json:"-"`
	ControllerID       string        `json:"-"`
	StartupOperationID string        `json:"-"`
	WorkspaceRoot      string        `json:"-"`
	PreparedAt         time.Time     `json:"-"`
}

// B 复用当前事务核原操作者权限，只返回明确尚未尝试发送的准备。
// 发送在途、未知及已受理均拒绝；读取不访问正文或启动控制器。
type RelayPreparedInputReader interface {
	ReadPreparedNativeSession(context.Context, string, string) (RelayPreparedInput, error)
}
