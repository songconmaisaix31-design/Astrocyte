package domain

import (
	"math"
	"testing"
)

func TestDecayAttention(t *testing.T) {
	// λ·dt = ln(2) halves attention (half-life = 1 day).
	got := DecayAttention(1.0, math.Ln2, 1.0)
	if math.Abs(got-0.5) > 1e-12 {
		t.Fatalf("DecayAttention = %v, want 0.5", got)
	}
}

func TestUpdateAttentionFloorAndCeiling(t *testing.T) {
	// An idle old topic never decays below the floor.
	if got := UpdateAttention(AttentionFloor, 0.1, 30.0, 0.0); got != AttentionFloor {
		t.Fatalf("idle floor = %v, want %v", got, AttentionFloor)
	}
	// A strong signal clamps to the ceiling.
	if got := UpdateAttention(0.9, 0.0, 0.0, 1.5); got != AttentionCeiling {
		t.Fatalf("ceiling = %v, want %v", got, AttentionCeiling)
	}
}

func TestUpdateMomentumAlpha(t *testing.T) {
	// α = 0.3 fixed: (1-α)·(A(t)-A(t-1)) with zero prior momentum.
	got := UpdateMomentum(0.0, 0.0, 1.0)
	want := 0.7
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("UpdateMomentum = %v, want %v", got, want)
	}
}

func TestUpdateConfidenceClamps(t *testing.T) {
	if got := UpdateConfidence(0.95, 0.1); got != 1.0 {
		t.Fatalf("upper clamp = %v, want 1.0", got)
	}
	if got := UpdateConfidence(0.05, -0.1); got != 0.0 {
		t.Fatalf("lower clamp = %v, want 0.0", got)
	}
}
