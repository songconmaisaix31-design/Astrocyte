package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayLifecycle = (*relayService)(nil)

// ObserveOwned is a bounded pass over already accepted tasks. The entrypoint
// owns invocation/lifetime. This is observation, with no task creation, resume,
// automatic dispatch, or failover. An unknown owner is not probed repeatedly.
func (s *relayService) ObserveOwned(ctx context.Context) error {
	return s.ownedPass(ctx, "observe")
}

// Shutdown contains only processes belonging to this Service's existing task
// records. It never releases an unfinished workspace or treats stop as done.
func (s *relayService) Shutdown(ctx context.Context) error { return s.ownedPass(ctx, "stop") }

func (s *relayService) ownedPass(ctx context.Context, action string) error {
	if s.store == nil {
		return nil
	}
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	var failures []error
	for _, task := range st.Tasks {
		if err = ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		terminal, exists := st.Terminals[task.TerminalID]
		if !exists || terminal.ExitedAt != nil || task.State == domain.RelayReleased {
			continue
		}
		if action == "observe" {
			if task.State != domain.RelayAccepted && task.State != domain.RelayWorking && task.State != domain.RelayDone {
				continue
			}
			uncertain := false
			for _, r := range st.Receipts {
				if r.TaskID == task.ID && r.Action == "observe" && r.Status == "delivery_unknown" {
					uncertain = true
					break
				}
			}
			if uncertain {
				continue
			}
		}
		if err = s.ownedAction(ctx, task.ID, action); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *relayService) ownedAction(ctx context.Context, id, action string) error {
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	t, ok := st.Tasks[id]
	if !ok {
		return apierrors.NewNotFound("relay task", id)
	}
	w, ok := st.Workspaces[t.WorktreeID]
	if !ok || w.ActiveTaskID != id {
		return relayError(apierrors.VersionConflict, "owned task no longer reserves its workspace")
	}
	terminal := st.Terminals[t.TerminalID]
	spec, err := s.spec(t.Agent)
	if err != nil {
		return err
	}
	// Containment stop remains possible after grant revocation, as in the
	// existing native Shutdown; it does not grant new tools or send input.
	var p domain.LocalProject
	if action == "observe" {
		p, err = s.permission(ctx, domain.Caller{}, t, "observe")
	} else {
		p, err = s.projects.LoadProject(ctx, t.ProjectID)
	}
	if err != nil {
		return err
	}
	launch := RelayLaunch{Spec: spec, Project: p, Workspace: w, Terminal: terminal, Session: st.Sessions[terminal.ID]}
	driver := s.structured
	if terminal.DriveMode == "terminal" {
		driver = s.terminal
	}
	if driver == nil {
		return relayError(apierrors.UnsupportedCapability, "owned task driver is not connected")
	}
	operation := uuid.NewString()
	operator := "service:observer"
	if action == "stop" {
		operator = "service:shutdown"
	}
	r := domain.RelayReceipt{ID: operator + ":" + action + ":" + operation, OperationID: operation, TaskID: id, Action: action, Operator: operator, Status: "pending", Input: []byte("null"), CreatedAt: s.now()}
	err = s.store.Update(ctx, func(current *RelayState) error {
		if current.PendingActions[id] != "" {
			return relayError(apierrors.VersionConflict, "owned task has an action in progress")
		}
		current.PendingActions[id] = r.ID
		current.Receipts[r.ID] = r
		return nil
	})
	if err != nil {
		return err
	}
	var obs RelayObservation
	if action == "stop" {
		obs, err = driver.Stop(ctx, launch)
	} else {
		obs, err = driver.Read(ctx, launch)
	}
	if action == "observe" && err == nil && terminal.DriveMode == "terminal" && s.workspaces != nil {
		err = s.observeGit(ctx, id, launch)
	}
	_, err = s.finish(ctx, domain.Caller{}, r, obs, err, false)
	return err
}

func (s *relayService) observeGit(ctx context.Context, id string, launch RelayLaunch) error {
	git, err := s.workspaces.Observe(ctx, launch.Workspace)
	if err != nil {
		return err
	}
	if git == launch.Workspace.GitObservation {
		return nil
	}
	e := domain.RelayEvent{ID: uuid.NewString(), TaskID: id, TerminalID: launch.Terminal.ID, WorktreeID: launch.Workspace.ID, Agent: launch.Spec.Name, Kind: "git_activity", Source: "git", Topic: "relay:" + id, Scope: launch.Project.ID, TS: s.now()}
	return s.store.Update(ctx, func(st *RelayState) error {
		workspace := st.Workspaces[launch.Workspace.ID]
		if workspace.GitObservation == git {
			return nil
		}
		workspace.GitObservation = git
		st.Workspaces[workspace.ID] = workspace
		st.Events = append(st.Events, e)
		terminal := st.Terminals[launch.Terminal.ID]
		when := e.TS
		if terminal.LastEventAt == nil || when.After(*terminal.LastEventAt) {
			terminal.LastEventAt = &when
		}
		terminal.Liveness = domain.RelayLiveness(terminal.LastEventAt, terminal.ExitedAt, s.now(), s.options.LivenessWindow)
		st.Terminals[terminal.ID] = terminal
		t := st.Tasks[id]
		if t.State == domain.RelayAccepted {
			s.move(st, &t, domain.RelayWorking, "observation:git")
		}
		return nil
	})
}

// Reconciliation is explicit human work on the same stable event origins. It
// never replays a launch/send, and original error receipts remain immutable.
func (s *relayService) ReconcileEvents(ctx context.Context, c domain.Caller, id string) (domain.RelayTaskView, error) {
	r, fresh, err := s.begin(ctx, c, id, "reconcile_events", nil)
	if err != nil {
		return domain.RelayTaskView{}, err
	}
	if !fresh {
		view, e := s.GetTask(ctx, c, id)
		if e != nil {
			return view, e
		}
		return view, receiptError(r)
	}
	view, err := s.GetTask(ctx, c, id)
	if err == nil {
		_, err = s.permission(ctx, c, view.Task, "observe")
	}
	if err == nil {
		err = s.store.Update(ctx, func(st *RelayState) error { delete(st.PublicationErrors, id); return nil })
	}
	if err == nil {
		err = s.publish(ctx, id)
	}
	result, finishErr := s.finish(ctx, c, r, RelayObservation{}, err, true)
	return result.View, finishErr
}
