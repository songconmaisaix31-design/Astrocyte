package agents

import (
	"context"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"time"
)

// 仅在原持有进程锁内追加；时间来自事实产生位置。
func (p *nativeProcess) relayFact(kind, source, operation string) {
	if p.relayEvidence.ControllerID == "" {
		return
	}
	facts := p.relayEvidence.Facts
	p.relayEvidence.Facts = append(facts, domain.RelayDriverFact{OperationID: operation, Kind: kind, Status: "confirmed", Source: source, Sequence: int64(len(facts) + 1), TS: time.Now().UTC()})
}

func (p *nativeProcess) consumptionUnknown(operation string) {
	if p.relayEvidence.ControllerID == "" {
		return
	}
	p.relayEvidence.Facts = append(p.relayEvidence.Facts, domain.RelayDriverFact{OperationID: operation, Kind: "input_consumption", Status: "unknown", Source: "native_protocol", Reason: "client_consumption_not_acknowledged", Sequence: int64(len(p.relayEvidence.Facts) + 1), TS: time.Now().UTC()})
}

func (n *Native) OwnedRelayControllerEvidence(ctx context.Context, session domain.NativeSession) (domain.RelayControllerEvidence, error) {
	if err := ctx.Err(); err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p, err := n.get(session)
	if err != nil {
		return domain.RelayControllerEvidence{}, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// 旧进程缺少原启动 operation，只返回未记录，不推导新的归属证据。
	if p.relayEvidence.StartupOperationID == "" {
		return domain.RelayControllerEvidence{}, nil
	}
	if p.projectID != session.ProjectID || session.Ownership != "owned" || session.CLI != n.id || p.relayEvidence.SessionID != session.ID || p.relayEvidence.ControllerID == "" {
		return domain.RelayControllerEvidence{}, nativeError(apierrors.ScopeDenied, "original controller identity differs")
	}
	result := p.relayEvidence
	result.NativeID = p.nativeID
	result.Facts = append([]domain.RelayDriverFact{}, result.Facts...)
	return result, nil
}

func (n *Native) StartOperation(ctx context.Context, request domain.NativeRequest, operation string) (domain.NativeSession, error) {
	if operation == "" {
		return request.Session, &domain.NativePrelaunchFailure{Cause: nativeError(apierrors.ValidationFailed, "relay startup operation is required")}
	}
	return n.start(ctx, request, false, operation)
}

func (n *Native) ResumeOperation(ctx context.Context, request domain.NativeRequest, operation string) (domain.NativeSession, error) {
	if err := n.validateRelayResume(ctx, request, operation); err != nil {
		return request.Session, err
	}
	return n.start(ctx, request, true, operation)
}

func (n *Native) validateRelayResume(ctx context.Context, request domain.NativeRequest, operation string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if operation == "" {
		return nativeError(apierrors.ValidationFailed, "relay resume operation is required")
	}
	p, err := n.get(request.Session)
	if err != nil {
		return err
	}
	p.mu.Lock()
	identity := p.projectID == request.Project.ID && sameRelayMetadataCWD(p.cmd.Dir, request.Project.Root) && request.Session.Ownership == "owned" && request.SessionID == request.Session.ID && p.relayEvidence.StartupOperationID != ""
	p.mu.Unlock()
	if !identity {
		return nativeError(apierrors.ScopeDenied, "resume differs from the original controller workspace or operation")
	}
	if !request.Session.StopConfirmed || !p.observe().StopConfirmed || request.Session.NativeID == "" {
		return nativeError(apierrors.DeliveryUnknown, "resume requires original controller exit and native identity")
	}
	return nil
}

func (n *Native) SendOperation(ctx context.Context, session domain.NativeSession, input, operation string) (domain.NativeObservation, error) {
	if operation == "" {
		return domain.NativeObservation{}, nativeError(apierrors.ValidationFailed, "relay input operation is required")
	}
	if evidence, err := n.OwnedRelayControllerEvidence(ctx, session); err != nil {
		return domain.NativeObservation{}, err
	} else if evidence.ControllerID == "" {
		return domain.NativeObservation{}, nativeError(apierrors.DeliveryUnknown, "original relay controller operation is unrecorded")
	}
	return n.send(ctx, session, input, operation)
}
