package app

import (
	"context"
	"encoding/json"
)

// MCPBackend connects the five cognitive tools to application use cases.
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

type EvidenceItem = ExplanationSignal
type RelatedTopic = RelatedNode
type ExplainFocusResult = FocusExplanation

type GetExperiencesInput struct {
	Pattern string `json:"pattern"`
	Scope   string `json:"scope,omitempty"`
}

type ExperienceItem = StoredExperience

type GetExperiencesResult struct {
	Experiences []ExperienceItem `json:"experiences"`
}

// ReportEventInput mirrors the unified Event contract (task book §4.1) as the
// MCP wire shape for external agents reporting an event.
type ReportEventInput struct {
	Source          string          `json:"source"`
	Session         string          `json:"session"`
	TS              int64           `json:"ts"`
	Kind            string          `json:"kind"`
	Tokens          *int            `json:"tokens,omitempty"`
	Scope           string          `json:"scope"`
	Topic           string          `json:"topic"`
	Raw             json.RawMessage `json:"raw,omitempty"`
	OriginalPayload json.RawMessage `json:"-"`
}

type ReportEventResult struct {
	Accepted bool `json:"accepted,omitempty"`
	MCPReceiptDelivery
}

// MCPReceiptDelivery preserves acknowledgement uncertainty without implying
// that an errored operation did not commit. Reconcile the origin/receipt;
// submitting a fresh origin can duplicate the original write.
type MCPReceiptDelivery struct {
	EventID        string   `json:"event_id,omitempty"`
	OriginKey      string   `json:"origin_key,omitempty"`
	DeliveryState  string   `json:"delivery_state,omitempty"`
	CapturePending bool     `json:"capture_pending"`
	ReceiptIDs     []string `json:"receipt_ids,omitempty"`
	Recovery       string   `json:"recovery,omitempty"`
}

func (d MCPReceiptDelivery) MCPDeliveryState() string { return d.DeliveryState }

// VetoInput carries a user veto of a topic or a task.
type VetoInput struct {
	Topic  string `json:"topic,omitempty"`
	Task   string `json:"task,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type VetoResult struct {
	Acknowledged bool   `json:"acknowledged,omitempty"`
	Target       string `json:"target"`
	Pending      bool   `json:"pending"`
	MCPReceiptDelivery
}
