package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func (s *Service) readWithin(ctx context.Context, p domain.Principal, id string, fn func(context.Context, Transaction) error) error {
	m, err := s.repo.Load(ctx, id)
	if err != nil {
		return err
	}
	read, ok := s.grants.(MissionReadAuthorizer)
	if !ok {
		return domain.Failure(apierrors.UnsupportedCapability, "actual current project read authorization is unavailable")
	}
	return read.WithMissionRead(ctx, p, m.Grant.ID, func(ctx context.Context) error {
		return s.repo.Within(ctx, func(tx Transaction) error { return fn(ctx, tx) })
	})
}

func (s *Service) ListMissionsFor(ctx context.Context, p domain.Principal) ([]domain.Mission, error) {
	if p.ID == "" || (p.Kind != "human" && p.Kind != "agent") {
		return nil, domain.Failure(apierrors.ScopeDenied, "authenticated project identity required")
	}
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := []domain.Mission{}
	for _, m := range rows {
		if p.Kind == "agent" && p.ProjectID != m.Grant.ProjectID {
			continue
		}
		if err = s.readWithin(ctx, p, m.ID, func(_ context.Context, tx Transaction) error {
			current, err := tx.Mission(m.ID)
			if err != nil {
				return err
			}
			out = append(out, current)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Service) ReadMission(ctx context.Context, p domain.Principal, id string) (MissionDetail, error) {
	out := MissionDetail{WorkItems: []domain.WorkItem{}, Leases: []domain.Lease{}, Artifacts: []domain.Artifact{}, Timeline: []TimelineEvent{}, Executions: []ExecutionReference{}}
	err := s.readWithin(ctx, p, id, func(ctx context.Context, tx Transaction) error {
		var err error
		out.Mission, err = tx.Mission(id)
		if err != nil {
			return err
		}
		out.WorkItems, err = tx.WorkItems(id)
		if err != nil {
			return err
		}
		leases, err := tx.ActiveLeases(out.Mission.Grant.ProjectID)
		if err != nil {
			return err
		}
		for _, l := range leases {
			if l.MissionID == id {
				out.Leases = append(out.Leases, l)
			}
		}
		for _, w := range out.WorkItems {
			if w.ArtifactID != "" {
				a, err := tx.Artifact(w.ArtifactID)
				if err != nil {
					return err
				}
				out.Artifacts = append(out.Artifacts, a)
			}
		}
		history, ok := s.repo.(ExecutionHistoryReader)
		if !ok {
			return domain.Failure(apierrors.UnsupportedCapability, "actual operation history query is unavailable")
		}
		out.Executions, err = history.ExecutionHistory(ctx, id)
		if err != nil {
			return err
		}
		if reader, ok := s.repo.(TimelineReader); ok {
			out.Timeline, err = reader.Timeline(ctx, id)
			return err
		}
		return domain.Failure(apierrors.UnsupportedCapability, "durable mission timeline query is unavailable")
	})
	if err != nil {
		return out, err
	}
	if reader, ok := s.ownership.(ToolTraceReader); ok {
		for i := range out.Executions {
			e := &out.Executions[i]
			e.ToolEventsStatus, e.ToolEvents, err = reader.ReadToolEventRefs(ctx, p, e.OperationID, e.SessionID)
			if err != nil {
				return out, err
			}
		}
	}
	return out, err
}

func (s *Service) ReadArtifact(ctx context.Context, p domain.Principal, missionID, id string) (domain.Artifact, error) {
	var out domain.Artifact
	err := s.readWithin(ctx, p, missionID, func(_ context.Context, tx Transaction) error {
		var err error
		out, err = tx.Artifact(id)
		if err != nil {
			return err
		}
		w, err := tx.WorkItem(out.WorkItemID)
		if err != nil {
			return err
		}
		if w.MissionID != missionID {
			return domain.Failure(apierrors.ScopeDenied, "artifact belongs to another mission")
		}
		return nil
	})
	return out, err
}

func (s *Service) SelectWorkItems(ctx context.Context, p domain.Principal, missionID string, member Member, inputs []domain.SourceVersion) ([]domain.Candidate, error) {
	out := []domain.Candidate{}
	err := s.authorized(ctx, p, missionID, "read_swarm", domain.Resources{}, func(ctx context.Context, g domain.Grant) error {
		descriptor, err := s.execution(ctx, p, member)
		if err != nil {
			return err
		}
		return s.repo.Within(ctx, func(tx Transaction) error {
			m, err := tx.Mission(missionID)
			if err != nil {
				return err
			}
			items, err := tx.WorkItems(missionID)
			if err != nil {
				return err
			}
			leases, err := tx.ActiveLeases(g.ProjectID)
			if err != nil {
				return err
			}
			out, err = domain.SelectWorkItems(m, items, leases, domain.SelectionContext{Inputs: inputs, Capabilities: descriptor.Capabilities}, time.Now().UTC())
			return err
		})
	})
	return out, err
}
