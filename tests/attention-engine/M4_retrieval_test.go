package attention_engine_test

import (
	"math"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// M4 acceptance §5.6: retrieval ranking must be independently recomputable at a
// fixed attention state. This test recomputes the score formula by hand and
// asserts the ranker reproduces it. λr is a configurable parameter with no
// frozen production default; the value here is local to the test.
const testLambdaR = 1.0 / (7 * 24 * 3600)

func TestM4RetrievalRankingRecomputableAtFixedAttention(t *testing.T) {
	// Fixed attention: every candidate carries the same A(topic). Relevance,
	// access boost and age vary so the rank is non-trivial.
	fixedAttention := 0.6
	candidates := []domain.RetrievalCandidate{
		{ID: "a", Relevance: 0.9, Attention: fixedAttention, AccessBoost: 1, AgeSeconds: 0},
		{ID: "b", Relevance: 0.5, Attention: fixedAttention, AccessBoost: 4, AgeSeconds: 0},
		{ID: "c", Relevance: 0.9, Attention: fixedAttention, AccessBoost: 1, AgeSeconds: 86400},
		{ID: "d", Relevance: 0.9, Attention: fixedAttention, AccessBoost: 1, AgeSeconds: 0},
	}

	// Independent hand recomputation of score = relevance · e^(-λr·Δt) · A · boost.
	hand := map[string]float64{}
	for _, c := range candidates {
		hand[c.ID] = c.Relevance * math.Exp(-testLambdaR*c.AgeSeconds) * c.Attention * c.AccessBoost
	}

	first := domain.RankRetrieval(candidates, testLambdaR)
	second := domain.RankRetrieval(candidates, testLambdaR)

	if len(first) != len(candidates) {
		t.Fatalf("len = %d, want %d", len(first), len(candidates))
	}
	for i := range first {
		if first[i].ID != second[i].ID || first[i].Score != second[i].Score {
			t.Fatalf("rank %d not reproducible: %+v vs %+v", i, first[i], second[i])
		}
		if math.Abs(first[i].Score-hand[first[i].ID]) > 1e-12 {
			t.Fatalf("score for %s = %v, hand = %v", first[i].ID, first[i].Score, hand[first[i].ID])
		}
		if i > 0 && first[i-1].Score < first[i].Score {
			t.Fatalf("order not descending at %d: %+v then %+v", i, first[i-1], first[i])
		}
	}
	// b wins on access boost; a beats c (c is decayed by age); d ties a by ID.
	want := []string{"b", "a", "d", "c"}
	for i, id := range want {
		if first[i].ID != id {
			t.Fatalf("rank[%d] = %s, want %s (full order %v)", i, first[i].ID, id, idsOf(first))
		}
	}
}

// M4 §5.6 separation: retrieval is a read-only consumer of attention. It takes
// A(topic) as an input and must never write it back.
func TestM4RetrievalNeverUpdatesAttention(t *testing.T) {
	candidates := []domain.RetrievalCandidate{
		{ID: "a", Relevance: 0.9, Attention: 0.6, AccessBoost: 1, AgeSeconds: 100},
	}
	before := append([]domain.RetrievalCandidate(nil), candidates...)
	domain.RankRetrieval(candidates, testLambdaR)
	for i := range candidates {
		if candidates[i] != before[i] {
			t.Fatalf("candidate %d mutated by retrieval: %+v", i, candidates[i])
		}
	}
}

func TestM4RetrievalLambdaRIsAnIndependentQueryParameter(t *testing.T) {
	candidates := []domain.RetrievalCandidate{
		{ID: "old", Relevance: 1, Attention: 0.8, AccessBoost: 2, AgeSeconds: 86400},
		{ID: "fresh", Relevance: 1, Attention: 0.4, AccessBoost: 1, AgeSeconds: 0},
	}
	for _, tc := range []struct {
		lambda float64
		winner string
	}{
		{lambda: 0, winner: "old"},
		{lambda: 2.0 / 86400, winner: "fresh"},
	} {
		ranked := domain.RankRetrieval(candidates, tc.lambda)
		if ranked[0].ID != tc.winner {
			t.Fatalf("lambda=%g: winner=%s, want %s", tc.lambda, ranked[0].ID, tc.winner)
		}
		want := map[string]float64{"old": 1.6 * math.Exp(-tc.lambda*86400), "fresh": 0.4}
		for _, result := range ranked {
			if math.Abs(result.Score-want[result.ID]) > 1e-12 {
				t.Fatalf("lambda=%g: score for %s=%g, want %g", tc.lambda, result.ID, result.Score, want[result.ID])
			}
		}
	}
	if candidates[0].Attention != 0.8 || candidates[1].Attention != 0.4 {
		t.Fatal("query parameter changed attention")
	}
}

func idsOf(rank []domain.RetrievalRanking) []string {
	out := make([]string, len(rank))
	for i, r := range rank {
		out[i] = r.ID
	}
	return out
}
