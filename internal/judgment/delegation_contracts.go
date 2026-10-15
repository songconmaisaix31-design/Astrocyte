package judgment

import (
	"context"
	"time"
)

const AdmissionPurpose = "opportunity_admission"

type DelegationRule struct {
	ID                       string        `json:"id"`
	ProjectID                string        `json:"project_id"`
	WorkflowID               string        `json:"workflow_id"`
	OwnerID                  string        `json:"owner_id"`
	Revision                 int           `json:"revision"`
	State                    string        `json:"state"`
	Purpose                  string        `json:"purpose"`
	Question                 Question      `json:"question"`
	PolicyRevision           int           `json:"policy_revision"`
	Language                 string        `json:"language"`
	Materials                []EvidenceRef `json:"materials"`
	AdmissionChoice          *string       `json:"admission_choice"`
	AdmissionValueMin        *float64      `json:"admission_value_min"`
	AdmissionValueMax        *float64      `json:"admission_value_max"`
	AllowOpportunityExternal *bool         `json:"allow_opportunity_external"`
	ReviewConditions         []string      `json:"review_conditions"`
}

type OpportunityEvidence struct {
	Reference EvidenceRef `json:"reference"`
	Locator   string      `json:"locator"`
	Span      *string     `json:"span"`
}

type OpportunitySnapshot struct {
	ID              string                `json:"id"`
	Revision        int                   `json:"revision"`
	Version         int                   `json:"version"`
	State           string                `json:"state"`
	Title           string                `json:"title"`
	NextStep        string                `json:"next_step"`
	GoalRefs        []string              `json:"goal_refs"`
	MissingEvidence []string              `json:"missing_evidence"`
	Evidence        []OpportunityEvidence `json:"evidence"`
}

type DecisionContext struct {
	Purpose     string              `json:"purpose"`
	RuleID      string              `json:"rule_id"`
	RuleVersion int                 `json:"rule_version"`
	WorkflowID  string              `json:"workflow_id"`
	OwnerID     string              `json:"owner_id"`
	Opportunity OpportunitySnapshot `json:"opportunity"`
}

type AdmissionBinding struct {
	ResultID      string          `json:"result_id"`
	ResultVersion int             `json:"result_version"`
	ProjectID     string          `json:"project_id"`
	Actor         Actor           `json:"actor"`
	Context       DecisionContext `json:"context"`
	CreatedAt     time.Time       `json:"created_at"`
}

type DelegationRepository interface {
	WithJudgmentTx(context.Context, func(context.Context) error) error
	LoadDelegationRule(context.Context, string) (DelegationRule, error)
	ListDelegationRules(context.Context, string) ([]DelegationRule, error)
	SaveDelegationRule(context.Context, DelegationRule, int) error
	LoadAdmissionBinding(context.Context, string) (AdmissionBinding, error)
	SaveAdmissionBinding(context.Context, AdmissionBinding) error
}

type OpportunityReader interface {
	ReadJudgmentOpportunity(context.Context, Actor, string, int) (OpportunitySnapshot, error)
}

type DelegationPolicyAccess interface {
	Authorizer
	ReadPolicy(context.Context, Actor, string, QuestionType) (Policy, error)
}

type DelegatedAdmissionCommand struct {
	SchemaVersion              int    `json:"schema_version"`
	RequestID                  string `json:"request_id"`
	ExpectedOpportunityVersion int    `json:"expected_opportunity_version"`
	IdempotencyKey             string `json:"-"`
}

type DelegatedAdmissionReceipt struct {
	AdmissionID      string `json:"admission_id"`
	AdmissionVersion int    `json:"admission_version"`
	ProposalID       string `json:"proposal_id"`
	ProposalVersion  int    `json:"proposal_version"`
}

type DelegatedAdmissions interface {
	AdmitJudgment(context.Context, Actor, string, DelegatedAdmissionCommand) (DelegatedAdmissionReceipt, error)
}
