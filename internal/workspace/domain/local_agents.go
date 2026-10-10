package domain

import "time"

// Observation describes one independent readiness check, never inferred from
// another check. A CLI entry on PATH does not establish configuration or login.
type Observation struct {
	Status    string     `json:"status"`
	Reason    string     `json:"reason"`
	CheckedAt *time.Time `json:"checked_at"`
}

// CapabilityObservation uses the native capability vocabulary from SPEC 8.1.
// Help text is a hint; supported requires a successful native runtime check.
type CapabilityObservation struct {
	Status    string     `json:"status"`
	Reason    string     `json:"reason"`
	CheckedAt *time.Time `json:"checked_at"`
}

// LocalAgent is installation inventory, not a discovered or bound session.
// Executable paths and private native state never enter this transport value.
type LocalAgent struct {
	NativeAdapterRegistered bool                             `json:"native_adapter_registered"`
	ID                      string                           `json:"id"`
	DisplayName             string                           `json:"display_name"`
	Version                 *string                          `json:"version"`
	Installed               Observation                      `json:"installed"`
	Configured              Observation                      `json:"configured"`
	Startable               Observation                      `json:"startable"`
	Capabilities            map[string]CapabilityObservation `json:"capabilities"`
}

// UnknownNativeCapabilities returns a fresh map so callers cannot mutate
// another installation's observations.
func UnknownNativeCapabilities() map[string]CapabilityObservation {
	result := make(map[string]CapabilityObservation, 8)
	for _, name := range []string{"discover", "read_context", "start", "resume", "send", "stop", "observe", "reconcile"} {
		result[name] = CapabilityObservation{Status: "unknown", Reason: "native_runtime_not_tested"}
	}
	return result
}
