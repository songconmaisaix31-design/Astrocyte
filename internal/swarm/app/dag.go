package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

type DAGNode struct {
	WorkItem      domain.WorkItem `json:"work_item"`
	CreationEvent *TimelineEvent  `json:"creation_event"`
}

type DAGEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Required bool   `json:"required"`
	EventID  string `json:"event_id"`
}

type WorkerObservation struct {
	State          string     `json:"state"`
	Reason         string     `json:"reason"`
	Source         string     `json:"source"`
	ObservedAt     time.Time  `json:"observed_at"`
	ExitConfirmed  bool       `json:"exit_confirmed"`
	ExitObservedAt *time.Time `json:"exit_observed_at"`
	CurrentInput   *bool      `json:"current_input"`
}

type WorkerParticipationReader interface {
	ReadWorkerParticipation(context.Context, domain.Mission, ExecutionReference) (WorkerObservation, error)
}

type DAGWorker struct {
	Execution   ExecutionReference `json:"execution"`
	LeaseState  string             `json:"lease_state"`
	Observation WorkerObservation  `json:"observation"`
}

type MissionDAG struct {
	SchemaVersion  int             `json:"schema_version"`
	MissionID      string          `json:"mission_id"`
	MissionVersion int             `json:"mission_version"`
	MissionState   string          `json:"mission_state"`
	Reason         string          `json:"reason"`
	ObservedAt     time.Time       `json:"observed_at"`
	Nodes          []DAGNode       `json:"nodes"`
	Edges          []DAGEdge       `json:"edges"`
	Workers        []DAGWorker     `json:"workers"`
	Events         []TimelineEvent `json:"events"`
}

func (s *Service) ReadMissionDAG(ctx context.Context, p domain.Principal, id string) (MissionDAG, error) {
	out := MissionDAG{SchemaVersion: 1, Nodes: []DAGNode{}, Edges: []DAGEdge{}, Workers: []DAGWorker{}, Events: []TimelineEvent{}}
	m, err := s.repo.Load(ctx, id)
	if err != nil {
		return out, err
	}
	authorizer, ok := s.grants.(MissionDAGReadAuthorizer)
	if !ok {
		return out, domain.Failure(apierrors.UnsupportedCapability, "current project read_swarm authorization is unavailable")
	}
	err = authorizer.WithMissionDAGRead(ctx, p, m.Grant.ID, func(ctx context.Context) error {
		detail, err := s.ReadMission(ctx, p, id)
		if err != nil {
			return err
		}
		if err = domain.ValidateDAG(id, detail.WorkItems); err != nil {
			return err
		}
		now := time.Now().UTC()
		out.MissionID, out.MissionVersion, out.MissionState, out.Reason, out.ObservedAt = id, detail.Mission.Version, detail.Mission.State, detail.Mission.Reason, now
		out.Events = detail.Timeline
		creation := map[string]TimelineEvent{}
		for _, event := range detail.Timeline {
			if event.Type == "swarm.work_item_created" {
				if _, exists := creation[event.AggregateID]; exists {
					return domain.Failure(apierrors.EvidenceMissing, "work item has multiple creation events")
				}
				creation[event.AggregateID] = event
			}
		}
		for _, work := range detail.WorkItems {
			node := DAGNode{WorkItem: work}
			if event, exists := creation[work.ID]; exists {
				node.CreationEvent = &event
			}
			out.Nodes = append(out.Nodes, node)
			for _, dependency := range work.Dependencies {
				out.Edges = append(out.Edges, DAGEdge{From: dependency, To: work.ID, Required: true, EventID: creation[work.ID].ID})
			}
		}
		leases := map[string]domain.Lease{}
		workItems := map[string]domain.WorkItem{}
		for _, work := range detail.WorkItems {
			workItems[work.ID] = work
		}
		for _, lease := range detail.Leases {
			leases[lease.OperationID] = lease
		}
		for _, execution := range detail.Executions {
			if _, exists := workItems[execution.WorkItemID]; !exists {
				return domain.Failure(apierrors.EvidenceMissing, "execution history references a missing work item")
			}
			worker := DAGWorker{Execution: execution, LeaseState: "absent", Observation: WorkerObservation{State: "unverifiable", Reason: "native_observer_unavailable", Source: "swarm_history", ObservedAt: now}}
			if lease, exists := leases[execution.OperationID]; exists {
				if lease.SessionID != execution.SessionID || lease.WorkItemID != execution.WorkItemID || lease.Fence != execution.Fence {
					return domain.Failure(apierrors.ContextStale, "current lease differs from its original execution history")
				}
				worker.LeaseState = "current"
				if workItems[lease.WorkItemID].Fence != lease.Fence || detail.Mission.State != "ready" && detail.Mission.State != "running" {
					worker.LeaseState = "retained"
				}
				if !now.Before(lease.ExpiresAt) {
					worker.LeaseState = "expired"
				}
			}
			if execution.ExitConfirmed {
				if execution.ExitEvent == nil || execution.ExitEvent.Type != "swarm.execution_reconciled" && execution.ExitEvent.Type != "swarm.execution_exit_observed" {
					return domain.Failure(apierrors.EvidenceMissing, "confirmed exit lacks its original reconciliation event")
				}
				worker.Observation = WorkerObservation{State: "exited", ExitConfirmed: true, Reason: execution.Reason, Source: execution.ExitEvent.Type, ObservedAt: now}
			} else if reader, ok := s.ownership.(WorkerParticipationReader); ok {
				worker.Observation, err = reader.ReadWorkerParticipation(ctx, detail.Mission, execution)
				if err != nil {
					return err
				}
			}
			out.Workers = append(out.Workers, worker)
		}
		return nil
	})
	return out, err
}
