package mcp

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Event exposure deliberately has no original bytes, source or session. Keep
// this adapter whitelist separate from StoredEvent and the core explanation.
type exposedEvent struct {
	EventID string `json:"event_id"`
	Kind    string `json:"kind"`
	Topic   string `json:"topic"`
	Tokens  *int   `json:"tokens"`
	Scope   string `json:"scope"`
	TS      *int64 `json:"ts"`
}

func exposeEvent(row app.StoredEvent) exposedEvent {
	out := exposedEvent{EventID: row.ID}
	if row.Event != nil {
		event := row.Event
		out.Kind, out.Topic, out.Scope = event.Kind, event.Topic, event.Scope
		tokens, ts := event.Tokens, event.TS
		out.Tokens, out.TS = &tokens, &ts
	}
	return out
}

type exposedSignal struct {
	Event                  exposedEvent       `json:"event"`
	EventID                string             `json:"event_id"`
	Kind                   string             `json:"kind"`
	Class                  domain.SignalClass `json:"class"`
	Weight                 float64            `json:"weight"`
	TS                     int64              `json:"ts"`
	Strength               float64            `json:"strength"`
	ElapsedDays            float64            `json:"elapsed_days"`
	Decay                  float64            `json:"decay"`
	DecayFactor            float64            `json:"decay_factor"`
	AttentionContribution  float64            `json:"attention_contribution"`
	ConfidenceContribution float64            `json:"confidence_contribution"`
	ConfidenceOnly         bool               `json:"confidence_only"`
}

type exposedExplanation struct {
	Topic           string            `json:"topic"`
	Attention       float64           `json:"attention"`
	Momentum        float64           `json:"momentum"`
	Evidence        []exposedSignal   `json:"evidence"`
	Related         []app.RelatedNode `json:"related"`
	RelatedProjects []string          `json:"related_projects"`
	UpdatedAt       int64             `json:"updated_at"`
}

func exposeExplanation(result app.ExplainFocusResult) exposedExplanation {
	out := exposedExplanation{Topic: result.Topic, Attention: result.Attention, Momentum: result.Momentum,
		Related: result.Related, RelatedProjects: result.RelatedProjects, UpdatedAt: result.UpdatedAt}
	if result.Evidence != nil {
		out.Evidence = make([]exposedSignal, 0, len(result.Evidence))
	}
	for _, signal := range result.Evidence {
		out.Evidence = append(out.Evidence, exposedSignal{
			Event: exposeEvent(signal.Original), EventID: signal.EventID, Kind: signal.Kind, Class: signal.Class,
			Weight: signal.Weight, TS: signal.TS, Strength: signal.Strength, ElapsedDays: signal.ElapsedDays,
			Decay: signal.Decay, DecayFactor: signal.DecayFactor, AttentionContribution: signal.AttentionContribution,
			ConfidenceContribution: signal.ConfidenceContribution, ConfidenceOnly: signal.ConfidenceOnly,
		})
	}
	return out
}

// Human-approved structured experience is a separate derived capability.
// Internal approval/share audit actors and request metadata are not exposure.
type exposedExperience struct {
	Experience    domain.Experience `json:"experience"`
	OwnerProject  string            `json:"owner_project"`
	OriginEventID string            `json:"origin_event_id"`
	Version       int               `json:"version"`
	CreatedAt     int64             `json:"created_at"`
	UpdatedAt     int64             `json:"updated_at"`
}

func exposeExperiences(result app.GetExperiencesResult) any {
	var rows []exposedExperience
	if result.Experiences != nil {
		rows = make([]exposedExperience, 0, len(result.Experiences))
	}
	for _, row := range result.Experiences {
		rows = append(rows, exposedExperience{Experience: row.Experience, OwnerProject: row.OwnerProject,
			OriginEventID: row.OriginEventID, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return struct {
		Experiences []exposedExperience `json:"experiences"`
	}{rows}
}

type exposedDelivery struct {
	EventID         string   `json:"event_id,omitempty"`
	DeliveryState   string   `json:"delivery_state,omitempty"`
	CapturePending  bool     `json:"capture_pending"`
	DynamicsPending bool     `json:"dynamics_pending"`
	ReceiptIDs      []string `json:"receipt_ids,omitempty"`
	Recovery        string   `json:"recovery,omitempty"`
}

func (d exposedDelivery) MCPDeliveryState() string { return d.DeliveryState }

func exposeDelivery(delivery app.MCPReceiptDelivery) exposedDelivery {
	out := exposedDelivery{EventID: delivery.EventID, DeliveryState: delivery.DeliveryState,
		CapturePending: delivery.CapturePending, DynamicsPending: delivery.DynamicsPending, ReceiptIDs: delivery.ReceiptIDs}
	if delivery.Recovery != "" {
		// Do not interpolate backend text or expose the session-bearing origin.
		out.Recovery = "reconcile event_id and receipt_ids; if delivery is unknown, request local reconciliation; do not resubmit with a new origin key"
	}
	return out
}

type exposedReport struct {
	Accepted bool `json:"accepted,omitempty"`
	exposedDelivery
}

type exposedVeto struct {
	Acknowledged bool   `json:"acknowledged,omitempty"`
	Target       string `json:"target"`
	Pending      bool   `json:"pending"`
	exposedDelivery
}
