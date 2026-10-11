package domain

import (
	"cmp"
	"math"
	"slices"
)

// RetrievalCandidate is the minimal input the retrieval ranker reads. It is
// deliberately narrower than KnowledgeEntry: retrieval consumes relevance,
// attention, access boost and age, and never touches fluid/crystallized content
// or status. Retrieval is a read-only consumer of attention state.
type RetrievalCandidate struct {
	ID          string
	Relevance   float64 // sqlite-vec similarity in [0,1]
	Attention   float64 // A(topic) at query time
	AccessBoost float64 // human access boost, base 1.0
	AgeSeconds  float64 // entry age at query time (>= 0)
}

// RetrievalRanking is one entry's computed score. The candidate input is never
// mutated.
type RetrievalRanking struct {
	ID    string
	Score float64
}

// RetrievalScore applies §5.6: score = relevance × e^(-λr·Δt) × A(topic) ×
// access_boost. It only reads attention; no code path here writes A(t).
func RetrievalScore(c RetrievalCandidate, lambdaR float64) float64 {
	return c.Relevance * math.Exp(-lambdaR*c.AgeSeconds) * c.Attention * c.AccessBoost
}

// RankRetrieval returns candidates ordered by descending RetrievalScore. Equal
// scores break ties by ascending ID so the order is fully reproducible for a
// fixed attention state and fixed inputs.
func RankRetrieval(candidates []RetrievalCandidate, lambdaR float64) []RetrievalRanking {
	out := make([]RetrievalRanking, len(candidates))
	for i, c := range candidates {
		out[i] = RetrievalRanking{ID: c.ID, Score: RetrievalScore(c, lambdaR)}
	}
	slices.SortStableFunc(out, func(a, b RetrievalRanking) int {
		if n := cmp.Compare(b.Score, a.Score); n != 0 {
			return n
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return out
}
