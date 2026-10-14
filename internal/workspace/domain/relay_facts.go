package domain

import "time"

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
	ControllerID         string            `json:"-"`
	PreviousControllerID string            `json:"-"`
	StartupOperationID   string            `json:"-"`
	SessionID            string            `json:"-"`
	NativeID             string            `json:"-"`
	Provider             string            `json:"-"`
	WorkspaceRoot        string            `json:"-"`
	Relation             string            `json:"-"`
	Facts                []RelayDriverFact `json:"-"`
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
