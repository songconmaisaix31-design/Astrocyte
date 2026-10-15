package app

import "context"

// SessionEvents is independent local admission, never a cognitive snapshot.
type SessionEvents struct {
	Events   []Event           `json:"events"`
	Coverage []SessionCoverage `json:"coverage"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type SessionCoverage struct {
	Source           string `json:"source"`
	TaskID           string `json:"task_id"`
	TerminalID       string `json:"terminal_id"`
	WorktreeID       string `json:"worktree_id"`
	ProjectID        string `json:"project_id"`
	Status           string `json:"status"` // unknown, partial, complete
	Gap              string `json:"gap"`
	Lines            int64  `json:"lines"`
	Projected        int64  `json:"projected"`
	Duplicates       int64  `json:"duplicates"`
	Unknown          int64  `json:"unknown"`
	Errors           int64  `json:"errors"`
	UnknownTokens    int64  `json:"unknown_tokens"`
	RemainingBytes   int64  `json:"remaining_bytes"`
	TailBytes        int64  `json:"tail_bytes"`
	SnapshotComplete bool   `json:"snapshot_complete"`
	SessionComplete  bool   `json:"session_complete"`
}

type SessionEventsQuery struct {
	TaskID, Start, End string
	Page, PageSize     int
}

// Transport calls this read port only; ownership and current grant/revision are
// rechecked by its entrypoint bridge before reading cached derived admissions.
type SessionEventsReader interface {
	ReadSessionEvents(context.Context, SessionEventsQuery) (SessionEvents, error)
}

// Internal selection from the owned runtime, never decoded from HTTP. The
// repository returns committed receipt IDs, not parser or cursor event IDs.
type SessionAdmissionScope struct {
	TaskID, ProjectID, TerminalID, NativeSessionID, NativeBindingID string
	RelayEventIDs                                                   []string
	NativeNotifications                                             bool
}
type SessionAdmissions interface {
	ReadCommittedSessionEvents(context.Context, []SessionAdmissionScope, SessionEventsQuery) ([]Event, int, error)
}
