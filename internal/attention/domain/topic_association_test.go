package domain

import "testing"

func TestSustainedMergeRequiresTimeAndOverlap(t *testing.T) {
	const duration = int64(86400000)
	run, ready := ObserveMerge(AssociationRun{}, 100, .9, true, duration, 2)
	if ready {
		t.Fatal("one high association is not sustained")
	}
	run, ready = ObserveMerge(run, 100, .9, true, duration, 2)
	if ready || run.Observations != 1 {
		t.Fatal("replay counted as another observation")
	}
	run, ready = ObserveMerge(run, 100+duration, .9, false, duration, 2)
	if ready || run.Observations != 0 {
		t.Fatal("missing evidence overlap did not reset the run")
	}
	run, _ = ObserveMerge(run, 100+duration, .9, true, duration, 2)
	run, ready = ObserveMerge(run, 100+2*duration, .8, true, duration, 2)
	if ready || run.Observations != 0 {
		t.Fatal("threshold equality did not reset the run")
	}
	run, _ = ObserveMerge(run, 100+2*duration, .9, true, duration, 2)
	_, ready = ObserveMerge(run, 100+3*duration, .9, true, duration, 2)
	if !ready {
		t.Fatal("two qualifying observations across configured duration did not merge")
	}
}
