package domain

import "time"

// ProgressBasis records the fixed files actually read to produce a stage. A
// basis is observed file provenance, never a claim the model supplied on its own.
type ProgressBasis struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// ProjectProgress is a human-authored or model-inferred project stage. It is an
// advisory observation: activity, commit time, and session headers never produce
// a stage, and a stage never grants or expands any permission.
type ProjectProgress struct {
	ProjectID  string          `json:"project_id"`
	Stage      string          `json:"stage"`
	Percent    *int            `json:"percent,omitempty"`
	Source     string          `json:"source"` // "human" | "agent_inferred"
	Basis      []ProgressBasis `json:"basis"`
	NativeID   string          `json:"native_id,omitempty"`
	Model      *string         `json:"model,omitempty"`
	ObservedAt time.Time       `json:"observed_at"`
	Revision   int             `json:"revision"`
}

// ProgressInput is the human-authored payload. Percent is optional and must be
// 0..100 inclusive when present.
type ProgressInput struct {
	Stage   string `json:"stage"`
	Percent *int   `json:"percent,omitempty"`
}

// ProgressCommand names the bounded fixed files (for example TASK.md or
// STATUS.md) already inside the approved project directory. The service
// re-validates every path and never accepts an arbitrary disk path or a
// credential/native-state component.
type ProgressCommand struct {
	Files []string `json:"files"`
}

// ValidProgressStage bounds the free-form stage label so a hostile or oversized
// model output cannot grow the stored record without limit.
func ValidProgressStage(stage string) bool {
	return len(stage) >= 1 && len(stage) <= 200
}

// ValidProgressPercent accepts only an explicit 0..100 integer, rejecting a
// fabricated or out-of-range completion value.
func ValidProgressPercent(p *int) bool {
	if p == nil {
		return true
	}
	return *p >= 0 && *p <= 100
}
