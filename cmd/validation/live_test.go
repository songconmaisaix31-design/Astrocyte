package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sleepruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/attentionruntime"
	recordruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/internal/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func liveFixtureState(t *testing.T) (*sqlite.DB, *app.DynamicsEngine, *time.Time, liveRecorder, liveSleeper) {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	clock := v1Window().AddDate(0, 0, 29)
	clockFn := func() time.Time { return clock }
	options := events.ValidationDynamicsOptions(app.DynamicsOptions{Clock: clockFn})
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), options)
	for _, topic := range events.SeedTopics() {
		if _, err = engine.SeedTopic(ctx, app.Principal{ID: "fixture-human", Kind: "human"}, topic, clock.UnixMilli()); err != nil {
			t.Fatal(err)
		}
	}
	repo := sqlite.NewEventRepository(db)
	recorder := &recordruntime.EventRecorder{Events: app.NewEventService(repo, events.ValidationSignalStrength), Repository: repo, Engine: engine}
	sleep := sleepruntime.NewSleepService(db, options, app.SleepOptions{Clock: clockFn})
	clock = clock.AddDate(0, 0, 1).Add(-time.Millisecond)
	if _, err = sleep.Run(ctx, v1Window().AddDate(0, 0, 29)); err != nil {
		t.Fatal(err)
	}
	return db, engine, &clock, recorder, sleep
}

func liveDay() (liveConfig, liveCoverage) {
	start := v1Window().AddDate(0, 0, 30)
	return liveConfig{Start: start, End: start.AddDate(0, 0, 1)}, liveCoverage{Day: start.Format(time.DateOnly), WindowStart: start.UnixMilli(), WindowEnd: start.AddDate(0, 0, 1).UnixMilli(), SleepStatus: "not_run", Status: "failed", ReceiptIDs: []string{}}
}

type observingLiveRecorder struct {
	liveRecorder
	order []int64
}

func (r *observingLiveRecorder) Record(ctx context.Context, rows []app.StoredEvent) ([]app.StoredEvent, error) {
	if len(rows) != 1 {
		return nil, errors.New("batch forbidden")
	}
	r.order = append(r.order, rows[0].Observation.TS)
	return r.liveRecorder.Record(ctx, rows)
}

func TestLiveAllSourceSortDuplicateAndEmptyDailySleep(t *testing.T) {
	db, engine, clock, recorder, sleep := liveFixtureState(t)
	c, coverage := liveDay()
	start := c.Start.UnixMilli()
	first := replayFixture("first", "Morphogenesis", start+1000).Row
	second := replayFixture("second", "McPet", start+2000).Row
	third := replayFixture("unassigned", "", start+3000).Row
	pending := replayFixture("structure", "", start+4000).Row
	pending.Event = nil
	rows := []app.StoredEvent{third, second, pending, first, first}
	observing := &observingLiveRecorder{liveRecorder: recorder}
	var safe bytes.Buffer
	result, err := executeLive(context.Background(), db, rows, c, coverage, clock, observing, sleep, json.NewEncoder(&safe))
	if err != nil || result.Recorded != 4 || result.CognitiveSuccess != 3 || result.Unassigned != 1 || result.CaptureDuplicates != 1 || !result.CaptureComplete || result.SleepStatus != "completed" {
		t.Fatalf("real live pipeline mismatch: %+v %v", result, err)
	}
	for i, ts := range observing.order {
		if ts != start+int64(i+1)*1000 {
			t.Fatal("all-source chronological ordering failed")
		}
	}
	if strings.Contains(safe.String(), "PRIVATE_RAW_SENTINEL") || strings.Contains(safe.String(), "origin_key") {
		t.Fatal("raw escaped receipts")
	}
	before, err := engine.StoredNode(context.Background(), "Morphogenesis")
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := executeLive(context.Background(), db, rows, c, coverage, clock, observing, sleep, json.NewEncoder(&safe))
	after, _ := engine.StoredNode(context.Background(), "Morphogenesis")
	if err != nil || repeated.Recorded != 0 || repeated.PersistedDuplicates != 4 || repeated.SleepStatus != "already_completed" || after.Node.Attention != before.Node.Attention || after.Node.Momentum != before.Node.Momentum {
		t.Fatal("repeat caused contribution or history decay")
	}
	c.Start = c.End
	c.End = c.Start.AddDate(0, 0, 1)
	coverage.Day = c.Start.Format(time.DateOnly)
	coverage.WindowStart = c.Start.UnixMilli()
	coverage.WindowEnd = c.End.UnixMilli()
	empty, err := executeLive(context.Background(), db, nil, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe))
	if err != nil || !empty.CaptureComplete || empty.Recorded != 0 || empty.SleepReceiptID != "sleep:2026-10-13" {
		t.Fatalf("empty day did not run production Sleep: %+v %v", empty, err)
	}
}

type unknownLiveRecorder struct{ liveRecorder }

func (r unknownLiveRecorder) Record(ctx context.Context, rows []app.StoredEvent) ([]app.StoredEvent, error) {
	actual, err := r.liveRecorder.Record(ctx, rows)
	if err != nil {
		return actual, err
	}
	return actual, &app.EventCapturePendingError{ReceiptIDs: []string{actual[0].ID}, Cause: errors.New("PRIVATE_RAW_SENTINEL cause")}
}

type unknownLiveSleeper struct{ liveSleeper }

func (s unknownLiveSleeper) Run(ctx context.Context, day time.Time) (app.SleepResult, error) {
	result, err := s.liveSleeper.Run(ctx, day)
	if err != nil {
		return result, err
	}
	return app.SleepResult{}, errors.New("PRIVATE_RAW_SENTINEL committed Sleep transport error")
}

func TestLiveUnknownCommittedRecorderAndSleepKeepIDs(t *testing.T) {
	for _, phase := range []string{"record", "sleep"} {
		t.Run(phase, func(t *testing.T) {
			db, _, clock, recorder, sleep := liveFixtureState(t)
			c, coverage := liveDay()
			var safe bytes.Buffer
			if phase == "record" {
				recorder = unknownLiveRecorder{recorder}
			} else {
				sleep = unknownLiveSleeper{sleep}
			}
			result, err := executeLive(context.Background(), db, []app.StoredEvent{replayFixture("known-commit", "Morphogenesis", c.Start.UnixMilli()+1000).Row}, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe))
			if err == nil || result.CaptureComplete || result.Status != "delivery_unknown" || strings.Contains(safe.String(), "PRIVATE_RAW_SENTINEL") {
				t.Fatal("unknown was promoted or raw escaped")
			}
			if phase == "record" && len(result.ReceiptIDs) != 1 {
				t.Fatal("committed event ID lost")
			}
			if phase == "sleep" && result.SleepReceiptID != "sleep:2026-10-12" {
				t.Fatal("committed Sleep ID lost")
			}
		})
	}
}

func TestLiveLateAndConflictingSourceRefuseBeforeMutation(t *testing.T) {
	for _, kind := range []string{"late", "conflict", "gap"} {
		t.Run(kind, func(t *testing.T) {
			db, _, clock, recorder, sleep := liveFixtureState(t)
			c, coverage := liveDay()
			row := replayFixture("x", "Morphogenesis", c.Start.UnixMilli()+1000).Row
			rows := []app.StoredEvent{row}
			if kind == "late" {
				rows[0].Event.TS = c.Start.UnixMilli() - 1
				rows[0].Observation.TS = rows[0].Event.TS
			}
			if kind == "conflict" {
				duplicate := row
				duplicate.RawBytes = []byte(`{"changed":true}`)
				rows = append(rows, duplicate)
			}
			if kind == "gap" {
				c.Start = c.End
				c.End = c.Start.AddDate(0, 0, 1)
				coverage.Day = c.Start.Format(time.DateOnly)
			}
			var safe bytes.Buffer
			result, err := executeLive(context.Background(), db, rows, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe))
			var n int
			_ = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_events").Scan(&n)
			if err == nil || result.CaptureComplete || n != 0 {
				t.Fatal("late/conflict/gap silently changed model")
			}
		})
	}
}

func TestLiveNativeHeaderWindowsLayoutsAndCompleteChatSelection(t *testing.T) {
	base := t.TempDir()
	ctx := context.Background()
	start := v1Window().AddDate(0, 0, 30)
	for _, source := range []string{"codex", "claude_code"} {
		root := filepath.Join(base, source)
		directory := filepath.Join(root, "subagents")
		if source == "codex" {
			directory = filepath.Join(root, "2026/10/12")
		}
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		var lines []string
		if source == "codex" {
			lines = []string{`{"timestamp":"2026-10-11T00:00:00+08:00","type":"session_meta","payload":{"id":"fixture","cwd":"` + filepath.ToSlash(base) + `"}}`, `{"timestamp":"2026-10-12T00:00:01+08:00","type":"response_item","payload":{"type":"function_call","call_id":"one","name":"apply_patch","arguments":"synthetic edit"}}`, `{"timestamp":"2026-10-13T00:00:00+08:00","type":"response_item","payload":{"type":"function_call","call_id":"end-exclusive","name":"apply_patch","arguments":"synthetic edit"}}`}
		} else {
			lines = []string{`{"timestamp":"2026-10-11T00:00:00+08:00","type":"user","sessionId":"fixture","cwd":"` + filepath.ToSlash(base) + `","message":{"content":"synthetic header"}}`, `{"timestamp":"2026-10-12T00:00:01+08:00","type":"assistant","message":{"content":[{"type":"tool_use","id":"one","name":"Edit","input":{"file_path":"synthetic"}}]}}`}
		}
		if err := os.WriteFile(filepath.Join(directory, "tool.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "pure-chat.jsonl"), []byte(`{"timestamp":"2026-10-12T00:00:00+08:00","type":"user","message":{"content":"synthetic pure chat"}}`+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		var raw bytes.Buffer
		c := liveCapture{Clients: map[string]map[string]int{}, seen: map[string]app.StoredEvent{}, mapper: events.NewNativeSemanticMapper(), raw: json.NewEncoder(&raw)}
		c.captureNative(ctx, source, root, start, start.AddDate(0, 0, 1))
		if c.Errors != 0 || len(c.Rows) != 1 || c.Rows[0].Event == nil || c.Rows[0].Event.Kind != "edit_file" || c.Rows[0].Observation.Session != "fixture" || !equalValidationPath(c.Rows[0].Observation.Scope, base) || len(c.selection) != 1 || len(c.selection[0]["files"].([]string)) != 2 {
			t.Fatalf("header/window/full selection mismatch %s: errors=%d rows=%d", source, c.Errors, len(c.Rows))
		}
	}
}

func TestLiveAuthorizationAndClosedWindowRejectInvalidArgs(t *testing.T) {
	home, _ := os.UserHomeDir()
	c, _ := liveDay()
	c.Phase = "prelock"
	c.State = filepath.Join(v4Root, "prelock-fixture/state.sqlite")
	c.Out = filepath.Join(filepath.Dir(c.State), "capture-new")
	c.Native = map[string]string{"claude_code": filepath.Join(home, ".claude/projects"), "codex": filepath.Join(home, ".codex/sessions")}
	c.Git = map[string]string{"Morphogenesis": "C:/Users/DW/orca/Morphogenesis", "Astrocyte/Attention Engine": "C:/Users/DW/orca/Astrocyte", "McPet": "C:/Users/DW/orca/mcpet"}
	// Root existence is environmental; invalid constraints must fail regardless.
	for _, change := range []func(*liveConfig){func(c *liveConfig) { c.Start = c.Start.Add(time.Second) }, func(c *liveConfig) { c.End = c.Start.Add(12 * time.Hour) }, func(c *liveConfig) { c.Out = "C:/personal" }, func(c *liveConfig) { c.State = filepath.Join(v4Root, "state.sqlite") }, func(c *liveConfig) { c.Native = map[string]string{"claude_code": "C:/private"} }} {
		copy := c
		change(&copy)
		if validateLiveConfig(copy, c.End) == nil {
			t.Fatal("invalid root/window accepted")
		}
	}
	if validateLiveConfig(c, c.Start) == nil {
		t.Fatal("future day accepted")
	}
}

func TestLiveGitFixedHEADCommitterWindowAndActualRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mcpet")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	start := v1Window().AddDate(0, 0, 30)
	run := func(timestamp time.Time, args ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", root}, args...)...)
		command.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+timestamp.Format(time.RFC3339), "GIT_COMMITTER_DATE="+timestamp.Format(time.RFC3339))
		if b, err := command.CombinedOutput(); err != nil {
			t.Fatalf("synthetic git failed: %v %s", err, b)
		}
	}
	run(start, "init", "-q")
	for i, ts := range []time.Time{start, start.AddDate(0, 0, 1).Add(-time.Second), start.AddDate(0, 0, 1)} {
		run(ts, "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "commit", "--allow-empty", "-m", fmt.Sprintf("synthetic%d", i))
	}
	var raw bytes.Buffer
	c := liveCapture{Git: map[string]int{}, Heads: map[string]string{}, seen: map[string]app.StoredEvent{}, mapper: events.NewNativeSemanticMapper(), raw: json.NewEncoder(&raw)}
	c.captureGit(context.Background(), "McPet", root, start, start.AddDate(0, 0, 1))
	if c.Errors != 0 || c.Git["McPet"] != 2 || len(c.Rows) != 2 || len(c.Heads["McPet"]) != 40 {
		t.Fatalf("fixed HEAD/committer window failed: errors=%d count=%d", c.Errors, len(c.Rows))
	}
	for _, row := range c.Rows {
		if row.Event.Kind != "commit" || row.Event.Topic != "McPet" || row.Event.TS >= start.AddDate(0, 0, 1).UnixMilli() {
			t.Fatal("wrong Git fact identity")
		}
	}
}

func TestLiveIncompleteNativeSourceExplicitlyFailsCapture(t *testing.T) {
	root := t.TempDir()
	start := v1Window().AddDate(0, 0, 30)
	if err := os.WriteFile(filepath.Join(root, "broken.jsonl"), []byte("{broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var raw bytes.Buffer
	c := liveCapture{Clients: map[string]map[string]int{}, seen: map[string]app.StoredEvent{}, mapper: events.NewNativeSemanticMapper(), raw: json.NewEncoder(&raw)}
	c.captureNative(context.Background(), "codex", root, start, start.AddDate(0, 0, 1))
	if c.Errors != 1 || c.Clients["codex"]["parse_errors"] != 1 || len(c.Rows) != 0 {
		t.Fatal("incomplete raw JSON silently omitted")
	}
}

func TestLiveNativeRecordAboveFormerScannerLimit(t *testing.T) {
	root := t.TempDir()
	start := v1Window().AddDate(0, 0, 30)
	raw := `{"timestamp":"2026-10-11T00:00:00+08:00","type":"user","message":{"content":"` + strings.Repeat("x", 64*1024*1024) + `"}}` + "\n"
	if err := os.WriteFile(filepath.Join(root, "large.jsonl"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	var local bytes.Buffer
	c := liveCapture{Clients: map[string]map[string]int{}, seen: map[string]app.StoredEvent{}, mapper: events.NewNativeSemanticMapper(), raw: json.NewEncoder(&local)}
	c.captureNative(context.Background(), "codex", root, start, start.AddDate(0, 0, 1))
	if c.Errors != 0 || c.Clients["codex"]["scanned_lines"] != 1 {
		t.Fatal("large native JSONL record truncated")
	}
}

func TestLivePublicUnknownDepthDoesNotBecomeSuccessfulSignal(t *testing.T) {
	db, _, clock, recorder, sleep := liveFixtureState(t)
	c, coverage := liveDay()
	row := replayFixture("unknown-depth", "Morphogenesis", c.Start.UnixMilli()+1000).Row
	row.Event.Kind = "read_paper"
	var safe bytes.Buffer
	result, err := executeLive(context.Background(), db, []app.StoredEvent{row}, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe))
	var signals int
	_ = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_engine_signals").Scan(&signals)
	if err == nil || result.CaptureComplete || result.NormalizationFailed != 1 || result.Recorded != 0 || result.CognitiveSuccess != 0 || signals != 0 || !strings.Contains(safe.String(), "normalization_failed") {
		t.Fatal("unknown depth silently gained signal or success")
	}
	if got := liveTimestamp(json.RawMessage(`{"timestamp":null,"type":"session_meta","payload":{"timestamp":"2026-10-12T00:00:00+08:00"}}`)); got != c.Start.UnixMilli() {
		t.Fatal("public header timestamp fallback lost")
	}
}

func TestLiveFormalPhaseRequiresDeclaredLockAndPrelockCannotClaimIt(t *testing.T) {
	c, _ := liveDay()
	for _, phase := range []string{"prelock", "bridge"} {
		c.Phase = phase
		if validateLivePhase(c) != nil {
			t.Fatal("nonformal capture rejected")
		}
	}
	c.Phase = "formal"
	if validateLivePhase(c) == nil {
		t.Fatal("formal accepted without declared lock")
	}
	c.ProtocolSHA = strings.Repeat("a", 64)
	c.Day0 = c.Start.Format(time.DateOnly)
	if validateLivePhase(c) != nil {
		t.Fatal("root-declared formal fields rejected")
	}
	c.Phase = "prelock"
	if validateLivePhase(c) == nil {
		t.Fatal("prelock claimed experiment lock")
	}
	c.Phase = "formal"
	c.Start = c.Start.AddDate(0, 0, 9)
	if validateLivePhase(c) != nil {
		t.Fatal("last D+3 prediction maturity collection rejected")
	}
	c.Start = c.Start.AddDate(0, 0, 1)
	if validateLivePhase(c) == nil {
		t.Fatal("formal accepted out-of-experiment day")
	}
}

func TestLiveSnapshotFailureRetainsCommittedIDsAndNoCompletedCoverage(t *testing.T) {
	db, _, clock, recorder, sleep := liveFixtureState(t)
	c, coverage := liveDay()
	c.Out = t.TempDir()
	var safe bytes.Buffer
	actual, err := executeLive(context.Background(), db, []app.StoredEvent{replayFixture("snapshot-commit", "McPet", c.Start.UnixMilli()+1000).Row}, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe))
	if err != nil {
		t.Fatal(err)
	}
	failed, err := finalizeLiveSnapshot(context.Background(), db, c, actual, func() error { return errors.New("PRIVATE_RAW_SENTINEL backup error") })
	if err == nil || failed.CaptureComplete || failed.Status != "delivery_unknown" || len(failed.ReceiptIDs) != 1 || failed.SleepReceiptID == "" {
		t.Fatal("committed IDs lost on snapshot failure")
	}
	var receipt map[string]any
	readReplayJSON(t, filepath.Join(c.Out, "failure-safe.json"), &receipt)
	// Advancing a real next day makes the old mutable model invalid for a new
	// old-day snapshot; the previous immutable snapshot remains the consumer.
	next := c
	next.Start = c.End
	next.End = next.Start.AddDate(0, 0, 1)
	nextCoverage := coverage
	nextCoverage.Day = next.Start.Format(time.DateOnly)
	if _, err = executeLive(context.Background(), db, nil, next, nextCoverage, clock, recorder, sleep, json.NewEncoder(&safe)); err != nil {
		t.Fatal(err)
	}
	if result, err := executeLive(context.Background(), db, nil, c, coverage, clock, recorder, sleep, json.NewEncoder(&safe)); err == nil || result.CaptureComplete || result.Late != 1 {
		t.Fatal("D+1 model served as old-day state")
	}
}
