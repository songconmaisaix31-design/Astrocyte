package agents

import (
	"errors"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只核原协议终态的拒绝边界，不构造会话、模型结果或成功通知。
func TestRelayInputOutcomeRejectsMissingAndConflictingTerminalStatus(t *testing.T) {
	for _, status := range []string{"", "inProgress", "done", "released", "exited", "success", "unknown"} {
		if err := validateRelayInputOutcome(status, false); err == nil {
			t.Fatal("nonterminal or inferred task outcome was accepted as a native result")
		}
	}
	if err := validateRelayInputOutcome("completed", true); err == nil {
		t.Fatal("conflicting native completion and failure produced a result")
	}
}

func TestRelayInputOutcomeRejectsUnownedIdentityWithoutFacts(t *testing.T) {
	process := &nativeProcess{cli: "codex", relayEvidence: domain.RelayControllerEvidence{Provider: "codex", StartupOperationID: "unowned-startup"}}
	fresh, current, err := process.recordRelayTurnCompletion("unowned-thread", "unowned-turn", "failed", true)
	var denied *apierrors.ServiceError
	if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied || fresh || current {
		t.Fatal("unowned result acquired an input association")
	}
	if err = process.bindRelayTurn(domain.NativeSession{}, "unowned-input", "unowned-turn"); !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
		t.Fatal("missing original controller tuple acquired a turn binding")
	}
	if len(process.relayEvidence.Turns) != 0 || len(process.relayEvidence.Facts) != 0 || len(process.relayCompletedTurns) != 0 || len(process.events) != 0 || process.turnID != "" || process.status != "" {
		t.Fatal("rejected identity changed accepted input, result, start or process facts")
	}
}
