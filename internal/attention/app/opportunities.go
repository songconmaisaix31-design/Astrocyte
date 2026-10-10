package app

import (
	"context"
	"crypto/rand"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *Service) GetOpportunity(ctx context.Context, p Principal, id string) (OpportunityDetail, error) {
	var result OpportunityDetail
	if err := authorize(p, false); err != nil {
		return result, err
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		var err error
		result, err = tx.LoadOpportunity(id)
		if err != nil {
			return err
		}
		if p.Kind == "agent" {
			if err = validateOpportunityAccess(tx, result.Opportunity); err != nil {
				return err
			}
			for _, revision := range result.Revisions {
				if err = validateOpportunityAccess(tx, revision); err != nil {
					return err
				}
			}
		}
		profile, err := rankingProfile(tx)
		if err != nil {
			return err
		}
		result.Opportunity = projectOpportunityRanking(result.Opportunity, profile)
		return nil
	})
	return result, mapError(err, "")
}

func validateOpportunityAccess(tx AttentionTx, o Opportunity) error {
	if err := validateSourceRefs(tx, o.EvidenceRefs); err != nil {
		return err
	}
	all, err := tx.ListDistillations()
	if err != nil {
		return err
	}
	for _, id := range o.DistillationIDs {
		found := false
		for _, d := range all {
			if d.ID != id {
				continue
			}
			found = true
			if err := validateSourceRefs(tx, d.InputRefs); err != nil {
				return err
			}
			if err := validateSourceRefs(tx, d.RelatedRefs); err != nil {
				return err
			}
		}
		if !found {
			return apierrors.NewNotFound("distillation", id)
		}
	}
	return nil
}

func (s *Service) buildOpportunity(tx AttentionTx, c OpportunityCommand, id string, revision, version int) (Opportunity, error) {
	if strings.TrimSpace(c.Title) == "" {
		return Opportunity{}, domain.ErrInvalid
	}
	if err := domain.ValidateDimensions(dimensions(c.Dimensions)); err != nil {
		return Opportunity{}, err
	}
	if err := validateSourceRefs(tx, c.EvidenceRefs); err != nil {
		return Opportunity{}, err
	}
	state, err := domain.OpportunityState(domainRefs(c.EvidenceRefs), c.Purpose, c.NextStep)
	if err != nil {
		return Opportunity{}, err
	}
	all, err := tx.ListDistillations()
	if err != nil {
		return Opportunity{}, err
	}
	associatedOrQuestion := false
	for _, distillationID := range c.DistillationIDs {
		found := false
		for _, d := range all {
			if d.ID == distillationID {
				found = true
				if d.Status != "succeeded" {
					return Opportunity{}, domain.ErrEvidence
				}
				if err := validateSourceRefs(tx, d.InputRefs); err != nil {
					return Opportunity{}, err
				}
				// A linked successful record must concern this candidate's evidence.
				// A source/use/next-step alone does not complete theme association.
				if overlapsEvidence(d.InputRefs, c.EvidenceRefs) && hasAssociationOrQuestion(d) {
					associatedOrQuestion = true
				}
			}
		}
		if !found {
			return Opportunity{}, apierrors.NewNotFound("distillation", distillationID)
		}
	}
	if state == "ready_for_review" && !associatedOrQuestion {
		state = "incubating"
	}
	return Opportunity{ID: id, Version: version, Revision: revision, State: state, Title: c.Title, EvidenceRefs: nonNil(normalizeRefs(c.EvidenceRefs)), GoalRefs: nonNil(c.GoalRefs), Dimensions: c.Dimensions, NextStep: c.NextStep, MissingEvidence: nonNil(c.MissingEvidence), Purpose: c.Purpose, DistillationIDs: nonNil(c.DistillationIDs), CreatedAt: s.options.Clock()}, nil
}

func overlapsEvidence(inputs, evidence []SourceRef) bool {
	for _, input := range inputs {
		for _, ref := range evidence {
			if input.MaterialID == ref.MaterialID && input.Revision == ref.Revision {
				return true
			}
		}
	}
	return false
}

func hasAssociationOrQuestion(d Distillation) bool {
	for _, question := range d.PendingQuestions {
		if strings.TrimSpace(question) != "" {
			return true
		}
	}
	if d.NextQuestion != nil && strings.TrimSpace(*d.NextQuestion) != "" {
		return true
	}
	if d.Stage != "topic" && d.Stage != "project" {
		return false
	}
	if len(d.RelatedRefs) > 0 {
		return true
	}
	for _, values := range [][]string{d.RelatedIdeas, d.Conflicts, d.GoalRefs, d.ExistingAssets} {
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
	}
	return false
}

func (s *Service) CreateOpportunity(ctx context.Context, p Principal, c OpportunityCommand) (OpportunityDetail, error) {
	return command(s, ctx, p, c.CommandMeta, "CreateOpportunity", c, func(tx AttentionTx) (OpportunityDetail, error) {
		if c.ExpectedVersion != 1 {
			return OpportunityDetail{}, domain.ErrVersion
		}
		opportunity, err := s.buildOpportunity(tx, c, rand.Text(), 1, 1)
		if err != nil {
			return OpportunityDetail{}, err
		}
		row := OpportunityDetail{SchemaVersion: 1, Opportunity: opportunity, Revisions: []Opportunity{opportunity}, Reviews: []Review{}}
		if err = tx.SaveOpportunity(row, 0); err != nil {
			return OpportunityDetail{}, err
		}
		if err = s.event(tx, "opportunity_created", opportunity.ID, 1, c.CommandMeta, map[string]any{"opportunity_id": opportunity.ID, "revision": 1, "state": opportunity.State}); err != nil {
			return OpportunityDetail{}, err
		}
		return row, nil
	})
}

func (s *Service) ReviseOpportunity(ctx context.Context, p Principal, id string, c OpportunityCommand) (OpportunityDetail, error) {
	input := struct {
		ID string
		OpportunityCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "ReviseOpportunity", input, func(tx AttentionTx) (OpportunityDetail, error) {
		row, err := tx.LoadOpportunity(id)
		if err != nil {
			return OpportunityDetail{}, err
		}
		if row.Opportunity.Version != c.ExpectedVersion {
			return OpportunityDetail{}, domain.ErrVersion
		}
		oldVersion := row.Opportunity.Version
		opportunity, err := s.buildOpportunity(tx, c, id, row.Opportunity.Revision+1, oldVersion+1)
		if err != nil {
			return OpportunityDetail{}, err
		}
		row.Opportunity = opportunity
		row.Revisions = append(row.Revisions, opportunity)
		if err = tx.SaveOpportunity(row, oldVersion); err != nil {
			return OpportunityDetail{}, err
		}
		if err = s.event(tx, "opportunity_revised", id, opportunity.Version, c.CommandMeta, map[string]any{"opportunity_id": id, "revision": opportunity.Revision, "state": opportunity.State}); err != nil {
			return OpportunityDetail{}, err
		}
		return row, nil
	})
}

func (s *Service) ReviewOpportunity(ctx context.Context, p Principal, id string, c ReviewOpportunityCommand) (VersionResult, error) {
	input := struct {
		ID string
		ReviewOpportunityCommand
	}{id, c}
	return command(s, ctx, p, c.CommandMeta, "ReviewOpportunity", input, func(tx AttentionTx) (VersionResult, error) {
		row, err := tx.LoadOpportunity(id)
		if err != nil {
			return VersionResult{}, err
		}
		if row.Opportunity.Version != c.ExpectedVersion {
			return VersionResult{}, domain.ErrVersion
		}
		reasons := map[string]string{}
		if c.Dimensions != nil {
			if err := domain.ValidateDimensions(dimensions(*c.Dimensions)); err != nil {
				return VersionResult{}, err
			}
			for name, score := range dimensions(*c.Dimensions) {
				if strings.TrimSpace(score.Reason) != "" {
					reasons[name] = score.Reason
				}
			}
		}
		state, _, err := domain.FeedbackTransition(row.Opportunity.State, c.Feedback, reasons)
		if err != nil {
			return VersionResult{}, err
		}
		oldVersion := row.Opportunity.Version
		row.Reviews = append(row.Reviews, Review{ID: rand.Text(), OpportunityID: id, Revision: row.Opportunity.Revision, Feedback: c.Feedback, Reason: c.Reason, Dimensions: c.Dimensions, ActorID: p.ID, CreatedAt: s.options.Clock()})
		row.Opportunity.State = state
		row.Opportunity.Version++
		if err = tx.SaveOpportunity(row, oldVersion); err != nil {
			return VersionResult{}, err
		}
		if err = s.event(tx, "opportunity_reviewed", id, row.Opportunity.Version, c.CommandMeta, map[string]any{"opportunity_id": id, "revision": row.Opportunity.Revision, "feedback": c.Feedback, "state": state}); err != nil {
			return VersionResult{}, err
		}
		return VersionResult{SchemaVersion: 1, ID: id, Version: row.Opportunity.Version}, nil
	})
}
