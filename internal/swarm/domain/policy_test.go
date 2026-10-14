package domain

import (
	"testing"
	"time"
)

func TestLocalPolicyFiltersPermissionDependenciesOccupancyAndCapabilities(t *testing.T) {
	now := time.Now()
	m := Mission{ID: "mission", State: "ready", Grant: Grant{WritePaths: []string{"src"}, Limits: Limits{MaxActive: 2, MaxAttempts: 3, MaxNoProgress: 3, Deadline: now.Add(time.Hour)}}}
	input := SourceVersion{ID: "explicit-input", Revision: "fixed-version"}
	items := []WorkItem{
		{ID: "accepted", MissionID: m.ID, State: "accepted", WritePaths: []string{"src/accepted"}},
		{ID: "ready-with-context", MissionID: m.ID, State: "open", WritePaths: []string{"src/ready"}, Dependencies: []string{"accepted"}, Inputs: []SourceVersion{input}, RequiredCapabilities: []string{"read_context"}},
		{ID: "ready-without-context", MissionID: m.ID, State: "open", WritePaths: []string{"src/other"}},
		{ID: "unapproved", MissionID: m.ID, State: "open", WritePaths: []string{"private"}},
		{ID: "dependency-unready", MissionID: m.ID, State: "open", WritePaths: []string{"src/later"}, Dependencies: []string{"missing"}},
		{ID: "capability-unknown", MissionID: m.ID, State: "open", WritePaths: []string{"src/native"}, RequiredCapabilities: []string{"resume"}},
		{ID: "occupied", MissionID: m.ID, State: "open", WritePaths: []string{"src/owned/file"}},
	}
	lease := Lease{WritePaths: []string{"src/owned"}, ExpiresAt: now.Add(-time.Hour)}
	selected, err := SelectWorkItems(m, items, []Lease{lease}, SelectionContext{Inputs: []SourceVersion{input}, Capabilities: []string{"read_context"}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].WorkItem.ID != "ready-with-context" || selected[1].WorkItem.ID != "ready-without-context" {
		t.Fatalf("unexpected authorized selection: %v", selected)
	}
	for _, c := range selected {
		if c.PolicyVersion != LocalPolicyVersion {
			t.Fatal("selection lost policy version")
		}
	}
	selected, err = SelectWorkItems(m, items, []Lease{lease, lease}, SelectionContext{Capabilities: []string{"read_context"}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 0 {
		t.Fatal("project executor limit did not constrain selection")
	}
	m.NoProgress = 3
	if _, err = SelectWorkItems(m, items, nil, SelectionContext{}, now); err == nil {
		t.Fatal("no-progress stop condition was ignored")
	}
}

func TestLocalPolicyFiltersExhaustedOrUnknownBoundedBudgetBeforeRanking(t *testing.T) {
	now := time.Now()
	limit, used := int64(2), int64(1)
	m := Mission{ID: "mission", State: "ready", Used: Usage{Calls: &used}, Reserved: Resources{Calls: 1}, Grant: Grant{WritePaths: []string{"src"}, Budget: Budget{Calls: &limit}, Limits: Limits{MaxActive: 2, MaxAttempts: 3, Deadline: now.Add(time.Hour)}}}
	items := []WorkItem{{ID: "available", MissionID: m.ID, State: "open", WritePaths: []string{"src/file"}}}
	check := func(want int, capabilities []string) {
		t.Helper()
		result, err := SelectWorkItems(m, items, nil, SelectionContext{Capabilities: capabilities}, now)
		if err != nil || len(result) != want {
			t.Fatalf("unexpected budget-filtered selection: count=%d error=%v", len(result), err)
		}
	}
	check(0, nil)
	m.Reserved.Calls = 0
	check(1, nil)
	m.Used.Calls = nil
	check(0, nil)
	m.Used.Calls = &used
	m.Grant.Budget.Tokens = &limit
	check(0, []string{"token_budget"})
	m.Used.Tokens = &used
	check(0, nil)
	check(1, []string{"token_budget"})
}
