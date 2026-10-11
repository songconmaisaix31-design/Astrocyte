package domain

import (
	"errors"
	"testing"
	"time"
)

func TestVerdictAndStatusVocabulary(t *testing.T) {
	for _, v := range []string{VerdictAvoid, VerdictPrefer, VerdictCondition} {
		if !ValidVerdict(v) {
			t.Fatalf("valid verdict %q rejected", v)
		}
	}
	for _, v := range []string{"", "neutral", "AVOID"} {
		if ValidVerdict(v) {
			t.Fatalf("invalid verdict %q accepted", v)
		}
	}
	for _, s := range []string{StatusQuarantine, StatusApproved} {
		if !ValidStatus(s) {
			t.Fatalf("valid status %q rejected", s)
		}
	}
	if ValidStatus("inbox") || ValidStatus("") {
		t.Fatal("invalid status accepted")
	}
}

func TestOnlyAgentTaskFailureIsCaptureEligible(t *testing.T) {
	if !CaptureEligible(KindAgentTask, OutcomeFailed) {
		t.Fatal("agent_task failure not captured")
	}
	for _, tt := range []struct{ kind, outcome string }{
		{KindAgentTask, OutcomeSucceeded},
		{"job_failed", OutcomeFailed},
		{"import", OutcomeFailed},
		{"distillation", OutcomeFailed},
		{"edit_file", OutcomeFailed},
		{"", OutcomeFailed},
	} {
		if CaptureEligible(tt.kind, tt.outcome) {
			t.Fatalf("%q/%q disguised as agent_task failure", tt.kind, tt.outcome)
		}
	}
}

func TestCaptureRequiresPatternConditionAndEvidence(t *testing.T) {
	if err := ValidateCapture("", VerdictAvoid, "", 1); !errors.Is(err, ErrInvalid) {
		t.Fatal("empty pattern accepted")
	}
	if err := ValidateCapture("centralized scheduler", "neutral", "", 1); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown verdict accepted")
	}
	if err := ValidateCapture("centralized scheduler", VerdictCondition, "", 1); !errors.Is(err, ErrInvalid) {
		t.Fatal("condition verdict without trigger condition accepted")
	}
	if err := ValidateCapture("centralized scheduler", VerdictAvoid, "", 0); !errors.Is(err, ErrEvidence) {
		t.Fatal("candidate without evidence accepted")
	}
	if err := ValidateCapture("centralized scheduler", VerdictCondition, ">100 agents", 2); err != nil {
		t.Fatal(err)
	}
}

func TestPromotionRequiresHumanGateTransitions(t *testing.T) {
	approved, err := Promote(StatusQuarantine, ActionApprove)
	if err != nil || approved != StatusApproved {
		t.Fatalf("quarantine approve: %s %v", approved, err)
	}
	demoted, err := Promote(StatusApproved, ActionDemote)
	if err != nil || demoted != StatusQuarantine {
		t.Fatalf("approved demote: %s %v", demoted, err)
	}
	for _, tt := range []struct{ status, action string }{
		{StatusQuarantine, ActionDemote},
		{StatusApproved, ActionApprove},
		{StatusQuarantine, "delete"},
		{StatusApproved, "delete"},
		{"inbox", ActionApprove},
	} {
		if _, err := Promote(tt.status, tt.action); !errors.Is(err, ErrInvalid) {
			t.Fatalf("illegal transition %q/%q accepted", tt.status, tt.action)
		}
	}
}

func TestProjectIsolationAndAuditedShare(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	shares := []ShareRecord{{SharedBy: "human", SharedToProject: "B", SharedAt: now}}

	if !ScopedForProject("A", shares, "A") {
		t.Fatal("owner project not visible to itself")
	}
	if ScopedForProject("A", shares, "C") {
		t.Fatal("unshared project sees the experience")
	}
	if !ScopedForProject("A", shares, "B") {
		t.Fatal("shared project not visible")
	}

	if err := ValidateShare(ShareRecord{SharedBy: "human", SharedToProject: "B", SharedAt: now}, "A"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateShare(ShareRecord{SharedBy: "human", SharedToProject: "A", SharedAt: now}, "A"); !errors.Is(err, ErrInvalid) {
		t.Fatal("share back to owner accepted")
	}
	if err := ValidateShare(ShareRecord{SharedBy: "", SharedToProject: "B", SharedAt: now}, "A"); !errors.Is(err, ErrInvalid) {
		t.Fatal("share without actor accepted")
	}
	if err := ValidateShare(ShareRecord{SharedBy: "human", SharedToProject: "B"}, "A"); !errors.Is(err, ErrInvalid) {
		t.Fatal("share without timestamp accepted")
	}
}
