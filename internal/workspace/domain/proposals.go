package domain

import (
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type ProposalStatus string

const (
	ProposalDraft ProposalStatus = "draft"
	ProposalInReview ProposalStatus = "in_review"
	ProposalApproved ProposalStatus = "approved"
	ProposalDeclined ProposalStatus = "declined"
	ProposalSuperseded ProposalStatus = "superseded"
)

type GoalReference struct {
	ID string `json:"id"`
	Revision int `json:"revision"`
}

type ResourceBudget struct {
	Calls *int64 `json:"calls"`
	Tokens *int64 `json:"tokens"`
	MoneyMinor *int64 `json:"money_minor"`
	Currency string `json:"currency"`
}

type ExecutionScope struct {
	ProjectID string `json:"project_id"`
	Roots []string `json:"roots"`
	Actions []string `json:"actions"`
	Tools []string `json:"tools"`
	ExternalTargets []string `json:"external_targets"`
	MaxConcurrency int `json:"max_concurrency"`
}

type StopConditions struct {
	Deadline time.Time `json:"deadline"`
	Conditions []string `json:"conditions"`
	MaxWorkItems int `json:"max_work_items"`
	MaxAttempts int `json:"max_attempts"`
	NoProgressSeconds int `json:"no_progress_seconds"`
}

type ProposalContent struct {
	Title string `json:"title"`
	Goal GoalReference `json:"goal"`
	Inputs []FixedReference `json:"inputs"`
	Scope ExecutionScope `json:"scope"`
	Budget ResourceBudget `json:"budget"`
	Stops StopConditions `json:"stops"`
	Deliverables []string `json:"deliverables"`
	SuccessCriteria []string `json:"success_criteria"`
}

type Proposal struct {
	ID string `json:"id"`
	Version int `json:"version"`
	Revision int `json:"revision"`
	Status ProposalStatus `json:"status"`
	Content ProposalContent `json:"content"`
	AdmissionID string `json:"admission_id,omitempty"`
	OpportunityID string `json:"opportunity_id,omitempty"`
	OpportunityRevision int `json:"opportunity_revision,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func WorkspaceError(code apierrors.Code, message string) error {
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: "review_workspace_version_and_permissions"}
}

func (b ResourceBudget) Validate() error {
	for _, limit := range []*int64{b.Calls,b.Tokens,b.MoneyMinor} {
		if limit != nil && *limit < 0 { return WorkspaceError(apierrors.ValidationFailed,"resource limits cannot be negative") }
	}
	if b.MoneyMinor != nil && (len(b.Currency)!=3 || strings.ToUpper(b.Currency)!=b.Currency) { return WorkspaceError(apierrors.ValidationFailed,"money requires an uppercase three-letter currency") }
	if b.MoneyMinor == nil && b.Currency != "" { return WorkspaceError(apierrors.ValidationFailed,"currency requires a money budget") }
	return nil
}

func (c ProposalContent) Validate(now time.Time, executable bool) error {
	if strings.TrimSpace(c.Title)=="" || c.Goal.ID=="" || c.Goal.Revision<1 || c.Scope.ProjectID=="" { return WorkspaceError(apierrors.ValidationFailed,"title, fixed goal and project are required") }
	if err:=c.Budget.Validate(); err!=nil { return err }
	for _, ref:=range c.Inputs { if ref.MaterialID=="" || ref.Revision<1 { return WorkspaceError(apierrors.ValidationFailed,"input references require a fixed material revision") } }
	if !executable { return nil }
	if len(c.Scope.Roots)==0 || len(c.Scope.Actions)==0 || c.Scope.MaxConcurrency<1 || len(c.Deliverables)==0 || len(c.SuccessCriteria)==0 || len(c.Stops.Conditions)==0 || c.Stops.MaxWorkItems<1 || c.Stops.MaxAttempts<1 || c.Stops.NoProgressSeconds<1 || !c.Stops.Deadline.After(now) { return WorkspaceError(apierrors.ValidationFailed,"execution requires roots, actions, concurrency, outputs, success and stopping conditions") }
	if c.Budget.Calls==nil && c.Budget.Tokens==nil && c.Budget.MoneyMinor==nil { return WorkspaceError(apierrors.ValidationFailed,"execution requires an explicit resource limit") }
	for _, root:=range c.Scope.Roots { if !filepath.IsAbs(root) { return WorkspaceError(apierrors.ValidationFailed,"allowed roots must be absolute") } }
	for _, action:=range c.Scope.Actions { if strings.TrimSpace(action)=="" { return WorkspaceError(apierrors.ValidationFailed,"allowed actions cannot be empty") } }
	return nil
}

func (p *Proposal) Transition(expected int, target ProposalStatus, now time.Time) error {
	if p.Version!=expected { return WorkspaceError(apierrors.VersionConflict,"proposal version changed") }
	valid:=p.Status==ProposalDraft && target==ProposalInReview || p.Status==ProposalInReview && slices.Contains([]ProposalStatus{ProposalApproved,ProposalDeclined},target)
	if !valid { return WorkspaceError(apierrors.ValidationFailed,"proposal transition is not allowed") }
	if target==ProposalInReview || target==ProposalApproved { if err:=p.Content.Validate(now,true); err!=nil { return err } }
	p.Status=target; p.Version++; p.UpdatedAt=now
	return nil
}
