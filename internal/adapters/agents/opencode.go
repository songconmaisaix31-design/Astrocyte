package agents

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
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

// opencodeDenyPermission denies every tool explicitly. It is written into the
// owned config and, as OPENCODE_PERMISSION, merged last by config.ts after
// managed/well-known/account config. `*` alone is not enough because a deep
// merge does not erase an inherited specific read/bash allow, so each named key
// is denied too. Agent-specific permission rules (agent rules take precedence)
// from a managed/well-known source are still not guaranteed overridden: this
// adapter never claims that residual path is suppressed and stays fail-closed
// (Configured=unknown, no model readiness claimed).
const opencodeDenyPermission = `{"*":"deny","read":"deny","edit":"deny","bash":"deny","glob":"deny","grep":"deny","webfetch":"deny","websearch":"deny","task":"deny","skill":"deny","lsp":"deny","question":"deny","external_directory":"deny","doom_loop":"deny"}`

// opencodePermissionRule is the session permission ruleset entry from the
// official 1.18.35 v2 SDK: SessionCreateData/SessionUpdateData accept
// permission: PermissionRuleset where each rule is {permission, pattern, action}.
type opencodePermissionRule struct {
	Permission string `json:"permission"`
	Pattern    string `json:"pattern"`
	Action     string `json:"action"`
}

// opencodeSessionDenyPermission is written into the POST /session body. The
// prompt/tools/processor paths merge the session ruleset AFTER the agent
// permission (Permission.merge(agent.permission, session.permission ?? [])) and
// evaluate with findLast, so this last deny rule overrides any agent-specific
// allow inherited from config or managed/well-known sources. A single wildcard
// rule is sufficient here because findLast is order-based, unlike the config
// OPENCODE_PERMISSION deep-merge that needs per-key denies.
var opencodeSessionDenyPermission = []opencodePermissionRule{{Permission: "*", Pattern: "*", Action: "deny"}}

// opencodeNative drives `opencode serve` over loopback HTTP with Basic auth.
// It is a distinct session driver from the stdio JSON-RPC Native adapter: the
// protocol, process lifecycle and event decoding differ, while the ownership,
// positive-stop and unknown-delivery disciplines are shared by construction.
//
// Selected-context isolation is established by launching the server with
// OPENCODE_CONFIG_DIR pointing at a controller-owned empty config directory, so
// the user's global AGENTS.md, MCP servers, plugins and instructions are not
// loaded; `--pure` + OPENCODE_DISABLE_DEFAULT_PLUGINS suppress external and
// default plugins, and a written `permission: deny` config blocks every tool.
// Native provider authentication remains read by the CLI itself and is never
// copied, exported or logged. This does not redirect the platform-managed
// config directory (e.g. `%ProgramData%\opencode`) or well-known/account remote
// config, which can still declare MCP/plugins; permission deny still blocks
// their tool use, and this adapter never claims those sources are suppressed.
type opencodeNative struct {
	mu            sync.Mutex
	id, stateRoot string
	slots         chan struct{}
	processes     map[string]*opencodeProcess
	capabilities  map[string]domain.CapabilityObservation
	version       string
}

var _ app.NativeAdapter = (*opencodeNative)(nil)

func newOpencodeNative(stateRoot string, slots chan struct{}) *opencodeNative {
	return &opencodeNative{
		id:           "opencode",
		stateRoot:    stateRoot,
		slots:        slots,
		processes:    map[string]*opencodeProcess{},
		capabilities: domain.UnknownNativeCapabilities(),
	}
}

func (n *opencodeNative) ID() string { return n.id }
func (n *opencodeNative) Version() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.version == "" {
		return "native_version_not_observed"
	}
	return n.version
}
func (n *opencodeNative) Capabilities() map[string]domain.CapabilityObservation {
	n.mu.Lock()
	defer n.mu.Unlock()
	result := map[string]domain.CapabilityObservation{}
	for k, v := range n.capabilities {
		result[k] = v
	}
	return result
}
func (n *opencodeNative) observed(capability string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UTC()
	n.capabilities[capability] = domain.CapabilityObservation{Status: "supported", Reason: "owned_native_protocol_succeeded", CheckedAt: &now}
}

// snapshot renders the registry observation for this adapter. Configured stays
// unknown: OpenCode does not report its current model before a paid turn, and
// the selected-text processing path requires that observation, so it remains
// gated rather than invented.
func (n *opencodeNative) snapshot() domain.LocalAgent {
	caps := n.Capabilities()
	item := domain.LocalAgent{
		ID:                      n.id,
		NativeAdapterRegistered: true,
		Capabilities:            caps,
		Configured:              domain.Observation{Status: "unknown", Reason: "native_current_model_not_observed_before_paid_turn"},
		Startable:               domain.Observation{Status: "unknown", Reason: "native_start_not_authorized_or_tested"},
	}
	n.mu.Lock()
	if n.version != "" {
		v := n.version
		item.Version = &v
	}
	n.mu.Unlock()
	if caps["start"].Status == "supported" {
		item.Startable = domain.Observation{Status: "available", Reason: "native_protocol_session_start_observed_not_model_readiness", CheckedAt: caps["start"].CheckedAt}
	}
	return item
}

func (n *opencodeNative) get(s domain.NativeSession) (*opencodeProcess, error) {
	n.mu.Lock()
	p, ok := n.processes[s.ID]
	n.mu.Unlock()
	if !ok {
		return nil, nativeError(apierrors.DeliveryUnknown, "native process ownership is not established in this service instance")
	}
	p.mu.Lock()
	identityMatches := p.nativeID == s.NativeID
	p.mu.Unlock()
	if !identityMatches {
		return nil, nativeError(apierrors.DeliveryUnknown, "native process ownership is not established in this service instance")
	}
	return p, nil
}

func (n *opencodeNative) Discover(ctx context.Context, p domain.LocalProject) ([]domain.NativeSession, error) {
	// No verified OpenCode history-header scope protocol; global session listing
	// would read private history without a registered root.
	return nil, nativeError(apierrors.UnsupportedCapability, "this CLI history header scope protocol has not been verified")
}

func (n *opencodeNative) ReadContext(ctx context.Context, s domain.NativeSession) ([]domain.NativeEvent, error) {
	p, err := n.get(s)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	events, err := p.readMessages(ctx, s.NativeID)
	if err == nil {
		n.observed("read_context")
	}
	return events, err
}

func (n *opencodeNative) Start(ctx context.Context, r domain.NativeRequest) (domain.NativeSession, error) {
	return n.start(ctx, r, false)
}

func (n *opencodeNative) Resume(ctx context.Context, r domain.NativeRequest) (domain.NativeSession, error) {
	if !r.Session.StopConfirmed || r.Session.NativeID == "" {
		return r.Session, nativeError(apierrors.DeliveryUnknown, "resume requires confirmed stop and an original native ID")
	}
	return n.start(ctx, r, true)
}

func (n *opencodeNative) start(ctx context.Context, r domain.NativeRequest, resume bool) (domain.NativeSession, error) {
	s := r.Session
	s.Ownership = "owned"
	s.ID = r.SessionID
	s.ProjectID = r.Project.ID
	s.CLI = n.id
	s.Version = n.Version()
	s.ContextPacket = r.Packet
	s.Mode = "native"
	s.Status = "starting"
	s.StopConfirmed = false
	s.Limitations = []string{"cooperative native CLI; no operating-system read isolation guarantee", "tools denied by permission:deny; user global config isolated via empty OPENCODE_CONFIG_DIR", "platform-managed and well-known remote config are not redirected and may still declare MCP/plugins", "existing external sessions are not taken over", "current model is not reported before a paid turn"}
	n.mu.Lock()
	existing, ok := n.processes[s.ID]
	n.mu.Unlock()
	if ok && !existing.observe().StopConfirmed {
		return s, nativeError(apierrors.VersionConflict, "owned session is still running")
	}
	select {
	case n.slots <- struct{}{}:
	default:
		s.Ownership, s.Status, s.StopConfirmed = "unstarted", "failed", true
		if resume {
			s.Ownership = r.Session.Ownership
		}
		return s, &domain.NativePrelaunchFailure{Cause: nativeError(apierrors.BudgetExhausted, "controller native concurrency limit is four")}
	}
	p, err := n.launch(ctx, r, resume)
	if err != nil {
		<-n.slots
		var uncertain *nativeOwnershipError
		if !errors.As(err, &uncertain) {
			s.Ownership = "unstarted"
			s.StopConfirmed = true
			s.Status = "failed"
			if resume {
				s.Ownership = r.Session.Ownership
			}
			return s, &domain.NativePrelaunchFailure{Cause: err}
		}
		return s, err
	}
	go func() { <-p.done; <-n.slots }()
	n.mu.Lock()
	if len(n.processes) >= 256 {
		for id, old := range n.processes {
			if id != s.ID && old.observe().StopConfirmed {
				delete(n.processes, id)
				break
			}
		}
	}
	n.processes[s.ID] = p
	n.mu.Unlock()

	fail := func(err error) (domain.NativeSession, error) {
		_ = p.stop(context.Background())
		s.Status = "unknown"
		s.StopConfirmed = p.observe().StopConfirmed
		return s, err
	}
	var sessionID string
	if resume {
		// Same-native-ID resume: the original session is loaded by its opaque ID
		// from OpenCode's persisted state and messaging continues against it. A
		// new session would silently change the native identity.
		sessionID = r.Session.NativeID
		if err := p.verifySession(ctx, sessionID); err != nil {
			return fail(err)
		}
	} else {
		sessionID, err = p.createSession(ctx, r.Project.ID)
		if err != nil {
			return fail(err)
		}
		// Fail-closed before any message: confirm the session really carries the
		// deny-all permission ruleset so an inherited agent allow cannot survive.
		if err := p.verifySession(ctx, sessionID); err != nil {
			return fail(err)
		}
	}
	p.mu.Lock()
	p.nativeID = sessionID
	p.mu.Unlock()
	s.NativeID = sessionID
	s.Version = n.Version()
	s.Status = "idle"
	s.UpdatedAt = time.Now().UTC()
	if resume {
		n.observed("resume")
	} else {
		n.observed("start")
	}
	if strings.TrimSpace(r.Command.Message) != "" {
		obs, err := n.Send(ctx, s, packetPrompt(r.Packet, r.Command.Message))
		if err != nil {
			return s, err
		}
		s.Status = obs.Status
	}
	return s, nil
}

func (n *opencodeNative) Send(ctx context.Context, s domain.NativeSession, message string) (domain.NativeObservation, error) {
	if err := ctx.Err(); err != nil {
		return domain.NativeObservation{}, err
	}
	p, err := n.get(s)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	obs := p.observe()
	if obs.StopConfirmed || obs.Status == "running" || obs.Status == "blocked" {
		return obs, nativeError(apierrors.VersionConflict, "native session is stopped, occupied or blocked")
	}
	p.mu.Lock()
	p.status = "running"
	p.mu.Unlock()
	text, err := p.sendMessage(ctx, s.NativeID, message)
	if err != nil {
		return p.observe(), err
	}
	p.mu.Lock()
	if p.status != "failed" {
		p.status = "completed"
	}
	p.mu.Unlock()
	if text != "" {
		p.addEvent("text", text)
	}
	n.observed("send")
	return p.observe(), nil
}

func (n *opencodeNative) Stop(ctx context.Context, s domain.NativeSession) (domain.NativeObservation, error) {
	p, err := n.get(s)
	if err != nil {
		return domain.NativeObservation{Status: "blocked"}, err
	}
	err = p.stop(ctx)
	obs := p.observe()
	if err == nil && obs.StopConfirmed {
		n.observed("stop")
	}
	return obs, err
}

func (n *opencodeNative) Observe(ctx context.Context, s domain.NativeSession) (domain.NativeObservation, error) {
	if err := ctx.Err(); err != nil {
		return domain.NativeObservation{}, err
	}
	p, err := n.get(s)
	if err != nil {
		return domain.NativeObservation{Status: "unknown"}, err
	}
	n.observed("observe")
	return p.observe(), nil
}

// opencodeProcess owns one `opencode serve` subprocess and its HTTP client.
type opencodeProcess struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	cleanup    func()
	done       chan struct{}
	client     *http.Client
	baseURL    string
	username   string
	password   string
	nativeID   string
	status     string
	events     []domain.NativeEvent
	sequence   int
	bytes      int
	truncated  bool
	stopping   bool
	version    string
	model      string
	provider   string
	observedAt time.Time
}

func (p *opencodeProcess) observe() domain.NativeObservation {
	p.mu.Lock()
	defer p.mu.Unlock()
	stopped := false
	select {
	case <-p.done:
		stopped = true
	default:
	}
	return domain.NativeObservation{Status: p.status, Events: append([]domain.NativeEvent{}, p.events...), StopConfirmed: stopped, OutputTruncated: p.truncated}
}

func (p *opencodeProcess) addEvent(kind, text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.truncated {
		return
	}
	if len(p.events) >= 256 || p.bytes+len(text) > 128*1024 {
		p.truncated = true
		p.status = "blocked"
		return
	}
	p.sequence++
	p.bytes += len(text)
	p.events = append(p.events, domain.NativeEvent{Sequence: p.sequence, Kind: kind, Text: text})
}

func (p *opencodeProcess) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.SetBasicAuth(p.username, p.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

// createSession POSTs /session with the deny-all permission ruleset and returns
// the server-assigned session ID, which is the opaque native identity reused for
// resume and read.
func (p *opencodeProcess) createSession(ctx context.Context, title string) (string, error) {
	data, status, err := p.do(ctx, http.MethodPost, "/session", map[string]any{"title": title, "permission": opencodeSessionDenyPermission})
	if err != nil {
		return "", nativeError(apierrors.DeliveryUnknown, "OpenCode session create delivery is unknown")
	}
	if status < 200 || status >= 300 {
		return "", nativeError(apierrors.ProviderUnavailable, "OpenCode server rejected session create")
	}
	var session struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(data, &session) != nil || session.ID == "" {
		return "", nativeError(apierrors.EvidenceMissing, "OpenCode session identity was not reported")
	}
	return session.ID, nil
}

// verifySession loads an existing session by its opaque native ID so a resume
// continues the same session rather than silently creating a new one, and
// fail-closes unless the session carries the deny-all permission ruleset. If the
// server dropped or overrode the deny rule, no message is ever sent.
func (p *opencodeProcess) verifySession(ctx context.Context, sessionID string) error {
	data, status, err := p.do(ctx, http.MethodGet, "/session/"+sessionID, nil)
	if err != nil {
		return nativeError(apierrors.DeliveryUnknown, "OpenCode session resume verification delivery is unknown")
	}
	if status == http.StatusNotFound {
		return nativeError(apierrors.DeliveryUnknown, "original native session is unavailable")
	}
	if status < 200 || status >= 300 {
		return nativeError(apierrors.ProviderUnavailable, "OpenCode server rejected the session resume")
	}
	var session struct {
		ID         string                   `json:"id"`
		Permission []opencodePermissionRule `json:"permission"`
	}
	if json.Unmarshal(data, &session) != nil || session.ID != sessionID {
		return nativeError(apierrors.ScopeDenied, "native session identity differs")
	}
	denyAll := false
	for _, rule := range session.Permission {
		if rule.Permission == "*" && rule.Action == "deny" {
			denyAll = true
		}
	}
	if !denyAll {
		return nativeError(apierrors.ScopeDenied, "session deny-by-default permission is not effective; refusing to send")
	}
	return nil
}

// sendMessage POSTs /session/:id/message and waits for the synchronous reply.
// The returned text is the assistant reply; tool use is denied by the isolated
// permission config, so a text-only reply is the expected shape.
func (p *opencodeProcess) sendMessage(ctx context.Context, sessionID, message string) (string, error) {
	body := map[string]any{"parts": []map[string]any{{"type": "text", "text": message}}}
	data, status, err := p.do(ctx, http.MethodPost, "/session/"+sessionID+"/message", body)
	if err != nil {
		if ctx.Err() != nil {
			return "", nativeError(apierrors.DeliveryUnknown, "OpenCode reply deadline elapsed; do not replay automatically")
		}
		return "", nativeError(apierrors.DeliveryUnknown, "OpenCode message delivery is unknown")
	}
	if status < 200 || status >= 300 {
		return "", nativeError(apierrors.ProviderUnavailable, "OpenCode server rejected the message")
	}
	var reply struct {
		Info struct {
			Role       string `json:"role"`
			ModelID    string `json:"modelID"`
			ProviderID string `json:"providerID"`
		} `json:"info"`
		Parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parts"`
	}
	if json.Unmarshal(data, &reply) != nil {
		return "", nativeError(apierrors.EvidenceMissing, "OpenCode message schema is unavailable")
	}
	if reply.Info.ModelID != "" {
		p.mu.Lock()
		p.model, p.provider, p.observedAt = reply.Info.ModelID, reply.Info.ProviderID, time.Now()
		p.mu.Unlock()
	}
	var text strings.Builder
	for _, part := range reply.Parts {
		if part.Type == "text" {
			text.WriteString(part.Text)
		}
	}
	return text.String(), nil
}

// readMessages GETs /session/:id/message and projects the transcript text into
// bounded native events. It performs no model turn and no state mutation.
func (p *opencodeProcess) readMessages(ctx context.Context, sessionID string) ([]domain.NativeEvent, error) {
	data, status, err := p.do(ctx, http.MethodGet, "/session/"+sessionID+"/message", nil)
	if err != nil {
		return nil, nativeError(apierrors.DeliveryUnknown, "OpenCode context read delivery is unknown")
	}
	if status < 200 || status >= 300 {
		return nil, nativeError(apierrors.ProviderUnavailable, "OpenCode server rejected the context read")
	}
	var messages []struct {
		Info struct {
			Role string `json:"role"`
		} `json:"info"`
		Parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"parts"`
	}
	if json.Unmarshal(data, &messages) != nil {
		return nil, nativeError(apierrors.EvidenceMissing, "OpenCode message list schema is unavailable")
	}
	projection := contextProjection{events: []domain.NativeEvent{}}
	for _, message := range messages {
		for _, part := range message.Parts {
			if part.Type == "text" {
				if err := projection.add(message.Info.Role, part.Text); err != nil {
					return nil, err
				}
			}
		}
	}
	return projection.events, nil
}

func (p *opencodeProcess) stop(ctx context.Context) error {
	select {
	case <-p.done:
		return nil
	default:
	}
	p.mu.Lock()
	p.stopping = true
	p.mu.Unlock()
	// Best-effort graceful abort then dispose; only positive process-tree exit
	// counts as a confirmed stop.
	short, cancel := context.WithTimeout(ctx, time.Second)
	_, _, _ = p.do(short, http.MethodPost, "/session/"+p.nativeID+"/abort", nil)
	_, _, _ = p.do(short, http.MethodPost, "/instance/dispose", nil)
	cancel()
	if err := p.cmd.Cancel(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return nativeError(apierrors.DeliveryUnknown, "native process tree stop is unconfirmed")
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return nativeError(apierrors.DeliveryUnknown, "native process exit is unconfirmed")
	}
}

// launch starts an isolated `opencode serve` and waits for its health endpoint.
func (n *opencodeNative) launch(ctx context.Context, r domain.NativeRequest, resume bool) (*opencodeProcess, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	probe, err := installedCommand(n.id, []string{"--version"})
	if err != nil {
		return nil, err
	}
	probeOutput := &boundedOutput{}
	probe.Stdout = probeOutput
	probe.Stderr = probeOutput
	cleanup, err := startOwnedNative(probe)
	if err != nil {
		return nil, err
	}
	probeDone := make(chan error, 1)
	go func() { probeDone <- probe.Wait(); cleanup() }()
	select {
	case err := <-probeDone:
		if err != nil {
			return nil, nativeError(apierrors.ProviderUnavailable, "native version probe failed")
		}
	case <-probeCtx.Done():
		_ = probe.Cancel()
		<-probeDone
		return nil, nativeError(apierrors.ProviderUnavailable, "native version probe timed out")
	}
	version := versionPattern.FindString(string(probeOutput.bytes))
	if version == "" {
		return nil, nativeError(apierrors.EvidenceMissing, "native version was not reported")
	}
	n.mu.Lock()
	n.version = version
	n.mu.Unlock()

	seconds := r.Command.DeadlineSeconds
	if seconds == 0 {
		seconds = 180
	}
	if seconds < 1 || seconds > 1800 {
		return nil, nativeError(apierrors.BudgetExhausted, "native deadline must be 1 to 1800 seconds")
	}
	if !filepath.IsAbs(n.stateRoot) {
		return nil, nativeError(apierrors.ProviderUnavailable, "controller native state root is not configured")
	}

	port, err := freeLoopbackPort()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(n.stateRoot, r.Project.ID, n.id, r.SessionID)
	configDir := filepath.Join(dir, "config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(configDir, "opencode.json"), []byte(`{"$schema":"https://opencode.ai/config.json","permission":`+opencodeDenyPermission+`}`+"\n"), 0600); err != nil {
		return nil, err
	}
	username, password, err := randomBasicAuth()
	if err != nil {
		return nil, err
	}
	args := []string{"serve", "--hostname", "127.0.0.1", "--port", port, "--pure"}
	cmd, err := installedCommand(n.id, args)
	if err != nil {
		return nil, err
	}
	cmd.Dir = r.Project.Root
	cmd.Env = opencodeIsolationEnv(configDir, username, password)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	processCtx, processCancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	process := &opencodeProcess{
		cmd:      cmd,
		done:     make(chan struct{}),
		client:   &http.Client{Timeout: time.Duration(seconds) * time.Second},
		baseURL:  "http://127.0.0.1:" + port,
		username: username,
		password: password,
		version:  version,
		status:   "idle",
	}
	cleanup, err = startOwnedNative(cmd)
	if err != nil {
		processCancel()
		return nil, err
	}
	process.cleanup = cleanup
	go func() {
		waitErr := cmd.Wait()
		cleanup()
		processCancel()
		process.mu.Lock()
		if waitErr != nil && !process.stopping && process.status != "blocked" && process.status != "completed" {
			process.status = "failed"
			process.addEventLocked("process_exit", fmt.Sprintf("exit_code:%d", exitCode(waitErr)))
			slog.Warn("owned OpenCode process exited", "exit_code", exitCode(waitErr))
		} else if process.status != "completed" && process.status != "failed" && process.status != "blocked" {
			process.status = "stopped"
		}
		process.mu.Unlock()
		close(process.done)
	}()
	go func() {
		select {
		case <-processCtx.Done():
			_ = process.stop(context.Background())
		case <-process.done:
		}
	}()
	// Wait for the loopback server to report healthy.
	if err := waitHealth(ctx, process); err != nil {
		_ = process.stop(context.Background())
		return nil, err
	}
	return process, nil
}

func (p *opencodeProcess) addEventLocked(kind, text string) {
	if p.truncated {
		return
	}
	if len(p.events) >= 256 || p.bytes+len(text) > 128*1024 {
		p.truncated = true
		p.status = "blocked"
		return
	}
	p.sequence++
	p.bytes += len(text)
	p.events = append(p.events, domain.NativeEvent{Sequence: p.sequence, Kind: kind, Text: text})
}

func exitCode(err error) int {
	var nativeExit *exec.ExitError
	if errors.As(err, &nativeExit) {
		return nativeExit.ExitCode()
	}
	return -1
}

func waitHealth(ctx context.Context, p *opencodeProcess) error {
	deadline := time.Now().Add(20 * time.Second)
	for {
		// Bound each probe so a half-open connection cannot stall the whole
		// launch for the client's full timeout.
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		data, status, err := p.do(reqCtx, http.MethodGet, "/global/health", nil)
		cancel()
		if err == nil && status == http.StatusOK {
			var health struct {
				Healthy bool `json:"healthy"`
			}
			if json.Unmarshal(data, &health) == nil && health.Healthy {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return nativeError(apierrors.ProviderUnavailable, "OpenCode server did not become healthy")
		case <-p.done:
			return nativeError(apierrors.ProviderUnavailable, "OpenCode server exited before becoming healthy")
		default:
		}
		if time.Now().After(deadline) {
			return nativeError(apierrors.ProviderUnavailable, "OpenCode server did not become healthy in time")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func freeLoopbackPort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	return port, err
}

func randomBasicAuth() (string, string, error) {
	usernameBytes := make([]byte, 12)
	passwordBytes := make([]byte, 24)
	if _, err := rand.Read(usernameBytes); err != nil {
		return "", "", err
	}
	if _, err := rand.Read(passwordBytes); err != nil {
		return "", "", err
	}
	return hex.EncodeToString(usernameBytes), hex.EncodeToString(passwordBytes), nil
}

// opencodeIsolationEnv keeps the user's provider authentication (read natively
// by the CLI from its own data/state directories) while replacing the config
// directory so the user's global AGENTS.md, MCP servers and plugins are not
// loaded. Inherited OpenCode-specific environment is dropped; unrelated
// environment (including provider credentials the CLI reads itself) is kept.
func opencodeIsolationEnv(configDir, username, password string) []string {
	inherited := os.Environ()
	out := make([]string, 0, len(inherited)+8)
	for _, kv := range inherited {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "OPENCODE_") || key == "XDG_CONFIG_HOME" {
			continue
		}
		out = append(out, kv)
	}
	return append(out,
		"OPENCODE_CONFIG_DIR="+configDir,
		"XDG_CONFIG_HOME="+configDir,
		"OPENCODE_DISABLE_PROJECT_CONFIG=1",
		"OPENCODE_DISABLE_CLAUDE_CODE_PROMPT=1",
		"OPENCODE_PURE=1",
		"OPENCODE_DISABLE_DEFAULT_PLUGINS=1",
		"OPENCODE_PERMISSION="+opencodeDenyPermission,
		"OPENCODE_SERVER_USERNAME="+username,
		"OPENCODE_SERVER_PASSWORD="+password,
	)
}
