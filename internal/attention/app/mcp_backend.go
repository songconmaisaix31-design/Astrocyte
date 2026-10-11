package app

import (
	"context"
	"encoding/json"
)

// MCPBackend answers the five cognitive-interface tools (task book §7). It is
// the seam where the stdio MCP adapter meets the attention/experience read-side
// use cases: the concrete implementation is wired once M0 publishes the engine
// and experience contracts, mapping their types onto the DTOs below. The
// interface itself is fixed by §7 and does not change when the backend lands.
type MCPBackend interface {
	GetAttentionState(context.Context) (GetAttentionStateResult, error)
	ExplainFocus(context.Context, ExplainFocusInput) (ExplainFocusResult, error)
	GetExperiences(context.Context, GetExperiencesInput) (GetExperiencesResult, error)
	ReportEvent(context.Context, ReportEventInput) (ReportEventResult, error)
	Veto(context.Context, VetoInput) (VetoResult, error)
}

// FocusItem is one ranked entry of the current attention state.
type FocusItem struct {
	Topic     string  `json:"topic"`
	Attention float64 `json:"attention"`
	Momentum  float64 `json:"momentum"`
}

type GetAttentionStateResult struct {
	Topics []FocusItem `json:"topics"`
}

type ExplainFocusInput struct {
	Topic string `json:"topic"`
}

// EvidenceItem is an EvidenceRef view of one event: which event, which kind of
// signal, its weight in the attention path and its timestamp.
type EvidenceItem struct {
	EventID string  `json:"event_id"`
	Kind    string  `json:"kind"`
	Source  string  `json:"source"`
	Weight  float64 `json:"weight"`
	TS      int64   `json:"ts"`
}

// RelatedTopic is one adjacent node and its association strength (MAGMA-style).
type RelatedTopic struct {
	Topic    string  `json:"topic"`
	Strength float64 `json:"strength"`
}

type ExplainFocusResult struct {
	Topic     string         `json:"topic"`
	Attention float64        `json:"attention"`
	Momentum  float64        `json:"momentum"`
	Evidence  []EvidenceItem `json:"evidence"`
	Related   []RelatedTopic `json:"related_topics"`
	Projects  []string       `json:"projects"`
}

type GetExperiencesInput struct {
	Pattern string `json:"pattern"`
	Scope   string `json:"scope,omitempty"`
}

// ExperienceItem is an approved experience, its verdict and the project scope
// it is visible in.
type ExperienceItem struct {
	ID        string   `json:"id"`
	Pattern   string   `json:"pattern"`
	Verdict   string   `json:"verdict"`
	Condition string   `json:"condition"`
	Scope     string   `json:"scope"`
	Evidence  []string `json:"evidence"`
}

type GetExperiencesResult struct {
	Experiences []ExperienceItem `json:"experiences"`
}

// ReportEventInput mirrors the unified Event contract (task book §4.1) as the
// MCP wire shape for external agents reporting an event.
type ReportEventInput struct {
	Source  string          `json:"source"`
	Session string          `json:"session"`
	TS      int64           `json:"ts"`
	Kind    string          `json:"kind"`
	Tokens  int             `json:"tokens"`
	Scope   string          `json:"scope"`
	Topic   string          `json:"topic"`
	Raw     json.RawMessage `json:"raw,omitempty"`
}

type ReportEventResult struct {
	Accepted bool   `json:"accepted"`
	EventID  string `json:"event_id"`
}

// VetoInput carries a user veto of a topic or a task.
type VetoInput struct {
	Target string `json:"target"`
	Reason string `json:"reason,omitempty"`
}

type VetoResult struct {
	Acknowledged bool   `json:"acknowledged"`
	Target       string `json:"target"`
}
