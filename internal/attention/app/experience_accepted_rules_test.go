package app

import (
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// 仅验证明确的规则输入与期望值，不创建来源、模型调用、控制器或持久化收据。
func TestCompletionInterpretationRulesRetainUnknownModelVersion(t *testing.T) {
	complete := false
	draft := InterpretedExperienceDraft{
		Pattern: "rule input", Verdict: domain.ExperienceVerdictCondition, Condition: "explicit rule condition",
		Refused: &complete, Truncated: &complete,
		Provenance: ExperienceInterpretationProvenance{
			Backend: "codex", Model: "gpt-6.1-sol", ModelIdentitySource: "configured_route",
			ModelVersion: "unknown", ModelVersionSource: "unknown", PolicyVersion: "rule-input-v1", OperationID: uuid.NewString(),
		},
	}
	if err := validateExperienceInterpretation(draft); err != nil {
		t.Fatal("explicitly unknown independent version rejected a complete rule input", err)
	}
	for _, provenance := range []ExperienceInterpretationProvenance{
		{Backend: "codex", Model: "gpt-6.1-sol", ModelVersion: "unknown", ModelVersionSource: "unknown", PolicyVersion: "rule-input-v1", OperationID: draft.Provenance.OperationID},
		{Backend: "codex", Model: "gpt-6.1-sol", ModelIdentitySource: "configured_route", ModelVersion: "gpt-6.1-sol", ModelVersionSource: "provider_reported", PolicyVersion: "rule-input-v1", OperationID: draft.Provenance.OperationID},
		{Backend: "codex", Model: "gpt-6.1-sol", ModelIdentitySource: "configured_route", ModelVersion: "0.162.0", ModelVersionSource: "unknown", PolicyVersion: "rule-input-v1", OperationID: draft.Provenance.OperationID},
		{Backend: "codex", Model: "gpt-6.1-sol", ModelIdentitySource: "provider_reported", ModelVersion: "unknown", ModelVersionSource: "provider_reported", PolicyVersion: "rule-input-v1", OperationID: draft.Provenance.OperationID},
	} {
		invalid := draft
		invalid.Provenance = provenance
		if err := validateExperienceInterpretation(invalid); err == nil {
			t.Fatal("missing or contradictory observation source was accepted")
		}
	}
	refused := true
	for _, incomplete := range []InterpretedExperienceDraft{
		{Pattern: draft.Pattern, Verdict: draft.Verdict, Condition: draft.Condition, Provenance: draft.Provenance, Refused: nil, Truncated: &complete},
		{Pattern: draft.Pattern, Verdict: draft.Verdict, Condition: draft.Condition, Provenance: draft.Provenance, Refused: &refused, Truncated: &complete},
		{Pattern: draft.Pattern, Verdict: draft.Verdict, Condition: draft.Condition, Provenance: draft.Provenance, Refused: &complete, Truncated: &refused},
	} {
		if err := validateExperienceInterpretation(incomplete); err == nil {
			t.Fatal("unknown model version bypassed complete non-refused response requirements")
		}
	}
}
