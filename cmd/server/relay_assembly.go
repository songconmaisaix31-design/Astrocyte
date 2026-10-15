package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// This entrypoint owns cross-context assembly. The runtime remains the sole
// task/terminal/worktree writer; event observations enter the existing app port.
func assembleRelay(services *httpapi.Services, db *sqlite.DB, registry workspace.NativeRegistry, contextReaders ...workspace.RelayLaunchContextReader) (func() error, error) {
	if len(contextReaders) > 1 || len(contextReaders) == 1 && contextReaders[0] == nil {
		return nil, errors.New("relay launch context requires one explicit trusted reader")
	}
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
	deadlineSeconds, err := relayNativeDeadlineSeconds()
	if err != nil {
		return nil, err
	}
	store, err := sqlite.OpenRelayStore(root)
	if err != nil {
		return nil, errors.New("isolated relay store unavailable")
	}
	identity, err := os.Stat(root)
	if err != nil || !identity.IsDir() {
		return nil, errors.Join(errors.New("trusted registration root unavailable"), store.Close())
	}
	sink := events.NewRelayEventSink(attention.NewPendingEventService(sqlite.NewEventRepository(db)), nil)
	neutral, err := relayWorkspaceProvider(root)
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	notifications, err := relayNativeNotifications()
	if err != nil {
		return nil, errors.Join(err, store.Close())
	}
	sourceGuard := newRelaySourceGuard(registry)
	structured := agents.NewRelayNativeDriver(registry, agents.RelayNativeDriverOptions{CaptureNativeNotifications: true, SessionDeadlineSeconds: deadlineSeconds, BeforeFirstInput: sourceGuard.Check})
	terminal := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{})
	options := workspace.RelayServiceOptions{Agents: specs, LivenessWindow: window, OwnedRoots: agents.RelayOwnedRoots{}}
	if len(contextReaders) == 1 {
		options.LaunchContext = contextReaders[0]
	}
	runtime := workspace.NewRelayService(store, services.Works, db, neutral, structured, terminal, sink, options)
	owned, ok := runtime.(workspace.RelayOwnedSessionReader)
	if !ok {
		_ = store.Close()
		return nil, errors.New("owned session read port unavailable")
	}
	lifecycle, ok := runtime.(workspace.RelayLifecycle)
	if !ok {
		_ = store.Close()
		return nil, errors.New("owned session lifecycle unavailable")
	}
	collector := newRelaySessionCollector(owned, events.NewRelaySessionCapture(attention.NewPendingEventService(sqlite.NewEventRepository(db)), nil), relaySessionFileState{root: root}, sqlite.NewAccessSessionReader(db))
	collector.archiveOnly = !notifications
	services.Relay = &relaySessionRuntime{RelayRuntime: runtime, RelayLifecycle: lifecycle, collector: collector, sourceGuard: sourceGuard, registration: &relayTrustedRegistration{root: root, identity: identity, window: window}}
	services.RelaySessionEvents = collector
	return func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var shutdownErr error
		if lifecycle, ok := services.Relay.(workspace.RelayLifecycle); ok {
			shutdownErr = lifecycle.Shutdown(shutdownCtx)
		}
		return errors.Join(shutdownErr, store.Close())
	}, nil
}
