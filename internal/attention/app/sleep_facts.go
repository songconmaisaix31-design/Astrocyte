package app

import (
	"encoding/json"
	"slices"
	"strings"
)

// SleepFact is a literal claim from an explicitly configured projection. It
// carries no inferred truth or confidence; conflicting values need a human gate.
type SleepFact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SleepFactProjector is a deterministic source-format rule, never an LLM.
// Nil means fact recall is unknown, rather than recall of event identifiers.
type SleepFactProjector func(json.RawMessage) ([]SleepFact, error)

type sleepFluid struct {
	Format         string         `json:"format"`
	Facts          []SleepFact    `json:"facts"`
	Observations   map[string]int `json:"observations"`
	ObservationDay string         `json:"observation_day"`
	PreviousFluid  string         `json:"previous_fluid,omitempty"`
}

func projectSleepFacts(project SleepFactProjector, raw json.RawMessage) ([]SleepFact, error) {
	if project == nil || len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	facts, err := project(raw)
	if err != nil {
		return nil, err
	}
	return uniqueSleepFacts(facts), nil
}

func uniqueSleepFacts(facts []SleepFact) []SleepFact {
	result := make([]SleepFact, 0, len(facts))
	seen := map[SleepFact]bool{}
	for _, fact := range facts {
		if strings.TrimSpace(fact.Key) == "" || seen[fact] {
			continue
		}
		seen[fact] = true
		result = append(result, fact)
	}
	slices.SortFunc(result, func(a, b SleepFact) int {
		if n := strings.Compare(a.Key, b.Key); n != 0 {
			return n
		}
		return strings.Compare(a.Value, b.Value)
	})
	return result
}

func decodeSleepFluid(fluid string) sleepFluid {
	var previous sleepFluid
	if json.Unmarshal([]byte(fluid), &previous) == nil && previous.Format == "sleep-v1" {
		return previous
	}
	return sleepFluid{PreviousFluid: fluid}
}

func storedSleepFacts(project SleepFactProjector, entry StoredKnowledge) ([]SleepFact, error) {
	facts := decodeSleepFluid(entry.Entry.Fluid).Facts
	crystal, err := projectSleepFacts(project, entry.Entry.Crystallized)
	if err != nil {
		return nil, err
	}
	return uniqueSleepFacts(append(slices.Clone(facts), crystal...)), nil
}

func sleepFactKeys(owner, topic string, facts []SleepFact) []string {
	keys := make([]string, 0, len(facts))
	for _, fact := range facts {
		key, _ := json.Marshal([]string{owner, topic, fact.Key, fact.Value})
		keys = append(keys, string(key))
	}
	return keys
}

func sleepFactContradictions(left, right []SleepFact) [][2]SleepFact {
	var pairs [][2]SleepFact
	for _, a := range uniqueSleepFacts(left) {
		for _, b := range uniqueSleepFacts(right) {
			if a.Key == b.Key && a.Value != b.Value {
				pairs = append(pairs, [2]SleepFact{a, b})
			}
		}
	}
	return pairs
}
