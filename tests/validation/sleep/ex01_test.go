package sleep_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

var day = time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
var human = app.Principal{ID: "fixture-human", Kind: "human"}

type fixture struct {
	db      *sqlite.DB
	engine  *app.DynamicsEngine
	service *app.SleepService
	now     time.Time
}

func open(t *testing.T) *fixture {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "sleep.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	f := &fixture{db: db, now: day.Add(23 * time.Hour)}
	decay, confidence := 0.0, .5
	duration, observations := int64(86400000), 2
	f.engine = app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence, MergeDurationMS: &duration, MergeObservations: &observations, Clock: func() time.Time { return f.now }})
	f.service = app.NewSleepService(sqlite.NewSleepRepository(db), f.engine, app.SleepOptions{Clock: func() time.Time { return f.now }, ProjectFacts: app.LiteralSleepFactsMap})
	return f
}

func (f *fixture) seed(t *testing.T, topic string) {
	t.Helper()
	if _, err := f.engine.SeedTopic(context.Background(), human, topic, day.Add(7*time.Hour).UnixMilli()); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) record(t *testing.T, key, topic string, mapped bool, hour int) app.StoredEvent {
	t.Helper()
	ts := day.Add(time.Duration(hour) * time.Hour).UnixMilli()
	raw := []byte(`{"facts":{"method":"fixture-ex01"}}`)
	record := app.StoredEvent{OriginKey: "fixture:" + key, ReceivedAt: ts, RawBytes: raw, Observation: app.Observation{Source: "manual", SourceKind: "collect", Session: "fixture", Scope: "project-a", TS: ts, Outcome: "unknown", Depth: "unknown", Provenance: "contract_local:EX01"}}
	if mapped {
		record.Event = &domain.Event{Source: "manual", Session: "fixture", Scope: "project-a", TS: ts, Kind: "collect", Topic: topic}
	}
	rows, err := app.NewEventService(sqlite.NewEventRepository(f.db), .2).Record(context.Background(), []app.StoredEvent{record})
	if err != nil || len(rows) != 1 {
		t.Fatalf("record: %v", err)
	}
	return rows[0]
}

func (f *fixture) nodeCount(t *testing.T) int {
	t.Helper()
	var count int
	if err := f.db.Conn().QueryRow("SELECT count(*) FROM attention_engine_nodes").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func candidates(t *testing.T, result app.SleepResult) []app.SleepCandidate {
	t.Helper()
	var rows []app.SleepCandidate
	if err := json.Unmarshal(result.Run.Candidates, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestEX01TaggedOnlyCannotBirth(t *testing.T) {
	for _, adjacent := range []bool{false, true} {
		t.Run(fmt.Sprintf("adjacent=%v", adjacent), func(t *testing.T) {
			f := open(t)
			before := 0
			if adjacent {
				f.seed(t, "existing")
				before = 1
				f.record(t, "existing-1", "existing", true, 8)
				f.record(t, "existing-2", "existing", true, 9)
				if err := sqlite.NewDynamicsRepository(f.db).WithDynamicsTx(context.Background(), func(tx app.DynamicsTx) error {
					n, err := tx.GetNode("existing")
					if err != nil {
						return err
					}
					n.Node.RelatedTopics = map[string]float64{"tagged-missing": .9}
					n.Version++
					return tx.SaveNode(n)
				}); err != nil {
					t.Fatal(err)
				}
			}
			f.record(t, "missing-1", "tagged-missing", true, 10)
			if !adjacent {
				f.record(t, "missing-2", "tagged-missing", true, 11)
			}
			result, err := f.service.Run(context.Background(), day)
			if err != nil {
				t.Fatal(err)
			}
			if got := f.nodeCount(t); got != before {
				t.Fatalf("tagged-only birth: nodes=%d before=%d", got, before)
			}
			found := false
			for _, c := range candidates(t, result) {
				if c.Topic == "tagged-missing" && c.Kind == "concept_pending" && c.Note == "tagged_birth_disabled_v0_2" {
					found = true
				}
			}
			if !found {
				t.Fatal("tagged candidate disposition missing")
			}
		})
	}
}

func TestEX01UnknownInputsRetainReceiptReferencesWithoutInventingTopics(t *testing.T) {
	f := open(t)
	unassigned := f.record(t, "unassigned-1", "", true, 8)
	other := f.record(t, "unassigned-2", "", true, 9)
	pending := f.record(t, "observation-only", "", false, 10)
	unknownTime := f.record(t, "observation-unknown-time", "", false, 11)
	// Legacy rows can have a storage window timestamp but no known timestamp
	// in their observation. Intake must not substitute ReceivedAt for this.
	if _, err := f.db.Conn().Exec("UPDATE attention_engine_events SET observation=json_set(observation,'$.ts',0) WHERE id=?", unknownTime.ID); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewEventRepository(f.db)
	before, err := repo.ListEvents(context.Background(), day.UnixMilli(), f.now.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	result, err := f.service.Run(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.nodeCount(t); got != 0 {
		t.Fatalf("invented unassigned topic: %d", got)
	}
	want := map[string]string{unassigned.ID: "unassigned_pending", other.ID: "unassigned_pending", pending.ID: "mapping_pending", unknownTime.ID: "mapping_pending"}
	seen := map[string]bool{}
	for _, c := range candidates(t, result) {
		if c.Kind != "unassigned_pending" && c.Kind != "mapping_pending" {
			continue
		}
		wantNote := "topic_candidate_unavailable"
		if c.Kind == "mapping_pending" {
			wantNote = "mapping_pending_bucket_unknown"
		}
		if len(c.Sources) == 1 && c.Sources[0].EventID == unknownTime.ID {
			wantNote += ";observed_ts_unknown"
		}
		if c.Topic != "" || c.Note != wantNote || len(c.Sources) != 1 {
			t.Fatal("unknown candidate was labeled or lost original reference")
		}
		ref := c.Sources[0]
		if want[ref.EventID] != c.Kind || seen[ref.EventID] {
			t.Fatal("input bucket/reference duplicated or changed")
		}
		seen[ref.EventID] = true
		row, err := repo.GetEvent(context.Background(), ref.EventID)
		if err != nil {
			t.Fatal(err)
		}
		if ref.TS != row.Observation.TS {
			t.Fatal("candidate uses ingestion time instead of observed time")
		}
	}
	if len(seen) != len(want) {
		t.Fatalf("Sleep skipped unknown inputs: retained=%d expected=%d", len(seen), len(want))
	}
	after, err := repo.ListEvents(context.Background(), day.UnixMilli(), f.now.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("Sleep relabeled original events")
	}
	replayed, err := f.service.Run(context.Background(), day)
	if err != nil || !replayed.Replayed || !reflect.DeepEqual(result.Run, replayed.Run) {
		t.Fatalf("daily pending receipt replay: %v", err)
	}
	t.Log("real SQLite + production Sleep retain unlabeled receipt references; candidate discovery and positive unassigned birth remain unavailable")
}

func TestEX01MixedPreservesTaggedFluidDenoiseAndSustainedMerge(t *testing.T) {
	f := open(t)
	ctx := context.Background()
	for _, topic := range []string{"left", "right", "noise"} {
		f.seed(t, topic)
	}
	shared := f.record(t, "left-1", "left", true, 8)
	f.record(t, "left-2", "left", true, 9)
	f.record(t, "right-1", "right", true, 10)
	f.record(t, "right-2", "right", true, 11)
	f.record(t, "noise-1", "noise", true, 12)
	f.record(t, "tagged-missing", "tagged-missing", true, 13)
	unassigned := f.record(t, "unassigned", "", true, 14)
	f.record(t, "observation", "", false, 15)
	// Existing production consolidation can attach shared evidence to an
	// existing topic; the EX-01 birth gate must not disable this merge support.
	if err := sqlite.NewDynamicsRepository(f.db).WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		if _, err := f.engine.ConsolidateTopicInTx(tx, "right", []domain.EvidenceRef{{EventID: shared.ID, TS: shared.Event.TS}}, nil, day.Add(8*time.Hour).UnixMilli()); err != nil {
			return err
		}
		n, err := tx.GetNode("left")
		if err != nil {
			return err
		}
		n.Node.RelatedTopics = map[string]float64{"right": .9, "tagged-missing": .9}
		n.Version++
		return tx.SaveNode(n)
	}); err != nil {
		t.Fatal(err)
	}
	result, err := f.service.Run(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	if f.nodeCount(t) != 3 {
		t.Fatal("mixed input birthed a tagged or fabricated unassigned topic")
	}
	for _, topic := range []string{"left", "right"} {
		var data string
		if err := f.db.Conn().QueryRow("SELECT data FROM attention_engine_knowledge WHERE topic=?", topic).Scan(&data); err != nil {
			t.Fatalf("tagged fluid skipped for %s: %v", topic, err)
		}
		var entry domain.KnowledgeEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			t.Fatal(err)
		}
		var fluid struct{ Facts []app.SleepFact }
		if err := json.Unmarshal([]byte(entry.Fluid), &fluid); err != nil {
			t.Fatal(err)
		}
		if len(fluid.Facts) != 1 || fluid.Facts[0].Value != "fixture-ex01" {
			t.Fatal("tagged fluid facts changed")
		}
	}
	var noiseKnowledge int
	if err := f.db.Conn().QueryRow("SELECT count(*) FROM attention_engine_knowledge WHERE topic='noise'").Scan(&noiseKnowledge); err != nil || noiseKnowledge != 0 {
		t.Fatalf("isolated tagged denoise changed: %d %v", noiseKnowledge, err)
	}
	for _, topic := range []string{"left", "right"} {
		n, err := f.engine.StoredNode(ctx, topic)
		if err != nil {
			t.Fatal(err)
		}
		for _, ref := range n.Node.Evidence {
			if ref.EventID == unassigned.ID {
				t.Fatal("unassigned neighbor became tagged evidence")
			}
		}
	}
	found := false
	for _, c := range candidates(t, result) {
		if c.Kind == "unassigned_pending" && len(c.Sources) == 1 && c.Sources[0].EventID == unassigned.ID {
			found = true
		}
	}
	if !found {
		t.Error("mixed Sleep omitted unassigned receipt")
	}
	f.now = f.now.AddDate(0, 0, 1)
	if _, err := f.service.Run(ctx, day.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	left, err := f.engine.StoredNode(ctx, "left")
	if err != nil {
		t.Fatal(err)
	}
	right, err := f.engine.StoredNode(ctx, "right")
	if err != nil {
		t.Fatal(err)
	}
	var history app.DynamicsHistory
	if err := json.Unmarshal(right.History, &history); err != nil {
		t.Fatal(err)
	}
	if right.Lifecycle != "cold" || !reflect.DeepEqual(history.RetiredTo, []string{"left"}) || left.Lifecycle == "cold" {
		t.Fatal("tagged sustained merge path blocked")
	}
	if f.nodeCount(t) != 3 {
		t.Fatal("merge created or removed historical nodes")
	}
}
