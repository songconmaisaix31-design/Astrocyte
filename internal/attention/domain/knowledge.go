package domain

import (
	"encoding/json"
	"slices"
	"time"
)

// KnowledgeStatus is the admission state of a KnowledgeEntry (§4.4). An entry
// is captured into inbox and becomes referenceable by a Project Workspace only
// after a human confirms it.
type KnowledgeStatus string

const (
	KnowledgeInbox     KnowledgeStatus = "inbox"
	KnowledgeConfirmed KnowledgeStatus = "confirmed"
)

// ConfirmKnowledge advances inbox → confirmed. Confirmed is idempotent; any
// other status is rejected rather than silently promoted.
func ConfirmKnowledge(status KnowledgeStatus) (KnowledgeStatus, error) {
	switch status {
	case KnowledgeInbox, KnowledgeConfirmed:
		return KnowledgeConfirmed, nil
	default:
		return "", ErrInvalid
	}
}

// ReferencableKnowledge reports whether a Project Workspace may reference the
// entry: only confirmed entries qualify.
func ReferencableKnowledge(status KnowledgeStatus) bool {
	return status == KnowledgeConfirmed
}

// CrystallizedVersion is one immutable snapshot of structured facts. The
// crystallized layer is append-only: each sleep consolidation appends a new
// timestamped version instead of overwriting, so fact-evolution history is
// never lost (Zep temporal graph, simplified). The fluid layer, by contrast, is
// a single mutable summary with no history.
type CrystallizedVersion struct {
	Content    json.RawMessage `json:"content"`
	RecordedAt time.Time       `json:"recorded_at"`
}

// AppendCrystallized returns the history with a new version appended. Existing
// versions are never mutated, and a version whose timestamp does not strictly
// follow the previous one is rejected so fact-evolution order is preserved.
func AppendCrystallized(history []CrystallizedVersion, content json.RawMessage, at time.Time) ([]CrystallizedVersion, error) {
	if len(content) == 0 {
		return nil, ErrEvidence
	}
	if len(history) > 0 && !at.After(history[len(history)-1].RecordedAt) {
		return nil, ErrVersion
	}
	out := make([]CrystallizedVersion, len(history)+1)
	copy(out, history)
	out[len(history)] = CrystallizedVersion{Content: content, RecordedAt: at}
	return out, nil
}

// KnowledgeVisible reports whether an entry scoped to ownerProject is visible
// from project. Default isolation: only the owner sees it. An explicit
// cross-project share (a human action, audited separately) widens visibility.
func KnowledgeVisible(ownerProject string, sharedProjects []string, project string) bool {
	if ownerProject == project {
		return true
	}
	return slices.Contains(sharedProjects, project)
}
