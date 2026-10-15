package domain

import (
	"testing"
	"time"
)

func TestResumeAfterReconciliationPreservesConcurrentControlAndTerminalStates(t *testing.T) {
	now := time.Now().UTC()
	for _, state := range []string{"paused", "blocked", "ready", "running", "cancelled", "completed"} {
		t.Run(state, func(t *testing.T) {
			m := Mission{Version: 5, State: state, Reason: "current human control", Grant: Grant{Limits: Limits{Deadline: now.Add(time.Hour)}}}
			before := m
			if err := m.ResumeAfterReconciliation(4, "older recovery", now); err == nil {
				t.Fatal("recovery overwrote a newer mission version")
			}
			if m.Version != before.Version || m.State != before.State || m.Reason != before.Reason {
				t.Fatal("rejected recovery changed current mission state")
			}
			if state == "cancelled" || state == "completed" {
				if err := m.ResumeAfterReconciliation(5, "terminal recovery", now); err == nil {
					t.Fatal("terminal mission resumed")
				}
				if m.Version != before.Version || m.State != before.State || m.Reason != before.Reason {
					t.Fatal("terminal recovery changed current mission state")
				}
				return
			}
			if err := m.ResumeAfterReconciliation(5, "current explicit recovery", now); err != nil {
				t.Fatal(err)
			}
			if m.Version != 6 || m.State != "ready" || m.Reason != "current explicit recovery" {
				t.Fatal("current recovery did not save the expected state")
			}
		})
	}
}

func TestResumeAfterReconciliationRetainsReachedStopCondition(t *testing.T) {
	now := time.Now().UTC()
	m := Mission{Version: 3, State: "paused", NoProgress: 2, Grant: Grant{Limits: Limits{MaxNoProgress: 2, Deadline: now.Add(time.Hour)}}}
	if err := m.ResumeAfterReconciliation(3, "explicit recovery", now); err != nil {
		t.Fatal(err)
	}
	if m.Version != 4 || m.State != "blocked" || m.Reason != "mission stop condition reached" {
		t.Fatal("recovery bypassed the original stop condition")
	}
}
