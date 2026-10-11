package app

import (
	"context"
	"fmt"
	"strings"
)

// MCPBinding is supplied by startup, never decoded from MCP arguments.
type MCPBinding struct {
	Principal Principal
	Scope     string
	Session   string
	CanVeto   bool
}

type boundMCPBackend struct {
	backend MCPBackend
	binding MCPBinding
}

func BindMCPBackend(backend MCPBackend, binding MCPBinding) (MCPBackend, error) {
	if backend == nil || strings.TrimSpace(binding.Principal.ID) == "" || (binding.Principal.Kind != "agent" && binding.Principal.Kind != "human") || strings.TrimSpace(binding.Scope) == "" || strings.TrimSpace(binding.Session) == "" {
		return nil, fmt.Errorf("MCP requires a backend and startup-bound principal, project scope and session")
	}
	if binding.CanVeto && binding.Principal.Kind != "human" {
		return nil, fmt.Errorf("MCP veto requires a human principal")
	}
	return &boundMCPBackend{backend: backend, binding: binding}, nil
}

func (b *boundMCPBackend) GetAttentionState(ctx context.Context) (GetAttentionStateResult, error) {
	return b.backend.GetAttentionState(ctx)
}

func (b *boundMCPBackend) ExplainFocus(ctx context.Context, in ExplainFocusInput) (ExplainFocusResult, error) {
	return b.backend.ExplainFocus(ctx, in)
}

func (b *boundMCPBackend) GetExperiences(ctx context.Context, in GetExperiencesInput) (GetExperiencesResult, error) {
	if in.Scope != "" && in.Scope != b.binding.Scope {
		return GetExperiencesResult{}, fmt.Errorf("experience scope differs from startup project")
	}
	in.Scope = b.binding.Scope
	return b.backend.GetExperiences(ctx, in)
}

func (b *boundMCPBackend) ReportEvent(ctx context.Context, in ReportEventInput) (ReportEventResult, error) {
	if in.Scope != "" && in.Scope != b.binding.Scope {
		return ReportEventResult{}, fmt.Errorf("event scope differs from startup project")
	}
	switch in.Kind {
	case "search", "edit_file", "read_paper", "watch_video", "bookmark", "commit", "agent_task", "collect":
	default:
		return ReportEventResult{}, fmt.Errorf("event kind is not an external observation")
	}
	in.Source, in.Scope, in.Session = "mcp", b.binding.Scope, b.binding.Session
	return b.backend.ReportEvent(ctx, in)
}

func (b *boundMCPBackend) Veto(ctx context.Context, in VetoInput) (VetoResult, error) {
	if !b.binding.CanVeto || b.binding.Principal.Kind != "human" {
		return VetoResult{}, fmt.Errorf("veto requires the trusted human startup channel")
	}
	if (in.Topic == "") == (in.Task == "") {
		return VetoResult{}, fmt.Errorf("veto requires exactly one topic or task")
	}
	return b.backend.Veto(ctx, in)
}
