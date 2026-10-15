package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func relayRegistrationService(t *testing.T) (app.RelayRuntime, *sqlite.RelayStore, app.RelayIntentCommand) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "public-registration")
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("# Public trusted registration check\n\nRegistration changes are checked without starting a native controller.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", source}, {"-C", source, "add", "README.md"}, {"-C", source, "-c", "user.name=Relay registration check", "-c", "user.email=relay-registration@example.invalid", "-c", "core.hooksPath=", "commit", "-m", "Document trusted registration"}} {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatal("public registration Git creation failed")
		}
	}
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
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(filepath.Join(root, "controller"))
	projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	project, err := projects.RegisterProject(ctx, caller(), domain.RegisterProjectCommand{Name: "Public registration check", Root: source, SpaceID: "public"})
	if err != nil {
		t.Fatal(err)
	}
	project, err = projects.SetProjectSettings(ctx, caller(), project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: domain.ProjectSettings{ExternalModelCLI: "codex", AllowedActions: []string{"observe"}}})
	if err != nil {
		t.Fatal(err)
	}
	works := app.NewWorksService(db, nil, worksscan.Delivery{}, nil, app.WorksServiceOptions{})
	delivery, err := json.Marshal(map[string]any{"works": []any{map[string]any{"platform": "codex", "name": "Public registration check", "path": project.Root, "kind": "folder", "handoff": "Review the public trusted registration policy."}}})
	if err != nil {
		t.Fatal(err)
	}
	album, err := works.ImportWorksDelivery(ctx, caller(), app.WorksDeliveryCommand{Content: delivery})
	if err != nil || len(album.Works) != 1 {
		t.Fatal("public registration work import failed")
	}
	service := app.NewRelayService(store, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), nil, nil, app.RelayServiceOptions{})
	return service, store, app.RelayIntentCommand{WorkID: album.Works[0].ID, ProjectID: project.ID, Agent: "codex", Task: "Review the documented registration policy"}
}

func TestRelayTrustedRegistrationSameServiceAtomicAndReservedIntent(t *testing.T) {
	ctx := context.Background()
	service, store, command := relayRegistrationService(t)
	registrar := service.(app.RelayTrustedAgentRegistrar)
	initial, err := service.ListAgents(ctx, caller())
	if err != nil || initial == nil || len(initial) != 0 {
		t.Fatal("original service did not start with an empty registration")
	}
	when := time.Now().UTC()
	input := []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}, Capabilities: map[string]domain.CapabilityObservation{"start": {Status: "unknown", Reason: "not_run", CheckedAt: &when}}}}
	if err = registrar.RegisterTrustedAgents(ctx, input); err != nil {
		t.Fatal(err)
	}
	active, err := service.ListAgents(ctx, caller())
	if err != nil || !reflect.DeepEqual(active, input) {
		t.Fatal("same service did not expose the complete trusted registration")
	}
	input[0].LaunchCmd[0] = "mutated"
	input[0].Capabilities["start"] = domain.CapabilityObservation{Status: "supported"}
	when = when.Add(time.Hour)
	after, err := service.ListAgents(ctx, caller())
	if err != nil || !reflect.DeepEqual(after, active) {
		t.Fatal("caller mutation changed executable or capability policy")
	}
	after[0].LaunchCmd[0] = "mutated-return"
	*after[0].Capabilities["start"].CheckedAt = time.Time{}
	before, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range [][]domain.AgentSpec{
		{active[0], active[0]},
		{active[0], {Name: "invalid", DriveMode: "terminal"}},
		{{Name: "codex", DriveMode: "interactive", LaunchCmd: []string{"codex"}}},
		{{Name: "codex", DriveMode: "terminal", InputMode: "paste", LaunchCmd: []string{"codex"}}},
		{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex", "bad\x00argument"}}},
	} {
		checkCode(t, registrar.RegisterTrustedAgents(ctx, invalid), apierrors.ValidationFailed)
		current, e := service.ListAgents(ctx, caller())
		if e != nil || !reflect.DeepEqual(current, active) {
			t.Fatal("failed registration partially activated or changed prior policy")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err = registrar.RegisterTrustedAgents(cancelled, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled registration did not fail")
	}
	unchanged, err := store.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(before, unchanged) {
		t.Fatal("registration created a receipt, session, task or controller fact")
	}
	intent, err := service.CreateIntent(ctx, caller(), command)
	if err != nil || intent.Task.State != domain.RelayReady {
		t.Fatal("original service could not create an intent after trusted activation")
	}
	if err = registrar.RegisterTrustedAgents(ctx, active); err != nil {
		t.Fatal("identical trusted reload conflicted with the current intent")
	}
	checkCode(t, registrar.RegisterTrustedAgents(ctx, nil), apierrors.VersionConflict)
	changed := []domain.AgentSpec{{Name: "codex", DriveMode: "terminal", LaunchCmd: []string{"codex"}}}
	checkCode(t, registrar.RegisterTrustedAgents(ctx, changed), apierrors.VersionConflict)
	extended := append(active, domain.AgentSpec{Name: "claude", DriveMode: "structured", LaunchCmd: []string{"claude"}})
	if err = registrar.RegisterTrustedAgents(ctx, extended); err != nil {
		t.Fatal("independent registration could not be added")
	}
	stored, err := store.Snapshot(ctx)
	if err != nil || len(stored.Tasks) != 1 || len(stored.Workspaces) != 0 || len(stored.Sessions) != 0 || len(stored.ControllerEvidence) != 0 {
		t.Fatal("trusted registration or ready intent started execution")
	}
}

func TestRelayTrustedRegistrationConcurrentCompleteSnapshotsAndIntent(t *testing.T) {
	ctx := context.Background()
	service, _, command := relayRegistrationService(t)
	registrar := service.(app.RelayTrustedAgentRegistrar)
	first := []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}
	second := []domain.AgentSpec{{Name: "claude", DriveMode: "structured", LaunchCmd: []string{"claude"}}, {Name: "pi", DriveMode: "structured", LaunchCmd: []string{"pi"}}}
	if err := registrar.RegisterTrustedAgents(ctx, first); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	failures := make(chan error, 16)
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func(writer bool) {
			defer wait.Done()
			for step := 0; step < 20; step++ {
				if writer {
					snapshot := first
					if step%2 == 1 {
						snapshot = second
					}
					if err := registrar.RegisterTrustedAgents(ctx, snapshot); err != nil {
						failures <- err
						return
					}
				} else {
					items, err := service.ListAgents(ctx, caller())
					if err != nil || !reflect.DeepEqual(items, first) && !reflect.DeepEqual(items, second) {
						failures <- errors.New("concurrent reader observed a partial registration")
						return
					}
				}
			}
		}(worker%2 == 0)
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	if err := registrar.RegisterTrustedAgents(ctx, first); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	registration := make(chan error, 1)
	intent := make(chan error, 1)
	go func() { <-start; registration <- registrar.RegisterTrustedAgents(ctx, nil) }()
	go func() { <-start; _, err := service.CreateIntent(ctx, caller(), command); intent <- err }()
	close(start)
	registrationErr, intentErr := <-registration, <-intent
	if registrationErr == nil && intentErr == nil {
		t.Fatal("concurrent removal succeeded after the original intent reserved its Agent")
	}
	if registrationErr != nil {
		checkCode(t, registrationErr, apierrors.VersionConflict)
	}
	if intentErr != nil {
		checkCode(t, intentErr, apierrors.UnsupportedCapability)
	}
	items, err := service.ListAgents(ctx, caller())
	if err != nil || registrationErr == nil && len(items) != 0 || intentErr == nil && !reflect.DeepEqual(items, first) {
		t.Fatal("concurrent operation result differs from the activated registration")
	}
}
