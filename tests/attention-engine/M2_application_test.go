package attention_engine_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func m2ptr[T any](v T) *T { return &v }
func m2options() app.DynamicsOptions {
	return app.DynamicsOptions{Decay: m2ptr(.1), InitialConfidence: m2ptr(.5), SuccessConfidenceDelta: m2ptr(.1), ConfidenceDropThreshold: m2ptr(.05), DensityThreshold: m2ptr(1.0), SplitThreshold: m2ptr(.3), SplitState: func(parent domain.AttentionNode, _ int) domain.AttentionNode { return parent }, MergeDurationMS: m2ptr(int64(86400000)), MergeObservations: m2ptr(2), Clock: func() time.Time { return time.UnixMilli(1767225600000) }}
}

var m2human = app.Principal{ID: "m2-human", Kind: "human"}

//go:embed M2_application_golden.json
var applicationGolden []byte

func m2open(t *testing.T, path string) (*sqlite.DB, *app.DynamicsEngine) {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db, app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), m2options())
}

// The M2 task book prescribes fake inputs. These rows retain original bytes and
// use the real W1 receipt/normalisation service, migration, repository and app.
func m2signal(t *testing.T, db *sqlite.DB, topic, id string, ts int64, class domain.SignalClass, strength float64) domain.Signal {
	t.Helper()
	kind := map[domain.SignalClass]string{domain.SignalImplicitRegular: domain.EventKindEditFile, domain.SignalImplicitDepth: domain.EventKindReadPaper, domain.SignalImplicitLight: domain.EventKindBookmark, domain.SignalExplicitPositive: domain.EventKindUserApprove, domain.SignalExplicitNegative: domain.EventKindUserVeto, domain.SignalFeedbackSuccess: domain.EventKindAgentTask, domain.SignalFeedbackFailure: domain.EventKindAgentTask}[class]
	raw := []byte(fmt.Sprintf("{ \"fixture_event\": %q, \"depth\": \"long\" }\n", id))
	event := domain.Event{Source: domain.EventSourceCodex, Session: "m2-synthetic", TS: ts, Kind: kind, Scope: "m2-project", Topic: topic, Raw: raw}
	outcome := "unknown"
	if class == domain.SignalFeedbackSuccess {
		outcome = "success"
	}
	if class == domain.SignalFeedbackFailure {
		outcome = "failure"
	}
	if class == domain.SignalExplicitPositive || class == domain.SignalExplicitNegative {
		event.Source = domain.EventSourceManual
	}
	repo := sqlite.NewEventRepository(db)
	records, err := app.NewEventService(repo, strength).Record(context.Background(), []app.StoredEvent{{OriginKey: id, ReceivedAt: ts, Event: &event, RawBytes: raw, Observation: app.Observation{Source: event.Source, SourceKind: kind, Session: event.Session, Scope: event.Scope, TS: ts, Outcome: outcome, Depth: "deep", Provenance: "M2 synthetic application acceptance"}}})
	if err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListSignals(context.Background(), ts, ts)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range list {
		if s.EventID == records[0].ID {
			if s.Class != class {
				t.Fatal("unexpected W1 signal class", s.Class)
			}
			return s
		}
	}
	t.Fatal("W1 did not persist the normalized signal")
	return domain.Signal{}
}

func m2nodeBytes(t *testing.T, db *sqlite.DB, topic string) []byte {
	t.Helper()
	var data string
	if err := db.Conn().QueryRow("SELECT data FROM attention_engine_nodes WHERE topic=?", topic).Scan(&data); err != nil {
		t.Fatal(err)
	}
	return []byte(data)
}

func TestM2ApplicationThirtyDaySQLiteReplay(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	if retained := os.Getenv("ASTROCYTE_M2_DB"); retained != "" {
		if _, err := os.Stat(retained); !os.IsNotExist(err) {
			t.Fatal("retained M2 database must be a fresh path")
		}
		path = retained
	}
	db, engine := m2open(t, path)
	defer func() { db.Close() }()
	var doc goldenDoc
	if err := json.Unmarshal(applicationGolden, &doc); err != nil {
		t.Fatal(err)
	}
	initial, err := engine.SeedTopic(ctx, m2human, "Agent Swarm", doc.Input[0].TsMs-doc.DayMs)
	if err != nil {
		t.Fatal(err)
	}
	if initial.Attention != domain.AttentionFloor {
		t.Fatal("manual seed persisted below floor")
	}
	type step struct {
		TS      int64                   `json:"ts"`
		Signals []domain.Signal         `json:"signals"`
		State   app.StoredAttentionNode `json:"state"`
	}
	steps := []step{}
	for i, in := range doc.Input {
		signals := []domain.Signal{}
		class := map[float64]domain.SignalClass{.6: domain.SignalImplicitRegular, 1: domain.SignalImplicitDepth, .3: domain.SignalImplicitLight, .8: domain.SignalFeedbackSuccess, 1.5: domain.SignalExplicitPositive, -.8: domain.SignalExplicitNegative}[in.SignalSum]
		if in.ConfidenceDelta < 0 {
			class = domain.SignalFeedbackFailure
		}
		if class != "" {
			signals = append(signals, m2signal(t, db, "Agent Swarm", fmt.Sprintf("m2-day-%02d", i), in.TsMs, class, 1))
		}
		n, err := engine.ApplySignals(ctx, "Agent Swarm", signals, in.TsMs)
		if err != nil {
			t.Fatalf("day %d: %v", i, err)
		}
		want := doc.Output[i]
		if math.Abs(n.Attention-want.Attention) > 1e-12 || math.Abs(n.Momentum-want.Momentum) > 1e-12 || math.Abs(n.Confidence-want.Confidence) > 1e-12 {
			t.Fatalf("day %d application A/M/C=(%.16f,%.16f,%.16f), independent golden=(%.16f,%.16f,%.16f)", i, n.Attention, n.Momentum, n.Confidence, want.Attention, want.Momentum, want.Confidence)
		}
		before := m2nodeBytes(t, db, n.Topic)
		if _, err = engine.ApplySignals(ctx, n.Topic, signals, in.TsMs); err != nil {
			t.Fatal(err)
		}
		if _, err = engine.AttentionState(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err = engine.ExplainFocus(ctx, n.Topic); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, m2nodeBytes(t, db, n.Topic)) {
			t.Fatalf("day %d replay/read wrote attention state", i)
		}
		saved, err := engine.StoredNode(ctx, n.Topic)
		if err != nil {
			t.Fatal(err)
		}
		if saved.UpdatedAt != in.TsMs {
			t.Fatal("decay time not persisted")
		}
		if len(signals) == 0 && i > 0 && saved.Node.LastActive != steps[len(steps)-1].State.Node.LastActive {
			t.Fatal("idle step refreshed last_active")
		}
		steps = append(steps, step{TS: in.TsMs, Signals: signals, State: saved})
		if i == 10 || i == 20 {
			if err = db.Close(); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestM2ProcessReload$", "-test.v")
			cmd.Env = append(os.Environ(), "ASTROCYTE_M2_RELOAD_PATH="+path, "ASTROCYTE_M2_RELOAD_TIME="+strconv.FormatInt(in.TsMs, 10))
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("new process reload: %v\n%s", err, out)
			}
			db, engine = m2open(t, path)
			if _, err = engine.ApplySignals(ctx, n.Topic, signals, in.TsMs); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, m2nodeBytes(t, db, n.Topic)) {
				t.Fatal("restart/replay decayed twice")
			}
		}
	}
	explanation, err := engine.ExplainFocus(ctx, "Agent Swarm")
	if err != nil {
		t.Fatal(err)
	}
	if len(explanation.Evidence) != 9 {
		t.Fatalf("explanation has %d references", len(explanation.Evidence))
	}
	for _, s := range explanation.Evidence {
		if s.Original.Event == nil || len(s.Original.RawBytes) == 0 || s.Strength != 1 {
			t.Fatal("incomplete weighted original evidence")
		}
		if s.ConfidenceOnly && (s.AttentionContribution != 0 || s.ConfidenceContribution != -.1) {
			t.Fatal("failure weighted attention")
		}
	}
	if output := os.Getenv("ASTROCYTE_M2_OUTPUT"); output != "" {
		b, err := json.MarshalIndent(struct {
			EvidenceClass  string               `json:"evidence_class"`
			DBRestartDays  []int                `json:"db_restart_days"`
			ProcessRestart bool                 `json:"process_restart"`
			Initial        domain.AttentionNode `json:"initial"`
			Steps          []step               `json:"steps"`
			Explanation    app.FocusExplanation `json:"explanation"`
		}{"synthetic_application_sqlite", []int{10, 20}, true, initial, steps, explanation}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(output, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// Launched by the acceptance test as a new Go process against the closed DB.
func TestM2ProcessReload(t *testing.T) {
	path := os.Getenv("ASTROCYTE_M2_RELOAD_PATH")
	if path == "" {
		t.Skip("child-process entry used by the 30-day acceptance")
	}
	now, err := strconv.ParseInt(os.Getenv("ASTROCYTE_M2_RELOAD_TIME"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	db, engine := m2open(t, path)
	defer db.Close()
	before := m2nodeBytes(t, db, "Agent Swarm")
	if _, err = engine.ExplainFocus(context.Background(), "Agent Swarm"); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ApplySignals(context.Background(), "Agent Swarm", nil, now); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, m2nodeBytes(t, db, "Agent Swarm")) {
		t.Fatal("cold process applied idle decay twice")
	}
}

func TestM2FailureBatchEvidenceAndAtomicRollback(t *testing.T) {
	ctx := context.Background()
	db, engine := m2open(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	const ts = int64(1767225600000)
	if _, err := engine.SeedTopic(ctx, m2human, "focus", ts); err != nil {
		t.Fatal(err)
	}
	a := m2signal(t, db, "focus", "a", ts, domain.SignalImplicitLight, 1)
	b := m2signal(t, db, "focus", "b", ts, domain.SignalImplicitLight, 1)
	n, err := engine.ApplySignals(ctx, "focus", []domain.Signal{a, b, a}, ts)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(n.Attention-.65) > 1e-12 || math.Abs(n.Momentum-.42) > 1e-12 {
		t.Fatal("batch did not sum one step or counted duplicate")
	}
	failure := m2signal(t, db, "focus", "fail", ts, domain.SignalFeedbackFailure, 1)
	f, err := engine.ApplySignals(ctx, "focus", []domain.Signal{failure}, ts)
	if err != nil {
		t.Fatal(err)
	}
	if f.Attention != n.Attention || f.Momentum != n.Momentum || f.Confidence != .4 {
		t.Fatal("failure did not lower confidence only")
	}
	before := m2nodeBytes(t, db, "focus")
	good := m2signal(t, db, "focus", "good", ts+86400000, domain.SignalImplicitRegular, 1)
	bad := good
	bad.EventID = "missing"
	if _, err = engine.ApplySignals(ctx, "focus", []domain.Signal{good, bad}, good.TS); err == nil {
		t.Fatal("missing evidence accepted")
	}
	if !bytes.Equal(before, m2nodeBytes(t, db, "focus")) {
		t.Fatal("partial failed batch persisted")
	}
	if _, err = engine.ApplySignals(ctx, "focus", []domain.Signal{good}, good.TS); err != nil {
		t.Fatal(err)
	}
	if _, err = engine.ApplySignals(ctx, "focus", []domain.Signal{a}, ts); err != nil {
		t.Fatal("old already-applied replay rejected", err)
	}
}
