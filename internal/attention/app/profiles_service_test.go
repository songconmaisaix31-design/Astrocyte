package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestProfileCASHistoryReceiptAndRollback(t *testing.T) {
	ctx := context.Background()
	s, db, _ := fixture(t)
	empty, err := s.GetRankingProfile(ctx, human)
	if err != nil || empty.Configured || empty.Profile != nil || len(empty.Versions) != 0 {
		t.Fatal("empty profile changed")
	}
	weights := func() map[string]float64 {
		return map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}
	}
	c := UpdateRankingProfileCommand{CommandMeta: meta("profile", 1), Enabled: true, Weights: weights()}
	if _, err = s.UpdateRankingProfile(ctx, human, c); err != nil {
		t.Fatal("save first profile")
	}
	c.Weights["goal_progress"] = 8
	c.CommandMeta = meta("profile2", 1)
	second, err := s.UpdateRankingProfile(ctx, human, c)
	if err != nil || second.Profile.Version != 2 || second.Versions[0].Weights["goal_progress"] != 4 {
		t.Fatal("profile history changed")
	}
	replay, err := s.UpdateRankingProfile(ctx, human, c)
	if err != nil || !reflect.DeepEqual(second, replay) {
		t.Fatal("profile receipt changed")
	}
	c.CommandMeta = meta("stale-profile", 1)
	_, err = s.UpdateRankingProfile(ctx, human, c)
	errorCode(t, err, apierrors.VersionConflict)
	c.CommandMeta = meta("bad-profile", 2)
	c.Weights = map[string]float64{"goal_progress": 1}
	_, err = s.UpdateRankingProfile(ctx, human, c)
	errorCode(t, err, apierrors.ValidationFailed)
	restore := refuseOutbox(t, db)
	c.CommandMeta = meta("rollback-profile", 2)
	c.Weights = weights()
	if _, err = s.UpdateRankingProfile(ctx, human, c); err == nil {
		t.Fatal("outbox storage failure accepted")
	}
	restore()
	current, err := s.GetRankingProfile(ctx, human)
	if err != nil || current.Profile.Version != 2 || len(current.Versions) != 2 {
		t.Fatal("failed transaction changed profile")
	}
	_, err = s.GetRankingProfile(ctx, agent)
	errorCode(t, err, apierrors.ScopeDenied)
}
