package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func (r *SwarmRepository) ReadAcceptedArtifactEvent(ctx context.Context, id string) (swarm.AcceptedArtifactEvent, error) {
	var out swarm.AcceptedArtifactEvent
	var at, payload string
	err := r.db.conn.QueryRowContext(ctx, "SELECT occurred_at,payload FROM foundation_outbox WHERE event_id=? AND event_type='swarm.artifact_accepted' AND schema_version=1", id).Scan(&at, &payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, apierrors.NewNotFound("accepted artifact event", id)
		}
		return out, err
	}
	out.EventID = id
	out.OccurredAt, err = time.Parse(time.RFC3339Nano, at)
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal([]byte(payload), &out.Artifact); err != nil {
		return out, err
	}
	a := out.Artifact
	if a.ID == "" || a.State != "accepted" || a.AcceptanceID == "" || a.AcceptedBy == "" || a.AcceptedActorKind != "human" || a.Commit == "" || a.ObjectRef == "" || a.ProjectChecksRef == "" {
		return out, domain.Failure(apierrors.EvidenceMissing, "accepted artifact event lacks immutable human adoption provenance")
	}
	out.ActorID = a.AcceptedBy
	out.ActorKind = a.AcceptedActorKind
	err = r.db.conn.QueryRowContext(ctx, "SELECT w.mission_id,m.project_id FROM swarm_work_items w JOIN swarm_missions m ON m.id=w.mission_id WHERE w.id=?", a.WorkItemID).Scan(&out.MissionID, &out.ProjectID)
	return out, err
}

// 从既有 outbox 的真实 operation 事实构造只读视图，不创建新的执行身份。
func (r *SwarmRepository) ExecutionHistory(ctx context.Context, id string) ([]swarm.ExecutionReference, error) {
	var executor SwarmSQLExecutor = r.db.conn
	if tx, ok := ctx.Value(swarmTransactionKey{}).(*swarmTransaction); ok {
		if tx.db != r.db {
			return nil, domain.Failure(apierrors.ScopeDenied, "execution history transaction belongs to another database")
		}
		executor = tx.exec
	}
	rows, err := executor.QueryContext(ctx, "SELECT event_id,event_type,aggregate_id,aggregate_version,occurred_at,payload FROM foundation_outbox WHERE event_type IN ('swarm.work_item_claimed','swarm.artifact_submitted','swarm.execution_blocked','swarm.execution_reconciled','swarm.artifact_revision_requested','swarm.artifact_accepted') AND (aggregate_id IN (SELECT id FROM swarm_work_items WHERE mission_id=?) OR aggregate_id IN (SELECT a.id FROM swarm_artifacts a JOIN swarm_work_items w ON w.id=a.work_item_id WHERE w.mission_id=?)) ORDER BY occurred_at,event_id", id, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []swarm.ExecutionReference{}
	byOperation := map[string]int{}
	for rows.Next() {
		var e swarm.TimelineEvent
		var at, payload string
		if err = rows.Scan(&e.ID, &e.Type, &e.AggregateID, &e.Version, &at, &payload); err != nil {
			return nil, err
		}
		e.OccurredAt, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, err
		}
		if e.Type == "swarm.work_item_claimed" {
			var l domain.Lease
			if err = json.Unmarshal([]byte(payload), &l); err != nil {
				return nil, err
			}
			if l.OperationID == "" {
				return nil, domain.Failure(apierrors.EvidenceMissing, "claim event lacks its actual operation reference")
			}
			if _, exists := byOperation[l.OperationID]; exists {
				return nil, domain.Failure(apierrors.VersionConflict, "one operation has multiple claim facts")
			}
			byOperation[l.OperationID] = len(out)
			out = append(out, swarm.ExecutionReference{OperationID: l.OperationID, SessionID: l.SessionID, WorkItemID: l.WorkItemID, BindingID: l.BindingID, Fence: l.Fence, ContextPacketID: l.ContextPacketID, PolicyVersion: l.PolicyVersion, Events: []swarm.TimelineEvent{e}, ToolEventsStatus: "unsupported", ToolEvents: []swarm.ToolEventReference{}})
			continue
		}
		var ref struct {
			OperationID string `json:"operation_id"`
			SessionID   string `json:"session_id"`
		}
		if err = json.Unmarshal([]byte(payload), &ref); err != nil {
			return nil, err
		}
		index, exists := byOperation[ref.OperationID]
		if !exists {
			return nil, domain.Failure(apierrors.EvidenceMissing, "execution event lacks its original claim fact")
		}
		if out[index].SessionID != ref.SessionID {
			return nil, domain.Failure(apierrors.ContextStale, "execution event changed the original session identity")
		}
		out[index].Events = append(out[index].Events, e)
	}
	return out, rows.Err()
}
