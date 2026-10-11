package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type MCPAttentionReader interface {
	AttentionState(context.Context) ([]domain.AttentionNode, error)
	ExplainFocus(context.Context, string) (FocusExplanation, error)
}

type MCPExperienceReader interface {
	GetExperiences(context.Context, Principal, string, string) ([]StoredExperience, error)
}

type MCPEventRecorder interface {
	Record(context.Context, []StoredEvent) ([]StoredEvent, error)
}

type MCPBackendOptions struct{ Clock func() time.Time }

type engineMCPBackend struct {
	binding     MCPBinding
	attention   MCPAttentionReader
	experiences MCPExperienceReader
	events      MCPEventRecorder
	clock       func() time.Time
}

func NewEngineMCPBackend(binding MCPBinding, attention MCPAttentionReader, experiences MCPExperienceReader, events MCPEventRecorder, options MCPBackendOptions) (MCPBackend, error) {
	if attention == nil || experiences == nil || events == nil {
		return nil, fmt.Errorf("MCP requires attention, experience and event use cases")
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return BindMCPBackend(&engineMCPBackend{binding: binding, attention: attention, experiences: experiences, events: events, clock: options.Clock}, binding)
}

func (b *engineMCPBackend) GetAttentionState(ctx context.Context) (GetAttentionStateResult, error) {
	nodes, err := b.attention.AttentionState(ctx)
	if err != nil {
		return GetAttentionStateResult{}, err
	}
	topics := make([]FocusItem, 0, len(nodes))
	for _, n := range nodes {
		topics = append(topics, FocusItem{Topic: n.Topic, Attention: n.Attention, Momentum: n.Momentum})
	}
	return GetAttentionStateResult{Topics: topics}, nil
}

func (b *engineMCPBackend) ExplainFocus(ctx context.Context, in ExplainFocusInput) (ExplainFocusResult, error) {
	return b.attention.ExplainFocus(ctx, in.Topic)
}

func (b *engineMCPBackend) GetExperiences(ctx context.Context, in GetExperiencesInput) (GetExperiencesResult, error) {
	rows, err := b.experiences.GetExperiences(ctx, b.binding.Principal, b.binding.Scope, in.Pattern)
	if err != nil {
		return GetExperiencesResult{}, err
	}
	if rows == nil {
		rows = []StoredExperience{}
	}
	return GetExperiencesResult{Experiences: rows}, nil
}

func (b *engineMCPBackend) ReportEvent(ctx context.Context, in ReportEventInput) (ReportEventResult, error) {
	raw := in.OriginalPayload
	if len(raw) == 0 {
		var err error
		raw, err = json.Marshal(in)
		if err != nil {
			return ReportEventResult{}, err
		}
	}
	observation := Observation{Source: domain.EventSourceMCP, SourceKind: in.Kind, Session: b.binding.Session, Scope: b.binding.Scope, TS: in.TS, Tokens: in.Tokens, Outcome: "unknown", Depth: "unknown", Provenance: "mcp_unverified:" + b.binding.Principal.ID}
	var event *domain.Event
	if in.Topic != "" {
		tokens := 0
		if in.Tokens != nil {
			tokens = *in.Tokens
		}
		event = &domain.Event{Source: domain.EventSourceMCP, Session: b.binding.Session, TS: in.TS, Kind: in.Kind, Tokens: tokens, Scope: b.binding.Scope, Topic: in.Topic, Raw: in.Raw}
	}
	id, err := b.record(ctx, observation, event, raw)
	return ReportEventResult{Accepted: err == nil, EventID: id}, err
}

func (b *engineMCPBackend) Veto(ctx context.Context, in VetoInput) (VetoResult, error) {
	if in.Topic != "" {
		if _, err := b.attention.ExplainFocus(ctx, in.Topic); err != nil {
			return VetoResult{}, err
		}
	}
	raw, err := json.Marshal(struct {
		VetoInput
		ActorID   string `json:"actor_id"`
		ActorKind string `json:"actor_kind"`
	}{in, b.binding.Principal.ID, b.binding.Principal.Kind})
	if err != nil {
		return VetoResult{}, err
	}
	now := b.clock().UnixMilli()
	observation := Observation{Source: domain.EventSourceManual, SourceKind: domain.EventKindUserVeto, Session: b.binding.Session, Scope: b.binding.Scope, TS: now, Provenance: "mcp_human:" + b.binding.Principal.ID}
	var event *domain.Event
	if in.Topic != "" {
		event = &domain.Event{Source: domain.EventSourceManual, Session: b.binding.Session, Scope: b.binding.Scope, TS: now, Kind: domain.EventKindUserVeto, Topic: in.Topic, Raw: raw}
	}
	id, err := b.record(ctx, observation, event, raw)
	return VetoResult{Acknowledged: err == nil, Target: in.Topic + in.Task, EventID: id, Pending: event == nil}, err
}

func (b *engineMCPBackend) record(ctx context.Context, observation Observation, event *domain.Event, raw []byte) (string, error) {
	rows, err := b.events.Record(ctx, []StoredEvent{{OriginKey: "mcp:" + b.binding.Session + ":" + rand.Text(), ReceivedAt: b.clock().UnixMilli(), Observation: observation, Event: event, RawBytes: raw}})
	if err != nil {
		return "", err
	}
	if len(rows) != 1 || rows[0].ID == "" {
		return "", fmt.Errorf("event use case did not return a persisted receipt")
	}
	return rows[0].ID, nil
}
