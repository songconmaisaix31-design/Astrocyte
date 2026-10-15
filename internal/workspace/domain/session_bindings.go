package domain

import "time"

type SessionBinding struct {
	ID                 string                           `json:"id"`
	Version            int                              `json:"version"`
	ProjectID          string                           `json:"project_id"`
	SessionID          string                           `json:"session_id"`
	AdapterID          string                           `json:"adapter_id"`
	NativeSessionID    string                           `json:"-"`
	ControllerID       string                           `json:"-"`
	StartupOperationID string                           `json:"-"`
	WorktreeID         string                           `json:"worktree_id"`
	GrantID            string                           `json:"grant_id"`
	GrantEpoch         int                              `json:"grant_epoch"`
	PacketID           string                           `json:"packet_id"`
	MemberID           string                           `json:"member_id"`
	Status             string                           `json:"status"`
	Capabilities       map[string]CapabilityObservation `json:"capabilities"`
	CreatedAt          time.Time                        `json:"created_at"`
	UpdatedAt          time.Time                        `json:"updated_at"`
}

// 仅来自当前Registry所持有的原生会话，不接受HTTP正文提供能力或所有权。
type OwnedSessionFacts struct {
	SessionID          string
	ProjectID          string
	AdapterID          string
	NativeID           string
	WorktreeID         string
	Root               string
	Owned              bool
	ControllerID       string
	StartupOperationID string
	Liveness           string
	Capabilities       map[string]CapabilityObservation
	Snapshot           ProjectContextSnapshot `json:"-"`
}

// 原 Registry 读取器的私有返回值，绝不直接成为 HTTP DTO。
type OwnedNativeSessionRecord struct {
	Session            NativeSession                    `json:"-"`
	ControllerID       string                           `json:"-"`
	StartupOperationID string                           `json:"-"`
	EvidenceSessionID  string                           `json:"-"`
	EvidenceNativeID   string                           `json:"-"`
	EvidenceAdapterID  string                           `json:"-"`
	WorkspaceRoot      string                           `json:"-"`
	Liveness           string                           `json:"-"`
	Capabilities       map[string]CapabilityObservation `json:"-"`
}
