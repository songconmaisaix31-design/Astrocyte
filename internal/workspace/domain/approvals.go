package domain

import (
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type ApprovalGrant struct {
	ID string `json:"id"`
	ProposalID string `json:"proposal_id"`
	ProposalRevision int `json:"proposal_revision"`
	Goal GoalReference `json:"goal"`
	Epoch int `json:"epoch"`
	Status string `json:"status"`
	Scope ExecutionScope `json:"scope"`
	Budget ResourceBudget `json:"budget"`
	Stops StopConditions `json:"stops"`
	ApproverID string `json:"approver_id"`
	ApprovedAt time.Time `json:"approved_at"`
	ExpiresAt time.Time `json:"expires_at"`
	ChangedAt time.Time `json:"changed_at"`
	Reason string `json:"reason,omitempty"`
}

// Used 包含本次动作预留后的累计用量，由执行事务核实。
type GrantCheck struct {
	GrantID string `json:"grant_id"`
	Epoch int `json:"epoch"`
	ProposalRevision int `json:"proposal_revision"`
	ProjectID string `json:"project_id"`
	Root string `json:"root"`
	Action string `json:"action"`
	Tools []string `json:"tools"`
	ExternalTargets []string `json:"external_targets"`
	Concurrency int `json:"concurrency"`
	Used ResourceBudget `json:"used"`
}

func (g ApprovalGrant) Check(c GrantCheck, now time.Time) error {
	if g.Status!="active" || !g.ExpiresAt.After(now) || !g.Stops.Deadline.After(now) { return WorkspaceError(apierrors.ApprovalRevoked,"approval is revoked, expired or completed") }
	if c.GrantID!=g.ID || c.Epoch!=g.Epoch || c.ProposalRevision!=g.ProposalRevision { return WorkspaceError(apierrors.VersionConflict,"approval epoch or proposal revision changed") }
	if c.ProjectID!=g.Scope.ProjectID || !slices.Contains(g.Scope.Actions,c.Action) || c.Concurrency<1 || c.Concurrency>g.Scope.MaxConcurrency { return WorkspaceError(apierrors.ScopeDenied,"protected action exceeds approved project, action or concurrency") }
	if !filepath.IsAbs(c.Root) { return WorkspaceError(apierrors.ScopeDenied,"protected root must be absolute") }
	rootAllowed:=false
	for _, root:=range g.Scope.Roots {
		rel,err:=filepath.Rel(root,c.Root)
		if err==nil && rel!=".." && !strings.HasPrefix(rel,".."+string(filepath.Separator)) && !filepath.IsAbs(rel) { rootAllowed=true }
	}
	if !rootAllowed { return WorkspaceError(apierrors.ScopeDenied,"protected root is outside approval") }
	for _, tool:=range c.Tools { if !slices.Contains(g.Scope.Tools,tool) { return WorkspaceError(apierrors.ScopeDenied,"tool is outside approval") } }
	for _, target:=range c.ExternalTargets { if !slices.Contains(g.Scope.ExternalTargets,target) { return WorkspaceError(apierrors.ScopeDenied,"external target is outside approval") } }
	if err:=c.Used.Validate(); err!=nil { return err }
	limits:=[]*int64{g.Budget.Calls,g.Budget.Tokens,g.Budget.MoneyMinor}
	used:=[]*int64{c.Used.Calls,c.Used.Tokens,c.Used.MoneyMinor}
	for i,limit:=range limits { if limit!=nil && (used[i]==nil || *used[i]>*limit) { return WorkspaceError(apierrors.BudgetExhausted,"protected action lacks usage evidence or exceeds budget") } }
	if g.Budget.MoneyMinor!=nil && c.Used.Currency!=g.Budget.Currency { return WorkspaceError(apierrors.ScopeDenied,"budget currency differs") }
	return nil
}

type ProposalApprovedEvent struct {
	ProposalID string `json:"proposal_id"`
	ProposalRevision int `json:"proposal_revision"`
	GrantID string `json:"grant_id"`
	Epoch int `json:"epoch"`
}
