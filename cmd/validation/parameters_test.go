package main

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestV22ProductionBoostCapCooldownWithPersistedAuditTimestamps(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "boost.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db))
	const day = int64(86400000)
	const first = int64(1000)
	actions := []app.ActionAudit{}
	for i := 0; i < 12; i++ {
		actions = append(actions, app.ActionAudit{Action: "reference", ActorKind: "human", TS: first + int64(i), RequestID: "synthetic"})
	}
	actions = append(actions, app.ActionAudit{Action: "reference", ActorKind: "agent", TS: first + 12}, app.ActionAudit{Action: "confirm", ActorKind: "human", TS: first + 12}, app.ActionAudit{Action: "reference", ActorKind: "human", TS: 0}, app.ActionAudit{Action: "reference", ActorKind: "human", TS: first + day + 11, RequestID: "after-cooldown"}, app.ActionAudit{Action: "reference", ActorKind: "human", TS: first + 3*day, RequestID: "future"})
	// Save through the real repository then reopen; retrieval runs query_only.
	err = repo.WithKnowledgeTx(ctx, func(tx app.KnowledgeTx) error {
		if err := tx.SaveKnowledge(app.StoredKnowledge{Entry: domain.KnowledgeEntry{ID: "boost", Topic: "fixture", Status: "confirmed", Fluid: "fixture"}, OwnerProject: "p", Version: 1, CreatedAt: 1, UpdatedAt: 1}, 0); err != nil {
			return err
		}
		for i, action := range actions {
			if action.ActorKind != "human" {
				continue
			} // existing SQLite human-only audit contract
			action.RequestID = fmt.Sprintf("synthetic-%d", i)
			action.ActorID = "fixture"
			action.Version = 1
			if err := tx.AddKnowledgeAction("boost", action); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// KnowledgeAttention falls back to zero only when absent; give actual stored
	// attention via the established repository so score also tests lambda units.
	base := sqlite.NewDynamicsRepository(db)
	err = base.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		return tx.SaveNode(app.StoredAttentionNode{Node: domain.AttentionNode{ID: "fixture", Topic: "fixture", Attention: .6, Confidence: .5, RelatedTopics: map[string]float64{}}, UpdatedAt: 1, Version: 1})
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	service := app.NewKnowledgeService(sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db)), app.KnowledgeServiceOptions{})
	for _, c := range []struct {
		now   int64
		boost float64
	}{{first - 1, 1}, {first, 1.1}, {first + 11, 2}, {first + day + 10, 2}, {first + day + 11, 1.1}, {first + 2*day + 10, 1.1}, {first + 2*day + 11, 1}} {
		q := events.ValidationKnowledgeQuery("fixture", "", c.now, app.KnowledgeAgeCreated)
		matches, e := service.RetrieveKnowledge(ctx, app.Principal{ID: "human", Kind: "human"}, "p", q)
		if e != nil || len(matches) != 1 {
			t.Fatalf("query time %d: count %d error %v", c.now, len(matches), e)
		}
		if math.Abs(matches[0].Inputs.AccessBoost-c.boost) > 1e-12 {
			t.Fatalf("time %d boost=%g want=%g", c.now, matches[0].Inputs.AccessBoost, c.boost)
		}
		want := math.Exp(-.05*(float64(c.now-1)/86400000)) * .6 * c.boost
		if math.Abs(matches[0].Score-want) > 1e-12 || *q.LambdaR != .05/86400 {
			t.Fatal("daily retrieval decay unit mismatch")
		}
	}
	if events.ValidationSeedConfidence != .5 || events.ValidationReferenceIncrement != .1 {
		t.Fatal("validation parameter drift")
	}
}

func TestV22AllSeedsBindConfidenceThroughProductionEngine(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "seed.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), events.ValidationDynamicsOptions(app.DynamicsOptions{}))
	for _, topic := range events.SeedTopics() {
		node, e := engine.SeedTopic(ctx, app.Principal{ID: "human", Kind: "human"}, topic, 1000)
		if e != nil || node.Confidence != .5 || node.Decay != .1 || node.Attention != .05 {
			t.Fatal("production seed binding drift")
		}
	}
}
