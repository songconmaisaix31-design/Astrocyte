package app

import (
	"context"
	"crypto/rand"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type ExperienceRepository interface {
	WithExperienceTx(context.Context, func(ExperienceTx) error) error
}

type ExperienceTx interface {
	FindExperience(ownerProject, eventID string) (StoredExperience, bool, error)
	LoadExperience(id string) (StoredExperience, error)
	ListApprovedExperiences(project, pattern string) ([]StoredExperience, error)
	InsertExperience(StoredExperience) error
	SaveExperience(StoredExperience, int) error
	AppendExperienceShare(string, ShareAudit) error
	AppendExperienceAction(string, ActionAudit) error
}

// ProjectID is bound by trusted application setup, never an event or tool body.
type ExperienceOptions struct {
	ProjectID string
	Clock     func() time.Time
}

type ExperienceService struct {
	repo    ExperienceRepository
	events  EventReader
	options ExperienceOptions
}

var _ ExperienceGovernanceService = (*ExperienceService)(nil)

func NewExperienceService(repo ExperienceRepository, events EventReader, options ExperienceOptions) *ExperienceService {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &ExperienceService{repo: repo, events: events, options: options}
}

func (s *ExperienceService) project(project string) error {
	if strings.TrimSpace(s.options.ProjectID) == "" || project != s.options.ProjectID {
		return serviceError(apierrors.ScopeDenied, "Experience project differs from trusted application binding", "use_bound_project")
	}
	return nil
}

func experiencePrincipal(p Principal, humanOnly bool) error {
	if strings.TrimSpace(p.ID) == "" || (p.Kind != "human" && p.Kind != "agent") {
		return serviceError(apierrors.ScopeDenied, "Trusted caller required", "authenticate")
	}
	if humanOnly && p.Kind != "human" {
		return serviceError(apierrors.ApprovalRequired, "Experience governance requires a human", "request_human_review")
	}
	return nil
}

// CaptureFailure consumes a persisted event ID. A non-task, pending receipt or
// unknown outcome produces no candidate. Import/distillation failures retain
// their source meaning even if a caller mislabeled the mapped event.
func (s *ExperienceService) CaptureFailure(ctx context.Context, eventID string) (*StoredExperience, error) {
	if err := s.project(s.options.ProjectID); err != nil {
		return nil, err
	}
	event, err := s.events.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event.Event == nil || !domain.CaptureEligible(event.Event.Kind, event.Observation.Outcome) {
		return nil, nil
	}
	switch event.Observation.SourceKind {
	case "import", "distillation", "job_failed", "attention.job_failed":
		return nil, nil
	}
	if event.Event.Scope != s.options.ProjectID || event.Observation.Scope != s.options.ProjectID {
		return nil, serviceError(apierrors.ScopeDenied, "Failure event belongs to another project", "use_bound_project")
	}
	if event.ID != eventID || event.OriginKey == "" || event.RawBytes == nil {
		return nil, serviceError(apierrors.EvidenceMissing, "Failure must resolve to an original persisted receipt", "persist_original_event")
	}
	pattern, verdict, condition, err := domain.StructureFailure(event.Event.Kind, event.Observation.Outcome)
	if err != nil {
		return nil, err
	}
	var result StoredExperience
	err = s.repo.WithExperienceTx(ctx, func(tx ExperienceTx) error {
		previous, found, err := tx.FindExperience(s.options.ProjectID, eventID)
		if err != nil {
			return err
		}
		if found {
			result = previous
			return nil
		}
		now := s.options.Clock().UnixMilli()
		result = StoredExperience{
			Experience: domain.Experience{ID: rand.Text(), Pattern: pattern, Verdict: verdict, Condition: condition,
				Evidence: []domain.EventRef{{EventID: eventID, TS: event.Event.TS}}, Status: domain.StatusQuarantine},
			OwnerProject: s.options.ProjectID, OriginEventID: eventID, Version: 1, CreatedAt: now, UpdatedAt: now,
			Shares: []ShareAudit{}, Actions: []ActionAudit{},
		}
		return tx.InsertExperience(result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CaptureFailures catches up original receipts after restart without scheduling
// tasks or recapturing an already approved/demoted candidate.
func (s *ExperienceService) CaptureFailures(ctx context.Context, from, to int64) ([]StoredExperience, error) {
	events, err := s.events.ListEvents(ctx, from, to)
	if err != nil {
		return nil, err
	}
	result := []StoredExperience{}
	for _, event := range events {
		if event.Event == nil || event.Event.Scope != s.options.ProjectID {
			continue
		}
		captured, err := s.CaptureFailure(ctx, event.ID)
		if err != nil {
			return nil, err
		}
		if captured != nil {
			result = append(result, *captured)
		}
	}
	return result, nil
}

func experienceVisible(record StoredExperience, project string) bool {
	if record.OwnerProject == project && project != "" {
		return true
	}
	for _, share := range record.Shares {
		if share.TargetProject == project && share.OwnerProject == record.OwnerProject && share.ActorKind == "human" {
			return true
		}
	}
	return false
}

func (s *ExperienceService) GetExperience(ctx context.Context, p Principal, id, project string) (StoredExperience, error) {
	var result StoredExperience
	if err := experiencePrincipal(p, false); err != nil {
		return result, err
	}
	if err := s.project(project); err != nil {
		return result, err
	}
	err := s.repo.WithExperienceTx(ctx, func(tx ExperienceTx) error {
		record, err := tx.LoadExperience(id)
		if err != nil {
			return err
		}
		if !experienceVisible(record, project) || (record.Experience.Status != domain.StatusApproved && (p.Kind != "human" || record.OwnerProject != project)) {
			return serviceError(apierrors.ScopeDenied, "Experience is not visible in this project", "request_human_share")
		}
		result = record
		return nil
	})
	return result, err
}

// GetExperiences is the approved-only read side used by MCP. Empty pattern
// lists approved records in the trusted project; otherwise it is a substring.
func (s *ExperienceService) GetExperiences(ctx context.Context, p Principal, project, pattern string) ([]StoredExperience, error) {
	if err := experiencePrincipal(p, false); err != nil {
		return nil, err
	}
	if err := s.project(project); err != nil {
		return nil, err
	}
	var result []StoredExperience
	err := s.repo.WithExperienceTx(ctx, func(tx ExperienceTx) error {
		var err error
		result, err = tx.ListApprovedExperiences(project, pattern)
		return err
	})
	return result, err
}

// Evidence remains a view of original persisted Event receipts, including raw
// bytes. Shared evidence is readable only through its explicitly shared record.
func (s *ExperienceService) ResolveExperienceEvidence(ctx context.Context, p Principal, id, project string) ([]StoredEvent, error) {
	record, err := s.GetExperience(ctx, p, id, project)
	if err != nil {
		return nil, err
	}
	result := make([]StoredEvent, 0, len(record.Experience.Evidence))
	for _, ref := range record.Experience.Evidence {
		event, err := s.events.GetEvent(ctx, ref.EventID)
		if err != nil {
			return nil, err
		}
		if event.ID != ref.EventID || event.Event == nil || event.Event.TS != ref.TS || event.RawBytes == nil {
			return nil, serviceError(apierrors.EvidenceMissing, "Original failure evidence does not resolve", "inspect_original_event")
		}
		result = append(result, event)
	}
	return result, nil
}

func (s *ExperienceService) ApproveExperience(ctx context.Context, p Principal, id string, cmd ApproveExperienceCommand) (ExperienceResult, error) {
	return s.govern(ctx, p, id, cmd.OwnerProject, "", domain.ActionApprove, cmd.CommandMeta)
}

func (s *ExperienceService) DemoteExperience(ctx context.Context, p Principal, id string, cmd DemoteMemoryCommand) (ExperienceResult, error) {
	return s.govern(ctx, p, id, cmd.OwnerProject, "", domain.ActionDemote, cmd.CommandMeta)
}

func (s *ExperienceService) ShareExperience(ctx context.Context, p Principal, id string, cmd ShareMemoryCommand) (ExperienceResult, error) {
	return s.govern(ctx, p, id, cmd.OwnerProject, cmd.TargetProject, "share", cmd.CommandMeta)
}

func (s *ExperienceService) govern(ctx context.Context, p Principal, id, project, target, action string, cmd CommandMeta) (ExperienceResult, error) {
	var result ExperienceResult
	if err := experiencePrincipal(p, true); err != nil {
		return result, err
	}
	if err := s.project(project); err != nil {
		return result, err
	}
	if cmd.SchemaVersion != 1 || strings.TrimSpace(cmd.RequestID) == "" || cmd.ExpectedVersion < 1 {
		return result, serviceError(apierrors.ValidationFailed, "Experience command needs schema, request and expected version", "correct_command")
	}
	err := s.repo.WithExperienceTx(ctx, func(tx ExperienceTx) error {
		record, err := tx.LoadExperience(id)
		if err != nil {
			return err
		}
		if record.OwnerProject != project {
			return serviceError(apierrors.ScopeDenied, "Only the owning project can govern an experience", "use_owner_project")
		}
		if record.Version != cmd.ExpectedVersion {
			return serviceError(apierrors.VersionConflict, "Experience version changed", "reload_current_version")
		}
		now := s.options.Clock().UnixMilli()
		if action == "share" {
			if err := domain.ValidateShare(domain.ShareRecord{SharedBy: p.ID, SharedToProject: target, SharedAt: time.UnixMilli(now)}, project); err != nil {
				return serviceError(apierrors.ValidationFailed, "Share requires a different target project", "correct_target_project")
			}
			if record.Experience.Status != domain.StatusApproved {
				return serviceError(apierrors.ApprovalRequired, "Quarantined experience cannot be shared", "approve_experience")
			}
			for _, share := range record.Shares {
				if share.TargetProject == target {
					return serviceError(apierrors.VersionConflict, "Experience is already shared to this project", "reload_current_version")
				}
			}
			share := ShareAudit{OwnerProject: project, TargetProject: target, ActorID: p.ID, ActorKind: p.Kind, TS: now, RequestID: cmd.RequestID}
			if err := tx.AppendExperienceShare(id, share); err != nil {
				return err
			}
		} else {
			status, err := domain.Promote(record.Experience.Status, action)
			if err != nil {
				return serviceError(apierrors.ValidationFailed, "Illegal experience status transition", "review_current_status")
			}
			record.Experience.Status = status
		}
		record.Version++
		record.UpdatedAt = now
		if err := tx.SaveExperience(record, cmd.ExpectedVersion); err != nil {
			return err
		}
		if err := tx.AppendExperienceAction(id, ActionAudit{Action: action, ActorID: p.ID, ActorKind: p.Kind, TS: now, RequestID: cmd.RequestID, Version: record.Version}); err != nil {
			return err
		}
		result = ExperienceResult{SchemaVersion: 1, Experience: record.Experience, OwnerProject: project, Version: record.Version}
		return nil
	})
	return result, err
}
