package domain

import "time"

const (
	RelayVerifiedComplete      = "verified_complete"
	RelayUserConfirmedComplete = "user_confirmed_complete"
)

// 完成记录保留原人工动作与目标核验出处；不生成开始或退出事实。
type RelayCompletion struct {
	ID               string     `json:"completion_id"`
	TaskID           string     `json:"task_id"`
	WorktreeID       string     `json:"worktree_id"`
	TerminalID       string     `json:"terminal_id"`
	Channel          string     `json:"channel"`
	Source           string     `json:"source"`
	Operator         string     `json:"operator"`
	ActorID          string     `json:"actor_id"`
	ActorKind        string     `json:"actor_kind"`
	ReceiptID        string     `json:"receipt_id"`
	OperationID      string     `json:"operation_id"`
	TS               time.Time  `json:"ts"`
	EvidenceID       string     `json:"evidence_id,omitempty"`
	InputOperationID string     `json:"input_operation_id,omitempty"`
	OutputEventIDs   []string   `json:"output_event_ids,omitempty"`
	CheckEvidenceIDs []string   `json:"check_evidence_ids,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
}

// 受信读取者核合法 owned 输出和固定目标检查；HTTP 只提交证据引用。
type RelayGoalVerification struct {
	EvidenceID         string    `json:"-"`
	TaskID             string    `json:"-"`
	WorktreeID         string    `json:"-"`
	TerminalID         string    `json:"-"`
	ControllerID       string    `json:"-"`
	StartupOperationID string    `json:"-"`
	SessionID          string    `json:"-"`
	NativeID           string    `json:"-"`
	InputOperationID   string    `json:"-"`
	TurnID             string    `json:"-"`
	OutputEventIDs     []string  `json:"-"`
	CheckEvidenceIDs   []string  `json:"-"`
	PassedAt           time.Time `json:"-"`
}
