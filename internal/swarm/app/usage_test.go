package app

import (
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestSettledOutcomeAcceptsOriginalInterruptionAndRejectsUnknown(t *testing.T) {
	for _, outcome := range []string{"completed", "failed", "interrupted", "cancelled", "stopped", "not_started"} {
		if !settledOutcome(outcome) {
			t.Fatalf("terminal input outcome %q was not recognized", outcome)
		}
	}
	for _, outcome := range []string{"", "unknown", "inProgress", "running"} {
		if settledOutcome(outcome) {
			t.Fatalf("unresolved input outcome %q was treated as settled", outcome)
		}
	}
}

func TestActualUsageKeepsUnknownAndRejectsContradictoryEvidence(t *testing.T) {
	for _, fact := range []ExecutionFact{{CallsKnown: true}, {TokensKnown: true}, {CallsKnown: true, Usage: &domain.Usage{}}, {TokensKnown: true, Usage: &domain.Usage{}}} {
		if _, err := actualUsage(fact); err == nil {
			t.Fatal("known usage without actual evidence was accepted")
		}
	}
	calls, tokens := int64(1), int64(0)
	u, err := actualUsage(ExecutionFact{CallsKnown: true, Usage: &domain.Usage{Calls: &calls, Tokens: &tokens}})
	if err != nil {
		t.Fatal(err)
	}
	if u.Calls == nil || *u.Calls != 1 || u.Tokens != nil || u.Money != nil {
		t.Fatal("an unknown usage dimension was reported as zero")
	}
	u, err = actualUsage(ExecutionFact{CallsKnown: true, TokensKnown: true, Usage: &domain.Usage{Calls: &calls, Tokens: &tokens}})
	if err != nil || u.Tokens == nil || *u.Tokens != 0 {
		t.Fatal("actual known zero usage was lost", err)
	}
}
