package domain

import "time"

// RelayOwnedSession is an internal read binding, not an HTTP session DTO or a
// capability grant. Paths and native identifiers must never be serialized to UI.
type RelayOwnedSession struct {
	BindingID    string     `json:"binding_id"`
	TaskID       string     `json:"task_id"`
	TerminalID   string     `json:"terminal_id"`
	WorktreeID   string     `json:"worktree_id"`
	ProjectID    string     `json:"project_id"`
	Agent        string     `json:"agent"`
	DriveMode    string     `json:"drive_mode"`
	SourceStatus string     `json:"source_status"`
	Gap          string     `json:"gap,omitempty"`
	ExitedAt     *time.Time `json:"exited_at"`
	// Process exit does not prove archive flushing or append-only semantics.
	ArchiveFinalized  string                     `json:"archive_finalized"`
	ArchiveAppendOnly string                     `json:"archive_append_only"`
	WorkspaceRoot     string                     `json:"-"`
	SourceRoot        string                     `json:"-"`
	SessionID         string                     `json:"-"`
	NativeID          string                     `json:"-"`
	Ownership         string                     `json:"-"`
	ArchivePath       string                     `json:"-"`
	ArchiveRoot       string                     `json:"-"`
	ArchiveFormat     string                     `json:"-"`
	NativeCapture     *RelayNativeSessionCapture `json:"-"`
}

// RelaySessionArchive is positive source selection supplied by a trusted
// controller port. It is never decoded from public registration or HTTP input.
// The identity and CWD must match the owned session before accepting this source.
type RelaySessionArchive struct {
	SessionID     string `json:"-"`
	NativeID      string `json:"-"`
	WorkspaceRoot string `json:"-"`
	Path          string `json:"-"`
	Root          string `json:"-"`
	Format        string `json:"-"`
	Finalized     bool   `json:"-"`
	AppendOnly    bool   `json:"-"`
}

// RelayArchiveObservationFailure 表示原生操作之后发生的来源读取失败。
// 该错误不撤销已经接受的输入，也不授权重发或启动第二个控制器。
type RelayArchiveObservationFailure struct{ Cause error }

func (e *RelayArchiveObservationFailure) Error() string { return e.Cause.Error() }
func (e *RelayArchiveObservationFailure) Unwrap() error { return e.Cause }

// RelayNativeSessionEvent 只保留控制器通知的派生信息；TS 是实际接收时间。
type RelayNativeSessionEvent struct {
	Sequence int64     `json:"sequence"`
	Kind     string    `json:"kind"`
	Tokens   *int      `json:"tokens"`
	TS       time.Time `json:"ts"`
}

// 覆盖范围是已持有 app-server 暴露的通知流，内部未暴露事件保持限制。
type RelayNativeSessionCapture struct {
	Events         []RelayNativeSessionEvent `json:"events"`
	SourceStatus   string                    `json:"source_status"`
	PrefixComplete string                    `json:"prefix_complete"`
	Finalized      string                    `json:"finalized"`
	LastSequence   int64                     `json:"last_sequence"`
	Errors         int64                     `json:"errors"`
	Unknown        int64                     `json:"unknown"`
	FirstGap       string                    `json:"first_gap,omitempty"`
}

type RelayNativeSessionBatch struct {
	OwnedSession RelayOwnedSession `json:"-"`
	RelayNativeSessionCapture
}
