package domain

import (
	"encoding/json"
	"testing"
)

func TestRelayCompletionEdgesKeepStartAndReleaseIndependent(t *testing.T) {
	for _, from := range []string{RelayAccepted, RelayWorking} {
		if !ValidRelayTransition(from, RelayDone) {
			t.Fatal("an authorized human completion still requires an inferred start")
		}
		if ValidRelayTransition(from, RelayReleased) {
			t.Fatal("unfinished task bypassed the independent release audit")
		}
	}
	if ValidRelayTransition(RelayReady, RelayDone) || ValidRelayTransition(RelayReady, RelayReleased) || !ValidRelayTransition(RelayDone, RelayReleased) {
		t.Fatal("completion changed undispatched or released task boundaries")
	}
	encoded, err := json.Marshal(RelayGoalVerification{NativeID: "PRIVATE_NATIVE_CANARY", ControllerID: "PRIVATE_CONTROLLER_CANARY", TurnID: "PRIVATE_TURN_CANARY"})
	if err != nil || string(encoded) != "{}" {
		t.Fatal("private goal verification tuple entered JSON")
	}
}
