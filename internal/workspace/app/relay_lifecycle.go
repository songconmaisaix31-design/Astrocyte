package app

import (
	"context"
	"errors"
	"reflect"

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
				if r.TaskID == task.ID && r.Action == "observe" && (r.Status == "delivery_unknown" || r.Status == "rejected") {
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
	if action == "observe" && (st.PendingActions[id] != "" || terminal.ExitedAt != nil) {
		return nil
	}
	spec, err := s.spec(t.Agent)
	if err != nil && action != "observe" {
		return err
	}
	// Containment stop remains possible after grant revocation, as in the
	// existing native Shutdown; it does not grant new tools or send input.
	var p domain.LocalProject
	if err == nil {
		if action == "observe" {
			p, err = s.permission(ctx, domain.Caller{}, t, "observe")
			if err == nil && !sameRelayRoot(w.SourceRoot, p.Root) {
				err = relayError(apierrors.ScopeDenied, "confirmed source root changed")
			}
		} else {
			p, err = s.projects.LoadProject(ctx, t.ProjectID)
		}
	}
	if err != nil && action != "observe" {
		return err
	}
	launch := RelayLaunch{Spec: spec, Project: p, Workspace: w, Terminal: terminal, Session: st.Sessions[terminal.ID]}
	driver := s.structured
	if terminal.DriveMode == "terminal" {
		driver = s.terminal
	}
	if driver == nil && err == nil {
		err = relayError(apierrors.UnsupportedCapability, "owned task driver is not connected")
	}
	if action == "observe" {
		return s.collectOwned(ctx, id, launch, driver, err)
	}
	if err != nil {
		return err
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
	obs, err := driver.Stop(ctx, launch)
	_, err = s.finish(ctx, domain.Caller{}, r, obs, err, false)
	return err
}

var errRelayObservationUnchanged = errors.New("relay observation unchanged")

func (s *relayService) observationChanged(st RelayState, launch RelayLaunch, obs RelayObservation, git *string) bool {
	t := st.Terminals[launch.Terminal.ID]
	if obs.ExitConfirmed && t.ExitedAt == nil {
		return true
	}
	for _, event := range obs.Events {
		if event.Sequence > t.LastSequence && !event.TS.IsZero() && !event.TS.After(s.now()) {
			return true
		}
	}
	if obs.Session.ID != "" && !reflect.DeepEqual(obs.Session, st.Sessions[t.ID]) {
		return true
	}
	if obs.Archive != nil {
		bound := st.OwnedSessions[st.Workspaces[launch.Workspace.ID].ActiveTaskID]
		if bound.Gap != "session_identity_changed" && bound.Gap != "archive_identity_changed" && (bound.SourceStatus != "available" || bound.ArchivePath != obs.Archive.Path || bound.ArchiveRoot != obs.Archive.Root || bound.ArchiveFormat != obs.Archive.Format || !relayArchiveMatches(*obs.Archive, bound) || obs.Archive.Finalized && bound.ArchiveFinalized != "confirmed" || obs.Archive.AppendOnly && bound.ArchiveAppendOnly != "confirmed") {
			return true
		}
	}
	return git != nil && *git != st.Workspaces[launch.Workspace.ID].GitObservation
}

// Read is a cache/observation port. A background tick with no new evidence is
// not a user action and creates neither a receipt nor a state write.
func (s *relayService) collectOwned(ctx context.Context, id string, launch RelayLaunch, driver RelayDriver, captureErr error) error {
	var obs RelayObservation
	var git *string
	if captureErr == nil {
		obs, captureErr = driver.Read(ctx, launch)
	}
	if captureErr == nil && launch.Terminal.DriveMode == "terminal" && s.workspaces != nil {
		var value string
		value, captureErr = s.workspaces.Observe(ctx, launch.Workspace)
		if captureErr == nil {
			git = &value
		}
	}
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	if captureErr == nil && !s.observationChanged(st, launch, obs, git) {
		return nil
	}
	if st.Workspaces[launch.Workspace.ID].ActiveTaskID != id {
		return nil
	}
	operation := uuid.NewString()
	r := domain.RelayReceipt{ID: "service:observer:observe:" + operation, OperationID: operation, TaskID: id, Action: "observe", Operator: "service:observer", Status: "pending", Input: []byte("null"), CreatedAt: s.now()}
	err = s.store.Update(context.WithoutCancel(ctx), func(current *RelayState) error {
		t := current.Tasks[id]
		if t.TerminalID != launch.Terminal.ID || current.Workspaces[launch.Workspace.ID].ActiveTaskID != id || current.Terminals[t.TerminalID].ExitedAt != nil || current.PendingActions[id] != "" || !reflect.DeepEqual(current.Sessions[t.TerminalID], launch.Session) {
			return errRelayObservationUnchanged
		}
		for _, prior := range current.Receipts {
			if prior.TaskID == id && prior.Action == "observe" && (prior.Status == "rejected" || prior.Status == "delivery_unknown") {
				return errRelayObservationUnchanged
			}
		}
		if captureErr == nil && !s.observationChanged(*current, launch, obs, git) {
			return errRelayObservationUnchanged
		}
		current.Receipts[r.ID] = r
		if git != nil {
			s.recordGit(current, id, launch, *git)
		}
		var finalErr error
		r, finalErr = s.finalize(current, r, obs, captureErr, false)
		return finalErr
	})
	if errors.Is(err, errRelayObservationUnchanged) {
		return nil
	}
	if err != nil {
		return err
	}
	_ = s.publish(context.WithoutCancel(ctx), id)
	return receiptError(r)
}

func (s *relayService) observeGit(ctx context.Context, id string, launch RelayLaunch) error {
	git, err := s.workspaces.Observe(ctx, launch.Workspace)
	if err != nil {
		return err
	}
	if git == launch.Workspace.GitObservation {
		return nil
	}
	return s.store.Update(ctx, func(st *RelayState) error {
		s.recordGit(st, id, launch, git)
		return nil
	})
}

func (s *relayService) recordGit(st *RelayState, id string, launch RelayLaunch, git string) {
	workspace := st.Workspaces[launch.Workspace.ID]
	if workspace.GitObservation == git {
		return
	}
	e := domain.RelayEvent{ID: uuid.NewString(), TaskID: id, TerminalID: launch.Terminal.ID, WorktreeID: launch.Workspace.ID, Agent: launch.Spec.Name, Kind: "git_activity", Source: "git", Topic: "relay:" + id, Scope: launch.Project.ID, TS: s.now()}
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
