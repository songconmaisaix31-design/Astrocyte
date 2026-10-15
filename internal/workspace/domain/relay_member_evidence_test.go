package domain

import (
	"encoding/json"
	"testing"
	"time"
)

// 仅检查序列化隐私，不创建实际会话、事件、所有权或执行结果。
func TestRelayMemberEvidencePrivateSerialization(t *testing.T) {
	observed := time.Now().UTC()
	turn := "PRIVATE_TURN_CANARY"
	fact := RelayMemberEvidence{
		Purpose: "submit_artifact", SessionID: "PRIVATE_SESSION_CANARY", NativeID: "PRIVATE_NATIVE_CANARY",
		ControllerID: "PRIVATE_CONTROLLER_CANARY", StartupOperationID: "PRIVATE_STARTUP_CANARY", InputOperationID: "PRIVATE_INPUT_CANARY",
		ProjectID: "PRIVATE_PROJECT_CANARY", CWD: "PRIVATE_CWD_CANARY", CLI: "PRIVATE_CLI_CANARY", CLIVersion: "PRIVATE_VERSION_CANARY",
		AcceptedTurnID: &turn, ControllerStartedAt: &observed, InputAcceptedAt: &observed, TurnStartedAt: &observed, ExitObservedAt: &observed,
		Outcome: &turn, FailureCode: &turn, ResultSource: "PRIVATE_SOURCE_CANARY", ResultObservedAt: &observed,
		Capabilities: []string{"PRIVATE_CAPABILITY_CANARY"}, Tools: []string{"PRIVATE_TOOL_CANARY"}, ExternalTargets: []string{"PRIVATE_TARGET_CANARY"},
	}
	encoded, err := json.Marshal(fact)
	if err != nil || string(encoded) != "{}" {
		t.Fatal("portable membership evidence disclosed private data")
	}
}
