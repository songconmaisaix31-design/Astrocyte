package domain

// ColdArchiveInactivityDays is the inactivity window before a bottomed-out
// topic moves to the cold archive (§5.4, fixed at 90).
const ColdArchiveInactivityDays = 90.0

// ShouldArchive reports whether a bottomed-out topic has been inactive for 90
// days. daysInactive = now - lastActive, supplied by the caller in days.
func ShouldArchive(attention, daysInactive float64) bool {
	return IsBottomedOut(attention) && daysInactive >= ColdArchiveInactivityDays
}

// AwakenAttention returns the attention a revived topic starts from: A_floor.
// The cold archive never deletes; a re-surface draw can awaken the node.
func AwakenAttention() float64 {
	return AttentionFloor
}
