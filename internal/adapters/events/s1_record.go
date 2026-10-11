package events

import (
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func s1Record(v app.OutboxEvent) nativeObservation {
	p := object(v.Payload)
	o := nativeObservation{kind: "unknown", outcome: "unknown", depth: "unknown", cwd: str(p, "project_id")}
	if !v.OccurredAt.IsZero() {
		o.ts = v.OccurredAt.UnixMilli()
	}
	// Correlation IDs are request IDs, not native agent sessions.
	switch strings.TrimPrefix(v.Type, "attention.") {
	case "material_imported", "source_items_selected", "project_space_referenced":
		o.kind = "collect"
	case "material_updated":
		if str(p, "lifecycle") == "withdrawn" {
			o.kind = "user_veto"
		} else if string(p["pinned"]) == "true" {
			o.kind = "user_approve"
		}
	case "opportunity_reviewed":
		switch str(p, "feedback") {
		case "adopt":
			o.kind = "user_approve"
		case "reject", "revise":
			o.kind = "user_veto"
		}
	case "project_space_reference_removed":
		o.kind = "user_veto"
	}
	// Import/distillation jobs and native operation acceptance do not establish
	// agent-task completion. Ordinary reads do not establish depth or duration.
	return o
}
