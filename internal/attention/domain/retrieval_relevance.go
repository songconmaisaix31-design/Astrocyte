package domain

import (
	"math"
	"strings"
)

// KeywordKnowledgeRelevance is the explicitly cut vector requirement's fallback.
// Topic equality, persisted topic adjacency and literal terms are independently visible.
func KeywordKnowledgeRelevance(entry KnowledgeEntry, node AttentionNode, topic, text string) float64 {
	if topic == "" && strings.TrimSpace(text) == "" {
		return 1
	}
	relevance := 0.0
	if topic != "" {
		if entry.Topic == topic {
			return 1
		}
		relevance = node.RelatedTopics[topic]
	}
	terms := strings.Fields(strings.ToLower(text))
	if len(terms) > 0 {
		haystack := strings.ToLower(entry.Topic + " " + entry.Fluid + " " + string(entry.Crystallized))
		matched := 0
		for _, term := range terms {
			if strings.Contains(haystack, term) {
				matched++
			}
		}
		relevance = math.Max(relevance, float64(matched)/float64(len(terms)))
	}
	return math.Min(1, math.Max(0, relevance))
}
