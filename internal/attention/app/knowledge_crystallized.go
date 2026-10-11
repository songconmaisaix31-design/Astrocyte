package app

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeVersionPolicy string

const (
	KnowledgeReconfirmVersion KnowledgeVersionPolicy = "require_reconfirmation"
	KnowledgeKeepConfirmation KnowledgeVersionPolicy = "retain_confirmation"
)

type KnowledgeCrystallizedUpdate struct {
	ID              string
	OwnerProject    string
	Content         json.RawMessage
	TS              int64
	ExpectedVersion int
	Policy          KnowledgeVersionPolicy
}

func AppendKnowledgeCrystallizedTx(tx KnowledgeTx, u KnowledgeCrystallizedUpdate) (StoredKnowledge, error) {
	var v StoredKnowledge
	if u.Policy != KnowledgeReconfirmVersion && u.Policy != KnowledgeKeepConfirmation {
		return v, serviceError(apierrors.ApprovalRequired, "Crystallized version confirmation policy is undecided", "choose_knowledge_version_policy")
	}
	if u.ExpectedVersion < 1 || !json.Valid(u.Content) || u.TS <= 0 {
		return v, domain.ErrInvalid
	}
	v, err := tx.LoadKnowledge(u.ID)
	if err != nil {
		return v, err
	}
	if v.OwnerProject == "" || v.OwnerProject != u.OwnerProject {
		return v, serviceError(apierrors.ScopeDenied, "Crystallized versions require the owning project", "select_owning_project")
	}
	if v.Version != u.ExpectedVersion || u.TS < v.UpdatedAt {
		return v, domain.ErrVersion
	}
	if err = validateKnowledgeSources(tx, v.OwnerProject, v.Entry.Sources); err != nil {
		return v, err
	}
	history := make([]domain.CrystallizedVersion, len(v.CrystallizedVersions))
	for i, version := range v.CrystallizedVersions {
		history[i] = domain.CrystallizedVersion{Content: version.Content, RecordedAt: time.UnixMilli(version.RecordedAt)}
	}
	if _, err = domain.AppendCrystallized(history, u.Content, time.UnixMilli(u.TS)); err != nil {
		return v, err
	}
	v.CrystallizedVersions = append(v.CrystallizedVersions, KnowledgeVersion{Version: len(history) + 1, Content: slices.Clone(u.Content), RecordedAt: u.TS})
	v.Entry.Crystallized = slices.Clone(u.Content)
	if u.Policy == KnowledgeReconfirmVersion {
		v.Entry.Status = domain.KnowledgeInbox
	}
	v.Version, v.UpdatedAt = u.ExpectedVersion+1, u.TS
	v.ConsolidatedAt = &u.TS
	return v, tx.SaveKnowledge(v, u.ExpectedVersion)
}

func (s *KnowledgeService) AppendCrystallized(ctx context.Context, update KnowledgeCrystallizedUpdate) (StoredKnowledge, error) {
	var result StoredKnowledge
	err := s.repo.WithKnowledgeTx(ctx, func(tx KnowledgeTx) error {
		var err error
		result, err = AppendKnowledgeCrystallizedTx(tx, update)
		return err
	})
	return result, mapError(err, "")
}
