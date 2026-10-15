package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type SleepOptions struct {
	Clock        func() time.Time
	ProjectFacts SleepFactProjector
	Interpreter  domain.ClusterInterpreter
}

type SleepCandidate struct {
	ID           string            `json:"id"`
	Kind         string            `json:"kind"`
	Topic        string            `json:"topic"`
	OwnerProject string            `json:"owner_project,omitempty"`
	Sources      []domain.EventRef `json:"sources,omitempty"`
	Facts        []SleepFact       `json:"facts,omitempty"`
	ConflictID   string            `json:"conflict_id,omitempty"`
	RelatedTopic string            `json:"related_topic,omitempty"`
	Note         string            `json:"note,omitempty"`
}

type SleepExplanation struct {
	Topics      []string                  `json:"topics"`
	Explanation domain.ClusterExplanation `json:"explanation"`
	Error       string                    `json:"error,omitempty"`
}

type SleepResult struct {
	Run          EngineSleepRun     `json:"run"`
	Replayed     bool               `json:"replayed"`
	Explanations []SleepExplanation `json:"explanations,omitempty"`
}

type SleepService struct {
	repo    SleepRepository
	engine  *DynamicsEngine
	options SleepOptions
}

func NewSleepService(repo SleepRepository, engine *DynamicsEngine, options SleepOptions) *SleepService {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &SleepService{repo: repo, engine: engine, options: options}
}

// Run is the fixed daily entry point; day.Location selects the calendar window.
// The daily receipt, graph, fluid and pending conflicts commit together.
func (s *SleepService) Run(ctx context.Context, day time.Time) (SleepResult, error) {
	var result SleepResult
	if day.IsZero() || s.engine == nil {
		return result, domain.ErrInvalid
	}
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	started := s.options.Clock().UnixMilli()
	if start.UnixMilli() > started {
		return result, domain.ErrInvalid
	}
	until := min(start.AddDate(0, 0, 1).UnixMilli()-1, started)
	var clusters []domain.Cluster
	err := s.repo.WithSleepTx(ctx, func(tx SleepTx) error {
		run, found, err := tx.LoadSleepRun(start.Format(time.DateOnly))
		if err != nil {
			return err
		}
		if found {
			result.Run, result.Replayed = run, true
			return nil
		}
		run = EngineSleepRun{ID: "sleep:" + start.Format(time.DateOnly), Day: start.Format(time.DateOnly), StartedAt: started, ConflictIDs: []string{}}
		events, err := tx.ListEvents(start.UnixMilli(), until)
		if err != nil {
			return err
		}
		signals, err := tx.ListSignals(start.UnixMilli(), until)
		if err != nil {
			return err
		}
		nodes, err := tx.ListNodes()
		if err != nil {
			return err
		}
		related := map[string][]string{}
		weights := map[string]map[string]float64{}
		for _, node := range nodes {
			weights[node.Node.Topic] = node.Node.RelatedTopics
			for topic, strength := range node.Node.RelatedTopics {
				if strength > 0 {
					related[node.Node.Topic] = append(related[node.Node.Topic], topic)
				}
			}
		}
		byTopic := map[string][]StoredEvent{}
		views := []domain.EventTopic{}
		candidates := []SleepCandidate{}
		for _, event := range events {
			if event.Event == nil || strings.TrimSpace(event.Event.Topic) == "" {
				ts, kind, note := event.Observation.TS, "unassigned_pending", "topic_candidate_unavailable"
				if event.Event == nil {
					kind = "mapping_pending"
					note = "mapping_pending_bucket_unknown"
				} else if ts <= 0 {
					ts = event.Event.TS
				}
				if ts <= 0 {
					note += ";observed_ts_unknown"
				}
				candidates = append(candidates, SleepCandidate{ID: fmt.Sprintf("receipt:%q", event.ID), Kind: kind, Sources: []domain.EventRef{{EventID: event.ID, TS: ts}}, Note: note})
				continue
			}
			topic := strings.TrimSpace(event.Event.Topic)
			byTopic[topic] = append(byTopic[topic], event)
			views = append(views, domain.EventTopic{EventID: event.ID, Topic: topic})
		}
		core := domain.Consolidate(views, related)
		clusters = core.Clusters
		created := map[string]bool{}
		for _, cluster := range clusters {
			for _, topic := range cluster.Topics {
				own := byTopic[topic]
				refs := sleepEvidence(own)
				_, existsErr := tx.GetNode(topic)
				if existsErr != nil && !dynamicsNotFound(existsErr) {
					return existsErr
				}
				if dynamicsNotFound(existsErr) {
					// Current cluster views contain only tagged events. Unlabeled
					// receipts stay pending until a production candidate exists.
					candidates = append(candidates, SleepCandidate{ID: fmt.Sprintf("concept:%q", topic), Kind: "concept_pending", Topic: topic, Sources: sleepSources(own), Note: "tagged_birth_disabled_v0_2"})
					continue
				}
				var batch []domain.Signal
				for _, signal := range signals {
					if signal.Topic == topic {
						batch = append(batch, signal)
					}
				}
				// W2 verifies and deduplicates persisted signals; Sleep does not
				// recalculate strength, weights, A or M.
				if _, err = s.engine.ApplySignalsInTx(tx, topic, batch, started); err != nil {
					return err
				}
				if _, err = s.engine.ConsolidateTopicInTx(tx, topic, refs, nil, started); err != nil {
					return err
				}
				created[topic] = true
				candidates = append(candidates, SleepCandidate{ID: fmt.Sprintf("concept:%q", topic), Kind: "concept", Topic: topic, Sources: sleepSources(own)})
			}
		}
		// Attach only observed edges, after both endpoints can exist. A cluster
		// never invents a stronger association or authorizes a topic merge.
		for _, decision := range core.Decisions {
			if !created[decision.Topic] {
				continue
			}
			edges := map[string]float64{}
			for neighbor, strength := range weights[decision.Topic] {
				if _, err := tx.GetNode(neighbor); err == nil {
					edges[neighbor] = strength
				} else if !dynamicsNotFound(err) {
					return err
				}
			}
			if _, err = s.engine.ConsolidateTopicInTx(tx, decision.Topic, sleepEvidence(byTopic[decision.Topic]), edges, started); err != nil {
				return err
			}
		}
		if err = s.observeSleepAssociations(tx, started, &candidates); err != nil {
			return err
		}
		before, after, beforeBytes, afterBytes, err := s.consolidateSleepKnowledge(tx, byTopic, created, &candidates, run.ID, started)
		if err != nil {
			return err
		}
		if err = s.engine.AdvanceLifecycleInTx(tx, started); err != nil {
			return err
		}
		rng := rand.New(rand.NewPCG(uint64(start.UnixMilli()), uint64(start.AddDate(0, 0, 1).UnixMilli())))
		resurfaced, err := s.engine.ResurfaceInTx(tx, started, rng)
		if err != nil {
			return err
		}
		for _, node := range resurfaced {
			candidates = append(candidates, SleepCandidate{ID: fmt.Sprintf("resurface:%q", node.Node.Topic), Kind: "resurface", Topic: node.Node.Topic})
		}
		conflicts, err := tx.ListConflicts()
		if err != nil {
			return err
		}
		for _, conflict := range conflicts {
			if conflict.Conflict.Status != domain.ConflictStatusPending {
				continue
			}
			run.ConflictIDs = append(run.ConflictIDs, conflict.Conflict.ID)
			if !slices.ContainsFunc(candidates, func(c SleepCandidate) bool { return c.ConflictID == conflict.Conflict.ID }) {
				candidates = append(candidates, SleepCandidate{ID: "conflict:" + conflict.Conflict.ID, Kind: "conflict", Topic: conflict.Conflict.Topic, ConflictID: conflict.Conflict.ID})
			}
		}
		slices.SortFunc(candidates, func(a, b SleepCandidate) int { return strings.Compare(a.ID, b.ID) })
		for _, candidate := range candidates {
			data, err := json.Marshal(candidate)
			if err != nil {
				return err
			}
			afterBytes += int64(len(data))
			if candidate.OwnerProject != "" {
				after = append(after, sleepFactKeys(candidate.OwnerProject, candidate.Topic, candidate.Facts)...)
			}
		}
		m := domain.ComputeMetrics(before, after, beforeBytes, afterBytes)
		retained := 0
		for fact := range sleepFactSetForRun(before) {
			if slices.Contains(after, fact) {
				retained++
			}
		}
		run.Metrics = EngineSleepMetrics{RetentionMeasurable: m.RetentionMeasurable && s.options.ProjectFacts != nil, StorageMeasurable: m.StorageMeasurable,
			Representation: "mapped event raw BLOB + touched knowledge.data UTF-8 -> touched knowledge.data UTF-8 + cycle candidates JSON; raw events retained; excludes graph and SQLite pages",
			KeyFactsBefore: m.BeforeFacts, KeyFactsRetained: retained, BytesBefore: beforeBytes, BytesAfter: afterBytes, RetentionPrecision: m.RetentionPrecision, StorageReduction: m.StorageReduction}
		run.Candidates, err = json.Marshal(candidates)
		if err != nil {
			return err
		}
		run.CompletedAt = s.options.Clock().UnixMilli()
		if err = tx.SaveSleepRun(run); err != nil {
			return err
		}
		result.Run = run
		return nil
	})
	if err != nil {
		return SleepResult{}, err
	}
	// The interpreter runs after commit and only sees a detached topic view.
	if !result.Replayed && s.options.Interpreter != nil {
		for _, cluster := range clusters {
			copy := domain.Cluster{Topics: slices.Clone(cluster.Topics), Events: cluster.Events}
			explanation, err := s.options.Interpreter.Explain(copy)
			row := SleepExplanation{Topics: slices.Clone(cluster.Topics), Explanation: explanation}
			if err != nil {
				row.Error = err.Error()
			}
			result.Explanations = append(result.Explanations, row)
		}
	}
	return result, nil
}

func (s *SleepService) observeSleepAssociations(tx SleepTx, now int64, candidates *[]SleepCandidate) error {
	nodes, err := tx.ListNodes()
	if err != nil {
		return err
	}
	slices.SortFunc(nodes, func(a, b StoredAttentionNode) int { return strings.Compare(a.Node.Topic, b.Node.Topic) })
	seen := map[[2]string]bool{}
	for _, node := range nodes {
		neighbors := []string{}
		for neighbor := range node.Node.RelatedTopics {
			neighbors = append(neighbors, neighbor)
		}
		slices.Sort(neighbors)
		for _, neighbor := range neighbors {
			pair := [2]string{node.Node.Topic, neighbor}
			if pair[1] < pair[0] {
				pair[0], pair[1] = pair[1], pair[0]
			}
			if pair[0] == pair[1] || seen[pair] {
				continue
			}
			seen[pair] = true
			left, err := tx.GetNode(pair[0])
			if dynamicsNotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			right, err := tx.GetNode(pair[1])
			if dynamicsNotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			lh, err := readDynamicsHistory(left)
			if err != nil {
				return err
			}
			rh, err := readDynamicsHistory(right)
			if err != nil {
				return err
			}
			if len(lh.RetiredTo) > 0 || len(rh.RetiredTo) > 0 {
				continue
			}
			if s.engine.options.MergeDurationMS == nil || s.engine.options.MergeObservations == nil {
				*candidates = append(*candidates, SleepCandidate{ID: fmt.Sprintf("association-pending:%q:%q", pair[0], pair[1]), Kind: "association_pending", Topic: pair[0], RelatedTopic: pair[1], Note: "merge_policy_undecided"})
				continue
			}
			if _, err = s.engine.ObserveAssociationInTx(tx, pair[0], pair[1], node.Node.RelatedTopics[neighbor], now); err != nil {
				return err
			}
		}
	}
	return nil
}

func sleepFactSetForRun(facts []string) map[string]bool {
	set := map[string]bool{}
	for _, fact := range facts {
		set[fact] = true
	}
	return set
}

func sleepSources(events []StoredEvent) []domain.EventRef {
	refs := []domain.EventRef{}
	for _, event := range events {
		refs = append(refs, domain.EventRef{EventID: event.ID, TS: event.Event.TS})
	}
	return refs
}

func sleepEvidence(events []StoredEvent) []domain.EvidenceRef {
	refs := []domain.EvidenceRef{}
	for _, event := range events {
		refs = append(refs, domain.EvidenceRef{EventID: event.ID, TS: event.Event.TS})
	}
	return refs
}

func sleepBirthSupport(tx SleepTx, topic string, cluster domain.Cluster, byTopic map[string][]StoredEvent, nodes []StoredAttentionNode) ([]domain.EvidenceRef, []domain.EventRef, error) {
	scopes := map[string][]StoredEvent{}
	for _, name := range cluster.Topics {
		for _, event := range byTopic[name] {
			if event.Event.Scope != "" {
				scopes[event.Event.Scope] = append(scopes[event.Event.Scope], event)
			}
		}
	}
	for _, node := range nodes {
		if !slices.Contains(cluster.Topics, node.Node.Topic) && node.Node.RelatedTopics[topic] <= 0 {
			continue
		}
		for _, ref := range node.Node.Evidence {
			event, err := tx.GetEvent(ref.EventID)
			if err != nil {
				return nil, nil, err
			}
			if event.Event != nil && event.Event.Scope != "" {
				scopes[event.Event.Scope] = append(scopes[event.Event.Scope], event)
			}
		}
	}
	owners := []string{}
	for owner := range scopes {
		owners = append(owners, owner)
	}
	slices.Sort(owners)
	for _, owner := range owners {
		unique := map[string]bool{}
		for _, event := range scopes[owner] {
			unique[event.ID] = true
		}
		if len(unique) < 2 {
			continue
		}
		own := []StoredEvent{}
		for _, event := range byTopic[topic] {
			if event.Event.Scope == owner {
				own = append(own, event)
			}
		}
		if len(own) > 0 {
			return sleepEvidence(own), sleepSources(scopes[owner]), nil
		}
	}
	return nil, nil, nil
}
