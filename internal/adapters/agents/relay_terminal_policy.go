package agents

import (
	"context"
	"os"
	"path/filepath"
	"reflect"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只解析已安装的公开 CLI 入口，不启动 CLI、不读取认证或私人配置。
// 公开根目录许可仍由 H 的受信组装核实，本函数不授予来源目录权限。
func RelayPublicEOFTerminalSpec() (domain.AgentSpec, error) {
	cmd, err := installedCommand("claude", []string{"-p", "--input-format", "text", "--output-format", "stream-json", "--verbose", "--safe-mode", "--tools=", "--disallowedTools", "mcp__*", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--no-chrome"})
	if err != nil {
		return domain.AgentSpec{}, err
	}
	return domain.AgentSpec{Name: "claude", DriveMode: "terminal", InputMode: "stdin_eof", LaunchCmd: append([]string(nil), cmd.Args...), Capabilities: map[string]domain.CapabilityObservation{}}, nil
}

// 核对一次性输入和固定禁用工具命令；调用方必须先核获准的公开根目录。
// 不能接受 HTTP 命令、自选脚本、附加工具参数或自由工具注册。
func relayPublicEOFTerminalPolicy(ctx context.Context, launch app.RelayLaunch) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	spec, err := RelayPublicEOFTerminalSpec()
	if err != nil {
		return err
	}
	if launch.Spec.Name != spec.Name || launch.Spec.InputMode != spec.InputMode || launch.Spec.DriveMode != spec.DriveMode || !reflect.DeepEqual(launch.Spec.LaunchCmd, spec.LaunchCmd) || len(launch.Project.Settings.AllowedTools) != 0 {
		return nativeError(apierrors.UnsupportedCapability, "public EOF terminal requires the installed fixed CLI with all tools disabled")
	}
	return nil
}

// 对唯一显式获准根目录保存文件身份，每次启动重新核实；默认组装仍关闭。
func NewRelayPublicEOFTerminalPolicy(approvedRoot string) (func(context.Context, app.RelayLaunch) error, error) {
	if !filepath.IsAbs(approvedRoot) {
		return nil, nativeError(apierrors.ScopeDenied, "public EOF root must be explicitly selected")
	}
	canonical, err := (RelayOwnedRoots{}).CanonicalRoot(approvedRoot)
	if err != nil {
		return nil, err
	}
	approved, err := os.Stat(canonical)
	if err != nil || !approved.IsDir() {
		return nil, nativeError(apierrors.ScopeDenied, "approved public EOF root is unavailable")
	}
	return func(ctx context.Context, launch app.RelayLaunch) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, root := range []string{launch.Project.Root, launch.Workspace.SourceRoot} {
			canonical, err := (RelayOwnedRoots{}).CanonicalRoot(root)
			if err != nil {
				return err
			}
			info, err := os.Stat(canonical)
			if err != nil || !os.SameFile(approved, info) {
				return nativeError(apierrors.ScopeDenied, "EOF launch does not match the approved public source")
			}
		}
		return relayPublicEOFTerminalPolicy(ctx, launch)
	}, nil
}
