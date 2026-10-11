package app

import (
	"context"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (e *DynamicsEngine) StoredNode(ctx context.Context, topic string) (StoredAttentionNode, error) {
	var n StoredAttentionNode
	err := e.store.ViewDynamics(ctx, func(tx DynamicsTx) error { var err error; n, err = tx.GetNode(topic); return err })
	return n, err
}

func (e *DynamicsEngine) AttentionNode(ctx context.Context, topic string) (domain.AttentionNode, error) {
	n, err := e.StoredNode(ctx, topic)
	return n.Node, err
}

// AttentionState reads persisted state without advancing time or touching A.
// Cold/retired history is retained in storage but is not a current focus.
func (e *DynamicsEngine) AttentionState(ctx context.Context) ([]domain.AttentionNode, error) {
	result := []domain.AttentionNode{}
	err := e.store.ViewDynamics(ctx, func(tx DynamicsTx) error {
		nodes, err := tx.ListNodes()
		if err != nil {
			return err
		}
		for _, n := range nodes {
			if n.Lifecycle != "cold" {
				result = append(result, n.Node)
			}
		}
		return nil
	})
	slices.SortFunc(result, func(a, b domain.AttentionNode) int {
		if a.Attention > b.Attention {
			return -1
		}
		if a.Attention < b.Attention {
			return 1
		}
		if a.Topic < b.Topic {
			return -1
		}
		if a.Topic > b.Topic {
			return 1
		}
		return 0
	})
	return result, err
}

func (e *DynamicsEngine) ExplainFocus(ctx context.Context, topic string) (FocusExplanation, error) {
	var result FocusExplanation
	err := e.store.ViewDynamics(ctx, func(tx DynamicsTx) error {
		n, err := tx.GetNode(topic)
		if err != nil {
			return err
		}
		h, err := readDynamicsHistory(n)
		if err != nil {
			return err
		}
		result = FocusExplanation{Topic: topic, Attention: n.Node.Attention, Momentum: n.Node.Momentum, UpdatedAt: n.UpdatedAt, Evidence: []ExplanationSignal{}, Related: []RelatedNode{}, RelatedProjects: n.Node.RelatedProjects, RelatedAgents: n.Node.RelatedAgents}
		for _, used := range h.Applied {
			raw, err := tx.GetEvent(used.Signal.EventID)
			if err != nil {
				return err
			}
			if raw.Event == nil {
				return dynamicsInvalid("Explanation evidence is not a mapped event")
			}
			s := used.Signal
			result.Evidence = append(result.Evidence, ExplanationSignal{Original: raw, EventID: s.EventID, Kind: raw.Event.Kind, Class: s.Class, Weight: s.Weight, TS: s.TS, Strength: s.Strength, ElapsedDays: used.ElapsedDays, Decay: used.Decay, DecayFactor: used.DecayFactor, AttentionContribution: used.AttentionContribution, ConfidenceContribution: used.ConfidenceContribution, ConfidenceOnly: s.Class == domain.SignalFeedbackFailure})
		}
		// Clustering evidence can be attached without creating a behaviour Signal.
		for _, ref := range n.Node.Evidence {
			if slices.ContainsFunc(result.Evidence, func(s ExplanationSignal) bool { return s.EventID == ref.EventID }) {
				continue
			}
			raw, err := tx.GetEvent(ref.EventID)
			if err != nil {
				return err
			}
			if raw.Event == nil {
				return dynamicsInvalid("Clustering evidence is not a mapped event")
			}
			result.Evidence = append(result.Evidence, ExplanationSignal{Original: raw, EventID: ref.EventID, Kind: raw.Event.Kind, TS: ref.TS, Weight: ref.WeightAtUse})
		}
		for topic, strength := range n.Node.RelatedTopics {
			result.Related = append(result.Related, RelatedNode{Topic: topic, Strength: strength, TS: h.RelatedAt[topic]})
		}
		return nil
	})
	slices.SortFunc(result.Evidence, func(a, b ExplanationSignal) int {
		if a.TS < b.TS {
			return -1
		}
		if a.TS > b.TS {
			return 1
		}
		if a.EventID < b.EventID {
			return -1
		}
		if a.EventID > b.EventID {
			return 1
		}
		if a.Class < b.Class {
			return -1
		}
		if a.Class > b.Class {
			return 1
		}
		return 0
	})
	slices.SortFunc(result.Related, func(a, b RelatedNode) int {
		if a.Topic < b.Topic {
			return -1
		}
		if a.Topic > b.Topic {
			return 1
		}
		return 0
	})
	return result, err
}
