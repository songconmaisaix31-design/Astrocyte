package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeServiceOptions struct {
	Clock func() time.Time
	// ProjectAccess checks a trusted Agent's granted scope. Absent grants deny Agents.
	ProjectAccess func(context.Context, Principal, string) error
}

type KnowledgeService struct {
	repo    KnowledgeRepository
	options KnowledgeServiceOptions
}

func NewKnowledgeService(repo KnowledgeRepository, options KnowledgeServiceOptions) *KnowledgeService {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &KnowledgeService{repo: repo, options: options}
}

var _ KnowledgeGovernanceService = (*KnowledgeService)(nil)

func (s *KnowledgeService) projectAccess(ctx context.Context, p Principal, project string) error {
	if p.ID == "" || strings.TrimSpace(project) == "" || (p.Kind != "human" && p.Kind != "agent") {
		return serviceError(apierrors.ScopeDenied, "Knowledge requires an authenticated project scope", "select_authorized_project")
	}
	if p.Kind == "agent" {
		if s.options.ProjectAccess == nil {
			return serviceError(apierrors.ScopeDenied, "Agent project access has not been granted", "grant_project_access")
		}
		return s.options.ProjectAccess(ctx, p, project)
	}
	return nil
}

func knowledgeVisible(v StoredKnowledge, project string) bool {
	if v.OwnerProject == "" {
		return false
	}
	shared := make([]string, 0, len(v.Shares))
	for _, share := range v.Shares {
		if share.ActorKind == "human" && share.OwnerProject == v.OwnerProject {
			shared = append(shared, share.TargetProject)
		}
	}
	return domain.KnowledgeVisible(v.OwnerProject, shared, project)
}

func (s *KnowledgeService) GetKnowledge(ctx context.Context, p Principal, project, id string) (StoredKnowledge, error) {
	var result StoredKnowledge
	if err := s.projectAccess(ctx, p, project); err != nil {
		return result, err
	}
	err := s.repo.WithKnowledgeTx(ctx, func(tx KnowledgeTx) error {
		var err error
		result, err = tx.LoadKnowledge(id)
		if err == nil && (!knowledgeVisible(result, project) || (p.Kind == "agent" && !domain.ReferencableKnowledge(result.Entry.Status))) {
			return serviceError(apierrors.ScopeDenied, "Knowledge is outside the confirmed project scope", "confirm_or_share_knowledge")
		}
		return err
	})
	if err != nil {
		return StoredKnowledge{}, mapError(err, "")
	}
	return result, nil
}

func knowledgeResult(v StoredKnowledge) KnowledgeEntryResult {
	return KnowledgeEntryResult{SchemaVersion: 1, Entry: v.Entry, OwnerProject: v.OwnerProject, Version: v.Version}
}

func (s *KnowledgeService) action(ctx context.Context, p Principal, id, project, action string, meta CommandMeta, input any, mutate func(KnowledgeTx, *StoredKnowledge) error) (KnowledgeEntryResult, error) {
	var result KnowledgeEntryResult
	if err := authorize(p, true); err != nil {
		return result, mapError(err, meta.RequestID)
	}
	if err := s.projectAccess(ctx, p, project); err != nil {
		return result, mapError(err, meta.RequestID)
	}
	if err := validateMeta(meta); err != nil {
		return result, mapError(err, meta.RequestID)
	}
	digest, err := semanticDigest(struct {
		ID    string
		Input any
	}{id, input})
	if err != nil {
		return result, err
	}
	err = s.repo.WithKnowledgeTx(ctx, func(tx KnowledgeTx) error {
		caller, command := p.Kind+":"+p.ID, "Knowledge."+action
		receipt, err := tx.LoadKnowledgeReceipt(caller, command, meta.IdempotencyKey)
		if err == nil {
			if receipt.Digest != digest {
				return domain.ErrVersion
			}
			return json.Unmarshal(receipt.Result, &result)
		}
		if !isMissing(err) {
			return err
		}
		v, err := tx.LoadKnowledge(id)
		if err != nil {
			return err
		}
		if !knowledgeVisible(v, project) || (action != "reference" && v.OwnerProject != project) {
			return serviceError(apierrors.ScopeDenied, "This project does not own or reference this knowledge", "select_owning_project")
		}
		if v.Version != meta.ExpectedVersion {
			return domain.ErrVersion
		}
		if err = mutate(tx, &v); err != nil {
			return err
		}
		v.Version++
		v.UpdatedAt = s.options.Clock().UnixMilli()
		if err = tx.SaveKnowledge(v, meta.ExpectedVersion); err != nil {
			return err
		}
		if err = tx.AddKnowledgeAction(id, ActionAudit{Action: action, ActorID: p.ID, ActorKind: p.Kind, TS: v.UpdatedAt, RequestID: meta.RequestID, Version: v.Version}); err != nil {
			return err
		}
		result = knowledgeResult(v)
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return tx.SaveKnowledgeReceipt(caller, command, meta.IdempotencyKey, Receipt{Digest: digest, Result: data})
	})
	return result, mapError(err, meta.RequestID)
}

func (s *KnowledgeService) ConfirmKnowledge(ctx context.Context, p Principal, id string, c ConfirmKnowledgeCommand) (KnowledgeEntryResult, error) {
	return s.action(ctx, p, id, c.OwnerProject, "confirm", c.CommandMeta, c, func(tx KnowledgeTx, v *StoredKnowledge) error {
		if err := validateKnowledgeSources(tx, v.OwnerProject, v.Entry.Sources); err != nil {
			return err
		}
		status, err := domain.ConfirmKnowledge(v.Entry.Status)
		v.Entry.Status = status
		return err
	})
}

func (s *KnowledgeService) DemoteKnowledge(ctx context.Context, p Principal, id string, c DemoteMemoryCommand) (KnowledgeEntryResult, error) {
	return s.action(ctx, p, id, c.OwnerProject, "demote", c.CommandMeta, c, func(_ KnowledgeTx, v *StoredKnowledge) error {
		v.Entry.Status = domain.KnowledgeInbox
		return nil
	})
}

func (s *KnowledgeService) ShareKnowledge(ctx context.Context, p Principal, id string, c ShareMemoryCommand) (KnowledgeEntryResult, error) {
	return s.action(ctx, p, id, c.OwnerProject, "share", c.CommandMeta, c, func(tx KnowledgeTx, v *StoredKnowledge) error {
		if strings.TrimSpace(c.TargetProject) == "" || c.TargetProject == v.OwnerProject {
			return domain.ErrInvalid
		}
		for _, share := range v.Shares {
			if share.TargetProject == c.TargetProject {
				return nil
			}
		}
		return tx.AddKnowledgeShare(id, ShareAudit{OwnerProject: v.OwnerProject, TargetProject: c.TargetProject, ActorID: p.ID, ActorKind: p.Kind, TS: s.options.Clock().UnixMilli(), RequestID: c.RequestID})
	})
}

func (s *KnowledgeService) ReferenceKnowledge(ctx context.Context, p Principal, id string, c ReferenceKnowledgeCommand) (KnowledgeEntryResult, error) {
	return s.action(ctx, p, id, c.ProjectID, "reference", c.CommandMeta, c, func(_ KnowledgeTx, v *StoredKnowledge) error {
		if !domain.ReferencableKnowledge(v.Entry.Status) {
			return serviceError(apierrors.ApprovalRequired, "Knowledge must be confirmed before a project reference", "confirm_knowledge")
		}
		return nil
	})
}
