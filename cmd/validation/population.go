package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type populationEnvelope struct {
	Source         string                   `json:"source"`
	Project        string                   `json:"project"`
	FileIndex      int                      `json:"file_index"`
	Line           int                      `json:"line"`
	Blocks         []int                    `json:"blocks"`
	Association    events.NativeAssociation `json:"association"`
	AssociationBad bool                     `json:"association_bad"`
	RawBase64      string                   `json:"raw_base64"`
}

func localLines(path string, consume func([]byte) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open local population failed")
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 4096), 64*1024*1024)
	for s.Scan() {
		if err := consume(s.Bytes()); err != nil {
			return err
		}
	}
	if s.Err() != nil {
		return fmt.Errorf("read local population failed")
	}
	return nil
}

func countPopulationRow(row app.StoredEvent, d *diagnosis) {
	d.Counts["retained"]++
	a := events.ClassifyNativeReceipt(row)
	if row.Event != nil {
		a.Topic = row.Event.Topic
	}
	if a.Topic == "" {
		d.Counts["unassigned"]++
		d.Counts["unassigned_source:"+row.Observation.Source]++
		scopeGroup := events.MatchSeedTopic(row.Observation.Scope, "").Topic
		if scopeGroup == "" {
			scopeGroup = "outside_fixed_seed_scope"
		}
		d.Counts["unassigned_scope:"+scopeGroup]++
	} else {
		d.Counts["assigned"]++
		d.Topics[a.Topic]++
	}
	if row.Event == nil {
		d.Counts["unknown_kind"]++
		return
	}
	d.Counts["mapped"]++
	d.Kinds[row.Event.Kind]++
	if row.Event.Topic != "" {
		d.Counts["kind_topic_complete"]++
	}
	signal, err := domain.NormalizeSignal(row.OriginKey, *row.Event, row.Observation.Outcome, row.Observation.Depth, 1)
	if err != nil {
		d.Counts["normalization_error"]++
		return
	}
	if signal == nil {
		d.Counts["unknown_depth_or_outcome_blocks_signal"]++
		return
	}
	d.Counts["normalizable_strength_1"]++
	if signal.Topic != "" {
		d.Counts["signal_eligible_strength_1"]++
	}
}

func runPopulation(dir, gitPath, output string) error {
	d := &diagnosis{Counts: map[string]int{"negative_false_cognitive": 0, "adapter_rejected": 0, "association_rejected": 0, "negative_adapter_rejected": 0, "normalization_error": 0, "unknown_depth_or_outcome_blocks_signal": 0}, Kinds: map[string]int{}, Topics: map[string]int{}}
	err := localLines(filepath.Join(dir, "tool-records-local.jsonl"), func(line []byte) error {
		var v populationEnvelope
		if json.Unmarshal(line, &v) != nil || len(v.Blocks) == 0 {
			return fmt.Errorf("invalid tool envelope")
		}
		raw, e := base64.StdEncoding.DecodeString(v.RawBase64)
		if e != nil {
			return fmt.Errorf("invalid tool encoding")
		}
		countShape(raw, d)
		for _, block := range v.Blocks {
			d.RecordShapes["unit_form:"+events.NativeToolForm(v.Source, raw, block)]++
			d.Counts["tool_N"]++
			d.Counts[v.Source+"_N"]++
			if v.AssociationBad {
				d.Counts["association_rejected"]++
				continue
			}
			row, e := events.MapNativeToolRecord(v.Source, fmt.Sprintf("native:%s:%d:%d:block:%d", v.Source, v.FileIndex, v.Line, block), v.Project, raw, 0, v.Association, block)
			if e != nil {
				d.Counts["adapter_rejected"]++
				continue
			}
			if row.Event == nil {
				d.Counts["unknown_"+events.NativeToolForm(v.Source, raw, block)]++
				d.Counts["unknown_reason:"+events.NativeToolUnknownReason(v.Source, raw, block)]++
			}
			countPopulationRow(row, d)
		}
		return nil
	})
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	err = localLines(gitPath, func(line []byte) error {
		var v struct {
			Project string          `json:"project"`
			Record  json.RawMessage `json:"record"`
		}
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid frozen git envelope")
		}
		row, e := events.MapGitLogRecord(v.Project, v.Record, 0)
		if e != nil {
			return fmt.Errorf("frozen git rejected")
		}
		if seen[row.OriginKey] {
			return fmt.Errorf("duplicate frozen commit")
		}
		seen[row.OriginKey] = true
		var commit events.GitLogRecord
		_ = json.Unmarshal(v.Record, &commit)
		for _, topic := range events.ClassifyGitTopics(v.Project, commit.Subject).KeywordTopics {
			d.Counts["git_keyword_topic:"+topic]++
		}
		d.Counts["git_N"]++
		countPopulationRow(row, d)
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != 1893 {
		return fmt.Errorf("frozen git count changed")
	}
	err = localLines(filepath.Join(dir, "negative-200-local.jsonl"), func(line []byte) error {
		var v populationEnvelope
		if json.Unmarshal(line, &v) != nil {
			return fmt.Errorf("invalid negative envelope")
		}
		raw, e := base64.StdEncoding.DecodeString(v.RawBase64)
		if e != nil {
			return fmt.Errorf("invalid negative encoding")
		}
		d.Counts["negative_N"]++
		row, e := events.MapAssociatedNativeRecord(v.Source, "negative", v.Project, raw, 0, v.Association)
		if e != nil {
			d.Counts["negative_adapter_rejected"]++
			return nil
		}
		if row.Event != nil {
			d.Counts["negative_false_cognitive"]++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if d.Counts["negative_N"] != 200 {
		return fmt.Errorf("negative sample count changed")
	}
	d.Counts["N"] = d.Counts["tool_N"] + d.Counts["git_N"]
	b, e := json.MarshalIndent(d, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return fmt.Errorf("create new population result failed")
	}
	defer f.Close()
	if _, e = f.Write(b); e != nil {
		return fmt.Errorf("write population result failed")
	}
	fmt.Println(string(b))
	return nil
}
