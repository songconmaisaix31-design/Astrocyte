package domain

import (
	"math"
	"testing"
)

// testLambdaR is a fixed local value; λr is configurable and has no frozen
// production default.
const testLambdaR = 1.0 / (7 * 24 * 3600)

func TestRetrievalScoreMatchesFormula(t *testing.T) {
	lambdaR := testLambdaR
	cases := []struct {
		name string
		c    RetrievalCandidate
		want float64
	}{
		{
			name: "fresh, full relevance and attention, no boost",
			c:    RetrievalCandidate{Relevance: 1, Attention: 1, AccessBoost: 1, AgeSeconds: 0},
			want: 1,
		},
		{
			name: "one week age applies one e-fold decay",
			c:    RetrievalCandidate{Relevance: 1, Attention: 1, AccessBoost: 1, AgeSeconds: 7 * 24 * 3600},
			want: 1 / math.E,
		},
		{
			name: "all factors multiply",
			c:    RetrievalCandidate{Relevance: 0.5, Attention: 0.4, AccessBoost: 2, AgeSeconds: 0},
			want: 0.5 * 0.4 * 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RetrievalScore(tc.c, lambdaR)
			if math.Abs(got-tc.want) > 1e-12 {
				t.Fatalf("score = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRankRetrievalIsRecomputableAndDeterministic(t *testing.T) {
	// A fixed attention state (same Attention per candidate) must always yield
	// the same order. Relevance, boost and age vary to force a non-trivial rank.
	candidates := []RetrievalCandidate{
		{ID: "a", Relevance: 0.9, Attention: 0.5, AccessBoost: 1, AgeSeconds: 100},
		{ID: "b", Relevance: 0.7, Attention: 0.5, AccessBoost: 3, AgeSeconds: 100},
		{ID: "c", Relevance: 0.9, Attention: 0.5, AccessBoost: 1, AgeSeconds: 10000},
		{ID: "d", Relevance: 0.9, Attention: 0.5, AccessBoost: 1, AgeSeconds: 100},
	}
	first := RankRetrieval(candidates, testLambdaR)
	second := RankRetrieval(candidates, testLambdaR)
	if len(first) != len(second) {
		t.Fatal("length mismatch")
	}
	for i := range first {
		if first[i].ID != second[i].ID || first[i].Score != second[i].Score {
			t.Fatalf("rank %d differs: %+v vs %+v", i, first[i], second[i])
		}
	}
	// b wins on access boost, then a over c (age decays c), d ties a by ID.
	want := []string{"b", "a", "d", "c"}
	got := make([]string, len(first))
	for i, r := range first {
		got[i] = r.ID
	}
	if !equalStrings(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestRetrievalDoesNotMutateCandidates(t *testing.T) {
	candidates := []RetrievalCandidate{
		{ID: "a", Relevance: 0.9, Attention: 0.5, AccessBoost: 1, AgeSeconds: 100},
		{ID: "b", Relevance: 0.7, Attention: 0.5, AccessBoost: 3, AgeSeconds: 100},
	}
	before := append([]RetrievalCandidate(nil), candidates...)
	RankRetrieval(candidates, testLambdaR)
	for i := range candidates {
		if candidates[i] != before[i] {
			t.Fatalf("candidate %d mutated: %+v", i, candidates[i])
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
