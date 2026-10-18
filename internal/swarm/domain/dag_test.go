package domain

import (
	"testing"
	"time"
)

func TestDynamicDAGIntegrity(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items []WorkItem
		valid bool
	}{
		{"empty", nil, true},
		{"branch_and_join", []WorkItem{{ID: "read", MissionID: "review"}, {ID: "check", MissionID: "review", Dependencies: []string{"read"}}, {ID: "report", MissionID: "review", Dependencies: []string{"read", "check"}}}, true},
		{"cycle", []WorkItem{{ID: "read", MissionID: "review", Dependencies: []string{"report"}}, {ID: "report", MissionID: "review", Dependencies: []string{"read"}}}, false},
		{"missing_endpoint", []WorkItem{{ID: "report", MissionID: "review", Dependencies: []string{"read"}}}, false},
		{"duplicate_node", []WorkItem{{ID: "read", MissionID: "review"}, {ID: "read", MissionID: "review"}}, false},
		{"duplicate_edge", []WorkItem{{ID: "read", MissionID: "review"}, {ID: "report", MissionID: "review", Dependencies: []string{"read", "read"}}}, false},
		{"self_dependency", []WorkItem{{ID: "read", MissionID: "review", Dependencies: []string{"read"}}}, false},
		{"different_mission", []WorkItem{{ID: "read", MissionID: "another"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateDAG("review", tc.items); (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestDynamicDAGStopLimitsPreserveUnknownAndReservations(t *testing.T) {
	now := time.Now().UTC()
	limit := int64(3)
	m := Mission{LastProgressAt: now, Grant: Grant{Budget: Budget{Calls: &limit}, Limits: Limits{Deadline: now.Add(time.Hour), MaxNoProgress: 2, NoProgressSeconds: 60, MaxWorkItems: 2}}, WorkItemCount: 2, Reserved: Resources{Calls: 3}}
	if reason := m.StopReason(now); reason != "" {
		t.Fatalf("unknown usage, full reservations and task count must not stop existing work: %s", reason)
	}
	used := int64(3)
	m.Used.Calls = &used
	if reason := m.StopReason(now); reason != "call_budget_exhausted" {
		t.Fatal(reason)
	}
	m.Used.Calls = nil
	if reason := m.StopReason(now.Add(time.Minute)); reason != "no_progress_time_limit" {
		t.Fatal(reason)
	}
	m.NoProgress = 2
	if reason := m.StopReason(now); reason != "no_progress_round_limit" {
		t.Fatal(reason)
	}
	if reason := m.StopReason(now.Add(time.Hour)); reason != "deadline_reached" {
		t.Fatal(reason)
	}
}
