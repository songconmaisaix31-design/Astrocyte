package app

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// mappedRefs rejects fabricated evidence; clustering attaches raw events with
// weight zero because it does not assert a normalised behaviour contribution.
func mappedRefs(tx DynamicsTx, ids []string, start, end int64, scope string) ([]domain.EvidenceRef, error) {
	refs := []domain.EvidenceRef{}
	for _, id := range ids {
		if slices.ContainsFunc(refs, func(r domain.EvidenceRef) bool { return r.EventID == id }) {
			continue
		}
		raw, err := tx.GetEvent(id)
		if err != nil {
			return nil, err
		}
		if raw.Event == nil || raw.Event.TS < start || raw.Event.TS > end || (scope != "" && raw.Event.Scope != scope) {
			return nil, dynamicsInvalid("Cluster requires mapped events in the declared scope/time window")
		}
		refs = append(refs, domain.EvidenceRef{EventID: id, TS: raw.Event.TS})
	}
	slices.SortFunc(refs, func(a, b domain.EvidenceRef) int {
		if a.TS < b.TS {
			return -1
		}
		if a.TS > b.TS {
			return 1
		}
		if a.EventID < b.EventID {
			return -1
		}
		return 1
	})
	return refs, nil
}

func (e *DynamicsEngine) BirthByDensity(ctx context.Context, topic, scope string, ids []string, start, end int64) (StoredAttentionNode, error) {
	var n StoredAttentionNode
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		if e.options.DensityThreshold == nil || !dynamicsFinite(*e.options.DensityThreshold) || *e.options.DensityThreshold < 0 {
			return dynamicsPolicy("density threshold")
		}
		if scope == "" || start < 0 || end <= start {
			return dynamicsInvalid("Density birth needs a project and nonempty time window")
		}
		refs, err := mappedRefs(tx, ids, start, end, scope)
		if err != nil {
			return err
		}
		if !domain.ShouldBirthByDensity(float64(len(refs)), float64(end-start)/86400000, *e.options.DensityThreshold) {
			return dynamicsInvalid("Event density did not cross configured birth threshold")
		}
		n, err = e.seedInTx(tx, topic, refs[0].TS, "density_birth")
		return err
	})
	return n, err
}

// ConsolidateTopicInTx is W3's graph operation. Existing A/M/updated_at are
// unchanged. Birth support can include adjacent cluster events; only this
// topic's own refs are attached. A raw isolated event cannot create a concept.
func (e *DynamicsEngine) ConsolidateTopicInTx(tx DynamicsTx, topic string, refs []domain.EvidenceRef, related map[string]float64, now int64, support ...domain.EventRef) (StoredAttentionNode, error) {
	ids := []string{}
	for _, ref := range refs {
		if ref.WeightAtUse != 0 {
			return StoredAttentionNode{}, dynamicsInvalid("Clustering references cannot invent signal weights")
		}
		ids = append(ids, ref.EventID)
	}
	verified, err := mappedRefs(tx, ids, 0, now, "")
	if err != nil {
		return StoredAttentionNode{}, err
	}
	n, err := tx.GetNode(topic)
	if dynamicsNotFound(err) {
		supportIDs := append([]string(nil), ids...)
		for _, ref := range support {
			supportIDs = append(supportIDs, ref.EventID)
		}
		birthRefs, err := mappedRefs(tx, supportIDs, 0, now, "")
		if err != nil {
			return n, err
		}
		if len(verified) == 0 || len(birthRefs) < 2 {
			return n, dynamicsInvalid("An isolated event cannot create a sleep concept")
		}
		first, err := tx.GetEvent(verified[0].EventID)
		if err != nil {
			return n, err
		}
		if first.Event.Scope == "" {
			return n, dynamicsInvalid("New sleep concepts require a known project scope")
		}
		for _, ref := range birthRefs {
			r, err := tx.GetEvent(ref.EventID)
			if err != nil {
				return n, err
			}
			if r.Event.Scope != first.Event.Scope {
				return n, dynamicsInvalid("Sleep birth evidence crosses project scopes")
			}
		}
		n, err = e.seedInTx(tx, topic, verified[0].TS, "sleep_birth")
		if err != nil {
			return n, err
		}
		n.Node.Attention = domain.AttentionFloor
		n.Node.LastActive = verified[len(verified)-1].TS
	} else if err != nil {
		return n, err
	}
	h, err := readDynamicsHistory(n)
	if err != nil {
		return n, err
	}
	if now < n.UpdatedAt || len(h.RetiredTo) > 0 {
		return n, dynamicsInvalid("Cannot consolidate stale or retired topic")
	}
	changed := false
	for _, ref := range verified {
		if !slices.ContainsFunc(n.Node.Evidence, func(old domain.EvidenceRef) bool { return old.EventID == ref.EventID }) {
			n.Node.Evidence = append(n.Node.Evidence, ref)
			changed = true
		}
		raw, err := tx.GetEvent(ref.EventID)
		if err != nil {
			return n, err
		}
		if raw.Event.Scope != "" && !slices.Contains(n.Node.RelatedProjects, raw.Event.Scope) {
			n.Node.RelatedProjects = append(n.Node.RelatedProjects, raw.Event.Scope)
			changed = true
		}
		if (raw.Event.Source == domain.EventSourceClaudeCode || raw.Event.Source == domain.EventSourceCodex) && !slices.Contains(n.Node.RelatedAgents, raw.Event.Source) {
			n.Node.RelatedAgents = append(n.Node.RelatedAgents, raw.Event.Source)
			changed = true
		}
	}
	if n.Node.RelatedTopics == nil {
		n.Node.RelatedTopics = map[string]float64{}
	}
	for topic, strength := range related {
		if topic == n.Node.Topic || !dynamicsFinite(strength) || strength < 0 || strength > 1 {
			return n, dynamicsInvalid("Invalid topic association")
		}
		if _, err := tx.GetNode(topic); err != nil {
			return n, err
		}
		if n.Node.RelatedTopics[topic] != strength || h.RelatedAt[topic] != now {
			n.Node.RelatedTopics[topic] = strength
			h.RelatedAt[topic] = now
			changed = true
		}
	}
	if !changed {
		return n, nil
	}
	h.Transitions = append(h.Transitions, TopicTransition{Kind: "consolidate", TS: now})
	err = saveDynamicsNode(tx, &n, h)
	return n, err
}

func (e *DynamicsEngine) AdvanceLifecycleInTx(tx DynamicsTx, now int64) error {
	nodes, err := tx.ListNodes()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		h, err := readDynamicsHistory(n)
		if err != nil {
			return err
		}
		if len(h.RetiredTo) > 0 {
			continue
		}
		if _, err = e.ApplySignalsInTx(tx, n.Node.Topic, nil, now); err != nil {
			return err
		}
	}
	return nil
}

func (e *DynamicsEngine) AdvanceLifecycle(ctx context.Context, now int64) error {
	return e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error { return e.AdvanceLifecycleInTx(tx, now) })
}

func (e *DynamicsEngine) ResurfaceInTx(tx DynamicsTx, now int64, rng *rand.Rand) ([]StoredAttentionNode, error) {
	if rng == nil {
		return nil, dynamicsInvalid("Resurfacing requires an injected random source")
	}
	nodes, err := tx.ListNodes()
	if err != nil {
		return nil, err
	}
	candidates := []string{}
	for _, n := range nodes {
		h, err := readDynamicsHistory(n)
		if err != nil {
			return nil, err
		}
		if len(h.RetiredTo) == 0 && domain.IsBottomedOut(n.Node.Attention) && !domain.IsRising(n.Node.Momentum) {
			candidates = append(candidates, n.Node.Topic)
		}
	}
	slices.Sort(candidates)
	chosen := domain.SelectResurface(candidates, rng, 3)
	result := []StoredAttentionNode{}
	for _, topic := range chosen {
		n, err := tx.GetNode(topic)
		if err != nil {
			return nil, err
		}
		if now < n.UpdatedAt {
			return nil, dynamicsInvalid("Resurface time predates updated_at")
		}
		h, err := readDynamicsHistory(n)
		if err != nil {
			return nil, err
		}
		if n.Lifecycle == "cold" {
			n.Node.Attention = domain.AwakenAttention()
			n.Node.Momentum = 0
			n.UpdatedAt = now
			n.Lifecycle = "active"
			// Resurface is not user activity. Keep last_active from the raw event.
			h.Transitions = append(h.Transitions, TopicTransition{Kind: "resurface_awaken", TS: now})
			if err = saveDynamicsNode(tx, &n, h); err != nil {
				return nil, err
			}
		}
		result = append(result, n)
	}
	return result, nil
}

func (e *DynamicsEngine) Resurface(ctx context.Context, now int64, rng *rand.Rand) ([]StoredAttentionNode, error) {
	var selected []StoredAttentionNode
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error { var err error; selected, err = e.ResurfaceInTx(tx, now, rng); return err })
	return selected, err
}

func cloneDynamicsNode(n StoredAttentionNode) StoredAttentionNode {
	// All fields are JSON values from the persisted contract.
	b, _ := json.Marshal(n)
	var copy StoredAttentionNode
	_ = json.Unmarshal(b, &copy)
	return copy
}

func (e *DynamicsEngine) SplitTopic(ctx context.Context, parent string, children [2]string, clusters [2][]string, correlation float64, now int64) ([]StoredAttentionNode, error) {
	result := []StoredAttentionNode{}
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		if e.options.SplitThreshold == nil || !dynamicsFinite(*e.options.SplitThreshold) {
			return dynamicsPolicy("split threshold")
		}
		if e.options.SplitState == nil {
			return dynamicsPolicy("split child numeric state")
		}
		if !dynamicsFinite(correlation) || !domain.ShouldSplit(correlation, *e.options.SplitThreshold) || children[0] == "" || children[1] == "" || children[0] == children[1] || children[0] == parent || children[1] == parent {
			return dynamicsInvalid("Split requires two distinct clusters below the configured correlation threshold")
		}
		n, err := tx.GetNode(parent)
		if err != nil {
			return err
		}
		h, err := readDynamicsHistory(n)
		if err != nil {
			return err
		}
		if slices.Equal(h.RetiredTo, children[:]) {
			for _, topic := range children {
				child, err := tx.GetNode(topic)
				if err != nil {
					return err
				}
				result = append(result, child)
			}
			return nil
		}
		if len(h.RetiredTo) > 0 || now < n.UpdatedAt {
			return dynamicsInvalid("Cannot split a retired or stale topic")
		}
		seen := map[string]bool{}
		for _, cluster := range clusters {
			if len(cluster) == 0 {
				return dynamicsInvalid("Split clusters must both contain evidence")
			}
			for _, id := range cluster {
				if seen[id] || !slices.ContainsFunc(n.Node.Evidence, func(ref domain.EvidenceRef) bool { return ref.EventID == id }) {
					return dynamicsInvalid("Split evidence must partition the parent without overlap")
				}
				seen[id] = true
			}
		}
		for _, ref := range n.Node.Evidence {
			if !seen[ref.EventID] {
				return dynamicsInvalid("Split omitted parent evidence")
			}
		}
		for _, topic := range children {
			if _, err := tx.GetNode(topic); err == nil {
				return dynamicsInvalid("Split child already exists")
			} else if !dynamicsNotFound(err) {
				return err
			}
		}
		for i, topic := range children {
			child := cloneDynamicsNode(n)
			state := e.options.SplitState(cloneDynamicsNode(n).Node, i)
			if !dynamicsFinite(state.Attention) || state.Attention < domain.AttentionFloor || state.Attention > domain.AttentionCeiling || !dynamicsFinite(state.Momentum) || !dynamicsFinite(state.Confidence) || state.Confidence < 0 || state.Confidence > 1 || !dynamicsFinite(state.Decay) || state.Decay < 0 {
				return dynamicsInvalid("Configured split state violates attention/confidence/decay bounds")
			}
			child.Node.Attention, child.Node.Momentum, child.Node.Confidence, child.Node.Decay = state.Attention, state.Momentum, state.Confidence, state.Decay
			child.Node.ID = dynamicsID()
			child.Node.Topic = topic
			child.Lifecycle = "active"
			child.Version = 0
			child.Node.Evidence = []domain.EvidenceRef{}
			ch, err := readDynamicsHistory(child)
			if err != nil {
				return err
			}
			ch.Applied = map[string]AppliedDynamicsSignal{}
			ch.RetiredTo = nil
			ch.Associations = map[string]domain.AssociationRun{}
			for _, ref := range n.Node.Evidence {
				if slices.Contains(clusters[i], ref.EventID) {
					child.Node.Evidence = append(child.Node.Evidence, ref)
				}
			}
			for key, used := range h.Applied {
				if slices.Contains(clusters[i], used.Signal.EventID) {
					ch.Applied[key] = used
				}
			}
			child.Node.RelatedTopics[parent] = 1
			child.Node.RelatedTopics[children[1-i]] = 1
			ch.RelatedAt[parent] = now
			ch.RelatedAt[children[1-i]] = now
			ch.Transitions = append(ch.Transitions, TopicTransition{Kind: "split_child", TS: now, Topics: []string{parent}})
			if err = saveDynamicsNode(tx, &child, ch); err != nil {
				return err
			}
			result = append(result, child)
		}
		h.RetiredTo = children[:]
		h.Transitions = append(h.Transitions, TopicTransition{Kind: "split", TS: now, Topics: children[:]})
		n.Lifecycle = "cold"
		return saveDynamicsNode(tx, &n, h)
	})
	return result, err
}

// ObserveAssociation persists the sustained-affinity gate. Evidence overlap is
// derived from stored event references, never asserted by an external caller.
func (e *DynamicsEngine) ObserveAssociation(ctx context.Context, left, right string, strength float64, now int64) (bool, error) {
	merged := false
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		var err error
		merged, err = e.ObserveAssociationInTx(tx, left, right, strength, now)
		return err
	})
	return merged, err
}

func (e *DynamicsEngine) ObserveAssociationInTx(tx DynamicsTx, left, right string, strength float64, now int64) (bool, error) {
	if e.options.MergeDurationMS == nil || *e.options.MergeDurationMS <= 0 || e.options.MergeObservations == nil || *e.options.MergeObservations < 2 {
		return false, dynamicsPolicy("sustained merge duration and observation count")
	}
	if left == right || !dynamicsFinite(strength) || strength < 0 || strength > 1 {
		return false, dynamicsInvalid("Invalid merge association")
	}
	// Keep a stable survivor regardless of the observer's ordering.
	if right < left {
		left, right = right, left
	}
	a, err := tx.GetNode(left)
	if err != nil {
		return false, err
	}
	b, err := tx.GetNode(right)
	if err != nil {
		return false, err
	}
	ah, err := readDynamicsHistory(a)
	if err != nil {
		return false, err
	}
	bh, err := readDynamicsHistory(b)
	if err != nil {
		return false, err
	}
	if slices.Equal(bh.RetiredTo, []string{left}) {
		return true, nil
	}
	if len(ah.RetiredTo) > 0 || len(bh.RetiredTo) > 0 || now < a.UpdatedAt || now < b.UpdatedAt || now < ah.Associations[right].LastObservedAt {
		return false, dynamicsInvalid("Cannot merge stale or retired topics")
	}
	overlap := false
	for _, ref := range a.Node.Evidence {
		if slices.ContainsFunc(b.Node.Evidence, func(other domain.EvidenceRef) bool { return other.EventID == ref.EventID }) {
			overlap = true
			break
		}
	}
	old := ah.Associations[right]
	run, ready := domain.ObserveMerge(old, now, strength, overlap, *e.options.MergeDurationMS, *e.options.MergeObservations)
	if old == run && a.Node.RelatedTopics[right] == strength && ah.RelatedAt[right] == now {
		return false, nil
	}
	ah.Associations[right] = run
	a.Node.RelatedTopics[right] = strength
	b.Node.RelatedTopics[left] = strength
	ah.RelatedAt[right] = now
	bh.RelatedAt[left] = now
	if ready {
		a.Node.Attention = domain.MergeAttention(ah.MaxAttention, bh.MaxAttention)
		// §5.4 specifies only historical A=max. Preserve the survivor's
		// orthogonal confidence, momentum and per-topic decay.
		a.Node.LastActive = max(a.Node.LastActive, b.Node.LastActive)
		a.UpdatedAt = now
		a.Lifecycle = "active"
		for _, ref := range b.Node.Evidence {
			if !slices.ContainsFunc(a.Node.Evidence, func(old domain.EvidenceRef) bool {
				return old.EventID == ref.EventID && old.WeightAtUse == ref.WeightAtUse
			}) {
				a.Node.Evidence = append(a.Node.Evidence, ref)
			}
		}
		for key, used := range bh.Applied {
			ah.Applied[key] = used
		}
		for _, project := range b.Node.RelatedProjects {
			if !slices.Contains(a.Node.RelatedProjects, project) {
				a.Node.RelatedProjects = append(a.Node.RelatedProjects, project)
			}
		}
		for _, agent := range b.Node.RelatedAgents {
			if !slices.Contains(a.Node.RelatedAgents, agent) {
				a.Node.RelatedAgents = append(a.Node.RelatedAgents, agent)
			}
		}
		for topic, weight := range b.Node.RelatedTopics {
			if topic != left {
				a.Node.RelatedTopics[topic] = max(a.Node.RelatedTopics[topic], weight)
				ah.RelatedAt[topic] = max(ah.RelatedAt[topic], bh.RelatedAt[topic])
			}
		}
		ah.Transitions = append(ah.Transitions, TopicTransition{Kind: "merge", TS: now, Topics: []string{right}})
		bh.RetiredTo = []string{left}
		b.Lifecycle = "cold"
		bh.Transitions = append(bh.Transitions, TopicTransition{Kind: "merged_into", TS: now, Topics: []string{left}})
	}
	if err = saveDynamicsNode(tx, &a, ah); err != nil {
		return false, err
	}
	if err = saveDynamicsNode(tx, &b, bh); err != nil {
		return false, err
	}
	return ready, nil
}
