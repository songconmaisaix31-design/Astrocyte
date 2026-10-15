package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type AdmissionCommand struct {
	CommandMeta
	ExpectedRevision int    `json:"expected_revision"`
	ProjectID        string `json:"project_id,omitempty"`
}

type RevokeAdmissionCommand struct {
	CommandMeta
	Reason string `json:"reason"`
}

type AdmissionDraft struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type AdmissionResult struct {
	SchemaVersion int              `json:"schema_version"`
	Admission     domain.Admission `json:"admission"`
	Draft         *AdmissionDraft  `json:"draft,omitempty"`
}

type AdmittedOpportunitySnapshot struct {
	EventID             string                    `json:"event_id"`
	AdmissionID         string                    `json:"admission_id"`
	AdmissionVersion    int                       `json:"admission_version"`
	OpportunityID       string                    `json:"opportunity_id"`
	OpportunityRevision int                       `json:"opportunity_revision"`
	OwnerNamespace      string                    `json:"owner_namespace"`
	Title               string                    `json:"title"`
	Inputs              []SourceRef               `json:"inputs"`
	GoalRefs            []string                  `json:"goal_refs"`
	ProjectID           string                    `json:"project_id,omitempty"`
	Authority           domain.AdmissionAuthority `json:"authority"`
	PolicyVersion       string                    `json:"policy_version,omitempty"`
	Active              bool                      `json:"active"`
}

type AdmissionTx interface {
	AttentionTx
	Context() context.Context
	LoadAdmission(string) (domain.Admission, error)
	FindAdmission(string, int) (domain.Admission, bool, error)
	SaveAdmission(domain.Admission, int) error
	LoadAdmissionEvent(string) (OutboxEvent, error)
}

type AdmissionRepository interface {
	WithAdmissionTx(context.Context, func(AdmissionTx) error) error
}

type AdmissionDraftConsumer interface {
	ConsumeAdmittedOpportunity(context.Context, string) (AdmissionDraft, error)
}

type AdmissionProjectAccess interface {
	AuthorizeAdmissionProject(context.Context, Principal, string, []SourceRef) error
}

type AdmissionJudgmentRequest struct {
	WorkflowID          string
	DecisionID          string
	OpportunityID       string
	OpportunityRevision int
	ProjectID           string
	Inputs              []SourceRef
}

// 字段来自实际决策与当前委托记录；未知拒答或截断状态不能通过。
type AdmissionJudgmentProof struct {
	WorkflowID          string
	DecisionID          string
	OwnerID             string
	OpportunityID       string
	OpportunityRevision int
	ProjectID           string
	Inputs              []SourceRef
	Status              string
	Decision            string
	QuestionType        string
	Purpose             string
	Backend             string
	Model               string
	ModelVersion        string
	PolicyVersion       string
	Refused             *bool
	Truncated           *bool
	PolicyActive        bool
	CurrentPermissions  bool
	BudgetVerified      bool
}

type AdmissionJudgmentAuthorizer interface {
	AuthorizeAdmissionJudgment(context.Context, AdmissionJudgmentRequest) (AdmissionJudgmentProof, error)
}

type AdmissionOptions struct {
	Drafts   AdmissionDraftConsumer
	Projects AdmissionProjectAccess
	Judgment AdmissionJudgmentAuthorizer
}

type OpportunityAdmissions interface {
	AdmitOpportunity(context.Context, Principal, string, AdmissionCommand) (AdmissionResult, error)
	RevokeAdmission(context.Context, Principal, string, RevokeAdmissionCommand) (AdmissionResult, error)
	GetAdmission(context.Context, Principal, string) (AdmissionResult, error)
}
