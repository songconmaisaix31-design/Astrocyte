package judgmentadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

const JevModel = "jev-1.13.0"
const jevEndpoint = "https://api.typesafe.ai/v1/systemone"

type Jev struct {
	key    string
	client *http.Client
}

type ProviderError struct {
	Status int
	Body   []byte
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("Jev evaluation returned HTTP %d; original body retained by caller", e.Status)
}

func NewJev(apiKey string) (*Jev, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("TYPESAFE_API_KEY is missing")
	}
	// 禁止重定向外发鉴权信息；标准客户端不配置重试。
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return fmt.Errorf("Jev redirects are not authorized") }}
	return &Jev{key: apiKey, client: client}, nil
}

func (*Jev) Backend() judgment.Backend { return judgment.JevBackend }

type providerAnswer struct {
	Type          string             `json:"type"`
	Choice        *string            `json:"choice"`
	Score         *float64           `json:"score"`
	Noul          *float64           `json:"noul"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    *float64           `json:"confidence"`
	Legend        map[string]string  `json:"legend"`
}

type providerReply struct {
	Model   string                    `json:"model"`
	Answers map[string]providerAnswer `json:"answers"`
	Usage   json.RawMessage           `json:"usage"`
}

func providerQuestion(q judgment.Question) map[string]any {
	result := map[string]any{"type": string(q.Type), "instructions": q.Instructions}
	switch q.Type {
	case judgment.Choice:
		criteria := map[string]string{}
		for _, option := range q.Options {
			criteria[option.ID] = option.Description
		}
		result["criteria"] = criteria
	case judgment.Score:
		result["criteria"] = q.Levels
	case judgment.Truth:
		result["type"] = "noul"
	}
	return result
}

func (j *Jev) Evaluate(ctx context.Context, input judgment.BackendInput) (judgment.BackendResult, error) {
	var result judgment.BackendResult
	if err := judgment.ValidateBackendInputBytes(input); err != nil {
		return result, err
	}
	if input.ModelRevision != JevModel {
		return result, fmt.Errorf("Jev requires pinned %s", JevModel)
	}
	body, err := json.Marshal(map[string]any{"model": JevModel, "state": input.State, "questions": map[string]any{input.Question.ID: providerQuestion(input.Question)}})
	if err != nil {
		return result, err
	}
	if len(body) > input.Budget.MaxInputBytes {
		return result, fmt.Errorf("Jev complete REST input exceeds explicit byte budget")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, jevEndpoint, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	request.Header.Set("Authorization", "Bearer "+j.key)
	request.Header.Set("Content-Type", "application/json")
	response, err := j.client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	if response.StatusCode != http.StatusOK {
		return result, &ProviderError{Status: response.StatusCode, Body: data}
	}
	var reply providerReply
	if err := json.Unmarshal(data, &reply); err != nil {
		return result, err
	}
	if reply.Model != JevModel {
		return result, fmt.Errorf("Jev response actual model %q differs from pinned %s", reply.Model, JevModel)
	}
	if len(reply.Answers) != 1 {
		return result, fmt.Errorf("Jev response must contain the requested question only")
	}
	answer, found := reply.Answers[input.Question.ID]
	if !found {
		return result, fmt.Errorf("Jev response omitted question %q", input.Question.ID)
	}
	var usage struct {
		Input  *int `json:"input_tokens"`
		Output *int `json:"output_tokens"`
	}
	if err := json.Unmarshal(reply.Usage, &usage); err != nil {
		return result, err
	}
	if usage.Input == nil || usage.Output == nil {
		return result, fmt.Errorf("Jev response omitted actual token usage")
	}
	result = judgment.BackendResult{
		Selected: answer.Choice, Distribution: answer.Probabilities, Confidence: answer.Confidence,
		ProviderModel: reply.Model, ModelRevision: reply.Model, Options: map[string]judgment.OptionCoverage{},
		Usage: judgment.Usage{InputTokens: *usage.Input, OutputTokens: *usage.Output},
	}
	switch input.Question.Type {
	case judgment.Choice:
		if answer.Type != "choice" || answer.Score != nil || answer.Noul != nil {
			return result, fmt.Errorf("Jev choice answer type differs")
		}
	case judgment.Score:
		if answer.Type != "score" || answer.Choice != nil || answer.Noul != nil {
			return result, fmt.Errorf("Jev score answer type differs")
		}
		result.Value = answer.Score
		for i, level := range input.Question.Levels {
			if answer.Legend[fmt.Sprint(i)] != level {
				return result, fmt.Errorf("Jev score legend differs from requested rubric")
			}
		}
	case judgment.Truth:
		if answer.Type != "noul" || answer.Choice != nil || answer.Score != nil || answer.Confidence != nil {
			return result, fmt.Errorf("Jev Noul answer contains incompatible fields")
		}
		result.Value = answer.Noul
		result.Distribution = map[string]float64{}
		result.Confidence = nil
	default:
		return result, fmt.Errorf("unsupported Jev question type")
	}
	return result, judgment.ValidateModelResult(input.Question, result)
}
