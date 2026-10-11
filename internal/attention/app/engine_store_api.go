package app

import (
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// DynamicsStore keeps evidence checks, idempotence and state updates in one
// transaction. W3 may compose DynamicsTx with its daily transaction.
type DynamicsStore interface {
	WithDynamicsTx(context.Context, func(DynamicsTx) error) error
	ViewDynamics(context.Context, func(DynamicsTx) error) error
}

type DynamicsTx interface {
	GetNode(string) (StoredAttentionNode, error)
	ListNodes() ([]StoredAttentionNode, error)
	SaveNode(StoredAttentionNode) error
	GetEvent(string) (StoredEvent, error)
	GetSignal(domain.Signal) (domain.Signal, error)
	GetConflict(string) (DynamicsConflict, error)
	ListConflicts() ([]DynamicsConflict, error)
	SaveConflict(DynamicsConflict) error
	AddConflictAction(string, ActionAudit) error
}

type DynamicsConflict struct {
	Conflict  domain.Conflict `json:"conflict"`
	CreatedAt int64           `json:"created_at"`
	Version   int             `json:"version"`
	Actions   []ActionAudit   `json:"actions"`
}

type AppliedDynamicsSignal struct {
	Signal                 domain.Signal `json:"signal"`
	ElapsedDays            float64       `json:"elapsed_days"`
	Decay                  float64       `json:"decay"`
	DecayFactor            float64       `json:"decay_factor"`
	AttentionContribution  float64       `json:"attention_contribution"`
	ConfidenceContribution float64       `json:"confidence_contribution"`
}

type TopicTransition struct {
	Kind   string   `json:"kind"`
	TS     int64    `json:"ts"`
	Topics []string `json:"topics,omitempty"`
}

// DynamicsHistory is stored in M0's History extension, never in raw evidence.
type DynamicsHistory struct {
	MaxAttention float64                          `json:"max_attention"`
	Applied      map[string]AppliedDynamicsSignal `json:"applied"`
	Associations map[string]domain.AssociationRun `json:"associations"`
	RelatedAt    map[string]int64                 `json:"related_at"`
	Transitions  []TopicTransition                `json:"transitions"`
	RetiredTo    []string                         `json:"retired_to,omitempty"`
}

func readDynamicsHistory(n StoredAttentionNode) (DynamicsHistory, error) {
	h := DynamicsHistory{}
	if len(n.History) > 0 {
		if err := json.Unmarshal(n.History, &h); err != nil {
			return h, err
		}
	}
	if h.Applied == nil {
		h.Applied = map[string]AppliedDynamicsSignal{}
	}
	if h.Associations == nil {
		h.Associations = map[string]domain.AssociationRun{}
	}
	if h.RelatedAt == nil {
		h.RelatedAt = map[string]int64{}
	}
	if h.MaxAttention < n.Node.Attention {
		h.MaxAttention = n.Node.Attention
	}
	return h, nil
}

func saveDynamicsNode(tx DynamicsTx, n *StoredAttentionNode, h DynamicsHistory) error {
	if h.MaxAttention < n.Node.Attention {
		h.MaxAttention = n.Node.Attention
	}
	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	n.History = b
	n.Version++
	return tx.SaveNode(*n)
}
