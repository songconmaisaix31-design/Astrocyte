package domain

import "testing"

func TestShouldArchive(t *testing.T) {
	if ShouldArchive(AttentionFloor, 89.9) {
		t.Fatal("89.9 days inactive must not archive")
	}
	if !ShouldArchive(AttentionFloor, 90.0) {
		t.Fatal("90 days inactive at floor must archive")
	}
	if ShouldArchive(AttentionFloor+0.01, 100) {
		t.Fatal("above floor must not archive")
	}
}

func TestAwakenAttentionStartsAtFloor(t *testing.T) {
	if got := AwakenAttention(); got != AttentionFloor {
		t.Fatalf("AwakenAttention = %v, want %v", got, AttentionFloor)
	}
}
