package agents

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 观察器在原控制器启动前挂入接收器，只保留派生数据和必要来源身份。
type relayNativeCapture struct {
	mu                        sync.Mutex
	state                     domain.RelayNativeSessionCapture
	workspace, nativeID, path string
	previousPath              string
	closed, readerDone        chan struct{}
	closedSeen, readerEnded   bool
	diagnosticMethod          string
	failed                    bool
}

func newRelayNativeCapture(root string) *relayNativeCapture {
	return &relayNativeCapture{workspace: root, closed: make(chan struct{}), readerDone: make(chan struct{}), state: domain.RelayNativeSessionCapture{Events: []domain.RelayNativeSessionEvent{}, SourceStatus: "unknown", PrefixComplete: "unknown", Finalized: "unknown"}}
}

func (c *relayNativeCapture) fail(reason string, unknown bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failLocked(reason, unknown)
}

func (c *relayNativeCapture) failLocked(reason string, unknown bool) {
	c.failed = true
	if c.state.FirstGap == "" {
		c.state.FirstGap = reason
	}
	if unknown {
		c.state.Unknown++
	} else {
		c.state.Errors++
	}
	c.state.Finalized = "unknown"
}

func (c *relayNativeCapture) bind(raw json.RawMessage, request domain.NativeRequest, resume bool) error {
	var response struct {
		Thread struct {
			ID, Cwd      string
			Path         *string
			ForkedFromID *string
			Ephemeral    bool
		} `json:"thread"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		c.fail("native_start_metadata_decode_failed", false)
		return nativeError(apierrors.InternalError, "native source metadata response is invalid")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	source := response.Thread
	if source.ID == "" || !sameRelayMetadataCWD(source.Cwd, request.Project.Root) || c.nativeID != "" && c.nativeID != source.ID || resume && source.ID != request.Session.NativeID || !resume && source.ForkedFromID != nil {
		c.failLocked("native_start_identity_mismatch", false)
		return nativeError(apierrors.ScopeDenied, "native notification source identity differs")
	}
	c.nativeID = source.ID
	c.state.SourceStatus = "available"
	if !resume && c.state.Errors == 0 && c.state.Unknown == 0 {
		c.state.PrefixComplete = "confirmed"
	}
	if source.Path != nil && (!filepath.IsAbs(*source.Path) || filepath.Ext(*source.Path) != ".jsonl" || source.Ephemeral) {
		c.failLocked("native_start_archive_metadata_invalid", false)
		return nativeError(apierrors.ScopeDenied, "native start archive metadata is invalid")
	}
	if source.Path != nil {
		path := filepath.Clean(*source.Path)
		if resume && c.previousPath != "" && path != c.previousPath {
			// 恢复响应可使用 Windows 长路径别名，只核控制器正选的两个文件。
			previous, err := os.Stat(c.previousPath)
			if err != nil {
				c.failLocked("native_resume_archive_identity_unavailable", false)
				return nativeError(apierrors.EvidenceMissing, "original native archive identity is unavailable")
			}
			current, err := os.Stat(path)
			if err != nil {
				c.failLocked("native_resume_archive_identity_unavailable", false)
				return nativeError(apierrors.EvidenceMissing, "resumed native archive identity is unavailable")
			}
			if !previous.Mode().IsRegular() || !current.Mode().IsRegular() || !os.SameFile(previous, current) {
				c.failLocked("native_resume_archive_identity_changed", false)
				return nativeError(apierrors.ScopeDenied, "resumed native archive is a different file")
			}
			path = c.previousPath
		}
		c.path = path
	}
	return nil
}

// 使用现有 JSON 库解析得到的 RPC 对象，不增加 framing 或字节 parser。
func (c *relayNativeCapture) observe(record map[string]json.RawMessage) error {
	if _, ok := record["method"]; !ok {
		return nil
	}
	var method string
	if err := json.Unmarshal(record["method"], &method); err != nil {
		c.fail("native_notification_method_invalid", false)
		return nativeError(apierrors.InternalError, "native notification method is invalid")
	}
	var params struct {
		ThreadID   string                   `json:"threadId"`
		Thread     struct{ ID, Cwd string } `json:"thread"`
		TokenUsage *struct {
			Last struct {
				TotalTokens *int `json:"totalTokens"`
			} `json:"last"`
		} `json:"tokenUsage"`
	}
	if err := json.Unmarshal(record["params"], &params); err != nil {
		c.fail("native_notification_params_invalid", false)
		return nativeError(apierrors.InternalError, "native notification parameters are invalid")
	}
	if method == "thread/started" {
		params.ThreadID = params.Thread.ID
		if !sameRelayMetadataCWD(params.Thread.Cwd, c.workspace) {
			c.fail("native_notification_cwd_mismatch", false)
			return nativeError(apierrors.ScopeDenied, "native notification CWD differs")
		}
	}
	// 连接级账户和配置通知不属于本任务；不复制任何账户属性。
	if relayCodexConnectionNotification(method) && params.ThreadID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if params.ThreadID == "" || c.nativeID != "" && params.ThreadID != c.nativeID {
		c.failLocked("native_notification_identity_unverified", true)
		return nativeError(apierrors.ScopeDenied, "native notification thread identity is unavailable or differs")
	}
	if c.nativeID == "" {
		// start 响应与通知可在不同 goroutine 交错；bind 随后核实际 ID/CWD。
		c.nativeID = params.ThreadID
	}
	kind, known := relayCodexNotificationKind(method)
	if !known {
		c.diagnosticMethod = method
		c.failLocked("native_notification_semantics_unknown", true)
		return nativeError(apierrors.UnsupportedCapability, "native notification semantics are unknown")
	}
	if c.closedSeen || len(c.state.Events) >= 16384 {
		c.failLocked("native_notification_after_close_or_capacity", false)
		return nativeError(apierrors.BudgetExhausted, "native notification capture capacity or closure was exceeded")
	}
	var tokens *int
	if method == "thread/tokenUsage/updated" {
		if params.TokenUsage != nil {
			tokens = params.TokenUsage.Last.TotalTokens
		}
		if tokens != nil && *tokens < 0 {
			c.failLocked("native_token_usage_invalid", false)
			return nativeError(apierrors.InternalError, "native reported token usage is negative")
		}
	}
	c.state.LastSequence++
	c.state.Events = append(c.state.Events, domain.RelayNativeSessionEvent{Sequence: c.state.LastSequence, Kind: kind, Tokens: tokens, TS: time.Now().UTC()})
	if method == "thread/closed" {
		c.closedSeen = true
		close(c.closed)
	}
	return nil
}

func relayCodexConnectionNotification(method string) bool {
	switch method {
	case "skills/changed", "project/changed", "mcpServer/startupStatus/updated", "mcpServer/oauthLogin/completed", "account/updated", "account/gatewayOAuth/changed", "account/rateLimits/updated", "app/list/updated", "remoteControl/status/changed", "externalAgentConfig/import/progress", "externalAgentConfig/import/completed", "configWarning", "deprecationNotice", "windows/worldWritableWarning", "windowsSandbox/setupCompleted", "account/login/completed", "serverRequest/resolved":
		return true
	}
	return false
}

func relayCodexNotificationKind(method string) (string, bool) {
	if relayCodexConnectionNotification(method) {
		return "native_activity", true
	}
	switch method {
	case "thread/started":
		return "session_started", true
	case "thread/closed":
		return "session_closed", true
	case "turn/started":
		return "turn_started", true
	case "turn/completed":
		return "turn_completed", true
	case "item/reasoning/summaryTextDelta", "item/reasoning/summaryPartAdded", "item/reasoning/textDelta":
		return "reasoning", true
	case "item/agentMessage/delta":
		return "text", true
	case "thread/tokenUsage/updated":
		return "token_usage", true
	case "rawResponseItem/completed", "rawResponse/completed":
		return "model_response", true
	case "error", "thread/realtime/error":
		return "native_error", true
	case "warning", "guardianWarning":
		return "native_warning", true
	case "thread/status/changed", "thread/archived", "thread/deleted", "thread/unarchived", "thread/reverted", "thread/name/updated", "thread/attachment/updated", "thread/goal/updated", "thread/prediction/updated", "thread/goal/cleared", "thread/queue/changed", "thread/project/updated", "thread/environment/connected", "thread/environment/disconnected", "thread/settings/updated", "thread/compacted", "hook/started", "hook/completed", "turn/diff/updated", "turn/plan/updated", "item/started", "item/completed", "item/plan/delta", "item/autoApprovalReview/started", "item/autoApprovalReview/completed", "autoApprovalReview/strictReviewRequired", "item/commandExecution/outputDelta", "item/commandExecution/terminalInteraction", "item/fileChange/outputDelta", "item/fileChange/patchUpdated", "item/mcpToolCall/progress", "model/rerouted", "model/verification", "modelProvider/authRecoveryStarted", "modelProvider/authRecoveryCompleted", "turn/moderationMetadata", "model/safetyBuffering/updated", "thread/realtime/started", "thread/realtime/itemAdded", "thread/realtime/item/started", "thread/realtime/item/transcript/delta", "thread/realtime/item/completed", "thread/realtime/transcript/delta", "thread/realtime/transcript/done", "thread/realtime/outputAudio/delta", "thread/realtime/sdp", "thread/realtime/closed":
		return "native_activity", true
	}
	return "", false
}

func (c *relayNativeCapture) readerFinished() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readerEnded = true
	if !c.closedSeen {
		c.failLocked("native_stream_ended_without_shutdown_ack", true)
	}
	close(c.readerDone)
}

func (c *relayNativeCapture) snapshot(stopped bool) domain.RelayNativeSessionCapture {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := c.state
	result.Events = append([]domain.RelayNativeSessionEvent{}, c.state.Events...)
	if stopped && c.closedSeen && c.readerEnded && result.Errors == 0 && result.SourceStatus == "available" {
		result.Finalized = "confirmed"
	}
	return result
}

func (n *Native) StartRelay(ctx context.Context, r domain.NativeRequest) (domain.NativeSession, error) {
	return n.StartRelayOperation(ctx, r, "")
}

func (n *Native) StartRelayOperation(ctx context.Context, r domain.NativeRequest, operation string) (domain.NativeSession, error) {
	return n.startRelayWithGuard(ctx, r, operation, nil)
}

func (n *Native) StartRelayOperationWithInputGuard(ctx context.Context, r domain.NativeRequest, operation string, guard RelayBeforeFirstInput) (domain.NativeSession, error) {
	if n.id != "codex" || operation == "" || guard == nil {
		return r.Session, &domain.NativePrelaunchFailure{Cause: nativeError(apierrors.UnsupportedCapability, "first input source checkpoint requires an armed Codex operation")}
	}
	return n.startRelayWithGuard(ctx, r, operation, n.relaySourceCheckpoint(guard))
}

func (n *Native) relaySourceCheckpoint(guard RelayBeforeFirstInput) func(context.Context, domain.NativeSession) error {
	return func(ctx context.Context, session domain.NativeSession) error {
		p, err := n.get(session)
		if err != nil {
			return err
		}
		if p.relayCapture == nil {
			return nativeError(apierrors.DeliveryUnknown, "original source checkpoint was not armed")
		}
		// armed capture 的路径来自同次 thread/start 响应，此读取不发 RPC。
		archive, err := n.OwnedRelayArchive(ctx, session)
		if err != nil {
			var known *apierrors.ServiceError
			if errors.As(err, &known) && known.Code == apierrors.EvidenceMissing {
				if err := guard(ctx, session, nil, "archive_not_yet_available"); err != nil {
					return err
				}
				return nativeError(apierrors.EvidenceMissing, "selected archive is not yet available before input")
			}
			return err
		}
		return guard(ctx, session, &archive, "")
	}
}

func (n *Native) startRelayWithGuard(ctx context.Context, r domain.NativeRequest, operation string, guard func(context.Context, domain.NativeSession) error) (domain.NativeSession, error) {
	if n.id != "codex" {
		return n.start(ctx, r, false, operation)
	}
	capture := newRelayNativeCapture(r.Project.Root)
	n.mu.Lock()
	if n.relayCaptures == nil {
		n.relayCaptures = map[string]*relayNativeCapture{}
	}
	n.relayCaptures[r.SessionID] = capture
	n.mu.Unlock()
	return n.startWithInputGuard(ctx, r, false, operation, guard)
}

func (n *Native) ResumeRelay(ctx context.Context, r domain.NativeRequest) (domain.NativeSession, error) {
	return n.ResumeRelayOperation(ctx, r, "")
}

func (n *Native) ResumeRelayOperation(ctx context.Context, r domain.NativeRequest, operation string) (domain.NativeSession, error) {
	if operation != "" {
		if err := n.validateRelayResume(ctx, r, operation); err != nil {
			return r.Session, err
		}
	}
	if n.id != "codex" {
		if operation != "" {
			return n.ResumeOperation(ctx, r, operation)
		}
		return n.Resume(ctx, r)
	}
	n.mu.Lock()
	previous := n.relayCaptures[r.SessionID]
	n.mu.Unlock()
	if previous == nil {
		if operation != "" {
			return n.ResumeOperation(ctx, r, operation)
		}
		return n.Resume(ctx, r)
	}
	capture := newRelayNativeCapture(r.Project.Root)
	capture.state = previous.snapshot(false)
	if operation != "" {
		// 新连接从自身序号开始；旧连接的派生事件由原事件库保存。
		capture.state.Events = []domain.RelayNativeSessionEvent{}
		capture.state.LastSequence = 0
		capture.state.PrefixComplete, capture.state.Finalized = "unknown", "unknown"
	}
	previous.mu.Lock()
	capture.previousPath = previous.path
	previous.mu.Unlock()
	capture.nativeID = r.Session.NativeID
	capture.state.Unknown++
	if capture.state.FirstGap == "" {
		capture.state.FirstGap = "native_resume_raw_events_unavailable"
	}
	n.mu.Lock()
	n.relayCaptures[r.SessionID] = capture
	n.mu.Unlock()
	if operation != "" {
		return n.ResumeOperation(ctx, r, operation)
	}
	return n.Resume(ctx, r)
}

func (n *Native) OwnedRelayNativeCapture(ctx context.Context, session domain.NativeSession) (domain.RelayNativeSessionCapture, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayNativeSessionCapture{}, err
	}
	p, err := n.get(session)
	if err != nil {
		return domain.RelayNativeSessionCapture{}, err
	}
	if session.Ownership != "owned" || session.CLI != n.id || p.relayCapture == nil {
		return domain.RelayNativeSessionCapture{}, nativeError(apierrors.UnsupportedCapability, "native controller has no armed notification source")
	}
	capture := p.relayCapture
	state := capture.snapshot(p.observe().StopConfirmed)
	p.mu.Lock()
	state.ControllerID, state.StartupOperationID = p.relayEvidence.ControllerID, p.relayEvidence.StartupOperationID
	p.mu.Unlock()
	capture.mu.Lock()
	failed := capture.failed
	capture.mu.Unlock()
	if failed {
		return state, nativeError(apierrors.ProviderUnavailable, "native notification source failed; inspect its preserved coverage gap")
	}
	return state, nil
}

func (n *Native) StopRelay(ctx context.Context, session domain.NativeSession) (domain.NativeObservation, error) {
	p, err := n.get(session)
	if err != nil {
		return domain.NativeObservation{}, err
	}
	if p.relayCapture == nil || p.observe().StopConfirmed {
		return n.Stop(ctx, session)
	}
	capture := p.relayCapture
	graceCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	// 官方 unsubscribe 只取消订阅；thread/closed 才证明 writer shutdown 成功。
	_, graceErr := p.call(graceCtx, "thread/unsubscribe", map[string]any{"threadId": session.NativeID}, false)
	if graceErr == nil {
		select {
		case <-capture.closed:
		case <-graceCtx.Done():
			graceErr = nativeError(apierrors.EvidenceMissing, "native thread closure acknowledgement was not observed")
		}
	}
	if graceErr != nil {
		capture.fail("native_shutdown_ack_unavailable", true)
	}
	obs, stopErr := n.Stop(context.WithoutCancel(ctx), session)
	if stopErr != nil {
		return obs, stopErr
	}
	select {
	case <-capture.readerDone:
	case <-graceCtx.Done():
		capture.fail("native_reader_end_unconfirmed", true)
		graceErr = errors.Join(graceErr, nativeError(apierrors.EvidenceMissing, "native reader termination was not observed"))
	}
	if graceErr != nil {
		return obs, &domain.RelayArchiveObservationFailure{Cause: graceErr}
	}
	return obs, nil
}
