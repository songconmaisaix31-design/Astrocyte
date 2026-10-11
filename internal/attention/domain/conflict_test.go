package domain

import "testing"

func TestShouldConflictConfidenceDrop(t *testing.T) {
	if !ShouldConflictConfidenceDrop(-0.4, 0.3) {
		t.Fatal("drop 0.4 > 0.3 must conflict")
	}
	if ShouldConflictConfidenceDrop(-0.3, 0.3) {
		t.Fatal("drop 0.3 == 0.3 must not conflict")
	}
	if ShouldConflictConfidenceDrop(0.1, 0.3) {
		t.Fatal("positive delta must not conflict")
	}
}
