package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayService struct {
	store                RelayStore
	works                Works
	projects             LocalProjectRepository
	workspaces           RelayWorkspaces
	structured, terminal RelayDriver
	events               RelayEventSink
	options              RelayServiceOptions
}

var _ RelayRuntime = (*relayService)(nil)

func NewRelayService(store RelayStore, works Works, projects LocalProjectRepository, workspaces RelayWorkspaces, structured, terminal RelayDriver, events RelayEventSink, options RelayServiceOptions) RelayRuntime {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	// Copy registration data so external mutation cannot change executable policy.
	data, _ := json.Marshal(options.Agents)
	_ = json.Unmarshal(data, &options.Agents)
	return &relayService{store, works, projects, workspaces, structured, terminal, events, options}
}

func relayError(code apierrors.Code, message string) *apierrors.ServiceError {
	action := "review_relay_permissions_and_state"
	if code == apierrors.DeliveryUnknown {
		action = "reload_relay_without_replay"
	}
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: action}
}

func relayOperation(c domain.Caller) error {
	if err := human(c); err != nil {
		return err
	}
	if _, err := uuid.Parse(c.OperationID); err != nil {
		return relayError(apierrors.ValidationFailed, "a UUID operation identity is required")
	}
	return nil
}

func (s *relayService) now() time.Time  { return s.options.Clock().UTC() }
func relayActor(c domain.Caller) string { return c.Kind + ":" + c.ID }

func (s *relayService) spec(name string) (domain.AgentSpec, error) {
	for _, v := range s.options.Agents {
		if v.Name == name && (v.DriveMode == "structured" || v.DriveMode == "terminal" || v.DriveMode == "both") {
			return v, nil
		}
	}
	return domain.AgentSpec{}, relayError(apierrors.UnsupportedCapability, "selected Agent has no valid relay registration")
}

func (s *relayService) ListAgents(ctx context.Context, c domain.Caller) ([]domain.AgentSpec, error) {
	if err := human(c); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	items := []domain.AgentSpec{}
	data, _ := json.Marshal(s.options.Agents)
	_ = json.Unmarshal(data, &items)
	return items, nil
}

func (s *relayService) snapshot(ctx context.Context, c domain.Caller) (RelayState, error) {
	if err := human(c); err != nil {
		return RelayState{}, err
	}
	if s.store == nil {
		return RelayState{}, relayError(apierrors.UnsupportedCapability, "relay store is not connected")
	}
	return s.store.Snapshot(ctx)
}

func (s *relayService) view(st RelayState, id string) (domain.RelayTaskView, error) {
	t, ok := st.Tasks[id]
	if !ok {
		return domain.RelayTaskView{}, apierrors.NewNotFound("relay task", id)
	}
	v := domain.RelayTaskView{Task: t, Transitions: []domain.RelayTransition{}}
	v.EventPublication.Status = "none"
	for _, e := range st.Events {
		if e.TaskID == id {
			if st.Published[e.ID] {
				v.EventPublication.PublishedEvents++
			} else {
				v.EventPublication.PendingEvents++
			}
		}
	}
	if v.EventPublication.PublishedEvents > 0 {
		v.EventPublication.Status = "published"
	}
	if v.EventPublication.PendingEvents > 0 {
		v.EventPublication.Status = "pending"
	}
	if st.PublicationErrors[id] {
		v.EventPublication.Status = "error"
	}
	if terminal, ok := st.Terminals[t.TerminalID]; ok {
		terminal.Liveness = domain.RelayLiveness(terminal.LastEventAt, terminal.ExitedAt, s.now(), s.options.LivenessWindow)
		v.Terminal = &terminal
		if terminal.ExitedAt != nil && t.State != domain.RelayDone && t.State != domain.RelayReleased {
			v.Task.Condition = "terminal_exited"
		}
	}
	if workspace, ok := st.Workspaces[t.WorktreeID]; ok {
		v.Workspace = &workspace
	}
	for _, edge := range st.Transitions {
		if edge.TaskID == id {
			v.Transitions = append(v.Transitions, edge)
		}
	}
	return v, nil
}

func (s *relayService) GetTask(ctx context.Context, c domain.Caller, id string) (domain.RelayTaskView, error) {
	st, err := s.snapshot(ctx, c)
	if err != nil {
		return domain.RelayTaskView{}, err
	}
	return s.view(st, id)
}

func (s *relayService) ListTasks(ctx context.Context, c domain.Caller) ([]domain.RelayTaskView, error) {
	st, err := s.snapshot(ctx, c)
	if err != nil {
		return nil, err
	}
	items := []domain.RelayTaskView{}
	for id := range st.Tasks {
		v, e := s.view(st, id)
		if e != nil {
			return nil, e
		}
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Task.CreatedAt.Equal(items[j].Task.CreatedAt) {
			return items[i].Task.ID < items[j].Task.ID
		}
		return items[i].Task.CreatedAt.After(items[j].Task.CreatedAt)
	})
	return items, nil
}

func (s *relayService) ListEvents(ctx context.Context, c domain.Caller, id string) ([]domain.RelayEvent, error) {
	st, err := s.snapshot(ctx, c)
	if err != nil {
		return nil, err
	}
	if id != "" {
		if _, err = s.view(st, id); err != nil {
			return nil, err
		}
	}
	items := []domain.RelayEvent{}
	for _, v := range st.Events {
		if id == "" || v.TaskID == id {
			items = append(items, v)
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].TS.After(items[j].TS) })
	return items, nil
}

func (s *relayService) ListReceipts(ctx context.Context, c domain.Caller, id string) ([]domain.RelayReceipt, error) {
	st, err := s.snapshot(ctx, c)
	if err != nil {
		return nil, err
	}
	if id != "" {
		if _, err = s.view(st, id); err != nil {
			return nil, err
		}
	}
	items := []domain.RelayReceipt{}
	for _, v := range st.Receipts {
		if id == "" || v.TaskID == id {
			v.Input = nil
			items = append(items, v)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID < items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (s *relayService) move(st *RelayState, task *domain.RelayTask, to, actor string) {
	st.Transitions = append(st.Transitions, domain.RelayTransition{TaskID: task.ID, From: task.State, To: to, TS: s.now(), Operator: actor})
	task.State = to
	task.UpdatedAt = s.now()
	st.Tasks[task.ID] = *task
}

func sameRelayRoot(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func (s *relayService) work(ctx context.Context, c domain.Caller, workID, projectID string) (domain.Work, domain.LocalProject, error) {
	if s.works == nil || s.projects == nil {
		return domain.Work{}, domain.LocalProject{}, relayError(apierrors.UnsupportedCapability, "relay work/project source is not connected")
	}
	p, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return domain.Work{}, p, err
	}
	album, err := s.works.ListWorks(ctx, c)
	if err != nil {
		return domain.Work{}, p, err
	}
	for _, w := range album.Works {
		if w.ID == workID {
			if !sameRelayRoot(w.Root, p.Root) {
				return w, p, relayError(apierrors.ScopeDenied, "work does not match the approved project root")
			}
			return w, p, nil
		}
	}
	return domain.Work{}, p, apierrors.NewNotFound("work", workID)
}

// The A/B/C labels summarize existing, independent scope switches. A confirm
// must acknowledge the current scope, not silently reduce C or directory B.
func RelayPermissionLevel(settings domain.ProjectSettings) string {
	if settings.ExpandReferences {
		return "C"
	}
	if settings.AllowDirectory {
		return "B"
	}
	return "A"
}

func (s *relayService) permission(ctx context.Context, c domain.Caller, t domain.RelayTask, action string) (domain.LocalProject, error) {
	p, err := s.projects.LoadProject(ctx, t.ProjectID)
	if err != nil {
		return p, err
	}
	if p.Settings.Revision != t.SettingsRevision || RelayPermissionLevel(p.Settings) != t.PermissionLevel {
		return p, relayError(apierrors.ApprovalRevoked, "project permission changed; review a new intent")
	}
	if p.Settings.ExternalModelCLI != t.Agent || !has(p.Settings.AllowedActions, action) {
		return p, relayError(apierrors.ApprovalRequired, "project permission does not cover the Agent and action")
	}
	// Human controls keep the existing native human policy. Agent tokens cannot
	// invoke this human-only surface; grants are reloaded if the project already
	// has an explicit grant for the selected Agent.
	g, e := s.projects.LoadGrant(ctx, t.ProjectID, t.Agent)
	if e == nil {
		if g.RevokedAt != nil || (g.ExpiresAt != nil && !g.ExpiresAt.After(s.now())) {
			return p, relayError(apierrors.ApprovalRevoked, "selected Agent grant is revoked or expired")
		}
		if !has(g.Actions, action) {
			return p, relayError(apierrors.ScopeDenied, "selected Agent grant excludes the action")
		}
	} else {
		var known *apierrors.ServiceError
		if !errors.As(e, &known) || known.Code != apierrors.NotFound {
			return p, e
		}
	}
	return p, nil
}

// An exact action+operator+operation key is durable; final RED and unknown are
// replayed as evidence rather than run again. No process effect occurs here.
func (s *relayService) begin(ctx context.Context, c domain.Caller, id, action string, input any) (domain.RelayReceipt, bool, error) {
	var r domain.RelayReceipt
	if err := relayOperation(c); err != nil {
		return r, false, err
	}
	if s.store == nil {
		return r, false, relayError(apierrors.UnsupportedCapability, "relay store is not connected")
	}
	data, err := json.Marshal(struct {
		TaskID string
		Input  any
	}{id, input})
	if err != nil {
		return r, false, err
	}
	key := relayActor(c) + ":" + action + ":" + c.OperationID
	fresh := false
	err = s.store.Update(ctx, func(st *RelayState) error {
		if old, ok := st.Receipts[key]; ok {
			r = old
			if !bytes.Equal(old.Input, data) {
				return relayError(apierrors.VersionConflict, "operation identity has different input")
			}
			return nil
		}
		r = domain.RelayReceipt{ID: key, OperationID: c.OperationID, TaskID: id, Action: action, Operator: relayActor(c), Status: "pending", Input: data, CreatedAt: s.now()}
		st.Receipts[key] = r
		fresh = true
		return nil
	})
	return r, fresh, err
}

func receiptError(r domain.RelayReceipt) error {
	if r.Status == "accepted" {
		return nil
	}
	if r.Error != nil {
		return &apierrors.ServiceError{Code: apierrors.Code(r.Error.Code), Message: r.Error.Message, RequiredAction: r.Error.RequiredAction}
	}
	return relayError(apierrors.DeliveryUnknown, "action is pending or has unknown delivery; do not replay")
}

func (s *relayService) result(ctx context.Context, _ domain.Caller, r domain.RelayReceipt) (domain.RelayActionResult, error) {
	if r.TaskID == "" {
		r.Input = nil
		return domain.RelayActionResult{Receipt: r}, nil
	}
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return domain.RelayActionResult{}, err
	}
	v, err := s.view(st, r.TaskID)
	r.Input = nil
	return domain.RelayActionResult{View: v, Receipt: r}, err
}

func (s *relayService) finish(ctx context.Context, c domain.Caller, r domain.RelayReceipt, obs RelayObservation, effectErr error, knownUnstarted bool) (domain.RelayActionResult, error) {
	err := s.store.Update(context.WithoutCancel(ctx), func(st *RelayState) error {
		current, ok := st.Receipts[r.ID]
		if !ok || current.Status != "pending" {
			return relayError(apierrors.VersionConflict, "receipt already finalized")
		}
		r = current
		now := s.now()
		r.FinishedAt = &now
		r.Status = "accepted"
		if effectErr != nil {
			var known *apierrors.ServiceError
			if !errors.As(effectErr, &known) {
				known = relayError(apierrors.DeliveryUnknown, "action delivery could not be verified")
			}
			r.Status = "rejected"
			if known.Code == apierrors.DeliveryUnknown || (!knownUnstarted && r.Action == "dispatch") {
				r.Status = "delivery_unknown"
				known = relayError(apierrors.DeliveryUnknown, "launch delivery is unknown; no second writer may be started")
			}
			r.Error = &domain.RelayReceiptError{Code: string(known.Code), Message: known.Message, RequiredAction: known.RequiredAction}
		}
		t, exists := st.Tasks[r.TaskID]
		if exists {
			if r.Status == "delivery_unknown" {
				t.Condition = "delivery_unknown"
				st.Tasks[t.ID] = t
			}
			if knownUnstarted && r.Status != "delivery_unknown" && r.Action == "dispatch" && st.PendingActions[t.ID] == r.ID {
				w := st.Workspaces[t.WorktreeID]
				if w.ActiveTaskID == t.ID {
					w.ActiveTaskID = ""
					st.Workspaces[w.ID] = w
				}
			}
			if t.TerminalID != "" {
				s.observe(st, &t, obs)
			}
		}
		st.Receipts[r.ID] = r
		if st.PendingActions[r.TaskID] == r.ID {
			delete(st.PendingActions, r.TaskID)
		}
		return nil
	})
	if err != nil {
		return domain.RelayActionResult{}, relayError(apierrors.DeliveryUnknown, "effect receipt could not be finalized; reload without replay")
	}
	_ = s.publish(context.WithoutCancel(ctx), r.TaskID)
	result, err := s.result(context.WithoutCancel(ctx), c, r)
	if err != nil {
		return result, err
	}
	// Events remain durable and pending if the sink is not connected or rejects
	// them. Never label event publication as task completion.
	return result, receiptError(r)
}

func (s *relayService) CreateIntent(ctx context.Context, c domain.Caller, cmd RelayIntentCommand) (domain.RelayTaskView, error) {
	r, fresh, err := s.begin(ctx, c, "", "intent", cmd)
	if err != nil {
		return domain.RelayTaskView{}, err
	}
	if !fresh {
		v, e := s.GetTask(ctx, c, r.TaskID)
		if e != nil {
			return v, receiptError(r)
		}
		return v, receiptError(r)
	}
	validate := func() error {
		if strings.TrimSpace(cmd.Task) == "" || len(cmd.Task) > 64*1024 {
			return relayError(apierrors.ValidationFailed, "bounded nonempty task description required")
		}
		if _, e := s.spec(cmd.Agent); e != nil {
			return e
		}
		_, _, e := s.work(ctx, c, cmd.WorkID, cmd.ProjectID)
		return e
	}
	if e := validate(); e != nil {
		_, err = s.finish(ctx, c, r, RelayObservation{}, e, true)
		return domain.RelayTaskView{}, err
	}
	id := uuid.NewString()
	err = s.store.Update(ctx, func(st *RelayState) error {
		t := domain.RelayTask{ID: id, WorkID: cmd.WorkID, ProjectID: cmd.ProjectID, Agent: cmd.Agent, Description: cmd.Task, PermissionLevel: "A", CreatedAt: s.now()}
		s.move(st, &t, domain.RelayReady, relayActor(c))
		r.TaskID = id
		r.Status = "accepted"
		now := s.now()
		r.FinishedAt = &now
		st.Receipts[r.ID] = r
		return nil
	})
	if err != nil {
		return domain.RelayTaskView{}, relayError(apierrors.DeliveryUnknown, "intent receipt could not be persisted; reload without replay")
	}
	return s.GetTask(ctx, c, id)
}

func (s *relayService) ConfirmIntent(ctx context.Context, c domain.Caller, id string, cmd RelayConfirmCommand) (domain.RelayActionResult, error) {
	r, fresh, err := s.begin(ctx, c, id, "dispatch", cmd)
	if err != nil {
		return domain.RelayActionResult{}, err
	}
	if !fresh {
		result, e := s.result(ctx, c, r)
		if e != nil {
			return result, e
		}
		return result, receiptError(r)
	}
	reject := func(e error) (domain.RelayActionResult, error) {
		return s.finish(ctx, c, r, RelayObservation{}, e, true)
	}
	v, err := s.GetTask(ctx, c, id)
	if err != nil {
		return reject(err)
	}
	if !cmd.PermissionConfirmed {
		return reject(relayError(apierrors.ApprovalRequired, "permission confirmation is required before dispatch"))
	}
	if cmd.PermissionLevel == "" {
		cmd.PermissionLevel = "A"
	}
	w, p, err := s.work(ctx, c, v.Task.WorkID, v.Task.ProjectID)
	if err != nil {
		return reject(err)
	}
	if cmd.ProjectSettingsRevision != p.Settings.Revision || cmd.PermissionLevel != RelayPermissionLevel(p.Settings) {
		return reject(relayError(apierrors.VersionConflict, "permission confirmation does not match current project settings"))
	}
	v.Task.PermissionLevel = cmd.PermissionLevel
	v.Task.SettingsRevision = cmd.ProjectSettingsRevision
	if _, err = s.permission(ctx, c, v.Task, "start"); err != nil {
		return reject(err)
	}
	if strings.TrimSpace(w.Handoff.Content) == "" {
		return reject(relayError(apierrors.EvidenceMissing, "complete handoff manual is required"))
	}
	if s.workspaces == nil {
		return reject(relayError(apierrors.UnsupportedCapability, "neutral workspace provider is not connected"))
	}
	workspaceID := ""
	err = s.store.Update(ctx, func(st *RelayState) error {
		t := st.Tasks[id]
		if st.PendingActions[id] != "" {
			return relayError(apierrors.VersionConflict, "another task action is pending")
		}
		if t.State != domain.RelayReady || t.Condition == "delivery_unknown" {
			return relayError(apierrors.VersionConflict, "task has already been dispatched or has unknown delivery")
		}
		for _, existing := range st.Workspaces {
			if sameRelayRoot(existing.SourceRoot, p.Root) {
				workspaceID = existing.ID
				break
			}
		}
		if workspaceID == "" {
			workspaceID = uuid.NewString()
			st.Workspaces[workspaceID] = domain.RelayWorkspace{ID: workspaceID, ProjectID: p.ID, SourceRoot: p.Root, CreatedAt: s.now()}
		}
		workspace := st.Workspaces[workspaceID]
		if workspace.ActiveTaskID != "" {
			e := relayError(apierrors.VersionConflict, "neutral workspace already has an active task")
			e.RequiredAction = "view_existing_task:" + workspace.ActiveTaskID
			return e
		}
		workspace.ActiveTaskID = id
		st.Workspaces[workspaceID] = workspace
		t.WorktreeID = workspaceID
		t.PermissionLevel = cmd.PermissionLevel
		t.SettingsRevision = cmd.ProjectSettingsRevision
		st.Tasks[id] = t
		st.PendingActions[id] = r.ID
		return nil
	})
	if err != nil {
		return reject(err)
	}
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return s.finish(ctx, c, r, RelayObservation{}, err, false)
	}
	workspace, err := s.workspaces.Prepare(ctx, w, st.Workspaces[workspaceID])
	if err != nil {
		return reject(err)
	}
	// Attach only after the neutral workspace exists. Reuse an exited terminal's
	// durable identity; a new native session remains a separate identity.
	spec, _ := s.spec(v.Task.Agent)
	err = s.store.Update(ctx, func(st *RelayState) error {
		t := st.Tasks[id]
		workspace.ActiveTaskID = id
		st.Workspaces[workspace.ID] = workspace
		terminalID := ""
		createdAt := s.now()
		for _, candidate := range st.Terminals {
			if candidate.WorktreeID == workspace.ID && candidate.Agent == t.Agent {
				terminalID = candidate.ID
				createdAt = candidate.CreatedAt
				break
			}
		}
		if terminalID == "" {
			terminalID = uuid.NewString()
		}
		st.Terminals[terminalID] = domain.RelayTerminal{ID: terminalID, WorktreeID: workspace.ID, Agent: t.Agent, DriveMode: spec.DriveMode, Liveness: domain.RelayUnverifiable, CreatedAt: createdAt}
		t.TerminalID = terminalID
		st.Tasks[id] = t
		return nil
	})
	if err != nil {
		return s.finish(ctx, c, r, RelayObservation{}, err, false)
	}
	launch, err := s.launch(ctx, c, id, "start")
	if err != nil {
		return reject(err)
	}
	launch.Input = "HANDOFF_MANUAL:\n" + w.Handoff.Content + "\n\nTASK:\n" + v.Task.Description
	driver := s.structured
	mode := "structured"
	if spec.DriveMode == "terminal" {
		driver = s.terminal
		mode = "terminal"
	}
	var obs RelayObservation
	var launchErr error
	if driver == nil {
		launchErr = &domain.NativePrelaunchFailure{Cause: relayError(apierrors.UnsupportedCapability, "selected drive path is not connected")}
	} else {
		obs, launchErr = driver.Launch(ctx, launch)
	}
	var prelaunch *domain.NativePrelaunchFailure
	degraded := false
	if launchErr != nil && mode == "structured" && spec.DriveMode == "both" && errors.As(launchErr, &prelaunch) {
		var known *apierrors.ServiceError
		if errors.As(prelaunch.Cause, &known) && known.Code == apierrors.UnsupportedCapability && s.terminal != nil {
			degraded = true
			mode = "terminal"
			obs, launchErr = s.terminal.Launch(ctx, launch)
		}
	}
	prelaunch = nil
	knownUnstarted := launchErr != nil && errors.As(launchErr, &prelaunch)
	// Drive choice and accepted edge persist before observation edges.
	err = s.store.Update(context.WithoutCancel(ctx), func(st *RelayState) error {
		t := st.Tasks[id]
		t.DegradedLaunch = degraded
		terminal := st.Terminals[t.TerminalID]
		terminal.DriveMode = mode
		st.Terminals[terminal.ID] = terminal
		if launchErr == nil {
			s.move(st, &t, domain.RelayAccepted, relayActor(c))
		}
		// A proven unstarted process is not an observed process exit. Its durable
		// terminal identity remains unverifiable rather than fabricated exited.
		return nil
	})
	if err != nil {
		return domain.RelayActionResult{}, relayError(apierrors.DeliveryUnknown, "launch state could not be persisted; do not replay")
	}
	return s.finish(ctx, c, r, obs, launchErr, knownUnstarted)
}

func (s *relayService) launch(ctx context.Context, c domain.Caller, id, action string) (RelayLaunch, error) {
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return RelayLaunch{}, err
	}
	t, ok := st.Tasks[id]
	if !ok {
		return RelayLaunch{}, apierrors.NewNotFound("relay task", id)
	}
	p, err := s.permission(ctx, c, t, action)
	if err != nil {
		return RelayLaunch{}, err
	}
	spec, err := s.spec(t.Agent)
	if err != nil {
		return RelayLaunch{}, err
	}
	w, ok := st.Workspaces[t.WorktreeID]
	if !ok || !w.Prepared || w.ActiveTaskID != t.ID {
		return RelayLaunch{}, relayError(apierrors.VersionConflict, "task does not own a prepared neutral workspace")
	}
	if !sameRelayRoot(w.SourceRoot, p.Root) {
		return RelayLaunch{}, relayError(apierrors.ScopeDenied, "project root changed from the confirmed workspace source")
	}
	terminal, ok := st.Terminals[t.TerminalID]
	if !ok {
		return RelayLaunch{}, relayError(apierrors.EvidenceMissing, "task terminal identity is absent")
	}
	return RelayLaunch{Spec: spec, Project: p, Workspace: w, Terminal: terminal, Session: st.Sessions[terminal.ID]}, nil
}

func (s *relayService) observe(st *RelayState, t *domain.RelayTask, obs RelayObservation) {
	terminal := st.Terminals[t.TerminalID]
	if obs.Session.ID != "" {
		st.Sessions[terminal.ID] = obs.Session
	}
	for _, event := range obs.Events {
		if event.Sequence <= terminal.LastSequence || event.TS.IsZero() || event.TS.After(s.now()) {
			continue
		}
		terminal.LastSequence = event.Sequence
		when := event.TS.UTC()
		terminal.LastEventAt = &when
		event.ID = uuid.NewString()
		event.TaskID = t.ID
		event.TerminalID = terminal.ID
		event.WorktreeID = t.WorktreeID
		event.Agent = t.Agent
		event.Topic = "relay:" + t.ID
		event.Scope = t.ProjectID
		st.Events = append(st.Events, event)
		if t.State == domain.RelayAccepted {
			s.move(st, t, domain.RelayWorking, "observation:"+event.Source)
		}
	}
	if obs.ExitConfirmed {
		if terminal.ExitedAt == nil {
			now := s.now()
			terminal.ExitedAt = &now
		}
		if t.State != domain.RelayDone && t.State != domain.RelayReleased {
			t.Condition = "terminal_exited"
			st.Tasks[t.ID] = *t
		}
	}
	terminal.Liveness = domain.RelayLiveness(terminal.LastEventAt, terminal.ExitedAt, s.now(), s.options.LivenessWindow)
	st.Terminals[terminal.ID] = terminal
}

func (s *relayService) publish(ctx context.Context, id string) error {
	if s.events == nil {
		return nil
	}
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return err
	}
	if st.PublicationErrors[id] {
		return relayError(apierrors.DeliveryUnknown, "event publication is pending explicit reconciliation")
	}
	for _, event := range st.Events {
		if event.TaskID != id || st.Published[event.ID] {
			continue
		}
		if err = s.events.EmitRelayEvent(ctx, event); err != nil {
			_ = s.store.Update(context.WithoutCancel(ctx), func(current *RelayState) error { current.PublicationErrors[id] = true; return nil })
			return relayError(apierrors.DeliveryUnknown, "derived event publication could not be verified")
		}
		if err = s.store.Update(ctx, func(current *RelayState) error { current.Published[event.ID] = true; return nil }); err != nil {
			return err
		}
	}
	return nil
}

func (s *relayService) action(ctx context.Context, c domain.Caller, id, action string, input any) (domain.RelayActionResult, error) {
	r, fresh, err := s.begin(ctx, c, id, action, input)
	if err != nil {
		return domain.RelayActionResult{}, err
	}
	if !fresh {
		result, e := s.result(ctx, c, r)
		if e != nil {
			return result, e
		}
		return result, receiptError(r)
	}
	reject := func(e error) (domain.RelayActionResult, error) {
		return s.finish(ctx, c, r, RelayObservation{}, e, true)
	}
	permissionAction := action
	if action == "read" {
		permissionAction = "observe"
	}
	launch, err := s.launch(ctx, c, id, permissionAction)
	if err != nil {
		return reject(err)
	}
	view, err := s.GetTask(ctx, c, id)
	if err != nil {
		return reject(err)
	}
	if view.Task.State == domain.RelayReleased {
		return reject(relayError(apierrors.VersionConflict, "released task cannot control the terminal"))
	}
	if action == "send" && (view.Task.Condition == "delivery_unknown" || view.Terminal.ExitedAt != nil || view.Task.State == domain.RelayDone) {
		return reject(relayError(apierrors.VersionConflict, "task cannot accept another input"))
	}
	if action == "resume" && (view.Terminal.ExitedAt == nil || view.Task.State == domain.RelayDone) {
		return reject(relayError(apierrors.DeliveryUnknown, "resume requires a confirmed exit of the original task"))
	}
	// Serialize conflicting actions durably without holding SQL across process IO.
	err = s.store.Update(ctx, func(st *RelayState) error {
		if st.PendingActions[id] != "" {
			return relayError(apierrors.VersionConflict, "another task action is pending")
		}
		st.PendingActions[id] = r.ID
		return nil
	})
	if err != nil {
		return reject(err)
	}
	driver := s.structured
	if launch.Terminal.DriveMode == "terminal" {
		driver = s.terminal
	}
	if driver == nil {
		return reject(relayError(apierrors.UnsupportedCapability, "task drive path is not connected"))
	}
	var obs RelayObservation
	switch action {
	case "send":
		cmd := input.(RelaySendCommand)
		if strings.TrimSpace(cmd.Input) == "" || len(cmd.Input) > 64*1024 {
			return reject(relayError(apierrors.ValidationFailed, "bounded nonempty input required"))
		}
		obs, err = driver.Send(ctx, launch, cmd.Input)
	case "read":
		obs, err = driver.Read(ctx, launch)
		if err == nil && launch.Terminal.DriveMode == "terminal" && s.workspaces != nil {
			var git string
			git, err = s.workspaces.Observe(ctx, launch.Workspace)
			if err == nil && git != launch.Workspace.GitObservation {
				// Git events have an independent sequence namespace.
				e := domain.RelayEvent{ID: uuid.NewString(), TaskID: id, TerminalID: launch.Terminal.ID, WorktreeID: launch.Workspace.ID, Agent: launch.Spec.Name, Kind: "git_activity", Source: "git", Topic: "relay:" + id, Scope: launch.Project.ID, TS: s.now()}
				err = s.store.Update(ctx, func(st *RelayState) error {
					workspace := st.Workspaces[launch.Workspace.ID]
					if workspace.GitObservation == git {
						return nil
					}
					workspace.GitObservation = git
					st.Workspaces[workspace.ID] = workspace
					st.Events = append(st.Events, e)
					t := st.Tasks[id]
					if t.State == domain.RelayAccepted {
						s.move(st, &t, domain.RelayWorking, "observation:git")
					}
					return nil
				})
			}
		}
	case "stop":
		obs, err = driver.Stop(ctx, launch)
	case "resume":
		obs, err = driver.Resume(ctx, launch)
	}
	if action == "resume" && err == nil {
		if e := s.store.Update(context.WithoutCancel(ctx), func(st *RelayState) error {
			t := st.Tasks[id]
			t.Condition = ""
			st.Tasks[id] = t
			terminal := st.Terminals[t.TerminalID]
			terminal.ExitedAt = nil
			terminal.LastEventAt = nil
			terminal.LastSequence = 0
			st.Terminals[terminal.ID] = terminal
			return nil
		}); e != nil {
			return domain.RelayActionResult{}, relayError(apierrors.DeliveryUnknown, "resume state is uncertain; do not replay")
		}
	}
	return s.finish(ctx, c, r, obs, err, false)
}

func (s *relayService) Send(ctx context.Context, c domain.Caller, id string, cmd RelaySendCommand) (domain.RelayActionResult, error) {
	return s.action(ctx, c, id, "send", cmd)
}
func (s *relayService) Read(ctx context.Context, c domain.Caller, id string) (domain.RelayActionResult, error) {
	return s.action(ctx, c, id, "read", nil)
}
func (s *relayService) Stop(ctx context.Context, c domain.Caller, id string) (domain.RelayActionResult, error) {
	return s.action(ctx, c, id, "stop", nil)
}
func (s *relayService) Resume(ctx context.Context, c domain.Caller, id string) (domain.RelayActionResult, error) {
	return s.action(ctx, c, id, "resume", nil)
}

func (s *relayService) Transition(ctx context.Context, c domain.Caller, id string, cmd RelayTransitionCommand) (domain.RelayTaskView, error) {
	r, fresh, err := s.begin(ctx, c, id, "transition", cmd)
	if err != nil {
		return domain.RelayTaskView{}, err
	}
	if !fresh {
		v, e := s.GetTask(ctx, c, id)
		if e != nil {
			return v, e
		}
		return v, receiptError(r)
	}
	view, e := s.GetTask(ctx, c, id)
	if e == nil {
		_, e = s.permission(ctx, c, view.Task, "reconcile")
	}
	if e != nil {
		_, finishErr := s.finish(ctx, c, r, RelayObservation{}, e, true)
		return domain.RelayTaskView{}, finishErr
	}
	err = s.store.Update(ctx, func(st *RelayState) error {
		t, ok := st.Tasks[id]
		if !ok {
			return apierrors.NewNotFound("relay task", id)
		}
		if st.PendingActions[id] != "" {
			return relayError(apierrors.VersionConflict, "another task action is pending")
		}
		if t.State != cmd.From || !domain.ValidRelayTransition(cmd.From, cmd.To) {
			return relayError(apierrors.VersionConflict, "invalid or stale task transition")
		}
		if cmd.To != domain.RelayDone && cmd.To != domain.RelayReleased {
			return relayError(apierrors.ScopeDenied, "accepted and working are execution observations")
		}
		if cmd.To == domain.RelayDone && !cmd.CompletionConfirmed {
			return relayError(apierrors.ApprovalRequired, "human task completion confirmation is required")
		}
		if cmd.To == domain.RelayDone {
			t.Condition = ""
		}
		if cmd.To == domain.RelayReleased {
			terminal, ok := st.Terminals[t.TerminalID]
			if !ok || terminal.ExitedAt == nil {
				return relayError(apierrors.DeliveryUnknown, "release requires confirmed process exit")
			}
			for _, other := range st.Receipts {
				if other.ID != r.ID && other.TaskID == id && other.Status == "pending" {
					return relayError(apierrors.VersionConflict, "task action is pending")
				}
			}
			workspace := st.Workspaces[t.WorktreeID]
			if workspace.ActiveTaskID == id {
				workspace.ActiveTaskID = ""
				st.Workspaces[workspace.ID] = workspace
			}
		}
		s.move(st, &t, cmd.To, relayActor(c))
		return nil
	})
	_, finishErr := s.finish(ctx, c, r, RelayObservation{}, err, true)
	if finishErr != nil {
		return domain.RelayTaskView{}, finishErr
	}
	return s.GetTask(ctx, c, id)
}
