package sqlite

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestRankingProfileImmutableConcurrentCASAndRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	repo := NewAttentionRepository(db)
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.(app.RankingProfileTx).LoadRankingProfile()
		if v.Configured || v.Profile != nil || len(v.Versions) != 0 {
			t.Fatal("invented default ranking weights")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	first := app.RankingProfile{ID: "attention", Version: 1, Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}}
	one := app.RankingProfileDetail{SchemaVersion: 1, Configured: true, Profile: &first, Versions: []app.RankingProfile{first}}
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error { return tx.(app.RankingProfileTx).SaveRankingProfile(one, 0) }); err != nil {
		t.Fatal(err)
	}
	db2 := openAttentionDB(t, path)
	other := NewAttentionRepository(db2)
	second := first
	second.Version = 2
	second.Enabled = false
	two := one
	two.Profile = &second
	two.Versions = append(append([]app.RankingProfile{}, one.Versions...), second)
	results := make(chan error, 12)
	var group sync.WaitGroup
	for i := 0; i < 12; i++ {
		group.Add(1)
		r := repo
		if i%2 == 0 {
			r = other
		}
		go func() {
			defer group.Done()
			results <- r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.(app.RankingProfileTx).SaveRankingProfile(two, 1) })
		}()
	}
	group.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("expected one CAS winner, got %d", winners)
	}
	db.Close()
	db2.Close()
	restarted := openAttentionDB(t, path)
	again := NewAttentionRepository(restarted)
	if err := again.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.(app.RankingProfileTx).LoadRankingProfile()
		if err != nil {
			return err
		}
		if v.Profile.Version != 2 || v.Profile.Enabled || len(v.Versions) != 2 || !equalJSON(v.Versions[0], first) {
			t.Fatal("ranking profile/history lost on restart")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	third := second
	third.Version = 3
	bad := two
	bad.Profile = &third
	bad.Versions = append(append([]app.RankingProfile{}, two.Versions...), third)
	bad.Versions[0].Enabled = false
	if err := again.WithTx(ctx, func(tx app.AttentionTx) error { return tx.(app.RankingProfileTx).SaveRankingProfile(bad, 2) }); err == nil {
		t.Fatal("old profile revision was rewritten")
	}
	if err := again.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.(app.RankingProfileTx).LoadRankingProfile()
		if err == nil && (v.Profile.Version != 2 || !v.Versions[0].Enabled) {
			t.Fatal("failed write changed head/history")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := again.WithTx(ctx, func(tx app.AttentionTx) error { return tx.(app.RankingProfileTx).SaveRankingProfile(one, -1) }); err == nil {
		t.Fatal("negative expected version accepted")
	}
}
