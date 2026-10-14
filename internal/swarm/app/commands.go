package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type Service struct {
	repo      Repository
	grants    GrantAuthorizer
	members   MemberAuthorizer
	evidence  ArtifactEvidence
	ownership ExecutionOwnership
}

func NewSwarmLedger(repo Repository, grants GrantAuthorizer) (*Service, error) {
	if repo == nil || grants == nil {
		return nil, domain.Failure(apierrors.ValidationFailed, "actual storage and Workspace approval service required")
	}
	return &Service{repo: repo, grants: grants}, nil
}

func NewSwarmService(repo Repository, grants GrantAuthorizer, members MemberAuthorizer, evidence ArtifactEvidence, ownership ExecutionOwnership) (*Service, error) {
	if repo == nil || grants == nil || members == nil || evidence == nil || ownership == nil {
		return nil, domain.Failure(apierrors.ValidationFailed, "swarm requires storage, current authorization, member binding, Git/object evidence and native ownership")
	}
	return &Service{repo: repo, grants: grants, members: members, evidence: evidence, ownership: ownership}, nil
}

func (s *Service) ListMissions(ctx context.Context) (apierrors.ListResult, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return apierrors.ListResult{}, err
	}
	out := apierrors.EmptyList()
	for _, m := range rows {
		out.Items = append(out.Items, m)
	}
	return out, nil
}
func (s *Service) GetMission(ctx context.Context, id string) (any, error) {
	return s.repo.Load(ctx, id)
}

func validateMeta(p domain.Principal, m domain.Meta) error {
	if p.ID == "" || (p.Kind != "human" && p.Kind != "agent" && p.Kind != "consumer") || m.SchemaVersion != 1 || m.RequestID == "" || m.IdempotencyKey == "" || m.ExpectedVersion < 0 {
		return domain.Failure(apierrors.ValidationFailed, "authenticated identity and command metadata are required")
	}
	return nil
}
func human(p domain.Principal) error {
	if p.Kind != "human" {
		return domain.Failure(apierrors.ScopeDenied, "human control session required")
	}
	return nil
}
func missing(err error) bool {
	var e *apierrors.ServiceError
	return errors.As(err, &e) && e.Code == apierrors.NotFound
}
func digest(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	d := sha256.Sum256(b)
	return hex.EncodeToString(d[:]), nil
}
func actor(p domain.Principal) string { return p.Kind + ":" + p.ID }

func resourcesEqual(a, b domain.Resources) bool {
	return a.Calls == b.Calls && a.Tokens == b.Tokens && a.Currency == b.Currency && ((a.Money == nil && b.Money == nil) || (a.Money != nil && b.Money != nil && *a.Money == *b.Money))
}

func settledOutcome(outcome string) bool {
	return slices.Contains([]string{"completed", "failed", "cancelled", "stopped", "not_started"}, outcome)
}
func safeFailureCode(code string) bool {
	return slices.Contains([]string{"", "unknown", "native_failure", "validation_failed", "unsupported_capability", "scope_denied", "approval_required", "approval_revoked", "lease_lost", "budget_exhausted", "provider_unavailable", "delivery_unknown", "evidence_missing", "context_stale", "version_conflict", "cancelled"}, code)
}

type executionCheckKey struct{}
type executionCheck struct {
	operationID string
	descriptor  ExecutionDescriptor
	bindingID   string
}

func (s *Service) execution(ctx context.Context, p domain.Principal, m Member) (ExecutionDescriptor, error) {
	if s.members == nil {
		return ExecutionDescriptor{}, domain.Failure(apierrors.UnsupportedCapability, "actual native member binding is not configured")
	}
	if err := s.members.CheckMember(ctx, p, m); err != nil {
		return ExecutionDescriptor{}, err
	}
	reader, ok := s.members.(ExecutionDescriptorReader)
	if !ok {
		return ExecutionDescriptor{}, domain.Failure(apierrors.UnsupportedCapability, "native binding does not expose its actual execution policy")
	}
	d, err := reader.DescribeExecution(ctx, p, m)
	if err != nil {
		return d, err
	}
	if !d.Known {
		return d, domain.Failure(apierrors.EvidenceMissing, "actual native tools and external targets remain unknown")
	}
	return d, nil
}

func replay[T any](tx Transaction, p domain.Principal, name string, m domain.Meta, command any) (T, bool, string, error) {
	var out T
	d, err := digest(command)
	if err != nil {
		return out, false, "", err
	}
	r, found, err := tx.Receipt(actor(p), name, m.IdempotencyKey)
	if err != nil {
		return out, false, "", err
	}
	if !found {
		return out, false, d, nil
	}
	if r.Digest != d {
		return out, false, d, domain.Failure(apierrors.VersionConflict, "idempotency key was already used with different input")
	}
	err = json.Unmarshal(r.Result, &out)
	return out, true, d, err
}
func finish(tx Transaction, p domain.Principal, name string, m domain.Meta, d string, result any, event string, id string, version int) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err = tx.SaveReceipt(actor(p), name, m.IdempotencyKey, Receipt{Digest: d, Result: b}); err != nil {
		return err
	}
	if event == "" {
		return nil
	}
	return tx.Event(Event{ID: uuid.NewString(), Type: "swarm." + event, AggregateID: id, Version: version, At: time.Now().UTC(), RequestID: m.RequestID, Payload: result})
}
func checkVersion(current, expected int) error {
	if current != expected {
		return domain.Failure(apierrors.VersionConflict, "aggregate version changed")
	}
	return nil
}

func (s *Service) authorized(ctx context.Context, p domain.Principal, id, action string, resources domain.Resources, fn func(context.Context, domain.Grant) error) error {
	m, err := s.repo.Load(ctx, id)
	if err != nil {
		return err
	}
	if p.Kind == "agent" && p.ProjectID != m.Grant.ProjectID {
		return domain.Failure(apierrors.ScopeDenied, "agent belongs to a different project")
	}
	check := GrantCheck{GrantID: m.Grant.ID, Epoch: m.Grant.Epoch, ProposalRevision: m.Grant.ProposalRevision, ProjectID: m.Grant.ProjectID, Root: m.Grant.Root, Action: action, Resources: resources, Principal: p}
	if e, ok := ctx.Value(executionCheckKey{}).(executionCheck); ok {
		check.OperationID = e.operationID
		check.BindingIDs = []string{e.bindingID}
		check.ExecutionKnown = e.descriptor.Known
		check.Tools = e.descriptor.Tools
		check.ExternalTargets = e.descriptor.ExternalTargets
	}
	return s.grants.WithAuthorizedGrant(ctx, check, fn)
}

func (s *Service) CreateMission(ctx context.Context, p domain.Principal, c CreateMissionCommand) (domain.Mission, error) {
	var out domain.Mission
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if err := human(p); err != nil && p.Kind != "consumer" {
		return out, err
	}
	err := s.grants.WithAuthorizedGrant(ctx, GrantCheck{GrantID: c.GrantID, Epoch: c.Epoch, ProposalRevision: c.ProposalRevision, ProjectID: c.ProjectID, Root: c.Root, Action: "create_mission", Principal: p}, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.Mission](tx, p, "create_mission", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			if c.Meta.ExpectedVersion != 0 {
				return domain.Failure(apierrors.VersionConflict, "new mission expected version must be zero")
			}
			if g.ID != c.GrantID || g.Epoch != c.Epoch || g.ProposalRevision != c.ProposalRevision || g.ProjectID != c.ProjectID || g.Root != c.Root {
				return domain.Failure(apierrors.ApprovalRevoked, "approval snapshot changed")
			}
			out, err = tx.MissionByGrant(g.ID)
			if err != nil && !missing(err) {
				return err
			}
			created := false
			if missing(err) {
				created = true
				out, err = domain.NewMission(uuid.NewString(), g, time.Now().UTC())
				if err != nil {
					return err
				}
				if err = tx.SaveMission(out, 0); err != nil {
					return err
				}
			} else if out.Grant.Root != g.Root || out.Grant.Epoch != g.Epoch || out.Grant.ProposalRevision != g.ProposalRevision {
				return domain.Failure(apierrors.VersionConflict, "existing mission has a different fixed approval scope")
			}
			if c.EventID != "" {
				if _, err := tx.Consume("swarm.create_mission", c.EventID); err != nil {
					return err
				}
			}
			event := ""
			if created {
				event = "mission_created"
			}
			return finish(tx, p, "create_mission", c.Meta, d, out, event, out.ID, out.Version)
		})
	})
	return out, err
}

func (s *Service) AddWorkItem(ctx context.Context, p domain.Principal, c AddWorkItemCommand) (domain.WorkItem, error) {
	var out domain.WorkItem
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	err := s.authorized(ctx, p, c.MissionID, "manage_work_items", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		if len(c.ArtifactInputs) > 0 {
			if err := s.grants.WithAuthorizedGrant(ctx, GrantCheck{GrantID: g.ID, Epoch: g.Epoch, ProposalRevision: g.ProposalRevision, ProjectID: g.ProjectID, Root: g.Root, Action: "read_swarm", Principal: p}, func(context.Context, domain.Grant) error { return nil }); err != nil {
				return err
			}
		}
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.WorkItem](tx, p, "add_work_item", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			m, err := tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			if err = checkVersion(m.Version, c.Meta.ExpectedVersion); err != nil {
				return err
			}
			if err = m.CanExecute(time.Now().UTC()); err != nil {
				return err
			}
			if m.WorkItemCount >= g.Limits.MaxWorkItems {
				return domain.Failure(apierrors.BudgetExhausted, "mission work item limit reached")
			}
			out = domain.WorkItem{ID: uuid.NewString(), MissionID: m.ID, Version: 1, State: "open", Title: c.Title, Inputs: c.Inputs, Dependencies: c.Dependencies, WritePaths: c.WritePaths, RequiredCapabilities: c.RequiredCapabilities}
			for _, input := range c.Inputs {
				if !slices.Contains(g.Inputs, input) {
					return domain.Failure(apierrors.ScopeDenied, "work item input is not a fixed approved proposal source")
				}
			}
			for _, ref := range c.ArtifactInputs {
				a, err := tx.ArtifactVersion(ref.ID, ref.Version)
				if err != nil {
					return err
				}
				if a.State != "accepted" || a.AcceptanceID == "" || a.Commit != ref.Commit || a.ObjectRef != ref.ObjectRef {
					return domain.Failure(apierrors.ContextStale, "source artifact does not match the fixed human-accepted version")
				}
				sourceWork, err := tx.WorkItem(a.WorkItemID)
				if err != nil {
					return err
				}
				sourceMission, err := tx.Mission(sourceWork.MissionID)
				if err != nil {
					return err
				}
				if sourceMission.Grant.ProjectID != g.ProjectID {
					return domain.Failure(apierrors.ScopeDenied, "cross-project artifact sharing is not authorized")
				}
				out.ArtifactInputs = append(out.ArtifactInputs, ref)
			}
			if err = out.Validate(m); err != nil {
				return err
			}
			for _, id := range out.Dependencies {
				v, err := tx.WorkItem(id)
				if err != nil {
					return err
				}
				if v.MissionID != m.ID {
					return domain.Failure(apierrors.ScopeDenied, "dependencies must belong to this mission")
				}
			}
			old := m.Version
			m.Version++
			m.WorkItemCount++
			if err = tx.SaveMission(m, old); err != nil {
				return err
			}
			if err = tx.SaveWorkItem(out, 0); err != nil {
				return err
			}
			return finish(tx, p, "add_work_item", c.Meta, d, out, "work_item_created", out.ID, out.Version)
		})
	})
	return out, err
}

func (s *Service) ClaimWorkItem(ctx context.Context, p domain.Principal, c ClaimCommand) (domain.Lease, error) {
	var out domain.Lease
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if p.Kind != "agent" {
		return out, domain.Failure(apierrors.ScopeDenied, "bound agent identity required")
	}
	if s.members == nil || s.evidence == nil || s.ownership == nil {
		return out, domain.Failure(apierrors.UnsupportedCapability, "real native binding, evidence and ownership are not configured")
	}
	if _, err := uuid.Parse(c.OperationID); err != nil {
		return out, domain.Failure(apierrors.ValidationFailed, "stable UUID operation identity required")
	}
	if err := c.Resources.Validate(); err != nil {
		return out, err
	}
	if c.Resources.Calls < 1 || c.SessionID == "" || c.MemberVersion == "" || c.BindingID == "" {
		return out, domain.Failure(apierrors.ValidationFailed, "bound session, member version and a call reservation required")
	}
	m, err := s.repo.Load(ctx, c.MissionID)
	if err != nil {
		return out, err
	}
	replayed := false
	var descriptor ExecutionDescriptor
	err = s.authorized(ctx, p, c.MissionID, "read_swarm", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		var err error
		descriptor, err = s.execution(ctx, p, Member{BindingID: c.BindingID, OperationID: c.OperationID, HolderID: p.ID, SessionID: c.SessionID, ProjectID: g.ProjectID, Worktree: c.Worktree, Version: c.MemberVersion})
		if err != nil {
			return err
		}
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, _, err := replay[domain.Lease](tx, p, "claim_work_item", c.Meta, c)
			if err != nil {
				return err
			}
			if !found {
				return nil
			}
			current, err := tx.Lease(c.WorkItemID)
			if err != nil {
				return err
			}
			if err = current.Check(p.ID, cached.Fence, time.Now().UTC()); err != nil {
				return err
			}
			out = cached
			replayed = true
			return nil
		})
	})
	if err != nil || replayed {
		return out, err
	}
	ctx = context.WithValue(ctx, executionCheckKey{}, executionCheck{operationID: c.OperationID, descriptor: descriptor, bindingID: c.BindingID})
	baseCommit, err := s.evidence.VerifyWorktree(ctx, m.Grant.Root, c.Worktree)
	if err != nil {
		return out, err
	}
	err = s.authorized(ctx, p, c.MissionID, "claim_work_item", c.Resources, func(ctx context.Context, g domain.Grant) error {
		binding, err := authorizedBinding(ctx, c.BindingID, c.MemberVersion)
		if err != nil {
			return err
		}
		return s.repo.Within(ctx, func(tx Transaction) error {
			if err := s.members.CheckMember(ctx, p, Member{BindingID: c.BindingID, OperationID: c.OperationID, HolderID: p.ID, SessionID: c.SessionID, ProjectID: g.ProjectID, Worktree: c.Worktree, Version: c.MemberVersion}); err != nil {
				return err
			}
			cached, found, d, err := replay[domain.Lease](tx, p, "claim_work_item", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				current, e := tx.Lease(c.WorkItemID)
				if e != nil {
					return e
				}
				if err = current.Check(p.ID, out.Fence, time.Now().UTC()); err != nil {
					return err
				}
				return nil
			}
			m, err := tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err = m.CanExecute(now); err != nil {
				return err
			}
			w, err := tx.WorkItem(c.WorkItemID)
			if err != nil {
				return err
			}
			if w.MissionID != m.ID {
				return domain.Failure(apierrors.ScopeDenied, "work item belongs to another mission")
			}
			if err = checkVersion(w.Version, c.Meta.ExpectedVersion); err != nil {
				return err
			}
			if w.State != "open" && w.State != "revision_needed" {
				return domain.Failure(apierrors.LeaseLost, "work item is unavailable")
			}
			if w.Attempts >= g.Limits.MaxAttempts {
				return domain.Failure(apierrors.BudgetExhausted, "work item attempt limit reached")
			}
			for _, capability := range w.RequiredCapabilities {
				if !slices.Contains(descriptor.Capabilities, capability) {
					return domain.Failure(apierrors.UnsupportedCapability, "bound agent lacks a required capability")
				}
			}
			if !now.Before(c.ExpiresAt) || c.ExpiresAt.After(g.ExpiresAt) || c.ExpiresAt.After(g.Limits.Deadline) {
				return domain.Failure(apierrors.ValidationFailed, "lease expiry exceeds authorized execution window")
			}
			for _, id := range w.Dependencies {
				v, err := tx.WorkItem(id)
				if err != nil {
					return err
				}
				if v.State != "accepted" {
					return domain.Failure(apierrors.VersionConflict, "work item dependency is not accepted")
				}
			}
			leases, err := tx.ActiveLeases(g.ProjectID)
			if err != nil {
				return err
			}
			limit := g.Limits.MaxActive
			if len(leases) >= limit {
				return domain.Failure(apierrors.BudgetExhausted, "project active executor limit reached")
			}
			for _, l := range leases {
				if l.Worktree == c.Worktree || domain.Overlaps(l.WritePaths, w.WritePaths) {
					return domain.Failure(apierrors.ScopeDenied, "an existing executor occupies this worktree or write scope")
				}
				if l.HolderID == p.ID || l.SessionID == c.SessionID || l.OperationID == c.OperationID {
					return domain.Failure(apierrors.LeaseLost, "executor or operation is already leased")
				}
			}
			reserved, err := m.Reserved.Add(c.Resources)
			if err != nil {
				return err
			}
			total, err := reserved.Add(m.Used)
			if err != nil {
				return err
			}
			if !total.Within(g.Budget) {
				return domain.Failure(apierrors.BudgetExhausted, "mission budget cannot cover reservation")
			}
			if w.Fence == math.MaxInt64 {
				return domain.Failure(apierrors.LeaseLost, "fencing token exhausted")
			}
			oldW := w.Version
			w.Version++
			w.Fence++
			w.Attempts++
			w.State = "leased"
			oldM := m.Version
			m.Version++
			m.State = "running"
			m.Reserved = reserved
			out = domain.Lease{ContextPacketID: binding.PacketID, BindingID: c.BindingID, WorkItemID: w.ID, MissionID: m.ID, ProjectID: g.ProjectID, HolderID: p.ID, Fence: w.Fence, OperationID: c.OperationID, SessionID: c.SessionID, MemberVersion: c.MemberVersion, Worktree: c.Worktree, BaseCommit: baseCommit, WritePaths: w.WritePaths, Reserved: c.Resources, ExpiresAt: c.ExpiresAt, PolicyVersion: domain.LocalPolicyVersion}
			if err = tx.SaveMission(m, oldM); err != nil {
				return err
			}
			if err = tx.SaveWorkItem(w, oldW); err != nil {
				return err
			}
			if err = tx.SaveLease(out); err != nil {
				return err
			}
			return finish(tx, p, "claim_work_item", c.Meta, d, out, "work_item_claimed", w.ID, w.Version)
		})
	})
	return out, err
}

func (s *Service) SubmitArtifact(ctx context.Context, p domain.Principal, c SubmitCommand) (domain.Artifact, error) {
	var out domain.Artifact
	if s.members == nil || s.evidence == nil || s.ownership == nil {
		return out, domain.Failure(apierrors.UnsupportedCapability, "real native binding, evidence and ownership are not configured")
	}
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if p.Kind != "agent" {
		return out, domain.Failure(apierrors.ScopeDenied, "bound agent identity required")
	}
	var lease domain.Lease
	replayed := false
	var descriptor ExecutionDescriptor
	err := s.authorized(ctx, p, c.MissionID, "read_swarm", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, _, err := replay[domain.Artifact](tx, p, "submit_artifact", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				replayed = true
				return nil
			}
			lease, err = tx.Lease(c.WorkItemID)
			if err != nil {
				return err
			}
			if lease.MissionID != c.MissionID {
				return domain.Failure(apierrors.ScopeDenied, "lease belongs to a different mission")
			}
			descriptor, err = s.execution(ctx, p, Member{BindingID: lease.BindingID, OperationID: lease.OperationID, HolderID: p.ID, SessionID: lease.SessionID, ProjectID: g.ProjectID, Worktree: lease.Worktree, Version: lease.MemberVersion})
			return err
		})
	})
	if err != nil || replayed {
		return out, err
	}
	ctx = context.WithValue(ctx, executionCheckKey{}, executionCheck{operationID: lease.OperationID, descriptor: descriptor, bindingID: lease.BindingID})
	if err = lease.Check(p.ID, c.Fence, time.Now().UTC()); err != nil {
		return out, err
	}
	out = domain.Artifact{ContextPacketID: lease.ContextPacketID, ID: uuid.NewString(), Version: 1, WorkItemID: c.WorkItemID, OperationID: lease.OperationID, SessionID: lease.SessionID, Fence: c.Fence, Inputs: c.Inputs, Commit: c.Commit, BaseCommit: lease.BaseCommit, ObjectRef: c.ObjectRef, ChecksRef: c.ChecksRef, ChangedPaths: c.ChangedPaths, State: "submitted", CreatedAt: time.Now().UTC()}
	if c.Commit == "" || c.ObjectRef == "" || c.ChecksRef == "" || len(c.ChangedPaths) == 0 {
		return out, domain.Failure(apierrors.EvidenceMissing, "commit, published objects, checks and changed paths are required")
	}
	if err = c.Usage.Validate(); err != nil {
		return out, err
	}
	if !domain.Within(c.ChangedPaths, lease.WritePaths) {
		return out, domain.Failure(apierrors.ScopeDenied, "artifact exceeds leased write paths")
	}
	fact, err := s.ownership.Reconcile(ctx, lease.OperationID, lease.SessionID)
	if err != nil {
		return out, err
	}
	if fact.OperationID != lease.OperationID || fact.SessionID != lease.SessionID || !fact.StopConfirmed || !settledOutcome(fact.Outcome) {
		return out, domain.Failure(apierrors.DeliveryUnknown, "original executor must be reconciled before releasing its reservation")
	}
	if fact.Outcome == "not_started" || !safeFailureCode(fact.FailureCode) {
		return out, domain.Failure(apierrors.EvidenceMissing, "original execution has no safe actual artifact outcome")
	}
	if fact.Usage == nil || !fact.CallsKnown || !fact.TokensKnown {
		return out, domain.Failure(apierrors.EvidenceMissing, "original operation usage has not been reconciled")
	}
	if !resourcesEqual(c.Usage, *fact.Usage) {
		return out, domain.Failure(apierrors.EvidenceMissing, "submitted usage differs from reconciled operation usage")
	}
	verification, err := s.evidence.VerifyArtifact(ctx, lease, out)
	if err != nil {
		return out, err
	}
	if !verification.ApplicableChecksVerified {
		return out, domain.Failure(apierrors.EvidenceMissing, "the actual project tests and build have not been verified for this commit")
	}
	out.ProjectChecksRef = verification.ProjectChecksRef
	out.ExecutionOutcome = fact.Outcome
	out.ExecutionFailureCode = fact.FailureCode
	err = s.authorized(ctx, p, c.MissionID, "submit_artifact", c.Usage, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.Artifact](tx, p, "submit_artifact", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			m, err := tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			if err = m.CanExecute(time.Now().UTC()); err != nil {
				return err
			}
			w, err := tx.WorkItem(c.WorkItemID)
			if err != nil {
				return err
			}
			if w.MissionID != m.ID {
				return domain.Failure(apierrors.ScopeDenied, "work item belongs to another mission")
			}
			if err = checkVersion(w.Version, c.Meta.ExpectedVersion); err != nil {
				return err
			}
			current, err := tx.Lease(w.ID)
			if err != nil {
				return err
			}
			if err = current.Check(p.ID, c.Fence, time.Now().UTC()); err != nil {
				return err
			}
			if current.OperationID != lease.OperationID || current.SessionID != lease.SessionID {
				return domain.Failure(apierrors.LeaseLost, "execution identity changed")
			}
			if w.Fence != c.Fence {
				return domain.Failure(apierrors.LeaseLost, "work item fencing token has changed")
			}
			if !slices.Equal(w.Inputs, c.Inputs) {
				return domain.Failure(apierrors.ContextStale, "artifact input versions differ from work item")
			}
			if !c.Usage.Fits(current.Reserved) {
				return domain.Failure(apierrors.BudgetExhausted, "reported usage exceeds reserved resources")
			}
			remaining, err := m.Reserved.Sub(current.Reserved)
			if err != nil {
				return err
			}
			used, err := m.Used.Add(c.Usage)
			if err != nil {
				return err
			}
			if !used.Within(g.Budget) {
				return domain.Failure(apierrors.BudgetExhausted, "usage exceeds current approval")
			}
			oldM := m.Version
			m.Version++
			m.Reserved = remaining
			m.Used = used
			oldW := w.Version
			w.Version++
			w.ArtifactID = out.ID
			w.State = "submitted"
			if !verification.ChecksPassed {
				out.State = "revision_needed"
				w.State = "revision_needed"
				m.NoProgress++
				w.Reason = "artifact checks failed"
			} else {
				m.NoProgress = 0
				m.LastProgressAt = time.Now().UTC()
			}
			if err = tx.SaveMission(m, oldM); err != nil {
				return err
			}
			if err = tx.SaveWorkItem(w, oldW); err != nil {
				return err
			}
			if err = tx.SaveArtifact(out, 0); err != nil {
				return err
			}
			if err = tx.DeleteLease(w.ID, current.Fence); err != nil {
				return err
			}
			return finish(tx, p, "submit_artifact", c.Meta, d, out, "artifact_submitted", out.ID, out.Version)
		})
	})
	return out, err
}

func (s *Service) AcceptArtifact(ctx context.Context, p domain.Principal, c AcceptCommand) (domain.Artifact, error) {
	var out domain.Artifact
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if err := human(p); err != nil {
		return out, err
	}
	err := s.authorized(ctx, p, c.MissionID, "accept_artifact", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.Artifact](tx, p, "accept_artifact", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			out, err = tx.Artifact(c.ArtifactID)
			if err != nil {
				return err
			}
			if err = checkVersion(out.Version, c.Meta.ExpectedVersion); err != nil {
				return err
			}
			w, err := tx.WorkItem(out.WorkItemID)
			if err != nil {
				return err
			}
			if w.MissionID != c.MissionID {
				return domain.Failure(apierrors.ScopeDenied, "artifact belongs to another mission")
			}
			if out.State != "submitted" || w.State != "submitted" || w.ArtifactID != out.ID {
				return domain.Failure(apierrors.VersionConflict, "artifact is not the current successful submission")
			}
			oldA := out.Version
			out.Version++
			out.State = "accepted"
			out.AcceptanceID = uuid.NewString()
			out.AcceptedBy = p.ID
			out.AcceptedActorKind = p.Kind
			out.ExperienceStatus = "pending"
			oldW := w.Version
			w.Version++
			w.State = "accepted"
			if err = tx.SaveArtifact(out, oldA); err != nil {
				return err
			}
			if err = tx.SaveWorkItem(w, oldW); err != nil {
				return err
			}
			rows, err := tx.WorkItems(c.MissionID)
			if err != nil {
				return err
			}
			complete := len(rows) > 0
			for _, row := range rows {
				if row.State != "accepted" {
					complete = false
				}
			}
			if complete {
				m, err := tx.Mission(c.MissionID)
				if err != nil {
					return err
				}
				old := m.Version
				m.Version++
				m.State = "completed"
				m.Reason = "all work items accepted"
				if err = tx.SaveMission(m, old); err != nil {
					return err
				}
			}
			return finish(tx, p, "accept_artifact", c.Meta, d, out, "artifact_accepted", out.ID, out.Version)
		})
	})
	return out, err
}
