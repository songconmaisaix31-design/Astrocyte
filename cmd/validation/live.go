package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	sleepruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/attentionruntime"
	recordruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/internal/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

const v4Root = v1EvidenceRoot + "/v4-w1"

var liveZone = time.FixedZone("Asia/Shanghai", 8*3600)

type liveConfig struct {
	State, Out  string
	Native, Git map[string]string
	Start, End  time.Time
	Init        bool
}

type liveCoverage struct {
	Day                 string                    `json:"day"`
	WindowStart         int64                     `json:"window_start"`
	WindowEnd           int64                     `json:"window_end"`
	CaptureComplete     bool                      `json:"capture_complete"`
	SelectionComplete   bool                      `json:"selection_complete"`
	ClientCounts        map[string]map[string]int `json:"client_counts"`
	GitCounts           map[string]int            `json:"git_counts"`
	CapturedHeads       map[string]string         `json:"captured_git_heads"`
	CognitiveSuccess    int                       `json:"cognitive_success"`
	Unassigned          int                       `json:"unassigned"`
	StructureFailed     int                       `json:"structure_failed"`
	NormalizationFailed int                       `json:"normalization_failed"`
	Noncognitive        int                       `json:"noncognitive"`
	CaptureDuplicates   int                       `json:"capture_duplicates"`
	PersistedDuplicates int                       `json:"persisted_duplicates"`
	Late                int                       `json:"late"`
	SourceErrors        int                       `json:"source_errors"`
	Recorded            int                       `json:"recorded"`
	SleepStatus         string                    `json:"sleep_status"`
	SleepReceiptID      string                    `json:"sleep_receipt_id"`
	Status              string                    `json:"status"`
	ReceiptIDs          []string                  `json:"receipt_ids"`
	SemanticPolicy      string                    `json:"semantic_policy"`
	Experiment          string                    `json:"experiment"`
}

func runLiveCLI(arguments []string) error {
	f := flag.NewFlagSet("live", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	state := f.String("live-state", "", "existing isolated live state.sqlite")
	out := f.String("live-out", "", "exclusive new capture child of live state directory")
	start := f.String("live-start", "", "inclusive closed-day +08 RFC3339")
	end := f.String("live-end", "", "exclusive closed-day +08 RFC3339")
	claude := f.String("live-claude-root", "", "authorized native root")
	codex := f.String("live-codex-root", "", "authorized native root")
	morph := f.String("live-morph-root", "", "authorized Git root")
	astro := f.String("live-astro-root", "", "authorized Git root")
	mcpet := f.String("live-mcpet-root", "", "authorized Git root")
	init := f.Bool("live-init", false, "one-time fixed V1R read-only SQLite backup")
	if f.Parse(arguments) != nil || f.NArg() != 0 {
		return fmt.Errorf("invalid live arguments")
	}
	a, e := time.Parse(time.RFC3339, *start)
	b, g := time.Parse(time.RFC3339, *end)
	if e != nil || g != nil {
		return fmt.Errorf("explicit live window required")
	}
	config := liveConfig{State: *state, Out: *out, Start: a, End: b, Init: *init, Native: map[string]string{"claude_code": *claude, "codex": *codex}, Git: map[string]string{"Morphogenesis": *morph, "Astrocyte/Attention Engine": *astro, "McPet": *mcpet}}
	if err := validateLiveConfig(config, time.Now()); err != nil {
		return err
	}
	return runLive(context.Background(), config)
}

func validateLiveConfig(c liveConfig, now time.Time) error {
	_, a := c.Start.Zone()
	_, b := c.End.Zone()
	if a != 8*3600 || b != 8*3600 || c.Start.Hour() != 0 || c.Start.Minute() != 0 || c.Start.Second() != 0 || c.Start.Nanosecond() != 0 || !c.End.Equal(c.Start.AddDate(0, 0, 1)) || c.End.After(now) || c.Start.Before(v1Window().AddDate(0, 0, 30)) {
		return fmt.Errorf("require one closed +08 day after frozen history")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	for name, want := range map[string]string{"claude_code": filepath.Join(home, ".claude/projects"), "codex": filepath.Join(home, ".codex/sessions")} {
		if !equalValidationPath(c.Native[name], want) {
			return fmt.Errorf("native root outside authorization")
		}
	}
	for topic, want := range map[string]string{"Morphogenesis": "C:/Users/DW/orca/Morphogenesis", "Astrocyte/Attention Engine": "C:/Users/DW/orca/Astrocyte", "McPet": "C:/Users/DW/orca/mcpet"} {
		if !equalValidationPath(c.Git[topic], want) {
			return fmt.Errorf("Git root outside authorization")
		}
	}
	parent := filepath.Dir(c.State)
	if filepath.Base(c.State) != "state.sqlite" || !equalValidationPath(filepath.Dir(parent), v4Root) || !equalValidationPath(filepath.Dir(c.Out), parent) || equalValidationPath(c.Out, c.State) {
		return fmt.Errorf("live output outside isolated state root")
	}
	for _, path := range []string{parent, c.Out} {
		// Component validation is repeated after reservation, before reading raw.
		if _, err := os.Stat(path); err == nil && events.ValidateLocalReplayPath(path) != nil {
			return fmt.Errorf("redirected live output")
		}
	}
	if events.ValidateLocalReplayPath(v4Root) != nil {
		return fmt.Errorf("live evidence root unavailable")
	}
	return nil
}

func runLive(ctx context.Context, c liveConfig) (resultErr error) {
	if c.Init {
		command := exec.CommandContext(ctx, "python", "scripts/validation/bootstrap_v4.py", "--out", c.State)
		if command.Run() != nil {
			return fmt.Errorf("fixed read-only bootstrap failed")
		}
	}
	if events.ValidateLocalReplayPath(c.State) != nil || events.ValidateLocalReplayPath(filepath.Dir(c.Out)) != nil {
		return fmt.Errorf("isolated live state unavailable")
	}
	if err := os.Mkdir(c.Out, 0700); err != nil {
		return fmt.Errorf("live capture output must be new")
	}
	if events.ValidateLocalReplayPath(c.Out) != nil {
		return fmt.Errorf("redirected capture output")
	}
	// This ordinary exclusive writer lock also survives a crash. Its presence
	// means the previous run needs reconciliation; it is never removed on error.
	lockPath := c.State + ".live-lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("prior live operation needs reconciliation")
	}
	if err = json.NewEncoder(lock).Encode(map[string]any{"status": "in_progress", "out": c.Out}); err != nil {
		lock.Close()
		return err
	}
	if err = lock.Sync(); err != nil {
		lock.Close()
		return err
	}
	lock.Close()
	release := false
	defer func() {
		if release {
			_ = os.Remove(lockPath)
		}
	}()
	coverage := liveCoverage{Day: c.Start.Format(time.DateOnly), WindowStart: c.Start.UnixMilli(), WindowEnd: c.End.UnixMilli(), SleepStatus: "not_run", Status: "failed", ReceiptIDs: []string{}, SemanticPolicy: events.SemanticPolicyVersion, Experiment: "live_entry_only; lock/day0 owned by root"}
	coverageFile, err := os.OpenFile(filepath.Join(c.Out, "coverage-safe.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer coverageFile.Close()
	defer func() {
		if err := json.NewEncoder(coverageFile).Encode(coverage); err != nil {
			release = false
			resultErr = fmt.Errorf("safe coverage delivery unknown")
		}
		if err := coverageFile.Sync(); err != nil {
			release = false
			resultErr = fmt.Errorf("safe coverage delivery unknown")
		}
	}()
	capture, err := captureLive(ctx, c, c.Out)
	coverage.ClientCounts, coverage.GitCounts, coverage.CapturedHeads = capture.Clients, capture.Git, capture.Heads
	coverage.StructureFailed, coverage.Noncognitive, coverage.CaptureDuplicates, coverage.SourceErrors = capture.StructureFailed, capture.Noncognitive, capture.Duplicates, capture.Errors
	coverage.NormalizationFailed = capture.NormalizationFailed
	coverage.SelectionComplete = capture.Errors == 0
	if err != nil || capture.Errors != 0 {
		release = true
		return fmt.Errorf("incomplete authorized source capture")
	}
	db, err := sqlite.Open(c.State)
	if err != nil {
		return err
	}
	defer db.Close()
	clock := c.Start
	clockFn := func() time.Time { return clock }
	options := events.ValidationDynamicsOptions(app.DynamicsOptions{Clock: clockFn})
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), options)
	repo := sqlite.NewEventRepository(db)
	recorder := &recordruntime.EventRecorder{Events: app.NewEventService(repo, events.ValidationSignalStrength), Repository: repo, Engine: engine}
	sleep := sleepruntime.NewSleepService(db, options, app.SleepOptions{Clock: clockFn})
	receiptFile, err := os.OpenFile(filepath.Join(c.Out, "event-receipts-safe.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer receiptFile.Close()
	coverage, err = executeLive(ctx, db, capture.Rows, c, coverage, &clock, recorder, sleep, json.NewEncoder(receiptFile))
	if err != nil {
		_ = writeExclusive(filepath.Join(c.Out, "failure-safe.json"), map[string]any{"status": coverage.Status, "receipt_ids": coverage.ReceiptIDs, "sleep_receipt_id": coverage.SleepReceiptID, "sleep_status": coverage.SleepStatus, "late": coverage.Late, "source_errors": coverage.SourceErrors, "retry": "reconcile LOCAL receipts; do not blindly replay"})
		return err
	}
	if err = receiptFile.Sync(); err != nil {
		coverage.Status = "delivery_unknown"
		coverage.CaptureComplete = false
		return err
	}
	release = true
	return nil
}

type liveRecorder interface {
	Record(context.Context, []app.StoredEvent) ([]app.StoredEvent, error)
}
type liveSleeper interface {
	Run(context.Context, time.Time) (app.SleepResult, error)
}

func pendingReceiptIDs(rows []app.StoredEvent, err error) []string {
	ids := []string{}
	seen := map[string]bool{}
	for _, row := range rows {
		if row.ID != "" && !seen[row.ID] {
			ids = append(ids, row.ID)
			seen[row.ID] = true
		}
	}
	var pending *app.EventCapturePendingError
	if errors.As(err, &pending) {
		for _, id := range pending.ReceiptIDs {
			if !seen[id] {
				ids = append(ids, id)
				seen[id] = true
			}
		}
	}
	return ids
}

func executeLive(ctx context.Context, db *sqlite.DB, rows []app.StoredEvent, c liveConfig, coverage liveCoverage, clock *time.Time, recorder liveRecorder, sleep liveSleeper, receipts *json.Encoder) (liveCoverage, error) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Observation.TS != rows[j].Observation.TS {
			return rows[i].Observation.TS < rows[j].Observation.TS
		}
		return rows[i].OriginKey < rows[j].OriginKey
	})
	// All reads and late checks precede the first production mutation. The
	// completed Sleep boundary governs continuity, never a wall-clock estimate.
	var lastDay sql.NullString
	if err := db.Conn().QueryRowContext(ctx, "SELECT MAX(day) FROM attention_engine_sleep_runs").Scan(&lastDay); err != nil {
		return coverage, err
	}
	if !lastDay.Valid {
		return coverage, fmt.Errorf("initialized model requires completed historical Sleep")
	}
	day, err := time.ParseInLocation(time.DateOnly, lastDay.String, liveZone)
	if err != nil {
		return coverage, err
	}
	var existingSleepID string
	err = db.Conn().QueryRowContext(ctx, "SELECT id FROM attention_engine_sleep_runs WHERE day=?", coverage.Day).Scan(&existingSleepID)
	if err == nil {
		coverage.SleepReceiptID = existingSleepID
		coverage.SleepStatus = "already_completed"
	} else if !errors.Is(err, sql.ErrNoRows) {
		return coverage, err
	}
	if coverage.SleepStatus != "already_completed" && !day.AddDate(0, 0, 1).Equal(c.Start) {
		return coverage, fmt.Errorf("capture missing intermediate closed days")
	}
	var latest int64
	if err = db.Conn().QueryRowContext(ctx, "SELECT COALESCE(MAX(updated_at),0) FROM attention_engine_nodes").Scan(&latest); err != nil {
		return coverage, err
	}
	var selected []app.StoredEvent
	seen := map[string]app.StoredEvent{}
	for _, row := range rows {
		if previous, exists := seen[row.OriginKey]; exists {
			a, _ := json.Marshal(previous.Observation)
			b, _ := json.Marshal(row.Observation)
			if !bytes.Equal(previous.RawBytes, row.RawBytes) || !bytes.Equal(a, b) {
				coverage.SourceErrors++
			} else {
				coverage.CaptureDuplicates++
			}
			continue
		}
		seen[row.OriginKey] = row
		var id string
		var raw []byte
		var observation string
		err = db.Conn().QueryRowContext(ctx, "SELECT id,raw_bytes,observation FROM attention_engine_events WHERE origin_key=?", row.OriginKey).Scan(&id, &raw, &observation)
		if err == nil {
			observed, _ := json.Marshal(row.Observation)
			var old any
			var current any
			_ = json.Unmarshal([]byte(observation), &old)
			_ = json.Unmarshal(observed, &current)
			a, _ := json.Marshal(old)
			b, _ := json.Marshal(current)
			if !bytes.Equal(raw, row.RawBytes) || !bytes.Equal(a, b) {
				coverage.SourceErrors++
				continue
			}
			coverage.PersistedDuplicates++
			if receipts.Encode(map[string]any{"event_id": id, "ts": row.Observation.TS, "status": "duplicate_no_contribution"}) != nil {
				return coverage, fmt.Errorf("safe receipt write failed")
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return coverage, err
		}
		if row.Observation.TS < c.Start.UnixMilli() || row.Observation.TS >= c.End.UnixMilli() || row.Observation.TS < latest || coverage.SleepStatus == "already_completed" {
			coverage.Late++
			continue
		}
		selected = append(selected, row)
	}
	if coverage.SourceErrors > 0 || coverage.Late > 0 {
		return coverage, fmt.Errorf("late or conflicting event requires explicit reconciliation")
	}
	for _, row := range selected {
		*clock = time.UnixMilli(row.Observation.TS).In(liveZone)
		row.ReceivedAt = row.Observation.TS
		stored, err := recorder.Record(ctx, []app.StoredEvent{row})
		ids := pendingReceiptIDs(stored, err)
		if err != nil || len(stored) != 1 {
			// Re-read the immutable origin when an error obscures a committed ID.
			if len(ids) == 0 {
				var id string
				if db.Conn().QueryRowContext(ctx, "SELECT id FROM attention_engine_events WHERE origin_key=?", row.OriginKey).Scan(&id) == nil {
					ids = append(ids, id)
				}
			}
			coverage.ReceiptIDs = append(coverage.ReceiptIDs, ids...)
			coverage.Status = "delivery_unknown"
			_ = receipts.Encode(map[string]any{"ts": row.Observation.TS, "status": "delivery_unknown", "receipt_ids": ids})
			return coverage, fmt.Errorf("single production Record needs reconciliation")
		}
		actual := stored[0]
		coverage.Recorded++
		if actual.Event != nil {
			signal, err := events.NormalizeValidationSignal(actual)
			if err != nil {
				coverage.Status = "delivery_unknown"
				coverage.ReceiptIDs = append(coverage.ReceiptIDs, actual.ID)
				return coverage, fmt.Errorf("committed signal binding requires reconciliation")
			}
			if signal != nil {
				coverage.CognitiveSuccess++
			}
			if actual.Event.Topic == "" {
				coverage.Unassigned++
			}
		}
		ref := map[string]any{"event_id": actual.ID, "ts": actual.Observation.TS, "status": "recorded"}
		if actual.Event != nil {
			ref["kind"], ref["topic"] = actual.Event.Kind, actual.Event.Topic
			if signal, _ := events.NormalizeValidationSignal(actual); signal == nil {
				ref["status"] = "normalization_pending"
			}
		} else {
			ref["status"] = "structure_pending"
		}
		if receipts.Encode(ref) != nil {
			coverage.Status = "delivery_unknown"
			coverage.ReceiptIDs = append(coverage.ReceiptIDs, actual.ID)
			return coverage, fmt.Errorf("committed receipt export failed")
		}
	}
	if coverage.SleepStatus != "already_completed" {
		*clock = c.End.Add(-time.Millisecond)
		result, err := sleep.Run(ctx, c.Start)
		coverage.SleepReceiptID = result.Run.ID
		if err != nil || result.Replayed || result.Run.ID == "" {
			if coverage.SleepReceiptID == "" {
				_ = db.Conn().QueryRowContext(ctx, "SELECT id FROM attention_engine_sleep_runs WHERE day=?", coverage.Day).Scan(&coverage.SleepReceiptID)
			}
			coverage.SleepStatus = "delivery_unknown"
			coverage.Status = "delivery_unknown"
			return coverage, fmt.Errorf("production Sleep needs reconciliation")
		}
		coverage.SleepStatus = "completed"
	}
	coverage.Status = "completed"
	coverage.CaptureComplete = true
	return coverage, nil
}
