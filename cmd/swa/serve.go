package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	orcaadapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/orca"
)

func serve(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("swa serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	data := flags.String("data-dir", "", "明确的私有数据目录")
	port := flags.Int("port", 0, "明确的回环 HTTP 端口")
	asJSON := flags.Bool("json", false, "输出实际 readiness")
	auth := flags.String("codex-auth-file", os.Getenv("SWA_CODEX_AUTH_FILE"), "已配置的只读 Codex OAuth 文件")
	protected := flags.String("protected-roots", os.Getenv("SWA_PROTECTED_ROOTS"), "本机 Orca 及账户受保护目录列表")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || !filepath.IsAbs(*data) || !filepath.IsAbs(*auth) || *port < 1 || *port > 65535 || *protected == "" {
		return report(stderr, "validation_failed", "需要明确的数据目录、端口、只读认证来源及受保护目录", "correct_serve_configuration", 2)
	}
	if err := configureCurrentProcess(); err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_supported_platform", 4)
	}
	probe, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(*port))
	if err != nil {
		return report(stderr, "address_in_use", err.Error(), "choose_unused_explicit_port", 5)
	}
	if err := probe.Close(); err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_explicit_port", 4)
	}
	executable, err := os.Executable()
	if err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_installation", 4)
	}
	install, err := (agents.ProjectFiles{}).CanonicalRoot(filepath.Dir(executable))
	if err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_installation", 4)
	}
	roots := filepath.SplitList(*protected)
	canonicalData, err := privateDataDirectory(*data, install, roots)
	if err != nil {
		return report(stderr, "scope_denied", err.Error(), "choose_private_product_roots", 3)
	}
	for _, name := range []string{"home", "temp"} {
		if err := os.MkdirAll(filepath.Join(canonicalData, name), 0o700); err != nil {
			return report(stderr, "storage_failed", err.Error(), "inspect_private_data_root", 7)
		}
	}
	native, err := orcaadapter.New(orcaadapter.Config{InstallRoot: install, DataRoot: canonicalData, NodeExecutable: filepath.Join(install, "node", "node.exe"), ClientEntry: filepath.Join(install, "runtime", "client.cjs"), PiEntry: filepath.Join(install, "runtime", "pi-entry.mjs"), CodexAuthFile: *auth, ProtectedRoots: roots})
	if err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_installed_runtime", 4)
	}
	samples, err := resourceAdmission(ctx)
	_ = json.NewEncoder(stderr).Encode(map[string]any{"type": "resource_admission", "samples": samples, "accepted": err == nil})
	if err != nil {
		return report(stderr, "resource_refused", err.Error(), "inspect_current_resources", 5)
	}
	baseEnv := native.ChildEnvironment()
	baseEnv = append(baseEnv, "ASTROCYTE_PORT="+strconv.Itoa(*port), "SWA_LIFECYCLE_STDIN=1", "GOMAXPROCS=4", "GOMEMLIMIT=512MiB")
	node := exec.Command(filepath.Join(install, "node", "node.exe"), "--max-old-space-size=1024", filepath.Join(install, "runtime", "host.cjs"))
	node.Env, node.Dir, node.Stderr = baseEnv, install, stderr
	configureServiceChild(node)
	nodeOwner, err := node.StdinPipe()
	if err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_owned_service_pipes", 4)
	}
	nodeOutput, err := node.StdoutPipe()
	if err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_owned_service_pipes", 4)
	}
	if err := node.Start(); err != nil {
		return report(stderr, "provider_unavailable", err.Error(), "inspect_installed_node", 4)
	}
	nodeResult := make(chan error, 1)
	go func() { nodeResult <- node.Wait() }()
	defer nodeOwner.Close()
	ready := make(chan map[string]any, 1)
	readyError := make(chan error, 1)
	go func() {
		var actual map[string]any
		if err := json.NewDecoder(nodeOutput).Decode(&actual); err != nil {
			readyError <- err
			return
		}
		ready <- actual
	}()
	var nativeReady map[string]any
	select {
	case nativeReady = <-ready:
		if nativeReady["ok"] != true {
			_ = nodeOwner.Close()
			if waitOwnedService(node, nodeResult, stderr, 4) == 6 {
				return 6
			}
			return report(stderr, "provider_unavailable", "私有原生后台没有返回真实 readiness", "inspect_private_runtime_log", 4)
		}
	case err := <-readyError:
		_ = nodeOwner.Close()
		if waitOwnedService(node, nodeResult, stderr, 4) == 6 {
			return 6
		}
		return report(stderr, "provider_unavailable", err.Error(), "inspect_private_runtime_log", 4)
	case err := <-nodeResult:
		return report(stderr, "provider_unavailable", fmt.Sprintf("私有原生后台提前退出: %v", err), "inspect_private_runtime_log", 4)
	case <-ctx.Done():
		_ = nodeOwner.Close()
		return waitOwnedService(node, nodeResult, stderr, 130)
	case <-time.After(60 * time.Second):
		_ = nodeOwner.Close()
		return waitOwnedService(node, nodeResult, stderr, 6)
	}
	server := exec.Command(filepath.Join(install, "server.exe"))
	server.Env, server.Dir, server.Stderr = baseEnv, install, stderr
	configureServiceChild(server)
	serverOwner, err := server.StdinPipe()
	if err != nil {
		_ = nodeOwner.Close()
		return waitOwnedService(node, nodeResult, stderr, 7)
	}
	serverOutput, err := server.StdoutPipe()
	if err != nil {
		_ = nodeOwner.Close()
		return waitOwnedService(node, nodeResult, stderr, 4)
	}
	if err := server.Start(); err != nil {
		_ = nodeOwner.Close()
		_ = waitOwnedService(node, nodeResult, stderr, 4)
		return report(stderr, "provider_unavailable", err.Error(), "inspect_installed_server", 4)
	}
	serverResult := make(chan error, 1)
	go func() { serverResult <- server.Wait() }()
	defer serverOwner.Close()
	endpoint := "http://127.0.0.1:" + strconv.Itoa(*port)
	listening := make(chan error, 1)
	go func() {
		var announcement struct {
			Type   string `json:"type"`
			APIURL string `json:"api_url"`
		}
		if err := json.NewDecoder(serverOutput).Decode(&announcement); err != nil {
			listening <- err
			return
		}
		if announcement.Type != "http_listening" || announcement.APIURL != endpoint {
			listening <- errors.New("所属 HTTP 子进程未确认实际指定监听地址")
			return
		}
		listening <- waitHTTPHealth(ctx, endpoint)
	}()
	var startupError error
	select {
	case startupError = <-listening:
	case <-ctx.Done():
		startupError = ctx.Err()
	case <-time.After(15 * time.Second):
		startupError = errors.New("所属 HTTP 子进程监听未确认")
	}
	if startupError != nil {
		_ = serverOwner.Close()
		_ = nodeOwner.Close()
		serverCode := waitOwnedService(server, serverResult, stderr, 7)
		nodeCode := waitOwnedService(node, nodeResult, stderr, 7)
		if serverCode == 6 || nodeCode == 6 {
			return 6
		}
		return report(stderr, "provider_unavailable", startupError.Error(), "inspect_explicit_port_and_service_logs", 4)
	}
	state := map[string]any{"ok": true, "readiness": nativeReady["readiness"], "http": endpoint, "data_root": canonicalData, "install_root": install, "runtime_pid": node.Process.Pid, "server_pid": server.Process.Pid, "resource_samples": samples}
	encoder := json.NewEncoder(stdout)
	if !*asJSON {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(state); err != nil {
		_ = serverOwner.Close()
		_ = nodeOwner.Close()
		_ = waitOwnedService(server, serverResult, stderr, 7)
		return waitOwnedService(node, nodeResult, stderr, 7)
	}
	code := 0
	select {
	case <-ctx.Done():
		code = 130
	case err := <-serverResult:
		_, _ = fmt.Fprintf(stderr, "私有 HTTP 服务退出: %v\n", err)
		_ = nodeOwner.Close()
		return waitOwnedService(node, nodeResult, stderr, 7)
	case err := <-nodeResult:
		_, _ = fmt.Fprintf(stderr, "私有原生后台退出: %v\n", err)
		_ = serverOwner.Close()
		return waitOwnedService(server, serverResult, stderr, 7)
	}
	_ = serverOwner.Close()
	serverCode := waitOwnedService(server, serverResult, stderr, code)
	_ = nodeOwner.Close()
	nodeCode := waitOwnedService(node, nodeResult, stderr, code)
	if serverCode == 6 || nodeCode == 6 {
		return 6
	}
	return code
}

func waitOwnedService(command *exec.Cmd, result <-chan error, stderr io.Writer, code int) int {
	select {
	case err := <-result:
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "所属服务 PID %d 退出: %v\n", command.Process.Pid, err)
		}
		return code
	case <-time.After(15 * time.Second):
		_, _ = fmt.Fprintf(stderr, "所属服务 PID %d 退出尚未确认，保留资源及原请求。\n", command.Process.Pid)
		return 6
	}
}

func waitHTTPHealth(ctx context.Context, endpoint string) error {
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	client := &http.Client{Timeout: time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("redirect refused") }}
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/v1/health", nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("指定端口的产品 HTTP 健康检查未成功")
		case <-tick.C:
		}
	}
}

func privateDataDirectory(input, install string, protected []string) (string, error) {
	path := filepath.Clean(input)
	var parts []string
	ancestor := path
	for {
		_, err := os.Stat(ancestor)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) || filepath.Dir(ancestor) == ancestor {
			return "", err
		}
		parts = append(parts, filepath.Base(ancestor))
		ancestor = filepath.Dir(ancestor)
	}
	actual, err := (agents.ProjectFiles{}).CanonicalRoot(ancestor)
	if err != nil {
		return "", err
	}
	for i := len(parts) - 1; i >= 0; i-- {
		actual = filepath.Join(actual, parts[i])
	}
	for _, root := range append([]string{install}, protected...) {
		root, err := (agents.ProjectFiles{}).CanonicalRoot(root)
		if err != nil {
			return "", err
		}
		if pathContains(root, actual) || pathContains(actual, root) {
			return "", errors.New("产品数据目录与安装目录或受保护目录重叠")
		}
	}
	if err := os.MkdirAll(actual, 0o700); err != nil {
		return "", err
	}
	return (agents.ProjectFiles{}).CanonicalRoot(actual)
}

func pathContains(root, path string) bool {
	if runtime.GOOS == "windows" {
		root, path = strings.ToLower(root), strings.ToLower(path)
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && !filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
