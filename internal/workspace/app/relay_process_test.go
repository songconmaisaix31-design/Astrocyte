package app_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayControlledCLIProcess(t *testing.T) {
	if os.Getenv("ASTROCYTE_RELAY_CONTROLLED_CLI") != "1" {
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, "public cli active")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "__edit__" {
			_ = os.WriteFile("public.txt", []byte("actual controlled CLI change\n"), 0600)
		}
	}
	os.Exit(0)
}

func TestRelayControlledCLIFullStateAndGitObservation(t *testing.T) {
	f := newRelayFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	t.Setenv("ASTROCYTE_RELAY_CONTROLLED_CLI", "1")
	source := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", source}, args...)...)
		if e := c.Run(); e != nil {
			t.Fatalf("controlled Git setup failed: %v", e)
		}
	}
	git("init", "-q")
	if e := os.WriteFile(filepath.Join(source, "public.txt"), []byte("public baseline\n"), 0600); e != nil {
		t.Fatal(e)
	}
	git("add", "public.txt")
	git("-c", "user.name=Public Test", "-c", "user.email=public-test@example.invalid", "commit", "-qm", "baseline")
	f.works.work.Root = source
	f.projects.project.Root = source
	f.projects.project.Settings.ExternalModelCLI = "public-controlled-cli"
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	f.options.Agents = []domain.AgentSpec{{Name: "public-controlled-cli", DriveMode: "terminal", LaunchCmd: []string{executable, "-test.run=^TestRelayControlledCLIProcess$"}}}
	f.options.Clock = time.Now
	workspaces := relayworktree.New(filepath.Join(f.root, "worktrees"), relayworktree.Options{AllowCommittedGit: true})
	terminal := agents.NewRelayTerminalDriver(agents.RelayTerminalOptions{LaunchPolicy: func(_ context.Context, r app.RelayLaunch) error {
		if r.Spec.Name != "public-controlled-cli" || r.Spec.LaunchCmd[0] != executable {
			return fmt.Errorf("test permits only its public controlled executable")
		}
		return nil
	}})
	f.service = app.NewRelayService(f.store, f.works, f.projects, workspaces, nil, terminal, f.sink, f.options)
	id := f.intent()
	result, e := f.service.ConfirmIntent(ctx, caller(), id, confirm())
	if e != nil {
		t.Fatal(e)
	}
	lifecycle := f.service.(app.RelayLifecycle)
	t.Cleanup(func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = lifecycle.Shutdown(shutdown)
	})
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if e = lifecycle.ObserveOwned(ctx); e != nil {
			v, _ := f.service.GetTask(context.Background(), caller(), id)
			t.Fatalf("observer failed: %v; context=%v, state=%s, sequence=%d", e, ctx.Err(), v.Task.State, v.Terminal.LastSequence)
		}
		result.View, e = f.service.GetTask(ctx, caller(), id)
		if e != nil {
			t.Fatal(e)
		}
		if result.View.Task.State == "working" && result.View.Terminal.Liveness == "live" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if result.View.Task.State != "working" || result.View.Terminal.Liveness != "live" {
		t.Fatal("actual CLI activity not visible in persisted task")
	}
	_, e = f.service.Send(ctx, caller(), id, app.RelaySendCommand{Input: "__edit__"})
	if e != nil {
		t.Fatal(e)
	}
	sawGit := false
	for time.Now().Before(deadline) {
		if e = lifecycle.ObserveOwned(ctx); e != nil {
			t.Fatal(e)
		}
		events, e := f.service.ListEvents(ctx, caller(), id)
		if e != nil {
			t.Fatal(e)
		}
		for _, event := range events {
			if event.Kind == "git_activity" {
				sawGit = true
			}
		}
		if sawGit {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !sawGit {
		t.Fatal("actual CLI Git change did not reach the derived sink")
	}
	v, e := f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Task.State != "working" || v.EventPublication.Status != "published" {
		t.Fatal("Git completion inferred or event publication missing")
	}
	data, e := os.ReadFile(filepath.Join(v.Workspace.Root, "public.txt"))
	if e != nil || string(data) != "actual controlled CLI change\n" {
		t.Fatal("CLI did not change its neutral workspace")
	}
	data, e = os.ReadFile(filepath.Join(source, "public.txt"))
	if e != nil || string(data) != "public baseline\n" {
		t.Fatal("CLI changed original source tree")
	}
	if e = lifecycle.Shutdown(ctx); e != nil {
		t.Fatal(e)
	}
	v, e = f.service.GetTask(ctx, caller(), id)
	if e != nil || v.Terminal.Liveness != "exited" || v.Task.Condition != "terminal_exited" || v.Task.State != "working" {
		t.Fatal("actual process exit silently completed the task")
	}
}
