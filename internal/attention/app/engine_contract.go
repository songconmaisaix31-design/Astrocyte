package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// This file locks the M0 engine boundary and the manual human gates. The
// dynamics implementation (W2), the knowledge/experience/conflict stores
// (W4/W5/W2) and the MCP read side (W6) consume these contracts; nothing here
// performs I/O or mutates state.

// Engine is the Attention Dynamics boundary (task book §5). It owns state
// evolution and the read side. Retrieval is a strictly separate consumer and
// has no write path into A(t) (§5.6).
type Engine interface {
	// ApplySignals folds one batch of weighted signals into a topic and returns
	// the updated node (§5.1).
	ApplySignals(ctx context.Context, topic string, signals []domain.Signal, now int64) (domain.AttentionNode, error)
	// AttentionState returns the current focus ordering (topic, attention,
	// momentum) for the read side and for get_attention_state (§7).
	AttentionState(ctx context.Context) ([]domain.AttentionNode, error)
	// AttentionNode returns one topic's current state.
	AttentionNode(ctx context.Context, topic string) (domain.AttentionNode, error)
}

// FocusExplanation is the evidence chain behind one topic's attention
// (§7 explain_focus). An unexplainable attention output does not exist.
type FocusExplanation struct {
	Topic     string              `json:"topic"`
	Attention float64             `json:"attention"`
	Momentum  float64             `json:"momentum"`
	Evidence  []ExplanationSignal `json:"evidence"`
	Related   []RelatedNode       `json:"related"`
}

// ExplanationSignal is one weighted signal in a focus explanation, timestamped
// so the chain is auditable end to end.
type ExplanationSignal struct {
	EventID string             `json:"event_id"`
	Kind    string             `json:"kind"`
	Class   domain.SignalClass `json:"class"`
	Weight  float64            `json:"weight"`
	TS      int64              `json:"ts"`
}

// RelatedNode is an adjacent topic and its association strength.
type RelatedNode struct {
	Topic    string  `json:"topic"`
	Strength float64 `json:"strength"`
}

// ConfirmKnowledgeCommand confirms a KnowledgeEntry (inbox → confirmed) so a
// Project Workspace may reference it (§4.4). Only the human gate may do this.
type ConfirmKnowledgeCommand struct {
	CommandMeta
}

// KnowledgeEntryResult returns the confirmed entry.
type KnowledgeEntryResult struct {
	SchemaVersion int                   `json:"schema_version"`
	Entry         domain.KnowledgeEntry `json:"entry"`
}

// ApproveExperienceCommand approves an Experience (quarantine → approved) so it
// becomes retrievable (§4.3). Only the human gate may do this.
type ApproveExperienceCommand struct {
	CommandMeta
}

// ExperienceResult returns the approved experience.
type ExperienceResult struct {
	SchemaVersion int               `json:"schema_version"`
	Experience    domain.Experience `json:"experience"`
}

// Conflict decision values (§5.5): keep the left/right evidence, merge, or dismiss.
const (
	ConflictDecisionKeepLeft  = "keep_left"
	ConflictDecisionKeepRight = "keep_right"
	ConflictDecisionMerge     = "merge"
	ConflictDecisionDismiss   = "dismiss"
)

// AdjudicateConflictCommand records the human decision on a pending Conflict
// (§5.5). The decision is a value, never an automatic merge.
type AdjudicateConflictCommand struct {
	CommandMeta
	Decision string `json:"decision"`
}

// ConflictResult returns the adjudicated conflict.
type ConflictResult struct {
	SchemaVersion int             `json:"schema_version"`
	Conflict      domain.Conflict `json:"conflict"`
}

// KnowledgeGateService is the human gate for knowledge confirmation.
type KnowledgeGateService interface {
	ConfirmKnowledge(ctx context.Context, p Principal, id string, cmd ConfirmKnowledgeCommand) (KnowledgeEntryResult, error)
}

// ExperienceGateService is the human gate for experience approval.
type ExperienceGateService interface {
	ApproveExperience(ctx context.Context, p Principal, id string, cmd ApproveExperienceCommand) (ExperienceResult, error)
}

// ConflictGateService is the human gate for conflict adjudication.
type ConflictGateService interface {
	AdjudicateConflict(ctx context.Context, p Principal, id string, cmd AdjudicateConflictCommand) (ConflictResult, error)
}

// EngineGates bundles the three manual gates so the HTTP layer can expose them
// through one optional service. It is assembled only once W4/W5/W2 stores exist.
type EngineGates interface {
	KnowledgeGateService
	ExperienceGateService
	ConflictGateService
}
