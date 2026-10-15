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
	ControllerID       string        `json:"-"`
	StartupOperationID string        `json:"-"`
	WorkspaceRoot      string        `json:"-"`
	PreparedAt         time.Time     `json:"-"`
}

// B 复用当前事务核原操作者权限与尚未提交的原操作，不读取正文或启动控制器。
type RelayPreparedInputReader interface {
	ReadPreparedNativeSession(context.Context, string, string) (RelayPreparedInput, error)
}
