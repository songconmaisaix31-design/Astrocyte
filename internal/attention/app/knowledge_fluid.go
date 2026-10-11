package app

import (
	"context"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeFluidUpdate struct {
	ID              string
	OwnerProject    string
	Topic           string
	Fluid           string
	Sources         []domain.EventRef
	ExpectedVersion int
	TS              int64
}

// RewriteKnowledgeFluidTx joins sleep's existing transaction; it never opens one.
// Crystallized history and confirmation survive a fluid-only rewrite.
func RewriteKnowledgeFluidTx(tx KnowledgeTx, u KnowledgeFluidUpdate) (StoredKnowledge, error) {
	var v StoredKnowledge
	if strings.TrimSpace(u.ID) == "" || strings.TrimSpace(u.OwnerProject) == "" || strings.TrimSpace(u.Topic) == "" || u.TS <= 0 || u.ExpectedVersion < 0 {
		return v, domain.ErrInvalid
	}
	if err := validateKnowledgeSources(tx, u.OwnerProject, u.Sources); err != nil {
		return v, err
	}
	if u.ExpectedVersion == 0 {
		v = StoredKnowledge{Entry: domain.KnowledgeEntry{ID: u.ID, Topic: u.Topic, Status: domain.KnowledgeInbox}, OwnerProject: u.OwnerProject, CreatedAt: u.TS,
			CrystallizedVersions: []KnowledgeVersion{}, Shares: []ShareAudit{}, Actions: []ActionAudit{}}
	} else {
		var err error
		v, err = tx.LoadKnowledge(u.ID)
		if err != nil {
			return v, err
		}
		if v.OwnerProject != u.OwnerProject || v.Entry.Topic != u.Topic {
			return v, domain.ErrInvalid
		}
		if v.Version != u.ExpectedVersion || u.TS < v.UpdatedAt {
			return v, domain.ErrVersion
		}
	}
	v.Entry.Fluid = u.Fluid
	v.Entry.Sources = slices.Clone(u.Sources)
	if v.Entry.Sources == nil {
		v.Entry.Sources = []domain.EventRef{}
	}
	v.Version, v.UpdatedAt = u.ExpectedVersion+1, u.TS
	v.ConsolidatedAt = &u.TS
	return v, tx.SaveKnowledge(v, u.ExpectedVersion)
}

func validateKnowledgeSources(tx KnowledgeTx, owner string, sources []domain.EventRef) error {
	if len(sources) == 0 {
		return domain.ErrEvidence
	}
	for _, ref := range sources {
		source, err := tx.KnowledgeSource(ref.EventID)
		if err != nil {
			return err
		}
		if source.Observation.Scope != owner || source.Observation.TS != ref.TS || source.OriginKey == "" || len(source.RawBytes) == 0 {
			return domain.ErrEvidence
		}
	}
	return nil
}

func (s *KnowledgeService) RewriteFluid(ctx context.Context, update KnowledgeFluidUpdate) (StoredKnowledge, error) {
	var result StoredKnowledge
	err := s.repo.WithKnowledgeTx(ctx, func(tx KnowledgeTx) error {
		var err error
		result, err = RewriteKnowledgeFluidTx(tx, update)
		return err
	})
	return result, mapError(err, "")
}
