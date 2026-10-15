package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
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

func TestRankingProfileConcurrentConnectionsAndReopenKeepHistory(t *testing.T) {
	ctx := context.Background()
	s, db, objects := fixture(t)
	initial, err := s.UpdateRankingProfile(ctx, human, UpdateRankingProfileCommand{CommandMeta: meta("ranking-initial", 1), Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}})
	if err != nil || initial.Profile == nil || initial.Profile.Version != 1 {
		t.Fatal("create actual initial ranking profile")
	}
	var file string
	if err = db.Conn().QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&file); err != nil {
		t.Fatal("resolve actual ranking database")
	}
	otherDB, err := sqlite.Open(file)
	if err != nil {
		t.Fatal("open independent ranking connection")
	}
	t.Cleanup(func() {
		if err := otherDB.Close(); err != nil {
			t.Error("close independent ranking connection")
		}
	})
	other := NewAttentionService(sqlite.NewAttentionRepository(otherDB), importers.NewReader(nil), objects, ServiceOptions{})
	commands := []UpdateRankingProfileCommand{
		{CommandMeta: meta("ranking-concurrent-a", 1), Enabled: true, Weights: map[string]float64{"goal_progress": 8, "current_interest": 3, "project_improvement": 2, "originality": 1}},
		{CommandMeta: meta("ranking-concurrent-b", 1), Enabled: false, Weights: map[string]float64{"goal_progress": 6, "current_interest": 1, "project_improvement": 2, "originality": 3}},
	}
	beforeEvents, beforeReceipts := countRows(t, db, "attention_outbox"), countRows(t, db, "attention_receipts")
	type result struct {
		index   int
		profile RankingProfileDetail
		err     error
	}
	start := make(chan struct{})
	finished := make(chan result, 2)
	for i, service := range []*Service{s, other} {
		go func(i int, service *Service) {
			<-start
			profile, err := service.UpdateRankingProfile(ctx, human, commands[i])
			finished <- result{index: i, profile: profile, err: err}
		}(i, service)
	}
	close(start)
	winner, conflicts := -1, 0
	var saved RankingProfileDetail
	for range commands {
		outcome := <-finished
		if outcome.err != nil {
			errorCode(t, outcome.err, apierrors.VersionConflict)
			conflicts++
			continue
		}
		if winner != -1 || outcome.profile.Profile == nil || outcome.profile.Profile.Version != 2 {
			t.Fatal("competing ranking writes did not retain a single CAS winner")
		}
		winner, saved = outcome.index, outcome.profile
	}
	if winner < 0 || conflicts != 1 || countRows(t, db, "attention_ranking_profiles") != 1 || countRows(t, db, "attention_ranking_profile_revisions") != 2 || countRows(t, db, "attention_outbox") != beforeEvents+1 || countRows(t, db, "attention_receipts") != beforeReceipts+1 {
		t.Fatal("ranking contention retained a losing revision, receipt or event")
	}
	if len(saved.Versions) != 2 || !reflect.DeepEqual(initial.Versions[0], saved.Versions[0]) {
		t.Fatal("competing ranking update changed immutable history")
	}
	if err = otherDB.Close(); err != nil {
		t.Fatal("close independent ranking database before reopen")
	}
	if err = db.Close(); err != nil {
		t.Fatal("close original ranking database before reopen")
	}
	reopenedDB, err := sqlite.Open(file)
	if err != nil {
		t.Fatal("reopen persisted ranking database")
	}
	t.Cleanup(func() {
		if err := reopenedDB.Close(); err != nil {
			t.Error("close reopened ranking database")
		}
	})
	reopened := NewAttentionService(sqlite.NewAttentionRepository(reopenedDB), importers.NewReader(nil), objects, ServiceOptions{})
	loaded, err := reopened.GetRankingProfile(ctx, human)
	if err != nil || !sameAdmissionJSON(t, loaded, saved) {
		t.Fatal("ranking head or complete revision history changed after reopen")
	}
	replayed, err := reopened.UpdateRankingProfile(ctx, human, commands[winner])
	if err != nil || !sameAdmissionJSON(t, replayed, saved) {
		t.Fatal("original ranking receipt changed after reopen")
	}
	stale := commands[1-winner]
	stale.CommandMeta = meta("ranking-stale-after-reopen", 1)
	_, err = reopened.UpdateRankingProfile(ctx, human, stale)
	errorCode(t, err, apierrors.VersionConflict)
	denied := commands[winner]
	denied.CommandMeta = meta("ranking-agent-after-reopen", saved.Profile.Version)
	_, err = reopened.UpdateRankingProfile(ctx, agent, denied)
	errorCode(t, err, apierrors.ScopeDenied)
	afterDenied, err := reopened.GetRankingProfile(ctx, human)
	if err != nil || !sameAdmissionJSON(t, afterDenied, saved) {
		t.Fatal("denied ranking writes changed the persisted head or history")
	}
	if countRows(t, reopenedDB, "attention_ranking_profile_revisions") != 2 || countRows(t, reopenedDB, "attention_outbox") != beforeEvents+1 || countRows(t, reopenedDB, "attention_receipts") != beforeReceipts+1 || countRows(t, reopenedDB, "local_agent_sessions") != 0 {
		t.Fatal("replay or denied ranking writes changed durable state or started native work")
	}
}
