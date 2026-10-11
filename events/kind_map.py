TOOL_NAMESPACES = ("functions.", "tools.", "container.")
PAPER_THRESHOLD = 1
RULES = (
    ("tool", "Edit", "edit_file"),
    ("tool", "Write", "edit_file"),
    ("tool", "MultiEdit", "edit_file"),
    ("tool", "apply_patch", "edit_file"),
    ("tool", "Grep", "search"),
    ("tool", "Glob", "search"),
    ("tool", "WebSearch", "search"),
    ("tool", "WebFetch", "search"),
    ("tool", "web.run", "search"),
    ("tool", "ReadPaper", "read_paper"),
    ("wrapper", "exec", "unknown"),
    ("shell_search", "Bash", "search"),
    ("shell_search", "exec_command", "search"),
    ("shell_search", "shell_command", "search"),
    ("assistant_marker", r"(?i)\b(arxiv|pdf)\b|arxiv\.org/(abs|pdf)/", "read_paper"),
    ("git_record", "commit", "commit"),
)
TOPIC_RULES = (
    ("Morphogenesis", ("morphogenesis",), ("morphogenesis",)),
    ("Agent Swarm", ("agent-swarm", "agent_swarm", "agentswarm"), ("agent swarm", "agentswarm")),
    ("Astrocyte/Attention Engine", ("astrocyte", "attention-engine", "attention_engine"), ("astrocyte", "attention engine")),
    ("Agent Memory", ("agent-memory", "agent_memory"), ("agent memory",)),
    ("MCP", ("mcp",), ("mcp", "model context protocol")),
)
