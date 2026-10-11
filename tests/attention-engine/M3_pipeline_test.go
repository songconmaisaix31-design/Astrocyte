package attention_engine_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func m3Open(t *testing.T, path string) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func m3Engine(db *sqlite.DB, now time.Time) *app.DynamicsEngine {
	decay, confidence := 0.0, 0.5
	duration, observations := int64(86400000), 2
	return app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence, MergeDurationMS: &duration, MergeObservations: &observations, Clock: func() time.Time { return now }})
}

func m3Service(db *sqlite.DB, now time.Time, interpreter domain.ClusterInterpreter) *app.SleepService {
	return app.NewSleepService(sqlite.NewSleepRepository(db), m3Engine(db, now), app.SleepOptions{Clock: func() time.Time { return now }, ProjectFacts: app.LiteralSleepFactsMap, Interpreter: interpreter})
}

func m3Record(t *testing.T, db *sqlite.DB, key, topic, project string, ts time.Time, facts map[string]string) app.StoredEvent {
	t.Helper()
	raw, err := json.Marshal(struct {
		Facts    map[string]string `json:"facts"`
		Original string            `json:"original"`
	}{facts, strings.Repeat("原始审计载荷", 80)})
	if err != nil {
		t.Fatal(err)
	}
	event := domain.Event{Source: domain.EventSourceManual, Session: "m3", TS: ts.UnixMilli(), Kind: domain.EventKindCollect, Scope: project, Topic: topic, Raw: raw}
	input := app.StoredEvent{OriginKey: key, ReceivedAt: ts.UnixMilli(), Observation: app.Observation{Source: domain.EventSourceManual, Session: "m3", SourceKind: domain.EventKindCollect, Scope: project, TS: ts.UnixMilli(), Provenance: "contract_local:M3"}, Event: &event, RawBytes: raw}
	rows, err := app.NewEventService(sqlite.NewEventRepository(db), .2).Record(context.Background(), []app.StoredEvent{input})
	if err != nil {
		t.Fatal(err)
	}
	return rows[0]
}

func m3Node(t *testing.T, db *sqlite.DB, topic string) app.StoredAttentionNode {
	t.Helper()
	var node app.StoredAttentionNode
	err := sqlite.NewDynamicsRepository(db).ViewDynamics(context.Background(), func(tx app.DynamicsTx) error { var err error; node, err = tx.GetNode(topic); return err })
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func TestM3PipelineBirthSignalsMetricsDailyReplayAndRestart(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	path := filepath.Join(t.TempDir(), "m3.sqlite")
	db := m3Open(t, path)
	defer func() { db.Close() }()
	first := m3Record(t, db, "fresh-1", "fresh", "project-a", day.Add(9*time.Hour), map[string]string{"method": "邻接聚类"})
	last := m3Record(t, db, "fresh-2", "fresh", "project-a", day.Add(10*time.Hour), map[string]string{"method": "邻接聚类"})
	m3Record(t, db, "noise", "noise", "project-a", day.Add(11*time.Hour), map[string]string{"one-off": "discarded"})
	m3Record(t, db, "yesterday", "fresh", "project-a", day.Add(-time.Millisecond), map[string]string{"previous": "excluded"})
	m3Record(t, db, "tomorrow", "fresh", "project-a", day.AddDate(0, 0, 1), map[string]string{"next": "excluded"})
	pending := app.StoredEvent{OriginKey: "unknown-source", ReceivedAt: now.UnixMilli(), Observation: app.Observation{Source: "unsupported", SourceKind: "unknown", Provenance: "contract_local:M3", Scope: "project-a", TS: day.Add(12 * time.Hour).UnixMilli()}, RawBytes: []byte(`{"facts":{"unknown":"must not become a Signal"}}`)}
	if _, err := app.NewEventService(sqlite.NewEventRepository(db), .2).Record(ctx, []app.StoredEvent{pending}); err != nil {
		t.Fatal(err)
	}
	result, err := m3Service(db, now, nil).Run(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	node := m3Node(t, db, "fresh")
	// Two +.06 steps yield M=.0546; the end-of-day zero-delta step multiplies
	// it by alpha=.3 again, while lambda=0 leaves A=.17.
	if math.Abs(node.Node.Attention-.17) > 1e-12 || math.Abs(node.Node.Momentum-.01638) > 1e-12 || node.Node.LastActive != last.Event.TS {
		t.Fatalf("same-day A/M or observed activity wrong: %+v", node.Node)
	}
	for _, id := range []string{first.ID, last.ID} {
		if !containsM3Evidence(node.Node.Evidence, id) {
			t.Fatalf("missing evidence %s", id)
		}
	}
	var noiseNodes, signals int
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_nodes WHERE topic='noise'`).Scan(&noiseNodes); err != nil || noiseNodes != 0 {
		t.Fatalf("isolated node promoted: %d %v", noiseNodes, err)
	}
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_signals s JOIN attention_engine_events e ON e.id=s.event_id WHERE e.mapping_status='pending'`).Scan(&signals); err != nil || signals != 0 {
		t.Fatalf("pending Signal: %d %v", signals, err)
	}
	m := result.Run.Metrics
	if !m.RetentionMeasurable || !m.StorageMeasurable || m.KeyFactsBefore != 2 || m.KeyFactsRetained != 1 || m.RetentionPrecision != .5 {
		t.Fatalf("fact metrics: %+v", m)
	}
	var beforeBytes, afterBytes int64
	if err = db.Conn().QueryRowContext(ctx, `SELECT sum(length(raw_bytes)) FROM attention_engine_events WHERE ts>=? AND ts<=? AND mapping_status='mapped' AND topic!=''`, day.UnixMilli(), now.UnixMilli()).Scan(&beforeBytes); err != nil {
		t.Fatal(err)
	}
	if err = db.Conn().QueryRowContext(ctx, `SELECT (SELECT COALESCE(sum(length(CAST(data AS BLOB))),0) FROM attention_engine_knowledge)+(SELECT COALESCE(sum(length(CAST(data AS BLOB))),0) FROM attention_engine_sleep_candidates WHERE run_id=?)`, result.Run.ID).Scan(&afterBytes); err != nil {
		t.Fatal(err)
	}
	if m.BytesBefore != beforeBytes || m.BytesAfter != afterBytes || math.Abs(m.StorageReduction-float64(beforeBytes-afterBytes)/float64(beforeBytes)) > 1e-12 {
		t.Fatalf("bytes differ from persisted BLOB/UTF-8 representation: %+v actual=%d/%d", m, beforeBytes, afterBytes)
	}
	if m.StorageReduction <= 0 {
		t.Fatalf("repeated large input did not reduce working representation: %+v", m)
	}
	t.Logf("M3 observed working representation: facts=%d retained=%d recall=%.6f UTF8 bytes=%d -> %d reduction=%.6f; original events retained", m.KeyFactsBefore, m.KeyFactsRetained, m.RetentionPrecision, m.BytesBefore, m.BytesAfter, m.StorageReduction)
	knownBefore, conflictsBefore := m3StoredTexts(t, db, "fresh")
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestM3ColdProcessReopen$")
	child.Env = append(os.Environ(), "ASTROCYTE_M3_COLD_DB="+path, "ASTROCYTE_M3_COLD_DAY="+day.Format(time.RFC3339), "ASTROCYTE_M3_COLD_NOW="+now.Format(time.RFC3339))
	output, err := child.CombinedOutput()
	if err != nil {
		t.Fatalf("cold process: %s %v", output, err)
	}
	var cold struct {
		Run       app.EngineSleepRun
		Node      app.StoredAttentionNode
		Knowledge string
		Conflicts string
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "M3-COLD:") {
			if err = json.Unmarshal([]byte(strings.TrimPrefix(line, "M3-COLD:")), &cold); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !reflect.DeepEqual(cold.Run, result.Run) || !reflect.DeepEqual(cold.Node, node) || cold.Knowledge != knownBefore || cold.Conflicts != conflictsBefore {
		t.Fatalf("cold process changed state: %s", output)
	}
	db = m3Open(t, path)
	replayed, err := m3Service(db, now, nil).Run(ctx, day)
	if err != nil || !replayed.Replayed || !reflect.DeepEqual(replayed.Run, result.Run) {
		t.Fatalf("daily restart replay: %+v %v", replayed, err)
	}
	if got := m3Node(t, db, "fresh"); !reflect.DeepEqual(got, node) {
		t.Fatalf("repeat changed graph: before=%+v after=%+v", node, got)
	}
	var count int
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_engine_events`).Scan(&count); err != nil || count != 6 {
		t.Fatalf("original events removed: %d %v", count, err)
	}
}

func TestM3ColdProcessReopen(t *testing.T) {
	path := os.Getenv("ASTROCYTE_M3_COLD_DB")
	if path == "" {
		t.Skip("child-process entry")
	}
	day, err := time.Parse(time.RFC3339, os.Getenv("ASTROCYTE_M3_COLD_DAY"))
	if err != nil {
		t.Fatal(err)
	}
	now, err := time.Parse(time.RFC3339, os.Getenv("ASTROCYTE_M3_COLD_NOW"))
	if err != nil {
		t.Fatal(err)
	}
	db := m3Open(t, path)
	result, err := m3Service(db, now, nil).Run(context.Background(), day)
	if err != nil || !result.Replayed {
		t.Fatalf("cold replay: %+v %v", result, err)
	}
	topic := os.Getenv("ASTROCYTE_M3_COLD_TOPIC")
	if topic == "" {
		topic = "fresh"
	}
	known, conflicts := m3StoredTexts(t, db, topic)
	data, err := json.Marshal(struct {
		Run       app.EngineSleepRun
		Node      app.StoredAttentionNode
		Knowledge string
		Conflicts string
	}{result.Run, m3Node(t, db, topic), known, conflicts})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("M3-COLD:%s\n", data)
}

func m3StoredTexts(t *testing.T, db *sqlite.DB, topic string) (string, string) {
	t.Helper()
	var knowledge, conflicts string
	if err := db.Conn().QueryRowContext(context.Background(), `SELECT json_group_array(data) FROM (SELECT data FROM attention_engine_knowledge WHERE owner_project='project-a' AND topic=? ORDER BY id)`, topic).Scan(&knowledge); err != nil {
		t.Fatal(err)
	}
	if err := db.Conn().QueryRowContext(context.Background(), `SELECT json_group_array(data) FROM (SELECT data FROM attention_engine_conflicts WHERE topic=? ORDER BY id)`, topic).Scan(&conflicts); err != nil {
		t.Fatal(err)
	}
	return knowledge, conflicts
}

type m3FailRepository struct{ app.SleepRepository }
type m3FailTx struct{ app.SleepTx }

func (r m3FailRepository) WithSleepTx(ctx context.Context, fn func(app.SleepTx) error) error {
	return r.SleepRepository.WithSleepTx(ctx, func(tx app.SleepTx) error { return fn(m3FailTx{tx}) })
}
func (m3FailTx) SaveSleepRun(app.EngineSleepRun) error {
	return errors.New("M3 deliberate transaction abort")
}

func TestM3PipelineRollsBackGraphFluidAndDailyReceipt(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	db := m3Open(t, filepath.Join(t.TempDir(), "rollback.sqlite"))
	m3Record(t, db, "a", "fresh", "project-a", day.Add(9*time.Hour), map[string]string{"method": "same"})
	m3Record(t, db, "b", "fresh", "project-a", day.Add(10*time.Hour), map[string]string{"method": "same"})
	s := app.NewSleepService(m3FailRepository{sqlite.NewSleepRepository(db)}, m3Engine(db, now), app.SleepOptions{Clock: func() time.Time { return now }, ProjectFacts: app.LiteralSleepFactsMap})
	if _, err := s.Run(ctx, day); err == nil || !strings.Contains(err.Error(), "deliberate transaction abort") {
		t.Fatalf("expected final write failure: %v", err)
	}
	for _, table := range []string{"attention_engine_nodes", "attention_engine_knowledge", "attention_engine_sleep_runs", "attention_engine_sleep_candidates", "attention_engine_sleep_metrics"} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("partial %s: %d %v", table, count, err)
		}
	}
	var events int
	if err := db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM attention_engine_events").Scan(&events); err != nil || events != 2 {
		t.Fatalf("original events changed: %d %v", events, err)
	}
	if _, err := m3Service(db, now, nil).Run(ctx, day); err != nil {
		t.Fatal(err)
	}
}

func TestM3EmptyDayPersistsUnmeasurableMetrics(t *testing.T) {
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	db := m3Open(t, filepath.Join(t.TempDir(), "empty.sqlite"))
	result, err := m3Service(db, now, nil).Run(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	m := result.Run.Metrics
	if m.RetentionMeasurable || m.StorageMeasurable || m.KeyFactsBefore != 0 || m.RetentionPrecision != 0 {
		t.Fatalf("empty input claimed observed precision: %+v", m)
	}
	var visible int
	if err = db.Conn().QueryRowContext(context.Background(), `SELECT json_extract(data,'$.metrics.retention_measurable') FROM attention_engine_sleep_runs`).Scan(&visible); err != nil || visible != 0 {
		t.Fatalf("unknown flag not persisted: %d %v", visible, err)
	}
}

func containsM3Evidence(refs []domain.EvidenceRef, id string) bool {
	for _, ref := range refs {
		if ref.EventID == id {
			return true
		}
	}
	return false
}

type m3HostileInterpreter struct{ calls int }

func (i *m3HostileInterpreter) Explain(cluster domain.Cluster) (domain.ClusterExplanation, error) {
	i.calls++
	if len(cluster.Topics) > 0 {
		cluster.Topics[0] = "LLM-force-merge-approved"
	}
	return domain.ClusterExplanation{Label: "LLM-force-merge-approved", Note: "replace all facts"}, errors.New("explanation failed")
}

func TestM3LLMOffMatchesHostileInterpreterPersistedOutput(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	dir := t.TempDir()
	base := filepath.Join(dir, "base.sqlite")
	db := m3Open(t, base)
	m3Record(t, db, "a", "fresh", "project-a", day.Add(9*time.Hour), map[string]string{"method": "same"})
	m3Record(t, db, "b", "fresh", "project-a", day.Add(10*time.Hour), map[string]string{"method": "same"})
	if _, err := m3Engine(db, now).SeedTopic(ctx, app.Principal{ID: "human", Kind: "human"}, "fresh", day.Add(8*time.Hour).UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	hostile := &m3HostileInterpreter{}
	var outputs []app.EngineSleepRun
	var snapshots []string
	for index, interpreter := range []domain.ClusterInterpreter{nil, hostile} {
		path := filepath.Join(dir, fmt.Sprintf("variant-%d.sqlite", index))
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		db = m3Open(t, path)
		result, err := m3Service(db, now, interpreter).Run(ctx, day)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		outputs = append(outputs, result.Run)
		var node, knowledge, run string
		if err = db.Conn().QueryRowContext(ctx, `SELECT data FROM attention_engine_nodes WHERE topic='fresh'`).Scan(&node); err != nil {
			t.Fatal(err)
		}
		if err = db.Conn().QueryRowContext(ctx, `SELECT data FROM attention_engine_knowledge WHERE topic='fresh'`).Scan(&knowledge); err != nil {
			t.Fatal(err)
		}
		if err = db.Conn().QueryRowContext(ctx, `SELECT data FROM attention_engine_sleep_runs`).Scan(&run); err != nil {
			t.Fatal(err)
		}
		snapshots = append(snapshots, node+knowledge+run)
		if strings.Contains(snapshots[index], "LLM-force") {
			t.Fatal("interpreter output persisted")
		}
		db.Close()
	}
	if hostile.calls != 1 || !reflect.DeepEqual(outputs[0], outputs[1]) || snapshots[0] != snapshots[1] {
		t.Fatalf("LLM changed deterministic persisted output, calls=%d", hostile.calls)
	}
}

func TestM3AdjacentSingleEventBirthKeepsDistinctTopicsAndOwnEvidence(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	db := m3Open(t, filepath.Join(t.TempDir(), "adjacent.sqlite"))
	engine := m3Engine(db, now)
	if _, err := engine.SeedTopic(ctx, app.Principal{ID: "human", Kind: "human"}, "swarm", day.Add(7*time.Hour).UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewDynamicsRepository(db).WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		node, err := tx.GetNode("swarm")
		if err != nil {
			return err
		}
		node.Node.RelatedTopics = map[string]float64{"paper": .9}
		node.Version++
		return tx.SaveNode(node)
	}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		m3Record(t, db, fmt.Sprintf("swarm-%d", i), "swarm", "project-a", day.Add(time.Duration(8+i)*time.Hour), map[string]string{"swarm": "distributed"})
	}
	paper := m3Record(t, db, "paper-1", "paper", "project-a", day.Add(11*time.Hour), map[string]string{"paper": "reading"})
	m3Record(t, db, "one-off", "noise", "project-a", day.Add(12*time.Hour), map[string]string{"noise": "one-off"})
	if _, err := m3Service(db, now, nil).Run(ctx, day); err != nil {
		t.Fatal(err)
	}
	p := m3Node(t, db, "paper")
	swarm := m3Node(t, db, "swarm")
	if !containsM3Evidence(p.Node.Evidence, paper.ID) || p.Node.RelatedTopics["swarm"] != .9 || swarm.Node.RelatedTopics["paper"] != .9 {
		t.Fatalf("association or single-topic birth missing: %+v %+v", p, swarm)
	}
	for _, node := range []app.StoredAttentionNode{p, swarm} {
		for _, ref := range node.Node.Evidence {
			event, err := sqlite.NewEventRepository(db).GetEvent(ctx, ref.EventID)
			if err != nil || event.Event.Topic != node.Node.Topic {
				t.Fatalf("borrowed neighbor evidence on %s: %+v %v", node.Node.Topic, event, err)
			}
		}
	}
	if _, err := m3Service(db, now.AddDate(0, 0, 1), nil).Run(ctx, day.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}
	for _, topic := range []string{"paper", "swarm"} {
		if got := m3Node(t, db, topic); got.Lifecycle == "cold" {
			t.Fatalf("disjoint topics merged: %+v", got)
		}
	}
}

func TestM3KnowledgeConflictRetainsConfirmedSnapshotAndRequiresHuman(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	path := filepath.Join(t.TempDir(), "conflict.sqlite")
	db := m3Open(t, path)
	for i := 0; i < 2; i++ {
		m3Record(t, db, fmt.Sprintf("left-%d", i), "design", "project-a", day.Add(time.Duration(9+i)*time.Hour), map[string]string{"method": "centralized"})
	}
	if _, err := m3Service(db, now, nil).Run(ctx, day); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewKnowledgeRepository(sqlite.NewEngineRepository(db))
	var old app.StoredKnowledge
	if err := repo.WithKnowledgeTx(ctx, func(tx app.KnowledgeTx) error {
		rows, err := tx.ListKnowledge("project-a", false)
		if err != nil {
			return err
		}
		old = rows[0]
		expected := old.Version
		old.Entry.Crystallized = json.RawMessage(`{"facts":{"method":"centralized"}}`)
		old.CrystallizedVersions = []app.KnowledgeVersion{{Version: 1, Content: old.Entry.Crystallized, RecordedAt: now.UnixMilli()}}
		old.Version++
		return tx.SaveKnowledge(old, expected)
	}); err != nil {
		t.Fatal(err)
	}
	knowledge := app.NewKnowledgeService(repo, app.KnowledgeServiceOptions{Clock: func() time.Time { return now }})
	if _, err := knowledge.ConfirmKnowledge(ctx, app.Principal{ID: "human", Kind: "human"}, old.Entry.ID, app.ConfirmKnowledgeCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "confirm", IdempotencyKey: "confirm", ExpectedVersion: old.Version}, OwnerProject: "project-a"}); err != nil {
		t.Fatal(err)
	}
	old, err := knowledge.GetKnowledge(ctx, app.Principal{ID: "human", Kind: "human"}, old.Entry.ID, "project-a")
	if err != nil {
		t.Fatal(err)
	}
	second := day.AddDate(0, 0, 1)
	secondNow := now.AddDate(0, 0, 1)
	for i := 0; i < 2; i++ {
		m3Record(t, db, fmt.Sprintf("right-%d", i), "design", "project-a", second.Add(time.Duration(9+i)*time.Hour), map[string]string{"method": "distributed"})
	}
	result, err := m3Service(db, secondNow, nil).Run(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Run.ConflictIDs) != 1 {
		t.Fatalf("conflict queue: %+v", result.Run)
	}
	current, err := knowledge.GetKnowledge(ctx, app.Principal{ID: "human", Kind: "human"}, old.Entry.ID, "project-a")
	if err != nil || !reflect.DeepEqual(current, old) {
		t.Fatalf("conflict overwrote fluid/status/snapshot: %+v %v", current, err)
	}
	if result.Run.Metrics.KeyFactsBefore != 2 || result.Run.Metrics.KeyFactsRetained != 2 || result.Run.Metrics.RetentionPrecision != 1 {
		t.Fatalf("pending claims lost: %+v", result.Run.Metrics)
	}
	t.Logf("M3 pending conflict: original confirmed entry unchanged; facts=%d retained=%d recall=%.6f, includes quarantined literal claims", result.Run.Metrics.KeyFactsBefore, result.Run.Metrics.KeyFactsRetained, result.Run.Metrics.RetentionPrecision)
	knownBefore, conflictsBefore := m3StoredTexts(t, db, "design")
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestM3ColdProcessReopen$")
	child.Env = append(os.Environ(), "ASTROCYTE_M3_COLD_DB="+path, "ASTROCYTE_M3_COLD_TOPIC=design", "ASTROCYTE_M3_COLD_DAY="+second.Format(time.RFC3339), "ASTROCYTE_M3_COLD_NOW="+secondNow.Format(time.RFC3339))
	output, err := child.CombinedOutput()
	if err != nil {
		t.Fatalf("pending-conflict cold process: %s %v", output, err)
	}
	var cold struct {
		Run       app.EngineSleepRun
		Knowledge string
		Conflicts string
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "M3-COLD:") {
			if err = json.Unmarshal([]byte(strings.TrimPrefix(line, "M3-COLD:")), &cold); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !reflect.DeepEqual(cold.Run, result.Run) || cold.Knowledge != knownBefore || cold.Conflicts != conflictsBefore {
		t.Fatalf("pending conflict/confirmed knowledge changed in cold process: %s", output)
	}
	db = m3Open(t, path)
	engine := m3Engine(db, secondNow)
	command := app.AdjudicateConflictCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "dismiss", IdempotencyKey: "dismiss", ExpectedVersion: 1}, Decision: app.ConflictDecisionDismiss}
	if _, err = engine.AdjudicateConflict(ctx, app.Principal{ID: "agent", Kind: "agent"}, result.Run.ConflictIDs[0], command); err == nil {
		t.Fatal("Agent adjudicated conflict")
	}
	decision, err := engine.AdjudicateConflict(ctx, app.Principal{ID: "human", Kind: "human"}, result.Run.ConflictIDs[0], command)
	if err != nil || decision.Conflict.Status != domain.ConflictStatusDismissed {
		t.Fatalf("human gate: %+v %v", decision, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = m3Open(t, path)
	list, err := m3Engine(db, secondNow).Conflicts(ctx)
	if err != nil || len(list) != 1 || list[0].Conflict.Status != domain.ConflictStatusDismissed || len(list[0].Actions) != 1 {
		t.Fatalf("human decision cold persistence: %+v %v", list, err)
	}
	replayed, err := m3Service(db, secondNow, nil).Run(ctx, second)
	if err != nil || !replayed.Replayed || !reflect.DeepEqual(replayed.Run, result.Run) {
		t.Fatalf("human decision overwritten by daily replay: %+v %v", replayed, err)
	}
}

func TestM3CompactInputReportsActualStorageGrowth(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	db := m3Open(t, filepath.Join(t.TempDir(), "growth.sqlite"))
	service := app.NewEventService(sqlite.NewEventRepository(db), .2)
	for i := 0; i < 2; i++ {
		ts := day.Add(time.Duration(9+i) * time.Hour).UnixMilli()
		raw := json.RawMessage(`{"facts":{"x":"1"}}`)
		event := domain.Event{Source: "manual", Session: "m3", TS: ts, Kind: "collect", Scope: "project-a", Topic: "compact", Raw: raw}
		if _, err := service.Record(ctx, []app.StoredEvent{{OriginKey: fmt.Sprintf("compact-%d", i), ReceivedAt: ts, Observation: app.Observation{Source: "manual", SourceKind: "collect", Provenance: "contract_local:M3", Scope: "project-a", TS: ts}, Event: &event, RawBytes: raw}}); err != nil {
			t.Fatal(err)
		}
	}
	result, err := m3Service(db, now, nil).Run(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	m := result.Run.Metrics
	if !m.StorageMeasurable || m.BytesAfter <= m.BytesBefore || m.StorageReduction >= 0 {
		t.Fatalf("growth hidden: %+v", m)
	}
	t.Logf("M3 observed growth: UTF8 bytes=%d -> %d reduction=%.6f", m.BytesBefore, m.BytesAfter, m.StorageReduction)
	var bytes int64
	if err = db.Conn().QueryRowContext(ctx, `SELECT (SELECT sum(length(CAST(data AS BLOB))) FROM attention_engine_knowledge)+(SELECT sum(length(CAST(data AS BLOB))) FROM attention_engine_sleep_candidates WHERE run_id=?)`, result.Run.ID).Scan(&bytes); err != nil || bytes != m.BytesAfter {
		t.Fatalf("growth bytes not actual persisted UTF-8: %d %+v %v", bytes, m, err)
	}
}

func TestM3UndecidedMergePolicyPersistsPendingAssociation(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	now := day.Add(23 * time.Hour)
	db := m3Open(t, filepath.Join(t.TempDir(), "undecided.sqlite"))
	engine := m3Engine(db, now)
	for _, topic := range []string{"left", "right"} {
		if _, err := engine.SeedTopic(ctx, app.Principal{ID: "human", Kind: "human"}, topic, day.Add(8*time.Hour).UnixMilli()); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			m3Record(t, db, fmt.Sprintf("%s-%d", topic, i), topic, "project-a", day.Add(time.Duration(9+i)*time.Hour), map[string]string{"fact": topic})
		}
	}
	if err := sqlite.NewDynamicsRepository(db).WithDynamicsTx(ctx, func(tx app.DynamicsTx) error {
		n, err := tx.GetNode("left")
		if err != nil {
			return err
		}
		n.Node.RelatedTopics = map[string]float64{"right": .9}
		n.Version++
		return tx.SaveNode(n)
	}); err != nil {
		t.Fatal(err)
	}
	decay, confidence := 0.0, 0.5
	unconfigured := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), app.DynamicsOptions{Decay: &decay, InitialConfidence: &confidence})
	service := app.NewSleepService(sqlite.NewSleepRepository(db), unconfigured, app.SleepOptions{Clock: func() time.Time { return now }, ProjectFacts: app.LiteralSleepFactsMap})
	result, err := service.Run(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	var candidates []app.SleepCandidate
	if err = json.Unmarshal(result.Run.Candidates, &candidates); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range candidates {
		if candidate.Kind == "association_pending" && candidate.Note == "merge_policy_undecided" {
			found = true
		}
	}
	if !found {
		t.Fatalf("undecided policy hidden: %s", result.Run.Candidates)
	}
	if m3Node(t, db, "left").Lifecycle == "cold" || m3Node(t, db, "right").Lifecycle == "cold" {
		t.Fatal("unset merge policy made a decision")
	}
}
