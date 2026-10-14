package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type WorkspaceCommandMeta struct {
	SchemaVersion int `json:"schema_version"`
	RequestID string `json:"request_id"`
	ExpectedVersion int `json:"expected_version"`
	IdempotencyKey string `json:"-"`
}

type WorkspaceEvent struct {
	ID string `json:"event_id"`
	Type string `json:"type"`
	SchemaVersion int `json:"schema_version"`
	AggregateID string `json:"aggregate_id"`
	AggregateVersion int `json:"aggregate_version"`
	OccurredAt time.Time `json:"occurred_at"`
	CorrelationID string `json:"correlation_id"`
	CausationID string `json:"causation_id"`
	Payload json.RawMessage `json:"payload"`
}

type WorkspaceReceipt struct { Digest string; Result json.RawMessage }

type WorkspaceTx interface {
	Context() context.Context
	LoadGoal(string,int) (domain.Goal,error)
	SaveGoal(domain.Goal,int) error
	LoadProposal(string,int) (domain.Proposal,error)
	SaveProposal(domain.Proposal,int) error
	FindAdmission(string,int) (domain.Proposal,error)
	LoadApproval(string) (domain.ApprovalGrant,error)
	SaveApproval(domain.ApprovalGrant,int) error
	LoadProject(string) (domain.LocalProject,error)
	LoadProjectGrant(string,string) (domain.ProjectGrant,error)
	LoadReceipt(string,string,string) (WorkspaceReceipt,error)
	SaveReceipt(string,string,string,WorkspaceReceipt) error
	AppendWorkspaceEvent(WorkspaceEvent) error
}

type WorkspaceRepository interface {
	WithWorkspaceTx(context.Context,func(WorkspaceTx)error) error
	ListWorkspaceProposals(context.Context,string) ([]domain.Proposal,error)
	ReadWorkspaceEvents(context.Context,string,int) ([]WorkspaceEvent,error)
}

// 回调使用同一事务context；消费者的SQLite adapter必须复用其连接。
type GrantAuthorizer interface {
	QueryGrant(context.Context,string) (domain.ApprovalGrant,error)
	WithAuthorizedGrant(context.Context,domain.GrantCheck,func(context.Context,domain.ApprovalGrant)error) error
}

type CreateProposalCommand struct { WorkspaceCommandMeta; Content domain.ProposalContent `json:"content"` }
type ApproveProposalCommand struct { WorkspaceCommandMeta; ExpectedRevision int `json:"expected_revision"`; ExpiresAt time.Time `json:"expires_at"` }
type ChangeApprovalCommand struct { WorkspaceCommandMeta; Reason string `json:"reason"` }
type GoalCommand struct { WorkspaceCommandMeta; Goal domain.Goal `json:"goal"` }
type AdmissionProposalCommand struct {
	EventID string `json:"event_id"`
	AdmissionID string `json:"admission_id"`
	OpportunityID string `json:"opportunity_id"`
	OpportunityRevision int `json:"opportunity_revision"`
	Content domain.ProposalContent `json:"content"`
}
