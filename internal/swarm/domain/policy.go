package domain

import (
	"slices"
	"sort"
	"time"
)

const LocalPolicyVersion = "swarm.fixed.v1"

type Candidate struct {
	WorkItem       WorkItem `json:"work_item"`
	PolicyVersion  string   `json:"policy_version"`
	MatchingInputs int      `json:"matching_inputs"`
}

type SelectionContext struct {
	Inputs       []SourceVersion
	Capabilities []string
}

// 先过滤当前批准与占用，再按固定版本上下文匹配排序。
func SelectWorkItems(m Mission, items []WorkItem, leases []Lease, c SelectionContext, now time.Time) ([]Candidate, error) {
	if err := m.CanExecute(now); err != nil {
		return nil, err
	}
	result := []Candidate{}
	if len(leases) >= m.Grant.Limits.MaxActive {
		return result, nil
	}
	usage, err := m.Used.AddReserved(m.Reserved)
	if err != nil {
		return nil, err
	}
	if !usage.Within(m.Grant.Budget) || m.Grant.Budget.Calls != nil && (usage.Calls == nil || *usage.Calls >= *m.Grant.Budget.Calls) {
		return result, nil
	}
	if m.Grant.Budget.Tokens != nil && !slices.Contains(c.Capabilities, "token_budget") || m.Grant.Budget.Money != nil && !slices.Contains(c.Capabilities, "money_budget") {
		return result, nil
	}
	states := map[string]string{}
	for _, w := range items {
		states[w.ID] = w.State
	}
	for _, w := range items {
		if w.MissionID != m.ID || (w.State != "open" && w.State != "revision_needed") || w.Attempts >= m.Grant.Limits.MaxAttempts || !Within(w.WritePaths, m.Grant.WritePaths) {
			continue
		}
		ready := true
		for _, id := range w.Dependencies {
			if states[id] != "accepted" {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		for _, capability := range w.RequiredCapabilities {
			if !slices.Contains(c.Capabilities, capability) {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		for _, l := range leases {
			if Overlaps(l.WritePaths, w.WritePaths) {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		matches := 0
		for _, input := range w.Inputs {
			if slices.Contains(c.Inputs, input) {
				matches++
			}
		}
		result = append(result, Candidate{WorkItem: w, PolicyVersion: LocalPolicyVersion, MatchingInputs: matches})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].MatchingInputs != result[j].MatchingInputs {
			return result[i].MatchingInputs > result[j].MatchingInputs
		}
		return result[i].WorkItem.ID < result[j].WorkItem.ID
	})
	return result, nil
}
