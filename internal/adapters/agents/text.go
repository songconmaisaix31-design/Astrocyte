package agents

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 来源及项目许可由调用者核实，此接口只处理调用者已经选择的文本。
func (r *Registry) textSession(ctx context.Context, cli, prompt string, seconds int) (*Native, domain.NativeSession, error) {
	a, err := r.Adapter(cli)
	if err != nil {
		return nil, domain.NativeSession{}, err
	}
	n, ok := a.(*Native)
	if !ok {
		return nil, domain.NativeSession{}, nativeError(apierrors.UnsupportedCapability, "this CLI adapter does not provide the selected-text processing path")
	}
	root, err := selectedTextRoot(n.stateRoot)
	if err != nil {
		return nil, domain.NativeSession{}, err
	}
	// 保留专属目录供错误核查，清理目录不改变原生会话身份。
	id := uuid.NewString()
	p := domain.LocalProject{ID: id, Root: root, Settings: domain.ProjectSettings{Revision: 1}}
	s, err := n.Start(ctx, domain.NativeRequest{SessionID: id, Project: p, Session: domain.NativeSession{ID: id, ProjectID: id}, Command: domain.NativeCommand{Message: prompt, DeadlineSeconds: seconds}, Packet: domain.ContextPacket{SchemaVersion: 1, ProjectID: id, Mode: "selected_text"}})
	return n, s, err
}

func selectedTextRoot(stateRoot string) (string, error) {
	if !filepath.IsAbs(stateRoot) {
		return "", nativeError(apierrors.ValidationFailed, "absolute controller-owned native state directory required")
	}
	parent := filepath.Join(stateRoot, "selected-text")
	if err := os.MkdirAll(parent, 0700); err != nil {
		return "", err
	}
	return os.MkdirTemp(parent, "astrocyte-selected-text-")
}

func (r *Registry) ConfigurationID(ctx context.Context, cli string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	a, err := r.Adapter(cli)
	if err != nil {
		return "", err
	}
	n, ok := a.(*Native)
	if !ok {
		return "", nativeError(apierrors.UnsupportedCapability, "this CLI adapter does not expose a model configuration identity")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.configID == "" || time.Since(n.configAt) > 5*time.Minute {
		return "", nativeError(apierrors.EvidenceMissing, "native configuration is unobserved or stale; explicitly probe the approved CLI")
	}
	return n.configID, nil
}

func (r *Registry) ProcessSelectedText(ctx context.Context, input domain.TextRequest) (result domain.TextResult, resultErr error) {
	if err := ValidateSelectedTextPrompt(input.Prompt); err != nil {
		return result, err
	}
	seconds := input.DeadlineSeconds
	if seconds == 0 {
		seconds = 180
	}
	if seconds < 1 || seconds > 1800 {
		return result, nativeError(apierrors.ValidationFailed, "selected text deadline is outside 1 to 1800 seconds")
	}
	limit := input.OutputLimit
	if limit == 0 {
		limit = 64 * 1024
	}
	if limit < 1 || limit > 128*1024 {
		return result, nativeError(apierrors.ValidationFailed, "selected text output limit is invalid")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	defer cancel()
	expectedConfig, err := r.ConfigurationID(ctx, input.CLI)
	if err != nil {
		return result, err
	}
	n, s, err := r.textSession(ctx, input.CLI, "", seconds)
	if err != nil {
		return result, err
	}
	var source *selectedTextSource
	defer func() {
		if source != nil {
			if err := source.retainOriginal(); err != nil {
				resultErr = errors.Join(resultErr, nativeError(apierrors.EvidenceMissing, "selected-text original source retention failed"))
			}
		}
		obs, err := n.Stop(context.Background(), s)
		if err != nil || !obs.StopConfirmed {
			resultErr = errors.Join(resultErr, nativeError(apierrors.DeliveryUnknown, "selected-text owned process stop remains unconfirmed"))
		}
	}()
	p, err := n.get(s)
	if err != nil {
		return result, err
	}
	actualConfig := n.processConfiguration(p)
	if actualConfig == "" {
		return result, nativeError(apierrors.EvidenceMissing, "native protocol does not report the current model before a paid turn; selected-text processing is unavailable")
	}
	if actualConfig != expectedConfig {
		return result, nativeError(apierrors.ContextStale, "native model configuration changed; refresh before processing")
	}
	source, err = n.selectedTextSource(ctx, s, input)
	if err != nil {
		return result, err
	}
	wirePrompt := packetPrompt(s.ContextPacket, input.Prompt)
	if len(wirePrompt) > 512*1024 {
		return result, nativeError(apierrors.ValidationFailed, "final selected text input exceeds512KiB")
	}
	// 诊断只记录大小，不记录文本、路径、提供者身份、凭据或对话。
	slog.Info("selected_text_input", "input_bytes", len(wirePrompt), "input_limit_bytes", 512*1024, "output_limit_bytes", limit)
	source.Phase = "send_attempted"
	if err = source.save(); err != nil {
		return result, err
	}
	if _, err := n.Send(ctx, s, wirePrompt); err != nil {
		return result, err
	}
	p.mu.Lock()
	source.TurnID = p.turnID
	p.mu.Unlock()
	if n.id == "codex" && source.TurnID == "" {
		return result, nativeError(apierrors.DeliveryUnknown, "native turn acceptance has no original turn identity")
	}
	source.Phase = "send_accepted"
	if err = source.save(); err != nil {
		return result, err
	}
	// 仅在原生 Send 实际受理后发布安全关联，原档路径与输出只在本机文件。
	slog.Info("selected_text_send_accepted", "project_id", input.ProjectID, "job_id", input.JobID, "operation_id", input.OperationID, "session_id", s.ID, "native_id", s.NativeID, "owned_process_id", source.Execution.ProcessID, "evidence_file", source.file)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		obs, err := n.Observe(ctx, s)
		if err != nil {
			return result, err
		}
		if obs.OutputTruncated || obs.Status == "blocked" {
			return result, nativeError(apierrors.BudgetExhausted, "native output exceeded bounds")
		}
		if obs.Status == "failed" || (obs.StopConfirmed && obs.Status != "completed") {
			return result, nativeError(apierrors.ProviderUnavailable, "native selected-text processing failed")
		}
		if obs.Status == "completed" {
			var text strings.Builder
			for _, event := range obs.Events {
				if event.Kind == "text" {
					text.WriteString(event.Text)
				}
			}
			if text.Len() == 0 || text.Len() > limit {
				return result, nativeError(apierrors.EvidenceMissing, "native output is empty or exceeds selected-text bound")
			}
			result.Text = text.String()
			result.NativeID = s.NativeID
			result.Version = s.Version
			p.mu.Lock()
			if p.model != "" {
				model := p.model
				result.Model = &model
			}
			p.mu.Unlock()
			// 此投影尚无原生 usage/cost 观测，保留 null。
			return result, nil
		}
		select {
		case <-ctx.Done():
			return result, nativeError(apierrors.DeliveryUnknown, "native selected-text result is unknown after deadline")
		case <-ticker.C:
		}
	}
}

// 在配置检查与原生启动之前核查包含固定指令和上下文的完整 UTF-8 输入。
func ValidateSelectedTextPrompt(prompt string) error {
	if !utf8.ValidString(prompt) || strings.TrimSpace(prompt) == "" || SelectedTextPromptBytes(prompt) > 512*1024 {
		return nativeError(apierrors.ValidationFailed, "selected text input including fixed instruction/context exceeds 512 KiB or is invalid UTF-8")
	}
	return nil
}

// SelectedTextPromptBytes 测量完整 UTF-8 输入，固定 UUID 长度为 36 bytes。
// 此纯函数不启动进程或记录文本。
func SelectedTextPromptBytes(prompt string) int {
	packet := domain.ContextPacket{SchemaVersion: 1, ProjectID: "00000000-0000-0000-0000-000000000000", Mode: "selected_text"}
	return len(packetPrompt(packet, prompt))
}
