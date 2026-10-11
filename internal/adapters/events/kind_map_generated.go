// Code generated from events/kind_map.py; DO NOT EDIT.
package events

var nativeToolNamespaces = []string{"functions.", "tools.", "container."}

const nativePaperThreshold = 1

var nativeKindRules = []kindRule{
	{"tool", "Edit", "edit_file"},
	{"tool", "Write", "edit_file"},
	{"tool", "MultiEdit", "edit_file"},
	{"tool", "apply_patch", "edit_file"},
	{"tool", "Grep", "search"},
	{"tool", "Glob", "search"},
	{"tool", "WebSearch", "search"},
	{"tool", "WebFetch", "search"},
	{"tool", "web.run", "search"},
	{"tool", "ReadPaper", "read_paper"},
	{"wrapper", "exec", "unknown"},
	{"shell_search", "Bash", "search"},
	{"shell_search", "exec_command", "search"},
	{"shell_search", "shell_command", "search"},
	{"assistant_marker", "(?i)\\b(arxiv|pdf)\\b|arxiv\\.org/(abs|pdf)/", "read_paper"},
	{"git_record", "commit", "commit"},
}
var nativeTopicRules = []topicRule{
	{"Morphogenesis", []string{"morphogenesis"}, []string{"morphogenesis"}},
	{"Agent Swarm", []string{"agent-swarm", "agent_swarm", "agentswarm"}, []string{"agent swarm", "agentswarm"}},
	{"Astrocyte/Attention Engine", []string{"astrocyte", "attention-engine", "attention_engine"}, []string{"astrocyte", "attention engine"}},
	{"Agent Memory", []string{"agent-memory", "agent_memory"}, []string{"agent memory"}},
	{"MCP", []string{"mcp"}, []string{"mcp", "model context protocol"}},
}
