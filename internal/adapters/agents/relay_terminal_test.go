package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// Public controlled CLI prevalidation only. This executable does not invoke a
// provider or pretend to be Codex/Claude or an authenticated coding Agent.
func TestRelayCLIHelperProcess(t *testing.T) {
	if os.Getenv("ASTROCYTE_RELAY_TEST_HELPER") != "1" {
		return
	}
	mode := os.Getenv("ASTROCYTE_RELAY_TEST_MODE")
	if mode == "blocked_input" {
		for {
			time.Sleep(time.Second)
		} // deliberately never consumes stdin
	}
	if mode == "eof" {
		input, e := io.ReadAll(os.Stdin)
		if e != nil {
			os.Exit(2)
		}
		if e = os.WriteFile("public-eof-input.txt", input, 0600); e != nil {
			os.Exit(3)
		}
		_, _ = fmt.Fprintln(os.Stdout, "public input fully consumed at EOF")
		for {
			time.Sleep(time.Second)
		} // remains owned until explicit Stop
	}
	if mode == "exit" {
		os.Exit(0)
	}
	if mode == "active" {
		_, _ = fmt.Fprintln(os.Stdout, "PRIVATE_TERMINAL_TEST_SENTINEL")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "__exit__" {
			os.Exit(0)
		}
		if scanner.Text() == "__write__" {
			_ = os.WriteFile("public-helper-result.txt", []byte("controlled workspace edit"), 0600)
			_, _ = fmt.Fprintln(os.Stdout, "actual output event")
		}
	}
	os.Exit(0)
}

func relayTerminalTestLaunch(t *testing.T, mode string) app.RelayLaunch {
	t.Helper()
	t.Setenv("ASTROCYTE_RELAY_TEST_HELPER", "1")
	t.Setenv("ASTROCYTE_RELAY_TEST_MODE", mode)
	executable, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	return app.RelayLaunch{Spec: domain.AgentSpec{Name: "public-controlled-cli", LaunchCmd: []string{executable, "-test.run=^TestRelayCLIHelperProcess$"}, DriveMode: "terminal"}, Workspace: domain.RelayWorkspace{Root: t.TempDir(), Prepared: true}, Terminal: domain.RelayTerminal{ID: "stable-terminal"}}
}

func relayTerminalTestPolicy(_ context.Context, r app.RelayLaunch) error {
	if r.Spec.Name != "public-controlled-cli" || len(r.Spec.LaunchCmd) != 2 || r.Spec.LaunchCmd[1] != "-test.run=^TestRelayCLIHelperProcess$" {
		return nativeError(apierrors.ScopeDenied, "test permits only the controlled public CLI")
	}
	return nil
}

func waitRelayObservation(t *testing.T, d *RelayTerminalDriver, r app.RelayLaunch, condition func(app.RelayObservation) bool) app.RelayObservation {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		obs, e := d.Read(context.Background(), r)
		if e != nil {
			t.Fatal(e)
		}
		if condition(obs) {
			return obs
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("real process observation deadline elapsed")
	return app.RelayObservation{}
}

func TestRelayTerminalActualActiveSilentAndExitEvidence(t *testing.T) {
	for _, mode := range []string{"active", "silent", "exit"} {
		t.Run(mode, func(t *testing.T) {
			r := relayTerminalTestLaunch(t, mode)
			d := NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, e := d.Launch(ctx, r); e != nil {
				t.Fatal(e)
			}
			t.Cleanup(func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = d.Stop(cleanup, r)
			})
			var obs app.RelayObservation
			switch mode {
			case "active":
				obs = waitRelayObservation(t, d, r, func(v app.RelayObservation) bool { return len(v.Events) > 0 })
				encoded, _ := json.Marshal(obs.Events)
				if strings.Contains(string(encoded), "PRIVATE_TERMINAL_TEST_SENTINEL") {
					t.Fatal("terminal bytes leaked into events")
				}
				last := obs.Events[0].TS
				if domain.RelayLiveness(&last, nil, time.Now(), time.Second) != "live" {
					t.Fatal("real active process not live")
				}
				if _, e := d.Send(ctx, r, "__exit__"); e != nil {
					t.Fatal(e)
				}
				obs = waitRelayObservation(t, d, r, func(v app.RelayObservation) bool { return v.ExitConfirmed })
			case "silent":
				time.Sleep(100 * time.Millisecond)
				obs, _ = d.Read(ctx, r)
				if len(obs.Events) != 0 || obs.ExitConfirmed || domain.RelayLiveness(nil, nil, time.Now(), time.Second) != "unverifiable" {
					t.Fatal("silence interpreted as process exit or liveness")
				}
				obs, e := d.Stop(ctx, r)
				if e != nil || !obs.ExitConfirmed {
					t.Fatal("actual process stop unconfirmed")
				}
			case "exit":
				obs = waitRelayObservation(t, d, r, func(v app.RelayObservation) bool { return v.ExitConfirmed })
			}
			if mode != "silent" && !obs.ExitConfirmed {
				t.Fatal("actual exit not observed")
			}
		})
	}
}

func TestRelayTerminalRegistrationDataAndUnknownOwnership(t *testing.T) {
	r := relayTerminalTestLaunch(t, "silent")
	d := NewRelayTerminalDriver(RelayTerminalOptions{})
	_, e := d.Launch(context.Background(), r)
	var pre *domain.NativePrelaunchFailure
	if !errors.As(e, &pre) {
		t.Fatal("unconfigured terminal execution not refused before start")
	}
	d = NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, e = d.Launch(ctx, r); e != nil {
		t.Fatal(e)
	}
	defer d.Stop(ctx, r)
	_, e = d.Launch(ctx, r)
	if !errors.As(e, &pre) {
		t.Fatal("same terminal launched two processes")
	}
	if _, e = d.Send(ctx, r, "__write__"); e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(5 * time.Second)
	exists := false
	for time.Now().Before(deadline) {
		if _, e = os.Stat(filepath.Join(r.Workspace.Root, "public-helper-result.txt")); e == nil {
			exists = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !exists {
		t.Fatal("configured public CLI did not execute in the neutral workspace")
	}
	restarted := NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
	_, e = restarted.Read(ctx, r)
	var known *apierrors.ServiceError
	if !errors.As(e, &known) || known.Code != apierrors.DeliveryUnknown {
		t.Fatal("restart inferred ownership from terminal ID or PID")
	}
	if _, e = d.Resume(ctx, r); e == nil {
		t.Fatal("unverified terminal resume claimed success")
	}
}

func TestRelayTerminalEOFCompleteFirstInputAndNoSecondWrite(t *testing.T) {
	r := relayTerminalTestLaunch(t, "eof")
	r.Spec.InputMode = "stdin_eof"
	r.Input = "HANDOFF_MANUAL:\nfull public manual\nlast manual line\n\nTASK:\npublic explicit task"
	d := NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, e := d.Launch(ctx, r); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = d.Stop(cleanup, r)
	})
	waitRelayObservation(t, d, r, func(obs app.RelayObservation) bool { return len(obs.Events) > 0 })
	path := filepath.Join(r.Workspace.Root, "public-eof-input.txt")
	data, e := os.ReadFile(path)
	if e != nil || string(data) != r.Input {
		t.Fatal("EOF did not deliver exactly one complete first input")
	}
	r.Spec.InputMode = "line" // later registration cannot reopen the owned stdin
	for i := 0; i < 2; i++ {
		_, e = d.Send(ctx, r, "must never be written")
		var known *apierrors.ServiceError
		if !errors.As(e, &known) || known.Code != apierrors.UnsupportedCapability {
			t.Fatalf("EOF Send must be unsupported: %v", e)
		}
	}
	current, e := os.ReadFile(path)
	if e != nil || string(current) != string(data) {
		t.Fatal("follow-up send changed EOF input")
	}
	obs, e := d.Stop(ctx, r)
	if e != nil || !obs.ExitConfirmed {
		t.Fatal("EOF owned process did not produce a real stop/exit")
	}
}

func TestRelayTerminalUnknownInputModeCannotStartAndEOFDoesNotGrantTools(t *testing.T) {
	r := relayTerminalTestLaunch(t, "eof")
	d := NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
	r.Spec.InputMode = "unregistered_framing"
	_, e := d.Launch(context.Background(), r)
	var pre *domain.NativePrelaunchFailure
	var known *apierrors.ServiceError
	if !errors.As(e, &pre) || !errors.As(e, &known) || known.Code != apierrors.UnsupportedCapability || len(d.processes) != 0 {
		t.Fatal("unknown framing started a process")
	}
	r.Spec.InputMode = "stdin_eof"
	d = NewRelayTerminalDriver(RelayTerminalOptions{})
	_, e = d.Launch(context.Background(), r)
	if !errors.As(e, &pre) || len(d.processes) != 0 {
		t.Fatal("EOF registration bypassed nil tools policy")
	}
}

func TestRelayTerminalEOFUnknownWriteNeverLaunchesSecondProcess(t *testing.T) {
	r := relayTerminalTestLaunch(t, "blocked_input")
	r.Spec.InputMode = "stdin_eof"
	r.Input = strings.Repeat("public blocked stdin\n", 1<<16)
	d := NewRelayTerminalDriver(RelayTerminalOptions{LaunchPolicy: relayTerminalTestPolicy})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	t.Cleanup(func() {
		cleanup, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopCancel()
		_, _ = d.Stop(cleanup, r)
	})
	_, e := d.Launch(ctx, r)
	var known *apierrors.ServiceError
	if !errors.As(e, &known) || known.Code != apierrors.DeliveryUnknown {
		t.Fatalf("blocked EOF write must remain unknown: %v", e)
	}
	cleanup, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_, e = d.Launch(cleanup, r)
	var pre *domain.NativePrelaunchFailure
	if !errors.As(e, &pre) || !errors.As(e, &known) || known.Code != apierrors.VersionConflict || len(d.processes) != 1 {
		t.Fatal("unknown EOF delivery launched a second writer")
	}
	_, e = d.Send(cleanup, r, "do not resend")
	if !errors.As(e, &known) || known.Code != apierrors.UnsupportedCapability {
		t.Fatal("EOF unknown delivery permitted a second write")
	}
	obs, e := d.Stop(cleanup, r)
	if e != nil || !obs.ExitConfirmed {
		t.Fatal("unknown EOF process could not be contained")
	}
}
