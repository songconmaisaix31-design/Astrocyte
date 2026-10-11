package domain

// AssociationRun records consecutive qualifying observations for a topic pair.
// Duration and observation count are supplied by the caller; §5.4 sets only
// the relatedness threshold, not what duration constitutes sustained affinity.
type AssociationRun struct {
	StartedAt      int64 `json:"started_at"`
	LastObservedAt int64 `json:"last_observed_at"`
	Observations   int   `json:"observations"`
}

// ObserveMerge resets the run whenever relatedness or overlap stops qualifying.
// Replaying the same observation time never increases the observation count.
func ObserveMerge(previous AssociationRun, now int64, relatedness float64, overlap bool, requiredDurationMS int64, requiredObservations int) (AssociationRun, bool) {
	if !ShouldMerge(relatedness, overlap) {
		return AssociationRun{}, false
	}
	if previous.Observations == 0 {
		previous = AssociationRun{StartedAt: now, LastObservedAt: now, Observations: 1}
	} else if now > previous.LastObservedAt {
		previous.LastObservedAt = now
		previous.Observations++
	}
	return previous, requiredDurationMS > 0 && requiredObservations > 1 &&
		previous.LastObservedAt-previous.StartedAt >= requiredDurationMS &&
		previous.Observations >= requiredObservations
}
