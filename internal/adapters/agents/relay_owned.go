package agents

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// RelayOwnedRoots reuses existing project canonicalization and redirect checks.
// It reads directory identity only, never session files or private home roots.
type RelayOwnedRoots struct{}

var _ app.RelayOwnedRoots = RelayOwnedRoots{}

func (RelayOwnedRoots) CanonicalRoot(root string) (string, error) {
	if hasRedirectedAncestor(root) {
		return "", scopeError("registered relay root is redirected")
	}
	return CanonicalRoot(root)
}

var _ app.RelayOwnedArchiveProvider = (*Native)(nil)

// OwnedRelayArchive reads metadata through an already-owned controller only.
// Codex 0.162's public ThreadRead schema declares nullable Thread.path and CWD.
// No turns/items/list, home discovery, archive contents, or model call occurs.
func (n *Native) OwnedRelayArchive(ctx context.Context, session domain.NativeSession) (domain.RelaySessionArchive, error) {
	var archive domain.RelaySessionArchive
	if err := ctx.Err(); err != nil {
		return archive, err
	}
	if session.Ownership != "owned" || session.ID == "" || session.NativeID == "" || session.ProjectID == "" || session.CLI != n.id {
		return archive, nativeError(apierrors.ScopeDenied, "native archive metadata requires the exact owned controller session")
	}
	if n.id != "codex" {
		return archive, nativeError(apierrors.UnsupportedCapability, "this native protocol has no verified metadata-only archive path and CWD operation")
	}
	p, err := n.get(session)
	if err != nil {
		return archive, err
	}
	if p.relayCapture != nil {
		capture := p.relayCapture
		capture.mu.Lock()
		path, nativeID, workspace := capture.path, capture.nativeID, capture.workspace
		capture.mu.Unlock()
		if nativeID != session.NativeID || !sameRelayMetadataCWD(workspace, p.cmd.Dir) {
			return archive, nativeError(apierrors.ScopeDenied, "captured native source identity differs")
		}
		if path == "" {
			return archive, nativeError(apierrors.EvidenceMissing, "native start response did not select an archive")
		}
		closed := capture.snapshot(p.observe().StopConfirmed).Finalized == "confirmed"
		return domain.RelaySessionArchive{SessionID: session.ID, NativeID: nativeID, WorkspaceRoot: workspace, Path: path, Root: filepath.Dir(path), Format: "codex", Finalized: closed}, nil
	}
	if p.observe().StopConfirmed {
		return archive, nativeError(apierrors.EvidenceMissing, "stopped native controller cannot resolve new archive metadata; retain its previously bound source")
	}
	response, err := p.call(ctx, "thread/read", map[string]any{"threadId": session.NativeID, "includeTurns": false}, false)
	if err != nil {
		return archive, err
	}
	// 只解码身份元数据，原始 preview 和 turns 不进入状态、日志或事件。
	var result struct {
		Thread struct {
			ID   string  `json:"id"`
			Cwd  string  `json:"cwd"`
			Path *string `json:"path"`
		} `json:"thread"`
	}
	if err = json.Unmarshal(response["result"], &result); err != nil {
		return archive, nativeError(apierrors.InternalError, "owned archive metadata response is invalid")
	}
	if result.Thread.ID != session.NativeID || !filepath.IsAbs(result.Thread.Cwd) || !sameRelayMetadataCWD(result.Thread.Cwd, p.cmd.Dir) {
		return archive, nativeError(apierrors.ScopeDenied, "owned archive metadata session or captured CWD differs")
	}
	if result.Thread.Path == nil || !filepath.IsAbs(*result.Thread.Path) || !strings.EqualFold(filepath.Ext(*result.Thread.Path), ".jsonl") {
		return archive, nativeError(apierrors.EvidenceMissing, "owned metadata does not identify a native JSONL archive path")
	}
	current, err := n.get(session)
	if err != nil || current != p {
		return archive, nativeError(apierrors.DeliveryUnknown, "native controller binding changed during metadata observation")
	}
	path := filepath.Clean(*result.Thread.Path)
	return domain.RelaySessionArchive{
		SessionID: session.ID, NativeID: session.NativeID, WorkspaceRoot: p.cmd.Dir,
		Path: path, Root: filepath.Dir(path), Format: "codex",
		// Parent is containment for this exact metadata-selected file, never a
		// directory discovery scope or permission to collect adjacent sessions.
		// Neither process exit nor the protocol path proves final/append-only.
	}, nil
}

func sameRelayMetadataCWD(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
