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
	ArchiveFinalized  string `json:"archive_finalized"`
	ArchiveAppendOnly string `json:"archive_append_only"`
	WorkspaceRoot     string `json:"-"`
	SourceRoot        string `json:"-"`
	SessionID         string `json:"-"`
	NativeID          string `json:"-"`
	Ownership         string `json:"-"`
	ArchivePath       string `json:"-"`
	ArchiveRoot       string `json:"-"`
	ArchiveFormat     string `json:"-"`
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
