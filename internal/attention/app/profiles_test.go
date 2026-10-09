package app

import (
	"context"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (t *memoryTx) LoadRankingProfile() (RankingProfileDetail, error) {
	r := t.state.Profile
	r.SchemaVersion = 1
	r.Versions = nonNil(r.Versions)
	return r, nil
}
func (t *memoryTx) SaveRankingProfile(r RankingProfileDetail, expected int) error {
	version := 0
	if t.state.Profile.Profile != nil {
		version = t.state.Profile.Profile.Version
	}
	if version != expected {
		return domain.ErrVersion
	}
	t.state.Profile = r
	return nil
}
func weights() map[string]float64 {
	return map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}
}
func TestProfileCASHistoryReceiptAndRollback(t *testing.T) {
	ctx := context.Background()
	r := newMemoryRepo()
	s := NewAttentionService(r, nil, nil, ServiceOptions{})
	empty, err := s.GetRankingProfile(ctx, human)
	if err != nil || empty.Configured || empty.Profile != nil || len(empty.Versions) != 0 {
		t.Fatalf("default: %+v %v", empty, err)
	}
	c := UpdateRankingProfileCommand{CommandMeta: meta("profile", 1), Enabled: true, Weights: weights()}
	_, err = s.UpdateRankingProfile(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	c.Weights["goal_progress"] = 8
	c.CommandMeta = meta("profile2", 1)
	second, err := s.UpdateRankingProfile(ctx, human, c)
	if err != nil || second.Profile.Version != 2 || second.Versions[0].Weights["goal_progress"] != 4 {
		t.Fatalf("history: %+v %v", second, err)
	}
	replay, err := s.UpdateRankingProfile(ctx, human, c)
	if err != nil || !reflect.DeepEqual(clone(second), replay) {
		t.Fatalf("receipt: %+v %v", replay, err)
	}
	c.CommandMeta = meta("stale-profile", 1)
	_, err = s.UpdateRankingProfile(ctx, human, c)
	errorCode(t, err, apierrors.VersionConflict)
	c.CommandMeta = meta("bad-profile", 2)
	c.Weights = map[string]float64{"goal_progress": 1}
	_, err = s.UpdateRankingProfile(ctx, human, c)
	errorCode(t, err, apierrors.ValidationFailed)
	r.failEvent = true
	c.CommandMeta = meta("rollback-profile", 2)
	c.Weights = weights()
	_, err = s.UpdateRankingProfile(ctx, human, c)
	if err == nil {
		t.Fatal("outbox failure ignored")
	}
	r.failEvent = false
	current, _ := s.GetRankingProfile(ctx, human)
	if current.Profile.Version != 2 || len(current.Versions) != 2 {
		t.Fatal("failed tx changed profile")
	}
	_, err = s.GetRankingProfile(ctx, agent)
	errorCode(t, err, apierrors.ScopeDenied)
}
func TestProfileRanksKnownSubsetPreservesUnknownAndHistory(t *testing.T) {
	score := func(v float64) Dimensions {
		d := DimensionScore{Value: &v}
		return Dimensions{GoalProgress: d, CurrentInterest: d, ProjectImprovement: d, Originality: d}
	}
	items := []Opportunity{{ID: "low", Dimensions: score(1)}, {ID: "unknown"}, {ID: "high", Dimensions: score(8)}}
	manual := rankOpportunities(clone(items), nil)
	if manual[0].ID != "low" || manual[0].CompositeScore != nil {
		t.Fatal("default reordered")
	}
	ranked := rankOpportunities(clone(items), &RankingProfile{Version: 3, Enabled: true, Weights: weights()})
	if ranked[0].ID != "high" || ranked[1].ID != "unknown" || ranked[2].ID != "low" || ranked[1].CompositeScore != nil || *ranked[0].RankingProfileVersion != 3 {
		t.Fatalf("ranking: %+v", ranked)
	}
	disabled := rankOpportunities(clone(items), &RankingProfile{Enabled: false, Weights: weights()})
	if disabled[0].ID != "low" || disabled[2].CompositeScore != nil {
		t.Fatal("disabled profile ranks")
	}
}
