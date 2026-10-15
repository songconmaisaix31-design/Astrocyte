package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relaySourceIdentity struct {
	nativeID, controllerID, operationID, fileID string
}

// 身份仅保留在本进程内存中，重启后不能据此重建控制器。
type relaySourceGuard struct {
	mu       sync.Mutex
	registry workspace.NativeRegistry
	selected map[string]relaySourceIdentity
}

func newRelaySourceGuard(registry workspace.NativeRegistry) *relaySourceGuard {
	return &relaySourceGuard{registry: registry, selected: map[string]relaySourceIdentity{}}
}

func relaySourceDenied(code apierrors.Code, reason string) error {
	return &apierrors.ServiceError{Code: code, Message: reason, RequiredAction: "inspect_original_owned_source_without_resending"}
}

// 预查和首输入使用同一检查；只读取原Registry已持有的元数据与正选文件身份。
func (g *relaySourceGuard) Check(ctx context.Context, session domain.NativeSession, archive *domain.RelaySessionArchive, gap string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.registry == nil || session.ID == "" || session.NativeID == "" || session.ProjectID == "" || session.CLI != "codex" || session.Ownership != "owned" {
		return relaySourceDenied(apierrors.ScopeDenied, "original_controller_identity_unverified")
	}
	adapter, err := g.registry.Adapter(session.CLI)
	if err != nil {
		return err
	}
	reader, ok := adapter.(interface {
		OwnedRelayControllerEvidence(context.Context, domain.NativeSession) (domain.RelayControllerEvidence, error)
		OwnedRelayExecution(context.Context, domain.NativeSession) (agents.RelayOwnedExecution, error)
	})
	if !ok {
		return relaySourceDenied(apierrors.UnsupportedCapability, "original_controller_metadata_unsupported")
	}
	evidence, err := reader.OwnedRelayControllerEvidence(ctx, session)
	if err != nil {
		return err
	}
	if evidence.SessionID != session.ID || evidence.NativeID != session.NativeID || evidence.Provider != session.CLI || evidence.ControllerID == "" || evidence.StartupOperationID == "" || evidence.ExitConfirmed {
		return relaySourceDenied(apierrors.ScopeDenied, "original_controller_identity_changed")
	}
	execution, err := reader.OwnedRelayExecution(ctx, session)
	if err != nil {
		return err
	}
	if execution.ExitConfirmed || execution.ProcessID < 1 || execution.MetadataStatus != "available" || execution.LaunchPolicy != "selected_context_tools_disabled_v1" || execution.PermissionPolicy != "never_read_only" {
		return relaySourceDenied(apierrors.EvidenceMissing, "original_tool_closure_unverified")
	}
	if archive == nil || gap != "" {
		return relaySourceDenied(apierrors.EvidenceMissing, "original_archive_not_yet_available")
	}
	if archive.SessionID != session.ID || archive.NativeID != session.NativeID || archive.Format != "codex" || !filepath.IsAbs(archive.Path) || !filepath.IsAbs(archive.Root) || !filepath.IsAbs(archive.WorkspaceRoot) {
		return relaySourceDenied(apierrors.ScopeDenied, "original_archive_identity_changed")
	}
	if _, err := os.Lstat(archive.Path); errors.Is(err, os.ErrNotExist) {
		return relaySourceDenied(apierrors.EvidenceMissing, "original_archive_not_yet_available")
	} else if err != nil {
		return relaySourceDenied(apierrors.EvidenceMissing, "original_archive_unavailable")
	}
	workspaceInfo, err := os.Stat(evidence.WorkspaceRoot)
	if err != nil || !workspaceInfo.IsDir() {
		return relaySourceDenied(apierrors.EvidenceMissing, "original_workspace_unavailable")
	}
	archiveWorkspace, err := os.Stat(archive.WorkspaceRoot)
	if err != nil || !os.SameFile(workspaceInfo, archiveWorkspace) {
		return relaySourceDenied(apierrors.ScopeDenied, "original_archive_workspace_changed")
	}
	fileID, err := events.VerifyRelaySessionArchive(ctx, archive.Path, archive.WorkspaceRoot, []string{archive.Root}, archive.NativeID, archive.Format)
	if err != nil {
		return err
	}
	current := relaySourceIdentity{session.NativeID, evidence.ControllerID, evidence.StartupOperationID, fileID}
	key := session.ID + ":" + evidence.ControllerID + ":" + evidence.StartupOperationID
	if previous, exists := g.selected[key]; exists && previous != current {
		return relaySourceDenied(apierrors.ScopeDenied, "original_source_identity_changed")
	}
	g.selected[key] = current
	return nil
}
