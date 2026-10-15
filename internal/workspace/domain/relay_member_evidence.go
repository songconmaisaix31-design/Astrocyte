package domain

import (
	"context"
	"time"
)

// 独立私有契约只使用标准库类型；不依赖完整 Relay 状态或 NativeSession。
type RelayMemberEvidence struct {
	Purpose                    string     `json:"-"`
	SessionID                  string     `json:"-"`
	NativeID                   string     `json:"-"`
	ControllerID               string     `json:"-"`
	StartupOperationID         string     `json:"-"`
	InputOperationID           string     `json:"-"`
	ProjectID                  string     `json:"-"`
	CWD                        string     `json:"-"`
	CLI                        string     `json:"-"`
	CLIVersion                 string     `json:"-"`
	ControllerHeld             bool       `json:"-"`
	CurrentInput               bool       `json:"-"`
	ExitConfirmed              bool       `json:"-"`
	InputAccepted              bool       `json:"-"`
	AcceptedTurnID             *string    `json:"-"`
	ControllerStartedAt        *time.Time `json:"-"`
	InputAcceptedAt            *time.Time `json:"-"`
	TurnStartedAt              *time.Time `json:"-"`
	ExitObservedAt             *time.Time `json:"-"`
	Capabilities               []string   `json:"-"`
	Tools                      []string   `json:"-"`
	ExternalTargets            []string   `json:"-"`
	ToolsKnown                 bool       `json:"-"`
	TargetsKnown               bool       `json:"-"`
	CallsKnown                 bool       `json:"-"`
	TokensKnown                bool       `json:"-"`
	ObservedCompletedResponses *int64     `json:"-"`
	ReportedInputTokens        *int64     `json:"-"`
	ReportedOutputTokens       *int64     `json:"-"`
	ReportedTotalTokens        *int64     `json:"-"`
}

// H 注入原 Registry；S3 的桥接自行核对 S2 binding / epoch / grant。
type RelayMemberEvidenceSource interface {
	ReadRelayMemberEvidence(context.Context, string, string, string) (RelayMemberEvidence, error)
	StopRelayMemberEvidence(context.Context, string, string) (RelayMemberEvidence, error)
}
