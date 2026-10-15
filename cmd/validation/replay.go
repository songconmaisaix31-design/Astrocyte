package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	sleepruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/attentionruntime"
	recordruntime "github.com/songconmaisaix31-design/Astrocyte/cmd/internal/attentionruntime"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

const v1EvidenceRoot = "C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012"
const v1WeightSampleSeed = int64(2026101201)

// IDs in the frozen V0 audit identify source semantic units. The store assigns
// separate immutable receipt IDs. A sorted source ordinal gives a safe, stable
// replay identity without exposing source origins/session IDs or changing M0.
type replayInput struct {
	StableID string
	Row      app.StoredEvent
}

// This is the existing six-field exposure projection, never StoredEvent JSON.
type replayEventRef struct {
	EventID string `json:"event_id"`
	Kind    string `json:"kind"`
	Topic   string `json:"topic"`
	Tokens  int    `json:"tokens"`
	Scope   string `json:"scope"`
	TS      int64  `json:"ts"`
}

type replayPoint struct {
	Topic      string  `json:"topic"`
	TS         int64   `json:"ts"`
	UpdatedAt  int64   `json:"updated_at"`
	Attention  float64 `json:"attention"`
	Momentum   float64 `json:"momentum"`
	Confidence float64 `json:"confidence"`
	Lifecycle  string  `json:"lifecycle"`
	Phase      string  `json:"phase"`
	ReplayID   string  `json:"replay_id,omitempty"`
}

type replayOperation struct {
	Phase       string          `json:"phase"`
	TS          int64           `json:"ts"`
	ReplayID    string          `json:"replay_id,omitempty"`
	Event       *replayEventRef `json:"event,omitempty"`
	Strength    float64         `json:"strength,omitempty"`
	Weight      float64         `json:"weight,omitempty"`
	Points      []replayPoint   `json:"points"`
	Day         string          `json:"day,omitempty"`
	SignalIDs   []string        `json:"signal_ids,omitempty"`
	InputEvents int             `json:"input_events,omitempty"`
}

type replaySleep struct {
	Day          string   `json:"day"`
	WindowStart  int64    `json:"window_start"`
	WindowEnd    int64    `json:"window_end"`
	StartedAt    int64    `json:"started_at"`
	CompletedAt  int64    `json:"completed_at"`
	InputEvents  int      `json:"input_events"`
	InputSignals int      `json:"input_signals"`
	Unassigned   int      `json:"unassigned"`
	Pending      int      `json:"unassigned_pending"`
	ReceiptID    string   `json:"receipt_id"`
	Replayed     bool     `json:"replayed"`
	SignalIDs    []string `json:"signal_ids"`
	Result       string   `json:"result"`
}

type replayRanking struct {
	Day     string        `json:"day"`
	Top5    []replayPoint `json:"top5"`
	Changes []rankChange  `json:"changes"`
}

type rankChange struct {
	Topic    string `json:"topic"`
	Previous int    `json:"previous_rank"`
	Current  int    `json:"current_rank"`
}

func v1Window() time.Time {
	return time.Date(2026, 9, 12, 0, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*3600))
}

func loadV1Inputs(nativePath, gitPath string) ([]replayInput, map[string]int, error) {
	return loadV1InputsWithMapper(nativePath, gitPath, events.NewNativeSemanticMapper())
}

func loadV1InputsWithMapper(nativePath, gitPath string, mapper *events.NativeSemanticMapper) ([]replayInput, map[string]int, error) {
	counts := map[string]int{}
	var rows []app.StoredEvent
	seen := map[string]bool{}
	add := func(row app.StoredEvent, bucket string) error {
		counts["semantic_units"]++
		if row.Event == nil {
			counts[bucket]++
			return nil
		}
		if seen[row.OriginKey] || row.Event.TS != row.Observation.TS {
			return fmt.Errorf("duplicate or inconsistent frozen identity")
		}
		seen[row.OriginKey] = true
		counts["mapped"]++
		counts["kind:"+row.Event.Kind]++
		if row.Event.Topic == "" {
			counts["unassigned"]++
		}
		rows = append(rows, row)
		return nil
	}
	err := localLines(nativePath, func(line []byte) error {
		var v populationEnvelope
		if json.Unmarshal(line, &v) != nil || v.AssociationBad || len(v.Blocks) == 0 {
			return fmt.Errorf("invalid frozen native envelope")
		}
		raw, err := base64.StdEncoding.DecodeString(v.RawBase64)
		if err != nil {
			return fmt.Errorf("invalid frozen native encoding")
		}
		for _, block := range v.Blocks {
			counts["native_parents"]++
			origin := fmt.Sprintf("native:%s:%d:%d:block:%d", v.Source, v.FileIndex, v.Line, block)
			children, semantic, err := mapper.MapTool(v.Source, origin, v.Project, raw, 0, v.Association, block)
			if err != nil {
				return fmt.Errorf("frozen semantic mapping failed")
			}
			for i, row := range children {
				if err := add(row, semantic[i].Bucket); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, counts, err
	}
	err = localLines(gitPath, func(line []byte) error {
		var v struct {
			Project string          `json:"project"`
			Record  json.RawMessage `json:"record"`
		}
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid frozen Git envelope")
		}
		row, err := mapper.MapGit(v.Project, v.Record, 0)
		if err != nil {
			return fmt.Errorf("frozen Git mapping failed")
		}
		counts["git_parents"]++
		return add(row, "cognitive_commit")
	})
	if err != nil {
		return nil, counts, err
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].OriginKey < rows[j].OriginKey })
	inputs := make([]replayInput, len(rows))
	for i, row := range rows {
		inputs[i] = replayInput{StableID: fmt.Sprintf("v1-semantic-%06d", i+1), Row: row}
	}
	return inputs, counts, nil
}

func runV1(populationDir, gitPath, out string) error {
	equalPath := func(a, b string) bool {
		x, e := filepath.Abs(a)
		y, f := filepath.Abs(b)
		return e == nil && f == nil && filepath.Clean(x) == filepath.Clean(y)
	}
	if !equalPath(populationDir, filepath.Join(v1EvidenceRoot, "v21-w1")) ||
		!equalPath(gitPath, filepath.Join(v1EvidenceRoot, "diagnosis-w1/p0-git/git-log-local.jsonl")) ||
		!equalPath(filepath.Dir(out), filepath.Join(v1EvidenceRoot, "v1-w1")) {
		return fmt.Errorf("require fixed authorized inputs and exclusive v1-w1 output")
	}
	// Reserve evidence before reading any payload. Never overwrite or retry in
	// a used directory; a failed run remains available for first-RED review.
	if err := os.Mkdir(out, 0700); err != nil {
		return fmt.Errorf("output must be exclusively new")
	}
	started := time.Now().UTC()
	inputs, counts, err := loadV1Inputs(filepath.Join(populationDir, "tool-records-local.jsonl"), gitPath)
	if err == nil && (counts["native_parents"] != 4123 || counts["git_parents"] != 1893 || counts["mapped"] != 2832 || counts["unassigned"] != 849 || counts["semantic_units"] != 8914) {
		err = fmt.Errorf("approved V0 population changed")
	}
	if err == nil {
		err = replayV1Records(inputs, out, counts, started)
	}
	if err != nil {
		_ = writeExclusive(filepath.Join(out, "first-red.json"), map[string]any{"result": "RED", "stage": "production_replay", "started_at": started, "ended_at": time.Now().UTC(), "error": err.Error()})
	}
	return err
}

func replayV1Records(inputs []replayInput, out string, counts map[string]int, started time.Time) error {
	ctx := context.Background()
	start := v1Window()
	end := start.AddDate(0, 0, 30)
	seenStable, seenOrigin := map[string]bool{}, map[string]bool{}
	for _, input := range inputs {
		if input.StableID == "" || seenStable[input.StableID] || seenOrigin[input.Row.OriginKey] || input.Row.Event == nil || input.Row.Event.TS < start.UnixMilli() || input.Row.Event.TS >= end.UnixMilli() {
			return fmt.Errorf("invalid replay identity or time window")
		}
		seenStable[input.StableID], seenOrigin[input.Row.OriginKey] = true, true
	}
	sort.Slice(inputs, func(i, j int) bool {
		if inputs[i].Row.Event.TS != inputs[j].Row.Event.TS {
			return inputs[i].Row.Event.TS < inputs[j].Row.Event.TS
		}
		return inputs[i].StableID < inputs[j].StableID
	})
	db, err := sqlite.Open(filepath.Join(out, "state.sqlite"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		return err
	}
	clock := start
	clockFn := func() time.Time { return clock }
	options := events.ValidationDynamicsOptions(app.DynamicsOptions{Clock: clockFn})
	engine := app.NewDynamicsEngine(sqlite.NewDynamicsRepository(db), options)
	repo := sqlite.NewEventRepository(db)
	recorder := &recordruntime.EventRecorder{Events: app.NewEventService(repo, events.ValidationSignalStrength), Repository: repo, Engine: engine}
	sleep := sleepruntime.NewSleepService(db, options, app.SleepOptions{Clock: clockFn})
	topics := events.SeedTopics()
	sort.Strings(topics)
	series := map[string][]replayPoint{}
	point := func(topic, phase, stable string) (replayPoint, error) {
		n, err := engine.StoredNode(ctx, topic)
		return replayPoint{Topic: topic, TS: clock.UnixMilli(), UpdatedAt: n.UpdatedAt, Attention: n.Node.Attention, Momentum: n.Node.Momentum, Confidence: n.Node.Confidence, Lifecycle: n.Lifecycle, Phase: phase, ReplayID: stable}, err
	}
	traceFile, err := os.OpenFile(filepath.Join(out, "operations-safe.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer traceFile.Close()
	trace := json.NewEncoder(traceFile)
	seed := replayOperation{Phase: "seed", TS: clock.UnixMilli(), Points: []replayPoint{}}
	for _, topic := range topics {
		if _, err = engine.SeedTopic(ctx, app.Principal{ID: "validation-seed-human", Kind: "human"}, topic, clock.UnixMilli()); err != nil {
			return err
		}
		p, err := point(topic, "seed", "")
		if err != nil {
			return err
		}
		seed.Points = append(seed.Points, p)
		series[topic] = append(series[topic], p)
	}
	if err = trace.Encode(seed); err != nil {
		return err
	}
	var weights []replayOperation
	var sleeps []replaySleep
	var rankings []replayRanking
	previousRanks := map[string]int{}
	index, equalTSPairs := 0, 0
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		dayEnd := day.AddDate(0, 0, 1).Add(-time.Millisecond)
		dayInputs := 0
		for index < len(inputs) && inputs[index].Row.Event.TS <= dayEnd.UnixMilli() {
			input := inputs[index]
			clock = time.UnixMilli(input.Row.Event.TS).In(start.Location())
			input.Row.ReceivedAt = clock.UnixMilli()
			rows, err := recorder.Record(ctx, []app.StoredEvent{input.Row})
			if err != nil || len(rows) != 1 {
				return fmt.Errorf("single event production recorder failed at input %d", index)
			}
			row := rows[0]
			ref := replayEventRef{EventID: row.ID, Kind: row.Event.Kind, Topic: row.Event.Topic, Tokens: row.Event.Tokens, Scope: row.Event.Scope, TS: row.Event.TS}
			signal, err := events.NormalizeValidationSignal(row)
			if err != nil || signal == nil {
				return fmt.Errorf("approved input has no normalized signal")
			}
			op := replayOperation{Phase: "event", TS: clock.UnixMilli(), ReplayID: input.StableID, Event: &ref, Strength: signal.Strength, Weight: signal.Weight, Points: []replayPoint{}}
			if row.Event.Topic != "" {
				p, err := point(row.Event.Topic, "event", input.StableID)
				if err != nil {
					return err
				}
				op.Points = append(op.Points, p)
				series[p.Topic] = append(series[p.Topic], p)
			}
			if err = trace.Encode(op); err != nil {
				return err
			}
			weights = append(weights, op)
			if index > 0 && inputs[index-1].Row.Event.TS == row.Event.TS {
				equalTSPairs++
			}
			index++
			dayInputs++
		}
		clock = dayEnd
		dailyEvents, err := repo.ListEvents(ctx, day.UnixMilli(), dayEnd.UnixMilli())
		if err != nil || len(dailyEvents) != dayInputs {
			return fmt.Errorf("daily receipt count differs from sequential inputs")
		}
		signals, err := repo.ListSignals(ctx, day.UnixMilli(), dayEnd.UnixMilli())
		if err != nil {
			return err
		}
		daily := replaySleep{Day: day.Format(time.DateOnly), WindowStart: day.UnixMilli(), WindowEnd: dayEnd.UnixMilli(), InputEvents: len(dailyEvents), InputSignals: len(signals), SignalIDs: []string{}, Result: "PASS"}
		for _, row := range dailyEvents {
			if row.Event.Topic == "" {
				daily.Unassigned++
			}
		}
		for _, signal := range signals {
			daily.SignalIDs = append(daily.SignalIDs, signal.EventID)
		}
		result, err := sleep.Run(ctx, day)
		if err != nil {
			return fmt.Errorf("production Sleep failed for %s", daily.Day)
		}
		daily.StartedAt, daily.CompletedAt, daily.ReceiptID, daily.Replayed = result.Run.StartedAt, result.Run.CompletedAt, result.Run.ID, result.Replayed
		var candidates []app.SleepCandidate
		if err := json.Unmarshal(result.Run.Candidates, &candidates); err != nil {
			return err
		}
		for _, candidate := range candidates {
			if candidate.Kind == "unassigned_pending" {
				daily.Pending++
			}
		}
		if daily.Pending != daily.Unassigned || daily.Replayed || daily.StartedAt != clock.UnixMilli() || daily.CompletedAt != clock.UnixMilli() {
			return fmt.Errorf("Sleep pending bucket or simulated clock mismatch")
		}
		sleeps = append(sleeps, daily)
		op := replayOperation{Phase: "sleep", TS: clock.UnixMilli(), Day: daily.Day, SignalIDs: daily.SignalIDs, InputEvents: dayInputs, Points: []replayPoint{}}
		for _, topic := range topics {
			p, err := point(topic, "sleep", "")
			if err != nil {
				return err
			}
			op.Points = append(op.Points, p)
			series[topic] = append(series[topic], p)
		}
		if err = trace.Encode(op); err != nil {
			return err
		}
		focus, err := engine.AttentionState(ctx)
		if err != nil {
			return err
		}
		ranking := replayRanking{Day: daily.Day, Top5: []replayPoint{}, Changes: []rankChange{}}
		for i, n := range focus {
			if i == 5 {
				break
			}
			p, err := point(n.Topic, "sleep", "")
			if err != nil {
				return err
			}
			ranking.Top5 = append(ranking.Top5, p)
		}
		current := map[string]int{}
		for i, p := range ranking.Top5 {
			current[p.Topic] = i + 1
		}
		for _, topic := range topics {
			if previousRanks[topic] != current[topic] {
				ranking.Changes = append(ranking.Changes, rankChange{topic, previousRanks[topic], current[topic]})
			}
		}
		previousRanks = current
		rankings = append(rankings, ranking)
	}
	if index != len(inputs) {
		return fmt.Errorf("not all inputs recorded")
	}
	persisted := map[string]int{}
	for key, query := range map[string]string{
		"events":             "SELECT COUNT(*) FROM attention_engine_events",
		"signals":            "SELECT COUNT(*) FROM attention_engine_signals",
		"unassigned_events":  "SELECT COUNT(*) FROM attention_engine_events WHERE topic=''",
		"unassigned_signals": "SELECT COUNT(*) FROM attention_engine_signals WHERE topic=''",
		"unique_commits":     "SELECT COUNT(DISTINCT origin_key) FROM attention_engine_events WHERE kind='commit'",
		"sleep_runs":         "SELECT COUNT(*) FROM attention_engine_sleep_runs",
		"nodes":              "SELECT COUNT(*) FROM attention_engine_nodes",
	} {
		var count int
		if err = db.Conn().QueryRowContext(ctx, query).Scan(&count); err != nil {
			return err
		}
		persisted[key] = count
	}
	if persisted["events"] != len(inputs) || persisted["signals"] != len(inputs) || persisted["sleep_runs"] != 30 || persisted["nodes"] != len(topics) {
		return fmt.Errorf("final persistence count mismatch")
	}
	emptyDays, pending := 0, 0
	for _, run := range sleeps {
		if run.InputEvents == 0 {
			emptyDays++
		}
		pending += run.Pending
	}
	summary := map[string]any{"semantic_policy": events.SemanticPolicyVersion, "started_at": started, "ended_at": time.Now().UTC(), "window_start": start, "window_end_exclusive": end, "source_counts": counts, "persisted": persisted, "empty_days": emptyDays, "unassigned_pending_total": pending, "equal_ts_adjacent_pairs": equalTSPairs, "single_record_calls": len(inputs), "sort": "ts then stable semantic event identity (source-origin lexical ordinal); store receipt ID assigned once by M0", "weight_sample_seed": v1WeightSampleSeed, "production_replay": "PASS", "independent_M2": "PENDING", "user_focus_confirmation": "PENDING", "human_50_weight_review": "PENDING", "V1_human_acceptance": "PENDING", "EXPERIMENT": "UNLOCKED", "V4": "NOT_RUN", "facts_projector": "NOT_CONFIGURED", "cluster_interpreter": "NOT_CONFIGURED"}
	for _, file := range []struct {
		name string
		data any
	}{{"summary-safe.json", summary}, {"sleep-runs-safe.json", sleeps}, {"topic-timeseries-safe.json", series}, {"daily-top5-safe.json", rankings}} {
		if err = writeExclusive(filepath.Join(out, file.name), file.data); err != nil {
			return err
		}
	}
	if err = writeFocusObservations(out, rankings); err != nil {
		return err
	}
	return writeWeightSample(out, weights)
}

func writeFocusObservations(out string, rankings []replayRanking) error {
	var days []map[string]any
	septemberDays, septemberFirst, septemberWithEvidence, septemberEvidenceFirst := 0, 0, 0, 0
	afterOct8, afterOct8Top2, fromOct8, fromOct8Top2 := 0, 0, 0, 0
	for _, ranking := range rankings {
		ranks := map[string]int{}
		allFloor := true
		for i, point := range ranking.Top5 {
			ranks[point.Topic] = i + 1
			if point.Attention > .05 {
				allFloor = false
			}
		}
		morph, astro := ranks["Morphogenesis"], ranks["Astrocyte/Attention Engine"]
		if ranking.Day < "2026-10-01" {
			septemberDays++
			if morph == 1 {
				septemberFirst++
			}
			if !allFloor {
				septemberWithEvidence++
				if morph == 1 {
					septemberEvidenceFirst++
				}
			}
		}
		if ranking.Day >= "2026-10-08" {
			fromOct8++
			if astro > 0 && astro <= 2 {
				fromOct8Top2++
			}
		}
		if ranking.Day > "2026-10-08" {
			afterOct8++
			if astro > 0 && astro <= 2 {
				afterOct8Top2++
			}
		}
		days = append(days, map[string]any{"day": ranking.Day, "morphogenesis_rank_in_top5": morph, "astrocyte_rank_in_top5": astro, "all_seed_attention_at_floor": allFloor})
	}
	return writeExclusive(filepath.Join(out, "focus-anchor-observations-safe.json"), map[string]any{
		"result": "MODEL_RECONSTRUCTION_ONLY", "user_confirmation": "PENDING", "rank_zero": "outside top5; six-way floor ties use lexical topic ordering",
		"september":                      map[string]int{"days": septemberDays, "morphogenesis_first": septemberFirst, "days_above_floor": septemberWithEvidence, "morphogenesis_first_on_above_floor_days": septemberEvidenceFirst},
		"astrocyte_after_oct8_exclusive": map[string]int{"days": afterOct8, "top2": afterOct8Top2},
		"astrocyte_from_oct8_inclusive":  map[string]int{"days": fromOct8, "top2": fromOct8Top2}, "days": days})
}

func writeWeightSample(out string, operations []replayOperation) error {
	// Sample from every mapped event, including unassigned, before human review.
	// Source ordinals make selection stable across random receipt IDs and runs.
	ordered := append([]replayOperation(nil), operations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ReplayID < ordered[j].ReplayID })
	indices := rand.New(rand.NewSource(v1WeightSampleSeed)).Perm(len(ordered))
	var sample []map[string]any
	for _, i := range indices[:min(50, len(indices))] {
		op := ordered[i]
		sample = append(sample, map[string]any{"replay_id": op.ReplayID, "event": op.Event, "s": op.Strength, "w": op.Weight, "contribution": op.Strength * op.Weight, "attention_applied": op.Event.Topic != "", "human_review": "PENDING"})
	}
	return writeExclusive(filepath.Join(out, "weights-50-safe.json"), map[string]any{"seed": v1WeightSampleSeed, "population": len(ordered), "method": "Go math/rand permutation of lexical stable semantic IDs; first 50 without replacement", "human_review": "PENDING", "samples": sample})
}
