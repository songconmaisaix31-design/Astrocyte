package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type CaptureCompletionExperienceCommand struct {
	CommandMeta
	LocalProjectID string `json:"local_project_id"`
	EventID        string `json:"event_id"`
}

// 原治理服务可单独装配；完成来源接口仅由实际来源和 Interpreter 组合提供。
type CompletionExperienceGovernance interface {
	CaptureCompletionExperience(context.Context, Principal, CaptureCompletionExperienceCommand) (StoredExperience, error)
	ResolveExperienceEvidence(context.Context, Principal, string, string) ([]SafeExperienceEvidence, error)
}

func (s *AcceptedExperienceService) CaptureCompletionExperience(ctx context.Context, p Principal, command CaptureCompletionExperienceCommand) (StoredExperience, error) {
	if err := experiencePrincipal(p, true); err != nil {
		return StoredExperience{}, err
	}
	if err := validateMeta(command.CommandMeta); err != nil {
		return StoredExperience{}, err
	}
	if command.ExpectedVersion != 1 || strings.TrimSpace(command.LocalProjectID) == "" || strings.TrimSpace(command.EventID) == "" {
		return StoredExperience{}, serviceError(apierrors.ValidationFailed, "Completion capture requires version one and actual registered project and completion identities", "select_actual_completion_source")
	}
	return s.captureCompletion(ctx, p, command.LocalProjectID, command.EventID, &command)
}

// HTTP 命令的幂等键只固定原请求；模型操作继续复用原完成事件的既有收据。
func bindCompletionExperienceCommand(tx AcceptedExperienceTx, p Principal, command CaptureCompletionExperienceCommand) error {
	digest, err := semanticDigest(command)
	if err != nil {
		return err
	}
	caller, operation := p.Kind+":"+p.ID, "Experience/CaptureCompletion/"+command.LocalProjectID
	receipt, err := tx.LoadExperienceReceipt(caller, operation, command.IdempotencyKey)
	if err == nil {
		if receipt.Digest != digest {
			return serviceError(apierrors.VersionConflict, "Idempotency key binds different completion input", "use_new_idempotency_key")
		}
		var original CaptureCompletionExperienceCommand
		if err := json.Unmarshal(receipt.Result, &original); err != nil {
			return err
		}
		if original.SchemaVersion != 1 || original.ExpectedVersion != 1 || original.LocalProjectID != command.LocalProjectID || original.EventID != command.EventID {
			return serviceError(apierrors.DeliveryUnknown, "Original completion command identity is unavailable", "reconcile_original_completion")
		}
		return nil
	}
	if !isMissing(err) {
		return err
	}
	data, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return tx.SaveExperienceReceipt(caller, operation, command.IdempotencyKey, Receipt{Digest: digest, Result: data})
}
