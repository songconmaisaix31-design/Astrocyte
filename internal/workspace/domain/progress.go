package domain

import "time"

// ProgressEvidence records one fixed file actually read to produce a stage. It
// is observed file provenance, never a claim the model supplied on its own.
// Version embeds the observed file mtime/size (from the project files reader);
// the record-level ObservedAt is the freshness of the whole observation.
type ProgressEvidence struct {
	SourcePath string `json:"source_path"`
	Kind       string `json:"kind"` // "task" | "status" | "project_file"
	Version    string `json:"version"`
	Excerpt    string `json:"excerpt,omitempty"`
}

// ProgressOperation is a durable inference receipt, mirroring NativeOperation:
// an operation identity is persisted before the model turn so a retry cannot
// silently re-charge. accepted/failed/unknown match the native receipt statuses.
type ProgressOperation struct {
	Action     string     `json:"action"` // "infer"
	Status     string     `json:"status"` // pending | accepted | failed | unknown
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// ProjectProgress is a human-authored or model-inferred project stage. It is an
// advisory observation: activity, commit time, and session headers never produce
// a status, and a status never grants or expands any permission. Source
// distinguishes a human stage from an Agent inference.
type ProjectProgress struct {
	SchemaVersion int                `json:"schema_version"`
	ProjectID     string             `json:"project_id"`
	Status        string             `json:"status"`
	Summary       string             `json:"summary,omitempty"`
	Percent       *int               `json:"percent,omitempty"`
	Source        string             `json:"source"` // "human" | "agent_inferred"
	Evidence      []ProgressEvidence `json:"evidence"`
	NativeID      string             `json:"native_id,omitempty"`
	Model         *string            `json:"model,omitempty"`
	ObservedAt    time.Time          `json:"observed_at"`
	Warning       string             `json:"warning,omitempty"`
	Revision      int                `json:"revision"`

	PendingOperation string                       `json:"pending_operation,omitempty"`
	Operations       map[string]ProgressOperation `json:"operations,omitempty"`
}

// ProgressInput is the human-authored payload. Percent is optional and must be
// 0..100 inclusive when present.
type ProgressInput struct {
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
	Percent *int   `json:"percent,omitempty"`
}

// ProgressCommand names the bounded fixed files (for example TASK.md or
// STATUS.md) already inside the approved project directory, plus an explicit
// UUID operation identity for idempotent inference. The service re-validates
// every path and never accepts an arbitrary disk path or a credential/native
// state component.
type ProgressCommand struct {
	OperationID string   `json:"operation_id"`
	Files       []string `json:"files"`
}

// ValidProgressStatus bounds the free-form status label so a hostile or
// oversized model output cannot grow the stored record without limit.
func ValidProgressStatus(status string) bool {
	return len(status) >= 1 && len(status) <= 200
}

// ValidProgressPercent accepts only an explicit 0..100 integer, rejecting a
// fabricated or out-of-range completion value.
func ValidProgressPercent(p *int) bool {
	if p == nil {
		return true
	}
	return *p >= 0 && *p <= 100
}
