package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const (
	OwnedPurposeBind   = "bind_session"
	OwnedPurposeClaim  = "claim_work_item"
	OwnedPurposeSubmit = "submit_artifact"
)

type ownedSessionProvider struct {
	reader    OwnedNativeSessionReader
	snapshots ProjectSnapshotProvider
}

func NewOwnedSessionProvider(reader OwnedNativeSessionReader, snapshots ProjectSnapshotProvider) (OwnedSessionProvider, error) {
	if reader == nil || snapshots == nil {
		return nil, NormalizeWorkspaceError(domain.WorkspaceError(apierrors.UnsupportedCapability, "actual Registry reader and authorized snapshot provider are required"))
	}
	return &ownedSessionProvider{reader: reader, snapshots: snapshots}, nil
}

func (p *ownedSessionProvider) currentOwnedRecord(ctx context.Context, c domain.Caller, id, purpose string) (record domain.OwnedNativeSessionRecord, err error) {
	if !has([]string{OwnedPurposeBind, OwnedPurposeClaim, OwnedPurposeSubmit}, purpose) {
		return record, domain.WorkspaceError(apierrors.ValidationFailed, "owned session purpose is required")
	}
	if c.ID == "" || (c.Kind != "human" && c.Kind != "agent") {
		return record, domain.WorkspaceError(apierrors.ScopeDenied, "authenticated project caller is required")
	}
	record, err = p.reader.ReadOwnedNativeSessionRecord(ctx, id)
	if err != nil {
		return record, err
	}
	session := record.Session
	if session.ID != id || session.ID == "" || session.ProjectID == "" || session.CLI == "" || session.NativeID == "" || session.Version == "" || session.Ownership != "owned" || record.ControllerID == "" || record.StartupOperationID == "" || record.EvidenceSessionID != session.ID || record.EvidenceNativeID != session.NativeID || record.EvidenceAdapterID != session.CLI || record.WorkspaceRoot == "" {
		return record, domain.WorkspaceError(apierrors.EvidenceMissing, "Registry omitted the original native session and controller tuple")
	}
	if record.Liveness != "live" && record.Liveness != "exited" {
		return record, domain.WorkspaceError(apierrors.EvidenceMissing, "current controller lifecycle is unknown")
	}
	if purpose != OwnedPurposeSubmit && record.Liveness != "live" {
		return record, domain.WorkspaceError(apierrors.EvidenceMissing, "new bindings and claims require a live original controller")
	}
	return record, nil
}

func (p *ownedSessionProvider) ReadOwnedSessionFacts(ctx context.Context, c domain.Caller, id, purpose string) (facts domain.OwnedSessionFacts, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	record, err := p.currentOwnedRecord(ctx, c, id, purpose)
	if err != nil {
		return facts, err
	}
	session := record.Session
	snapshot, err := p.snapshots.CaptureProjectSnapshot(ctx, c, session.ProjectID, []string{record.WorkspaceRoot})
	if err != nil {
		return facts, err
	}
	if snapshot.ProjectID != session.ProjectID || snapshot.Root != record.WorkspaceRoot || snapshot.WorktreeID == "" {
		return facts, domain.WorkspaceError(apierrors.ScopeDenied, "native controller root differs from the authorized Git worktree")
	}
	if err = p.snapshots.ValidateProjectSnapshot(ctx, c, snapshot); err != nil {
		return facts, err
	}
	capabilities := map[string]domain.CapabilityObservation{}
	for name, observation := range record.Capabilities {
		capabilities[name] = observation
	}
	return domain.OwnedSessionFacts{SessionID: session.ID, ProjectID: session.ProjectID, AdapterID: session.CLI, NativeID: session.NativeID, WorktreeID: snapshot.WorktreeID, Root: snapshot.Root, Owned: true, ControllerID: record.ControllerID, StartupOperationID: record.StartupOperationID, Liveness: record.Liveness, Capabilities: capabilities, Snapshot: snapshot}, nil
}

func (p *ownedSessionProvider) ValidateOwnedSessionFacts(ctx context.Context, c domain.Caller, expected domain.OwnedSessionFacts, purpose string) (err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	record, err := p.currentOwnedRecord(ctx, c, expected.SessionID, purpose)
	if err != nil {
		return err
	}
	session := record.Session
	if !expected.Owned || expected.WorktreeID == "" || session.ProjectID != expected.ProjectID || session.CLI != expected.AdapterID || session.NativeID != expected.NativeID || record.ControllerID != expected.ControllerID || record.StartupOperationID != expected.StartupOperationID || record.WorkspaceRoot != expected.Root {
		return domain.WorkspaceError(apierrors.EvidenceMissing, "current Registry controller tuple changed after project observation")
	}
	if expected.Snapshot.ProjectID != expected.ProjectID || expected.Snapshot.WorktreeID != expected.WorktreeID || expected.Snapshot.Root != expected.Root {
		return domain.WorkspaceError(apierrors.ContextStale, "owned session observation omitted its fixed project snapshot")
	}
	return validateProjectSnapshotReference(ctx, p.snapshots, c, expected.Snapshot)
}
