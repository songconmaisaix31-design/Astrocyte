package domain

// MergeRelatednessThreshold is the merge gate (§5.4, fixed at 0.8, strictly
// greater).
const MergeRelatednessThreshold = 0.8

// ShouldMerge reports one merge step: relatedness strictly above 0.8 with
// overlapping evidence. "Sustained" is the caller's state across steps.
func ShouldMerge(relatedness float64, evidenceOverlap bool) bool {
	return relatedness > MergeRelatednessThreshold && evidenceOverlap
}

// MergeAttention keeps the maximum historical attention on merge (§5.4).
func MergeAttention(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ShouldSplit reports whether a topic's internal evidence forms two clusters
// with correlation below the configured threshold (§5.4).
func ShouldSplit(correlation, splitThreshold float64) bool {
	return correlation < splitThreshold
}

// ShouldBirthByDensity reports a density-based birth trigger: events per day in
// one scope exceed the configured threshold (§5.4).
func ShouldBirthByDensity(eventCount, windowDays, densityThreshold float64) bool {
	if windowDays <= 0 {
		return false
	}
	return eventCount/windowDays > densityThreshold
}
