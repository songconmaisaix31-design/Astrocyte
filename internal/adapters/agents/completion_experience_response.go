package agents

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	attentiondomain "github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (p *nativeProcess) completionExperienceSchema() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.completionExperience == nil {
		return nil
	}
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"pattern", "verdict", "condition", "evidence", "refused"},
		"properties": map[string]any{
			"pattern":   map[string]any{"type": "string"},
			"verdict":   map[string]any{"type": "string", "enum": []string{"avoid", "prefer", "condition"}},
			"condition": map[string]any{"type": "string"},
			"evidence":  map[string]any{"type": "string", "enum": []string{p.completionExperience.evidence.EventID}},
			"refused":   map[string]any{"type": "boolean"},
		},
	}
}

// 由原接收器在进程锁内调用；仅保存本次完整最终消息，不拼接 delta。
func (p *nativeProcess) observeCompletionExperience(method string, record map[string]json.RawMessage) error {
	operation := p.completionExperience
	if method != "item/completed" && method != "turn/completed" && method != "model/rerouted" {
		return nil
	}
	if method == "model/rerouted" {
		operation.failure = nativeError(apierrors.ContextStale, "completion model routing changed during the original operation")
		return operation.failure
	}
	var params struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Item     struct {
			ID    string  `json:"id"`
			Type  string  `json:"type"`
			Phase *string `json:"phase"`
			Text  string  `json:"text"`
		} `json:"item"`
		Turn struct {
			ID     string          `json:"id"`
			Status string          `json:"status"`
			Error  json.RawMessage `json:"error"`
		} `json:"turn"`
	}
	if err := decodeNativeValue(record["params"], &params); err != nil {
		return err
	}
	turn := params.TurnID
	if method == "turn/completed" {
		turn = params.Turn.ID
	}
	if params.ThreadID != p.nativeID || turn == "" || p.inputOperationID != operation.operationID || p.turnID != "" && p.turnID != turn || operation.turnID != "" && operation.turnID != turn {
		operation.failure = nativeError(apierrors.ScopeDenied, "completion response differs from its original controller turn")
		return operation.failure
	}
	operation.turnID = turn
	if method == "item/completed" {
		if params.Item.Type != "agentMessage" && params.Item.Type != "reasoning" && params.Item.Type != "userMessage" {
			operation.failure = nativeError(apierrors.ScopeDenied, "completion observed an unexpected item under the tools-disabled policy")
			return operation.failure
		}
		if params.Item.Type != "agentMessage" {
			return nil
		}
		if params.Item.Phase == nil || *params.Item.Phase != "final_answer" {
			return nil
		}
		if params.Item.ID == "" || params.Item.Text == "" || len(params.Item.Text) > 64*1024 || operation.itemID != "" && operation.itemID != params.Item.ID {
			operation.failure = nativeError(apierrors.EvidenceMissing, "completion lacks one bounded final native message")
			return operation.failure
		}
		if operation.itemID != "" && operation.text != params.Item.Text {
			operation.failure = nativeError(apierrors.ContextStale, "completed native message changed after observation")
			return operation.failure
		}
		operation.itemID, operation.text = params.Item.ID, params.Item.Text
		return nil
	}
	if params.Turn.Status != "completed" || len(params.Turn.Error) != 0 && string(params.Turn.Error) != "null" {
		operation.failure = nativeError(apierrors.ProviderUnavailable, "original completion turn did not finish successfully")
		return operation.failure
	}
	operation.completed = true
	return nil
}

func (i *CompletionExperienceInterpreter) readCompletionExperience(ctx context.Context, process *nativeProcess, operationID string) (attention.InterpretedExperienceDraft, error) {
	var empty attention.InterpretedExperienceDraft
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion result is not available before the deadline; do not resend")
		}
		process.mu.Lock()
		original := process.completionExperience
		if original == nil {
			process.mu.Unlock()
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion operation was not retained")
		}
		operation := *original
		truncated, status := process.truncated, process.status
		bound := false
		for _, turn := range process.relayEvidence.Turns {
			if turn.OperationID == operationID && turn.TurnID == operation.turnID && turn.NativeID == process.nativeID && turn.ControllerID == process.relayEvidence.ControllerID {
				bound = true
			}
		}
		process.mu.Unlock()
		if operation.failure != nil {
			return empty, operation.failure
		}
		if truncated || status == "blocked" {
			return empty, nativeError(apierrors.EvidenceMissing, "original completion stream is incomplete or truncated")
		}
		if operation.completed {
			if !bound || operation.itemID == "" || operation.model == "" || operation.provider == "" || operation.modelVersion == "" {
				return empty, nativeError(apierrors.EvidenceMissing, "original completion lacks final message, model provenance or accepted turn binding")
			}
			return decodeCompletionExperienceDraft(operation)
		}
		if status == "failed" || status == "interrupted" || status == "cancelled" || status == "stopped" {
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion has no verified complete response")
		}
		select {
		case <-ctx.Done():
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion remains unresolved; do not resend")
		case <-process.done:
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion controller exited without a verified final response")
		case <-ticker.C:
		}
	}
}

func decodeCompletionExperienceDraft(operation completionExperienceOperation) (attention.InterpretedExperienceDraft, error) {
	var empty attention.InterpretedExperienceDraft
	var output struct {
		Pattern   string `json:"pattern"`
		Verdict   string `json:"verdict"`
		Condition string `json:"condition"`
		Evidence  string `json:"evidence"`
		Refused   *bool  `json:"refused"`
	}
	decoder := json.NewDecoder(strings.NewReader(operation.text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&output); err != nil {
		return empty, nativeError(apierrors.EvidenceMissing, "original completion is not a valid structured experience draft")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return empty, nativeError(apierrors.EvidenceMissing, "original completion contains additional output beyond its draft")
	}
	if output.Refused == nil || *output.Refused || output.Evidence != operation.evidence.EventID || output.Evidence == "" || !attentiondomain.ValidVerdict(output.Verdict) || strings.TrimSpace(output.Pattern) == "" || strings.TrimSpace(output.Condition) == "" || len(output.Pattern) > 8192 || len(output.Condition) > 8192 {
		return empty, nativeError(apierrors.EvidenceMissing, "original completion refused or lacks a bounded draft and exact evidence reference")
	}
	complete := false
	versionSource := "unknown"
	if operation.modelVersion != "unknown" {
		versionSource = "provider_reported"
	}
	return attention.InterpretedExperienceDraft{
		Pattern: output.Pattern, Verdict: output.Verdict, Condition: output.Condition,
		Refused: output.Refused, Truncated: &complete,
		Provenance: attention.ExperienceInterpretationProvenance{Backend: "codex-app-server", Model: operation.model, ModelIdentitySource: "configured_route", ModelVersion: operation.modelVersion, ModelVersionSource: versionSource, PolicyVersion: completionExperiencePolicy, OperationID: operation.operationID},
	}, nil
}
