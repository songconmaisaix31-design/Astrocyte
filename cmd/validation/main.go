// Command validation diagnoses frozen local native samples without persistence.
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type sample struct {
	Source         string                    `json:"source"`
	Ordinal        int                       `json:"ordinal"`
	Project        string                    `json:"project"`
	HeaderComplete bool                      `json:"header_complete"`
	RawBase64      string                    `json:"raw_base64"`
	Association    *events.NativeAssociation `json:"-"`
}

type diagnosis struct {
	Counts       map[string]int `json:"counts"`
	Kinds        map[string]int `json:"mapped_kinds"`
	NativeTypes  map[string]int `json:"native_types"`
	RecordShapes map[string]int `json:"record_shapes"`
	Topics       map[string]int `json:"seed_topics"`
}

func schemaTag(value string, allowed ...string) string {
	for _, tag := range allowed {
		if value == tag {
			return tag
		}
	}
	return "other"
}

func countShape(raw []byte, d *diagnosis) {
	if d.NativeTypes == nil {
		d.NativeTypes = map[string]int{}
	}
	if d.RecordShapes == nil {
		d.RecordShapes = map[string]int{}
	}
	var v struct {
		Type    string `json:"type"`
		Payload struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"payload"`
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	_ = json.Unmarshal(raw, &v)
	d.NativeTypes[schemaTag(v.Type, "session_meta", "turn_context", "response_item", "event_msg", "assistant", "user", "system", "progress", "file-history-snapshot", "queue-operation", "summary")]++
	if v.Type == "response_item" || v.Type == "event_msg" {
		d.RecordShapes[v.Type+":"+schemaTag(v.Payload.Type, "message", "function_call", "function_call_output", "custom_tool_call", "custom_tool_call_output", "reasoning", "token_count", "agent_message", "user_message", "agent_reasoning", "task_started", "task_complete", "turn_aborted")]++
		if v.Payload.Type == "function_call" || v.Payload.Type == "custom_tool_call" {
			d.RecordShapes["tool:"+schemaTag(v.Payload.Name, "functions.exec", "functions.apply_patch", "functions.exec_command", "functions.shell_command", "functions.write_stdin", "exec_command", "shell_command", "apply_patch", "web.run")]++
		}
	} else if v.Type == "assistant" || v.Type == "user" {
		var blocks []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		}
		if json.Unmarshal(v.Message.Content, &blocks) == nil {
			for _, b := range blocks {
				d.RecordShapes["message_block:"+schemaTag(b.Type, "text", "thinking", "tool_use", "tool_result")]++
			}
		} else {
			d.RecordShapes["message_content:scalar_or_missing"]++
		}
	}
}

func diagnose(v sample, d *diagnosis) {
	d.Counts["sample_input"]++
	raw, err := base64.StdEncoding.DecodeString(v.RawBase64)
	if err != nil || !json.Valid(raw) {
		d.Counts["input_parse_rejected"]++
		return
	}
	d.Counts["input_parse_pass"]++
	countShape(raw, d)
	if !v.HeaderComplete {
		d.Counts["header_authorization_missing"]++
	}
	association := events.NativeAssociation{}
	if v.Association != nil {
		association = *v.Association
	}
	row, err := events.MapAssociatedNativeRecord(v.Source, fmt.Sprintf("diagnosis:%s:%d", v.Source, v.Ordinal), v.Project, raw, 0, association)
	if err != nil {
		d.Counts["adapter_project_or_source_rejected"]++
		return
	}
	d.Counts["adapter_pass"]++
	assignment := events.ClassifyNativeReceipt(row)
	d.Counts["retained_local_raw"]++
	d.Counts["classification_"+assignment.Bucket]++
	if assignment.Topic != "" {
		if d.Topics == nil {
			d.Topics = map[string]int{}
		}
		d.Topics[assignment.Topic]++
	}
	if row.Observation.Depth == "unknown" {
		d.Counts["unknown_depth"]++
	}
	if row.Observation.Outcome == "unknown" {
		d.Counts["unknown_outcome"]++
	}
	if row.Observation.Tokens == nil {
		d.Counts["unknown_tokens"]++
	}
	if row.Observation.Session == "" {
		d.Counts["adapter_session_empty"]++
	}
	if row.Event == nil {
		d.Counts["unknown_kind"]++
		d.Counts["retained_observation_candidate"]++
		return
	}
	d.Counts["mapped_event"]++
	d.Kinds[row.Event.Kind]++
	d.Counts["nil_strength_blocks_pending_service"]++
	if row.Event.Topic == "" {
		d.Counts["empty_topic"]++
		d.Counts["retained_unassigned_candidate"]++
	}
	// Explicit 1 is only a diagnostic semantic probe, never a product policy.
	signal, err := domain.NormalizeSignal("diagnosis", *row.Event, row.Observation.Outcome, row.Observation.Depth, 1)
	if err != nil {
		d.Counts["normalization_error"]++
		return
	}
	if signal == nil {
		d.Counts["normalization_nil_unknown_depth_or_outcome"]++
		return
	}
	d.Counts["normalizable_with_probe_strength"]++
	if signal.Topic == "" {
		d.Counts["topic_gate_blocked"]++
		return
	}
	d.Counts["topic_gate_pass"]++
	if row.Observation.Session == "" || row.Observation.Scope == "" {
		d.Counts["association_gate_blocked"]++
		return
	}
	d.Counts["cognitive_signal_candidate_with_probe_strength"]++
	d.Counts["nil_dynamics_configuration_blocked"]++
}

func run(input, output string) error {
	return runAssociated(input, output, "")
}

func runGit(input, output string) error {
	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("open local git input failed")
	}
	defer f.Close()
	d := &diagnosis{Counts: map[string]int{"eligible_current_configuration": 0}, Kinds: map[string]int{}, Topics: map[string]int{}}
	seen := map[string]bool{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var v struct {
			Project string          `json:"project"`
			Record  json.RawMessage `json:"record"`
		}
		if json.Unmarshal(scanner.Bytes(), &v) != nil {
			return fmt.Errorf("invalid git input envelope")
		}
		row, err := events.MapGitLogRecord(v.Project, v.Record, 0)
		if err != nil {
			return fmt.Errorf("git mapping rejected")
		}
		if seen[row.OriginKey] {
			d.Counts["duplicate_commit_id"]++
			continue
		}
		seen[row.OriginKey] = true
		d.Counts["retained_local_raw"]++
		d.Counts["mapped_event"]++
		d.Kinds[row.Event.Kind]++
		if row.Event.Topic != "" {
			d.Counts["classification_assigned"]++
			d.Topics[row.Event.Topic]++
		} else {
			d.Counts["classification_unassigned"]++
		}
		signal, err := domain.NormalizeSignal(row.OriginKey, *row.Event, row.Observation.Outcome, row.Observation.Depth, 1)
		if err != nil {
			return fmt.Errorf("git normalization failed")
		}
		if signal != nil {
			d.Counts["normalizable_with_probe_strength"]++
			if signal.Topic != "" {
				d.Counts["cognitive_signal_candidate_with_probe_strength"]++
			}
		}
	}
	if scanner.Err() != nil {
		return fmt.Errorf("read local git input failed")
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	fout, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create derived git output failed")
	}
	defer fout.Close()
	if _, err = fout.Write(b); err != nil {
		return fmt.Errorf("write derived git output failed")
	}
	fmt.Println(string(b))
	return nil
}

func runAssociated(input, output, contextPath string) error {
	contexts := map[string]events.NativeAssociation{}
	if contextPath != "" {
		b, err := os.ReadFile(contextPath)
		if err != nil {
			return fmt.Errorf("read observed association context failed")
		}
		var values []struct {
			Source  string `json:"source"`
			Ordinal int    `json:"ordinal"`
			events.NativeAssociation
		}
		if json.Unmarshal(b, &values) != nil || len(values) != 100 {
			return fmt.Errorf("incomplete association context")
		}
		for _, v := range values {
			key := fmt.Sprintf("%s:%d", v.Source, v.Ordinal)
			if _, exists := contexts[key]; exists {
				return fmt.Errorf("duplicate association context")
			}
			contexts[key] = v.NativeAssociation
		}
	}
	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("open frozen input failed")
	}
	defer f.Close()
	all := map[string]*diagnosis{}
	seen := map[string]bool{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 64*1024*1024)
	for scanner.Scan() {
		var v sample
		if json.Unmarshal(scanner.Bytes(), &v) != nil || (v.Source != "codex" && v.Source != "claude_code") {
			return fmt.Errorf("invalid sample envelope")
		}
		key := fmt.Sprintf("%s:%d", v.Source, v.Ordinal)
		if v.Ordinal < 0 || v.Ordinal >= 50 || seen[key] {
			return fmt.Errorf("invalid or duplicate sample ordinal")
		}
		seen[key] = true
		if contextPath != "" {
			association, ok := contexts[key]
			if !ok {
				return fmt.Errorf("missing association context")
			}
			v.Association = &association
		}
		d := all[v.Source]
		if d == nil {
			d = &diagnosis{Counts: map[string]int{"eligible_current_configuration": 0}, Kinds: map[string]int{}}
			all[v.Source] = d
		}
		diagnose(v, d)
	}
	if scanner.Err() != nil {
		return fmt.Errorf("read frozen input failed")
	}
	if len(seen) != 100 || len(all) != 2 {
		return fmt.Errorf("incomplete frozen sample: require exactly 50 per client")
	}
	b, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	fout, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create derived output failed")
	}
	defer fout.Close()
	if _, err = fout.Write(b); err != nil {
		return fmt.Errorf("write derived output failed")
	}
	fmt.Println(string(b))
	return nil
}

func main() {
	input := flag.String("input", "", "frozen local sample-first.jsonl")
	output := flag.String("output", "", "new derived stage-count JSON file")
	contextPath := flag.String("associations", "", "separate local observed file association context")
	gitInput := flag.String("git-input", "", "local actual git log input, separate from frozen100")
	populationDir := flag.String("population-dir", "", "frozen full structural tool population and independent negatives")
	flag.Parse()
	if *populationDir != "" {
		if *input != "" || *contextPath != "" || *gitInput == "" || *output == "" {
			fmt.Fprintln(os.Stderr, "population requires frozen git and output only")
			os.Exit(2)
		}
		if err := runPopulation(*populationDir, *gitInput, *output); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if (*input == "" && *gitInput == "") || (*input != "" && *gitInput != "") || *output == "" {
		fmt.Fprintln(os.Stderr, "input and output required")
		os.Exit(2)
	}
	if *gitInput != "" {
		if *contextPath != "" {
			fmt.Fprintln(os.Stderr, "git input cannot use native associations")
			os.Exit(2)
		}
		if err := runGit(*gitInput, *output); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := runAssociated(*input, *output, *contextPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
