package app

import (
	"context"
	"reflect"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayIntentPreparer = (*relayService)(nil)

// 首输入只消费原 prepare 的控制器；不调用 Launch 或恢复旧进程。
func (s *relayService) submitPreparedIntent(ctx context.Context, caller domain.Caller, receipt domain.RelayReceipt, task domain.RelayTask, work domain.Work, project domain.LocalProject) (domain.RelayActionResult, error) {
	reject := func(err error) (domain.RelayActionResult, error) {
		return s.finish(ctx, caller, receipt, RelayObservation{}, err, true)
	}
	if task.State != domain.RelayReady || task.Condition == "delivery_unknown" {
		return reject(relayError(apierrors.VersionConflict, "task already has a dispatch or unresolved delivery"))
	}
	launch, err := s.launch(ctx, caller, task.ID, "start")
	if err != nil {
		return reject(err)
	}
	if launch.Terminal.DriveMode != "structured" || launch.Session.Ownership != "owned" || launch.Session.NativeID == "" || launch.ControllerID == "" || launch.Terminal.ExitedAt != nil {
		return reject(relayError(apierrors.DeliveryUnknown, "original initialized controller is unavailable"))
	}
	provider, ok := s.structured.(RelayControllerEvidenceReader)
	if !ok {
		return reject(relayError(apierrors.UnsupportedCapability, "original controller evidence reader is unavailable"))
	}
	evidence, err := provider.CurrentControllerEvidence(ctx, launch)
	if err != nil {
		return reject(err)
	}
	if evidence.ControllerID != launch.ControllerID || evidence.SessionID != launch.Session.ID || evidence.NativeID != launch.Session.NativeID || evidence.ExitConfirmed || len(evidence.Turns) != 0 {
		return reject(relayError(apierrors.DeliveryUnknown, "initialized controller identity or input state differs"))
	}
	if receipt.OperationID == evidence.StartupOperationID {
		return reject(relayError(apierrors.VersionConflict, "first input requires its own operation identity"))
	}
	confirmed := false
	for _, fact := range evidence.Facts {
		if fact.Kind == "session_confirmed" && fact.Status == "confirmed" && fact.OperationID == evidence.StartupOperationID {
			confirmed = true
		}
		if fact.Kind == "input_written" || fact.Kind == "input_accepted" || fact.Kind == "turn_started" {
			return reject(relayError(apierrors.DeliveryUnknown, "initialized controller already has an input fact"))
		}
	}
	if !confirmed {
		return reject(relayError(apierrors.EvidenceMissing, "original native session initialization is unconfirmed"))
	}
	if s.options.LaunchContext != nil {
		if _, err = s.permission(ctx, caller, task, "read_context"); err != nil {
			return reject(err)
		}
		packet, err := s.options.LaunchContext.ReadRelayLaunchContext(ctx, caller, launch)
		if err != nil {
			return reject(err)
		}
		if packet.ProjectID != project.ID || packet.SettingsRevision != project.Settings.Revision {
			return reject(relayError(apierrors.ScopeDenied, "fixed input context differs from the confirmed project"))
		}
		current := relaySessionMetadata(domain.NativeSession{ContextPacket: packet}).ContextPacket
		original := launch.Session.ContextPacket
		current.ID, current.CreatedAt = original.ID, original.CreatedAt
		if !reflect.DeepEqual(current, original) {
			return reject(relayError(apierrors.ContextStale, "fixed context changed since initialization"))
		}
	}
	if _, err = s.permission(ctx, caller, task, "start"); err != nil {
		return reject(err)
	}
	err = s.store.Update(ctx, func(state *RelayState) error {
		current := state.Tasks[task.ID]
		session := state.Sessions[task.TerminalID]
		controller := state.ControllerEvidence[task.ID]
		if state.PendingActions[task.ID] != "" || current.State != domain.RelayReady || current.Condition == "delivery_unknown" || current.TerminalID != task.TerminalID || current.WorktreeID != task.WorktreeID || state.Workspaces[task.WorktreeID].ActiveTaskID != task.ID || session.ID != launch.Session.ID || session.NativeID != launch.Session.NativeID || controller.ControllerID != evidence.ControllerID || controller.StartupOperationID != evidence.StartupOperationID || state.Terminals[task.TerminalID].ExitedAt != nil {
			return relayError(apierrors.VersionConflict, "prepared task reservation or controller changed before first input")
		}
		prepared := false
		for _, previous := range state.Receipts {
			if previous.TaskID == task.ID && previous.Action == "prepare" && previous.OperationID == evidence.StartupOperationID && previous.Status == "accepted" {
				prepared = true
			}
		}
		if !prepared {
			return relayError(apierrors.DeliveryUnknown, "original initialization receipt is not accepted")
		}
		state.PendingActions[task.ID] = receipt.ID
		s.claimFact(state, receipt)
		return nil
	})
	if err != nil {
		return reject(err)
	}
	launch.OperationID = receipt.OperationID
	input := "HANDOFF_MANUAL:\n" + work.Handoff.Content + "\n\nTASK:\n" + task.Description
	observation, sendErr := s.structured.Send(ctx, launch, input)
	if relayNativeEffectAccepted(sendErr) {
		if err := s.store.Update(context.WithoutCancel(ctx), func(state *RelayState) error {
			current := state.Tasks[task.ID]
			if state.PendingActions[task.ID] != receipt.ID || current.State != domain.RelayReady {
				return relayError(apierrors.VersionConflict, "first input reservation changed during submission")
			}
			s.move(state, &current, domain.RelayAccepted, relayActor(caller))
			return nil
		}); err != nil {
			return domain.RelayActionResult{}, relayError(apierrors.DeliveryUnknown, "first input state could not be persisted; do not replay")
		}
	}
	return s.finish(ctx, caller, receipt, observation, sendErr, false)
}
