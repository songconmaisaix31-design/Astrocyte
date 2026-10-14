package judgmentadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	"github.com/sourcegraph/jsonrpc2"
)

type LayaConfig struct {
	Node              string
	Worker            string
	ModelDir          string
	AssetsPath        string
	SDKEntry          string
	RPCEntry          string
	Language          string
	ModelRevision     string
	Stderr            io.Writer
	Environment       []string
	StartupTimeoutMS  int64
	ShutdownTimeoutMS int64
}

type Laya struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	conn        *jsonrpc2.Conn
	config      LayaConfig
	processDone chan struct{}
	processErr  error
	closed      bool
}

type processStream struct {
	io.Reader
	writer io.WriteCloser
	reader io.Closer
}

func (s *processStream) Write(data []byte) (int, error) { return s.writer.Write(data) }
func (s *processStream) Close() error                   { return errors.Join(s.writer.Close(), s.reader.Close()) }

type LayaHealth struct {
	ModelRevision             string   `json:"model_revision"`
	ProviderModel             string   `json:"provider_model"`
	Device                    string   `json:"device"`
	Checkpoint                string   `json:"checkpoint"`
	Language                  string   `json:"language"`
	SDKReportedRevision       *string  `json:"sdk_reported_revision"`
	InvalidCalibrationBuckets []string `json:"invalid_calibration_buckets"`
	PID                       int      `json:"pid"`
}

func NewLaya(ctx context.Context, config LayaConfig) (*Laya, error) {
	if config.Node == "" || config.Language == "" || config.ModelRevision == "" || config.Stderr == nil || config.StartupTimeoutMS <= 0 || config.ShutdownTimeoutMS <= 0 {
		return nil, fmt.Errorf("Laya requires explicit Node, language, model revision and stderr destination")
	}
	for _, path := range []string{config.Worker, config.ModelDir, config.AssetsPath, config.SDKEntry, config.RPCEntry} {
		if !filepath.IsAbs(path) {
			return nil, fmt.Errorf("Laya paths must be explicit absolute configured paths")
		}
	}
	cmd := exec.CommandContext(ctx, config.Node, config.Worker, "--model-dir", config.ModelDir, "--assets", config.AssetsPath, "--sdk-entry", config.SDKEntry, "--rpc-entry", config.RPCEntry, "--language", config.Language)
	configureLayaProcess(cmd)
	cmd.Stderr = config.Stderr
	cmd.Env = append(os.Environ(), config.Environment...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}
	l := &Laya{cmd: cmd, config: config, processDone: make(chan struct{})}
	l.conn = jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(&processStream{Reader: stdout, reader: stdout, writer: stdin}, jsonrpc2.VSCodeObjectCodec{}), nil)
	go func() { l.processErr = cmd.Wait(); close(l.processDone) }()
	startup, startupCancel := context.WithTimeout(ctx, time.Duration(config.StartupTimeoutMS)*time.Millisecond)
	defer startupCancel()
	health, err := l.Health(startup)
	if err != nil {
		return nil, errors.Join(err, l.Close())
	}
	if health.Device != "cpu" || health.Language != config.Language || health.ModelRevision != config.ModelRevision || health.ProviderModel != "laya-rl-agent" || health.PID != cmd.Process.Pid || health.SDKReportedRevision != nil {
		return nil, errors.Join(fmt.Errorf("actual Laya worker/model/language/CPU identity differs from configured assets"), l.Close())
	}
	return l, nil
}

func (*Laya) Backend() judgment.Backend { return judgment.LayaBackend }

func (l *Laya) Health(ctx context.Context) (LayaHealth, error) {
	var health LayaHealth
	if err := l.conn.Call(ctx, "judgment/health", nil, &health); err != nil {
		return health, err
	}
	return health, nil
}

func (l *Laya) Evaluate(ctx context.Context, input judgment.BackendInput) (judgment.BackendResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var result judgment.BackendResult
	if err := judgment.ValidateBackendInputBytes(input); err != nil {
		return result, err
	}
	if l.closed {
		return result, fmt.Errorf("Laya worker is closed; explicit restart is required")
	}
	if input.Language != l.config.Language || input.ModelRevision != l.config.ModelRevision {
		return result, fmt.Errorf("Laya input language or pinned model revision differs")
	}
	var reply struct {
		ModelRevision string        `json:"model_revision"`
		Raw           providerReply `json:"raw"`
	}
	if err := l.conn.Call(ctx, "judgment/evaluate", input, &reply); err != nil {
		// 在途调用超时后关闭该 worker，后续请求不能重用未知状态或自动重发。
		if ctx.Err() != nil {
			return result, errors.Join(err, l.close())
		}
		return result, err
	}
	if reply.ModelRevision != l.config.ModelRevision || reply.Raw.Model != "laya-rl-agent" || len(reply.Raw.Answers) != 1 {
		return result, fmt.Errorf("actual Laya answer model, revision or question count differs")
	}
	answer, found := reply.Raw.Answers[input.Question.ID]
	if !found {
		return result, fmt.Errorf("Laya omitted the requested question")
	}
	var usage struct {
		InputTokens        *int                               `json:"input_tokens"`
		OutputTokens       *int                               `json:"output_tokens"`
		StateTokens        *int                               `json:"state_tokens"`
		StateTokensDropped *int                               `json:"state_tokens_dropped"`
		Truncated          *bool                              `json:"truncated"`
		TruncatedQuestions []string                           `json:"truncated_questions"`
		Options            map[string]judgment.OptionCoverage `json:"options"`
	}
	if err := json.Unmarshal(reply.Raw.Usage, &usage); err != nil {
		return result, err
	}
	if usage.InputTokens == nil || usage.OutputTokens == nil || usage.StateTokens == nil || usage.StateTokensDropped == nil || usage.Truncated == nil || *usage.StateTokens < 0 || *usage.StateTokensDropped < 0 || *usage.StateTokensDropped > *usage.StateTokens || *usage.Truncated != (*usage.StateTokensDropped > 0) {
		return result, fmt.Errorf("Laya usage or truncation metadata is missing or inconsistent")
	}
	options := usage.Options
	if options == nil {
		options = map[string]judgment.OptionCoverage{}
	}
	result = judgment.BackendResult{
		Selected: answer.Choice, Confidence: answer.Confidence, Distribution: answer.Probabilities,
		ProviderModel: reply.Raw.Model, ModelRevision: reply.ModelRevision, Truncated: *usage.Truncated, Options: options,
		Usage: judgment.Usage{InputTokens: *usage.InputTokens, OutputTokens: *usage.OutputTokens, StateTokens: usage.StateTokens, StateTokensDropped: usage.StateTokensDropped, Truncated: usage.Truncated, TruncatedQuestions: usage.TruncatedQuestions, Options: options},
	}
	switch input.Question.Type {
	case judgment.Choice:
		if answer.Type != "choice" || answer.Score != nil || answer.Noul != nil {
			return result, fmt.Errorf("Laya choice answer type differs")
		}
	case judgment.Score:
		if answer.Type != "score" || answer.Noul != nil || answer.Choice != nil {
			return result, fmt.Errorf("Laya score answer type differs")
		}
		result.Value = answer.Score
		for i, level := range input.Question.Levels {
			if answer.Legend[fmt.Sprint(i)] != level {
				return result, fmt.Errorf("Laya score rubric differs")
			}
		}
	case judgment.Truth:
		if answer.Type != "noul" || answer.Score != nil || answer.Choice != nil {
			return result, fmt.Errorf("Laya Noul answer type differs")
		}
		result.Value = answer.Noul
		result.Distribution = map[string]float64{}
	default:
		return result, fmt.Errorf("unsupported Laya question type")
	}
	return result, judgment.ValidateModelResult(input.Question, result)
}

func (l *Laya) close() error {
	if l.closed {
		return nil
	}
	l.closed = true
	streamErr := l.conn.Close()
	select {
	case <-l.processDone:
	case <-time.After(time.Duration(l.config.ShutdownTimeoutMS) * time.Millisecond):
		killErr := l.cmd.Process.Kill()
		<-l.processDone
		return errors.Join(streamErr, fmt.Errorf("Laya owned process did not exit within explicit shutdown budget"), killErr, l.processErr)
	}
	return errors.Join(streamErr, l.processErr)
}

func (l *Laya) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}
