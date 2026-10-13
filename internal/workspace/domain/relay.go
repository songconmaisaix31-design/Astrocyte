package domain

import (
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

const (
	RelayReady        = "ready"
	RelayAccepted     = "accepted"
	RelayWorking      = "working"
	RelayDone         = "done"
	RelayReleased     = "released"
	RelayLive         = "live"
	RelayUnverifiable = "unverifiable"
	RelayExited       = "exited"
)

// AgentSpec is trusted registration data, never command arguments from HTTP.
// Capabilities remain independent observations; registration does not prove readiness.
type AgentSpec struct {
	Name         string                           `json:"name"`
	LaunchCmd    []string                         `json:"launch_cmd"`
	Capabilities map[string]CapabilityObservation `json:"capabilities"`
	DriveMode    string                           `json:"drive_mode"`
}

// RelayWorkspace exists independently of an Agent. ActiveTaskID includes unknown
// deliveries and exited unfinished tasks until an explicit safe release.
type RelayWorkspace struct {
	ID             string    `json:"worktree_id"`
	ProjectID      string    `json:"project_id"`
	SourceRoot     string    `json:"-"`
	Root           string    `json:"root"`
	Branch         string    `json:"branch"`
	Prepared       bool      `json:"prepared"`
	ActiveTaskID   string    `json:"active_task_id"`
	GitObservation string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

// RelayTerminal.ID is durable and never a PID. Native session details are kept
// separately in RelayState, so a view cannot disclose raw context or history paths.
type RelayTerminal struct {
	ID           string     `json:"terminal_id"`
	WorktreeID   string     `json:"worktree_id"`
	Agent        string     `json:"agent"`
	DriveMode    string     `json:"drive_mode"`
	Liveness     string     `json:"liveness"`
	LastEventAt  *time.Time `json:"last_event_at"`
	ExitedAt     *time.Time `json:"exited_at"`
	LastSequence int        `json:"last_sequence"`
	CreatedAt    time.Time  `json:"created_at"`
}

type RelayTask struct {
	ID               string    `json:"task_id"`
	WorkID           string    `json:"work_id"`
	ProjectID        string    `json:"project_id"`
	WorktreeID       string    `json:"worktree_id"`
	TerminalID       string    `json:"terminal_id"`
	Agent            string    `json:"agent"`
	Description      string    `json:"description"`
	State            string    `json:"state"`
	Condition        string    `json:"condition"`
	PermissionLevel  string    `json:"permission_level"`
	SettingsRevision int       `json:"project_settings_revision"`
	DegradedLaunch   bool      `json:"degraded_launch"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RelayTransition struct {
	TaskID   string    `json:"task_id"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	TS       time.Time `json:"ts"`
	Operator string    `json:"operator"`
}

type RelayReceipt struct {
	ID          string                  `json:"receipt_id"`
	OperationID string                  `json:"operation_id"`
	TaskID      string                  `json:"task_id"`
	Action      string                  `json:"action"`
	Operator    string                  `json:"operator"`
	Status      string                  `json:"status"`
	Error       *apierrors.ServiceError `json:"error"`
	CreatedAt   time.Time               `json:"created_at"`
	FinishedAt  *time.Time              `json:"finished_at"`
	// Input is an internal exact idempotency comparison, never a public payload.
	Input []byte `json:"-"`
}

// RelayEvent contains derived facts only: no prompts, file names, terminal text,
// native payloads, credentials, or completion inference from Git activity.
type RelayEvent struct {
	ID         string    `json:"event_id"`
	TaskID     string    `json:"task_id"`
	TerminalID string    `json:"terminal_id"`
	WorktreeID string    `json:"worktree_id"`
	Agent      string    `json:"agent"`
	Kind       string    `json:"kind"`
	Source     string    `json:"source"`
	Sequence   int       `json:"sequence"`
	Topic      string    `json:"topic"`
	Tokens     *int64    `json:"tokens"`
	Scope      string    `json:"scope"`
	TS         time.Time `json:"ts"`
	Published  bool      `json:"-"`
}

type RelayTaskView struct {
	Task        RelayTask         `json:"task"`
	Terminal    *RelayTerminal    `json:"terminal"`
	Workspace   *RelayWorkspace   `json:"worktree"`
	Transitions []RelayTransition `json:"transitions"`
}

type RelayActionResult struct {
	View    RelayTaskView `json:"view"`
	Receipt RelayReceipt  `json:"receipt"`
}

// Every edge is persisted. Human completion is explicit; turn completion and
// successful Git commands do not establish that the task is done.
func ValidRelayTransition(from, to string) bool {
	return (from == RelayReady && to == RelayAccepted) ||
		(from == RelayAccepted && to == RelayWorking) ||
		(from == RelayWorking && to == RelayDone) ||
		(from == RelayDone && to == RelayReleased)
}

func RelayLiveness(last, exited *time.Time, now time.Time, window time.Duration) string {
	if exited != nil {
		return RelayExited
	}
	if last != nil && window > 0 && !last.After(now) && now.Sub(*last) <= window {
		return RelayLive
	}
	return RelayUnverifiable
}
