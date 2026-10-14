package domain

import (
	"path/filepath"
	"slices"
	"strings"
	"time"

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
	SettingsRevision int `json:"settings_revision"`
	Roots []string `json:"roots"`
	Actions []string `json:"actions"`
	Tools []string `json:"tools"`
	ExternalTargets []string `json:"external_targets"`
	MaxConcurrency int `json:"max_concurrency"`
	WritePaths []string `json:"write_paths"`
}

type StopConditions struct {
	Deadline time.Time `json:"deadline"`
	Conditions []string `json:"conditions"`
	MaxWorkItems int `json:"max_work_items"`
	MaxAttempts int `json:"max_attempts"`
	NoProgressSeconds int `json:"no_progress_seconds"`
	MaxNoProgressRounds int `json:"max_no_progress_rounds"`
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
	GrantID string `json:"grant_id,omitempty"`
	OwnerNamespace string `json:"owner_namespace,omitempty"`
	AssignmentStatus string `json:"assignment_status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 由Attention可信应用端口提供，包含真实准入和当前访问校验。
type AdmittedOpportunitySnapshot struct {
	EventID             string           `json:"event_id"`
	AdmissionID         string           `json:"admission_id"`
	OpportunityID       string           `json:"opportunity_id"`
	OpportunityRevision int              `json:"opportunity_revision"`
	OwnerNamespace      string           `json:"owner_namespace"`
	Title               string           `json:"title"`
	Inputs              []FixedReference `json:"inputs"`
	ProjectID           string           `json:"project_id,omitempty"`
	Authority           string           `json:"authority"`
	PolicyVersion       string           `json:"policy_version"`
	Active              bool             `json:"active"`
}

type WorkspaceRuleError struct { Code string; Message string }
func (e *WorkspaceRuleError) Error() string { return e.Code+": "+e.Message }
func WorkspaceError[C ~string](code C, message string) error {
	return &WorkspaceRuleError{Code: string(code), Message: message}
}

func (b ResourceBudget) Validate() error {
	for _, limit := range []*int64{b.Calls,b.Tokens,b.MoneyMinor} {
		if limit != nil && *limit < 0 { return WorkspaceError("validation_failed","resource limits cannot be negative") }
	}
	if b.MoneyMinor != nil && (len(b.Currency)!=3 || strings.ToUpper(b.Currency)!=b.Currency) { return WorkspaceError("validation_failed","money requires an uppercase three-letter currency") }
	if b.MoneyMinor == nil && b.Currency != "" { return WorkspaceError("validation_failed","currency requires a money budget") }
	return nil
}

func (c ProposalContent) Validate(now time.Time, executable bool) error {
	if strings.TrimSpace(c.Title)=="" || c.Goal.ID=="" || c.Goal.Revision<1 || c.Scope.ProjectID=="" { return WorkspaceError("validation_failed","title, fixed goal and project are required") }
	if err:=c.Budget.Validate(); err!=nil { return err }
	for _, ref:=range c.Inputs { if ref.MaterialID=="" || ref.Revision<1 { return WorkspaceError("validation_failed","input references require a fixed material revision") } }
	if !executable { return nil }
	if len(c.Scope.Roots)==0 || len(c.Scope.Actions)==0 || c.Scope.MaxConcurrency<1 || len(c.Deliverables)==0 || len(c.SuccessCriteria)==0 || len(c.Stops.Conditions)==0 || c.Stops.MaxWorkItems<1 || c.Stops.MaxAttempts<1 || c.Stops.MaxNoProgressRounds<1 || c.Stops.NoProgressSeconds<0 || !c.Stops.Deadline.After(now) { return WorkspaceError("validation_failed","execution requires roots, actions, concurrency, outputs, success and stopping conditions") }
	if c.Budget.Calls==nil && c.Budget.Tokens==nil && c.Budget.MoneyMinor==nil { return WorkspaceError("validation_failed","execution requires an explicit resource limit") }
	for _, root:=range c.Scope.Roots { if !filepath.IsAbs(root) { return WorkspaceError("validation_failed","allowed roots must be absolute") } }
	for _, path:=range c.Scope.WritePaths { if filepath.IsAbs(path)||filepath.Clean(path)==".."||strings.HasPrefix(filepath.Clean(path),".."+string(filepath.Separator)) { return WorkspaceError("validation_failed","write paths must be relative to approved roots") } }
	for _, action:=range c.Scope.Actions { if strings.TrimSpace(action)=="" { return WorkspaceError("validation_failed","allowed actions cannot be empty") } }
	return nil
}

func (p *Proposal) Transition(expected int, target ProposalStatus, now time.Time) error {
	if p.Version!=expected { return WorkspaceError("version_conflict","proposal version changed") }
	valid:=p.Status==ProposalDraft && target==ProposalInReview || p.Status==ProposalInReview && slices.Contains([]ProposalStatus{ProposalApproved,ProposalDeclined},target)
	if !valid { return WorkspaceError("validation_failed","proposal transition is not allowed") }
	if target==ProposalInReview || target==ProposalApproved { if err:=p.Content.Validate(now,true); err!=nil { return err } }
	p.Status=target; p.Version++; p.UpdatedAt=now
	return nil
}
