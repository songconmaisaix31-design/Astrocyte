package domain

import "maps"

// TasteProfile is the ranking portion of the user's explicit taste policy. S1
// does not enable model delegation or training. Every replacement makes a new
// snapshot; application/storage retain previous snapshots without overwriting.
type TasteProfile struct {
	Version int
	Enabled bool
	Weights map[string]float64
}

func (p TasteProfile) WithWeights(enabled bool, weights map[string]float64) (TasteProfile, error) {
	if !CompleteWeights(weights) {
		return TasteProfile{}, ErrInvalid
	}
	return TasteProfile{Version: p.Version + 1, Enabled: enabled, Weights: maps.Clone(weights)}, nil
}
