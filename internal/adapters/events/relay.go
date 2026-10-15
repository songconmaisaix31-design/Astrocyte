package events

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// RelayEventSink records observations in the isolated application's existing
// event repository. It never normalizes a signal or writes cognitive state.
type RelayEventSink struct {
	service *attention.EventService
	clock   func() time.Time
}

var _ workspace.RelayEventSink = (*RelayEventSink)(nil)

func NewRelayEventSink(service *attention.EventService, clock func() time.Time) *RelayEventSink {
	if clock == nil {
		clock = time.Now
	}
	return &RelayEventSink{service, clock}
}

func (s *RelayEventSink) EmitRelayEvent(ctx context.Context, v domain.RelayEvent) error {
	if s.service == nil || v.ID == "" || v.Kind == "" || v.Source == "" || v.Scope == "" || v.TS.IsZero() {
		return errors.New("relay event observation unavailable")
	}
	// Construct a closed derived envelope instead of serializing an arbitrary
	// provider record. No prompt, terminal text, payload or task outcome exists here.
	derived := struct {
		ID     string `json:"event_id"`
		Kind   string `json:"kind"`
		Topic  string `json:"topic"`
		Tokens *int64 `json:"tokens"`
		Scope  string `json:"scope"`
		TS     string `json:"ts"`
	}{v.ID, v.Kind, v.Topic, v.Tokens, v.Scope, v.TS.UTC().Format(time.RFC3339Nano)}
	b, err := json.Marshal(derived)
	if err != nil {
		return errors.New("relay event projection unavailable")
	}
	var tokens *int
	if v.Tokens != nil {
		n := int(*v.Tokens)
		if int64(n) != *v.Tokens || n < 0 {
			return errors.New("invalid relay token observation")
		}
		tokens = &n
	}
	_, err = s.service.Record(ctx, []attention.StoredEvent{{OriginKey: "relay:" + v.ID, ReceivedAt: s.clock().UnixMilli(), RawBytes: b, Observation: attention.Observation{Source: v.Source, SourceKind: v.Kind, Session: v.TerminalID, Scope: v.Scope, TS: v.TS.UnixMilli(), Tokens: tokens, Outcome: "unknown", Depth: "unknown", Provenance: "workspace:relay"}}})
	// Preserve sink failures; no new origin, retry or false acknowledgement.
	return err
}
