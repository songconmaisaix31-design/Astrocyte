// Sleep metrics (M3 acceptance: both metrics measurable, no target threshold).
package domain

// SleepMetrics is the nightly consolidation's measurable output.
type SleepMetrics struct {
	RetentionPrecision  float64 // recall of pre-consolidation key facts
	StorageReduction    float64 // fraction of bytes removed by consolidation
	RetentionMeasurable bool
	StorageMeasurable   bool
	BeforeFacts         int
	AfterFacts          int
	BeforeBytes         int64
	AfterBytes          int64
}

// ComputeMetrics measures retention and storage reduction across one pass.
func ComputeMetrics(beforeFacts, afterFacts []string, beforeBytes, afterBytes int64) SleepMetrics {
	return SleepMetrics{
		RetentionPrecision:  RetentionPrecision(beforeFacts, afterFacts),
		StorageReduction:    StorageReduction(beforeBytes, afterBytes),
		RetentionMeasurable: len(sleepFactSet(beforeFacts)) > 0,
		StorageMeasurable:   beforeBytes > 0 && afterBytes >= 0,
		BeforeFacts:         len(sleepFactSet(beforeFacts)),
		AfterFacts:          len(sleepFactSet(afterFacts)),
		BeforeBytes:         beforeBytes,
		AfterBytes:          afterBytes,
	}
}

// RetentionPrecision is the recall rate of pre-consolidation key facts after
// consolidation: facts surviving in after over facts present in before. An
// empty before yields 0; ComputeMetrics marks that recall as unmeasurable.
func RetentionPrecision(before, after []string) float64 {
	beforeSet, afterSet := sleepFactSet(before), sleepFactSet(after)
	if len(beforeSet) == 0 {
		return 0
	}
	survivors := 0
	for f := range beforeSet {
		if _, ok := afterSet[f]; ok {
			survivors++
		}
	}
	return float64(survivors) / float64(len(beforeSet))
}

func sleepFactSet(facts []string) map[string]struct{} {
	set := make(map[string]struct{}, len(facts))
	for _, fact := range facts {
		set[fact] = struct{}{}
	}
	return set
}

// StorageReduction is (before-after)/before. Growth produces a negative rate.
// Invalid byte counts yield 0; ComputeMetrics marks that rate unmeasurable.
func StorageReduction(beforeBytes, afterBytes int64) float64 {
	if beforeBytes <= 0 || afterBytes < 0 {
		return 0
	}
	return (float64(beforeBytes) - float64(afterBytes)) / float64(beforeBytes)
}
