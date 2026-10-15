package agents

import (
	"context"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 回调核查当前主数据库权限，返回真实 LocalProject.ID 并复用收到的事务 context。
type CompletionExperienceAccess func(context.Context, attention.Principal, string) (string, error)

type RelayCompletionExperienceSource struct {
	reader workspace.RelayCompletionEvidenceReader
	access CompletionExperienceAccess
}

var _ attention.ExperienceCompletionReader = (*RelayCompletionExperienceSource)(nil)

func NewRelayCompletionExperienceSource(reader workspace.RelayCompletionEvidenceReader, access CompletionExperienceAccess) (*RelayCompletionExperienceSource, error) {
	if reader == nil || access == nil {
		return nil, nativeError(apierrors.EvidenceMissing, "actual Relay completion and current project authorization readers are required")
	}
	return &RelayCompletionExperienceSource{reader: reader, access: access}, nil
}

func (s *RelayCompletionExperienceSource) currentOwner(ctx context.Context, principal attention.Principal, projectID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(principal.ID) == "" || principal.Kind != "human" && principal.Kind != "agent" || strings.TrimSpace(projectID) == "" {
		return "", nativeError(apierrors.ScopeDenied, "current authenticated completion project principal is required")
	}
	if principal.Kind == "agent" {
		return "", nativeError(apierrors.ScopeDenied, "Agent completion source action has not been authorized")
	}
	if s == nil || s.reader == nil || s.access == nil {
		return "", nativeError(apierrors.EvidenceMissing, "actual completion source authorization is not connected")
	}
	owner, err := s.access(ctx, principal, projectID)
	if err != nil {
		return "", completionExperienceCallbackFailure(err, "current completion project authorization failed")
	}
	if owner != projectID {
		return "", nativeError(apierrors.EvidenceMissing, "completion source owner must be its actual local project")
	}
	return owner, nil
}

// B 在事务外调用；只读原完成与最终收据，时间及引用不随观察或释放变化。
func (s *RelayCompletionExperienceSource) ReadExperienceCompletion(ctx context.Context, principal attention.Principal, projectID, completionID string) (attention.ExperienceCompletionEvidence, error) {
	var empty attention.ExperienceCompletionEvidence
	owner, err := s.currentOwner(ctx, principal, projectID)
	if err != nil {
		return empty, err
	}
	completion, err := s.reader.ReadRelayCompletion(ctx, projectID, completionID)
	if err != nil {
		return empty, completionExperienceCallbackFailure(err, "original finalized Relay completion is unavailable")
	}
	if completion.ID != completionID || completion.TaskID == "" || completion.WorktreeID == "" || completion.TerminalID == "" || completion.OperationID == "" || completion.ReceiptID == "" || completion.ActorID == "" || completion.ActorKind != "human" || completion.Operator != "human:"+completion.ActorID || completion.TS.IsZero() {
		return empty, nativeError(apierrors.EvidenceMissing, "Relay completion lacks its original human transition and identity")
	}
	if completion.Channel == domain.RelayVerifiedComplete {
		if completion.Source != "owned_output_and_goal_checks" || completion.VerifiedAt == nil || completion.VerifiedAt.IsZero() || completion.EvidenceID == "" || completion.InputOperationID == "" || len(completion.OutputEventIDs) == 0 || len(completion.CheckEvidenceIDs) == 0 {
			return empty, nativeError(apierrors.EvidenceMissing, "verified completion lacks original output and goal check references")
		}
	} else if completion.Channel != domain.RelayUserConfirmedComplete || completion.Source != "human_confirmation" {
		return empty, nativeError(apierrors.EvidenceMissing, "Relay completion channel and original source differ")
	}
	// Version 表示该不可变完成记录的首个版本；不读取当前 task revision。
	value := attention.ExperienceCompletionEvidence{
		SourceKind: attention.ExperienceSourceRelayCompletion, EventID: completion.ID, Version: 1,
		OccurredAt: completion.TS.UnixMilli(), LocalProjectID: projectID, OwnerProject: owner,
		ActorID: completion.ActorID, ActorKind: completion.ActorKind, TaskID: completion.TaskID,
		Outcome: "success", OriginEventIDs: append([]string(nil), completion.OutputEventIDs...),
		CompletionChannel: completion.Channel, CompletionReceiptID: completion.ReceiptID,
	}
	// Outcome 属于该完成通道的目标声明，不描述原生回合终态。
	if err = attention.ValidateCompletionSource(value, projectID, completionID, owner); err != nil {
		return empty, err
	}
	return value, nil
}

// B 在同一主数据库短事务中调用；不读取 Relay、原档、对象或控制器。
func (s *RelayCompletionExperienceSource) AuthorizeExperienceCompletion(ctx context.Context, principal attention.Principal, fixed attention.ExperienceCompletionEvidence) error {
	if fixed.SourceKind != attention.ExperienceSourceRelayCompletion || fixed.ActorKind != "human" {
		return nativeError(apierrors.ScopeDenied, "fixed completion does not belong to the Relay human completion source")
	}
	owner, err := s.currentOwner(ctx, principal, fixed.LocalProjectID)
	if err != nil {
		return err
	}
	return attention.ValidateCompletionSource(fixed, fixed.LocalProjectID, fixed.EventID, owner)
}
