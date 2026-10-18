package app

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func (s *Service) PauseMission(ctx context.Context, p domain.Principal, c TransitionCommand) (domain.Mission, error) {
	return s.transition(ctx, p, c, "paused")
}
func (s *Service) CancelMission(ctx context.Context, p domain.Principal, c TransitionCommand) (domain.Mission, error) {
	return s.transition(ctx, p, c, "cancelled")
}

func (s *Service) transition(ctx context.Context, p domain.Principal, c TransitionCommand, target string) (domain.Mission, error) {
	var out domain.Mission
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if err := human(p); err != nil {
		return out, err
	}
	if c.Reason == "" {
		return out, domain.Failure(apierrors.ValidationFailed, "stop reason required")
	}
	var leases []domain.Lease
	fresh := false
	err := s.readWithin(ctx, p, c.MissionID, func(_ context.Context, tx Transaction) error {
		cached, found, d, err := replay[domain.Mission](tx, p, target+"_mission", c.Meta, c)
		if err != nil {
			return err
		}
		if found {
			out = cached
			return nil
		}
		out, err = tx.Mission(c.MissionID)
		if err != nil {
			return err
		}
		if err = checkVersion(out.Version, c.Meta.ExpectedVersion); err != nil {
			return err
		}
		if out.State == "completed" || out.State == "cancelled" {
			return domain.Failure(apierrors.VersionConflict, "terminal mission cannot change state")
		}
		old := out.Version
		out.Version++
		out.State = target
		out.Reason = c.Reason
		if err = tx.SaveMission(out, old); err != nil {
			return err
		}
		all, err := tx.ActiveLeases(out.Grant.ProjectID)
		if err != nil {
			return err
		}
		for _, l := range all {
			if l.MissionID == out.ID {
				leases = append(leases, l)
			}
		}
		if target == "cancelled" {
			rows, err := tx.WorkItems(out.ID)
			if err != nil {
				return err
			}
			for _, w := range rows {
				if w.State == "accepted" {
					continue
				}
				old := w.Version
				w.Version++
				w.State = "cancelled"
				w.Reason = c.Reason
				if err = tx.SaveWorkItem(w, old); err != nil {
					return err
				}
			}
		}
		fresh = true
		return finish(tx, p, target+"_mission", c.Meta, d, out, "mission_"+target, out.ID, out.Version)
	})
	if err != nil || !fresh {
		return out, err
	}
	for _, l := range leases {
		if s.ownership == nil {
			if err := s.blockExecutor(ctx, l, "native ownership service unavailable; original reservation retained"); err != nil {
				return out, err
			}
			return out, domain.Failure(apierrors.UnsupportedCapability, "actual native ownership is not configured")
		}
		fact, err := s.ownership.Stop(ctx, l.OperationID, l.SessionID)
		if err != nil {
			if recordErr := s.blockExecutor(context.WithoutCancel(ctx), l, "native stop failed; reconcile original operation"); recordErr != nil {
				return out, recordErr
			}
			return out, err
		}
		if err = s.settleExecutor(ctx, l, fact, target == "cancelled"); err != nil {
			return out, err
		}
	}
	return s.repo.Load(ctx, out.ID)
}

func (s *Service) blockExecutor(ctx context.Context, l domain.Lease, reason string) error {
	return s.repo.Within(ctx, func(tx Transaction) error {
		current, err := tx.Lease(l.WorkItemID)
		if err != nil {
			return err
		}
		if current.Fence != l.Fence || current.OperationID != l.OperationID || current.SessionID != l.SessionID || current.MissionID != l.MissionID || current.ProjectID != l.ProjectID || current.BindingID != l.BindingID || current.HolderID != l.HolderID {
			return domain.Failure(apierrors.LeaseLost, "original lease changed during reconciliation")
		}
		w, err := tx.WorkItem(l.WorkItemID)
		if err != nil {
			return err
		}
		oldW := w.Version
		w.Version++
		w.State = "blocked"
		w.Reason = reason
		if w.Fence == l.Fence {
			if w.Fence == math.MaxInt64 {
				return domain.Failure(apierrors.LeaseLost, "fencing token exhausted")
			}
			w.Fence++
		}
		m, err := tx.Mission(l.MissionID)
		if err != nil {
			return err
		}
		oldM := m.Version
		m.Version++
		if m.State != "cancelled" {
			m.State = "blocked"
		}
		m.Reason = reason
		if err = tx.SaveWorkItem(w, oldW); err != nil {
			return err
		}
		if err = tx.SaveMission(m, oldM); err != nil {
			return err
		}
		return tx.Event(Event{ID: uuid.NewString(), Type: "swarm.execution_blocked", AggregateID: w.ID, Version: w.Version, At: time.Now().UTC(), Payload: map[string]any{"operation_id": l.OperationID, "session_id": l.SessionID, "reason": reason, "fence": w.Fence}})
	})
}

func (s *Service) settleExecutor(ctx context.Context, l domain.Lease, fact ExecutionFact, cancelled bool) error {
	if fact.OperationID != l.OperationID || fact.SessionID != l.SessionID {
		return domain.Failure(apierrors.ContextStale, "original executor observation changed its operation or session identity")
	}
	if fact.StopConfirmed {
		persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		err := s.recordExecutorExit(persistCtx, l)
		cancel()
		if err != nil {
			return err
		}
	}
	if !safeFailureCode(fact.FailureCode) {
		return domain.Failure(apierrors.EvidenceMissing, "original execution failure requires a safe classified code")
	}
	if fact.Outcome == "completed" && fact.FailureCode != "" {
		return domain.Failure(apierrors.EvidenceMissing, "completed execution contradicts its failure classification")
	}
	usage, err := actualUsage(fact)
	if err != nil {
		return err
	}
	unknownOutcome := fact.Outcome == "" || fact.Outcome == "unknown"
	if !unknownOutcome && !settledOutcome(fact.Outcome) {
		return domain.Failure(apierrors.EvidenceMissing, "original execution outcome contradicts its supported result contract")
	}
	if !fact.StopConfirmed || unknownOutcome {
		if err := s.blockExecutor(ctx, l, "original execution or usage remains unknown; reservation and write scope retained"); err != nil {
			return err
		}
		return &retainedExecutionStop{Cause: domain.Failure(apierrors.DeliveryUnknown, "original executor must be reconciled; no replacement can claim its scope")}
	}
	before, err := s.repo.Load(ctx, l.MissionID)
	if err != nil {
		return err
	}
	if !usage.WithinReservation(l.Reserved, before.Grant.Budget) {
		if err = s.blockExecutor(ctx, l, "bounded usage evidence is missing or exceeds reservation; original scope retained"); err != nil {
			return err
		}
		failure := domain.Failure(apierrors.EvidenceMissing, "bounded actual usage must be reconciled before release")
		if unknownUsageWithinKnownBounds(usage, l.Reserved, before.Grant.Budget) {
			return &retainedExecutionStop{Cause: failure}
		}
		return failure
	}
	return s.repo.Within(ctx, func(tx Transaction) error {
		current, err := tx.Lease(l.WorkItemID)
		if err != nil {
			return err
		}
		if current.Fence != l.Fence || current.OperationID != l.OperationID || current.SessionID != l.SessionID || current.MissionID != l.MissionID || current.ProjectID != l.ProjectID || current.BindingID != l.BindingID || current.HolderID != l.HolderID {
			return domain.Failure(apierrors.LeaseLost, "original lease changed during reconciliation")
		}
		m, err := tx.Mission(l.MissionID)
		if err != nil {
			return err
		}
		remaining, err := m.Reserved.Sub(l.Reserved)
		if err != nil {
			return err
		}
		if err = m.RecordUsage(usage); err != nil {
			return err
		}
		oldM := m.Version
		m.Version++
		m.Reserved = remaining
		m.NoProgress++
		w, err := tx.WorkItem(l.WorkItemID)
		if err != nil {
			return err
		}
		oldW := w.Version
		w.Version++
		if w.Fence == l.Fence {
			if w.Fence == math.MaxInt64 {
				return domain.Failure(apierrors.LeaseLost, "fencing token exhausted")
			}
			w.Fence++
		}
		if cancelled || m.State == "cancelled" {
			w.State = "cancelled"
		} else {
			w.State = "revision_needed"
		}
		w.Reason = "original executor stopped; explicit new claim required"
		if err = tx.SaveMission(m, oldM); err != nil {
			return err
		}
		if err = tx.SaveWorkItem(w, oldW); err != nil {
			return err
		}
		if err = tx.DeleteLease(w.ID, l.Fence); err != nil {
			return err
		}
		return tx.Event(Event{ID: uuid.NewString(), Type: "swarm.execution_reconciled", AggregateID: w.ID, Version: w.Version, At: time.Now().UTC(), Payload: map[string]any{"operation_id": l.OperationID, "session_id": l.SessionID, "outcome": fact.Outcome, "failure_code": fact.FailureCode, "fence": w.Fence, "usage": usage, "stop_confirmed": fact.StopConfirmed}})
	})
}

func (s *Service) recordExecutorExit(ctx context.Context, l domain.Lease) error {
	reader, ok := s.repo.(ExecutionHistoryReader)
	if !ok {
		return domain.Failure(apierrors.UnsupportedCapability, "exit observation requires the original execution history")
	}
	history, err := reader.ExecutionHistory(ctx, l.MissionID)
	if err != nil {
		return err
	}
	claimID := ""
	alreadyObserved := false
	for _, execution := range history {
		if execution.OperationID != l.OperationID {
			continue
		}
		if execution.SessionID != l.SessionID || execution.WorkItemID != l.WorkItemID || execution.BindingID != l.BindingID || execution.Fence != l.Fence {
			return domain.Failure(apierrors.ContextStale, "confirmed exit differs from its original claim identity")
		}
		for _, event := range execution.Events {
			if event.Type == "swarm.work_item_claimed" {
				claimID = event.ID
			}
		}
		alreadyObserved = execution.ExitConfirmed && execution.ExitEvent != nil
	}
	if claimID == "" {
		return domain.Failure(apierrors.EvidenceMissing, "confirmed exit lacks its original claim event")
	}
	return s.repo.Within(ctx, func(tx Transaction) error {
		current, err := tx.Lease(l.WorkItemID)
		if err != nil {
			return err
		}
		if current.Fence != l.Fence || current.OperationID != l.OperationID || current.SessionID != l.SessionID || current.MissionID != l.MissionID || current.ProjectID != l.ProjectID || current.BindingID != l.BindingID || current.HolderID != l.HolderID {
			return domain.Failure(apierrors.LeaseLost, "original lease changed before persisting its confirmed exit")
		}
		w, err := tx.WorkItem(l.WorkItemID)
		if err != nil {
			return err
		}
		if w.MissionID != l.MissionID {
			return domain.Failure(apierrors.ContextStale, "original exit observation work item changed its mission")
		}
		if alreadyObserved {
			return nil
		}
		fresh, err := tx.Consume("swarm.execution_exit_observed", claimID)
		if err != nil || !fresh {
			return err
		}
		return tx.Event(Event{ID: uuid.NewString(), Type: "swarm.execution_exit_observed", AggregateID: w.ID, Version: w.Version, At: time.Now().UTC(), Payload: map[string]any{"operation_id": l.OperationID, "session_id": l.SessionID, "stop_confirmed": true, "source": "original_controller"}})
	})
}

// 内部停止检查只处理已有租约，不接收外部主体或派发请求。
func (s *Service) EnforceStopConditions(ctx context.Context) error {
	missions, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	for _, snapshot := range missions {
		if err = ctx.Err(); err != nil {
			return err
		}
		if snapshot.State != "ready" && snapshot.State != "running" || snapshot.StopReason(time.Now().UTC()) == "" {
			continue
		}
		leases := []domain.Lease{}
		err = s.repo.Within(ctx, func(tx Transaction) error {
			m, err := tx.Mission(snapshot.ID)
			if err != nil {
				return err
			}
			if m.Version != snapshot.Version || m.State != "ready" && m.State != "running" {
				return nil
			}
			now := time.Now().UTC()
			reason := m.StopReason(now)
			if reason == "" {
				return nil
			}
			all, err := tx.ActiveLeases(m.Grant.ProjectID)
			if err != nil {
				return err
			}
			old := m.Version
			m.Version++
			m.State, m.Reason = "blocked", reason
			if err = tx.SaveMission(m, old); err != nil {
				return err
			}
			if err = tx.Event(Event{ID: uuid.NewString(), Type: "swarm.mission_stop_requested", AggregateID: m.ID, Version: m.Version, At: now, Payload: map[string]any{"source": "system:swarm-lifecycle", "reason": reason}}); err != nil {
				return err
			}
			for _, lease := range all {
				if lease.MissionID != m.ID {
					continue
				}
				w, err := tx.WorkItem(lease.WorkItemID)
				if err != nil {
					return err
				}
				if w.Fence != lease.Fence || w.Fence == math.MaxInt64 {
					return domain.Failure(apierrors.LeaseLost, "original stop lease fence no longer matches its work item")
				}
				old := w.Version
				w.Version++
				w.Fence++
				w.State, w.Reason = "blocked", reason
				if err = tx.SaveWorkItem(w, old); err != nil {
					return err
				}
				if err = tx.Event(Event{ID: uuid.NewString(), Type: "swarm.execution_stop_requested", AggregateID: w.ID, Version: w.Version, At: now, Payload: map[string]any{"source": "system:swarm-lifecycle", "operation_id": lease.OperationID, "session_id": lease.SessionID, "fence": lease.Fence, "reason": reason}}); err != nil {
					return err
				}
				leases = append(leases, lease)
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, lease := range leases {
			if s.ownership == nil {
				if err = s.blockExecutor(context.WithoutCancel(ctx), lease, "native ownership unavailable; stop intent and reservation retained"); err != nil {
					return err
				}
				continue
			}
			fact, stopErr := s.ownership.Stop(ctx, lease.OperationID, lease.SessionID)
			if stopErr != nil {
				if err = s.blockExecutor(context.WithoutCancel(ctx), lease, "original stop unconfirmed; explicit reconciliation required"); err != nil {
					return err
				}
				var unavailable *ExecutionObservationUnavailable
				if errors.As(stopErr, &unavailable) && (unavailable.Reason == "controller_not_held" || unavailable.Reason == "input_not_accepted") {
					stopErr = &retainedExecutionStop{Cause: stopErr}
				}
			} else {
				stopErr = s.settleExecutor(ctx, lease, fact, false)
			}
			if stopErr != nil && !retainedStopFailure(stopErr) {
				return stopErr
			}
		}
	}
	return nil
}

type retainedExecutionStop struct{ Cause error }

func (e *retainedExecutionStop) Error() string { return e.Cause.Error() }
func (e *retainedExecutionStop) Unwrap() error { return e.Cause }

func retainedStopFailure(err error) bool {
	var retained *retainedExecutionStop
	return errors.As(err, &retained)
}

func unknownUsageWithinKnownBounds(usage domain.Usage, reserved domain.Resources, budget domain.Budget) bool {
	unknown := false
	for _, dimension := range []struct {
		limit    *int64
		used     *int64
		reserved int64
	}{{budget.Calls, usage.Calls, reserved.Calls}, {budget.Tokens, usage.Tokens, reserved.Tokens}} {
		if dimension.limit == nil {
			continue
		}
		if dimension.used == nil {
			unknown = true
		} else if *dimension.used > dimension.reserved {
			return false
		}
	}
	if budget.Money != nil {
		if usage.Money == nil {
			unknown = true
		} else if reserved.Money == nil || usage.Currency != reserved.Currency || *usage.Money > *reserved.Money {
			return false
		}
	}
	return unknown
}

// 恢复只核对原 operation；新认领始终需要显式请求和当前批准。
func (s *Service) RecoverMission(ctx context.Context, p domain.Principal, c TransitionCommand) (domain.Mission, error) {
	var out domain.Mission
	if err := validateMeta(p, c.Meta); err != nil {
		return out, err
	}
	if err := human(p); err != nil {
		return out, err
	}
	if c.Reason == "" {
		return out, domain.Failure(apierrors.ValidationFailed, "recovery reason required")
	}
	var leases []domain.Lease
	replayed := false
	err := s.readWithin(ctx, p, c.MissionID, func(_ context.Context, tx Transaction) error {
		cached, found, _, err := replay[domain.Mission](tx, p, "recover_mission", c.Meta, c)
		if err != nil {
			return err
		}
		if found {
			out = cached
			replayed = true
			return nil
		}
		out, err = tx.Mission(c.MissionID)
		if err != nil {
			return err
		}
		if err = checkVersion(out.Version, c.Meta.ExpectedVersion); err != nil {
			return err
		}
		all, err := tx.ActiveLeases(out.Grant.ProjectID)
		if err != nil {
			return err
		}
		for _, l := range all {
			if l.MissionID == out.ID {
				leases = append(leases, l)
			}
		}
		return nil
	})
	if err != nil || replayed {
		return out, err
	}
	expectedVersion := out.Version
	cancelled := out.State == "cancelled"
	for _, l := range leases {
		if s.ownership == nil {
			return out, domain.Failure(apierrors.UnsupportedCapability, "actual native ownership is not configured")
		}
		fact, err := s.ownership.Reconcile(ctx, l.OperationID, l.SessionID)
		if err != nil {
			if recordErr := s.blockExecutor(context.WithoutCancel(ctx), l, "reconciliation failed; original executor remains reserved"); recordErr != nil {
				return out, recordErr
			}
			return out, err
		}
		if err = s.settleExecutor(ctx, l, fact, cancelled); err != nil {
			return out, err
		}
		expectedVersion++
	}
	if cancelled {
		err = s.readWithin(ctx, p, c.MissionID, func(_ context.Context, tx Transaction) error {
			cached, found, d, err := replay[domain.Mission](tx, p, "recover_mission", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			out, err = tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			if err = checkVersion(out.Version, expectedVersion); err != nil {
				return err
			}
			if out.State != "cancelled" {
				return domain.Failure(apierrors.VersionConflict, "cancelled mission changed during reconciliation")
			}
			return finish(tx, p, "recover_mission", c.Meta, d, out, "", out.ID, out.Version)
		})
		return out, err
	}
	err = s.authorized(ctx, p, c.MissionID, "recover_mission", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		return s.repo.Within(ctx, func(tx Transaction) error {
			cached, found, d, err := replay[domain.Mission](tx, p, "recover_mission", c.Meta, c)
			if err != nil {
				return err
			}
			if found {
				out = cached
				return nil
			}
			out, err = tx.Mission(c.MissionID)
			if err != nil {
				return err
			}
			old := out.Version
			if err = out.ResumeAfterReconciliation(expectedVersion, c.Reason, time.Now().UTC()); err != nil {
				return err
			}
			if err = tx.SaveMission(out, old); err != nil {
				return err
			}
			return finish(tx, p, "recover_mission", c.Meta, d, out, "mission_recovered", out.ID, out.Version)
		})
	})
	return out, err
}
