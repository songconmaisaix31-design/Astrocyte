package domain

import (
	"slices"
	"strings"
	"time"
)

type AdmissionAuthority string

const (
	AdmissionHuman             AdmissionAuthority = "human"
	AdmissionDelegatedJudgment AdmissionAuthority = "delegated_judgment"
	AdmissionActive                               = "active"
	AdmissionRevoked                              = "revoked"
)

type Admission struct {
	ID                  string             `json:"id"`
	Version             int                `json:"version"`
	OpportunityID       string             `json:"opportunity_id"`
	OpportunityRevision int                `json:"opportunity_revision"`
	OpportunityVersion  int                `json:"opportunity_version"`
	ProjectID           string             `json:"project_id,omitempty"`
	Inputs              []InputRef         `json:"inputs"`
	OwnerID             string             `json:"owner_id"`
	OwnerNamespace      string             `json:"owner_namespace"`
	ActorID             string             `json:"actor_id"`
	Authority           AdmissionAuthority `json:"authority"`
	DecisionID          string             `json:"decision_id,omitempty"`
	PolicyVersion       string             `json:"policy_version,omitempty"`
	Model               string             `json:"model,omitempty"`
	ModelVersion        string             `json:"model_version,omitempty"`
	Backend             string             `json:"backend,omitempty"`
	QuestionType        string             `json:"question_type,omitempty"`
	Purpose             string             `json:"purpose,omitempty"`
	Status              string             `json:"status"`
	AdmittedEventID     string             `json:"admitted_event_id"`
	RevokedEventID      string             `json:"revoked_event_id,omitempty"`
	Reason              string             `json:"reason,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

type AdmissionInput struct {
	Admission
	OpportunityState string
}

// 输入来源与权限由应用在同一事务核实，领域只判断固定版本和授权类型。
func NewAdmission(input AdmissionInput, now time.Time) (Admission, error) {
	value := input.Admission
	if value.ID == "" || value.AdmittedEventID == "" || value.OpportunityID == "" || value.OpportunityRevision < 1 || value.OpportunityVersion < 1 || value.OwnerID == "" || value.ActorID == "" || now.IsZero() {
		return Admission{}, ErrInvalid
	}
	if value.OwnerNamespace != "human:"+value.OwnerID {
		return Admission{}, ErrInvalid
	}
	if input.OpportunityState != "ready_for_review" {
		return Admission{}, ErrEvidence
	}
	if err := ValidateRefs(value.Inputs); err != nil {
		return Admission{}, err
	}
	switch value.Authority {
	case AdmissionHuman:
		if value.ActorID != value.OwnerID || value.DecisionID != "" || value.PolicyVersion != "" || value.Model != "" || value.ModelVersion != "" || value.Backend != "" || value.QuestionType != "" || value.Purpose != "" {
			return Admission{}, ErrInvalid
		}
	case AdmissionDelegatedJudgment:
		if value.DecisionID == "" || value.PolicyVersion == "" || value.Model == "" || value.ModelVersion == "" || value.Backend == "" || !slices.Contains([]string{"choice", "score", "truth"}, value.QuestionType) || value.Purpose != "opportunity_admission" {
			return Admission{}, ErrEvidence
		}
	default:
		return Admission{}, ErrInvalid
	}
	value.Inputs = slices.Clone(value.Inputs)
	value.Version, value.Status = 1, AdmissionActive
	value.CreatedAt, value.UpdatedAt = now.UTC(), now.UTC()
	value.RevokedEventID, value.Reason = "", ""
	return value, nil
}

// 撤销仅改变当前准入；已签发的执行批准由 Workspace 独立管理。
func (a *Admission) Revoke(expectedVersion int, actorID, eventID, reason string, now time.Time) error {
	if a.Version != expectedVersion {
		return ErrVersion
	}
	if a.Status != AdmissionActive || actorID != a.OwnerID || strings.TrimSpace(reason) == "" || eventID == "" || now.IsZero() {
		return ErrInvalid
	}
	a.Status, a.Version = AdmissionRevoked, a.Version+1
	a.RevokedEventID, a.Reason, a.UpdatedAt = eventID, reason, now.UTC()
	return nil
}
