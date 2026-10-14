package app

import (
	"context"
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
		if current.Fence != l.Fence || current.OperationID != l.OperationID {
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
	if fact.OperationID != l.OperationID || fact.SessionID != l.SessionID || !fact.StopConfirmed || !settledOutcome(fact.Outcome) {
		if err := s.blockExecutor(ctx, l, "original execution or usage remains unknown; reservation and write scope retained"); err != nil {
			return err
		}
		return domain.Failure(apierrors.DeliveryUnknown, "original executor must be reconciled; no replacement can claim its scope")
	}
	usage, err := actualUsage(fact)
	if err != nil {
		return err
	}
	before, err := s.repo.Load(ctx, l.MissionID)
	if err != nil {
		return err
	}
	if !usage.WithinReservation(l.Reserved, before.Grant.Budget) {
		if err = s.blockExecutor(ctx, l, "bounded usage evidence is missing or exceeds reservation; original scope retained"); err != nil {
			return err
		}
		return domain.Failure(apierrors.EvidenceMissing, "bounded actual usage must be reconciled before release")
	}
	return s.repo.Within(ctx, func(tx Transaction) error {
		current, err := tx.Lease(l.WorkItemID)
		if err != nil {
			return err
		}
		if current.Fence != l.Fence || current.OperationID != l.OperationID || current.SessionID != l.SessionID {
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
		return tx.Event(Event{ID: uuid.NewString(), Type: "swarm.execution_reconciled", AggregateID: w.ID, Version: w.Version, At: time.Now().UTC(), Payload: map[string]any{"operation_id": l.OperationID, "session_id": l.SessionID, "outcome": fact.Outcome, "fence": w.Fence, "usage": usage}})
	})
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
		if err = s.settleExecutor(ctx, l, fact, out.State == "cancelled"); err != nil {
			return out, err
		}
	}
	if out.State == "cancelled" {
		return s.repo.Load(ctx, out.ID)
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
			if out.State == "completed" {
				return domain.Failure(apierrors.VersionConflict, "completed mission cannot recover")
			}
			old := out.Version
			out.Version++
			out.State = "ready"
			out.Reason = c.Reason
			if out.StopReached(time.Now().UTC()) {
				out.State = "blocked"
				out.Reason = "mission stop condition reached"
			}
			if err = tx.SaveMission(out, old); err != nil {
				return err
			}
			return finish(tx, p, "recover_mission", c.Meta, d, out, "mission_recovered", out.ID, out.Version)
		})
	})
	return out, err
}
