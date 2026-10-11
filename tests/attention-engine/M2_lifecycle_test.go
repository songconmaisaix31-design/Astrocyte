package attention_engine_test

import (
	"bytes"
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM2LifecycleBirthSplitSustainedMergeRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, engine := m2open(t, path)
	defer func() { db.Close() }()
	const day = int64(86400000)
	const ts = int64(1767225600000)
	if _, err := engine.SeedTopic(ctx, app.Principal{ID: "agent", Kind: "agent"}, "denied", ts); err == nil {
		t.Fatal("agent created manual seed")
	}
	x := m2signal(t, db, "dense", "density-1", ts, domain.SignalImplicitRegular, 1)
	y := m2signal(t, db, "dense", "density-2", ts+day/2, domain.SignalImplicitRegular, 1)
	if _, err := engine.BirthByDensity(ctx, "dense", "m2-project", []string{x.EventID}, ts, ts+day); err == nil {
		t.Fatal("density threshold equality created topic")
	}
	if _, err := engine.BirthByDensity(ctx, "dense", "other-project", []string{x.EventID, y.EventID}, ts, ts+day); err == nil {
		t.Fatal("cross-project density accepted")
	}
	if _, err := engine.BirthByDensity(ctx, "dense", "m2-project", []string{x.EventID, y.EventID, x.EventID}, ts, ts+day); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ApplySignals(ctx, "dense", []domain.Signal{x, y}, ts+day); err != nil {
		t.Fatal(err)
	}
	children, err := engine.SplitTopic(ctx, "dense", [2]string{"dense-a", "dense-b"}, [2][]string{{x.EventID}, {y.EventID}}, .1, ts+day)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 2 || len(children[0].Node.Evidence) != 1 || len(children[1].Node.Evidence) != 1 {
		t.Fatal("split did not partition evidence")
	}
	if len(children[0].Node.RelatedTopics) < 2 {
		t.Fatal("split dropped relationships")
	}
	if _, err = engine.ExplainFocus(ctx, "dense-a"); err != nil {
		t.Fatal(err)
	}
	parent, err := engine.StoredNode(ctx, "dense")
	if err != nil || parent.Lifecycle != "cold" || len(parent.Node.Evidence) != 2 {
		t.Fatal("split lost parent history", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, engine = m2open(t, path)
	if _, err = engine.SplitTopic(ctx, "dense", [2]string{"dense-a", "dense-b"}, [2][]string{{x.EventID}, {y.EventID}}, .1, ts+day); err != nil {
		t.Fatal("split replay after restart", err)
	}
	// Merge evidence overlap comes from stored raw references. Sleep attachment
	// changes neither A nor updated_at for an existing topic.
	for _, topic := range []string{"merge-a", "merge-b"} {
		if _, err = engine.SeedTopic(ctx, m2human, topic, ts); err != nil {
			t.Fatal(err)
		}
	}
	peak := m2signal(t, db, "merge-a", "peak", ts, domain.SignalImplicitDepth, 1)
	if _, err = engine.ApplySignals(ctx, "merge-a", []domain.Signal{peak}, ts); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ApplySignals(ctx, "merge-a", nil, ts+day); err != nil {
		t.Fatal(err)
	}
	stored, err := engine.StoredNode(ctx, "merge-a")
	if err != nil {
		t.Fatal(err)
	}
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .9, ts+day); err != nil || merged {
		t.Fatal("no-overlap merge", err)
	}
	repo := sqlite.NewDynamicsRepository(db)
	err = repo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		_, err := engine.ConsolidateTopicInTx(tx, "merge-b", []domain.EvidenceRef{{EventID: peak.EventID}}, nil, ts+day)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := engine.StoredNode(ctx, "merge-a")
	if err != nil || unchanged.UpdatedAt != stored.UpdatedAt || unchanged.Node.Attention != stored.Node.Attention {
		t.Fatal("consolidation changed source A", err)
	}
	if merged, err := engine.ObserveAssociation(ctx, "merge-b", "merge-a", .9, ts+day); err != nil || merged {
		t.Fatal("one observation merged", err)
	}
	before := m2nodeBytes(t, db, "merge-a")
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .9, ts+day); err != nil || merged {
		t.Fatal("replayed observation merged", err)
	}
	if !bytes.Equal(before, m2nodeBytes(t, db, "merge-a")) {
		t.Fatal("replayed association changed history")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, engine = m2open(t, path)
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .8, ts+2*day); err != nil || merged {
		t.Fatal("threshold equality merged", err)
	}
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .9, ts+3*day); err != nil || merged {
		t.Fatal("reset did not restart sustained duration", err)
	}
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .9, ts+4*day); err != nil || !merged {
		t.Fatal("persisted sustained overlap did not merge", err)
	}
	a, err := engine.StoredNode(ctx, "merge-a")
	if err != nil {
		t.Fatal(err)
	}
	b, err := engine.StoredNode(ctx, "merge-b")
	if err != nil {
		t.Fatal(err)
	}
	if a.Node.Attention != 1 || b.Lifecycle != "cold" || len(b.Node.Evidence) == 0 {
		t.Fatal("merge did not retain historical max and source evidence")
	}
	if merged, err := engine.ObserveAssociation(ctx, "merge-a", "merge-b", .9, ts+4*day); err != nil || !merged {
		t.Fatal("completed merge replay", err)
	}
}

func TestM2ColdArchiveAndSeededThreeResurface(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, engine := m2open(t, path)
	defer func() { db.Close() }()
	const day = int64(86400000)
	const ts = int64(1767225600000)
	for _, topic := range []string{"cold-a", "cold-b", "cold-c", "cold-d"} {
		if _, err := engine.SeedTopic(ctx, m2human, topic, ts); err != nil {
			t.Fatal(err)
		}
		s := m2signal(t, db, topic, "event-"+topic, ts, domain.SignalImplicitRegular, 1)
		if _, err := engine.ApplySignals(ctx, topic, []domain.Signal{s}, ts); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.AdvanceLifecycle(ctx, ts+89*day); err != nil {
		t.Fatal(err)
	}
	before, err := engine.StoredNode(ctx, "cold-a")
	if err != nil || before.Lifecycle == "cold" {
		t.Fatal("archived before 90 days", err)
	}
	if err = engine.AdvanceLifecycle(ctx, ts+90*day); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, engine = m2open(t, path)
	current, err := engine.StoredNode(ctx, "cold-a")
	if err != nil || current.Lifecycle != "cold" || current.Node.Attention != .05 || current.Node.LastActive != ts {
		t.Fatal("cold archive not persisted", err)
	}
	bytesBefore := m2nodeBytes(t, db, "cold-a")
	if err = engine.AdvanceLifecycle(ctx, ts+90*day); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytesBefore, m2nodeBytes(t, db, "cold-a")) {
		t.Fatal("restarted idle decay applied twice")
	}
	state, err := engine.AttentionState(ctx)
	if err != nil || len(state) != 0 {
		t.Fatal("archive returned as active focus", err)
	}
	explanation, err := engine.ExplainFocus(ctx, "cold-a")
	if err != nil || len(explanation.Evidence) != 1 {
		t.Fatal("archive lost explanation", err)
	}
	chosen, err := engine.Resurface(ctx, ts+90*day, rand.New(rand.NewPCG(41, 9)))
	if err != nil || len(chosen) != 3 {
		t.Fatal("daily draw did not select 3", err)
	}
	for _, n := range chosen {
		if n.Lifecycle != "active" || n.Node.Attention != .05 || n.Node.LastActive != ts || n.UpdatedAt != ts+90*day {
			t.Fatal("awakening did not start at floor or invented activity")
		}
	}
	state, err = engine.AttentionState(ctx)
	if err != nil || len(state) != 3 {
		t.Fatal("awakened topics not in focus", err)
	}
	var count int
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_nodes").Scan(&count); err != nil || count != 4 {
		t.Fatal("archive deleted topics", err)
	}
}

func TestM2SleepAttachAtomicAndConflictHumanGate(t *testing.T) {
	ctx := context.Background()
	db, engine := m2open(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	const ts = int64(1767225600000)
	a := m2signal(t, db, "old", "sleep-a", ts, domain.SignalImplicitRegular, 1)
	b := m2signal(t, db, "old", "sleep-b", ts+1, domain.SignalFeedbackFailure, 1)
	repo := sqlite.NewDynamicsRepository(db)
	if err := repo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		_, err := engine.ConsolidateTopicInTx(tx, "isolated", []domain.EvidenceRef{{EventID: a.EventID}}, nil, ts+1)
		return err
	}); err == nil {
		t.Fatal("isolated raw event created concept")
	}
	err := repo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		_, err := engine.ConsolidateTopicInTx(tx, "concept", []domain.EvidenceRef{{EventID: a.EventID}}, nil, ts+1, domain.EventRef{EventID: b.EventID})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := engine.StoredNode(ctx, "concept")
	if err != nil || len(n.Node.Evidence) != 1 || n.Node.Attention != .05 {
		t.Fatal("adjacent support copied neighbor evidence or wrong seed", err)
	}
	before := m2nodeBytes(t, db, "concept")
	err = repo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		_, err := engine.ConsolidateTopicInTx(tx, "concept", []domain.EvidenceRef{{EventID: b.EventID}}, nil, ts+2)
		if err != nil {
			return err
		}
		return errors.New("simulated daily rollback")
	})
	if err == nil {
		t.Fatal("expected rollback")
	}
	if !bytes.Equal(before, m2nodeBytes(t, db, "concept")) {
		t.Fatal("daily rollback left graph changes")
	}
	if _, err = engine.SeedTopic(ctx, m2human, "old", ts); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ApplySignals(ctx, "old", []domain.Signal{a}, ts); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ApplySignals(ctx, "old", []domain.Signal{b}, ts+1); err != nil {
		t.Fatal(err)
	}
	list, err := engine.Conflicts(ctx)
	if err != nil || len(list) != 1 {
		t.Fatal("confidence drop did not create pending conflict", err)
	}
	if list[0].Conflict.LeftRef != a.EventID || list[0].Conflict.RightRef != b.EventID || list[0].Conflict.Status != domain.ConflictStatusPending {
		t.Fatal("conflict did not retain both refs")
	}
	cmd := app.AdjudicateConflictCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "human-decision", IdempotencyKey: "human-decision", ExpectedVersion: 1}, Decision: app.ConflictDecisionKeepLeft}
	_, err = engine.AdjudicateConflict(ctx, app.Principal{ID: "agent", Kind: "agent"}, list[0].Conflict.ID, cmd)
	var denied *apierrors.ServiceError
	if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
		t.Fatal("agent passed human gate", err)
	}
	result, err := engine.AdjudicateConflict(ctx, m2human, list[0].Conflict.ID, cmd)
	if err != nil || result.Conflict.Status != domain.ConflictStatusKeepLeft {
		t.Fatal("human adjudication failed", err)
	}
	if _, err = engine.AdjudicateConflict(ctx, m2human, list[0].Conflict.ID, cmd); err != nil {
		t.Fatal("human request replay", err)
	}
	cmd.RequestID = "retry-new-request"
	if _, err = engine.AdjudicateConflict(ctx, m2human, list[0].Conflict.ID, cmd); err != nil {
		t.Fatal("idempotency retry with new request id", err)
	}
	cmd.Decision = app.ConflictDecisionKeepRight
	if _, err = engine.AdjudicateConflict(ctx, m2human, list[0].Conflict.ID, cmd); err == nil {
		t.Fatal("changed decision reused idempotency key")
	}
	list, err = engine.Conflicts(ctx)
	if err != nil || len(list[0].Actions) != 1 {
		t.Fatal("human action audit was duplicated", err)
	}
	if _, err = engine.ReportFactConflict(ctx, "old", a.EventID, b.EventID, "contradicted crystallized fact", ts+2); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ReportFactConflict(ctx, "old", a.EventID, b.EventID, "contradicted crystallized fact", ts+2); err != nil {
		t.Fatal(err)
	}
	list, err = engine.Conflicts(ctx)
	if err != nil || len(list) != 2 {
		t.Fatal("fact conflict replay duplicated or overwrote earlier conflict", err)
	}
}

func TestM2SleepBirthAppliesPersistedDaySignalsOnce(t *testing.T) {
	ctx := context.Background()
	db, engine := m2open(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	const day = int64(86400000)
	const ts = int64(1767225600000)
	a := m2signal(t, db, "new-sleep-topic", "new-topic-first", ts, domain.SignalImplicitRegular, 1)
	b := m2signal(t, db, "new-sleep-topic", "new-topic-last", ts+day/2, domain.SignalImplicitRegular, 1)
	repo := sqlite.NewDynamicsRepository(db)
	err := repo.WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		born, err := engine.ConsolidateTopicInTx(tx, "new-sleep-topic", []domain.EvidenceRef{{EventID: a.EventID}, {EventID: b.EventID}}, nil, ts+day)
		if err != nil {
			return err
		}
		if born.Node.Attention != domain.AttentionFloor || born.UpdatedAt != ts || born.Node.LastActive != b.TS {
			t.Fatal("sleep seed time or floor wrong")
		}
		_, err = engine.ApplySignalsInTx(tx, "new-sleep-topic", []domain.Signal{a, b}, ts+day)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := engine.StoredNode(ctx, "new-sleep-topic")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(n.Node.Attention-math.Exp(-.1*.5)) > 1e-12 || n.UpdatedAt != ts+day || n.Node.LastActive != b.TS {
		t.Fatal("sleep birth signals skipped or timestamps rewritten")
	}
	before := m2nodeBytes(t, db, "new-sleep-topic")
	if _, err = engine.ApplySignals(ctx, "new-sleep-topic", []domain.Signal{a, b}, ts+day); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, m2nodeBytes(t, db, "new-sleep-topic")) {
		t.Fatal("sleep signal replay applied twice")
	}
}
