package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayCompletionEvidenceReader = (*relayService)(nil)

func (s *relayService) ReadRelayCompletion(ctx context.Context, projectID, completionID string) (domain.RelayCompletion, error) {
	if !relayCompletionReference(projectID) || !relayCompletionReference(completionID) {
		return domain.RelayCompletion{}, relayError(apierrors.ValidationFailed, "actual project and completion references are required")
	}
	if s.store == nil {
		return domain.RelayCompletion{}, relayError(apierrors.UnsupportedCapability, "relay store is not connected")
	}
	state, err := s.store.Snapshot(ctx)
	if err != nil {
		return domain.RelayCompletion{}, err
	}
	for taskID, completion := range state.Completions {
		if completion.ID != completionID {
			continue
		}
		task, found := state.Tasks[taskID]
		if !found || task.ProjectID != projectID || completion.TaskID != task.ID || completion.WorktreeID != task.WorktreeID || completion.TerminalID != task.TerminalID {
			return domain.RelayCompletion{}, relayError(apierrors.ScopeDenied, "completion does not belong to the original project and task")
		}
		if _, err = s.permission(ctx, domain.Caller{}, task, "observe"); err != nil {
			return domain.RelayCompletion{}, err
		}
		receipt, found := state.Receipts[completion.ReceiptID]
		if !found || receipt.TaskID != task.ID || receipt.Action != "transition" || receipt.OperationID != completion.OperationID || receipt.Operator != completion.Operator || receipt.Status != "accepted" || receipt.FinishedAt == nil || completion.TS.IsZero() || (task.State != domain.RelayDone && task.State != domain.RelayReleased) {
			return domain.RelayCompletion{}, relayError(apierrors.DeliveryUnknown, "completion has no finalized original transition receipt")
		}
		if completion.Channel != domain.RelayUserConfirmedComplete && completion.Channel != domain.RelayVerifiedComplete {
			return domain.RelayCompletion{}, relayError(apierrors.EvidenceMissing, "original completion channel is unrecorded")
		}
		return cloneRelayCompletion(completion), nil
	}
	return domain.RelayCompletion{}, apierrors.NewNotFound("relay completion", completionID)
}

func cloneRelayCompletion(value domain.RelayCompletion) domain.RelayCompletion {
	value.OutputEventIDs = append([]string(nil), value.OutputEventIDs...)
	value.CheckEvidenceIDs = append([]string(nil), value.CheckEvidenceIDs...)
	if value.VerifiedAt != nil {
		when := *value.VerifiedAt
		value.VerifiedAt = &when
	}
	return value
}

func relayCompletionReference(value string) bool {
	if value == "" || len(value) > 160 {
		return false
	}
	for _, character := range value {
		if character != '-' && character != '_' && character != ':' && !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'z') && !(character >= 'A' && character <= 'Z') {
			return false
		}
	}
	return true
}

func (s *relayService) prepareCompletion(ctx context.Context, caller domain.Caller, task domain.RelayTask, cmd RelayTransitionCommand, receipt domain.RelayReceipt) (domain.RelayCompletion, *domain.RelayGoalVerification, error) {
	var result domain.RelayCompletion
	if cmd.To != domain.RelayDone {
		if cmd.CompletionChannel != "" || cmd.CompletionEvidenceID != "" {
			return result, nil, relayError(apierrors.ValidationFailed, "completion references require a done transition")
		}
		return result, nil, nil
	}
	if task.State != cmd.From || !domain.ValidRelayTransition(cmd.From, cmd.To) {
		return result, nil, relayError(apierrors.VersionConflict, "invalid or stale task transition")
	}
	if !cmd.CompletionConfirmed {
		return result, nil, relayError(apierrors.ApprovalRequired, "human task completion confirmation is required")
	}
	channel := cmd.CompletionChannel
	if channel == "" {
		channel = domain.RelayUserConfirmedComplete
	}
	if channel != domain.RelayUserConfirmedComplete && channel != domain.RelayVerifiedComplete {
		return result, nil, relayError(apierrors.ValidationFailed, "unsupported completion channel")
	}
	result = domain.RelayCompletion{ID: uuid.NewString(), TaskID: task.ID, WorktreeID: task.WorktreeID, TerminalID: task.TerminalID, Channel: channel, Source: "human_confirmation", Operator: relayActor(caller), ActorID: caller.ID, ActorKind: caller.Kind, ReceiptID: receipt.ID, OperationID: receipt.OperationID}
	if channel == domain.RelayUserConfirmedComplete {
		if cmd.CompletionEvidenceID != "" {
			return domain.RelayCompletion{}, nil, relayError(apierrors.ValidationFailed, "human confirmation cannot claim verified goal evidence")
		}
		return result, nil, nil
	}
	if !relayCompletionReference(cmd.CompletionEvidenceID) {
		return domain.RelayCompletion{}, nil, relayError(apierrors.EvidenceMissing, "verified completion requires an actual goal evidence reference")
	}
	if s.options.GoalVerification == nil {
		return domain.RelayCompletion{}, nil, relayError(apierrors.EvidenceMissing, "owned output and fixed goal checks are not connected")
	}
	verification, err := s.options.GoalVerification.ReadRelayGoalVerification(ctx, caller, task.ID, cmd.CompletionEvidenceID)
	if err != nil {
		return domain.RelayCompletion{}, nil, relayCompletionEvidenceError(err)
	}
	if verification.EvidenceID != cmd.CompletionEvidenceID || verification.TaskID != task.ID || verification.WorktreeID != task.WorktreeID || verification.TerminalID != task.TerminalID || verification.PassedAt.IsZero() || verification.PassedAt.After(s.now()) || len(verification.OutputEventIDs) == 0 || len(verification.CheckEvidenceIDs) == 0 {
		return domain.RelayCompletion{}, nil, relayError(apierrors.EvidenceMissing, "fixed goal verification is incomplete or belongs to another task")
	}
	for _, references := range [][]string{verification.OutputEventIDs, verification.CheckEvidenceIDs} {
		seen := map[string]bool{}
		for _, reference := range references {
			if !relayCompletionReference(reference) || seen[reference] {
				return domain.RelayCompletion{}, nil, relayError(apierrors.EvidenceMissing, "fixed goal verification references are invalid or duplicated")
			}
			seen[reference] = true
		}
	}
	verification.OutputEventIDs = append([]string{}, verification.OutputEventIDs...)
	verification.CheckEvidenceIDs = append([]string{}, verification.CheckEvidenceIDs...)
	result.Source, result.EvidenceID = "owned_output_and_goal_checks", verification.EvidenceID
	result.InputOperationID = verification.InputOperationID
	result.OutputEventIDs, result.CheckEvidenceIDs = verification.OutputEventIDs, verification.CheckEvidenceIDs
	when := verification.PassedAt.UTC()
	result.VerifiedAt = &when
	return result, &verification, nil
}

func (s *relayService) checkGoalVerification(state RelayState, task domain.RelayTask, verification domain.RelayGoalVerification) error {
	if verification.TaskID != task.ID || verification.WorktreeID != task.WorktreeID || verification.TerminalID != task.TerminalID {
		return relayError(apierrors.VersionConflict, "task execution identity changed before completion")
	}
	if verification.ControllerID == "" || verification.StartupOperationID == "" || verification.SessionID == "" || verification.NativeID == "" || verification.InputOperationID == "" || verification.TurnID == "" {
		return relayError(apierrors.EvidenceMissing, "verified completion lacks the original accepted input controller tuple")
	}
	controllers := append([]domain.RelayControllerEvidence{}, state.ControllerHistory[task.ID]...)
	controllers = append(controllers, state.ControllerEvidence[task.ID])
	matched := false
	for _, controller := range controllers {
		if controller.ControllerID != verification.ControllerID || controller.StartupOperationID != verification.StartupOperationID || controller.SessionID != verification.SessionID || controller.NativeID != verification.NativeID || controller.Provider != task.Agent || !sameRelayRoot(controller.WorkspaceRoot, state.Workspaces[task.WorktreeID].Root) {
			continue
		}
		for _, turn := range controller.Turns {
			if turn.OperationID == verification.InputOperationID && turn.TurnID == verification.TurnID && turn.ControllerID == controller.ControllerID && turn.StartupOperationID == controller.StartupOperationID && turn.SessionID == controller.SessionID && turn.NativeID == controller.NativeID && !turn.AcceptedAt.IsZero() && !verification.PassedAt.Before(turn.AcceptedAt) {
				matched = true
			}
		}
	}
	if !matched {
		return relayError(apierrors.ScopeDenied, "verified output differs from the task's original accepted input")
	}
	for _, receipt := range state.Receipts {
		if receipt.TaskID == task.ID && receipt.OperationID == verification.InputOperationID && receipt.Status == "accepted" && (receipt.Action == "dispatch" || receipt.Action == "send") {
			return nil
		}
	}
	return relayError(apierrors.EvidenceMissing, "verified output has no original accepted input receipt")
}

func relayCompletionEvidenceError(err error) error {
	// 外部核验错误不携带原生正文、路径或实际检查命令。
	code := apierrors.EvidenceMissing
	var failure *apierrors.ServiceError
	if errors.As(err, &failure) {
		code = failure.Code
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return relayError(code, "owned output or fixed goal verification failed")
}
