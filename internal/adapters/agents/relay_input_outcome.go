package agents

import (
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayInputOutcome struct {
	OperationID string
	Outcome     string
	FailureCode string
	ObservedAt  time.Time
}

func validateRelayInputOutcome(status string, hasError bool) error {
	if status != "completed" && status != "failed" && status != "interrupted" {
		return nativeError(apierrors.DeliveryUnknown, "native turn result has no supported terminal status")
	}
	if status == "completed" && hasError {
		return nativeError(apierrors.DeliveryUnknown, "completed native turn carries a conflicting failure")
	}
	return nil
}

// 在原接收器锁内关联明确的 turn/completed；完成时间不产生 started。
func (p *nativeProcess) recordRelayTurnCompletion(thread, turn, status string, hasError bool) (bool, bool, error) {
	if p.relayEvidence.StartupOperationID == "" {
		return true, true, nil
	}
	if p.cli != "codex" || p.relayEvidence.Provider != "codex" {
		return false, false, nativeError(apierrors.ScopeDenied, "native turn result provider differs from its original controller")
	}
	if err := validateRelayInputOutcome(status, hasError); err != nil {
		return false, false, err
	}
	if thread == "" || thread != p.nativeID || turn == "" {
		return false, false, nativeError(apierrors.ScopeDenied, "native completion differs from the original thread")
	}
	result := relayInputOutcome{Outcome: status, ObservedAt: time.Now().UTC()}
	if hasError {
		result.FailureCode = "native_failure"
	}
	for index, binding := range p.relayEvidence.Turns {
		if binding.TurnID != turn {
			continue
		}
		result.OperationID = binding.OperationID
		fresh, err := p.applyRelayTurnCompletion(index, result)
		return fresh, p.inputOperationID == binding.OperationID && p.turnID == turn, err
	}
	if p.turnID != "" || p.inputOperationID == "" {
		return false, false, nativeError(apierrors.ScopeDenied, "native completion has no original pending or accepted input")
	}
	result.OperationID = p.inputOperationID
	for observed, previous := range p.relayCompletedTurns {
		if previous.OperationID == result.OperationID && observed != turn {
			return false, false, nativeError(apierrors.ScopeDenied, "native completion has ambiguous pending turn identities")
		}
	}
	if old, found := p.relayCompletedTurns[turn]; found {
		if old.OperationID != result.OperationID || old.Outcome != result.Outcome || old.FailureCode != result.FailureCode {
			return false, false, nativeError(apierrors.ScopeDenied, "native completion changed before input acknowledgement")
		}
		return false, false, nil
	}
	if len(p.relayCompletedTurns) >= 256 {
		return false, false, nativeError(apierrors.BudgetExhausted, "unbound native completion observations exceed the limit")
	}
	if p.relayCompletedTurns == nil {
		p.relayCompletedTurns = map[string]relayInputOutcome{}
	}
	p.relayCompletedTurns[turn] = result
	return false, false, nil
}

func (p *nativeProcess) applyRelayTurnCompletion(index int, result relayInputOutcome) (bool, error) {
	binding := p.relayEvidence.Turns[index]
	if binding.OperationID != result.OperationID || binding.ControllerID != p.relayEvidence.ControllerID || binding.StartupOperationID != p.relayEvidence.StartupOperationID || binding.SessionID != p.relayEvidence.SessionID || binding.NativeID != p.nativeID {
		return false, nativeError(apierrors.ScopeDenied, "native result differs from the original accepted turn tuple")
	}
	if binding.ResultObservedAt != nil {
		if binding.Outcome != result.Outcome || binding.FailureCode != result.FailureCode || binding.ResultSource != "codex_turn_completed" {
			return false, nativeError(apierrors.ScopeDenied, "native result changed after its original observation")
		}
		return false, nil
	}
	binding.Outcome, binding.FailureCode, binding.ResultSource = result.Outcome, result.FailureCode, "codex_turn_completed"
	when := result.ObservedAt
	binding.ResultObservedAt = &when
	p.relayEvidence.Turns[index] = binding
	p.relayEvidence.Facts = append(p.relayEvidence.Facts, domain.RelayDriverFact{OperationID: result.OperationID, Kind: "input_result", Status: "confirmed", Source: "codex_turn_completed", Reason: result.Outcome, Sequence: int64(len(p.relayEvidence.Facts) + 1), TS: result.ObservedAt})
	return true, nil
}
