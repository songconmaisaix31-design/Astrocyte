package domain

import (
	"strings"
	"time"
)

// Experience verdict vocabulary (§4.3).
const (
	VerdictAvoid     = ExperienceVerdictAvoid
	VerdictPrefer    = ExperienceVerdictPrefer
	VerdictCondition = ExperienceVerdictCondition
)

// Experience status vocabulary (§4.3). Candidates are born quarantined and only
// reach approved through the human gate.
const (
	StatusQuarantine = ExperienceStatusQuarantine
	StatusApproved   = ExperienceStatusApproved
)

// Promotion actions for the quarantine↔approved gate. The human gate itself is
// enforced in the application layer by principal kind; these are the only legal
// status transitions.
const (
	ActionApprove = "approve" // quarantine → approved
	ActionDemote  = "demote"  // approved → quarantine
)

// Agent-task event kind and execution outcome vocabulary (§4.1, §8 M5).
const (
	KindAgentTask    = EventKindAgentTask
	OutcomeFailed    = "failed"
	OutcomeSucceeded = "succeeded"
)

func ValidVerdict(v string) bool {
	return v == VerdictAvoid || v == VerdictPrefer || v == VerdictCondition
}

func ValidStatus(s string) bool {
	return s == StatusQuarantine || s == StatusApproved
}

// CaptureEligible reports whether an event must be captured as a quarantine
// Experience candidate: an agent-task that failed. Import/distillation/model job
// failures keep their own kind and are never relabeled as agent_task, so their
// source-type semantics are preserved.
func CaptureEligible(kind, outcome string) bool {
	return kind == KindAgentTask && (outcome == OutcomeFailed || outcome == "failure")
}

// StructureFailure records the observed failure without inferring a general
// recommendation from a task's private conversation. The candidate remains a
// condition for human review, not an automatic avoid verdict.
func StructureFailure(kind, outcome string) (pattern, verdict, condition string, err error) {
	if !CaptureEligible(kind, outcome) {
		return "", "", "", ErrInvalid
	}
	return kind, VerdictCondition, "outcome=" + outcome, nil
}

// ValidateCapture validates a candidate's core fields (§4.3/§4.5). A verdict of
// condition requires a non-empty trigger condition; every candidate must carry
// at least one evidence event so the evidence chain stays traceable.
func ValidateCapture(pattern, verdict, condition string, evidenceCount int) error {
	if strings.TrimSpace(pattern) == "" {
		return ErrInvalid
	}
	if !ValidVerdict(verdict) {
		return ErrInvalid
	}
	if verdict == VerdictCondition && strings.TrimSpace(condition) == "" {
		return ErrInvalid
	}
	if evidenceCount < 1 {
		return ErrEvidence
	}
	return nil
}

// Promote returns the next status for a human-gate action, or ErrInvalid for
// any transition outside the quarantine↔approved chain.
func Promote(status, action string) (string, error) {
	if !ValidStatus(status) {
		return "", ErrInvalid
	}
	switch action {
	case ActionApprove:
		if status == StatusQuarantine {
			return StatusApproved, nil
		}
	case ActionDemote:
		if status == StatusApproved {
			return StatusQuarantine, nil
		}
	}
	return "", ErrInvalid
}

// ShareRecord is the cross-project share audit entry: who shared, when, and to
// which project (§plan 决定4).
type ShareRecord struct {
	SharedBy        string    // who
	SharedToProject string    // where
	SharedAt        time.Time // when
}

// ValidateShare requires a complete audit entry and rejects a share back to the
// owning project, which is not a cross-project share.
func ValidateShare(s ShareRecord, ownerProject string) error {
	if strings.TrimSpace(s.SharedBy) == "" || strings.TrimSpace(s.SharedToProject) == "" || s.SharedAt.IsZero() {
		return ErrInvalid
	}
	if s.SharedToProject == ownerProject {
		return ErrInvalid
	}
	return nil
}

// ScopedForProject reports whether an experience owned by ownerProject is
// visible to queryProject: same project, or explicitly shared there. Default
// isolation means no other project sees it.
func ScopedForProject(ownerProject string, shares []ShareRecord, queryProject string) bool {
	if ownerProject == queryProject {
		return true
	}
	for _, s := range shares {
		if s.SharedToProject == queryProject {
			return true
		}
	}
	return false
}
