package agents

import "time"

// 该补充仅属于 R；普通 Native 消费方无需复制此文件或增加进程字段。
// 返回原启动记录的派生副本，不发 RPC、不推导旧进程或有效工具清单。
func (p *nativeProcess) relayExecutionMetadata() relayExecutionMetadata {
	p.mu.Lock()
	defer p.mu.Unlock()
	metadata := relayExecutionMetadata{
		ProjectID: p.projectID, StartedAt: p.startedAt, DeadlineAt: p.deadlineAt,
		PermissionPolicy: p.permissionPolicy, WriterShutdown: p.relayCapture != nil,
	}
	if p.firstProtocolFailure != nil {
		failure := *p.firstProtocolFailure
		metadata.FirstProtocolFailure = &failure
	}
	return metadata
}

// 保留原控制器第一条安全 RPC 错误；不扩大普通 Native 消费方依赖。
func (p *nativeProcess) recordProtocolFailure(method string, code int, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.firstProtocolFailure == nil {
		p.firstProtocolFailure = &RelayProtocolFailure{Method: method, Code: code, Reason: reason, ObservedAt: time.Now().UTC()}
	}
}
