package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Nil policy values remain undecided. Operations requiring them return a
// configuration error; the task book's fixed constants live in domain.
type DynamicsOptions struct {
	Decay                   *float64
	InitialConfidence       *float64
	SuccessConfidenceDelta  *float64
	ConfidenceDropThreshold *float64
	DensityThreshold        *float64
	SplitThreshold          *float64
	MergeDurationMS         *int64
	MergeObservations       *int
	Clock                   func() time.Time
}

type DynamicsEngine struct {
	store   DynamicsStore
	options DynamicsOptions
}

func dynamicsID() string { return rand.Text() }

var _ Engine = (*DynamicsEngine)(nil)
var _ ConflictGateService = (*DynamicsEngine)(nil)

func NewDynamicsEngine(store DynamicsStore, options DynamicsOptions) *DynamicsEngine {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &DynamicsEngine{store: store, options: options}
}

func dynamicsPolicy(name string) error {
	return serviceError(apierrors.ValidationFailed, "Dynamics policy remains undecided: "+name, "configure_dynamics_policy")
}
func dynamicsInvalid(message string) error {
	return serviceError(apierrors.ValidationFailed, message, "correct_dynamics_input")
}
func dynamicsNotFound(err error) bool {
	var e *apierrors.ServiceError
	return errors.As(err, &e) && e.Code == apierrors.NotFound
}
func dynamicsFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func dynamicsSignalKey(s domain.Signal) string {
	return fmt.Sprintf("%q/%q/%q", s.EventID, s.Topic, s.Class)
}

func (e *DynamicsEngine) seedInTx(tx DynamicsTx, topic string, now int64, kind string) (StoredAttentionNode, error) {
	if topic == "" || now < 0 {
		return StoredAttentionNode{}, dynamicsInvalid("Topic and nonnegative time are required")
	}
	if e.options.Decay == nil || !dynamicsFinite(*e.options.Decay) || *e.options.Decay < 0 {
		return StoredAttentionNode{}, dynamicsPolicy("decay")
	}
	if e.options.InitialConfidence == nil || !dynamicsFinite(*e.options.InitialConfidence) || *e.options.InitialConfidence < 0 || *e.options.InitialConfidence > 1 {
		return StoredAttentionNode{}, dynamicsPolicy("initial confidence")
	}
	n, err := tx.GetNode(topic)
	if err == nil {
		return n, nil
	}
	if !dynamicsNotFound(err) {
		return n, err
	}
	n = StoredAttentionNode{Node: domain.AttentionNode{ID: dynamicsID(), Topic: topic, Attention: domain.AttentionFloor, Decay: *e.options.Decay, Confidence: *e.options.InitialConfidence, Evidence: []domain.EvidenceRef{}, RelatedProjects: []string{}, RelatedAgents: []string{}, RelatedTopics: map[string]float64{}, LastActive: now}, UpdatedAt: now, Lifecycle: "born"}
	h, _ := readDynamicsHistory(n)
	h.Transitions = append(h.Transitions, TopicTransition{Kind: kind, TS: now})
	err = saveDynamicsNode(tx, &n, h)
	return n, err
}

func (e *DynamicsEngine) SeedTopic(ctx context.Context, p Principal, topic string, now int64) (domain.AttentionNode, error) {
	if err := authorize(p, true); err != nil {
		return domain.AttentionNode{}, err
	}
	var n StoredAttentionNode
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		var err error
		n, err = e.seedInTx(tx, topic, now, "manual_seed")
		return err
	})
	return n.Node, err
}

func (e *DynamicsEngine) ApplySignals(ctx context.Context, topic string, signals []domain.Signal, now int64) (domain.AttentionNode, error) {
	var n domain.AttentionNode
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		var err error
		n, err = e.ApplySignalsInTx(tx, topic, signals, now)
		return err
	})
	return n, err
}

// ApplySignalsInTx allows W3 to advance the same engine inside its atomic daily
// commit. Signals must already exist with a mapped raw receipt from W1.
func (e *DynamicsEngine) ApplySignalsInTx(tx DynamicsTx, topic string, signals []domain.Signal, now int64) (domain.AttentionNode, error) {
	n, err := tx.GetNode(topic)
	if err != nil {
		return n.Node, err
	}
	h, err := readDynamicsHistory(n)
	if err != nil {
		return n.Node, err
	}
	if len(h.RetiredTo) > 0 {
		return n.Node, dynamicsInvalid("Topic is retained as split/merge history; use its successor")
	}
	fresh := []domain.Signal{}
	receipts := map[string]StoredEvent{}
	seen := map[string]domain.Signal{}
	for _, s := range signals {
		if s.Topic != topic || s.EventID == "" || s.TS < 0 || s.TS > now || !dynamicsFinite(s.Strength) || s.Strength < 0 || !dynamicsFinite(s.Weight) || domain.SignalWeight(s.Class) == 0 || s.Weight != domain.SignalWeight(s.Class) {
			return n.Node, dynamicsInvalid("Signal must use its persisted topic, time, strength and task-book weight")
		}
		key := dynamicsSignalKey(s)
		if prev, ok := seen[key]; ok {
			if prev != s {
				return n.Node, dynamicsInvalid("Duplicate signal changed within batch")
			}
			continue
		}
		seen[key] = s
		if used, ok := h.Applied[key]; ok {
			if used.Signal != s {
				return n.Node, dynamicsInvalid("Applied signal cannot be changed")
			}
			continue
		}
		if s.TS < n.UpdatedAt {
			return n.Node, dynamicsInvalid("New signal predates persisted updated_at; replay must be chronological")
		}
		saved, err := tx.GetSignal(s)
		if err != nil {
			return n.Node, err
		}
		if saved != s {
			return n.Node, dynamicsInvalid("Signal differs from persisted normalization")
		}
		raw, err := tx.GetEvent(s.EventID)
		if err != nil {
			return n.Node, err
		}
		if raw.Event == nil || raw.Event.TS != s.TS || raw.Event.Topic != s.Topic {
			return n.Node, dynamicsInvalid("Signal has no matching mapped event")
		}
		if s.Class == domain.SignalFeedbackFailure && raw.Event.Kind != domain.EventKindAgentTask {
			return n.Node, dynamicsInvalid("Failure feedback requires an agent_task event")
		}
		receipts[s.EventID] = raw
		fresh = append(fresh, s)
	}
	if now < n.UpdatedAt {
		if len(signals) > 0 && len(fresh) == 0 {
			return n.Node, nil
		}
		return n.Node, dynamicsInvalid("Engine time predates updated_at")
	}
	if len(fresh) == 0 && now == n.UpdatedAt {
		return n.Node, nil
	}
	slices.SortFunc(fresh, func(a, b domain.Signal) int {
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
		return 0
	})
	for i := 0; i < len(fresh); {
		end := i + 1
		for end < len(fresh) && fresh[end].TS == fresh[i].TS {
			end++
		}
		dt := float64(fresh[i].TS-n.UpdatedAt) / 86400000
		prevA, prevC := n.Node.Attention, n.Node.Confidence
		sum, confidenceDelta := 0.0, 0.0
		attentionSignal := false
		for _, s := range fresh[i:end] {
			a, c := s.Strength*s.Weight, 0.0
			if s.Class == domain.SignalFeedbackFailure {
				c = a
				a = 0
			} else {
				attentionSignal = true
			}
			if s.Class == domain.SignalFeedbackSuccess {
				if e.options.SuccessConfidenceDelta == nil || !dynamicsFinite(*e.options.SuccessConfidenceDelta) || *e.options.SuccessConfidenceDelta < 0 {
					return n.Node, dynamicsPolicy("success confidence increment")
				}
				c = s.Strength * *e.options.SuccessConfidenceDelta
			}
			sum += a
			confidenceDelta += c
			h.Applied[dynamicsSignalKey(s)] = AppliedDynamicsSignal{Signal: s, ElapsedDays: dt, Decay: n.Node.Decay, DecayFactor: math.Exp(-n.Node.Decay * dt), AttentionContribution: a, ConfidenceContribution: c}
			n.Node.Evidence = append(n.Node.Evidence, domain.EvidenceRef{EventID: s.EventID, WeightAtUse: s.Weight, TS: s.TS})
			raw := receipts[s.EventID]
			if raw.Event.Scope != "" && !slices.Contains(n.Node.RelatedProjects, raw.Event.Scope) {
				n.Node.RelatedProjects = append(n.Node.RelatedProjects, raw.Event.Scope)
			}
			if raw.Event.Source == domain.EventSourceClaudeCode || raw.Event.Source == domain.EventSourceCodex {
				if !slices.Contains(n.Node.RelatedAgents, raw.Event.Source) {
					n.Node.RelatedAgents = append(n.Node.RelatedAgents, raw.Event.Source)
				}
			}
			if s.Class == domain.SignalExplicitPositive || s.Class == domain.SignalExplicitNegative {
				n.Node.EmotionalWeight += a
			}
		}
		if dt > 0 || attentionSignal {
			n.Node.Attention = domain.UpdateAttention(prevA, n.Node.Decay, dt, sum)
			n.Node.Momentum = domain.UpdateMomentum(n.Node.Momentum, prevA, n.Node.Attention)
		}
		n.Node.Confidence = domain.UpdateConfidence(prevC, confidenceDelta)
		if err = e.confidenceConflict(tx, n, h, prevC, fresh[i:end]); err != nil {
			return n.Node, err
		}
		n.UpdatedAt = fresh[i].TS
		n.Node.LastActive = fresh[i].TS
		if n.Lifecycle == "cold" {
			h.Transitions = append(h.Transitions, TopicTransition{Kind: "signal_awaken", TS: fresh[i].TS})
		}
		n.Lifecycle = "active"
		if h.MaxAttention < n.Node.Attention {
			h.MaxAttention = n.Node.Attention
		}
		i = end
	}
	if now > n.UpdatedAt {
		a := n.Node.Attention
		n.Node.Attention = domain.UpdateAttention(a, n.Node.Decay, float64(now-n.UpdatedAt)/86400000, 0)
		n.Node.Momentum = domain.UpdateMomentum(n.Node.Momentum, a, n.Node.Attention)
		n.UpdatedAt = now
	}
	if !domain.IsRising(n.Node.Momentum) && domain.ShouldArchive(n.Node.Attention, float64(now-n.Node.LastActive)/86400000) && n.Lifecycle != "cold" {
		n.Lifecycle = "cold"
		h.Transitions = append(h.Transitions, TopicTransition{Kind: "archive", TS: now})
	}
	if err = saveDynamicsNode(tx, &n, h); err != nil {
		return n.Node, err
	}
	return n.Node, nil
}

func (e *DynamicsEngine) confidenceConflict(tx DynamicsTx, n StoredAttentionNode, h DynamicsHistory, previous float64, signals []domain.Signal) error {
	delta := n.Node.Confidence - previous
	if delta >= 0 {
		return nil
	}
	if e.options.ConfidenceDropThreshold == nil || !dynamicsFinite(*e.options.ConfidenceDropThreshold) || *e.options.ConfidenceDropThreshold < 0 {
		return dynamicsPolicy("confidence drop threshold")
	}
	if !domain.ShouldConflictConfidenceDrop(delta, *e.options.ConfidenceDropThreshold) {
		return nil
	}
	right := signals[len(signals)-1]
	left := n.Node.ID
	for _, r := range n.Node.Evidence {
		if !slices.ContainsFunc(signals, func(s domain.Signal) bool { return s.EventID == r.EventID }) {
			left = r.EventID
		}
	}
	detail := fmt.Sprintf("confidence_drop: previous=%g current=%g threshold=%g ts=%d", previous, n.Node.Confidence, *e.options.ConfidenceDropThreshold, right.TS)
	return tx.SaveConflict(DynamicsConflict{Conflict: domain.Conflict{ID: dynamicsID(), Topic: n.Node.Topic, Kind: domain.ConflictKindAttentionNode, LeftRef: left, RightRef: right.EventID, Detail: detail, Status: domain.ConflictStatusPending}, CreatedAt: right.TS, Version: 1})
}
