package agents

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

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

func TestRelayFirstInputCheckpointRefusesUnarmedController(t *testing.T) {
	registry := NewRegistry("unopened-checkpoint-root")
	callbackCalled := false
	guard := func(context.Context, domain.NativeSession, *domain.RelaySessionArchive, string) error {
		callbackCalled = true
		return nil
	}
	driver := NewRelayNativeDriver(registry, RelayNativeDriverOptions{BeforeFirstInput: guard})
	_, err := driver.Launch(context.Background(), app.RelayLaunch{Spec: domain.AgentSpec{Name: "codex"}, OperationID: uuid.NewString()})
	var known *apierrors.ServiceError
	var unstarted *domain.NativePrelaunchFailure
	if !errors.As(err, &known) || known.Code != apierrors.UnsupportedCapability || !errors.As(err, &unstarted) || callbackCalled {
		t.Fatal("unarmed controller reached the first input checkpoint")
	}
	adapter, err := registry.Adapter("claude")
	if err != nil {
		t.Fatal(err)
	}
	_, err = adapter.(*Native).StartRelayOperationWithInputGuard(context.Background(), domain.NativeRequest{}, uuid.NewString(), guard)
	if !errors.As(err, &known) || known.Code != apierrors.UnsupportedCapability || callbackCalled {
		t.Fatal("unsupported provider fabricated a Codex initialization checkpoint")
	}
}

// 核已安装入口和真实目录身份，但不运行 CLI 或发送任何任务。
func TestRelayPublicEOFPolicyRejectsExpandedCommandAndSource(t *testing.T) {
	spec, err := RelayPublicEOFTerminalSpec()
	if err != nil {
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != apierrors.ProviderUnavailable || len(spec.LaunchCmd) != 0 {
			t.Fatal("unavailable installation was not explicitly rejected without a command")
		}
	}
	root := t.TempDir()
	policy, err := NewRelayPublicEOFTerminalPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	launch := app.RelayLaunch{Spec: spec, Project: domain.LocalProject{Root: root}, Workspace: domain.RelayWorkspace{SourceRoot: root}}
	// 此处只验证 launch 门禁，不将门禁通过作为原生能力/工具清单证据。
	gateErr := policy(context.Background(), launch)
	if err == nil {
		if gateErr != nil {
			t.Fatal(gateErr)
		}
	} else {
		var known *apierrors.ServiceError
		if !errors.As(gateErr, &known) || known.Code != apierrors.ProviderUnavailable {
			t.Fatal("missing installed CLI passed the public launch gate")
		}
		// 无 CLI 仍执行下方全部目录/命令/工具拒绝检查，不伪造可启动能力。
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
	t.Run("unavailable_cli_is_a_rejection", func(t *testing.T) {
		// 真实缺失入口的拒绝路径；不放替身可执行文件或启动 helper。
		t.Setenv("PATH", t.TempDir())
		unavailable, err := RelayPublicEOFTerminalSpec()
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != apierrors.ProviderUnavailable || len(unavailable.LaunchCmd) != 0 {
			t.Fatal("missing CLI produced a launchable registration")
		}
		if err := policy(context.Background(), launch); !errors.As(err, &known) || known.Code != apierrors.ProviderUnavailable {
			t.Fatal("missing CLI was silently allowed by the public policy")
		}
	})
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

// 仅核首错误安全投影，不制造控制器、成功响应或已完成状态。
func TestRelayProtocolFirstFailureIsSafeAndPreserved(t *testing.T) {
	p := &nativeProcess{}
	p.recordProtocolFailure("thread/read", -32603, protocolReason("credentials PRIVATE_RPC_ERROR_CANARY"))
	first := *p.firstProtocolFailure
	p.recordProtocolFailure("turn/start", -32600, "reply_deadline")
	encoded, err := json.Marshal(p.firstProtocolFailure)
	if err != nil || p.firstProtocolFailure.Method != first.Method || p.firstProtocolFailure.Code != first.Code || p.firstProtocolFailure.ObservedAt != first.ObservedAt || first.ObservedAt.IsZero() || first.Reason != "credentials_rejected" || strings.Contains(string(encoded), "PRIVATE_RPC_ERROR_CANARY") {
		t.Fatal("first failure was overwritten, incomplete or leaked its raw response")
	}
}
