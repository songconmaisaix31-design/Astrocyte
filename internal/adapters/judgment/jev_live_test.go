package judgmentadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

type actualJevTransport struct {
	transport http.RoundTripper
	directory string
	key       string
	question  string
	input     string
}

func (r *actualJevTransport) save(suffix string, data []byte) error {
	if bytes.Contains(data, []byte(r.key)) {
		return errors.New("Jev receipt contains the credential; persistence refused")
	}
	file, err := os.OpenFile(filepath.Join(r.directory, r.question+suffix), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	return errors.Join(writeErr, file.Close())
}

func (r *actualJevTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodPost || request.URL.String() != jevEndpoint {
		return nil, errors.New("only the fixed official Jev evaluation endpoint is authorized")
	}
	if request.GetBody == nil {
		return nil, errors.New("the actual public request body is unavailable for the receipt")
	}
	body, err := request.GetBody()
	if err != nil {
		return nil, err
	}
	input, readErr := io.ReadAll(body)
	if err := errors.Join(readErr, body.Close(), r.save(".request.body", input)); err != nil {
		return nil, err
	}
	r.input = judgment.DigestText(string(input))
	// 实际请求交给标准 HTTP transport；保存公开正文和真实响应，不记录请求头。
	response, err := r.transport.RoundTrip(request)
	if err != nil {
		return nil, errors.Join(err, r.save(".transport-error.txt", []byte(err.Error())))
	}
	metadata, err := json.Marshal(struct {
		Status      int       `json:"http_status"`
		ContentType string    `json:"content_type"`
		ObservedAt  time.Time `json:"observed_at"`
	}{response.StatusCode, response.Header.Get("Content-Type"), time.Now().UTC()})
	if err != nil {
		return nil, errors.Join(err, response.Body.Close())
	}
	if err := r.save(".http.json", metadata); err != nil {
		return nil, errors.Join(err, response.Body.Close())
	}
	data, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close(), r.save(".response.body", data)); err != nil {
		return nil, err
	}
	response.Body = io.NopCloser(bytes.NewReader(data))
	return response, nil
}

func TestJevActualPinnedPublicThreeTypes(t *testing.T) {
	if os.Getenv("ASTROCYTE_S5_JEV_LIVE") != "1" {
		t.Skip("requires explicit authorization for the fixed public Jev task")
	}
	key := os.Getenv("TYPESAFE_API_KEY")
	if key == "" {
		t.Fatal("TYPESAFE_API_KEY is absent; no provider request was sent")
	}
	directory := os.Getenv("ASTROCYTE_S5_JEV_RUN_ROOT")
	if !filepath.IsAbs(directory) {
		t.Fatal("an explicit absolute owned receipt directory is required")
	}
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		t.Fatal("the owned receipt directory is unavailable")
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal("the fixed official public Laya README is unavailable")
	}
	client, err := NewJev(key)
	if err != nil {
		t.Fatal("the fixed Jev adapter could not be configured")
	}
	transport := &actualJevTransport{transport: http.DefaultTransport, directory: directory, key: key}
	client.client.Transport = transport
	questions := []judgment.Question{
		{ID: "official-export-format", Type: judgment.Choice, Instructions: "Which model format does the Export weights section produce?", Options: []judgment.Option{{ID: "onnx", Description: "Split encoder and head ONNX files"}, {ID: "safetensors", Description: "Only the source safetensors checkpoint"}}},
		{ID: "official-export-verification", Type: judgment.Score, Instructions: "How explicitly does the Export weights section describe numerical verification?", Levels: []string{"No verification described", "Verification mentioned", "Torch and ONNX agreement and tolerance described"}},
		{ID: "official-export-parity", Type: judgment.Truth, Instructions: "The documented export verifies torch and ONNX numerical agreement unless --no-verify is explicitly supplied."},
	}
	for _, question := range questions {
		transport.question = question.ID
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		input := judgment.BackendInput{Question: question, State: string(source), Language: "en", ModelRevision: JevModel, Budget: judgment.Budget{TimeoutMS: 60000, MaxInputBytes: 131072, MaxInputTokens: 8192, MaxOutputTokens: 256}}
		result, err := client.Evaluate(ctx, input)
		cancel()
		if err != nil {
			if saveErr := transport.save(".evaluation-error.txt", []byte(err.Error())); saveErr != nil {
				t.Fatal("actual Jev request or safe receipt persistence failed; no automatic retry")
			}
			t.Fatal("actual Jev evaluation failed; inspect the owned original receipt; no automatic retry")
		}
		if result.ProviderModel != JevModel || result.ModelRevision != JevModel || result.Usage.InputTokens <= 0 || result.Usage.InputTokens > input.Budget.MaxInputTokens || result.Usage.OutputTokens > input.Budget.MaxOutputTokens || question.Type == judgment.Truth && result.Confidence != nil {
			t.Fatal("actual Jev identity, token usage or nullable Noul confidence differs")
		}
		receipt, err := json.Marshal(struct {
			Source       string                 `json:"source"`
			InputDigest  string                 `json:"input_digest"`
			Question     judgment.Question      `json:"question"`
			ActualResult judgment.BackendResult `json:"actual_result"`
		}{"https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md", transport.input, question, result})
		if err != nil || transport.save(".result.json", receipt) != nil {
			t.Fatal("actual Jev normalized receipt could not be safely saved")
		}
		t.Logf("actual %s model=%s input_tokens=%d output_tokens=%d confidence_null=%t", question.Type, result.ProviderModel, result.Usage.InputTokens, result.Usage.OutputTokens, result.Confidence == nil)
	}
}
