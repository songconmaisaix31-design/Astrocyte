package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"testing"
)

type relaySessionOwnedStub struct{ bindings []domain.RelayOwnedSession }

func (s *relaySessionOwnedStub) ListOwnedRelaySessions(context.Context) ([]domain.RelayOwnedSession, error) {
	return s.bindings, nil
}

type relaySessionMemoryState struct {
	entries relaySessionCheckpoints
	saves   int
}

func (s *relaySessionMemoryState) Load() (relaySessionCheckpoints, error) { return s.entries, nil }
func (s *relaySessionMemoryState) Save(v relaySessionCheckpoints) error {
	s.saves++
	data, _ := json.Marshal(v)
	return json.Unmarshal(data, &s.entries)
}

func TestRelaySessionFirstGapPersistsOnceAndSilentPollWritesNothing(t *testing.T) {
	owned := &relaySessionOwnedStub{bindings: []domain.RelayOwnedSession{{TaskID: uuid.NewString(), TerminalID: uuid.NewString(), WorktreeID: uuid.NewString(), ProjectID: "public-project", SourceStatus: "unknown", Gap: "archive_not_selected"}}}
	state := &relaySessionMemoryState{}
	collector := newRelaySessionCollector(owned, nil, state, nil)
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 1 {
		t.Fatal("first source gap not recorded exactly once")
	}
	for i := 0; i < 3; i++ {
		if collector.ObserveOwned(context.Background()) != nil || state.saves != 1 {
			t.Fatal("silent gap poll rewrote metadata")
		}
	}
	owned.bindings[0].Gap = "archive_metadata_unsupported"
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 2 {
		t.Fatal("changed gap not recorded exactly once")
	}
	if collector.ObserveOwned(context.Background()) != nil || state.saves != 2 {
		t.Fatal("unchanged gap rewrote metadata")
	}
}
