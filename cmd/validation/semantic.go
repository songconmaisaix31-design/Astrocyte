package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

const semanticSampleSeed = int64(2026101222)

type semanticAudit struct {
	EventID              string   `json:"event_id"`
	ParentID             string   `json:"parent_id"`
	OriginalTool         string   `json:"original_tool"`
	InnerTool            string   `json:"inner_tool"`
	Structure            string   `json:"structure"`
	Head                 string   `json:"command_head,omitempty"`
	ShellHeads           []string `json:"shell_heads,omitempty"`
	NonCognitiveSegments int      `json:"noncognitive_segments,omitempty"`
	ParseErrors          []string `json:"parse_errors,omitempty"`
	Bucket               string   `json:"bucket"`
	Reason               string   `json:"reason"`
	Kind                 string   `json:"kind"`
	Topic                string   `json:"topic"`
	Scope                string   `json:"scope"`
	RepositoryStatus     string   `json:"repository_status"`
	TS                   int64    `json:"ts"`
	Strength             float64  `json:"strength"`
	Weight               *float64 `json:"weight,omitempty"`
}

func writeExclusive(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("refuse overwriting evidence")
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}

func runSemantic(dir, gitPath, out, policyPath string) error {
	policy, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("missing preregistered policy")
	}
	var p struct {
		Version string `json:"version"`
		Seed    int64  `json:"sample_seed"`
	}
	if json.Unmarshal(policy, &p) != nil || p.Version != events.SemanticPolicyVersion || p.Seed != semanticSampleSeed {
		return fmt.Errorf("policy/code disagreement")
	}
	if err = os.Mkdir(out, 0700); err != nil {
		return fmt.Errorf("require exclusively new output directory")
	}
	if err = writeExclusive(filepath.Join(out, "semantic-policy.json"), json.RawMessage(policy)); err != nil {
		return err
	}
	counts, buckets, kinds, topics, oldDirect, parents := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	repos := map[string]events.RepositoryObservation{}
	mapper := events.NewNativeSemanticMapper()
	observe := func(cwd string) events.RepositoryObservation {
		if o, ok := repos[cwd]; ok {
			return o
		}
		o := mapper.Repository(cwd)
		repos[cwd] = o
		return o
	}
	var audits []semanticAudit
	local := map[string]any{}
	add := func(row app.StoredEvent, s events.SemanticTool, parent, original string, repo events.RepositoryObservation, rawLocal any) {
		a := semanticAudit{EventID: row.OriginKey, ParentID: parent, OriginalTool: original, InnerTool: s.Tool, Structure: s.Structure, Head: s.Head, Bucket: s.Bucket, Reason: s.Reason, Scope: row.Observation.Scope, RepositoryStatus: repo.Status, TS: row.Observation.TS}
		a.ShellHeads, a.NonCognitiveSegments = s.ShellHeads, s.NonCognitiveSegments
		a.ParseErrors = s.ParseErrors
		counts["noncognitive_command_segments"] += s.NonCognitiveSegments
		counts["semantic_total"]++
		buckets[s.Bucket]++
		if s.Bucket == "unresolved_structure" || strings.HasPrefix(s.Bucket, "cognitive_") {
			counts["cognitive_denominator"]++
		} else {
			counts["semantic_noncognitive"]++
			if row.Event != nil {
				counts["semantic_noncognitive_false_positive"]++
			}
		}
		if row.Event != nil {
			a.Kind, a.Topic = row.Event.Kind, row.Event.Topic
			a.Strength = events.ValidationSignalStrength
			kinds[a.Kind]++
			counts["kind_mapped"]++
			if a.Topic != "" {
				counts["topic_assigned"]++
				topics[a.Topic]++
			} else {
				counts["mapped_unassigned"]++
			}
			signal, e := events.NormalizeValidationSignal(row)
			if e != nil {
				counts["normalization_error"]++
			} else if signal != nil {
				a.Weight = &signal.Weight
				counts["normalizable"]++
			}
			// Completeness is kind and explicit s binding. Unknown read depth remains
			// separately visible and cannot manufacture a deep-read signal.
			wantWeight := .6
			if a.Kind == "read_paper" {
				wantWeight = 1
			}
			if e == nil && signal != nil && signal.Strength == 1 && signal.Weight == wantWeight && strings.HasPrefix(s.Bucket, "cognitive_") {
				counts["kind_strength_success"]++
			} else {
				counts["nonstructure_mapping_failure"]++
			}
		} else if strings.HasPrefix(s.Bucket, "cognitive_") {
			counts["nonstructure_mapping_failure"]++
		}
		audits = append(audits, a)
		local[a.EventID] = rawLocal
	}
	err = localLines(filepath.Join(dir, "tool-records-local.jsonl"), func(line []byte) error {
		var v populationEnvelope
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid native envelope")
		}
		raw, e := base64.StdEncoding.DecodeString(v.RawBase64)
		if e != nil {
			return fmt.Errorf("invalid frozen encoding")
		}
		repo := observe(v.Association.Scope)
		for _, block := range v.Blocks {
			parent := fmt.Sprintf("native:%s:%d:%d:block:%d", v.Source, v.FileIndex, v.Line, block)
			if _, ok := parents[parent]; ok {
				return fmt.Errorf("duplicate native parent")
			}
			original := events.NativeToolName(v.Source, raw, block)
			// Historical direct-612 vocabulary is independent of revised mapper table.
			name := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(original, "functions."), "tools."), "container.")
			switch name {
			case "Edit", "Write", "MultiEdit", "apply_patch", "Grep", "Glob", "WebSearch", "WebFetch", "web.run", "ReadPaper", "exec", "Bash", "exec_command", "shell_command":
			default:
				oldDirect[name]++
			}
			if v.AssociationBad {
				return fmt.Errorf("frozen association conflict")
			}
			rows, semantic, e := mapper.MapTool(v.Source, parent, v.Project, raw, 0, v.Association, block)
			if e != nil {
				return fmt.Errorf("native mapping rejected")
			}
			parents[parent] = len(rows)
			counts["native_parent_units"]++
			counts["native_children"] += len(rows)
			forms := map[string]bool{}
			for i, row := range rows {
				forms[semantic[i].Bucket] = true
				add(row, semantic[i], parent, original, repo, map[string]any{"parent": v, "block": block, "semantic_path": semantic[i].Path})
			}
			if len(forms) > 1 {
				counts["mixed_bucket_parents"]++
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	seenGit := map[string]bool{}
	err = localLines(gitPath, func(line []byte) error {
		var v struct {
			Project string          `json:"project"`
			Record  json.RawMessage `json:"record"`
		}
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid frozen git envelope")
		}
		row, e := mapper.MapGit(v.Project, v.Record, 0)
		if e != nil {
			return fmt.Errorf("frozen Git mapping rejected")
		}
		if seenGit[row.OriginKey] {
			return fmt.Errorf("duplicate frozen Git fact")
		}
		seenGit[row.OriginKey] = true
		var commit events.GitLogRecord
		_ = json.Unmarshal(v.Record, &commit)
		for _, t := range events.ClassifyGitTopics(v.Project, commit.Subject).KeywordTopics {
			counts["git_keyword_topic:"+t]++
		}
		repo := observe(v.Project)
		counts["git_parent_units"]++
		parents[row.OriginKey] = 1
		add(row, events.SemanticTool{Path: "0", Tool: "git_adapter", Bucket: "cognitive_commit", Reason: "frozen_committed_fact", Structure: "git_fact"}, row.OriginKey, "git_adapter", repo, json.RawMessage(append([]byte(nil), line...)))
		return nil
	})
	if err != nil {
		return err
	}
	if counts["native_parent_units"] != 4123 || len(seenGit) != 1893 {
		return fmt.Errorf("frozen parent counts changed")
	}
	counts["frozen_parent_total"] = counts["native_parent_units"] + counts["git_parent_units"]
	err = localLines(filepath.Join(dir, "negative-200-local.jsonl"), func(line []byte) error {
		var v populationEnvelope
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid negative envelope")
		}
		raw, e := base64.StdEncoding.DecodeString(v.RawBase64)
		if e != nil {
			return fmt.Errorf("invalid negative encoding")
		}
		row, e := events.MapAssociatedNativeRecord(v.Source, "negative", v.Project, raw, 0, v.Association)
		counts["negative_N"]++
		if e != nil {
			counts["negative_adapter_rejected"]++
		} else if row.Event != nil {
			counts["negative_false_cognitive"]++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if counts["negative_N"] != 200 {
		return fmt.Errorf("negative count changed")
	}
	sort.Slice(audits, func(i, j int) bool { return audits[i].EventID < audits[j].EventID })
	groups := map[string][]semanticAudit{}
	for _, a := range audits {
		groups[a.Bucket] = append(groups[a.Bucket], a)
	}
	sampleSafe := map[string][]semanticAudit{}
	sampleLocal := map[string][]any{}
	for bucket, rows := range groups {
		rng := rand.New(rand.NewSource(semanticSampleSeed))
		indices := rng.Perm(len(rows))
		n := 20
		if len(rows) < n {
			n = len(rows)
		}
		for _, i := range indices[:n] {
			sampleSafe[bucket] = append(sampleSafe[bucket], rows[i])
			sampleLocal[bucket] = append(sampleLocal[bucket], map[string]any{"audit": rows[i], "source_local": local[rows[i].EventID]})
		}
	}
	completeness := float64(counts["kind_strength_success"]) / float64(counts["cognitive_denominator"])
	coverage := float64(counts["topic_assigned"]) / float64(counts["kind_mapped"])
	// Always show explicit zeroes, rather than relying on absent JSON keys.
	for _, key := range []string{"negative_false_cognitive", "negative_adapter_rejected", "semantic_noncognitive_false_positive", "normalization_error", "nonstructure_mapping_failure"} {
		counts[key] += 0
	}
	summary := map[string]any{"version": p.Version, "sample_seed": p.Seed, "counts": counts, "buckets": buckets, "mapped_kinds": kinds, "seed_topics": topics, "old_direct_612_names": oldDirect, "completeness": completeness, "topic_coverage": coverage, "numerical_gate": completeness >= .95 && counts["nonstructure_mapping_failure"] == 0 && counts["negative_false_cognitive"] == 0 && counts["semantic_noncognitive_false_positive"] == 0, "root_review": "PENDING", "V0": "NOT_SELF_APPROVED", "V1": "NOT_RUN"}
	for _, file := range []struct {
		name  string
		value any
	}{{"stages-first.json", summary}, {"semantic-audit-safe.json", audits}, {"samples-safe.json", sampleSafe}, {"samples-local.json", sampleLocal}, {"repositories-local.json", repos}, {"parent-child-counts.json", parents}} {
		if err = writeExclusive(filepath.Join(out, file.name), file.value); err != nil {
			return err
		}
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(b))
	return nil
}
