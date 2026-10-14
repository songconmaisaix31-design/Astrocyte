package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type GrantCheck struct {
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

type Member struct {
	HolderID  string
	SessionID string
	ProjectID string
	Worktree  string
	Version   string
}

type MemberAuthorizer interface {
	CheckMember(context.Context, domain.Principal, Member) error
}

type ArtifactEvidence interface {
	VerifyWorktree(context.Context, string, string) (string, error)
	VerifyArtifact(context.Context, domain.Lease, domain.Artifact) (ArtifactVerification, error)
}

type ArtifactVerification struct{ ChecksPassed bool }

type ExperienceRecorder interface {
	RecordAccepted(context.Context, domain.Principal, domain.Artifact) (string, error)
}

// 生命周期只对账和停止已绑定的原生身份，不创建替代执行者。
type ExecutionOwnership interface {
	Reconcile(context.Context, string, string) (ExecutionFact, error)
	Stop(context.Context, string, string) (ExecutionFact, error)
}

type ExecutionFact struct {
	OperationID   string
	SessionID     string
	StopConfirmed bool
	Outcome       string
	Usage         *domain.Resources
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
	MissionID    string                 `json:"mission_id"`
	Title        string                 `json:"title"`
	Inputs       []domain.SourceVersion `json:"inputs"`
	Dependencies []string               `json:"dependencies"`
	WritePaths   []string               `json:"write_paths"`
}

type ClaimCommand struct {
	domain.Meta
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
	Usage        domain.Resources       `json:"usage"`
}

type AcceptCommand struct {
	domain.Meta
	MissionID  string `json:"mission_id"`
	ArtifactID string `json:"artifact_id"`
}

type TransitionCommand struct {
	domain.Meta
	MissionID string `json:"mission_id"`
	Reason    string `json:"reason"`
}
