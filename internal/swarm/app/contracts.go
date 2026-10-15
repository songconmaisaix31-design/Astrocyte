package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type GrantCheck struct {
	BindingPurpose   string
	Usage            domain.Usage
	ExecutionRoot    string
	BindingIDs       []string
	OperationID      string
	Tools            []string
	ExternalTargets  []string
	ExecutionKnown   bool
	GrantID          string
	Epoch            int
	ProposalRevision int
	ProjectID        string
	Root             string
	Action           string
	Resources        domain.Resources
	Principal        domain.Principal
}

// 回调与授权读取在同一写事务执行，不在回调内调用外部工具。
type GrantAuthorizer interface {
	WithAuthorizedGrant(context.Context, GrantCheck, func(context.Context, domain.Grant) error) error
}

type MissionReadAuthorizer interface {
	WithMissionRead(context.Context, domain.Principal, string, func(context.Context) error) error
}

type TimelineReader interface {
	Timeline(context.Context, string) ([]TimelineEvent, error)
}

type ExecutionHistoryReader interface {
	ExecutionHistory(context.Context, string) ([]ExecutionReference, error)
}

type ExecutionReference struct {
	OperationID      string               `json:"operation_id"`
	SessionID        string               `json:"session_id"`
	WorkItemID       string               `json:"work_item_id"`
	BindingID        string               `json:"binding_id"`
	Fence            int64                `json:"fence"`
	ContextPacketID  string               `json:"context_packet_id"`
	PolicyVersion    string               `json:"policy_version"`
	Events           []TimelineEvent      `json:"events"`
	ToolEventsStatus string               `json:"tool_events_status"`
	ToolEvents       []ToolEventReference `json:"tool_events"`
}

type ToolEventReference struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	Tool       string    `json:"tool"`
	OccurredAt time.Time `json:"occurred_at"`
}

type ToolTraceReader interface {
	ReadToolEventRefs(context.Context, domain.Principal, string, string) (string, []ToolEventReference, error)
}

type AcceptedArtifactEvent struct {
	ActorID    string          `json:"actor_id"`
	ActorKind  string          `json:"actor_kind"`
	EventID    string          `json:"event_id"`
	MissionID  string          `json:"mission_id"`
	ProjectID  string          `json:"project_id"`
	OccurredAt time.Time       `json:"occurred_at"`
	Artifact   domain.Artifact `json:"artifact"`
}

type AcceptedArtifactEventReader interface {
	ReadAcceptedArtifactEvent(context.Context, string) (AcceptedArtifactEvent, error)
}

type AcceptedArtifactEvidenceReader interface {
	ReadAcceptedArtifactEvidence(context.Context, domain.Principal, string, string) (AcceptedArtifactEvent, error)
}

type ExperienceCandidateSource struct {
	LocalProjectID    string
	ID                string
	Version           int
	OwnerProjectID    string
	AcceptanceEventID string
	ArtifactID        string
}
type ExperienceCandidateReader interface {
	ReadArtifactCandidateSource(context.Context, string) (ExperienceCandidateSource, error)
}

type TimelineEvent struct {
	ID          string    `json:"event_id"`
	Type        string    `json:"type"`
	AggregateID string    `json:"aggregate_id"`
	Version     int       `json:"aggregate_version"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type MissionDetail struct {
	MaxActiveExecutors int                  `json:"max_active_executors"`
	Executions         []ExecutionReference `json:"executions"`
	Mission            domain.Mission       `json:"mission"`
	WorkItems          []domain.WorkItem    `json:"work_items"`
	Leases             []domain.Lease       `json:"leases"`
	Artifacts          []domain.Artifact    `json:"artifacts"`
	Timeline           []TimelineEvent      `json:"timeline"`
}

type ServiceOptions struct{ ProjectMaxActive int }

type Member struct {
	Purpose     string
	BindingID   string
	OperationID string
	HolderID    string
	SessionID   string
	ProjectID   string
	Worktree    string
	Version     string
}

type MemberAuthorizer interface {
	CheckMember(context.Context, domain.Principal, Member) error
}

type ExecutionDescriptor struct {
	Capabilities    []string
	Known           bool
	Tools           []string
	ExternalTargets []string
}

type ExecutionDescriptorReader interface {
	DescribeExecution(context.Context, domain.Principal, Member) (ExecutionDescriptor, error)
}

type ArtifactEvidence interface {
	VerifyWorktree(context.Context, string, string) (string, error)
	VerifyArtifact(context.Context, domain.Lease, domain.Artifact) (ArtifactVerification, error)
}

type ArtifactVerification struct {
	ChecksPassed             bool
	ApplicableChecksVerified bool
	ProjectChecksRef         string
}

type ArtifactDiffReader interface {
	ReadVerifiedDiff(context.Context, domain.Artifact) (string, error)
}

type ArtifactDiff struct {
	SchemaVersion int    `json:"schema_version"`
	ArtifactID    string `json:"artifact_id"`
	Commit        string `json:"commit"`
	BaseCommit    string `json:"base_commit"`
	Diff          string `json:"diff"`
}

// 生命周期只对账和停止已绑定的原生身份，不创建替代执行者。
type ExecutionOwnership interface {
	Reconcile(context.Context, string, string) (ExecutionFact, error)
	Stop(context.Context, string, string) (ExecutionFact, error)
}

type ExecutionFact struct {
	FailureCode   string
	CallsKnown    bool
	TokensKnown   bool
	OperationID   string
	SessionID     string
	StopConfirmed bool
	Outcome       string
	Usage         *domain.Usage
}

type Event struct {
	ID          string
	Type        string
	AggregateID string
	Version     int
	At          time.Time
	RequestID   string
	Payload     any
}

type Receipt struct {
	Digest string
	Result []byte
}

type Transaction interface {
	Mission(string) (domain.Mission, error)
	MissionByGrant(string) (domain.Mission, error)
	SaveMission(domain.Mission, int) error
	WorkItem(string) (domain.WorkItem, error)
	WorkItems(string) ([]domain.WorkItem, error)
	SaveWorkItem(domain.WorkItem, int) error
	Lease(string) (domain.Lease, error)
	ActiveLeases(string) ([]domain.Lease, error)
	SaveLease(domain.Lease) error
	DeleteLease(string, int64) error
	Artifact(string) (domain.Artifact, error)
	ArtifactVersion(string, int) (domain.Artifact, error)
	SaveArtifact(domain.Artifact, int) error
	Receipt(string, string, string) (Receipt, bool, error)
	SaveReceipt(string, string, string, Receipt) error
	Event(Event) error
	Consume(string, string) (bool, error)
}

type Repository interface {
	Within(context.Context, func(Transaction) error) error
	List(context.Context) ([]domain.Mission, error)
	Load(context.Context, string) (domain.Mission, error)
}

type CreateMissionCommand struct {
	domain.Meta
	GrantID          string `json:"grant_id"`
	Epoch            int    `json:"epoch"`
	ProposalRevision int    `json:"proposal_revision"`
	ProjectID        string `json:"project_id"`
	Root             string `json:"root"`
	EventID          string `json:"event_id"`
}

type AddWorkItemCommand struct {
	domain.Meta
	ArtifactInputs       []domain.ArtifactReference `json:"artifact_inputs"`
	RequiredCapabilities []string                   `json:"required_capabilities"`
	MissionID            string                     `json:"mission_id"`
	Title                string                     `json:"title"`
	Inputs               []domain.SourceVersion     `json:"inputs"`
	Dependencies         []string                   `json:"dependencies"`
	WritePaths           []string                   `json:"write_paths"`
}

type ClaimCommand struct {
	domain.Meta
	BindingID     string           `json:"binding_id"`
	MissionID     string           `json:"mission_id"`
	WorkItemID    string           `json:"work_item_id"`
	SessionID     string           `json:"session_id"`
	MemberVersion string           `json:"member_version"`
	Worktree      string           `json:"worktree"`
	OperationID   string           `json:"operation_id"`
	Resources     domain.Resources `json:"resources"`
	ExpiresAt     time.Time        `json:"expires_at"`
}

type SubmitCommand struct {
	domain.Meta
	MissionID    string                 `json:"mission_id"`
	WorkItemID   string                 `json:"work_item_id"`
	Fence        int64                  `json:"fence"`
	Commit       string                 `json:"commit"`
	ObjectRef    string                 `json:"object_ref"`
	ChecksRef    string                 `json:"checks_ref"`
	ChangedPaths []string               `json:"changed_paths"`
	Inputs       []domain.SourceVersion `json:"inputs"`
	Usage        domain.Usage           `json:"usage"`
}

type AcceptCommand struct {
	domain.Meta
	MissionID  string `json:"mission_id"`
	ArtifactID string `json:"artifact_id"`
}

type RejectCommand struct {
	domain.Meta
	MissionID  string `json:"mission_id"`
	ArtifactID string `json:"artifact_id"`
	Reason     string `json:"reason"`
}

type TransitionCommand struct {
	domain.Meta
	MissionID string `json:"mission_id"`
	Reason    string `json:"reason"`
}
