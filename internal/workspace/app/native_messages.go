package app

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type nativeMessageOwnedReader interface {
	ReadOwnedNativeSession(context.Context, string) (domain.RelayOwnedNativeSession, error)
}

// 此记录只在主数据库 adapter 与可信应用之间传递。
type NativeMessagePreparation struct {
	Input     domain.RelayPreparedInput `json:"-"`
	ObjectRef string                    `json:"-"`
	Phase     string                    `json:"-"`
	Scope     NativeMessageScope        `json:"-"`
	Lease     NativeMessageLease        `json:"-"`
}

type NativeMessageRepository interface {
	RequireNativeMessageOutsideTx(context.Context) error
	WithNativeMessageTx(context.Context, func(context.Context) error) error
	LoadNativeMessagePreparation(context.Context, string, string) (NativeMessagePreparation, error)
	SaveNativeMessagePreparation(context.Context, domain.NativeSession, NativeMessagePreparation, string) error
}

type nativeMessageOptions struct {
	objects    ContextObjectStore
	owned      nativeMessageOwnedReader
	authorizer NativeMessageAuthorizer
}

func (s *localProjectService) WithNativeMessageObjects(objects ContextObjectStore) *localProjectService {
	s.nativeMessageMu.Lock()
	defer s.nativeMessageMu.Unlock()
	s.nativeMessageOptions.objects = objects
	return s
}

func (s *localProjectService) WithOwnedNativeSessions(owned nativeMessageOwnedReader) *localProjectService {
	s.nativeMessageMu.Lock()
	defer s.nativeMessageMu.Unlock()
	s.nativeMessageOptions.owned = owned
	return s
}

func (s *localProjectService) nativeMessageServices() (NativeMessageRepository, nativeMessageOptions, error) {
	s.nativeMessageMu.RLock()
	options := s.nativeMessageOptions
	s.nativeMessageMu.RUnlock()
	repository, ok := s.repo.(NativeMessageRepository)
	if !ok || options.objects == nil || options.owned == nil || options.authorizer == nil {
		return nil, options, projectError(apierrors.UnsupportedCapability, "actual native message repository, original controller reader, object store and send authorizer are required")
	}
	return repository, options, nil
}

func validateOwnedNativeMessage(project domain.LocalProject, session domain.NativeSession, original domain.RelayOwnedNativeSession) error {
	e := original.Evidence
	o := original.Session
	if session.ID == "" || session.NativeID == "" || session.Ownership != "owned" || session.StopConfirmed || e.ExitConfirmed {
		return projectError(apierrors.DeliveryUnknown, "native message preparation requires the original live owned controller")
	}
	if o.ID != session.ID || o.NativeID != session.NativeID || o.ProjectID != session.ProjectID || o.CLI != session.CLI || o.Version != session.Version || o.Ownership != "owned" || e.SessionID != session.ID || e.NativeID != session.NativeID || e.Provider != session.CLI || e.WorkspaceRoot != project.Root || e.ControllerID == "" || e.StartupOperationID == "" {
		return projectError(apierrors.ScopeDenied, "original native controller and registered project tuple differ")
	}
	if session.ContextPacket.ID == "" || session.ContextPacket.ProjectID != project.ID || session.ContextPacket.SettingsRevision != project.Settings.Revision || !reflect.DeepEqual(nativeContextMetadata(o.ContextPacket), nativeContextMetadata(session.ContextPacket)) {
		return projectError(apierrors.ContextStale, "fixed native context or project settings changed")
	}
	return nil
}

func (s *localProjectService) PrepareNativeMessage(ctx context.Context, caller domain.Caller, projectID, sessionID, message string) (domain.NativeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prepareNativeMessage(ctx, caller, projectID, sessionID, message)
}

func (s *localProjectService) prepareNativeMessage(ctx context.Context, caller domain.Caller, projectID, sessionID, message string) (domain.NativeSession, error) {
	return s.prepareNativeMessageForScope(ctx, caller, projectID, sessionID, message, NativeMessageScope{Kind: "project"})
}

func (s *localProjectService) prepareNativeMessageForScope(ctx context.Context, caller domain.Caller, projectID, sessionID, message string, scope NativeMessageScope) (domain.NativeSession, error) {
	var result domain.NativeSession
	repository, options, err := s.nativeMessageServices()
	if err != nil {
		return result, err
	}
	if err := repository.RequireNativeMessageOutsideTx(ctx); err != nil {
		return result, err
	}
	if err := ValidateNativeMessageScope(scope); err != nil {
		return result, err
	}
	op, err := operationID(caller, domain.NativeCommand{})
	if err != nil {
		return result, err
	}
	if strings.TrimSpace(message) == "" || len(message) > 64*1024 || !utf8.ValidString(message) {
		return result, projectError(apierrors.ValidationFailed, "native message is empty, invalid UTF-8 or exceeds 64 KiB")
	}
	project, session, _, err := s.session(ctx, caller, projectID, sessionID, "send")
	if err != nil {
		return result, err
	}
	if err := s.nativePermission(ctx, caller, project, session.CLI, "send"); err != nil {
		return result, err
	}
	if err := s.checkPacket(ctx, caller, project, session.ContextPacket); err != nil {
		return result, err
	}
	owned, err := options.owned.ReadOwnedNativeSession(ctx, session.ID)
	if err != nil {
		return result, err
	}
	if err := validateOwnedNativeMessage(project, session, owned); err != nil {
		return result, err
	}
	if _, exists := session.Operations[op]; exists {
		previous, err := repository.LoadNativeMessagePreparation(ctx, sessionID, op)
		if err != nil {
			return result, err
		}
		if err := preparedNativeMessageCaller(previous, caller); err != nil {
			return result, err
		}
		if previous.Scope != scope {
			return result, projectError(apierrors.ScopeDenied, "original native input has a different fixed purpose")
		}
		if err := validatePreparedNativeMessage(project, session, previous, owned); err != nil {
			return result, err
		}
		original, err := options.objects.Read(ctx, previous.ObjectRef)
		if err != nil {
			return result, err
		}
		if !bytes.Equal(original, []byte(message)) {
			return result, projectError(apierrors.VersionConflict, "operation identity already binds a different fixed native input")
		}
		err = options.authorizer.WithNativeMessagePreparation(ctx, NativeMessageAuthorization{Caller: caller, Preparation: previous}, func(txContext context.Context) error {
			currentProject, current, _, err := s.session(txContext, caller, projectID, sessionID, "send")
			if err != nil {
				return err
			}
			if err := s.nativePermission(txContext, caller, currentProject, current.CLI, "send"); err != nil {
				return err
			}
			currentPrepared, err := repository.LoadNativeMessagePreparation(txContext, sessionID, op)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(currentPrepared, previous) || !reflect.DeepEqual(current, session) || !reflect.DeepEqual(currentProject, project) {
				return projectError(apierrors.ContextStale, "original native preparation, project or session changed during authorization")
			}
			if err := validatePreparedNativeMessage(currentProject, current, currentPrepared, owned); err != nil {
				return err
			}
			result = current
			return nil
		})
		return result, err
	}
	if session.PendingOperation != "" || len(session.Operations) >= 64 {
		return result, projectError(apierrors.DeliveryUnknown, "native session has pending delivery or exhausted operation receipts")
	}
	object, err := options.objects.Publish(ctx, []byte(message))
	if err != nil {
		return result, err
	}
	now := time.Now().UTC()
	candidate := NativeMessagePreparation{Input: domain.RelayPreparedInput{
		Session: session, OperationID: op, ActorID: caller.ID, ActorKind: caller.Kind, ActorProjectID: caller.ProjectID,
		ControllerID: owned.Evidence.ControllerID, StartupOperationID: owned.Evidence.StartupOperationID,
		WorkspaceRoot: owned.Evidence.WorkspaceRoot, PreparedAt: now,
	}, ObjectRef: object, Phase: "prepared", Scope: scope}
	err = options.authorizer.WithNativeMessagePreparation(ctx, NativeMessageAuthorization{Caller: caller, Preparation: candidate}, func(txContext context.Context) error {
		currentProject, current, _, err := s.session(txContext, caller, projectID, sessionID, "send")
		if err != nil {
			return err
		}
		if err := s.nativePermission(txContext, caller, currentProject, current.CLI, "send"); err != nil {
			return err
		}
		if currentProject.Settings.Revision != project.Settings.Revision || currentProject.Root != project.Root {
			return projectError(apierrors.ContextStale, "native session or project changed before preparation")
		}
		if _, exists := current.Operations[op]; exists {
			previous, err := repository.LoadNativeMessagePreparation(txContext, sessionID, op)
			if err != nil {
				return err
			}
			if previous.Phase != "prepared" || previous.ObjectRef != object || previous.Scope != scope || previous.Input.ActorID != caller.ID || previous.Input.ActorKind != caller.Kind || previous.Input.ActorProjectID != caller.ProjectID {
				return projectError(apierrors.DeliveryUnknown, "original native input is already attempted or binds different fixed input")
			}
			if err := validatePreparedNativeMessage(currentProject, current, previous, owned); err != nil {
				return err
			}
			result = current
			return nil
		}
		if !reflect.DeepEqual(current, session) {
			return projectError(apierrors.ContextStale, "native session changed before preparation")
		}
		if current.PendingOperation != "" || len(current.Operations) >= 64 {
			return projectError(apierrors.DeliveryUnknown, "native session has pending delivery or exhausted operation receipts")
		}
		current.PendingOperation, current.LastOperationID, current.UpdatedAt = op, op, now
		if current.Operations == nil {
			current.Operations = map[string]domain.NativeOperation{}
		}
		current.Operations[op] = domain.NativeOperation{Action: "send", Status: "pending", CreatedAt: now}
		prepared := candidate
		prepared.Input.Session = current
		if err := repository.SaveNativeMessagePreparation(txContext, current, prepared, ""); err != nil {
			return err
		}
		result = current
		return nil
	})
	return result, err
}

func preparedNativeMessageCaller(prepared NativeMessagePreparation, caller domain.Caller) error {
	input := prepared.Input
	if input.ActorID != caller.ID || input.ActorKind != caller.Kind || input.ActorProjectID != caller.ProjectID || input.OperationID != caller.OperationID {
		return projectError(apierrors.ScopeDenied, "fixed native preparation belongs to a different authenticated operation")
	}
	return nil
}

// 对象读取与原生发送均在事务外；已占用的原准备不会再次发送。
func (s *localProjectService) sendPreparedNativeMessage(ctx context.Context, caller domain.Caller, projectID, sessionID, message string) (domain.NativeObservation, error) {
	repository, options, err := s.nativeMessageServices()
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if err := repository.RequireNativeMessageOutsideTx(ctx); err != nil {
		return domain.NativeObservation{}, err
	}
	operation, err := operationID(caller, domain.NativeCommand{})
	if err != nil {
		return domain.NativeObservation{}, err
	}
	project, session, adapter, err := s.session(ctx, caller, projectID, sessionID, "send")
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if err := s.nativePermission(ctx, caller, project, session.CLI, "send"); err != nil {
		return domain.NativeObservation{}, err
	}
	if err := s.checkPacket(ctx, caller, project, session.ContextPacket); err != nil {
		return domain.NativeObservation{}, err
	}
	if _, exists := session.Operations[operation]; !exists {
		session, err = s.prepareNativeMessage(ctx, caller, projectID, sessionID, message)
		if err != nil {
			return domain.NativeObservation{}, err
		}
	}
	prepared, err := repository.LoadNativeMessagePreparation(ctx, sessionID, operation)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if err := preparedNativeMessageCaller(prepared, caller); err != nil {
		return domain.NativeObservation{}, err
	}
	receipt := session.Operations[operation]
	accepted := prepared.Phase == "send_attempted" && receipt.Action == "send" && receipt.Status == "accepted" && receipt.FinishedAt != nil
	// 已受理操作只允许核原输入后观察；在途和未知操作直接拒绝。
	if prepared.Phase != "prepared" && !accepted {
		return domain.NativeObservation{}, projectError(apierrors.DeliveryUnknown, "original native input was already claimed for sending; inspect its receipt")
	}
	input, err := options.objects.Read(ctx, prepared.ObjectRef)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if !bytes.Equal(input, []byte(message)) {
		return domain.NativeObservation{}, projectError(apierrors.VersionConflict, "native send differs from its fixed prepared input")
	}
	if strings.TrimSpace(string(input)) == "" || len(input) > 64*1024 || !utf8.Valid(input) {
		return domain.NativeObservation{}, projectError(apierrors.ValidationFailed, "fixed native input is empty, invalid UTF-8 or exceeds 64 KiB")
	}
	if accepted {
		return adapter.Observe(ctx, session)
	}
	owned, err := options.owned.ReadOwnedNativeSession(ctx, sessionID)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if err := validatePreparedNativeMessage(project, session, prepared, owned); err != nil {
		return domain.NativeObservation{}, err
	}
	var sendDeadline time.Time
	err = options.authorizer.WithNativeMessageSend(ctx, NativeMessageAuthorization{Caller: caller, Preparation: prepared}, func(txContext context.Context) error {
		if err := txContext.Err(); err != nil {
			return err
		}
		currentProject, current, _, err := s.session(txContext, caller, projectID, sessionID, "send")
		if err != nil {
			return err
		}
		if err := s.nativePermission(txContext, caller, currentProject, current.CLI, "send"); err != nil {
			return err
		}
		currentPrepared, err := repository.LoadNativeMessagePreparation(txContext, sessionID, operation)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(currentPrepared, prepared) || !reflect.DeepEqual(current, session) || !reflect.DeepEqual(currentProject, project) {
			return projectError(apierrors.ContextStale, "native preparation, project or session changed before its single send claim")
		}
		if err := validatePreparedNativeMessage(currentProject, current, currentPrepared, owned); err != nil {
			return err
		}
		currentPrepared.Phase = "send_attempted"
		sendDeadline, _ = txContext.Deadline()
		return repository.SaveNativeMessagePreparation(txContext, current, currentPrepared, "prepared")
	})
	if err != nil {
		return domain.NativeObservation{}, err
	}
	sendContext := ctx
	if !sendDeadline.IsZero() {
		var cancel context.CancelFunc
		sendContext, cancel = context.WithDeadline(ctx, sendDeadline)
		defer cancel()
	}
	if err := sendContext.Err(); err != nil {
		return domain.NativeObservation{}, err
	}
	observation, sendErr := adapter.Send(sendContext, session, string(input))
	finishErr := repository.WithNativeMessageTx(context.WithoutCancel(ctx), func(txContext context.Context) error {
		current, err := s.repo.LoadSession(txContext, sessionID)
		if err != nil {
			return err
		}
		currentReceipt, exists := current.Operations[operation]
		if current.ProjectID != projectID || current.NativeID != session.NativeID || current.LastOperationID != operation || !exists || currentReceipt.Action != "send" || currentReceipt.Status != "pending" || !currentReceipt.CreatedAt.Equal(prepared.Input.PreparedAt) {
			return projectError(apierrors.DeliveryUnknown, "native send returned after its original pending session changed; reconcile the original operation")
		}
		now := time.Now().UTC()
		currentReceipt.FinishedAt, currentReceipt.Status = &now, "accepted"
		if sendErr != nil {
			currentReceipt.Status = "unknown"
		}
		current.Operations[operation], current.UpdatedAt = currentReceipt, now
		current.StopConfirmed = current.StopConfirmed || observation.StopConfirmed
		if current.StopConfirmed {
			current.Status = "stopped"
		} else if sendErr != nil {
			current.Status = "unknown"
		} else {
			current.Status = observation.Status
		}
		if sendErr == nil && current.PendingOperation == operation {
			current.PendingOperation = ""
		}
		return s.repo.SaveSession(txContext, current)
	})
	return observation, errors.Join(sendErr, finishErr)
}

func validatePreparedNativeMessage(project domain.LocalProject, session domain.NativeSession, prepared NativeMessagePreparation, owned domain.RelayOwnedNativeSession) error {
	input := prepared.Input
	receipt, exists := session.Operations[input.OperationID]
	if prepared.Phase != "prepared" || prepared.ObjectRef == "" || input.PreparedAt.IsZero() || !exists || receipt.Action != "send" || receipt.Status != "pending" || receipt.FinishedAt != nil || !receipt.CreatedAt.Equal(input.PreparedAt) || session.PendingOperation != input.OperationID || session.LastOperationID != input.OperationID {
		return projectError(apierrors.DeliveryUnknown, "native input is not an actual unattempted preparation")
	}
	if err := validateOwnedNativeMessage(project, session, owned); err != nil {
		return err
	}
	if input.Session.ID != session.ID || input.Session.NativeID != session.NativeID || input.Session.ProjectID != session.ProjectID || input.Session.CLI != session.CLI || input.Session.Version != session.Version || input.Session.Mode != session.Mode || input.Session.SourceSessionID != session.SourceSessionID || input.ControllerID != owned.Evidence.ControllerID || input.StartupOperationID != owned.Evidence.StartupOperationID || input.WorkspaceRoot != owned.Evidence.WorkspaceRoot || !reflect.DeepEqual(input.Session.ContextPacket, session.ContextPacket) {
		return projectError(apierrors.ContextStale, "original native preparation tuple or fixed context changed")
	}
	return nil
}

// 认领事务只读取原准备和当前主数据库权限，不读取输入对象或启动控制器。
func (s *localProjectService) ReadPreparedNativeSession(ctx context.Context, sessionID, operationID string) (domain.RelayPreparedInput, error) {
	repository, options, err := s.nativeMessageServices()
	if err != nil {
		return domain.RelayPreparedInput{}, err
	}
	prepared, err := repository.LoadNativeMessagePreparation(ctx, sessionID, operationID)
	if err != nil {
		return domain.RelayPreparedInput{}, err
	}
	caller := domain.Caller{Kind: prepared.Input.ActorKind, ID: prepared.Input.ActorID, ProjectID: prepared.Input.ActorProjectID, OperationID: operationID}
	project, session, _, err := s.session(ctx, caller, prepared.Input.Session.ProjectID, sessionID, "send")
	if err != nil {
		return domain.RelayPreparedInput{}, err
	}
	if err := s.nativePermission(ctx, caller, project, session.CLI, "send"); err != nil {
		return domain.RelayPreparedInput{}, err
	}
	owned, err := options.owned.ReadOwnedNativeSession(ctx, sessionID)
	if err != nil {
		return domain.RelayPreparedInput{}, err
	}
	if prepared.Input.OperationID != operationID {
		return domain.RelayPreparedInput{}, projectError(apierrors.ScopeDenied, "native preparation belongs to another original operation")
	}
	if err := validatePreparedNativeMessage(project, session, prepared, owned); err != nil {
		return domain.RelayPreparedInput{}, err
	}
	prepared.Input.Session = session
	return prepared.Input, nil
}
