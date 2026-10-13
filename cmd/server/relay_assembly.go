package main

import (
	"errors"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// This entrypoint owns cross-context assembly. The runtime remains the sole
// task/terminal/worktree writer; event observations enter the existing app port.
func assembleRelay(services *httpapi.Services, db *sqlite.DB, registry workspace.NativeRegistry) (func() error, error) {
	enabled, err := relayEnabled()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return func() error { return nil }, nil
	}
	root, specs, window, err := relayConfiguration()
	if err != nil {
		return nil, err
	}
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		return nil, errors.New("isolated relay store unavailable")
	}
	sink := events.NewRelayEventSink(attention.NewPendingEventService(sqlite.NewEventRepository(db)), nil)
	neutral := relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: false})
	structured := agents.NewRelayNativeDriver(registry)
	terminal := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{})
	services.Relay = workspace.NewRelayService(store, services.Works, db, neutral, structured, terminal, sink, workspace.RelayServiceOptions{Agents: specs, LivenessWindow: window})
	return store.Close, nil
}
