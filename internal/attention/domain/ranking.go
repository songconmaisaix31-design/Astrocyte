package domain

import (
	"cmp"
	"slices"
	"time"
)

// MaterialRank keeps pinning and durable value separate from decaying human
// activity. This engineering strategy orders data; it never grants access.
type MaterialRank struct {
	ID            string
	Pinned        bool
	Activity      float64
	LongTermValue float64
	CreatedAt     time.Time
}

const MaterialRankingStrategy = "attention-v1"

func RankMaterials(values []MaterialRank) []MaterialRank {
	result := slices.Clone(values)
	slices.SortFunc(result, func(a, b MaterialRank) int {
		if a.Pinned != b.Pinned {
			if a.Pinned {
				return -1
			}
			return 1
		}
		if n := cmp.Compare(b.Activity, a.Activity); n != 0 {
			return n
		}
		if n := cmp.Compare(b.LongTermValue, a.LongTermValue); n != 0 {
			return n
		}
		if n := b.CreatedAt.Compare(a.CreatedAt); n != 0 {
			return n
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return result
}
