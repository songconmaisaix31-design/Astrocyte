package orca

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type Config struct {
	InstallRoot    string
	DataRoot       string
	NodeExecutable string
	ClientEntry    string
	PiEntry        string
	CodexAuthFile  string
	ProtectedRoots []string
}

type Runtime struct{ config Config }

var _ app.OrcaExecutionRuntime = (*Runtime)(nil)

func contains(root, path string) bool {
	if runtime.GOOS == "windows" {
		root, path = strings.ToLower(root), strings.ToLower(path)
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && !filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// 只接受已经安装的明确路径，缺失时不会启动本机 Orca 或选择其他入口。
func New(config Config) (*Runtime, error) {
	if len(config.ProtectedRoots) == 0 || !filepath.IsAbs(config.CodexAuthFile) {
		return nil, errors.New("explicit protected roots and read-only OAuth source are required")
	}
	for _, item := range []*string{&config.InstallRoot, &config.DataRoot, &config.NodeExecutable, &config.ClientEntry, &config.PiEntry} {
		if !filepath.IsAbs(*item) {
			return nil, errors.New("Orca runtime paths must be absolute")
		}
		actual, err := filepath.EvalSymlinks(*item)
		if err != nil {
			return nil, fmt.Errorf("resolve product runtime path: %w", err)
		}
		*item = actual
	}
	if !contains(config.InstallRoot, config.NodeExecutable) || !contains(config.InstallRoot, config.ClientEntry) || !contains(config.InstallRoot, config.PiEntry) || contains(config.InstallRoot, config.DataRoot) || contains(config.DataRoot, config.InstallRoot) {
		return nil, errors.New("product code and private writable roots must be disjoint")
	}
	for _, protected := range config.ProtectedRoots {
		if !filepath.IsAbs(protected) {
			return nil, errors.New("protected root must be absolute")
		}
		protected, err := filepath.EvalSymlinks(protected)
		if err != nil {
			return nil, fmt.Errorf("resolve protected root: %w", err)
		}
		if contains(protected, config.DataRoot) || contains(config.DataRoot, protected) || contains(protected, config.InstallRoot) || contains(config.InstallRoot, protected) {
			return nil, errors.New("product root overlaps a protected Orca or Agent root")
		}
	}
	return &Runtime{config: config}, nil
}

func (r *Runtime) Target() string { return r.config.PiEntry }

func (r *Runtime) ChildEnvironment() []string { return r.environment() }

type bridgeRequest struct {
	Operation  string              `json:"operation"`
	ProjectID  string              `json:"project_id,omitempty"`
	Command    *app.OrcaCommand    `json:"command,omitempty"`
	Invocation *app.OrcaInvocation `json:"invocation,omitempty"`
}

type bridgeResponse struct {
	Scope    *app.OrcaScope       `json:"scope,omitempty"`
	Envelope *app.OrcaEnvelope    `json:"envelope,omitempty"`
	Error    *app.OrcaNativeError `json:"error,omitempty"`
}

func (r *Runtime) environment() []string {
	env := make([]string, 0, 32)
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "ORCA_") || strings.HasPrefix(upper, "PI_") || strings.HasPrefix(upper, "CODEX_") || strings.HasPrefix(upper, "CLAUDE_") || strings.HasPrefix(upper, "ANTHROPIC_") || strings.HasPrefix(upper, "SWA_") || strings.HasPrefix(upper, "NODE_") || strings.HasPrefix(upper, "ASTROCYTE_") || upper == "OPENAI_API_KEY" || upper == "HOME" || upper == "USERPROFILE" || upper == "APPDATA" || upper == "LOCALAPPDATA" || upper == "TEMP" || upper == "TMP" {
			continue
		}
		env = append(env, item)
	}
	return append(env,
		"ORCA_USER_DATA="+r.config.DataRoot,
		"ORCA_USER_DATA_PATH="+r.config.DataRoot,
		"ORCA_TELEMETRY_DISABLED=1",
		"PI_CODING_AGENT_DIR="+filepath.Join(r.config.DataRoot, "pi"),
		"SWA_DATA_DIR="+r.config.DataRoot,
		"SWA_INSTALL_ROOT="+r.config.InstallRoot,
		"SWA_PROTECTED_ROOTS="+strings.Join(r.config.ProtectedRoots, string(filepath.ListSeparator)),
		"SWA_EXECUTABLE="+filepath.Join(r.config.InstallRoot, "swa.exe"),
		"SWA_CODEX_AUTH_FILE="+r.config.CodexAuthFile,
		"SWA_NODE_EXECUTABLE="+r.config.NodeExecutable,
		"SWA_PI_ENTRY="+r.config.PiEntry,
		"SWA_PI_PACKAGE_DIR="+filepath.Join(r.config.InstallRoot, "runtime", "node_modules", "@earendil-works", "pi-coding-agent"),
		"HOME="+filepath.Join(r.config.DataRoot, "home"),
		"USERPROFILE="+filepath.Join(r.config.DataRoot, "home"),
		"APPDATA="+filepath.Join(r.config.DataRoot, "home"),
		"LOCALAPPDATA="+filepath.Join(r.config.DataRoot, "home"),
		"TEMP="+filepath.Join(r.config.DataRoot, "temp"),
		"TMP="+filepath.Join(r.config.DataRoot, "temp"),
		"NODE_OPTIONS=--max-old-space-size=1024",
	)
}

func (r *Runtime) call(ctx context.Context, request bridgeRequest, mutation bool) (bridgeResponse, error) {
	var reply bridgeResponse
	body, err := json.Marshal(request)
	if err != nil {
		return reply, err
	}
	cmd := exec.CommandContext(ctx, r.config.NodeExecutable, "--max-old-space-size=1024", r.config.ClientEntry)
	cmd.Dir = r.config.InstallRoot
	cmd.Env = r.environment()
	cmd.Stdin = bytes.NewReader(body)
	cmd.WaitDelay = 5 * time.Second
	configureHiddenProcess(cmd)
	var stdout, stderr boundedBuffer
	stdout.maximum, stderr.maximum = 16*1024*1024, 256*1024
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		return reply, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "the installed native RuntimeClient could not start", RequiredAction: "inspect_product_installation"}
	}
	waitErr := cmd.Wait()
	if err := json.Unmarshal(stdout.Bytes(), &reply); err != nil {
		code := apierrors.ProviderUnavailable
		if mutation {
			code = apierrors.DeliveryUnknown
		}
		return reply, &apierrors.ServiceError{Code: code, Message: "native client response unavailable; preserve the original request", RequiredAction: "inspect_original_request"}
	}
	if reply.Error != nil {
		return reply, reply.Error
	}
	if waitErr != nil {
		return reply, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "native client exited without a successful transport result", RequiredAction: "inspect_original_request"}
	}
	return reply, nil
}

func (r *Runtime) ResolveScope(ctx context.Context, projectID string, command app.OrcaCommand) (app.OrcaScope, error) {
	response, err := r.call(ctx, bridgeRequest{Operation: "scope", ProjectID: projectID, Command: &command}, false)
	if err != nil {
		return app.OrcaScope{}, err
	}
	if response.Scope == nil {
		return app.OrcaScope{}, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "native controller scope is unavailable", RequiredAction: "inspect_controller_and_original_request"}
	}
	return *response.Scope, nil
}

func (r *Runtime) Execute(ctx context.Context, invocation app.OrcaInvocation) (app.OrcaEnvelope, error) {
	mutations := map[string]bool{"controller.open": true, "controller.close": true, "task.create": true, "task.complete": true, "worker.start": true, "worker.send": true, "worker.stop": true, "worker.release": true}
	response, err := r.call(ctx, bridgeRequest{Operation: "execute", Invocation: &invocation}, mutations[invocation.Command.Action])
	if err != nil {
		if response.Envelope != nil {
			return *response.Envelope, err
		}
		return app.OrcaEnvelope{}, err
	}
	if response.Envelope == nil {
		return app.OrcaEnvelope{}, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "native client omitted the original operation result", RequiredAction: "inspect_original_request"}
	}
	return *response.Envelope, nil
}

type boundedBuffer struct {
	bytes.Buffer
	maximum int
}

func (b *boundedBuffer) Write(value []byte) (int, error) {
	if len(value) > b.maximum-b.Len() {
		return 0, errors.New("native client output exceeds the declared bound")
	}
	return b.Buffer.Write(value)
}
