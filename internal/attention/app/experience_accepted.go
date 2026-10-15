package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"reflect"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

const (
	ExperienceSourceAcceptedArtifact = "swarm_artifact_accepted"
	ExperienceSourceRelayCompletion  = "relay_task_completed"
)

type CompletionArtifactReference struct {
	ID         string      `json:"id"`
	Version    int         `json:"version"`
	Commit     string      `json:"commit,omitempty"`
	ObjectRefs []string    `json:"object_refs"`
	Inputs     []SourceRef `json:"inputs"`
}

// 只携带实际业务收据的安全固定引用；原生正文和原始输出保留在来源端。
type ExperienceCompletionEvidence struct {
	SourceKind     string                        `json:"source_kind"`
	EventID        string                        `json:"event_id"`
	Version        int                           `json:"version"`
	OccurredAt     int64                         `json:"occurred_at"`
	LocalProjectID string                        `json:"local_project_id"`
	OwnerProject   string                        `json:"owner_project"`
	ActorID        string                        `json:"actor_id"`
	ActorKind      string                        `json:"actor_kind"`
	TaskID         string                        `json:"task_id"`
	Outcome        string                        `json:"outcome"`
	OriginEventIDs []string                      `json:"origin_event_ids"`
	Artifacts      []CompletionArtifactReference `json:"artifacts"`
}

type ExperienceCompletionReader interface {
	ReadExperienceCompletion(context.Context, Principal, string, string) (ExperienceCompletionEvidence, error)
}

type ExperienceInterpretationProvenance struct {
	Backend       string `json:"backend"`
	Model         string `json:"model"`
	ModelVersion  string `json:"model_version"`
	PolicyVersion string `json:"policy_version"`
	OperationID   string `json:"operation_id"`
}

type InterpretedExperienceDraft struct {
	Pattern    string                             `json:"pattern"`
	Verdict    string                             `json:"verdict"`
	Condition  string                             `json:"condition"`
	Provenance ExperienceInterpretationProvenance `json:"provenance"`
	Refused    *bool                              `json:"refused"`
	Truncated  *bool                              `json:"truncated"`
}

type CompletionExperienceRequest struct {
	OperationID string                       `json:"operation_id"`
	Evidence    ExperienceCompletionEvidence `json:"evidence"`
	ResumeOnly  bool                         `json:"resume_only"`
}

type CompletionExperienceInterpreter interface {
	// ResumeOnly 请求只读取和核对原操作，禁止创建会话或重新发送。
	DraftCompletionExperience(context.Context, CompletionExperienceRequest) (InterpretedExperienceDraft, error)
}

type completionInterpretationReceipt struct {
	SchemaVersion int    `json:"schema_version"`
	OperationID   string `json:"operation_id"`
	Status        string `json:"status"`
}

type AcceptedExperienceSource struct {
	Completion  ExperienceCompletionEvidence       `json:"completion"`
	Interpreter ExperienceInterpretationProvenance `json:"interpreter"`
}

type AcceptedExperienceTx interface {
	ExperienceTx
	Context() context.Context
	LoadAcceptedExperienceSource(string) (AcceptedExperienceSource, error)
	SaveAcceptedExperienceSource(string, AcceptedExperienceSource) error
}

type AcceptedExperienceRepository interface {
	ExperienceRepository
	WithAcceptedExperienceTx(context.Context, func(AcceptedExperienceTx) error) error
	ReadAcceptedExperienceSource(context.Context, string) (AcceptedExperienceSource, error)
}

type AcceptedExperienceService struct {
	governance  *ExperienceService
	repository  AcceptedExperienceRepository
	sources     ExperienceCompletionReader
	interpreter CompletionExperienceInterpreter
}

func NewAcceptedExperienceService(governance *ExperienceService, repository AcceptedExperienceRepository, sources ExperienceCompletionReader, interpreter CompletionExperienceInterpreter) *AcceptedExperienceService {
	return &AcceptedExperienceService{governance: governance, repository: repository, sources: sources, interpreter: interpreter}
}

func validateCompletionSource(source ExperienceCompletionEvidence, projectID, eventID, owner string) error {
	if source.EventID != eventID || source.LocalProjectID != projectID || source.OwnerProject != owner || source.EventID == "" || source.LocalProjectID == "" || source.OwnerProject == "" || source.Version < 1 || source.OccurredAt < 1 || source.ActorID == "" || source.TaskID == "" {
		return serviceError(apierrors.EvidenceMissing, "Completion lacks an actual fixed receipt and registered project mapping", "inspect_completion_source")
	}
	if source.Outcome != "success" && source.Outcome != "failure" {
		return serviceError(apierrors.DeliveryUnknown, "Completion outcome remains unknown", "reconcile_original_completion")
	}
	switch source.SourceKind {
	case ExperienceSourceAcceptedArtifact:
		if source.ActorKind != "human" || source.Outcome != "success" || len(source.Artifacts) == 0 {
			return serviceError(apierrors.EvidenceMissing, "Accepted artifact requires actual human acceptance and fixed artifacts", "inspect_actual_artifact_acceptance")
		}
	case ExperienceSourceRelayCompletion:
		if source.ActorKind != "human" && source.ActorKind != "agent" {
			return serviceError(apierrors.EvidenceMissing, "Relay completion has no actual actor", "inspect_original_completion")
		}
		if len(source.OriginEventIDs) == 0 {
			return serviceError(apierrors.EvidenceMissing, "Relay completion lacks original trace event references", "inspect_original_completion")
		}
	default:
		return serviceError(apierrors.ScopeDenied, "Source is not a supported actual completion receipt", "use_actual_completion_source")
	}
	for _, id := range source.OriginEventIDs {
		if strings.TrimSpace(id) == "" {
			return domain.ErrEvidence
		}
	}
	for _, artifact := range source.Artifacts {
		if artifact.ID == "" || artifact.Version < 1 {
			return domain.ErrEvidence
		}
		for _, object := range artifact.ObjectRefs {
			if strings.TrimSpace(object) == "" {
				return domain.ErrEvidence
			}
		}
	}
	return nil
}

func validateExperienceInterpretation(draft InterpretedExperienceDraft) error {
	if draft.Refused == nil || draft.Truncated == nil || *draft.Refused || *draft.Truncated {
		return serviceError(apierrors.EvidenceMissing, "Interpreter result lacks a complete non-refused response", "inspect_actual_interpreter_result")
	}
	if strings.TrimSpace(draft.Pattern) == "" || strings.TrimSpace(draft.Condition) == "" || !slices.Contains([]string{domain.ExperienceVerdictAvoid, domain.ExperienceVerdictPrefer, domain.ExperienceVerdictCondition}, draft.Verdict) {
		return domain.ErrInvalid
	}
	p := draft.Provenance
	if p.Backend == "" || p.Model == "" || p.ModelVersion == "" || p.PolicyVersion == "" || p.OperationID == "" {
		return serviceError(apierrors.EvidenceMissing, "Interpreter backend, model, policy and original operation must be recorded", "inspect_actual_interpreter_provenance")
	}
	return nil
}

// 可信收据消费者传实际 ID；候选始终停留 quarantine，人工批准复用既有服务。
func (s *AcceptedExperienceService) CaptureCompletion(ctx context.Context, p Principal, localProjectID, eventID string) (StoredExperience, error) {
	var result StoredExperience
	if err := experiencePrincipal(p, false); err != nil {
		return result, err
	}
	if s.governance == nil || s.repository == nil || s.sources == nil {
		return result, serviceError(apierrors.UnsupportedCapability, "Actual completion source and experience persistence are not connected", "connect_completion_experience_services")
	}
	owner := s.governance.options.ProjectID
	if err := s.governance.project(owner); err != nil {
		return result, err
	}
	source, err := s.sources.ReadExperienceCompletion(ctx, p, localProjectID, eventID)
	if err != nil {
		return result, err
	}
	if err := validateCompletionSource(source, localProjectID, eventID, owner); err != nil {
		return result, err
	}
	digest, err := semanticDigest(source)
	if err != nil {
		return result, err
	}
	request := CompletionExperienceRequest{Evidence: source}
	var found bool
	err = s.repository.WithAcceptedExperienceTx(ctx, func(tx AcceptedExperienceTx) error {
		var err error
		result, found, err = tx.FindExperience(owner, eventID)
		if err != nil {
			return err
		}
		if found {
			stored, err := tx.LoadAcceptedExperienceSource(result.Experience.ID)
			if err != nil {
				return err
			}
			if result.SourceKind != source.SourceKind || !reflect.DeepEqual(stored.Completion, source) {
				return serviceError(apierrors.ContextStale, "Original receipt differs from its fixed experience source", "inspect_existing_experience_candidate")
			}
			return nil
		}
		if s.interpreter == nil {
			return serviceError(apierrors.ProviderUnavailable, "Authorized completion Interpreter is not connected", "configure_current_interpreter_backend")
		}
		// 沿用既有幂等收据固定原操作；并发和重启只允许核对该操作。
		const caller = "system:completion-experience"
		command := "Experience/InterpretCompletion/" + owner
		receipt, err := tx.LoadExperienceReceipt(caller, command, eventID)
		if err == nil {
			if receipt.Digest != digest {
				return serviceError(apierrors.ContextStale, "Completion interpretation already binds different fixed evidence", "reconcile_original_completion")
			}
			var original completionInterpretationReceipt
			if err := json.Unmarshal(receipt.Result, &original); err != nil {
				return err
			}
			if _, err := uuid.Parse(original.OperationID); err != nil || original.SchemaVersion != 1 || original.Status != "prepared" {
				return serviceError(apierrors.DeliveryUnknown, "Original completion interpretation identity is unavailable", "reconcile_original_completion")
			}
			request.OperationID = original.OperationID
			request.ResumeOnly = true
			return nil
		}
		if !isMissing(err) {
			return err
		}
		request.OperationID = uuid.NewString()
		data, err := json.Marshal(completionInterpretationReceipt{SchemaVersion: 1, OperationID: request.OperationID, Status: "prepared"})
		if err != nil {
			return err
		}
		return tx.SaveExperienceReceipt(caller, command, eventID, Receipt{Digest: digest, Result: data})
	})
	if err != nil {
		return result, err
	}
	if found {
		if result.SourceKind != source.SourceKind {
			return result, serviceError(apierrors.VersionConflict, "Original receipt already binds another experience source kind", "inspect_existing_experience_candidate")
		}
		return result, nil
	}
	draft, err := s.interpreter.DraftCompletionExperience(ctx, request)
	if err != nil {
		return result, err
	}
	if err := validateExperienceInterpretation(draft); err != nil {
		return result, err
	}
	if draft.Provenance.OperationID != request.OperationID {
		return result, serviceError(apierrors.DeliveryUnknown, "Interpreter result belongs to another original operation", "reconcile_original_completion")
	}
	err = s.repository.WithAcceptedExperienceTx(ctx, func(tx AcceptedExperienceTx) error {
		current, err := s.sources.ReadExperienceCompletion(tx.Context(), p, localProjectID, eventID)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(source, current) {
			return serviceError(apierrors.ContextStale, "Actual completion authority or fixed evidence changed", "review_current_completion_source")
		}
		previous, found, err := tx.FindExperience(owner, eventID)
		if err != nil {
			return err
		}
		if found {
			if previous.SourceKind != source.SourceKind {
				return serviceError(apierrors.VersionConflict, "Original receipt already binds another experience source kind", "inspect_existing_experience_candidate")
			}
			stored, err := tx.LoadAcceptedExperienceSource(previous.Experience.ID)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(stored.Completion, source) {
				return serviceError(apierrors.ContextStale, "Original receipt differs from its fixed experience source", "inspect_existing_experience_candidate")
			}
			result = previous
			return nil
		}
		now := s.governance.options.Clock().UnixMilli()
		result = StoredExperience{SourceKind: source.SourceKind, Experience: domain.Experience{ID: rand.Text(), Pattern: draft.Pattern, Verdict: draft.Verdict, Condition: draft.Condition, Evidence: []domain.EventRef{{EventID: source.EventID, TS: source.OccurredAt}}, Status: domain.StatusQuarantine}, OwnerProject: owner, OriginEventID: source.EventID, Version: 1, CreatedAt: now, UpdatedAt: now, Shares: []ShareAudit{}, Actions: []ActionAudit{}}
		if err := tx.InsertExperience(result); err != nil {
			return err
		}
		return tx.SaveAcceptedExperienceSource(result.Experience.ID, AcceptedExperienceSource{Completion: source, Interpreter: draft.Provenance})
	})
	return result, err
}

type SafeExperienceEvidence struct {
	SourceKind string                        `json:"source_kind"`
	EventID    string                        `json:"event_id"`
	OccurredAt int64                         `json:"occurred_at"`
	Completion *ExperienceCompletionEvidence `json:"completion,omitempty"`
}

// 统一安全引用不把业务收据伪装成原生 Event，也不返回原始输入或响应。
func (s *AcceptedExperienceService) ResolveEvidence(ctx context.Context, p Principal, id, project string) ([]SafeExperienceEvidence, error) {
	if s.governance == nil || s.repository == nil {
		return nil, serviceError(apierrors.UnsupportedCapability, "Experience evidence services are not connected", "connect_experience_evidence_services")
	}
	record, err := s.governance.GetExperience(ctx, p, id, project)
	if err != nil {
		return nil, err
	}
	if record.SourceKind != "" {
		if s.sources == nil {
			return nil, serviceError(apierrors.UnsupportedCapability, "Current completion permission reader is not connected", "connect_completion_source_reader")
		}
		stored, err := s.repository.ReadAcceptedExperienceSource(ctx, id)
		if err != nil {
			return nil, err
		}
		current, err := s.sources.ReadExperienceCompletion(ctx, p, stored.Completion.LocalProjectID, record.OriginEventID)
		if err != nil {
			return nil, err
		}
		if record.SourceKind != stored.Completion.SourceKind || !reflect.DeepEqual(stored.Completion, current) {
			return nil, serviceError(apierrors.ContextStale, "Completion evidence differs from its immutable candidate source", "inspect_original_completion_source")
		}
		return []SafeExperienceEvidence{{SourceKind: record.SourceKind, EventID: current.EventID, OccurredAt: current.OccurredAt, Completion: &current}}, nil
	}
	if s.governance.events == nil {
		return nil, serviceError(apierrors.UnsupportedCapability, "Original event reader is not connected", "connect_original_event_reader")
	}
	result := make([]SafeExperienceEvidence, 0, len(record.Experience.Evidence))
	for _, ref := range record.Experience.Evidence {
		event, err := s.governance.events.GetEvent(ctx, ref.EventID)
		if err != nil {
			return nil, err
		}
		if event.ID != ref.EventID || event.Event == nil || event.Event.TS != ref.TS || event.OriginKey == "" || event.RawBytes == nil {
			return nil, serviceError(apierrors.EvidenceMissing, "Original event evidence does not resolve", "inspect_original_event")
		}
		result = append(result, SafeExperienceEvidence{SourceKind: "original_event", EventID: event.ID, OccurredAt: ref.TS})
	}
	return result, nil
}
