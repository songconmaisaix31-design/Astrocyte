package app

import (
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (e *DynamicsEngine) Conflicts(ctx context.Context) ([]DynamicsConflict, error) {
	var conflicts []DynamicsConflict
	err := e.store.ViewDynamics(ctx, func(tx DynamicsTx) error { var err error; conflicts, err = tx.ListConflicts(); return err })
	return conflicts, err
}

func (e *DynamicsEngine) ReportFactConflict(ctx context.Context, topic, leftEvent, rightEvent, detail string, now int64) (DynamicsConflict, error) {
	var c DynamicsConflict
	err := e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		var err error
		c, err = e.ReportFactConflictInTx(tx, topic, leftEvent, rightEvent, detail, now)
		return err
	})
	return c, err
}

func (e *DynamicsEngine) ReportFactConflictInTx(tx DynamicsTx, topic, leftEvent, rightEvent, detail string, now int64) (DynamicsConflict, error) {
	var c DynamicsConflict
	if _, err := tx.GetNode(topic); err != nil {
		return c, err
	}
	for _, id := range []string{leftEvent, rightEvent} {
		r, err := tx.GetEvent(id)
		if err != nil {
			return c, err
		}
		if r.Event == nil {
			return c, dynamicsInvalid("Conflict requires mapped evidence")
		}
	}
	list, err := tx.ListConflicts()
	if err != nil {
		return c, err
	}
	for _, old := range list {
		if old.Conflict.Topic == topic && old.Conflict.LeftRef == leftEvent && old.Conflict.RightRef == rightEvent && old.Conflict.Detail == detail {
			c = old
			return c, nil
		}
	}
	c = DynamicsConflict{Conflict: domain.Conflict{ID: dynamicsID(), Topic: topic, Kind: domain.ConflictKindKnowledge, LeftRef: leftEvent, RightRef: rightEvent, Detail: detail, Status: domain.ConflictStatusPending}, CreatedAt: now, Version: 1}
	return c, tx.SaveConflict(c)
}

func (e *DynamicsEngine) AdjudicateConflict(ctx context.Context, p Principal, id string, cmd AdjudicateConflictCommand) (ConflictResult, error) {
	if err := authorize(p, true); err != nil {
		return ConflictResult{}, err
	}
	if err := validateMeta(cmd.CommandMeta); err != nil {
		return ConflictResult{}, err
	}
	statuses := map[string]string{ConflictDecisionKeepLeft: domain.ConflictStatusKeepLeft, ConflictDecisionKeepRight: domain.ConflictStatusKeepRight, ConflictDecisionMerge: domain.ConflictStatusMerge, ConflictDecisionDismiss: domain.ConflictStatusDismissed}
	status, ok := statuses[cmd.Decision]
	if !ok {
		return ConflictResult{}, dynamicsInvalid("Unknown conflict decision")
	}
	var result ConflictResult
	digest, err := semanticDigest(cmd)
	if err != nil {
		return result, err
	}
	err = e.store.WithDynamicsTx(ctx, func(tx DynamicsTx) error {
		actor, operation := p.Kind+":"+p.ID, "engine.adjudicate_conflict:"+id
		receipt, err := tx.LoadDecisionReceipt(actor, operation, cmd.IdempotencyKey)
		if err == nil {
			if receipt.Digest != digest {
				return serviceError(apierrors.VersionConflict, "Conflict idempotency key binds a different decision", "use_new_idempotency_key")
			}
			if json.Unmarshal(receipt.Result, &result) != nil || result.SchemaVersion != 1 || result.Conflict.ID != id || result.Version < 1 {
				return serviceError(apierrors.DeliveryUnknown, "Conflict receipt outcome is unknown", "reconcile_original_operation")
			}
			return nil
		}
		if !dynamicsNotFound(err) {
			return err
		}
		c, err := tx.GetConflict(id)
		if err != nil {
			return err
		}
		if c.Version != cmd.ExpectedVersion {
			return serviceError(apierrors.VersionConflict, "Conflict version changed", "reload_and_retry")
		}
		if c.Conflict.Status != domain.ConflictStatusPending {
			return dynamicsInvalid("Conflict is already adjudicated")
		}
		c.Conflict.Status = status
		c.Version++
		result = ConflictResult{SchemaVersion: 1, Conflict: c.Conflict, Version: c.Version}
		if err = tx.SaveConflict(c); err != nil {
			return err
		}
		a := ActionAudit{Action: cmd.Decision, ActorID: p.ID, ActorKind: p.Kind, TS: e.options.Clock().UnixMilli(), RequestID: cmd.RequestID, Version: c.Version}
		if err = tx.AddConflictAction(id, a); err != nil {
			return err
		}
		result = ConflictResult{SchemaVersion: 1, Conflict: c.Conflict, Version: c.Version}
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return tx.SaveDecisionReceipt(actor, operation, cmd.IdempotencyKey, Receipt{Digest: digest, Result: data})
	})
	return result, err
}
