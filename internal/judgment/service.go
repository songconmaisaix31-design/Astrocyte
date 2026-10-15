package judgment

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type Service struct {
	repo   Repository
	access Authorizer
	models map[Backend]Model
}

func NewService(repo Repository, access Authorizer, models ...Model) (*Service, error) {
	if repo == nil || access == nil {
		return nil, invalid("judgment repository and authorization are required")
	}
	s := &Service{repo: repo, access: access, models: map[Backend]Model{}}
	for _, model := range models {
		if model == nil || (model.Backend() != JevBackend && model.Backend() != LayaBackend) || s.models[model.Backend()] != nil {
			return nil, invalid("each model backend must have exactly one adapter")
		}
		s.models[model.Backend()] = model
	}
	return s, nil
}

func (s *Service) resolve(ctx context.Context, actor Actor, request JudgmentRequest) (AuthorizedInput, error) {
	if actor.ID == "" || (actor.Kind != "human" && actor.Kind != "agent") || (actor.Kind == "agent" && actor.ProjectID != request.ProjectID) {
		return AuthorizedInput{}, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "judgment requires an authenticated actor in this project"}
	}
	input, err := s.access.Resolve(ctx, actor, request)
	if err != nil {
		return input, err
	}
	if err := ValidatePolicy(input.Policy); err != nil {
		return input, err
	}
	p := input.Policy
	if p.ProjectID != request.ProjectID || p.QuestionType != request.Question.Type || p.Versions != request.Versions || input.WorktreeID != request.WorktreeID || input.SnapshotID != request.SnapshotID {
		return input, conflict("judgment project snapshot or goal/taste/policy/model/access version changed")
	}
	allowed := false
	for _, language := range p.Languages {
		if language == request.Language {
			allowed = true
		}
	}
	if !allowed || request.Backend != p.Backend || request.ExternalScope != p.ExternalScope {
		return input, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "question language, backend or external data scope is not authorized"}
	}
	if request.Budget.TimeoutMS > p.Budget.TimeoutMS || request.Budget.MaxInputBytes > p.Budget.MaxInputBytes || request.Budget.MaxInputTokens > p.Budget.MaxInputTokens || request.Budget.MaxOutputTokens > p.Budget.MaxOutputTokens {
		return input, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "request exceeds authorized judgment budget"}
	}
	if len(input.Evidence) != len(request.EvidenceRefs) {
		return input, invalid("authorization must resolve every fixed evidence reference")
	}
	for i, evidence := range input.Evidence {
		if evidence.Reference != request.EvidenceRefs[i] || DigestText(evidence.Text) != evidence.Reference.Digest {
			return input, conflict("fixed evidence revision or digest differs")
		}
	}
	return input, nil
}

func (s *Service) Evaluate(ctx context.Context, actor Actor, request JudgmentRequest) (JudgmentResult, error) {
	var result JudgmentResult
	if err := ValidateRequest(request); err != nil {
		return result, err
	}
	input, err := s.resolve(ctx, actor, request)
	if err != nil {
		return result, err
	}
	key, err := digestJSON(struct {
		Actor         Actor
		Request       JudgmentRequest
		Policy        Policy
		AccessVersion string
	}{actor, request, input.Policy, input.AccessVersion})
	if err != nil {
		return result, err
	}
	record, found, err := s.repo.FindCached(ctx, key)
	if err != nil {
		return result, err
	}
	if found {
		if err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
			_, err := s.recheck(ctx, actor, request, input)
			return err
		}); err != nil {
			return result, err
		}
		if record.Result.ReasonCode == "backend_call_failed" || record.Result.ReasonCode == "backend_unavailable" {
			return result, errors.New("repository returned a failed model call as cache")
		}
		record.Result.CacheHit = true
		return record.Result, nil
	}
	stateParts := make([]string, len(input.Evidence))
	for i, evidence := range input.Evidence {
		stateParts[i] = evidence.Text
	}
	state := strings.Join(stateParts, "\n\n")
	inputDigest, err := digestJSON(struct {
		Question Question
		State    string
		Language string
	}{request.Question, state, request.Language})
	if err != nil {
		return result, err
	}
	result = JudgmentResult{
		SchemaVersion: "1.0", ID: uuid.NewString(), Version: 1,
		QuestionID: request.Question.ID, QuestionType: request.Question.Type,
		Backend: RuleBackend, InputDigest: inputDigest, Decision: NeedEvidence,
		ReasonCode: "missing_evidence", Versions: request.Versions,
		EvidenceRefs: append([]EvidenceRef{}, request.EvidenceRefs...), CreatedAt: time.Now().UTC(),
		Distribution: map[string]float64{}, Options: map[string]OptionCoverage{},
	}
	backendInput := BackendInput{Question: request.Question, State: state, Language: request.Language, ModelRevision: request.Versions.Model, Budget: request.Budget}
	if budgetErr := ValidateBackendInputBytes(backendInput); budgetErr != nil {
		result.Decision, result.ReasonCode = Abstain, "input_byte_budget_exceeded"
		err = budgetErr
	} else if strings.TrimSpace(state) != "" {
		if !applyRule(request.Question, input.Evidence, &result) {
			if request.Backend == RuleBackend {
				result.Decision, result.ReasonCode = Abstain, "rule_does_not_determine_answer"
			} else {
				err = s.evaluateModel(ctx, request, input.Policy, state, &result)
			}
		}
	}
	// 结果保存与使用前重新读取实际权限和输入，过期结果不可返回。
	if saveErr := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		if _, recheckErr := s.recheck(ctx, actor, request, input); recheckErr != nil {
			return recheckErr
		}
		return s.repo.SaveResult(ctx, Record{Actor: actor, Request: request, Result: result, CacheKey: key, Cacheable: err == nil})
	}); saveErr != nil {
		return result, saveErr
	}
	return result, err
}

func (s *Service) recheck(ctx context.Context, actor Actor, request JudgmentRequest, before AuthorizedInput) (AuthorizedInput, error) {
	after, err := s.resolve(ctx, actor, request)
	if err != nil {
		return after, err
	}
	if !reflect.DeepEqual(before, after) {
		return after, conflict("judgment authorization or fixed input changed during use")
	}
	return after, nil
}

func (s *Service) evaluateModel(ctx context.Context, request JudgmentRequest, policy Policy, state string, result *JudgmentResult) error {
	result.Backend, result.Decision = request.Backend, Abstain
	model := s.models[request.Backend]
	if model == nil {
		result.ReasonCode = "backend_unavailable"
		return errors.New("selected judgment backend is unavailable")
	}
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(request.Budget.TimeoutMS)*time.Millisecond)
	defer cancel()
	out, err := model.Evaluate(callCtx, BackendInput{Question: request.Question, State: state, Language: request.Language, ModelRevision: request.Versions.Model, Budget: request.Budget})
	if err != nil {
		result.ReasonCode = "backend_call_failed"
		return err
	}
	if err := ValidateModelResult(request.Question, out); err != nil {
		result.ReasonCode = "backend_call_failed"
		return err
	}
	if out.ModelRevision != request.Versions.Model {
		result.ReasonCode = "backend_call_failed"
		return conflict("actual judgment model revision differs from authorized version")
	}
	result.Selected, result.Value = out.Selected, out.Value
	result.Distribution, result.ProviderConfidence = out.Distribution, out.Confidence
	result.ModelRevision, result.ProviderModel = &out.ModelRevision, &out.ProviderModel
	result.Truncated, result.Options, result.Usage = out.Truncated, out.Options, &out.Usage
	switch {
	case out.Refused:
		result.ReasonCode = "provider_abstained"
	case out.Truncated:
		result.ReasonCode = "input_truncated"
	case incompleteOptions(out.Options):
		result.ReasonCode = "options_incomplete"
	case out.Usage.InputTokens > request.Budget.MaxInputTokens || out.Usage.OutputTokens > request.Budget.MaxOutputTokens:
		result.ReasonCode = "token_budget_exceeded"
	case request.Question.Type == Truth:
		if policy.TruthAcceptMin == nil {
			result.ReasonCode = "truth_threshold_unconfigured"
			break
		}
		if *out.Value < *policy.TruthAcceptMin || *out.Value > *policy.TruthAcceptMax {
			result.ReasonCode = "truth_outside_accepted_range"
			break
		}
		result.Decision, result.ReasonCode = Recommend, "explicit_truth_policy"
	default:
		if policy.MinConfidence == nil {
			result.ReasonCode = "confidence_threshold_unconfigured"
			break
		}
		if out.Confidence == nil || *out.Confidence < *policy.MinConfidence {
			result.ReasonCode = "low_provider_confidence"
			break
		}
		result.Decision, result.ReasonCode = Recommend, "explicit_confidence_policy"
	}
	return nil
}

func incompleteOptions(options map[string]OptionCoverage) bool {
	for _, stats := range options {
		if stats.Distinct < stats.Total {
			return true
		}
	}
	return false
}

func (s *Service) Read(ctx context.Context, actor Actor, id string) (JudgmentResult, error) {
	var result JudgmentResult
	err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		record, err := s.repo.LoadResult(ctx, id)
		if err != nil {
			return err
		}
		if err := s.recordAccess(ctx, actor, record); err != nil {
			return err
		}
		result = record.Result
		return nil
	})
	return result, err
}

func (s *Service) recordAccess(ctx context.Context, actor Actor, record Record) error {
	if actor.ProjectID != "" && actor.ProjectID != record.Request.ProjectID {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "judgment record belongs to a different project"}
	}
	if actor.Kind != "human" && (actor.ID != record.Actor.ID || actor.Kind != record.Actor.Kind) {
		return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "judgment record belongs to a different actor"}
	}
	access, ok := s.access.(RecordAuthorizer)
	if !ok {
		return apierrors.NewUnsupported("judgment_history_authorization")
	}
	return access.AuthorizeRecord(ctx, actor, record.Request)
}

func (s *Service) Feedback(ctx context.Context, actor Actor, id string, version int, kind, reason string) (Feedback, error) {
	var feedback Feedback
	if actor.Kind != "human" || actor.ID == "" {
		return feedback, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "judgment feedback requires an authenticated human"}
	}
	if kind != "adopt" && kind != "later" && kind != "reject" && kind != "revise" && kind != "already_solved" {
		return feedback, invalid("unsupported feedback kind")
	}
	if strings.TrimSpace(reason) == "" {
		return feedback, invalid("feedback requires a human reason")
	}
	err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		record, err := s.repo.LoadResult(ctx, id)
		if err != nil {
			return err
		}
		if record.Result.Version != version {
			return conflict("feedback targets an old judgment result version")
		}
		if err := s.recordAccess(ctx, actor, record); err != nil {
			return err
		}
		feedback = Feedback{ID: uuid.NewString(), ResultID: id, ResultVersion: version, ActorID: actor.ID, ActorKind: actor.Kind, Kind: kind, Reason: reason, CreatedAt: time.Now().UTC()}
		return s.repo.SaveFeedback(ctx, feedback)
	})
	return feedback, err
}

func (s *Service) FeedbackHistory(ctx context.Context, actor Actor, id string) ([]Feedback, error) {
	var history []Feedback
	err := s.repo.WithJudgmentTx(ctx, func(ctx context.Context) error {
		record, err := s.repo.LoadResult(ctx, id)
		if err != nil {
			return err
		}
		if err := s.recordAccess(ctx, actor, record); err != nil {
			return err
		}
		history, err = s.repo.ListFeedback(ctx, id)
		return err
	})
	return history, err
}

func ValidateModelResult(q Question, out BackendResult) error {
	if out.ProviderModel == "" || out.ModelRevision == "" || out.Usage.InputTokens < 0 || out.Usage.OutputTokens < 0 {
		return invalid("model identity or usage is missing or invalid")
	}
	if out.Confidence != nil && !unit(*out.Confidence) {
		return invalid("provider confidence is invalid")
	}
	if q.Type == Truth {
		if out.Value == nil || !unit(*out.Value) || out.Selected != nil {
			return invalid("truth requires an actual [0,1] value")
		}
		return nil
	}
	expected := map[string]bool{}
	if q.Type == Choice {
		for _, option := range q.Options {
			expected[option.ID] = true
		}
	}
	if q.Type == Score {
		for i := range q.Levels {
			expected[fmt.Sprint(i)] = true
		}
	}
	if len(out.Distribution) != len(expected) {
		return invalid("provider distribution omits or adds options")
	}
	sum := 0.0
	for option, probability := range out.Distribution {
		if !expected[option] || !unit(probability) {
			return invalid("provider distribution has unknown options or invalid values")
		}
		sum += probability
	}
	// SDK 对概率保留四位小数，误差界来自候选数量与显示精度。
	tolerance := float64(len(expected))*0.00005 + 1e-9
	if sum < 1-tolerance || sum > 1+tolerance {
		return invalid("provider probabilities do not sum to one")
	}
	if out.Confidence == nil {
		return invalid("choice/score confidence is missing")
	}
	if q.Type == Choice {
		if out.Selected == nil || !expected[*out.Selected] || out.Value != nil {
			return invalid("provider choice is missing or outside candidates")
		}
	} else if out.Value == nil || math.IsNaN(*out.Value) || math.IsInf(*out.Value, 0) || *out.Value < 0 || *out.Value > float64(len(q.Levels)-1) {
		return invalid("provider score is missing or outside rubric")
	}
	return nil
}
