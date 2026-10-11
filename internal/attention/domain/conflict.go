package domain

// ShouldConflictConfidenceDrop reports whether a single-step confidence drop
// (delta < 0) strictly exceeds the configured threshold and therefore requires
// a Conflict record (§5.5). Never overwrite or silently merge.
func ShouldConflictConfidenceDrop(confidenceDelta, dropThreshold float64) bool {
	return confidenceDelta < -dropThreshold
}
