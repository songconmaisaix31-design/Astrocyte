package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
)

const v1rRoot = v1EvidenceRoot + "/v1r-w1"
const mcpetCandidateRoot = "C:/Users/DW/orca/mcpet"
const mcpetCandidateRemote = "https://github.com/songconmaisaix31-design/mcpet"

type mcpetSource struct {
	Root                  string         `json:"approved_root"`
	Remote                string         `json:"remote"`
	CommonDir             string         `json:"common_dir"`
	Head                  string         `json:"head"`
	ConfirmationReference string         `json:"confirmation_reference"`
	WindowStart           int64          `json:"window_start"`
	WindowEnd             int64          `json:"window_end_exclusive"`
	UniqueCommits         int            `json:"unique_commits"`
	Days                  map[string]int `json:"days"`
}

func equalValidationPath(a, b string) bool {
	x, e := filepath.Abs(a)
	y, f := filepath.Abs(b)
	return e == nil && f == nil && filepath.Clean(x) == filepath.Clean(y)
}

func validateV1RPaths(population, originalGit, source, out string) error {
	if !equalValidationPath(population, filepath.Join(v1EvidenceRoot, "v21-w1")) ||
		!equalValidationPath(originalGit, filepath.Join(v1EvidenceRoot, "diagnosis-w1/p0-git/git-log-local.jsonl")) ||
		!equalValidationPath(filepath.Dir(source), v1rRoot) || !equalValidationPath(filepath.Dir(out), v1rRoot) || equalValidationPath(source, out) {
		return fmt.Errorf("V1R requires fixed frozen inputs and distinct direct children of v1r-w1")
	}
	return nil
}

func readMcPetSource(dir string) (mcpetSource, error) {
	var source mcpetSource
	for _, path := range []string{filepath.Join(dir, "source-local.json"), filepath.Join(dir, "git-log-local.jsonl")} {
		if events.ValidateLocalReplayPath(path) != nil {
			return source, fmt.Errorf("redirected or unavailable McPet input")
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "source-local.json"))
	if err != nil {
		return source, fmt.Errorf("missing confirmed local McPet source")
	}
	source, err = decodeMcPetSource(b)
	if err != nil {
		return source, err
	}
	observed := events.ObserveRepository(source.Root)
	if !equalValidationPath(observed.CommonDir, source.CommonDir) || observed.Remote != source.Remote || !equalValidationPath(observed.Root, source.Root) {
		return source, fmt.Errorf("confirmed McPet source identity changed")
	}
	return source, nil
}

func decodeMcPetSource(b []byte) (mcpetSource, error) {
	var source mcpetSource
	// Python's original capture emitted integral millisecond bounds as JSON
	// floats. Accept only the exact frozen bounds, preserving the first capture.
	type sourceAlias mcpetSource
	metadata := struct {
		*sourceAlias
		Start float64 `json:"window_start"`
		End   float64 `json:"window_end_exclusive"`
	}{sourceAlias: (*sourceAlias)(&source)}
	if json.Unmarshal(b, &metadata) != nil || metadata.Start != float64(v1Window().UnixMilli()) || metadata.End != float64(v1Window().AddDate(0, 0, 30).UnixMilli()) {
		return source, fmt.Errorf("missing confirmed local McPet source")
	}
	source.WindowStart, source.WindowEnd = int64(metadata.Start), int64(metadata.End)
	if !equalValidationPath(source.Root, mcpetCandidateRoot) || source.Remote != mcpetCandidateRemote || source.CommonDir == "" || source.ConfirmationReference == "" || len(source.Head) != 40 || source.WindowStart != v1Window().UnixMilli() || source.WindowEnd != v1Window().AddDate(0, 0, 30).UnixMilli() {
		return source, fmt.Errorf("McPet source outside the fixed confirmation/window")
	}
	return source, nil
}

func mergeMcPetFacts(inputs []replayInput, source mcpetSource, dir string, mapper *events.NativeSemanticMapper) ([]replayInput, int, int, error) {
	seen := map[string]events.GitLogRecord{}
	for _, input := range inputs {
		if input.Row.Event.Kind == "commit" {
			var record events.GitLogRecord
			if json.Unmarshal(input.Row.RawBytes, &record) != nil {
				return nil, 0, 0, fmt.Errorf("invalid original frozen Git fact")
			}
			seen[input.Row.OriginKey] = record
		}
	}
	newCount, duplicates, sourceCount := 0, 0, 0
	withinSource := map[string]bool{}
	err := localLines(filepath.Join(dir, "git-log-local.jsonl"), func(line []byte) error {
		var v struct {
			Project string          `json:"project"`
			Record  json.RawMessage `json:"record"`
		}
		if json.Unmarshal(line, &v) != nil || !equalValidationPath(v.Project, source.Root) {
			return fmt.Errorf("additional Git fact outside confirmed source")
		}
		row, err := mapper.MapGit(v.Project, v.Record, 0)
		if err != nil || row.Event == nil || row.Event.Topic != "McPet" || row.Event.TS < source.WindowStart || row.Event.TS >= source.WindowEnd || withinSource[row.OriginKey] {
			return fmt.Errorf("invalid additional Git identity/topic/window")
		}
		withinSource[row.OriginKey] = true
		sourceCount++
		var record events.GitLogRecord
		_ = json.Unmarshal(v.Record, &record)
		if previous, exists := seen[row.OriginKey]; exists {
			if previous.TS != record.TS || previous.Subject != record.Subject {
				return fmt.Errorf("same commit ID differs from frozen fact")
			}
			duplicates++
			return nil
		}
		seen[row.OriginKey] = record
		inputs = append(inputs, replayInput{Row: row})
		newCount++
		return nil
	})
	if err != nil || sourceCount != source.UniqueCommits {
		return nil, 0, 0, fmt.Errorf("additional Git source count/facts mismatch")
	}
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Row.OriginKey < inputs[j].Row.OriginKey })
	for i := range inputs {
		inputs[i].StableID = fmt.Sprintf("v1-semantic-%06d", i+1)
	}
	return inputs, newCount, duplicates, nil
}

func runV1R(population, originalGit, sourceDir, out string) error {
	if err := validateV1RPaths(population, originalGit, sourceDir, out); err != nil {
		return err
	}
	for _, path := range []string{filepath.Join(population, "tool-records-local.jsonl"), originalGit, filepath.Dir(out)} {
		if events.ValidateLocalReplayPath(path) != nil {
			return fmt.Errorf("redirected or unavailable V1R path")
		}
	}
	// Identity is confirmed before any raw additional facts are opened.
	source, err := readMcPetSource(sourceDir)
	if err != nil {
		return err
	}
	if err = os.Mkdir(out, 0700); err != nil {
		return fmt.Errorf("V1R output must be exclusively new")
	}
	started := time.Now().UTC()
	mapper := events.NewNativeSemanticMapper()
	// Read original data twice, never recapture: old default V1 must retain its
	// 2832/849 gate, then compare exact stable identities after reclassification.
	old, originalCounts, err := loadV1Inputs(filepath.Join(population, "tool-records-local.jsonl"), originalGit)
	if err != nil || originalCounts["mapped"] != 2832 || originalCounts["unassigned"] != 849 || originalCounts["git_parents"] != 1893 || originalCounts["native_parents"] != 4123 || originalCounts["semantic_units"] != 8914 {
		return fmt.Errorf("original approved V1 population changed")
	}
	inputs, counts, err := loadV1InputsWithMapper(filepath.Join(population, "tool-records-local.jsonl"), originalGit, mapper)
	if err != nil || len(inputs) != len(old) {
		return fmt.Errorf("V1R frozen reclassification changed kind population")
	}
	changed, moved := 0, 0
	for i := range inputs {
		a, b := old[i].Row, inputs[i].Row
		if a.OriginKey != b.OriginKey || a.Event.Kind != b.Event.Kind || a.Event.TS != b.Event.TS {
			return fmt.Errorf("V1R changed frozen kind/identity/time")
		}
		if a.Event.Topic != b.Event.Topic {
			changed++
			if a.Event.Topic == "" && b.Event.Topic == "McPet" {
				moved++
			} else {
				return fmt.Errorf("unexpected frozen source topic change")
			}
		}
	}
	inputs, added, duplicates, err := mergeMcPetFacts(inputs, source, sourceDir, mapper)
	if err != nil {
		return err
	}
	counts["mcpet_captured_commits"], counts["mcpet_added_commits"], counts["mcpet_duplicate_commits"] = source.UniqueCommits, added, duplicates
	counts["original_unassigned"], counts["original_topic_changes"], counts["original_unassigned_to_mcpet"] = 849, changed, moved
	counts["mapped"] += added
	counts["kind:commit"] += added
	counts["semantic_units"] += added
	counts["git_parents"] += added
	counts["cognitive_denominator"] = counts["mapped"] + counts["unresolved_structure"]
	metadata := map[string]any{"mode": "V1R", "repository_policy": events.V1RRepositoryPolicyVersion, "semantic_policy": events.SemanticPolicyVersion, "captured_head": source.Head, "captured_commits": source.UniqueCommits, "added_commits": added, "duplicate_commits": duplicates, "frozen_topic_changes": changed, "frozen_unassigned_to_mcpet": moved, "kind_strength_completeness": float64(counts["mapped"]) / float64(counts["cognitive_denominator"]), "human_acceptance": "PENDING"}
	if err := writeExclusive(filepath.Join(out, "v1r-source-safe.json"), metadata); err != nil {
		return err
	}
	// Export the complete kind-success population before production execution,
	// so an independent consumer can derive its reference without replay state.
	var audit []map[string]any
	for _, input := range inputs {
		signal, err := events.NormalizeValidationSignal(input.Row)
		if err != nil || signal == nil {
			return fmt.Errorf("V1R audit signal binding failed")
		}
		e := input.Row.Event
		audit = append(audit, map[string]any{"event_id": input.StableID, "kind": e.Kind, "topic": e.Topic, "tokens": e.Tokens, "scope": e.Scope, "ts": e.TS, "strength": signal.Strength, "weight": signal.Weight})
	}
	if err := writeExclusive(filepath.Join(out, "semantic-audit-safe.json"), audit); err != nil {
		return err
	}
	if err = replayV1Records(inputs, out, counts, started); err != nil {
		_ = writeExclusive(filepath.Join(out, "first-red.json"), map[string]any{"result": "RED", "error": err.Error(), "started_at": started, "ended_at": time.Now().UTC()})
		return err
	}
	if err := writeV1RCoverage(out, inputs); err != nil {
		return err
	}
	return writeMcPetSupplement(out)
}

func writeMcPetSupplement(out string) error {
	var primary struct {
		Samples []struct {
			ReplayID string         `json:"replay_id"`
			Event    replayEventRef `json:"event"`
		} `json:"samples"`
	}
	b, err := os.ReadFile(filepath.Join(out, "weights-50-safe.json"))
	if err != nil || json.Unmarshal(b, &primary) != nil {
		return fmt.Errorf("primary sample unavailable")
	}
	selected, covered := map[string]bool{}, 0
	for _, sample := range primary.Samples {
		selected[sample.ReplayID] = true
		if sample.Event.Topic == "McPet" {
			covered++
		}
	}
	var candidates []replayOperation
	err = localLines(filepath.Join(out, "operations-safe.jsonl"), func(line []byte) error {
		var operation replayOperation
		if json.Unmarshal(line, &operation) != nil {
			return fmt.Errorf("safe operation decode failed")
		}
		if operation.Phase == "event" && operation.Event.Topic == "McPet" && !selected[operation.ReplayID] {
			candidates = append(candidates, operation)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ReplayID < candidates[j].ReplayID })
	var samples []map[string]any
	if covered == 0 {
		indices := rand.New(rand.NewSource(v1WeightSampleSeed)).Perm(len(candidates))
		for _, i := range indices[:min(10, len(indices))] {
			op := candidates[i]
			samples = append(samples, map[string]any{"replay_id": op.ReplayID, "event": op.Event, "s": op.Strength, "w": op.Weight, "contribution": op.Strength * op.Weight, "human_review": "PENDING"})
		}
	}
	return writeExclusive(filepath.Join(out, "mcpet-supplement-safe.json"), map[string]any{"seed": v1WeightSampleSeed, "primary_mcpet_samples": covered, "eligible_mcpet_outside_primary": len(candidates), "supplement_count": len(samples), "method": "same Go permutation of sorted stable IDs, excluding primary sample", "samples": samples, "human_review": "PENDING"})
}

func writeV1RCoverage(out string, inputs []replayInput) error {
	kinds, topics, sources := map[string]int{}, map[string]int{}, map[string]int{}
	mcpetGit, mcpetNative, assigned := 0, 0, 0
	for _, input := range inputs {
		row := input.Row
		kinds[row.Event.Kind]++
		topics[row.Event.Topic]++
		sources[row.Observation.SourceKind]++
		if row.Event.Topic != "" {
			assigned++
		}
		if row.Event.Topic == "McPet" {
			if row.Event.Kind == "commit" {
				mcpetGit++
			} else {
				mcpetNative++
			}
		}
	}
	return writeExclusive(filepath.Join(out, "coverage-safe.json"), map[string]any{"events": len(inputs), "assigned": assigned, "unassigned": topics[""], "topic_coverage": float64(assigned) / float64(len(inputs)), "by_topic": topics, "by_kind": kinds, "by_source_kind": sources, "mcpet_git": mcpetGit, "mcpet_native": mcpetNative, "human_acceptance": "PENDING"})
}
