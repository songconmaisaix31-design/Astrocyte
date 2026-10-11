package domain

import "math/rand/v2"

// MomentumThreshold is the fast-candidate gate θ (§5.3, fixed at 0.15).
const MomentumThreshold = 0.15

// IsRising reports whether momentum exceeds θ. A rising topic qualifies as a
// fast candidate and anchors explanation; it injects no deviation into A(t).
func IsRising(momentum float64) bool {
	return momentum > MomentumThreshold
}

// IsBottomedOut reports whether attention has reached the decay floor.
func IsBottomedOut(attention float64) bool {
	return attention <= AttentionFloor
}

// SelectResurface draws up to k candidates uniformly without replacement using
// the injected deterministic RNG. The caller pre-filters bottomed-out topics;
// this keeps the draw reproducible for golden tests.
func SelectResurface(candidates []string, rng *rand.Rand, k int) []string {
	if k <= 0 {
		return nil
	}
	ids := append([]string(nil), candidates...)
	rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	if k > len(ids) {
		k = len(ids)
	}
	return ids[:k]
}
