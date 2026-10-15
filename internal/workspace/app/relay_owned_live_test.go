package app_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 使用真实 Native 端口，只在实际 Observe 返回后取消本次来源读取。
type relayLiveCancelAfterObserve struct {
	*agents.Native
	cancel context.CancelFunc
}

func (a *relayLiveCancelAfterObserve) Observe(ctx context.Context, s domain.NativeSession) (domain.NativeObservation, error) {
	obs, err := a.Native.Observe(ctx, s)
	if err == nil && a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	return obs, err
}

type relayLiveRegistry struct {
	app.NativeRegistry
	adapter app.NativeAdapter
}

func (r relayLiveRegistry) Adapter(id string) (app.NativeAdapter, error) {
	if id == "codex" {
		return r.adapter, nil
	}
	return r.NativeRegistry.Adapter(id)
}

func TestRelayOwnedActualCodexLifecycleAndSQLite(t *testing.T) {
	if os.Getenv("ASTROCYTE_RELAY_L1_NATIVE") != "1" {
		t.Skip("requires an explicitly authorized actual Codex public task")
	}
	base := os.Getenv("ASTROCYTE_RELAY_L1_EVIDENCE")
	if !filepath.IsAbs(base) {
		t.Fatal("absolute ignored evidence directory required")
	}
	root := filepath.Join(base, "actual-"+uuid.NewString())
	sourceRoot := filepath.Join(root, "public-git")
	if err := os.MkdirAll(sourceRoot, 0700); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", sourceRoot}, {"-C", sourceRoot, "-c", "user.name=Relay L1", "-c", "user.email=relay-l1@example.invalid", "-c", "core.hooksPath=", "commit", "--allow-empty", "-m", "Public L1 session task"}} {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatal("public Git initialization failed")
		}
	}
	canonical, err := agents.CanonicalRoot(sourceRoot)
	if err != nil {
		t.Fatal(err)
	}
	sourceRoot = canonical
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "local-state"))
	stateRoot, err := sqlite.RelayStateRoot()
	if err != nil {
		t.Fatal(err)
	}
	store, err := sqlite.OpenRelayStore(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	db, err := sqlite.Open(filepath.Join(root, "application.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(filepath.Join(root, "controller"))
	adapter, err := registry.Adapter("codex")
	if err != nil {
		t.Fatal(err)
	}
	native := adapter.(*agents.Native)
	controlled := &relayLiveCancelAfterObserve{Native: native}
	driver := agents.NewRelayNativeDriver(relayLiveRegistry{NativeRegistry: registry, adapter: controlled}, agents.RelayNativeDriverOptions{CaptureNativeNotifications: true})
	projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	human := func() domain.Caller {
		return domain.Caller{Kind: "human", ID: "public-l1-owner", OperationID: uuid.NewString()}
	}
	project, err := projects.RegisterProject(ctx, human(), domain.RegisterProjectCommand{Name: "Public L1 session", Root: sourceRoot, SpaceID: "public-l1"})
	if err != nil {
		t.Fatal(err)
	}
	actions := []string{"start", "send", "observe", "stop", "resume", "reconcile"}
	project, err = projects.SetProjectSettings(ctx, human(), project.ID, domain.SettingsCommand{ExpectedRevision: 1, Settings: domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: actions, AllowAgentControl: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projects.GrantProjectAgent(ctx, human(), project.ID, domain.ProjectGrant{AgentID: "codex", Actions: actions}); err != nil {
		t.Fatal(err)
	}
	works := app.NewWorksService(db, nil, worksscan.Delivery{}, nil, app.WorksServiceOptions{})
	delivery, err := json.Marshal(map[string]any{"works": []any{map[string]any{"platform": "codex", "name": "Public L1 session", "path": sourceRoot, "kind": "folder", "handoff": "Use only the public text in this task. Files and command tools remain disabled."}}})
	if err != nil {
		t.Fatal(err)
	}
	album, err := works.ImportWorksDelivery(ctx, human(), app.WorksDeliveryCommand{Content: delivery})
	if err != nil || len(album.Works) != 1 {
		t.Fatal("public owned work import failed")
	}
	spaces := relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true})
	options := app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}, OwnedRoots: agents.RelayOwnedRoots{}, LivenessWindow: 10 * time.Second}
	service := app.NewRelayService(store, works, db, spaces, driver, nil, nil, options)
	lifecycle := service.(app.RelayLifecycle)
	shutdownDone := false
	t.Cleanup(func() {
		if shutdownDone {
			return
		}
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		if err := lifecycle.Shutdown(stopCtx); err != nil {
			t.Error("actual owned cleanup did not confirm exit")
		}
	})
	intent, err := service.CreateIntent(ctx, human(), app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "codex", Task: "Reply with exactly PUBLIC_L1_FIRST. Do not use any tool."})
	if err != nil {
		t.Fatal(err)
	}
	// 真实输入成功后，在读取归档元数据之前取消；收据必须保留真实交付。
	launchCtx, cancelMetadata := context.WithCancel(ctx)
	controlled.cancel = cancelMetadata
	confirmCaller := human()
	confirmed, err := service.ConfirmIntent(launchCtx, confirmCaller, intent.Task.ID, app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: "A", ProjectSettingsRevision: project.Settings.Revision})
	if err == nil || confirmed.Receipt.Status != "accepted" || confirmed.Receipt.Error == nil || confirmed.View.Task.Condition == "delivery_unknown" {
		t.Fatalf("post-send metadata failure status=%s condition=%s error_present=%t", confirmed.Receipt.Status, confirmed.View.Task.Condition, err != nil)
	}
	replayed, replayErr := service.ConfirmIntent(ctx, confirmCaller, intent.Task.ID, app.RelayConfirmCommand{PermissionConfirmed: true, PermissionLevel: "A", ProjectSettingsRevision: project.Settings.Revision})
	if replayErr == nil || replayed.Receipt.ID != confirmed.Receipt.ID || replayed.Receipt.Status != "accepted" {
		t.Fatal("metadata failure replay did not preserve the original accepted receipt")
	}
	wait := func() domain.NativeSession {
		t.Helper()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			if err := lifecycle.ObserveOwned(ctx); err != nil {
				t.Fatal("actual owned observation failed; no automatic retry")
			}
			state, err := store.Snapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			session := state.Sessions[confirmed.View.Task.TerminalID]
			if session.Status == "failed" || session.Status == "blocked" || session.StopConfirmed {
				t.Fatal("actual native turn failed or exited before completion")
			}
			if session.Status == "completed" {
				return session
			}
			select {
			case <-ctx.Done():
				t.Fatal("actual turn remained unknown; input was not resent")
			case <-ticker.C:
			}
		}
	}
	session := wait()
	bound := findOwned(t, ownedReader(t, service), intent.Task.ID)
	if bound.SourceStatus != "available" || bound.NativeID != session.NativeID || bound.ArchiveFinalized != "unknown" || bound.ArchiveAppendOnly != "unknown" {
		t.Fatal("actual source failed exact metadata binding or fabricated lifecycle evidence")
	}
	first, err := os.Stat(bound.ArchivePath)
	if err != nil || !first.Mode().IsRegular() || first.Size() == 0 {
		t.Fatal("actual metadata-selected native archive was not written")
	}
	view, err := service.GetTask(ctx, human(), intent.Task.ID)
	if err != nil || view.Terminal.Liveness != domain.RelayLive {
		t.Fatal("real native activity did not establish configured live evidence")
	}
	before, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := lifecycle.ObserveOwned(ctx); err != nil {
		t.Fatal(err)
	}
	after, err := store.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("unchanged real metadata/cache observation created activity or a receipt")
	}
	lastEvent := *view.Terminal.LastEventAt
	time.Sleep(11 * time.Second)
	silent, err := service.GetTask(ctx, human(), intent.Task.ID)
	if err != nil || silent.Terminal.Liveness != domain.RelayUnverifiable || !silent.Terminal.LastEventAt.Equal(lastEvent) {
		t.Fatal("actual silence or cached read refreshed activity inside the explicit test window")
	}
	if _, err := service.Send(ctx, human(), intent.Task.ID, app.RelaySendCommand{Input: "Reply with exactly PUBLIC_L1_SECOND. Do not use any tool."}); err != nil {
		t.Fatal("actual second input failed; not retried")
	}
	_ = wait()
	second, err := os.Stat(bound.ArchivePath)
	if err != nil || !os.SameFile(first, second) || second.Size() <= first.Size() {
		t.Fatal("actual same-session archive did not grow")
	}
	stopped, err := service.Stop(ctx, human(), intent.Task.ID)
	if err != nil || stopped.View.Terminal.ExitedAt == nil || stopped.View.Terminal.Liveness != domain.RelayExited {
		t.Fatal("actual owned process exit was not confirmed")
	}
	closed := findOwned(t, ownedReader(t, service), intent.Task.ID)
	if closed.BindingID != bound.BindingID || closed.ArchivePath != bound.ArchivePath || closed.ExitedAt == nil || closed.ArchiveFinalized != "confirmed" || closed.ArchiveAppendOnly != "unknown" || closed.NativeCapture == nil || closed.NativeCapture.PrefixComplete != "confirmed" || closed.NativeCapture.Finalized != "confirmed" || closed.NativeCapture.Unknown != 0 || closed.NativeCapture.Errors != 0 {
		t.Fatal("actual shutdown acknowledgement and reader/exit chain was not confirmed")
	}
	finalArchive, err := native.OwnedRelayArchive(ctx, session)
	if err != nil || finalArchive.Path != bound.ArchivePath || !finalArchive.Finalized {
		t.Fatal("stopped controller lost its captured and flushed archive source")
	}
	resumed, err := service.Resume(ctx, human(), intent.Task.ID)
	if err != nil {
		t.Fatal("actual original-session resume failed; no replacement created")
	}
	current := findOwned(t, ownedReader(t, service), intent.Task.ID)
	if current.NativeID != bound.NativeID || current.BindingID != bound.BindingID || current.ExitedAt != nil || resumed.View.Terminal.Liveness == domain.RelayExited {
		state, stateErr := store.Snapshot(ctx)
		if stateErr != nil {
			t.Fatal(stateErr)
		}
		selected, selectedErr := native.OwnedRelayArchive(ctx, state.Sessions[bound.TerminalID])
		var sameFile bool
		if selectedErr == nil {
			oldInfo, oldErr := os.Stat(bound.ArchivePath)
			newInfo, newErr := os.Stat(selected.Path)
			sameFile = oldErr == nil && newErr == nil && os.SameFile(oldInfo, newInfo)
		}
		t.Fatalf("actual resume identity: gap=%s source=%s native_equal=%t binding_equal=%t exited=%t liveness=%s provider_error=%t path_equal=%t file_equal=%t", current.Gap, current.SourceStatus, current.NativeID == bound.NativeID, current.BindingID == bound.BindingID, current.ExitedAt != nil, resumed.View.Terminal.Liveness, selectedErr != nil, selected.Path == bound.ArchivePath, sameFile)
	}
	if _, err := service.Send(ctx, human(), intent.Task.ID, app.RelaySendCommand{Input: "Reply with exactly PUBLIC_L1_RESUMED. Do not use any tool."}); err != nil {
		t.Fatal("actual resumed input failed; not retried")
	}
	_ = wait()
	if err := lifecycle.Shutdown(ctx); err != nil {
		t.Fatal("actual shutdown failed")
	}
	shutdownDone = true
	final := findOwned(t, ownedReader(t, service), intent.Task.ID)
	if final.ExitedAt == nil || final.ArchiveFinalized != "confirmed" || final.NativeCapture == nil || final.NativeCapture.FirstGap != "native_resume_raw_events_unavailable" || final.NativeCapture.Unknown == 0 {
		t.Fatal("resumed source hid raw-event coverage gap or lost actual writer shutdown")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.OpenRelayStore(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restarted := app.NewRelayService(reopened, works, db, spaces, driver, nil, nil, options)
	restored := findOwned(t, ownedReader(t, restarted), intent.Task.ID)
	if !reflect.DeepEqual(final, restored) {
		t.Fatal("real SQLite reload changed source/lifecycle evidence")
	}
	readiness, err := registry.SnapshotNative(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range readiness {
		if item.ID == "codex" {
			t.Logf("cli=codex version=%s configured=%s startable=%s actual_turns=3 actual_archive_growth=true actual_stop=true original_resume=true sqlite_reload=true first_stream_prefix=confirmed first_stream_finalized=confirmed resumed_raw_gap=true append_only=unknown native_notifications=%d", native.Version(), item.Configured.Status, item.Startable.Status, len(final.NativeCapture.Events))
		}
	}
}
