package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestAdmissionExplicitRuleInputs(t *testing.T) {
	// 全部字段是领域函数的明确参数，不声明存在来源、用户批准或模型判断。
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	input := AdmissionInput{OpportunityState: "ready_for_review", Admission: Admission{
		ID: "admission-input", OpportunityID: "opportunity-input", OpportunityRevision: 2, OpportunityVersion: 3,
		OwnerID: "owner-input", OwnerNamespace: "human:owner-input", ActorID: "owner-input", Authority: AdmissionHuman,
		AdmittedEventID: "event-input", Inputs: []InputRef{{Kind: "material", ID: "material-input", Revision: 1}},
	}}
	admission, err := NewAdmission(input, now)
	if err != nil || admission.Status != AdmissionActive || admission.Version != 1 || !reflect.DeepEqual(admission.Inputs, input.Inputs) {
		t.Fatal("explicit valid human input was not preserved", err)
	}
	for _, state := range []string{"incubating", "admitted", "deferred", "rejected", "withdrawn"} {
		candidate := input
		candidate.OpportunityState = state
		if _, err := NewAdmission(candidate, now); !errors.Is(err, ErrEvidence) {
			t.Fatalf("state input %s obtained admission", state)
		}
	}
	delegated := input
	delegated.Authority, delegated.ActorID = AdmissionDelegatedJudgment, "workflow-input"
	if _, err := NewAdmission(delegated, now); !errors.Is(err, ErrEvidence) {
		t.Fatal("incomplete judgment input obtained admission")
	}
	delegated.DecisionID, delegated.PolicyVersion = "decision-input", "policy-input"
	delegated.Model, delegated.ModelVersion, delegated.Backend, delegated.QuestionType = "model-input", "model-version-input", "backend-input", "opportunity_admission"
	if _, err := NewAdmission(delegated, now); err != nil {
		t.Fatal("complete explicit judgment metadata was rejected", err)
	}
	before := admission
	if err := admission.Revoke(0, input.OwnerID, "revocation-input", "human reason input", now); !errors.Is(err, ErrVersion) || !reflect.DeepEqual(before, admission) {
		t.Fatal("stale revocation changed admission")
	}
	if err := admission.Revoke(1, "another-owner-input", "revocation-input", "human reason input", now); !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(before, admission) {
		t.Fatal("another owner input changed admission")
	}
	if err := admission.Revoke(1, input.OwnerID, "revocation-input", "human reason input", now.Add(time.Second)); err != nil || admission.Status != AdmissionRevoked || admission.Version != 2 || admission.AdmittedEventID != input.AdmittedEventID || !reflect.DeepEqual(admission.Inputs, before.Inputs) {
		t.Fatal("explicit revocation lost immutable source or version", err)
	}
}
