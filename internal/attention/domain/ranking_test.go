package domain

import (
	"slices"
	"testing"
	"time"
)

func TestActualActivityRankingPinAndStableTies(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	input := []MaterialRank{{ID: "new-unread", CreatedAt: now}, {ID: "old-active", Activity: 2, CreatedAt: now.Add(-time.Hour)}, {ID: "fixed-cold", Pinned: true, CreatedAt: now.Add(-24 * time.Hour)}, {ID: "durable", LongTermValue: 1, CreatedAt: now.Add(-time.Hour)}, {ID: "a", CreatedAt: now}}
	ranked := RankMaterials(input)
	ids := []string{}
	for _, item := range ranked {
		ids = append(ids, item.ID)
	}
	if !slices.Equal(ids, []string{"fixed-cold", "old-active", "durable", "a", "new-unread"}) {
		t.Fatal(ids)
	}
	if input[0].ID != "new-unread" {
		t.Fatal("query changed stored order")
	}
}
