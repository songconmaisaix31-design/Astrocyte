package agents

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 纯请求/拒绝检查，不创建原生进程或模拟成功协议响应。
func TestRelayNativeDeadlineRejectsBeforeController(t *testing.T) {
	for _, seconds := range []int{-1, 1801} {
		driver := NewRelayNativeDriver(nil, RelayNativeDriverOptions{SessionDeadlineSeconds: seconds})
		for _, effect := range []func(context.Context, app.RelayLaunch) (app.RelayObservation, error){driver.Launch, driver.Resume} {
			_, err := effect(context.Background(), app.RelayLaunch{})
			var prelaunch *domain.NativePrelaunchFailure
			var known *apierrors.ServiceError
			if !errors.As(err, &prelaunch) || !errors.As(err, &known) || known.Code != apierrors.BudgetExhausted {
				t.Fatal("invalid deadline reached an adapter or was classified as unknown delivery")
			}
		}
	}
	for _, seconds := range []int{0, 1, 900, 1800} {
		driver := NewRelayNativeDriver(nil, RelayNativeDriverOptions{SessionDeadlineSeconds: seconds})
		request := driver.request(app.RelayLaunch{Spec: domain.AgentSpec{Name: "codex"}, Workspace: domain.RelayWorkspace{Root: t.TempDir()}})
		if request.Command.DeadlineSeconds != seconds {
			t.Fatal("trusted lifetime was not passed to the existing Native request")
		}
	}
}

// 核已安装入口和真实目录身份，但不运行 CLI 或发送任何任务。
func TestRelayPublicEOFPolicyRejectsExpandedCommandAndSource(t *testing.T) {
	spec, err := RelayPublicEOFTerminalSpec()
	if err != nil {
		t.Skip("requires an installed Claude entrypoint; no CLI was started")
	}
	root := t.TempDir()
	policy, err := NewRelayPublicEOFTerminalPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	launch := app.RelayLaunch{Spec: spec, Project: domain.LocalProject{Root: root}, Workspace: domain.RelayWorkspace{SourceRoot: root}}
	// 此处只验证 launch 门禁，不将门禁通过作为原生能力/工具清单证据。
	if err := policy(context.Background(), launch); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*app.RelayLaunch){
		func(r *app.RelayLaunch) {
			r.Spec.LaunchCmd = append(append([]string(nil), r.Spec.LaunchCmd...), "--tools=Read")
		},
		func(r *app.RelayLaunch) { r.Spec.LaunchCmd = []string{"powershell", "-Command", "claude"} },
		func(r *app.RelayLaunch) { r.Spec.InputMode = "line" },
		func(r *app.RelayLaunch) { r.Project.Settings.AllowedTools = []string{"read"} },
		func(r *app.RelayLaunch) { r.Project.Root = t.TempDir() },
		func(r *app.RelayLaunch) { r.Workspace.SourceRoot = t.TempDir() },
	} {
		altered := launch
		change(&altered)
		if err := policy(context.Background(), altered); err == nil {
			t.Fatal("expanded tools, command or source passed the fixed public launch gate")
		}
	}
	if _, err := NewRelayPublicEOFTerminalPolicy(filepath.Join("relative", "public")); err == nil {
		t.Fatal("relative public root was implicitly approved")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(policy(ctx, launch), context.Canceled) {
		t.Fatal("canceled launch continued to inspect its source")
	}
}

func TestRelayOwnedExecutionRejectsUnownedAndMissingController(t *testing.T) {
	native := &Native{id: "codex", processes: map[string]*nativeProcess{}}
	for _, session := range []domain.NativeSession{
		{},
		{ID: "public", NativeID: "public-native", ProjectID: "public-project", CLI: "codex", Ownership: "external_observed"},
		{ID: "public", NativeID: "public-native", ProjectID: "public-project", CLI: "codex", Ownership: "owned"},
	} {
		if evidence, err := native.OwnedRelayExecution(context.Background(), session); err == nil || evidence.ProcessID != 0 || evidence.LaunchPolicy != "" {
			t.Fatal("unowned or missing controller produced execution evidence")
		}
	}
}
