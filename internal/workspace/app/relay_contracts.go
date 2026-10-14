package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type RelayIntentCommand struct {
	WorkID                  string `json:"work_id"`
	ProjectID               string `json:"project_id"`
	Agent                   string `json:"agent"`
	Task                    string `json:"task"`
	HandoffFromTaskID       string `json:"handoff_from_task_id,omitempty"`
	HandoffContextConfirmed bool   `json:"handoff_context_confirmed,omitempty"`
}

type RelayConfirmCommand struct {
	PermissionConfirmed     bool   `json:"permission_confirmed"`
	PermissionLevel         string `json:"permission_level"`
	ProjectSettingsRevision int    `json:"project_settings_revision"`
}

type RelaySendCommand struct {
	Input string `json:"input"`
}
type RelayTransitionCommand struct {
	From                string `json:"from"`
	To                  string `json:"to"`
	CompletionConfirmed bool   `json:"completion_confirmed"`
}

// RelayRuntime is the sole writer. Caller is supplied by session authentication,
// not decoded from command JSON. Every action requires a durable OperationID.
type RelayRuntime interface {
	ListAgents(context.Context, domain.Caller) ([]domain.AgentSpec, error)
	ListTasks(context.Context, domain.Caller) ([]domain.RelayTaskView, error)
	GetTask(context.Context, domain.Caller, string) (domain.RelayTaskView, error)
	ListEvents(context.Context, domain.Caller, string) ([]domain.RelayEvent, error)
	ListReceipts(context.Context, domain.Caller, string) ([]domain.RelayReceipt, error)
	CreateIntent(context.Context, domain.Caller, RelayIntentCommand) (domain.RelayTaskView, error)
	ConfirmIntent(context.Context, domain.Caller, string, RelayConfirmCommand) (domain.RelayActionResult, error)
	Send(context.Context, domain.Caller, string, RelaySendCommand) (domain.RelayActionResult, error)
	Read(context.Context, domain.Caller, string) (domain.RelayActionResult, error)
	Stop(context.Context, domain.Caller, string) (domain.RelayActionResult, error)
	Resume(context.Context, domain.Caller, string) (domain.RelayActionResult, error)
	Transition(context.Context, domain.Caller, string, RelayTransitionCommand) (domain.RelayTaskView, error)
}

// Internal storage state is never an HTTP DTO. Update is a single transaction;
// the callback changes data only and must never launch a process or emit events.
type RelayState struct {
	Tasks               map[string]domain.RelayTask
	Workspaces          map[string]domain.RelayWorkspace
	Terminals           map[string]domain.RelayTerminal
	Sessions            map[string]domain.NativeSession
	OwnedSessions       map[string]domain.RelayOwnedSession
	Receipts            map[string]domain.RelayReceipt
	Transitions         []domain.RelayTransition
	Events              []domain.RelayEvent
	Published           map[string]bool
	PendingActions      map[string]string
	PublicationErrors   map[string]bool
	RuntimeFacts        []domain.RelayRuntimeFact
	SessionRelations    []domain.RelaySessionRelation
	ControllerEvidence  map[string]domain.RelayControllerEvidence `json:"-"`
	HandoffDeclarations map[string]string
	ControllerHistory   map[string][]domain.RelayControllerEvidence `json:"-"`
}

type RelayStore interface {
	Snapshot(context.Context) (RelayState, error)
	Update(context.Context, func(*RelayState) error) error
}

type RelayWorkspaces interface {
	// Prepare receives no Agent identity or executable. A previously prepared
	// workspace must be reused as-is, including dirty/untracked relay changes.
	Prepare(context.Context, domain.Work, domain.RelayWorkspace) (domain.RelayWorkspace, error)
	Observe(context.Context, domain.RelayWorkspace) (string, error)
}

type RelayLaunch struct {
	Spec         domain.AgentSpec
	Project      domain.LocalProject
	Workspace    domain.RelayWorkspace
	Terminal     domain.RelayTerminal
	Session      domain.NativeSession
	Input        string
	OperationID  string
	ControllerID string
}

type RelayObservation struct {
	Session       domain.NativeSession
	NativeCapture *domain.RelayNativeSessionCapture
	Archive       *domain.RelaySessionArchive
	ArchiveGap    string
	Events        []domain.RelayEvent
	ExitConfirmed bool
	Evidence      *domain.RelayControllerEvidence
}

// 只读取已经持久化的安全事实；不执行控制器观察或就绪探测。
type RelayRuntimeFactsReader interface {
	ListRuntimeFacts(context.Context, domain.Caller, string) ([]domain.RelayRuntimeFact, []domain.RelaySessionRelation, error)
}

// 历史绑定供采集器核已提交事件，不授予当前进程所有权。
type RelayControllerHistoryReader interface {
	ListOwnedControllerHistory(context.Context, string) ([]domain.RelayControllerEvidence, error)
}

// RelayDriver is execution only; it cannot authorize, dispatch a task, persist
// state, choose a different Agent, or retry an unknown delivery.
type RelayDriver interface {
	Launch(context.Context, RelayLaunch) (RelayObservation, error)
	Send(context.Context, RelayLaunch, string) (RelayObservation, error)
	Read(context.Context, RelayLaunch) (RelayObservation, error)
	Stop(context.Context, RelayLaunch) (RelayObservation, error)
	Resume(context.Context, RelayLaunch) (RelayObservation, error)
}

type RelayEventSink interface {
	EmitRelayEvent(context.Context, domain.RelayEvent) error
}

// Lifecycle observes or contains already-owned processes only. It cannot
// create an intent, dispatch, resume, or choose a different execution path.
type RelayLifecycle interface {
	ObserveOwned(context.Context) error
	Shutdown(context.Context) error
	ReconcileEvents(context.Context, domain.Caller, string) (domain.RelayTaskView, error)
}

// RelayOwnedSessionReader is an internal, read-only collection seam. A reader
// rechecks permission and binding on every call, cannot start/resume a process,
// and cannot produce action receipts or fresh activity timestamps.
type RelayOwnedSessionReader interface {
	ListOwnedRelaySessions(context.Context) ([]domain.RelayOwnedSession, error)
}

// 只读取已经持久化的派生通知，不调用原生控制器或刷新活动时间。
type RelayOwnedNativeSessionReader interface {
	ListOwnedRelayNativeSessions(context.Context) ([]domain.RelayNativeSessionBatch, error)
}

// RelayOwnedArchiveProvider supplies exact positive native metadata. Absence is
// an explicit coverage gap; callers may not infer roots, filenames or formats.
type RelayOwnedArchiveProvider interface {
	OwnedRelayArchive(context.Context, domain.NativeSession) (domain.RelaySessionArchive, error)
}

// This narrow port reuses the existing ProjectFiles CanonicalRoot operation.
// The adapter validates only registered roots, never archives or native history.
type RelayOwnedRoots interface {
	CanonicalRoot(string) (string, error)
}

type RelayServiceOptions struct {
	Agents     []domain.AgentSpec
	OwnedRoots RelayOwnedRoots
	// No implicit 60-second decision. Zero means no configured live window.
	LivenessWindow time.Duration
	Clock          func() time.Time
}
