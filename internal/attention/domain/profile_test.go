package domain

import "testing"

func TestTasteProfileVersionsAndRealUserWeights(t *testing.T) {
	weights := map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}
	one, err := (TasteProfile{}).WithWeights(true, weights)
	if err != nil || one.Version != 1 {
		t.Fatal(err)
	}
	weights["goal_progress"] = 5
	two, err := one.WithWeights(true, weights)
	if err != nil || two.Version != 2 || one.Weights["goal_progress"] != 4 || two.Weights["goal_progress"] != 5 {
		t.Fatal("profile history mutated")
	}
	weights["originality"] = 0
	if _, err = two.WithWeights(true, weights); err == nil {
		t.Fatal("non-goal dimension disabled")
	}
	weights["originality"] = 6
	if _, err = two.WithWeights(true, weights); err == nil {
		t.Fatal("goal progress no longer primary")
	}
}
