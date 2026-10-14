package agents

import (
	"context"
	"errors"
	"github.com/google/uuid"
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
	mu        sync.Mutex
	writeMu   sync.Mutex
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	inputMode string
	done      chan struct{}
	sequence  int
	lastAt    time.Time
	exited    bool
	evidence  domain.RelayControllerEvidence
}

var _ app.RelayDriver = (*RelayTerminalDriver)(nil)

func NewRelayTerminalDriver(options RelayTerminalOptions) *RelayTerminalDriver {
	return &RelayTerminalDriver{processes: map[string]*relayTerminalProcess{}, options: options}
}

func (d *RelayTerminalDriver) CurrentControllerEvidence(ctx context.Context, r app.RelayLaunch) (domain.RelayControllerEvidence, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p, err := d.owned(r.Terminal.ID)
	if err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.evidence.ControllerID == "" || p.evidence.StartupOperationID == "" {
		return domain.RelayControllerEvidence{}, nativeError(apierrors.DeliveryUnknown, "original terminal controller operation is unrecorded")
	}
	if p.evidence.ControllerID != r.ControllerID || p.evidence.Provider != r.Spec.Name || !sameRelayMetadataCWD(p.cmd.Dir, r.Workspace.Root) {
		return domain.RelayControllerEvidence{}, nativeError(apierrors.ScopeDenied, "current terminal controller differs from the requested binding")
	}
	evidence := p.evidence
	evidence.ExitConfirmed = p.exited
	evidence.Facts = append([]domain.RelayDriverFact{}, evidence.Facts...)
	return evidence, nil
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
	if p.evidence.ControllerID != "" {
		evidence := p.evidence
		evidence.Facts = append([]domain.RelayDriverFact{}, p.evidence.Facts...)
		obs.Evidence = &evidence
	}
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
	mode := r.Spec.InputMode
	if mode == "" {
		mode = "line"
	}
	if mode != "line" && mode != "stdin_eof" {
		return unstarted(nativeError(apierrors.UnsupportedCapability, "registered terminal input mode is not supported"))
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
	p := &relayTerminalProcess{cmd: cmd, stdin: stdin, inputMode: mode, done: make(chan struct{})}
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
	p.mu.Lock()
	p.evidence = domain.RelayControllerEvidence{ControllerID: uuid.NewString(), StartupOperationID: r.OperationID, Provider: r.Spec.Name, WorkspaceRoot: r.Workspace.Root, Relation: "new", Facts: []domain.RelayDriverFact{}}
	p.fact("input_channel_open", "owned_stdin_pipe", r.OperationID)
	p.mu.Unlock()
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
	if r.Input != "" || mode == "stdin_eof" {
		return p.submit(ctx, r.Input, mode == "stdin_eof", r.OperationID)
	}
	return p.observation(), nil
}

func (d *RelayTerminalDriver) Send(ctx context.Context, r app.RelayLaunch, input string) (app.RelayObservation, error) {
	p, err := d.owned(r.Terminal.ID)
	if err != nil {
		return app.RelayObservation{}, err
	}
	// Framing belongs to the owned process, not a mutable later registration.
	// A one-shot input is already closed and can never receive a second write.
	if p.inputMode == "stdin_eof" {
		return p.observation(), nativeError(apierrors.UnsupportedCapability, "stdin EOF mode does not support follow-up input")
	}
	return p.submit(ctx, input, false, r.OperationID)
}

func (p *relayTerminalProcess) fact(kind, source, operation string) {
	p.evidence.Facts = append(p.evidence.Facts, domain.RelayDriverFact{OperationID: operation, Kind: kind, Status: "confirmed", Source: source, Sequence: int64(len(p.evidence.Facts) + 1), TS: time.Now().UTC()})
}

func (p *relayTerminalProcess) submit(ctx context.Context, input string, closeInput bool, operation string) (app.RelayObservation, error) {
	var err error
	if err = ctx.Err(); err != nil {
		return p.observation(), nativeError(apierrors.DeliveryUnknown, "owned terminal input deadline elapsed; do not resend")
	}
	if p.observation().ExitConfirmed {
		return p.observation(), nativeError(apierrors.VersionConflict, "terminal process has exited")
	}
	written := make(chan error, 1)
	go func() {
		p.writeMu.Lock()
		defer p.writeMu.Unlock()
		data := input
		if !closeInput {
			data += "\n"
		}
		n, e := io.WriteString(p.stdin, data)
		if e == nil && n != len(data) {
			e = io.ErrShortWrite
		}
		if e == nil {
			p.mu.Lock()
			p.fact("input_written", "stdin_pipe", operation)
			p.mu.Unlock()
		}
		if closeInput {
			closeErr := p.stdin.Close()
			if closeErr == nil {
				p.mu.Lock()
				p.fact("input_closed", "stdin_eof", operation)
				p.mu.Unlock()
			}
			if e == nil {
				e = closeErr
			}
		}
		if e == nil {
			p.mu.Lock()
			p.evidence.Facts = append(p.evidence.Facts, domain.RelayDriverFact{OperationID: operation, Kind: "input_consumption", Status: "unknown", Source: "stdin_pipe", Reason: "client_consumption_not_acknowledged", Sequence: int64(len(p.evidence.Facts) + 1), TS: time.Now().UTC()})
			p.mu.Unlock()
		}
		written <- e
	}()
	select {
	case err = <-written:
		if err != nil {
			return p.observation(), nativeError(apierrors.DeliveryUnknown, "terminal input delivery is unknown; do not resend")
		}
		return p.observation(), nil
	case <-ctx.Done():
		return p.observation(), nativeError(apierrors.DeliveryUnknown, "terminal input deadline elapsed; do not resend")
	}
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
