package domain

import "math"

// Fixed constants from the task book §5.1–5.5. Configurable values (per-topic
// decay λ, signal strength, split/density/confidence thresholds) are function
// parameters, never frozen defaults.
const (
	// AttentionFloor is the anti-Matthew decay floor: old topics never reach zero.
	AttentionFloor = 0.05
	// AttentionCeiling bounds attention to [0,1] per §4.2.
	AttentionCeiling = 1.0
	// MomentumAlpha is the momentum smoothing coefficient, fixed at 0.3.
	MomentumAlpha = 0.3
)

// DecayAttention applies exponential decay over dt days with per-topic λ
// (units 1/day).
func DecayAttention(attention, lambda, dt float64) float64 {
	return attention * math.Exp(-lambda*dt)
}

// UpdateAttention decays the previous attention, adds the weighted signal sum
// and clamps to [AttentionFloor, AttentionCeiling]. λ and signalSum are
// parameters; failed agent tasks never enter signalSum (§5.2).
func UpdateAttention(prevAttention, lambda, dt, signalSum float64) float64 {
	attention := DecayAttention(prevAttention, lambda, dt) + signalSum
	if attention < AttentionFloor {
		attention = AttentionFloor
	}
	if attention > AttentionCeiling {
		attention = AttentionCeiling
	}
	return attention
}

// UpdateMomentum applies exponential smoothing:
// M(t) = α·M(t-1) + (1-α)·(A(t)-A(t-1)).
func UpdateMomentum(prevMomentum, prevAttention, newAttention float64) float64 {
	return MomentumAlpha*prevMomentum + (1-MomentumAlpha)*(newAttention-prevAttention)
}

// UpdateConfidence shifts confidence by delta and clamps to [0,1]. Success is
// positive, failure negative; failure only lowers confidence (§5.2), it never
// reduces attention.
func UpdateConfidence(confidence, delta float64) float64 {
	confidence += delta
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	return confidence
}
