package app

import (
	"context"
	"math"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type KnowledgeAgeBasis string

const (
	KnowledgeAgeCreated      KnowledgeAgeBasis = "created_at"
	KnowledgeAgeConsolidated KnowledgeAgeBasis = "last_consolidation"
)

type KnowledgeQuery struct {
	Topic          string
	Text           string
	Now            int64    // milliseconds
	LambdaR        *float64 // seconds^-1; nil leaves retrieval unconfigured, zero is explicit
	AgeBasis       KnowledgeAgeBasis
	ReferenceBoost float64 // explicit per audited human reference; zero keeps base 1
}

type KnowledgeMatch struct {
	Knowledge     StoredKnowledge           `json:"knowledge"`
	Inputs        domain.RetrievalCandidate `json:"inputs"`
	Score         float64                   `json:"score"`
	AgeAt         int64                     `json:"age_at"`
	RelevanceMode string                    `json:"relevance_mode"`
}

func (s *KnowledgeService) RetrieveKnowledge(ctx context.Context, p Principal, project string, q KnowledgeQuery) ([]KnowledgeMatch, error) {
	if err := s.projectAccess(ctx, p, project); err != nil {
		return nil, err
	}
	if q.LambdaR == nil || (q.AgeBasis != KnowledgeAgeCreated && q.AgeBasis != KnowledgeAgeConsolidated) {
		return nil, serviceError(apierrors.ApprovalRequired, "Retrieval decay and age basis require explicit configuration", "configure_knowledge_retrieval")
	}
	if q.Now <= 0 || math.IsNaN(*q.LambdaR) || math.IsInf(*q.LambdaR, 0) || *q.LambdaR < 0 || math.IsNaN(q.ReferenceBoost) || math.IsInf(q.ReferenceBoost, 0) || q.ReferenceBoost < 0 {
		return nil, domain.ErrInvalid
	}
	matches := map[string]KnowledgeMatch{}
	candidates := []domain.RetrievalCandidate{}
	err := s.repo.WithKnowledgeReadTx(ctx, func(tx KnowledgeTx) error {
		entries, err := tx.ListKnowledge(project, true)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			node, err := tx.KnowledgeAttention(entry.Entry.Topic)
			if err != nil {
				return err
			}
			relevance := domain.KeywordKnowledgeRelevance(entry.Entry, node, q.Topic, q.Text)
			if relevance == 0 {
				continue
			}
			at := entry.CreatedAt
			if q.AgeBasis == KnowledgeAgeConsolidated {
				// Human governance changes updated_at too; only content writes reset age.
				at, err = tx.KnowledgeConsolidatedAt(entry.Entry.ID)
				if err != nil {
					return err
				}
			}
			if at <= 0 || q.Now < at {
				return domain.ErrInvalid
			}
			boost := 1.0
			for _, action := range entry.Actions {
				if action.Action == "reference" && action.ActorKind == "human" {
					boost += q.ReferenceBoost
				}
			}
			candidate := domain.RetrievalCandidate{ID: entry.Entry.ID, Relevance: relevance, Attention: node.Attention, AccessBoost: boost, AgeSeconds: float64(q.Now-at) / 1000}
			candidates = append(candidates, candidate)
			matches[candidate.ID] = KnowledgeMatch{Knowledge: entry, Inputs: candidate, AgeAt: at, RelevanceMode: s.repo.KnowledgeRelevanceMode()}
		}
		return nil
	})
	if err != nil {
		return nil, mapError(err, "")
	}
	result := make([]KnowledgeMatch, 0, len(candidates))
	for _, rank := range domain.RankRetrieval(candidates, *q.LambdaR) {
		match := matches[rank.ID]
		match.Score = rank.Score
		result = append(result, match)
	}
	return result, nil
}
