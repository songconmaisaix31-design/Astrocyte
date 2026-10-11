package events

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

//go:generate python ../../../scripts/validation/generate_kind_map.py

type kindRule struct{ selector, value, kind string }
type topicRule struct {
	topic            string
	scopes, keywords []string
}

func canonicalTool(name string) string {
	for _, prefix := range nativeToolNamespaces {
		if strings.HasPrefix(name, prefix) {
			return strings.TrimPrefix(name, prefix)
		}
	}
	return name
}

func toolKind(name string, raw json.RawMessage) string {
	name = canonicalTool(name)
	for _, rule := range nativeKindRules {
		if rule.value != name {
			continue
		}
		switch rule.selector {
		case "tool":
			return rule.kind
		case "shell_search":
			v := object(raw)
			if v == nil {
				var s string
				if json.Unmarshal(raw, &s) == nil {
					v = object([]byte(s))
				}
			}
			cmd := str(v, "command")
			if cmd == "" {
				cmd = str(v, "cmd")
			}
			fields := strings.Fields(cmd)
			if len(fields) == 0 || strings.ContainsAny(cmd, ";|&\n\r") {
				continue
			}
			if rule.selector == "shell_search" && (fields[0] == "rg" || fields[0] == "grep") {
				return rule.kind
			}
		}
	}
	return "unknown"
}

func assistantKind(text string) string {
	for _, rule := range nativeKindRules {
		if rule.selector == "assistant_marker" && len(regexp.MustCompile(rule.value).FindAllString(text, -1)) >= nativePaperThreshold {
			return rule.kind
		}
	}
	return "unknown"
}

func nativeText(source string, raw []byte) string {
	v := object(raw)
	if source == "claude_code" {
		if str(v, "type") != "assistant" && str(v, "type") != "user" {
			return ""
		}
		return messageText(object(v["message"])["content"])
	}
	if source == "codex" {
		p := object(v["payload"])
		if str(v, "type") == "response_item" && str(p, "type") == "message" {
			return messageText(p["content"])
		}
		if str(v, "type") == "event_msg" && (str(p, "type") == "agent_message" || str(p, "type") == "user_message") {
			return str(p, "message")
		}
	}
	return ""
}

func messageText(raw json.RawMessage) string {
	var scalar string
	if json.Unmarshal(raw, &scalar) == nil {
		return scalar
	}
	var blocks []map[string]json.RawMessage
	_ = json.Unmarshal(raw, &blocks)
	var texts []string
	for _, b := range blocks {
		if t := str(b, "type"); t == "text" || t == "output_text" || t == "input_text" {
			texts = append(texts, str(b, "text"))
		}
	}
	return strings.Join(texts, "\n")
}

func normalizedWords(text string) string {
	return " " + strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, text)), " ") + " "
}

// TopicAssignment describes a fixed seed match, never a new graph node.
type TopicAssignment struct {
	Bucket string `json:"bucket"`
	Topic  string `json:"topic,omitempty"`
	Rule   string `json:"rule"`
}

// SeedTopics returns matching labels; callers still need explicit engine policy.
func SeedTopics() []string {
	labels := make([]string, 0, len(nativeTopicRules))
	for _, rule := range nativeTopicRules {
		labels = append(labels, rule.topic)
	}
	return labels
}

func MatchSeedTopic(scope, text string) TopicAssignment {
	segments := strings.Split(strings.ToLower(strings.ReplaceAll(scope, "\\", "/")), "/")
	for _, rule := range nativeTopicRules {
		for _, segment := range segments {
			for _, prefix := range rule.scopes {
				if segment == prefix || strings.HasPrefix(segment, prefix+"-") || strings.HasPrefix(segment, prefix+"_") {
					return TopicAssignment{"assigned", rule.topic, "scope"}
				}
			}
		}
	}
	words := normalizedWords(text)
	for _, rule := range nativeTopicRules {
		for _, keyword := range rule.keywords {
			if strings.Contains(words, normalizedWords(keyword)) {
				return TopicAssignment{"assigned", rule.topic, "keyword"}
			}
		}
	}
	return TopicAssignment{Bucket: "unassigned", Rule: "no_seed_match"}
}

func ClassifyNativeReceipt(row app.StoredEvent) TopicAssignment {
	return MatchSeedTopic(row.Observation.Scope, nativeText(row.Observation.Source, row.RawBytes))
}

func nativeReceipt(origin, source, sourceKind, provenance string, raw []byte, o nativeObservation, receivedAt int64) app.StoredEvent {
	row := receipt(origin, source, sourceKind, provenance, raw, o, receivedAt)
	if row.Event != nil {
		row.Event.Topic = ClassifyNativeReceipt(row).Topic
	}
	return row
}
