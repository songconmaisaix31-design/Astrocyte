package agents

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// LaunchPolicy is supplied by a trusted assembly, never an HTTP callback.
// Default nil refuses free workspace tools until their execution scope is decided.
type RelayTerminalOptions struct {
	LaunchPolicy func(context.Context, app.RelayLaunch) error
}

type RelayTerminalDriver struct {
	mu        sync.Mutex
	processes map[string]*relayTerminalProcess
	options   RelayTerminalOptions
}

type relayTerminalProcess struct {
	mu       sync.Mutex
	writeMu  sync.Mutex
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	done     chan struct{}
	sequence int
	lastAt   time.Time
	exited   bool
}

var _ app.RelayDriver = (*RelayTerminalDriver)(nil)

func NewRelayTerminalDriver(options RelayTerminalOptions) *RelayTerminalDriver {
	return &RelayTerminalDriver{processes: map[string]*relayTerminalProcess{}, options: options}
}

// Output is discarded at the boundary. Only the fact and time of actual stdout
// bytes are retained; terminal text cannot enter a view, report or cognition event.
func (p *relayTerminalProcess) Write(b []byte) (int, error) {
	if len(b) > 0 {
		p.mu.Lock()
		p.sequence++
		p.lastAt = time.Now().UTC()
		p.mu.Unlock()
	}
	return len(b), nil
}

func (p *relayTerminalProcess) observation() app.RelayObservation {
	p.mu.Lock()
	defer p.mu.Unlock()
	obs := app.RelayObservation{ExitConfirmed: p.exited, Events: []domain.RelayEvent{}}
	if p.sequence > 0 {
		obs.Events = append(obs.Events, domain.RelayEvent{Kind: "cli_activity", Source: "terminal", Sequence: p.sequence, TS: p.lastAt})
	}
	return obs
}

func (d *RelayTerminalDriver) owned(id string) (*relayTerminalProcess, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.processes[id]
	if !ok {
		return nil, nativeError(apierrors.DeliveryUnknown, "terminal process ownership is not established after restart")
	}
	return p, nil
}

func (d *RelayTerminalDriver) Launch(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	unstarted := func(e error) (app.RelayObservation, error) {
		return app.RelayObservation{}, &domain.NativePrelaunchFailure{Cause: e}
	}
	if err := ctx.Err(); err != nil {
		return unstarted(err)
	}
	if d.options.LaunchPolicy == nil {
		return unstarted(nativeError(apierrors.UnsupportedCapability, "terminal workspace execution scope is not configured"))
	}
	if err := d.options.LaunchPolicy(ctx, r); err != nil {
		return unstarted(err)
	}
	if !r.Workspace.Prepared || !filepath.IsAbs(r.Workspace.Root) || r.Terminal.ID == "" || len(r.Spec.LaunchCmd) == 0 {
		return unstarted(nativeError(apierrors.ValidationFailed, "trusted terminal registration and prepared workspace are required"))
	}
	argv := r.Spec.LaunchCmd
	entry, err := exec.LookPath(argv[0])
	if err != nil {
		return unstarted(nativeError(apierrors.ProviderUnavailable, "registered terminal CLI is not available"))
	}
	// Trusted data can name node + a wrapper script explicitly. Never interpolate
	// a shell command or infer client-specific launch switches in dispatch code.
	if ext := strings.ToLower(filepath.Ext(entry)); ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		return unstarted(nativeError(apierrors.UnsupportedCapability, "registration must name a direct executable entrypoint"))
	}
	cmd := exec.Command(entry, argv[1:]...)
	cmd.Dir = r.Workspace.Root
	cmd.Env = os.Environ()
	cmd.WaitDelay = time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return unstarted(err)
	}
	p := &relayTerminalProcess{cmd: cmd, stdin: stdin, done: make(chan struct{})}
	cmd.Stdout = p
	cmd.Stderr = io.Discard
	d.mu.Lock()
	if existing, ok := d.processes[r.Terminal.ID]; ok && !existing.observation().ExitConfirmed {
		d.mu.Unlock()
		_ = stdin.Close()
		return unstarted(nativeError(apierrors.VersionConflict, "terminal identity already owns a process"))
	}
	cleanup, err := startOwnedNative(cmd)
	if err != nil {
		d.mu.Unlock()
		_ = stdin.Close()
		var unknown *nativeOwnershipError
		if errors.As(err, &unknown) {
			return app.RelayObservation{}, nativeError(apierrors.DeliveryUnknown, "terminal process-tree ownership is uncertain")
		}
		return unstarted(err)
	}
	d.processes[r.Terminal.ID] = p
	d.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		cleanup()
		_ = stdin.Close()
		p.mu.Lock()
		p.exited = true
		p.mu.Unlock()
		close(p.done)
	}()
	if r.Input != "" {
		return d.Send(ctx, r, r.Input)
	}
	return p.observation(), nil
}

func (d *RelayTerminalDriver) Send(ctx context.Context, r app.RelayLaunch, input string) (app.RelayObservation, error) {
	p, err := d.owned(r.Terminal.ID)
	if err != nil {
		return app.RelayObservation{}, err
	}
	if err = ctx.Err(); err != nil {
		return p.observation(), err
	}
	if p.observation().ExitConfirmed {
		return p.observation(), nativeError(apierrors.VersionConflict, "terminal process has exited")
	}
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_, err = io.WriteString(p.stdin, input+"\n")
	if err != nil {
		return p.observation(), nativeError(apierrors.DeliveryUnknown, "terminal input delivery is unknown; do not resend")
	}
	return p.observation(), nil
}

func (d *RelayTerminalDriver) Read(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	p, err := d.owned(r.Terminal.ID)
	if err != nil {
		return app.RelayObservation{}, err
	}
	if err = ctx.Err(); err != nil {
		return p.observation(), err
	}
	return p.observation(), nil
}

func (d *RelayTerminalDriver) Stop(ctx context.Context, r app.RelayLaunch) (app.RelayObservation, error) {
	p, err := d.owned(r.Terminal.ID)
	if err != nil {
		return app.RelayObservation{}, err
	}
	if p.observation().ExitConfirmed {
		return p.observation(), nil
	}
	if err = p.cmd.Cancel(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return p.observation(), nativeError(apierrors.DeliveryUnknown, "terminal process-tree stop is unconfirmed")
	}
	select {
	case <-p.done:
		return p.observation(), nil
	case <-ctx.Done():
		return p.observation(), nativeError(apierrors.DeliveryUnknown, "terminal process exit was not confirmed")
	}
}

func (d *RelayTerminalDriver) Resume(context.Context, app.RelayLaunch) (app.RelayObservation, error) {
	return app.RelayObservation{}, nativeError(apierrors.UnsupportedCapability, "terminal-only registration has no verified native resume operation")
}
