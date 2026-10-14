package agents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type RelayNativeDriver struct{ registry app.NativeRegistry }

var _ app.RelayDriver = (*RelayNativeDriver)(nil)

func NewRelayNativeDriver(registry app.NativeRegistry) *RelayNativeDriver {
	return &RelayNativeDriver{registry: registry}
}

func (d *RelayNativeDriver) adapter(r app.RelayLaunch) (app.NativeAdapter, error) {
	if d.registry == nil {
		return nil, nativeError(apierrors.UnsupportedCapability, "native registry is not connected")
	}
	return d.registry.Adapter(r.Spec.Name)
}

func (d *RelayNativeDriver) request(r app.RelayLaunch) domain.NativeRequest {
	p := r.Project
	p.Root = r.Workspace.Root
	session := r.Session
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	session.ProjectID = p.ID
	session.CLI = r.Spec.Name
	packet := domain.ContextPacket{ID: uuid.NewString(), SchemaVersion: 1, ProjectID: p.ID, SettingsRevision: p.Settings.Revision, Mode: "selected_context", Materials: []domain.ContextMaterial{}, Files: []domain.ContextFile{}, CreatedAt: time.Now().UTC()}
	return domain.NativeRequest{SessionID: session.ID, Project: p, Session: session, Command: domain.NativeCommand{CLI: r.Spec.Name, Message: r.Input, Mode: "native"}, Packet: packet}
}

func (d *RelayNativeDriver) Launch(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, &domain.NativePrelaunchFailure{Cause: err}
	}
	// Registration can explicitly mark workspace tools unavailable. This is a
	// proven prelaunch decision, never an ambiguous delivery fallback.
	if cap, ok := r.Spec.Capabilities["workspace_tools"]; ok && cap.Status == "unsupported" {
		return app.RelayObservation{}, &domain.NativePrelaunchFailure{Cause: nativeError(apierrors.UnsupportedCapability, "registered structured path cannot provide workspace tools")}
	}
	session, err := a.Start(ctx, d.request(r))
	obs := app.RelayObservation{Session: session}
	if err != nil {
		return obs, err
	}
	native, err := a.Observe(ctx, session)
	return d.withArchive(ctx, a, relayNativeObservation(session, native)), err
}

// An optional trusted adapter port must return positively selected metadata.
// Existing adapters without this port stay unknown; HistoryPath alone is not
// a root/format/identity proof and triggers no filesystem or home discovery.
func (d *RelayNativeDriver) withArchive(ctx context.Context, a app.NativeAdapter, obs app.RelayObservation) app.RelayObservation {
	if source, ok := a.(app.RelayOwnedArchiveProvider); ok && obs.Session.Ownership == "owned" {
		archive, err := source.OwnedRelayArchive(ctx, obs.Session)
		if err == nil {
			obs.Archive = &archive
		}
	}
	return obs
}

func relayNativeObservation(s domain.NativeSession, obs domain.NativeObservation) app.RelayObservation {
	result := app.RelayObservation{Session: s, ExitConfirmed: obs.StopConfirmed, Events: []domain.RelayEvent{}}
	result.Session.Status = obs.Status
	result.Session.StopConfirmed = obs.StopConfirmed
	for _, v := range obs.Events {
		kind := "native_activity"
		switch v.Kind {
		case "turn_started", "turn_completed", "turn_failed", "turn_cancelled":
			kind = v.Kind
		}
		result.Events = append(result.Events, domain.RelayEvent{Kind: kind, Source: "native", Sequence: v.Sequence, TS: time.Now().UTC()})
	}
	return result
}

func (d *RelayNativeDriver) Send(ctx context.Context, r app.RelayLaunch, input string) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	obs, err := a.Send(ctx, r.Session, input)
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs)), err
}
func (d *RelayNativeDriver) Read(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	obs, err := a.Observe(ctx, r.Session)
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs)), err
}
func (d *RelayNativeDriver) Stop(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	obs, err := a.Stop(ctx, r.Session)
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs)), err
}
func (d *RelayNativeDriver) Resume(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	request := d.request(r)
	request.Command.Message = "" // original input is never replayed
	session, err := a.Resume(ctx, request)
	if err != nil {
		return app.RelayObservation{Session: session}, err
	}
	obs, err := a.Observe(ctx, session)
	return d.withArchive(ctx, a, relayNativeObservation(session, obs)), err
}
