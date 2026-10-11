// Sleep consolidation (M3) owns the nightly pass that turns same-day events
// into stable concepts. Every storage decision in this package comes from
// deterministic rules; the interpreter only explains clustering.
package domain

import (
	"slices"
	"strings"
)

// EventTopic reduces a same-day event to its cognitive topic for clustering.
// Chain-1 events map onto this view; the full payload stays in the event store.
type EventTopic struct {
	EventID string
	Topic   string
}

// Cluster is a set of same-day events united by topic neighborhood.
type Cluster struct {
	Topics []string
	Events int
}

// ClusterTopics groups events into clusters by topic neighborhood using
// union-find over related-topic edges. Events sharing a topic, or whose topics
// are adjacent, land in the same cluster. Output is ordered by first topic.
func ClusterTopics(events []EventTopic, related map[string][]string) []Cluster {
	parent := make(map[string]string)
	count := make(map[string]int)
	for _, e := range events {
		t := strings.TrimSpace(e.Topic)
		if t == "" {
			continue
		}
		if _, ok := parent[t]; !ok {
			parent[t] = t
		}
		count[t]++
	}

	find := func(x string) string {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra == rb {
			return
		}
		parent[rb] = ra
	}

	for t, neighbors := range related {
		t = strings.TrimSpace(t)
		if _, ok := parent[t]; !ok {
			continue
		}
		for _, n := range neighbors {
			n = strings.TrimSpace(n)
			if _, ok := parent[n]; !ok {
				continue
			}
			union(t, n)
		}
	}

	rootTopics := make(map[string][]string)
	rootCount := make(map[string]int)
	for t := range parent {
		root := find(t)
		rootTopics[root] = append(rootTopics[root], t)
		rootCount[root] += count[t]
	}

	clusters := make([]Cluster, 0, len(rootTopics))
	for root, members := range rootTopics {
		slices.Sort(members)
		clusters = append(clusters, Cluster{Topics: members, Events: rootCount[root]})
	}
	slices.SortFunc(clusters, func(a, b Cluster) int {
		return strings.Compare(a.Topics[0], b.Topics[0])
	})
	return clusters
}

// Denoise drops clusters backed by a single isolated signal, which must not
// promote a transient topic to a node. A cluster is isolated when it carries
// exactly one event and none of its topics has an adjacent neighbor.
func Denoise(clusters []Cluster, related map[string][]string) []Cluster {
	kept := make([]Cluster, 0, len(clusters))
	for _, c := range clusters {
		if isolated(c, related) {
			continue
		}
		kept = append(kept, c)
	}
	return kept
}

// isolated reports whether the cluster is a single event with no neighbors.
func isolated(c Cluster, related map[string][]string) bool {
	if c.Events != 1 {
		return false
	}
	for from, neighbors := range related {
		from = strings.TrimSpace(from)
		for _, to := range neighbors {
			to = strings.TrimSpace(to)
			if from == "" || to == "" || from == to {
				continue
			}
			for _, topic := range c.Topics {
				if topic == from || topic == to {
					return false
				}
			}
		}
	}
	return true
}

// NodeDecisionKind enumerates the deterministic storage decisions a sleep
// cycle makes about a cluster.
type NodeDecisionKind int

const (
	NodeCreate NodeDecisionKind = iota // promote a topic to a concept candidate
)

// NodeDecision is a storage decision owned by deterministic rules, never by an
// interpreter. It concerns exactly one topic.
type NodeDecision struct {
	Kind  NodeDecisionKind
	Topic string
}

// ResolveClusters emits one NodeCreate per topic in each surviving cluster.
// Adjacent topics in a multi-topic cluster are related but never auto-merged:
// merging two existing topics is a W2 lifecycle decision (association strength
// sustained above 0.8 plus overlapping evidence, §5.4), so the association is
// retained instead. The interpreter is not a parameter, so its output can
// never change a decision.
func ResolveClusters(clusters []Cluster) []NodeDecision {
	decisions := make([]NodeDecision, 0, len(clusters))
	for _, c := range clusters {
		for _, topic := range c.Topics {
			decisions = append(decisions, NodeDecision{Kind: NodeCreate, Topic: topic})
		}
	}
	return decisions
}

// Consolidation is the deterministic concept stage of one sleep pass.
type Consolidation struct {
	Clusters  []Cluster      // surviving (post-denoise) clusters
	Decisions []NodeDecision // one create per topic
}

// Consolidate runs cluster -> denoise -> resolve. It takes no interpreter, so
// its decisions are identical with the LLM disconnected; the interpreter only
// decorates Clusters with explanations elsewhere.
func Consolidate(events []EventTopic, related map[string][]string) Consolidation {
	clusters := ClusterTopics(events, related)
	clusters = Denoise(clusters, related)
	return Consolidation{Clusters: clusters, Decisions: ResolveClusters(clusters)}
}

// ClusterExplanation is the interpreter's only output: a name and a note.
type ClusterExplanation struct {
	Label string
	Note  string
}

// ClusterInterpreter explains why a cluster coheres. It has no write path, and
// its output never feeds a NodeDecision; running the pipeline with a nil
// interpreter (LLM disconnected) leaves every decision unchanged.
type ClusterInterpreter interface {
	Explain(cluster Cluster) (ClusterExplanation, error)
}

// ConflictCandidate is a pending-adjudication queue item produced by the sleep
// cycle when consolidation surfaces a fact conflict. It records what conflicts
// without deciding how to resolve it; resolution belongs to the user gate.
type ConflictCandidate struct {
	Topic   string
	Summary string
}
