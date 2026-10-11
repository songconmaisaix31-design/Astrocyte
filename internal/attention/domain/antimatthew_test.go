package domain

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestIsRisingBoundary(t *testing.T) {
	if IsRising(MomentumThreshold) {
		t.Fatal("M == θ must not be rising")
	}
	if !IsRising(MomentumThreshold + 1e-9) {
		t.Fatal("M > θ must be rising")
	}
}

func TestIsBottomedOut(t *testing.T) {
	if !IsBottomedOut(AttentionFloor) {
		t.Fatal("A == floor must be bottomed out")
	}
	if IsBottomedOut(AttentionFloor + 0.01) {
		t.Fatal("A > floor must not be bottomed out")
	}
}

func TestSelectResurfaceDeterministicAndDistinct(t *testing.T) {
	candidates := []string{"a", "b", "c", "d", "e"}
	got1 := SelectResurface(candidates, rand.New(rand.NewPCG(7, 11)), 3)
	got2 := SelectResurface(candidates, rand.New(rand.NewPCG(7, 11)), 3)
	if !slices.Equal(got1, got2) {
		t.Fatalf("same seed diverged: %v vs %v", got1, got2)
	}
	if len(got1) != 3 {
		t.Fatalf("drew %d, want 3", len(got1))
	}
	seen := map[string]bool{}
	for _, id := range got1 {
		if seen[id] {
			t.Fatalf("duplicate draw %q", id)
		}
		seen[id] = true
	}
}

func TestSelectResurfaceBounds(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	if got := SelectResurface([]string{"a", "b"}, rng, 0); got != nil {
		t.Fatalf("k=0 must draw nothing, got %v", got)
	}
	if got := SelectResurface([]string{"a", "b"}, rng, 5); len(got) != 2 {
		t.Fatalf("k>len must cap at len, got %v", got)
	}
}
