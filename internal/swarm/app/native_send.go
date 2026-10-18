package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type NativeSendCheck struct {
	Input      PreparedMemberInput `json:"-"`
	BindingID  string              `json:"-"`
	MissionID  string              `json:"-"`
	WorkItemID string              `json:"-"`
	Fence      int64               `json:"-"`
}

// 外部成员观察先完成，最终授权与调用方的单次发送占用共用同一主库事务。
func (s *Service) WithAuthorizedNativeSend(ctx context.Context, check NativeSendCheck, fn func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.members == nil || fn == nil || check.BindingID == "" || check.MissionID == "" || check.WorkItemID == "" || check.Fence < 1 {
		return domain.Failure(apierrors.EvidenceMissing, "send requires its first bound Swarm preparation and lease")
	}
	input := check.Input
	p := domain.Principal{Kind: input.ActorKind, ID: input.ActorID, ProjectID: input.ActorProjectID}
	var original domain.Lease
	var mission domain.Mission
	if err := s.repo.Within(ctx, func(tx Transaction) error {
		var err error
		original, err = tx.Lease(check.WorkItemID)
		if err != nil {
			return err
		}
		mission, err = tx.Mission(check.MissionID)
		return err
	}); err != nil {
		return err
	}
	member := Member{Purpose: "claim_work_item", BindingID: check.BindingID, OperationID: input.OperationID, HolderID: p.ID, ProjectID: p.ProjectID, SessionID: input.SessionID, Worktree: input.Root, Version: original.MemberVersion}
	var descriptor ExecutionDescriptor
	if err := s.withExecutionForGrant(ctx, p, mission.Grant, member, func(ctx context.Context, _ domain.Grant, current ExecutionDescriptor) error {
		descriptor = current
		return ValidatePreparedMemberInput(ctx, p, member, input)
	}); err != nil {
		return err
	}
	ctx = context.WithValue(ctx, executionCheckKey{}, executionCheck{operationID: input.OperationID, executionRoot: input.Root, bindingID: check.BindingID, descriptor: descriptor})
	return s.authorized(ctx, p, check.MissionID, "claim_work_item", domain.Resources{}, func(ctx context.Context, grant domain.Grant) error {
		binding, err := AuthorizedBindingForMember(ctx, p, member)
		if err != nil {
			return err
		}
		if err := ValidatePreparedMemberInput(ctx, p, member, input); err != nil {
			return err
		}
		if err := s.members.CheckMember(ctx, p, member); err != nil {
			return err
		}
		return s.repo.Within(ctx, func(tx Transaction) error {
			lease, err := tx.Lease(check.WorkItemID)
			if err != nil {
				return err
			}
			mission, err := tx.Mission(check.MissionID)
			if err != nil {
				return err
			}
			work, err := tx.WorkItem(check.WorkItemID)
			if err != nil {
				return err
			}
			if lease.OperationID != input.OperationID || lease.BindingID != binding.ID || lease.MemberVersion != member.Version || lease.ContextPacketID != binding.PacketID || lease.SessionID != input.SessionID || lease.NativeSessionID != input.NativeSessionID || lease.AdapterID != input.AdapterID || lease.ControllerID != input.ControllerID || lease.StartupOperationID != input.StartupOperationID || lease.Worktree != input.Root || lease.ProjectID != input.ProjectID || mission.Grant.ID != grant.ID || mission.Grant.Epoch != grant.Epoch || mission.Grant.ProposalRevision != grant.ProposalRevision {
				return domain.Failure(apierrors.LeaseLost, "current send lease differs from the original preparation or approval")
			}
			if err := lease.CheckNativeSend(mission, work, p.ID, check.Fence, time.Now().UTC()); err != nil {
				return err
			}
			deadline := lease.ExpiresAt
			for _, end := range []time.Time{grant.ExpiresAt, grant.Limits.Deadline} {
				if end.Before(deadline) {
					deadline = end
				}
			}
			if grant.Limits.NoProgressSeconds > 0 {
				end := mission.LastProgressAt.Add(time.Duration(grant.Limits.NoProgressSeconds) * time.Second)
				if end.Before(deadline) {
					deadline = end
				}
			}
			bounded, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			return fn(bounded)
		})
	})
}
