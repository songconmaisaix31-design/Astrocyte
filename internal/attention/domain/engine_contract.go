package domain

import "encoding/json"

// This file locks the M0 cognitive-core contract (task book §4 + §5.2). Every
// type here is a pure value: no clock, no storage, no I/O. The engine interface
// that consumes these values lives in the app package (see engine_contract.go).

// Event is the unified event model (§4.1). Every observed behaviour, from every
// source, is normalised into this one shape before attention dynamics can see it.
type Event struct {
	// Source is the emitting origin: claude_code | codex | manual | mcp.
	Source string `json:"source"`
	// Session identifies the originating session within that source.
	Session string `json:"session"`
	// TS is the event time in milliseconds since the Unix epoch.
	TS int64 `json:"ts"`
	// Kind is the behavioural verb: search | edit_file | read_paper |
	// watch_video | bookmark | commit | agent_task | user_approve |
	// user_veto | collect.
	Kind string `json:"kind"`
	// Tokens is the token volume consumed, when the source reports one.
	Tokens int `json:"tokens"`
	// Scope is the project or path the behaviour belongs to.
	Scope string `json:"scope"`
	// Topic is the cognitive topic label; its value domain is the Attention Graph.
	Topic string `json:"topic"`
	// Raw is the original payload, retained unmodified for audit and evidence.
	Raw json.RawMessage `json:"raw"`
}

// Event kinds (§4.1). The value domain of Event.Kind is closed.
const (
	EventKindSearch      = "search"
	EventKindEditFile    = "edit_file"
	EventKindReadPaper   = "read_paper"
	EventKindWatchVideo  = "watch_video"
	EventKindBookmark    = "bookmark"
	EventKindCommit      = "commit"
	EventKindAgentTask   = "agent_task"
	EventKindUserApprove = "user_approve"
	EventKindUserVeto    = "user_veto"
	EventKindCollect     = "collect"
)

// Event sources (§4.1). The value domain of Event.Source is closed.
const (
	EventSourceClaudeCode = "claude_code"
	EventSourceCodex      = "codex"
	EventSourceManual     = "manual"
	EventSourceMCP        = "mcp"
)

// SignalClass is the closed set of signal categories from the weight table (§5.2).
type SignalClass string

// Signal is a normalised Event weighted for state update (§5.1). It carries the
// weight w_i assigned by the §5.2 table so the update A(t) += s_i * w_i is
// reproducible from the event stream alone.
type Signal struct {
	// EventID is the evidence trace back to the originating Event.
	EventID string `json:"event_id"`
	// Topic is the Attention Graph topic this signal feeds.
	Topic string `json:"topic"`
	// Class is the signal category; it determines the default weight.
	Class SignalClass `json:"class"`
	// Strength s_i is the signal intensity.
	Strength float64 `json:"strength"`
	// Weight w_i is the signal weight from the §5.2 table.
	Weight float64 `json:"weight"`
	// TS is the signal time in milliseconds since the Unix epoch.
	TS int64 `json:"ts"`
}

// Signal weight table (§5.2, hardcoded first version, calibrated offline only).
const (
	SignalImplicitDepth    SignalClass = "implicit_depth"    // video pause/replay/speed, long read: 1.0
	SignalImplicitRegular  SignalClass = "implicit_regular"  // edit file, commit, search: 0.6
	SignalImplicitLight    SignalClass = "implicit_light"    // bookmark, inbox: 0.3
	SignalExplicitPositive SignalClass = "explicit_positive" // user marked important: 1.5
	SignalExplicitNegative SignalClass = "explicit_negative" // user veto/reject: -0.8
	SignalFeedbackSuccess  SignalClass = "feedback_success"  // agent task closed successfully: 0.8
	SignalFeedbackFailure  SignalClass = "feedback_failure"  // agent task failed: -0.1 (lowers confidence only)
)

// signalWeights is the §5.2 hardcoded weight table. It is data, not logic:
// calibration happens offline and never from online learning.
var signalWeights = map[SignalClass]float64{
	SignalImplicitDepth:    1.0,
	SignalImplicitRegular:  0.6,
	SignalImplicitLight:    0.3,
	SignalExplicitPositive: 1.5,
	SignalExplicitNegative: -0.8,
	SignalFeedbackSuccess:  0.8,
	SignalFeedbackFailure:  -0.1,
}

// SignalWeight returns the §5.2 default weight for a signal class. An unknown
// class carries zero weight so it can never silently influence attention.
func SignalWeight(class SignalClass) float64 {
	return signalWeights[class]
}

// AttentionNode is the core cognitive object (§4.2). It is not a MemoryItem: it
// carries the state a topic needs to evolve and to be explained, nothing more.
type AttentionNode struct {
	ID              string             `json:"id"`
	Topic           string             `json:"topic"`
	Attention       float64            `json:"attention"`  // current strength [0,1]
	Momentum        float64            `json:"momentum"`   // rate of change; stable interest vs burst
	Confidence      float64            `json:"confidence"` // evidence sufficiency, orthogonal to attention
	Decay           float64            `json:"decay"`      // per-topic decay coefficient λ
	Evidence        []EvidenceRef      `json:"evidence"`
	RelatedProjects []string           `json:"related_projects"`
	RelatedAgents   []string           `json:"related_agents"`
	RelatedTopics   map[string]float64 `json:"related_topics"` // MAGMA-style adjacency strength
	LastActive      int64              `json:"last_active"`
	EmotionalWeight float64            `json:"emotional_weight"` // explicit user preference strength
}

// Experience is what an Agent learned about doing work (§4.3). It is not
// knowledge: it captures patterns and verdicts, not facts.
type Experience struct {
	ID        string     `json:"id"`
	Pattern   string     `json:"pattern"`   // e.g. "centralized scheduler"
	Verdict   string     `json:"verdict"`   // avoid | prefer | condition
	Condition string     `json:"condition"` // trigger condition, e.g. ">100 agents"
	Evidence  []EventRef `json:"evidence"`  // raw failure/success trace references
	Status    string     `json:"status"`    // quarantine | approved
}

// Experience verdicts (§4.3).
const (
	ExperienceVerdictAvoid     = "avoid"
	ExperienceVerdictPrefer    = "prefer"
	ExperienceVerdictCondition = "condition"
)

// Experience statuses (§4.3): quarantine → approved is the human gate.
const (
	ExperienceStatusQuarantine = "quarantine"
	ExperienceStatusApproved   = "approved"
)

// KnowledgeEntry is the Knowledge Graph carrier with two-level storage (§4.4).
type KnowledgeEntry struct {
	ID           string          `json:"id"`
	Topic        string          `json:"topic"` // hangs off an AttentionNode
	Fluid        string          `json:"fluid"`
	Crystallized json.RawMessage `json:"crystallized"`
	Sources      []EventRef      `json:"sources"`
	Status       KnowledgeStatus `json:"status"`
}

// KnowledgeStatus is the admission state of a KnowledgeEntry (§4.4). Only
// confirmed is referenceable by a Project Workspace. This is the single source
// for the knowledge status vocabulary; W4's wave-A knowledge.go must drop its
// duplicate KnowledgeStatus/KnowledgeInbox/KnowledgeConfirmed and use these.
type KnowledgeStatus string

const (
	KnowledgeInbox     KnowledgeStatus = "inbox"
	KnowledgeConfirmed KnowledgeStatus = "confirmed"
)

// EvidenceRef is Evidence, the view of an Event (§4.5). No independent object
// exists: the evidence chain always resolves back to the original event.
type EvidenceRef struct {
	EventID     string  `json:"event_id"`
	WeightAtUse float64 `json:"weight_at_use"`
	TS          int64   `json:"ts"`
}

// EventRef is a raw event-trace reference used by Experience and KnowledgeEntry
// (§4.3, §4.4). It is narrower than EvidenceRef: no weight is asserted.
type EventRef struct {
	EventID string `json:"event_id"`
	TS      int64  `json:"ts"`
}

// Conflict is the fact-conflict record (§5.5). Conflicting evidence is never
// silently merged; it becomes a Conflict for the sleep-cycle adjudication queue.
type Conflict struct {
	ID       string `json:"id"`
	Topic    string `json:"topic"`
	Kind     string `json:"kind"`     // attention_node | knowledge_entry
	LeftRef  string `json:"left_ref"` // existing evidence/entry id
	RightRef string `json:"right_ref"`
	Detail   string `json:"detail"`
	Status   string `json:"status"` // pending | resolved_keep_left | resolved_keep_right | resolved_merge | dismissed
}

// Conflict kinds (§5.5).
const (
	ConflictKindAttentionNode = "attention_node"
	ConflictKindKnowledge     = "knowledge_entry"
)

// Conflict statuses (§5.5).
const (
	ConflictStatusPending   = "pending"
	ConflictStatusKeepLeft  = "resolved_keep_left"
	ConflictStatusKeepRight = "resolved_keep_right"
	ConflictStatusMerge     = "resolved_merge"
	ConflictStatusDismissed = "dismissed"
)
