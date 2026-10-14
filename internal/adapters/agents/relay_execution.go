package agents

import (
	"context"
	"reflect"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 仅供本机 owner 的内部派生证据；不作为 HTTP DTO、权限或能力承诺。
// 不包含启动环境、原生输出、凭据、模型名称、工作目录或原档路径。
type RelayOwnedExecution struct {
	ProcessID            int                   `json:"owned_process_id"`
	StartedAt            time.Time             `json:"started_at"`
	DeadlineAt           time.Time             `json:"deadline_at"`
	ExitConfirmed        bool                  `json:"exit_confirmed"`
	CLI                  string                `json:"cli"`
	Version              string                `json:"version"`
	LaunchPolicy         string                `json:"launch_policy"`
	PermissionPolicy     string                `json:"permission_policy"`
	EffectiveToolSet     string                `json:"effective_tool_set"`
	DisabledFeatures     []string              `json:"disabled_features"`
	FirstProtocolFailure *RelayProtocolFailure `json:"first_protocol_failure,omitempty"`
	MetadataStatus       string                `json:"process_metadata_status"`
}

// 只保存第一条固定安全摘要，日志关闭时也不丢失本实例的首 RPC 错误。
type RelayProtocolFailure struct {
	Method     string    `json:"method"`
	Code       int       `json:"protocol_code"`
	Reason     string    `json:"reason"`
	ObservedAt time.Time `json:"observed_at"`
}

// 可选补充只由原控制器记录供给；普通 Native 基线缺少该接口时保持未知。
// 此文件不引用 Relay capture、应用服务、根计划或任何新增进程字段。
type relayExecutionMetadata struct {
	ProjectID            string
	StartedAt            time.Time
	DeadlineAt           time.Time
	PermissionPolicy     string
	FirstProtocolFailure *RelayProtocolFailure
	WriterShutdown       bool
}

type relayExecutionMetadataProvider interface {
	relayExecutionMetadata() relayExecutionMetadata
}

// 只核已经持有的精确 controller；无 RPC、无模型调用、无历史/文件扫描。
func (n *Native) OwnedRelayExecution(ctx context.Context, session domain.NativeSession) (RelayOwnedExecution, error) {
	var result RelayOwnedExecution
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if session.ID == "" || session.NativeID == "" || session.ProjectID == "" {
		return result, nativeError(apierrors.DeliveryUnknown, "original owned session identity is incomplete")
	}
	if session.CLI != n.id || session.Ownership != "owned" {
		return result, nativeError(apierrors.ScopeDenied, "execution evidence requires the exact owned controller session")
	}
	p, err := n.get(session)
	if err != nil {
		return result, err
	}
	metadata := relayExecutionMetadata{}
	metadataStatus := "unknown"
	if provider, ok := any(p).(relayExecutionMetadataProvider); ok {
		metadata = provider.relayExecutionMetadata()
		if metadata.ProjectID != session.ProjectID {
			return result, nativeError(apierrors.ScopeDenied, "owned process project binding differs")
		}
		if metadata.StartedAt.IsZero() {
			return result, nativeError(apierrors.DeliveryUnknown, "owned process start metadata is unavailable")
		}
		metadataStatus = "available"
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.nativeID != session.NativeID || p.cmd == nil || p.cmd.Process == nil || p.cmd.Process.Pid <= 0 || p.done == nil {
		return result, nativeError(apierrors.DeliveryUnknown, "original owned process identity is unavailable")
	}
	result = RelayOwnedExecution{ProcessID: p.cmd.Process.Pid, StartedAt: metadata.StartedAt, DeadlineAt: metadata.DeadlineAt, CLI: n.id, Version: session.Version, LaunchPolicy: "unknown", PermissionPolicy: "unknown", EffectiveToolSet: "unknown", DisabledFeatures: []string{}, MetadataStatus: metadataStatus}
	select {
	case <-p.done:
		result.ExitConfirmed = true
	default:
	}
	if metadata.PermissionPolicy != "" {
		result.PermissionPolicy = metadata.PermissionPolicy
	}
	if metadata.FirstProtocolFailure != nil {
		failure := *metadata.FirstProtocolFailure
		result.FirstProtocolFailure = &failure
	}
	if n.id == "codex" {
		args := []string{"app-server", "-c", "project_doc_max_bytes=0", "-c", `web_search="disabled"`, "-c", "agents.enabled=false", "-c", "skills.include_instructions=false", "-c", "include_apps_instructions=false", "-c", "mcp_servers={}", "--enable", "skip_host_skill_discovery"}
		for _, feature := range codexDisabledFeatures {
			args = append(args, "--disable", feature)
		}
		// 通知采集使用现有 writer 关闭参数，必须核同一次真实启动策略。
		if metadata.WriterShutdown {
			args = append(args, "-c", "thread_unload_delay_secs=0")
		}
		cmd, err := installedCommand("codex", args)
		if err != nil {
			return RelayOwnedExecution{}, err
		}
		if !reflect.DeepEqual(p.cmd.Args, cmd.Args) {
			return RelayOwnedExecution{}, nativeError(apierrors.ScopeDenied, "owned process launch policy differs")
		}
		result.LaunchPolicy = "selected_context_tools_disabled_v1"
		result.DisabledFeatures = append(result.DisabledFeatures, codexDisabledFeatures...)
	}
	// 启动策略和权限响应不等于完整有效工具清单，后者保持 unknown。
	return result, nil
}
