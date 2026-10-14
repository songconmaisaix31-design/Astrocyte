package agents

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type RelayNativeDriver struct {
	registry                   app.NativeRegistry
	captureNativeNotifications bool
}

type RelayNativeDriverOptions struct{ CaptureNativeNotifications bool }

type relayNativeLifecycle interface {
	StartRelay(context.Context, domain.NativeRequest) (domain.NativeSession, error)
	ResumeRelay(context.Context, domain.NativeRequest) (domain.NativeSession, error)
	StopRelay(context.Context, domain.NativeSession) (domain.NativeObservation, error)
	OwnedRelayNativeCapture(context.Context, domain.NativeSession) (domain.RelayNativeSessionCapture, error)
}

var _ app.RelayDriver = (*RelayNativeDriver)(nil)

func NewRelayNativeDriver(registry app.NativeRegistry, options ...RelayNativeDriverOptions) *RelayNativeDriver {
	driver := &RelayNativeDriver{registry: registry}
	if len(options) != 0 {
		driver.captureNativeNotifications = options[0].CaptureNativeNotifications
	}
	return driver
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
	request := d.request(r)
	var session domain.NativeSession
	if lifecycle, ok := a.(relayNativeLifecycle); ok && d.captureNativeNotifications {
		session, err = lifecycle.StartRelay(ctx, request)
	} else {
		session, err = a.Start(ctx, request)
	}
	obs := app.RelayObservation{Session: session}
	if err != nil {
		return obs, err
	}
	native, err := a.Observe(ctx, session)
	if err != nil {
		return relayNativeObservation(session, native), err
	}
	return d.withArchive(ctx, a, relayNativeObservation(session, native))
}

// An optional trusted adapter port must return positively selected metadata.
// Existing adapters without this port stay unknown; HistoryPath alone is not
// a root/format/identity proof and triggers no filesystem or home discovery.
func (d *RelayNativeDriver) withArchive(ctx context.Context, a app.NativeAdapter, obs app.RelayObservation) (app.RelayObservation, error) {
	var captureErr error
	obs, captureErr = d.withNativeCapture(a, obs)
	if captureErr != nil {
		return obs, captureErr
	}
	if source, ok := a.(app.RelayOwnedArchiveProvider); ok && obs.Session.Ownership == "owned" {
		archive, err := source.OwnedRelayArchive(ctx, obs.Session)
		if err == nil {
			obs.Archive = &archive
		} else {
			obs.ArchiveGap = "archive_metadata_unavailable"
			var failure *apierrors.ServiceError
			if errors.As(err, &failure) {
				switch failure.Code {
				case apierrors.UnsupportedCapability:
					obs.ArchiveGap = "archive_metadata_unsupported"
				case apierrors.DeliveryUnknown:
					obs.ArchiveGap = "archive_metadata_unverifiable"
				case apierrors.ScopeDenied:
					obs.ArchiveGap = "archive_metadata_identity_unverified"
				case apierrors.EvidenceMissing:
					obs.ArchiveGap = "archive_metadata_path_unavailable"
				}
				if failure.Code == apierrors.UnsupportedCapability || failure.Code == apierrors.EvidenceMissing {
					return obs, nil
				}
			}
			// RPC、解码和身份错误立即返回；输入交付与来源读取分别保留。
			return obs, &domain.RelayArchiveObservationFailure{Cause: err}
		}
	} else if obs.Session.Ownership == "owned" {
		obs.ArchiveGap = "archive_metadata_unsupported"
	}
	return obs, nil
}

func (d *RelayNativeDriver) withNativeCapture(a app.NativeAdapter, obs app.RelayObservation) (app.RelayObservation, error) {
	if !d.captureNativeNotifications {
		return obs, nil
	}
	if source, ok := a.(relayNativeLifecycle); ok && obs.Session.Ownership == "owned" {
		// 这是当前进程已产生数据的纯复制，不发 RPC 或更新时间。
		capture, err := source.OwnedRelayNativeCapture(context.Background(), obs.Session)
		if err != nil {
			var failure *apierrors.ServiceError
			if errors.As(err, &failure) && failure.Code == apierrors.UnsupportedCapability {
				obs.NativeCapture = &domain.RelayNativeSessionCapture{Events: []domain.RelayNativeSessionEvent{}, SourceStatus: "unknown", PrefixComplete: "unknown", Finalized: "unknown", FirstGap: "native_notification_source_unsupported"}
				return obs, nil
			}
			if capture.SourceStatus != "" {
				obs.NativeCapture = &capture
			}
			return obs, &domain.RelayArchiveObservationFailure{Cause: err}
		}
		obs.NativeCapture = &capture
	}
	return obs, nil
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
	if err != nil {
		return relayNativeObservation(r.Session, obs), err
	}
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs))
}
func (d *RelayNativeDriver) Read(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	obs, err := a.Observe(ctx, r.Session)
	if err != nil {
		return relayNativeObservation(r.Session, obs), err
	}
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs))
}
func (d *RelayNativeDriver) Stop(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	var obs domain.NativeObservation
	if lifecycle, ok := a.(relayNativeLifecycle); ok && d.captureNativeNotifications {
		obs, err = lifecycle.StopRelay(ctx, r.Session)
	} else {
		obs, err = a.Stop(ctx, r.Session)
	}
	if err != nil {
		result, captureErr := d.withNativeCapture(a, relayNativeObservation(r.Session, obs))
		if captureErr != nil {
			return result, errors.Join(err, captureErr)
		}
		return result, err
	}
	return d.withArchive(ctx, a, relayNativeObservation(r.Session, obs))
}
func (d *RelayNativeDriver) Resume(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	a, err := d.adapter(r)
	if err != nil {
		return app.RelayObservation{}, err
	}
	request := d.request(r)
	request.Command.Message = "" // original input is never replayed
	var session domain.NativeSession
	if lifecycle, ok := a.(relayNativeLifecycle); ok && d.captureNativeNotifications {
		session, err = lifecycle.ResumeRelay(ctx, request)
	} else {
		session, err = a.Resume(ctx, request)
	}
	if err != nil {
		return app.RelayObservation{Session: session}, err
	}
	obs, err := a.Observe(ctx, session)
	if err != nil {
		return relayNativeObservation(session, obs), err
	}
	return d.withArchive(ctx, a, relayNativeObservation(session, obs))
}
