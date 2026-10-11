package app

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (s *SleepService) consolidateSleepKnowledge(tx SleepTx, byTopic map[string][]StoredEvent, created map[string]bool, candidates *[]SleepCandidate, runID string, now int64) ([]string, []string, int64, int64, error) {
	var before, after []string
	var beforeBytes, afterBytes int64
	topics := []string{}
	for topic := range byTopic {
		topics = append(topics, topic)
	}
	slices.Sort(topics)
	for _, topic := range topics {
		groups := map[string][]StoredEvent{}
		for _, event := range byTopic[topic] {
			groups[event.Event.Scope] = append(groups[event.Event.Scope], event)
			beforeBytes += int64(len(event.RawBytes))
		}
		owners := []string{}
		for owner := range groups {
			owners = append(owners, owner)
		}
		slices.Sort(owners)
		for _, owner := range owners {
			events := groups[owner]
			incoming := []SleepFact{}
			for _, event := range events {
				facts, err := projectSleepFacts(s.options.ProjectFacts, event.Event.Raw)
				if err != nil {
					return nil, nil, 0, 0, err
				}
				incoming = append(incoming, facts...)
			}
			incoming = uniqueSleepFacts(incoming)
			before = append(before, sleepFactKeys(owner, topic, incoming)...)
			if owner == "" {
				if created[topic] {
					*candidates = append(*candidates, SleepCandidate{ID: fmt.Sprintf("knowledge-pending:%q:%q", owner, topic), Kind: "knowledge_pending", Topic: topic, Sources: sleepSources(events), Facts: incoming, Note: "project_scope_unknown"})
				}
				continue
			}
			all, err := tx.ListKnowledge(owner, false)
			if err != nil {
				return nil, nil, 0, 0, err
			}
			entries := []StoredKnowledge{}
			for _, entry := range all {
				if entry.OwnerProject == owner && entry.Entry.Topic == topic {
					entries = append(entries, entry)
				}
			}
			slices.SortFunc(entries, func(a, b StoredKnowledge) int { return strings.Compare(a.Entry.ID, b.Entry.ID) })
			oldSize, err := tx.SleepKnowledgeBytes(owner, topic)
			if err != nil {
				return nil, nil, 0, 0, err
			}
			beforeBytes += oldSize
			for _, entry := range entries {
				facts, err := storedSleepFacts(s.options.ProjectFacts, entry)
				if err != nil {
					return nil, nil, 0, 0, err
				}
				before = append(before, sleepFactKeys(owner, topic, facts)...)
			}
			if created[topic] {
				if len(entries) == 0 {
					entries = append(entries, StoredKnowledge{OwnerProject: owner, Entry: domain.KnowledgeEntry{ID: fmt.Sprintf("sleep-knowledge:%q:%q", owner, topic), Topic: topic}})
				}
				for _, entry := range entries {
					oldFacts, err := storedSleepFacts(s.options.ProjectFacts, entry)
					if err != nil {
						return nil, nil, 0, 0, err
					}
					pairs := sleepFactContradictions(oldFacts, incoming)
					for i, a := range incoming {
						for _, b := range incoming[i+1:] {
							if a.Key == b.Key && a.Value != b.Value {
								pairs = append(pairs, [2]SleepFact{a, b})
							}
						}
					}
					if len(pairs) > 0 {
						for index, pair := range pairs {
							left, right := sleepClaimSource(s.options.ProjectFacts, events, pair[0]), sleepClaimSource(s.options.ProjectFacts, events, pair[1])
							if left == "" && len(entry.Entry.Sources) > 0 {
								left = entry.Entry.Sources[0].EventID
							}
							detail, _ := json.Marshal(struct {
								Owner  string       `json:"owner_project"`
								Entry  string       `json:"entry_id"`
								Claims [2]SleepFact `json:"claims"`
							}{owner, entry.Entry.ID, pair})
							id := fmt.Sprintf("fact-conflict:%q:%q:%q:%d", runID, owner, entry.Entry.ID, index)
							candidate := SleepCandidate{ID: id, Kind: "conflict", Topic: topic, OwnerProject: owner, Sources: sleepSources(events), Facts: []SleepFact{pair[0], pair[1]}}
							if left != "" && right != "" {
								conflict, err := s.engine.ReportFactConflictInTx(tx, topic, left, right, string(detail), now)
								if err != nil {
									return nil, nil, 0, 0, err
								}
								candidate.ConflictID = conflict.Conflict.ID
							} else {
								candidate.Kind = "conflict_pending_evidence"
								candidate.Note = "original_left_or_right_evidence_unknown"
							}
							*candidates = append(*candidates, candidate)
						}
						continue // The original fluid and crystallized facts remain intact.
					}
					fluid := decodeSleepFluid(entry.Entry.Fluid)
					fluid.Format = "sleep-v1"
					fluid.Facts = uniqueSleepFacts(append(oldFacts, incoming...))
					fluid.Observations = map[string]int{}
					fluid.ObservationDay = strings.TrimPrefix(runID, "sleep:")
					for _, event := range events {
						fluid.Observations[event.Event.Kind]++
					}
					data, err := json.Marshal(fluid)
					if err != nil {
						return nil, nil, 0, 0, err
					}
					sources := slices.Clone(entry.Entry.Sources)
					for _, ref := range sleepSources(events) {
						if !slices.ContainsFunc(sources, func(old domain.EventRef) bool { return old.EventID == ref.EventID }) {
							sources = append(sources, ref)
						}
					}
					slices.SortFunc(sources, func(a, b domain.EventRef) int {
						if a.TS < b.TS {
							return -1
						}
						if a.TS > b.TS {
							return 1
						}
						return strings.Compare(a.EventID, b.EventID)
					})
					if _, err = RewriteKnowledgeFluidTx(tx, KnowledgeFluidUpdate{ID: entry.Entry.ID, OwnerProject: owner, Topic: topic, Fluid: string(data), Sources: sources, ExpectedVersion: entry.Version, TS: now}); err != nil {
						return nil, nil, 0, 0, err
					}
				}
			} else if slices.ContainsFunc(*candidates, func(c SleepCandidate) bool { return c.Topic == topic && c.Kind == "concept_pending" }) {
				*candidates = append(*candidates, SleepCandidate{ID: fmt.Sprintf("knowledge-pending:%q:%q", owner, topic), Kind: "knowledge_pending", Topic: topic, OwnerProject: owner, Sources: sleepSources(events), Facts: incoming, Note: "concept_not_admitted"})
			}
			current, err := tx.ListKnowledge(owner, false)
			if err != nil {
				return nil, nil, 0, 0, err
			}
			for _, entry := range current {
				if entry.OwnerProject == owner && entry.Entry.Topic == topic {
					facts, err := storedSleepFacts(s.options.ProjectFacts, entry)
					if err != nil {
						return nil, nil, 0, 0, err
					}
					after = append(after, sleepFactKeys(owner, topic, facts)...)
				}
			}
			size, err := tx.SleepKnowledgeBytes(owner, topic)
			if err != nil {
				return nil, nil, 0, 0, err
			}
			afterBytes += size
		}
	}
	return before, after, beforeBytes, afterBytes, nil
}

func sleepClaimSource(project SleepFactProjector, events []StoredEvent, fact SleepFact) string {
	for _, event := range events {
		facts, err := projectSleepFacts(project, event.Event.Raw)
		if err == nil && slices.Contains(facts, fact) {
			return event.ID
		}
	}
	return ""
}
