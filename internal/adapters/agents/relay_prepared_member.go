package agents

import (
	"context"
	"reflect"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 同一对象保留原 Registry 成员来源，并组合 B 的实际私有准备记录。
type RelayPreparedMemberEvidenceSource struct {
	registry *Registry
	prepared domain.RelayPreparedInputReader
}

var _ domain.RelayMemberEvidenceSource = (*RelayPreparedMemberEvidenceSource)(nil)
var _ domain.RelayPreparedInputReader = (*RelayPreparedMemberEvidenceSource)(nil)
var _ swarm.PreparedMemberInputReader = (*RelayPreparedMemberEvidenceSource)(nil)

func NewRelayPreparedMemberEvidenceSource(registry *Registry, prepared domain.RelayPreparedInputReader) (*RelayPreparedMemberEvidenceSource, error) {
	if registry == nil || prepared == nil {
		return nil, nativeError(apierrors.UnsupportedCapability, "original Registry and actual prepared input reader are required")
	}
	return &RelayPreparedMemberEvidenceSource{registry: registry, prepared: prepared}, nil
}

func (s *RelayPreparedMemberEvidenceSource) ReadRelayMemberEvidence(ctx context.Context, sessionID, operationID, purpose string) (domain.RelayMemberEvidence, error) {
	if s == nil || s.registry == nil {
		return domain.RelayMemberEvidence{}, nativeError(apierrors.UnsupportedCapability, "original member evidence source is not connected")
	}
	return s.registry.ReadRelayMemberEvidence(ctx, sessionID, operationID, purpose)
}

func (s *RelayPreparedMemberEvidenceSource) StopRelayMemberEvidence(ctx context.Context, sessionID, operationID string) (domain.RelayMemberEvidence, error) {
	if s == nil || s.registry == nil {
		return domain.RelayMemberEvidence{}, nativeError(apierrors.UnsupportedCapability, "original member evidence source is not connected")
	}
	return s.registry.StopRelayMemberEvidence(ctx, sessionID, operationID)
}

func (s *RelayPreparedMemberEvidenceSource) ReadPreparedMemberInput(ctx context.Context, sessionID, operationID string) (swarm.PreparedMemberInput, error) {
	input, err := s.ReadPreparedNativeSession(ctx, sessionID, operationID)
	if err != nil {
		return swarm.PreparedMemberInput{}, err
	}
	reader, ok := s.prepared.(workspace.SwarmNativeMessagePreparationStore)
	if !ok {
		return swarm.PreparedMemberInput{}, nativeError(apierrors.UnsupportedCapability, "actual explicit Swarm preparation reader is required")
	}
	prepared, err := reader.ReadSwarmNativeMessagePreparation(ctx, sessionID, operationID)
	if err != nil {
		return swarm.PreparedMemberInput{}, err
	}
	if !reflect.DeepEqual(input, prepared.Input) || prepared.Scope.Kind != "swarm" || prepared.Scope.BindingID == "" || prepared.Lease != (workspace.NativeMessageLease{}) {
		return swarm.PreparedMemberInput{}, nativeError(apierrors.ScopeDenied, "claim requires the original explicit Swarm input before its first lease")
	}
	return swarm.PreparedMemberInput{
		OperationID: input.OperationID, SessionID: input.Session.ID, NativeSessionID: input.Session.NativeID,
		AdapterID: input.Session.CLI, AdapterVersion: input.Session.Version,
		ControllerID: input.ControllerID, StartupOperationID: input.StartupOperationID,
		ProjectID: input.Session.ProjectID, Root: input.WorkspaceRoot,
		ActorID: input.ActorID, ActorKind: input.ActorKind, ActorProjectID: input.ActorProjectID, PreparedAt: input.PreparedAt,
	}, nil
}

// 在原事务 context 内只读 B 元数据和原 Registry；不读取输入对象或执行外部操作。
func (s *RelayPreparedMemberEvidenceSource) ReadPreparedNativeSession(ctx context.Context, sessionID, operationID string) (domain.RelayPreparedInput, error) {
	var empty domain.RelayPreparedInput
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if s == nil || s.registry == nil || s.prepared == nil {
		return empty, nativeError(apierrors.UnsupportedCapability, "actual prepared member source is not connected")
	}
	original, err := s.registry.unattemptedRelayInput(ctx, sessionID, operationID)
	if err != nil {
		return empty, err
	}
	prepared, err := s.prepared.ReadPreparedNativeSession(ctx, sessionID, operationID)
	if err != nil {
		return empty, err
	}
	if err = validatePreparedRelayInput(prepared, original); err != nil {
		return empty, err
	}
	current, err := s.registry.unattemptedRelayInput(ctx, sessionID, operationID)
	if err != nil {
		return empty, err
	}
	if current.Controller.ControllerID != original.Controller.ControllerID || current.Controller.StartupOperationID != original.Controller.StartupOperationID || current.Session.NativeID != original.Session.NativeID || !reflect.DeepEqual(current.Session.ContextPacket, original.Session.ContextPacket) {
		return empty, nativeError(apierrors.DeliveryUnknown, "original member controller changed during prepared input observation")
	}
	return prepared, nil
}

func (r *Registry) unattemptedRelayInput(ctx context.Context, sessionID, operationID string) (RelayOwnedInputOperation, error) {
	original, err := r.ReadOwnedInputOperation(ctx, sessionID, operationID, "claim_work_item")
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	if original.InputAccepted || original.Turn != nil || original.Execution.ExitConfirmed || !time.Now().Before(original.Execution.DeadlineAt) {
		return RelayOwnedInputOperation{}, nativeError(apierrors.VersionConflict, "original member input is no longer unattempted and live")
	}
	adapter, err := r.Adapter(original.Session.CLI)
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	process, err := adapter.(*Native).get(original.Session)
	if err != nil {
		return RelayOwnedInputOperation{}, err
	}
	process.mu.Lock()
	defer process.mu.Unlock()
	if process.inputOperationID == operationID || process.status == "running" || process.status == "blocked" || process.truncated || process.stopping || process.relayEvidence.ControllerID != original.Controller.ControllerID || process.relayEvidence.StartupOperationID != original.Controller.StartupOperationID {
		return RelayOwnedInputOperation{}, nativeError(apierrors.VersionConflict, "original input has been attempted or its controller is unavailable")
	}
	select {
	case <-process.done:
		return RelayOwnedInputOperation{}, nativeError(apierrors.DeliveryUnknown, "original prepared member controller has exited")
	default:
	}
	for _, fact := range process.relayEvidence.Facts {
		if fact.OperationID == operationID {
			return RelayOwnedInputOperation{}, nativeError(apierrors.VersionConflict, "original input already has a native controller fact")
		}
	}
	return original, nil
}

func validatePreparedRelayInput(prepared domain.RelayPreparedInput, original RelayOwnedInputOperation) error {
	session := prepared.Session
	if prepared.OperationID != original.OperationID || session.ID != original.Session.ID || session.NativeID != original.Session.NativeID || session.ProjectID != original.Session.ProjectID || session.CLI != original.Session.CLI || session.Version != original.Session.Version || session.Ownership != "owned" || session.StopConfirmed || prepared.ControllerID != original.Controller.ControllerID || prepared.StartupOperationID != original.Controller.StartupOperationID || !sameRelayMetadataCWD(prepared.WorkspaceRoot, original.Controller.WorkspaceRoot) {
		return nativeError(apierrors.ScopeDenied, "prepared input differs from its original owned controller tuple")
	}
	if prepared.ActorID == "" || prepared.ActorKind != "human" && prepared.ActorKind != "agent" || prepared.ActorKind == "agent" && prepared.ActorProjectID != session.ProjectID || prepared.PreparedAt.IsZero() || prepared.PreparedAt.After(time.Now()) {
		return nativeError(apierrors.EvidenceMissing, "prepared input lacks its actual actor scope or preparation time")
	}
	receipt, exists := session.Operations[prepared.OperationID]
	if !exists || receipt.Action != "send" || receipt.Status != "pending" || receipt.CreatedAt.IsZero() || receipt.CreatedAt.After(prepared.PreparedAt) || receipt.FinishedAt != nil || session.PendingOperation != prepared.OperationID || session.LastOperationID != prepared.OperationID {
		return nativeError(apierrors.DeliveryUnknown, "prepared input has no original pending send receipt")
	}
	packet := session.ContextPacket
	if packet.ID == "" || packet.SchemaVersion <= 0 || packet.SettingsRevision <= 0 || packet.ProjectID != session.ProjectID || packet.CreatedAt.IsZero() || !reflect.DeepEqual(preparedContextMetadata(packet), preparedContextMetadata(original.Session.ContextPacket)) {
		return nativeError(apierrors.ContextStale, "prepared input fixed native context differs from the original session")
	}
	return nil
}

func preparedContextMetadata(packet domain.ContextPacket) domain.ContextPacket {
	packet.Materials = append([]domain.ContextMaterial(nil), packet.Materials...)
	for i := range packet.Materials {
		packet.Materials[i].Text = ""
	}
	packet.Files = append([]domain.ContextFile(nil), packet.Files...)
	for i := range packet.Files {
		packet.Files[i].Text = ""
	}
	return packet
}
