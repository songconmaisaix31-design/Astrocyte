package domain

import "time"

// 当前所有权只来自此服务仍持有的原控制器，独立于历史事实和活性。
type RelayCurrentOwnership struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// 原完整 Start 返回会话只在本机可信闭包中传递，不成为 HTTP DTO。
type RelayOwnedNativeSession struct {
	Session  NativeSession           `json:"-"`
	Evidence RelayControllerEvidence `json:"-"`
}

// 事实独立于收据和任务状态；后到观察不得重写原动作结果。
type RelayRuntimeFact struct {
	ID          string    `json:"fact_id"`
	TaskID      string    `json:"task_id"`
	TerminalID  string    `json:"terminal_id"`
	WorktreeID  string    `json:"worktree_id"`
	ReceiptID   string    `json:"receipt_id,omitempty"`
	OperationID string    `json:"operation_id,omitempty"`
	Kind        string    `json:"kind"`
	Status      string    `json:"status"`
	Source      string    `json:"source"`
	Reason      string    `json:"reason,omitempty"`
	Sequence    int64     `json:"sequence"`
	TS          time.Time `json:"ts"`
}

// 此端口只复制原控制器的事实，不包含原生输出或输入。
type RelayDriverFact struct {
	OperationID string
	Kind        string
	Status      string
	Source      string
	Reason      string
	Sequence    int64
	TS          time.Time
}

// 身份资料只在隔离 store 的私有部分持久化，不序列化到产品视图。
type RelayControllerEvidence struct {
	ControllerID         string             `json:"-"`
	PreviousControllerID string             `json:"-"`
	StartupOperationID   string             `json:"-"`
	SessionID            string             `json:"-"`
	NativeID             string             `json:"-"`
	Provider             string             `json:"-"`
	WorkspaceRoot        string             `json:"-"`
	Relation             string             `json:"-"`
	Facts                []RelayDriverFact  `json:"-"`
	Turns                []RelayTurnBinding `json:"-"`
	ExitConfirmed        bool               `json:"-"`
}

// 原生 turn 标识仅供本机按准确输入核输出，不进入公共事实或事件 DTO。
type RelayTurnBinding struct {
	OperationID        string     `json:"-"`
	SessionID          string     `json:"-"`
	NativeID           string     `json:"-"`
	ControllerID       string     `json:"-"`
	StartupOperationID string     `json:"-"`
	TurnID             string     `json:"-"`
	AcceptedAt         time.Time  `json:"-"`
	StartedAt          *time.Time `json:"-"`
}

type RelaySessionRelation struct {
	TaskID       string    `json:"task_id"`
	PreviousTask string    `json:"previous_task_id,omitempty"`
	WorktreeID   string    `json:"worktree_id"`
	TerminalID   string    `json:"terminal_id"`
	OperationID  string    `json:"operation_id"`
	Kind         string    `json:"kind"`
	Status       string    `json:"status"`
	Source       string    `json:"source"`
	TS           time.Time `json:"ts"`
}
