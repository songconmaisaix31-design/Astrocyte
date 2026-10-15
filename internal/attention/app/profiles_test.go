package app

import (
	"testing"
)

func weights() map[string]float64 {
	return map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}
}
func TestProfileRanksKnownSubsetPreservesUnknownAndHistory(t *testing.T) {
	score := func(v float64) Dimensions {
		d := DimensionScore{Value: &v}
		return Dimensions{GoalProgress: d, CurrentInterest: d, ProjectImprovement: d, Originality: d}
	}
	items := []Opportunity{{ID: "low", Dimensions: score(1)}, {ID: "unknown"}, {ID: "high", Dimensions: score(8)}}
	manual := rankOpportunities(clone(items), nil)
	if manual[0].ID != "low" || manual[0].CompositeScore != nil {
		t.Fatal("default reordered")
	}
	ranked := rankOpportunities(clone(items), &RankingProfile{Version: 3, Enabled: true, Weights: weights()})
	if ranked[0].ID != "high" || ranked[1].ID != "unknown" || ranked[2].ID != "low" || ranked[1].CompositeScore != nil || *ranked[0].RankingProfileVersion != 3 {
		t.Fatalf("ranking: %+v", ranked)
	}
	disabled := rankOpportunities(clone(items), &RankingProfile{Enabled: false, Weights: weights()})
	if disabled[0].ID != "low" || disabled[2].CompositeScore != nil {
		t.Fatal("disabled profile ranks")
	}
}
