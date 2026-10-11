package domain

import "testing"

func TestShouldMerge(t *testing.T) {
	if ShouldMerge(MergeRelatednessThreshold, true) {
		t.Fatal("relatedness == 0.8 must not merge")
	}
	if !ShouldMerge(MergeRelatednessThreshold+1e-9, true) {
		t.Fatal("relatedness > 0.8 with overlap must merge")
	}
	if ShouldMerge(0.9, false) {
		t.Fatal("no evidence overlap must not merge")
	}
}

func TestMergeAttentionKeepsMax(t *testing.T) {
	if MergeAttention(0.2, 0.8) != 0.8 {
		t.Fatal("merge must keep max")
	}
	if MergeAttention(0.9, 0.1) != 0.9 {
		t.Fatal("merge must keep max")
	}
}

func TestShouldSplit(t *testing.T) {
	if !ShouldSplit(0.1, 0.3) {
		t.Fatal("correlation below threshold must split")
	}
	if ShouldSplit(0.3, 0.3) {
		t.Fatal("correlation at threshold must not split")
	}
}

func TestShouldBirthByDensity(t *testing.T) {
	if !ShouldBirthByDensity(5, 2.0, 2.0) {
		t.Fatal("2.5 events/day > 2.0 must birth")
	}
	if ShouldBirthByDensity(4, 2.0, 2.0) {
		t.Fatal("2.0 events/day == 2.0 must not birth")
	}
	if ShouldBirthByDensity(10, 0.0, 2.0) {
		t.Fatal("zero window must not birth")
	}
}
