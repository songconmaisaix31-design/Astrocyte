package agents

import (
	"context"
	"encoding/json"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"time"
)

// 普通接入只认调用方已记录的本次 pending 动作，不从 sessionID 命名推断。
func nativeRequestOperation(r domain.NativeRequest, action string) (string, error) {
	operation := r.Session.PendingOperation
	if operation == "" {
		return "", nil
	}
	receipt, exists := r.Session.Operations[operation]
	if !exists || receipt.Action != action || receipt.Status != "pending" || action == "start" && r.Session.LastOperationID != operation {
		return "", nativeError(apierrors.ScopeDenied, "native request operation does not match its original pending receipt")
	}
	if r.Command.OperationID != "" && r.Command.OperationID != operation {
		return "", nativeError(apierrors.ScopeDenied, "native request operation identities differ")
	}
	return operation, nil
}

func cloneOriginalNativeSession(session domain.NativeSession) (domain.NativeSession, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return domain.NativeSession{}, nativeError(apierrors.InternalError, "original native session could not be retained")
	}
	var result domain.NativeSession
	if err = json.Unmarshal(data, &result); err != nil {
		return domain.NativeSession{}, nativeError(apierrors.InternalError, "original native session could not be retained")
	}
	result.HistoryPath = session.HistoryPath
	return result, nil
}

func (p *nativeProcess) captureOriginalSession(session domain.NativeSession) error {
	copy, err := cloneOriginalNativeSession(session)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.originalSession = &copy
	p.mu.Unlock()
	return nil
}

func (n *Native) ReadOwnedNativeSession(ctx context.Context, id string) (domain.RelayOwnedNativeSession, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayOwnedNativeSession{}, err
	}
	n.mu.Lock()
	p := n.processes[id]
	n.mu.Unlock()
	if p == nil {
		return domain.RelayOwnedNativeSession{}, &domain.RelayObservationUnavailable{Reason: "controller_not_held", Cause: nativeError(apierrors.DeliveryUnknown, "original native session is not held by this controller")}
	}
	p.mu.Lock()
	if p.originalSession == nil {
		p.mu.Unlock()
		return domain.RelayOwnedNativeSession{}, nativeError(apierrors.DeliveryUnknown, "original Start session was not retained")
	}
	session, err := cloneOriginalNativeSession(*p.originalSession)
	p.mu.Unlock()
	if err != nil {
		return domain.RelayOwnedNativeSession{}, err
	}
	evidence, err := n.OwnedRelayControllerEvidence(ctx, session)
	if err != nil {
		return domain.RelayOwnedNativeSession{}, err
	}
	if evidence.ControllerID == "" {
		return domain.RelayOwnedNativeSession{}, nativeError(apierrors.DeliveryUnknown, "original startup operation was not recorded")
	}
	return domain.RelayOwnedNativeSession{Session: session, Evidence: evidence}, nil
}

func (r *Registry) ReadOwnedNativeSession(ctx context.Context, id string) (domain.RelayOwnedNativeSession, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayOwnedNativeSession{}, err
	}
	var result domain.RelayOwnedNativeSession
	found := false
	for _, adapter := range r.adapters {
		native, supported := adapter.(*Native)
		if !supported {
			continue
		}
		native.mu.Lock()
		_, held := native.processes[id]
		native.mu.Unlock()
		if !held {
			continue
		}
		if found {
			return domain.RelayOwnedNativeSession{}, nativeError(apierrors.ScopeDenied, "native session identity is ambiguous across current controllers")
		}
		var err error
		result, err = native.ReadOwnedNativeSession(ctx, id)
		if err != nil {
			return domain.RelayOwnedNativeSession{}, err
		}
		found = true
	}
	if !found {
		return domain.RelayOwnedNativeSession{}, &domain.RelayObservationUnavailable{Reason: "controller_not_held", Cause: nativeError(apierrors.DeliveryUnknown, "original native session is not held in this Registry")}
	}
	return result, nil
}

// 仅在原持有进程锁内追加；时间来自事实产生位置。
func (p *nativeProcess) relayFact(kind, source, operation string) {
	if p.relayEvidence.ControllerID == "" {
		return
	}
	facts := p.relayEvidence.Facts
	p.relayEvidence.Facts = append(facts, domain.RelayDriverFact{OperationID: operation, Kind: kind, Status: "confirmed", Source: source, Sequence: int64(len(facts) + 1), TS: time.Now().UTC()})
}

func (p *nativeProcess) consumptionUnknown(operation string) {
	if p.relayEvidence.ControllerID == "" {
		return
	}
	p.relayEvidence.Facts = append(p.relayEvidence.Facts, domain.RelayDriverFact{OperationID: operation, Kind: "input_consumption", Status: "unknown", Source: "native_protocol", Reason: "client_consumption_not_acknowledged", Sequence: int64(len(p.relayEvidence.Facts) + 1), TS: time.Now().UTC()})
}

func (n *Native) OwnedRelayControllerEvidence(ctx context.Context, session domain.NativeSession) (domain.RelayControllerEvidence, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p, err := n.get(session)
	if err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// 旧进程缺少原启动 operation，只返回未记录，不推导新的归属证据。
	if p.relayEvidence.StartupOperationID == "" {
		return domain.RelayControllerEvidence{}, nil
	}
	if session.ID == "" || session.NativeID == "" || session.ProjectID == "" || session.Version == "" || p.cmd == nil || p.cmd.Process == nil || p.cmd.Process.Pid <= 0 {
		return domain.RelayControllerEvidence{}, nativeError(apierrors.DeliveryUnknown, "exact native session or owned process evidence is missing")
	}
	if p.projectID != session.ProjectID || session.Ownership != "owned" || session.CLI != n.id || p.relayEvidence.SessionID != session.ID || p.relayEvidence.ControllerID == "" {
		return domain.RelayControllerEvidence{}, nativeError(apierrors.ScopeDenied, "original controller identity differs")
	}
	result := p.relayEvidence
	result.NativeID = p.nativeID
	select {
	case <-p.done:
		result.ExitConfirmed = true
	default:
		result.ExitConfirmed = false
	}
	result.Facts = append([]domain.RelayDriverFact{}, result.Facts...)
	result.Turns = copyRelayTurns(result.Turns)
	return result, nil
}

func copyRelayTurns(turns []domain.RelayTurnBinding) []domain.RelayTurnBinding {
	result := append([]domain.RelayTurnBinding{}, turns...)
	for i, turn := range result {
		if turn.StartedAt != nil {
			when := *turn.StartedAt
			result[i].StartedAt = &when
		}
		if turn.ResultObservedAt != nil {
			when := *turn.ResultObservedAt
			result[i].ResultObservedAt = &when
		}
	}
	return result
}

func (p *nativeProcess) bindRelayTurn(session domain.NativeSession, operation, turn string) error {
	if operation == "" || p.relayEvidence.StartupOperationID == "" {
		return nil
	}
	if p.cli != "codex" || p.relayEvidence.Provider != "codex" || turn == "" || session.ID == "" || session.ID != p.relayEvidence.SessionID || session.NativeID == "" || session.NativeID != p.nativeID || session.CLI != p.cli || session.ProjectID != p.projectID || session.Ownership != "owned" || p.relayEvidence.ControllerID == "" || p.inputOperationID != operation {
		return nativeError(apierrors.ScopeDenied, "native turn acknowledgement differs from its original controller input")
	}
	for _, previous := range p.relayEvidence.Turns {
		if previous.OperationID == operation || previous.TurnID == turn {
			return nativeError(apierrors.ScopeDenied, "native turn acknowledgement duplicates an original binding")
		}
	}
	for observed, pending := range p.relayCompletedTurns {
		if pending.OperationID == operation && observed != turn {
			return nativeError(apierrors.ScopeDenied, "native completion turn differs from input acknowledgement")
		}
	}
	pending, found := p.relayCompletedTurns[turn]
	if found && pending.OperationID != operation {
		return nativeError(apierrors.ScopeDenied, "native completion differs from the original pending input")
	}
	p.relayEvidence.Turns = append(p.relayEvidence.Turns, domain.RelayTurnBinding{OperationID: operation, SessionID: session.ID, NativeID: session.NativeID, ControllerID: p.relayEvidence.ControllerID, StartupOperationID: p.relayEvidence.StartupOperationID, TurnID: turn, AcceptedAt: time.Now().UTC()})
	if found {
		fresh, err := p.applyRelayTurnCompletion(len(p.relayEvidence.Turns)-1, pending)
		if err != nil {
			return err
		}
		if fresh {
			p.status = pending.Outcome
			p.modelObservedAt = pending.ObservedAt
			p.addEvent("turn_"+pending.Outcome, "")
		}
		delete(p.relayCompletedTurns, turn)
	}
	return nil
}

func (p *nativeProcess) markRelayTurnStarted(turn string, when time.Time) {
	for i, binding := range p.relayEvidence.Turns {
		if binding.TurnID == turn && binding.OperationID == p.inputOperationID && binding.StartedAt == nil {
			observed := when
			p.relayEvidence.Turns[i].StartedAt = &observed
			return
		}
	}
}

func (n *Native) StartOperation(ctx context.Context, request domain.NativeRequest, operation string) (domain.NativeSession, error) {
	if operation == "" {
		return request.Session, &domain.NativePrelaunchFailure{Cause: nativeError(apierrors.ValidationFailed, "relay startup operation is required")}
	}
	return n.start(ctx, request, false, operation)
}

func (n *Native) ResumeOperation(ctx context.Context, request domain.NativeRequest, operation string) (domain.NativeSession, error) {
	if err := n.validateRelayResume(ctx, request, operation); err != nil {
		return request.Session, err
	}
	return n.start(ctx, request, true, operation)
}

func (n *Native) validateRelayResume(ctx context.Context, request domain.NativeRequest, operation string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if operation == "" {
		return nativeError(apierrors.ValidationFailed, "relay resume operation is required")
	}
	p, err := n.get(request.Session)
	if err != nil {
		return err
	}
	p.mu.Lock()
	identity := p.projectID == request.Project.ID && sameRelayMetadataCWD(p.cmd.Dir, request.Project.Root) && request.Session.Ownership == "owned" && request.SessionID == request.Session.ID && p.relayEvidence.StartupOperationID != ""
	p.mu.Unlock()
	if !identity {
		return nativeError(apierrors.ScopeDenied, "resume differs from the original controller workspace or operation")
	}
	if !request.Session.StopConfirmed || !p.observe().StopConfirmed || request.Session.NativeID == "" {
		return nativeError(apierrors.DeliveryUnknown, "resume requires original controller exit and native identity")
	}
	return nil
}

func (n *Native) SendOperation(ctx context.Context, session domain.NativeSession, input, operation string) (domain.NativeObservation, error) {
	if operation == "" {
		return domain.NativeObservation{}, nativeError(apierrors.ValidationFailed, "relay input operation is required")
	}
	if evidence, err := n.OwnedRelayControllerEvidence(ctx, session); err != nil {
		return domain.NativeObservation{}, err
	} else if evidence.ControllerID == "" {
		return domain.NativeObservation{}, nativeError(apierrors.DeliveryUnknown, "original relay controller operation is unrecorded")
	}
	p, err := n.get(session)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if p.relayInputGuard != nil {
		original, err := p.relayInputSession(session)
		if err != nil {
			return domain.NativeObservation{}, err
		}
		input = packetPrompt(original.ContextPacket, input)
	}
	return n.send(ctx, session, input, operation)
}
