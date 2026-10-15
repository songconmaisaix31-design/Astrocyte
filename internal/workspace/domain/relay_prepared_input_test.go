package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRelayPreparedInputPrivateSerialization(t *testing.T) {
	value := RelayPreparedInput{
		Session:     NativeSession{ID: "PRIVATE_SESSION_CANARY", NativeID: "PRIVATE_NATIVE_CANARY"},
		OperationID: "PRIVATE_OPERATION_CANARY", ActorID: "PRIVATE_ACTOR_CANARY", ActorKind: "PRIVATE_KIND_CANARY",
		ControllerID: "PRIVATE_CONTROLLER_CANARY", StartupOperationID: "PRIVATE_STARTUP_CANARY",
		WorkspaceRoot: "PRIVATE_ROOT_CANARY", PreparedAt: time.Now(),
	}
	encoded, err := json.Marshal(value)
	if err != nil || string(encoded) != "{}" {
		t.Fatal("private prepared input record was exposed through JSON")
	}
}
