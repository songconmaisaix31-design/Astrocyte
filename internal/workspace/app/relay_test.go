package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayWorks struct {
	app.Works
	work domain.Work
}

func (w *relayWorks) ListWorks(context.Context, domain.Caller) (domain.WorksSnapshot, error) {
	return domain.WorksSnapshot{Works: []domain.Work{w.work}}, nil
}

type relayProjects struct {
	app.LocalProjectRepository
	project domain.LocalProject
	grant   *domain.ProjectGrant
}

func (p *relayProjects) LoadProject(context.Context, string) (domain.LocalProject, error) {
	return p.project, nil
}
func (p *relayProjects) LoadGrant(context.Context, string, string) (domain.ProjectGrant, error) {
	if p.grant == nil {
		return domain.ProjectGrant{}, apierrors.NewNotFound("grant", "test")
	}
	return *p.grant, nil
}

type relaySpaces struct {
	root     string
	prepares atomic.Int32
	git      string
}

func (w *relaySpaces) Prepare(_ context.Context, _ domain.Work, space domain.RelayWorkspace) (domain.RelayWorkspace, error) {
	if space.Prepared {
		return space, nil
	}
	w.prepares.Add(1)
	space.Root = filepath.Join(w.root, space.ID)
	space.Branch = "relay/" + space.ID
	space.Prepared = true
	if e := os.MkdirAll(space.Root, 0700); e != nil {
		return space, e
	}
	return space, nil
}
func (w *relaySpaces) Observe(context.Context, domain.RelayWorkspace) (string, error) {
	return w.git, nil
}

type relayDriver struct {
	launches, sends, reads, stops atomic.Int32
	launchErr                     error
	sendErr                       error
	enter, release                chan struct{}
	lastInput                     string
	silent                        bool
	now                           time.Time
}

func (d *relayDriver) observation(r app.RelayLaunch) app.RelayObservation {
	session := r.Session
	if session.ID == "" {
		session.ID = uuid.NewString()
		session.NativeID = "original-native"
		session.Ownership = "owned"
	}
	obs := app.RelayObservation{Session: session}
	if !d.silent {
		obs.Events = []domain.RelayEvent{{Sequence: 1, Kind: "turn_started", Source: "native", TS: d.now}}
	}
	return obs
}
func (d *relayDriver) Launch(_ context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	d.launches.Add(1)
	d.lastInput = r.Input
	if d.enter != nil {
		close(d.enter)
		<-d.release
	}
	return d.observation(r), d.launchErr
}
func (d *relayDriver) Send(_ context.Context, r app.RelayLaunch, _ string) (app.RelayObservation, error) {
	d.sends.Add(1)
	return d.observation(r), d.sendErr
}

func TestRelaySendUnknownReceiptBlocksFurtherInput(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	_, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	f.driver.sendErr = &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "original send unknown"}
	op := caller()
	input := app.RelaySendCommand{Input: "one approved message"}
	r, e := f.service.Send(ctx, op, id, input)
	checkCode(t, e, apierrors.DeliveryUnknown)
	if r.Receipt.Status != "delivery_unknown" {
		t.Fatal("uncertain send receipt lost")
	}
	_, e = f.service.Send(ctx, op, id, input)
	checkCode(t, e, apierrors.DeliveryUnknown)
	_, e = f.service.Send(ctx, caller(), id, input)
	checkCode(t, e, apierrors.VersionConflict)
	if f.driver.sends.Load() != 1 {
		t.Fatal("unknown input delivered twice")
	}
}

func TestRelayFallbackRequiresExplicitUnstartedUnsupported(t *testing.T) {
	for _, test := range []struct {
		name      string
		launchErr error
		code      apierrors.Code
	}{
		{"provider_rejected", &domain.NativePrelaunchFailure{Cause: &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "not installed"}}, apierrors.ProviderUnavailable},
		{"unsupported_without_exit_proof", &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "unknown effects"}, apierrors.DeliveryUnknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newRelayFixture(t)
			f.options.Agents[0].DriveMode = "both"
			f.driver.launchErr = test.launchErr
			terminal := &relayDriver{now: f.driver.now}
			f.service = f.assemble(f.store, f.driver, terminal)
			id := f.intent()
			r, e := f.service.ConfirmIntent(context.Background(), caller(), id, confirm())
			checkCode(t, e, test.code)
			if r.View.Terminal.Liveness == "exited" || r.View.Terminal.ExitedAt != nil {
				t.Fatal("unstarted launch fabricated main process exit")
			}
			if terminal.launches.Load() != 0 {
				t.Fatal("unsafe fallback started another writer")
			}
		})
	}
}
func (d *relayDriver) Read(_ context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	d.reads.Add(1)
	return d.observation(r), nil
}
func (d *relayDriver) Stop(_ context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	d.stops.Add(1)
	obs := d.observation(r)
	obs.ExitConfirmed = true
	obs.Session.StopConfirmed = true
	return obs, nil
}
func (d *relayDriver) Resume(_ context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	return d.observation(r), nil
}

type relaySink struct {
	mu      sync.Mutex
	events  []domain.RelayEvent
	failure error
}

func (s *relaySink) EmitRelayEvent(_ context.Context, v domain.RelayEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, v)
	return s.failure
}

func TestRelaySinkFailurePersistsVisiblePublicationWithoutReplayingLaunch(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	f.sink.failure = errors.New("PRIVATE_SINK_ERROR_SENTINEL")
	id := f.intent()
	r, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	if r.Receipt.Status != "accepted" || r.View.EventPublication.Status != "error" || r.View.EventPublication.PendingEvents != 1 {
		t.Fatal("launch acceptance and publication failure conflated")
	}
	encoded, _ := json.Marshal(r)
	if strings.Contains(string(encoded), "PRIVATE_SINK_ERROR_SENTINEL") {
		t.Fatal("sink error raw exposed")
	}
	if e = f.store.Close(); e != nil {
		t.Fatal(e)
	}
	f.store, e = sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	f.service = f.assemble(f.store, f.driver, nil)
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.EventPublication.Status != "error" {
		t.Fatal("restart lost publication failure")
	}
	_, e = f.service.Read(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	if len(f.sink.events) != 1 || f.driver.launches.Load() != 1 {
		t.Fatal("publication failure automatically retried sink or launch")
	}
	f.sink.failure = nil
	v, e = f.service.(app.RelayLifecycle).ReconcileEvents(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	if v.EventPublication.Status != "published" || v.EventPublication.PublishedEvents != 1 || v.EventPublication.PendingEvents != 0 || f.driver.launches.Load() != 1 {
		t.Fatal("explicit event reconciliation re-dispatched or lost publication evidence")
	}
	if len(f.sink.events) != 2 || f.sink.events[0].ID != f.sink.events[1].ID {
		t.Fatal("reconciliation changed stable event origin")
	}
}

func TestRelayLifecycleObservesOnlyOwnedTasksAndShutdownPersistsExit(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	_, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	lifecycle := f.service.(app.RelayLifecycle)
	if e = lifecycle.ObserveOwned(ctx); e != nil {
		t.Fatal(e)
	}
	if f.driver.launches.Load() != 1 || f.driver.reads.Load() != 1 {
		t.Fatal("observer launched or resumed a task")
	}
	revoked := time.Now()
	f.projects.grant = &domain.ProjectGrant{RevokedAt: &revoked}
	if e = lifecycle.Shutdown(ctx); e != nil {
		t.Fatal(e)
	}
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Terminal.Liveness != "exited" || v.Task.State != "working" || v.Workspace.ActiveTaskID != id {
		t.Fatal("shutdown released or completed unfinished task")
	}
	receipts, e := f.service.ListReceipts(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	saw := false
	for _, r := range receipts {
		if r.Action == "stop" && r.Operator == "service:shutdown" && r.Status == "accepted" {
			saw = true
		}
	}
	if !saw {
		t.Fatal("shutdown missing containment receipt")
	}
}

type relayFixture struct {
	t        *testing.T
	root     string
	store    *sqlite.RelayStore
	works    *relayWorks
	projects *relayProjects
	spaces   *relaySpaces
	driver   *relayDriver
	sink     *relaySink
	options  app.RelayServiceOptions
	service  app.RelayRuntime
}

func newRelayFixture(t *testing.T) *relayFixture {
	t.Helper()
	t.Setenv("LOCALAPPDATA", t.TempDir())
	root, e := sqlite.RelayStateRoot()
	if e != nil {
		t.Fatal(e)
	}
	store, e := sqlite.OpenRelayStore(root)
	if e != nil {
		t.Fatal(e)
	}
	source := t.TempDir()
	now := time.Now().UTC()
	f := &relayFixture{t: t, root: root, store: store, works: &relayWorks{work: domain.Work{ID: "work", Root: source, Kind: "folder", Handoff: domain.WorkHandoff{Content: "complete manual\nsecond line"}}},
		projects: &relayProjects{project: domain.LocalProject{ID: "project", Root: source, Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "alpha", AllowedActions: []string{"start", "send", "observe", "stop", "resume", "reconcile"}}}},
		spaces:   &relaySpaces{root: t.TempDir()}, driver: &relayDriver{now: now}, sink: &relaySink{}, options: app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "alpha", DriveMode: "structured"}, {Name: "beta", DriveMode: "structured"}}, LivenessWindow: time.Second, Clock: func() time.Time { return now }}}
	f.service = f.assemble(store, f.driver, nil)
	t.Cleanup(func() { _ = f.store.Close() })
	return f
}
func (f *relayFixture) assemble(store app.RelayStore, structured, terminal app.RelayDriver) app.RelayRuntime {
	return app.NewRelayService(store, f.works, f.projects, f.spaces, structured, terminal, f.sink, f.options)
}
func caller() domain.Caller {
	return domain.Caller{Kind: "human", ID: "operator", OperationID: uuid.NewString()}
}
func (f *relayFixture) intent() string {
	f.t.Helper()
	v, e := f.service.CreateIntent(context.Background(), caller(), app.RelayIntentCommand{WorkID: "work", ProjectID: "project", Agent: f.projects.project.Settings.ExternalModelCLI, Task: "explicit task"})
	if e != nil {
		f.t.Fatal(e)
	}
	return v.Task.ID
}
func confirm() app.RelayConfirmCommand {
	return app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: "A", ProjectSettingsRevision: 1}
}
func checkCode(t *testing.T, e error, code apierrors.Code) {
	t.Helper()
	var known *apierrors.ServiceError
	if !errors.As(e, &known) || known.Code != code {
		t.Fatalf("want %s, got %v", code, e)
	}
}

func TestRelayPersistenceTransitionsIdentityAndSameOperation(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	if f.spaces.prepares.Load() != 0 || f.driver.launches.Load() != 0 {
		t.Fatal("intent launched work")
	}
	op := caller()
	result, e := f.service.ConfirmIntent(ctx, op, id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	if result.View.Task.State != "working" || result.View.Terminal.Liveness != "live" {
		t.Fatal("actual event not represented")
	}
	if f.driver.lastInput != "HANDOFF_MANUAL:\ncomplete manual\nsecond line\n\nTASK:\nexplicit task" {
		t.Fatal("first input omitted manual or task")
	}
	terminalID := result.View.Terminal.ID
	nativeID := ""
	workspaceID := result.View.Workspace.ID
	state, e := f.store.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	nativeID = state.Sessions[terminalID].NativeID
	if state.Workspaces[workspaceID].SourceRoot != f.works.work.Root || nativeID == "" {
		t.Fatal("private workspace/session state lost")
	}
	if e = f.store.Close(); e != nil {
		t.Fatal(e)
	}
	f.store, e = sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	f.service = f.assemble(f.store, f.driver, nil)
	repeat, e := f.service.ConfirmIntent(ctx, op, id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	if f.driver.launches.Load() != 1 || repeat.View.Terminal.ID != terminalID {
		t.Fatal("restart repeated launch or changed identity")
	}
	wrong := confirm()
	wrong.PermissionConfirmed = false
	_, e = f.service.ConfirmIntent(ctx, op, id, wrong)
	checkCode(t, e, apierrors.VersionConflict)
	state, e = f.store.Snapshot(ctx)
	if e != nil || state.Sessions[terminalID].NativeID != nativeID {
		t.Fatal("restart lost native identity")
	}
	_, e = f.service.Stop(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Task.Condition != "terminal_exited" || v.Terminal.Liveness != "exited" {
		t.Fatal("exit silently completed task")
	}
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "working", To: "done"})
	checkCode(t, e, apierrors.ApprovalRequired)
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true})
	if e != nil {
		t.Fatal(e)
	}
	v, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "done", To: "released"})
	if e != nil {
		t.Fatal(e)
	}
	if v.Task.State != "released" || v.Workspace.ActiveTaskID != "" {
		t.Fatal("explicit release failed")
	}
	want := []string{"ready", "accepted", "working", "done", "released"}
	if len(v.Transitions) != len(want) {
		t.Fatal("state edges were not all persisted")
	}
	for i, edge := range v.Transitions {
		if edge.To != want[i] || edge.TaskID != id || edge.Operator == "" || edge.TS.IsZero() {
			t.Fatal("incomplete transition evidence")
		}
	}
}

func TestRelayRaceRejectsSecondWriterAcrossStores(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	first, second := f.intent(), f.intent()
	f.driver.enter = make(chan struct{})
	f.driver.release = make(chan struct{})
	other, e := sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	defer other.Close()
	service := f.assemble(other, f.driver, nil)
	done := make(chan error, 1)
	go func() { _, e := f.service.ConfirmIntent(ctx, caller(), first, confirm()); done <- e }()
	select {
	case <-f.driver.enter:
	case <-time.After(5 * time.Second):
		t.Fatal("first launch never entered")
	}
	_, e = service.ConfirmIntent(ctx, caller(), second, confirm())
	checkCode(t, e, apierrors.VersionConflict)
	var known *apierrors.ServiceError
	_ = errors.As(e, &known)
	if known.RequiredAction != "view_existing_task:"+first {
		t.Fatal("duplicate did not identify the active task")
	}
	_, e = service.ConfirmIntent(ctx, caller(), first, confirm())
	checkCode(t, e, apierrors.VersionConflict)
	state, e := other.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	for _, w := range state.Workspaces {
		if w.ActiveTaskID != first {
			t.Fatal("rejected duplicate cleared winning reservation")
		}
	}
	close(f.driver.release)
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if f.driver.launches.Load() != 1 || f.spaces.prepares.Load() != 1 {
		t.Fatal("race launched a second writer")
	}
}

func TestRelayPermissionDenialReceiptFirstREDAndRevision(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	op := caller()
	cmd := confirm()
	cmd.PermissionConfirmed = false
	r, e := f.service.ConfirmIntent(ctx, op, id, cmd)
	checkCode(t, e, apierrors.ApprovalRequired)
	if r.Receipt.Status != "rejected" || r.Receipt.Error == nil {
		t.Fatal("permission RED not stored")
	}
	original, _ := json.Marshal(r.Receipt)
	_, e = f.service.ConfirmIntent(ctx, op, id, cmd)
	checkCode(t, e, apierrors.ApprovalRequired)
	all, e := f.service.ListReceipts(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	for _, receipt := range all {
		if receipt.ID == r.Receipt.ID {
			data, _ := json.Marshal(receipt)
			if string(original) != string(data) {
				t.Fatal("first RED overwritten")
			}
		}
	}
	if f.driver.launches.Load() != 0 || f.spaces.prepares.Load() != 0 {
		t.Fatal("denial had workspace/process effects")
	}
	f.projects.project.Settings.Revision = 2
	_, e = f.service.ConfirmIntent(ctx, caller(), id, confirm())
	checkCode(t, e, apierrors.VersionConflict)
	f.projects.project.Settings.Revision = 1
	revoked := time.Now()
	f.projects.grant = &domain.ProjectGrant{RevokedAt: &revoked}
	_, e = f.service.ConfirmIntent(ctx, caller(), id, confirm())
	checkCode(t, e, apierrors.ApprovalRevoked)
	_, e = f.service.CreateIntent(ctx, domain.Caller{Kind: "agent", ID: "alpha", OperationID: uuid.NewString()}, app.RelayIntentCommand{})
	checkCode(t, e, apierrors.ScopeDenied)
}

func TestRelayUnknownNeverResendsOrDowngrades(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	f.options.Agents[0].DriveMode = "both"
	f.driver.launchErr = &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "first uncertain delivery"}
	terminal := &relayDriver{now: f.driver.now}
	f.service = f.assemble(f.store, f.driver, terminal)
	id := f.intent()
	op := caller()
	_, e := f.service.ConfirmIntent(ctx, op, id, confirm())
	checkCode(t, e, apierrors.DeliveryUnknown)
	if e = f.store.Close(); e != nil {
		t.Fatal(e)
	}
	f.store, e = sqlite.OpenRelayStore(f.root)
	if e != nil {
		t.Fatal(e)
	}
	f.service = f.assemble(f.store, f.driver, terminal)
	_, e = f.service.ConfirmIntent(ctx, op, id, confirm())
	checkCode(t, e, apierrors.DeliveryUnknown)
	_, e = f.service.ConfirmIntent(ctx, caller(), id, confirm())
	checkCode(t, e, apierrors.VersionConflict)
	second := f.intent()
	_, e = f.service.ConfirmIntent(ctx, caller(), second, confirm())
	checkCode(t, e, apierrors.VersionConflict)
	if f.driver.launches.Load() != 1 || terminal.launches.Load() != 0 {
		t.Fatal("unknown delivery retried or degraded")
	}
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Task.Condition != "delivery_unknown" || v.Workspace.ActiveTaskID != id {
		t.Fatal("unknown lost its writer reservation")
	}
}

func TestRelayKnownUnsupportedCanDegradeAndGitNeverCompletes(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	f.options.Agents[0].DriveMode = "both"
	f.driver.launchErr = &domain.NativePrelaunchFailure{Cause: &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "no structured tools"}}
	terminal := &relayDriver{now: f.driver.now, silent: true}
	f.service = f.assemble(f.store, f.driver, terminal)
	id := f.intent()
	r, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	if !r.View.Task.DegradedLaunch || r.View.Terminal.DriveMode != "terminal" || r.Receipt.Status != "accepted" || r.View.Task.State != "accepted" {
		t.Fatal("safe degradation misrepresented")
	}
	f.spaces.git = "actual new Git observation"
	r, e = f.service.Read(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	if r.View.Task.State != "working" || r.View.Terminal.Liveness != "unverifiable" {
		t.Fatal("Git became completion or terminal liveness proof")
	}
	if f.driver.launches.Load() != 1 || terminal.launches.Load() != 1 {
		t.Fatal("wrong degradation launch count")
	}
	events, e := f.service.ListEvents(ctx, caller(), id)
	if e != nil || len(events) != 1 || events[0].Kind != "git_activity" {
		t.Fatal("Git activity did not enter derived event sink")
	}
	if len(f.sink.events) != 1 {
		t.Fatal("Git event sink not called")
	}
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "done", To: "released"})
	checkCode(t, e, apierrors.DeliveryUnknown)
}

func TestRelaySameWorkspaceHandoffPreservesFiles(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	r, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(r.View.Workspace.Root, "unfinished.txt")
	if e = os.WriteFile(file, []byte("human-owned relay state"), 0600); e != nil {
		t.Fatal(e)
	}
	firstWorkspace, firstTerminal := r.View.Workspace.ID, r.View.Terminal.ID
	_, e = f.service.Stop(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "working", To: "done", CompletionConfirmed: true})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.service.Transition(ctx, caller(), id, app.RelayTransitionCommand{From: "done", To: "released"})
	if e != nil {
		t.Fatal(e)
	}
	f.projects.project.Settings.ExternalModelCLI = "beta"
	second := f.intent()
	r, e = f.service.ConfirmIntent(ctx, caller(), second, confirm())
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(file)
	if e != nil || string(data) != "human-owned relay state" || r.View.Workspace.ID != firstWorkspace || r.View.Terminal.ID == firstTerminal || f.spaces.prepares.Load() != 1 {
		t.Fatal("handoff rebuilt or discarded workspace")
	}
}

func TestRelayThreeValuesAndNoSilentTimestampRefresh(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	r, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	if r.View.Terminal.Liveness != "live" {
		t.Fatal("new event not live")
	}
	now := f.driver.now.Add(2 * time.Second)
	f.options.Clock = func() time.Time { return now }
	f.service = f.assemble(f.store, f.driver, nil)
	r, e = f.service.Read(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	if r.View.Terminal.Liveness != "unverifiable" {
		t.Fatal("re-reading old event refreshed liveness")
	}
	_, e = f.service.Stop(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Terminal.Liveness != "exited" || v.Task.State != "working" {
		t.Fatal("confirmed exit became task done")
	}
}

func TestRelayReceiptsAndEventsDoNotExposeInternalInput(t *testing.T) {
	f := newRelayFixture(t)
	ctx := context.Background()
	id := f.intent()
	_, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	for _, run := range []func() error{
		func() error {
			_, e := f.service.Send(ctx, caller(), id, app.RelaySendCommand{Input: "PRIVATE_SEND_SENTINEL"})
			return e
		},
		func() error { _, e := f.service.Read(ctx, caller(), id); return e },
		func() error { _, e := f.service.Stop(ctx, caller(), id); return e },
	} {
		if e = run(); e != nil {
			t.Fatal(e)
		}
	}
	receipts, e := f.service.ListReceipts(ctx, caller(), id)
	if e != nil {
		t.Fatal(e)
	}
	encoded, _ := json.Marshal(receipts)
	if strings.Contains(string(encoded), "PRIVATE_SEND_SENTINEL") || strings.Contains(string(encoded), "Input") {
		t.Fatal("private receipt input leaked")
	}
	seen := map[string]bool{}
	for _, r := range receipts {
		seen[r.Action] = true
	}
	for _, action := range []string{"dispatch", "send", "read", "stop"} {
		if !seen[action] {
			t.Fatalf("missing %s receipt", action)
		}
	}
}
