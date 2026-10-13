package app

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayOwnedSessionReader = (*relayService)(nil)

func (s *relayService) attachOwnedSession(st *RelayState, task domain.RelayTask, workspace domain.RelayWorkspace, terminal domain.RelayTerminal) {
	if st.OwnedSessions == nil {
		st.OwnedSessions = map[string]domain.RelayOwnedSession{}
	}
	if _, exists := st.OwnedSessions[task.ID]; exists {
		return
	}
	st.OwnedSessions[task.ID] = domain.RelayOwnedSession{
		BindingID: uuid.NewString(), TaskID: task.ID, TerminalID: terminal.ID,
		WorktreeID: workspace.ID, ProjectID: task.ProjectID, Agent: task.Agent,
		DriveMode: terminal.DriveMode, WorkspaceRoot: workspace.Root,
		SourceRoot: workspace.SourceRoot, SourceStatus: "unknown", Gap: "archive_not_selected",
		ArchiveFinalized: "unknown", ArchiveAppendOnly: "unknown",
	}
}

func relayArchiveMatches(archive domain.RelaySessionArchive, bound domain.RelayOwnedSession) bool {
	if archive.SessionID != bound.SessionID || archive.NativeID != bound.NativeID || !sameRelayRoot(archive.WorkspaceRoot, bound.WorkspaceRoot) || archive.Format == "" || !filepath.IsAbs(archive.Path) || !filepath.IsAbs(archive.Root) {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(archive.Root), filepath.Clean(archive.Path))
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func (s *relayService) observeOwnedSession(st *RelayState, task domain.RelayTask, terminal domain.RelayTerminal, obs RelayObservation) {
	bound, ok := st.OwnedSessions[task.ID]
	if !ok {
		return
	} // old states remain an explicit gap; read never backfills
	bound.DriveMode, bound.ExitedAt = terminal.DriveMode, terminal.ExitedAt
	if obs.Session.ID != "" {
		session := obs.Session
		if bound.SessionID == "" && (session.Ownership != "owned" || session.NativeID == "") {
			bound.Ownership = session.Ownership
			bound.SourceStatus, bound.Gap = "gap", "session_ownership_unverified"
		} else if bound.SessionID != "" && (bound.SessionID != session.ID || bound.NativeID != "" && bound.NativeID != session.NativeID) {
			bound.SourceStatus, bound.Gap = "gap", "session_identity_changed"
		} else {
			bound.SessionID, bound.NativeID, bound.Ownership = session.ID, session.NativeID, session.Ownership
			if session.ProjectID != task.ProjectID || session.CLI != task.Agent || session.Ownership != "owned" || session.NativeID == "" {
				bound.SourceStatus, bound.Gap = "gap", "session_ownership_unverified"
			} else if bound.SourceStatus == "unknown" || bound.Gap == "session_ownership_unverified" {
				bound.SourceStatus, bound.Gap = "unknown", "archive_not_selected"
			}
		}
	}
	if obs.Archive != nil && bound.Ownership == "owned" && bound.SessionID != "" && bound.NativeID != "" && bound.Gap != "session_identity_changed" {
		if !relayArchiveMatches(*obs.Archive, bound) || bound.ArchivePath != "" && (bound.ArchivePath != obs.Archive.Path || bound.ArchiveRoot != obs.Archive.Root || bound.ArchiveFormat != obs.Archive.Format) {
			bound.SourceStatus, bound.Gap = "gap", "archive_identity_changed"
		} else if bound.Gap != "archive_identity_changed" {
			bound.ArchivePath, bound.ArchiveRoot, bound.ArchiveFormat = obs.Archive.Path, obs.Archive.Root, obs.Archive.Format
			bound.SourceStatus, bound.Gap = "available", ""
			if obs.Archive.Finalized {
				bound.ArchiveFinalized = "confirmed"
			}
			if obs.Archive.AppendOnly {
				bound.ArchiveAppendOnly = "confirmed"
			}
		}
	}
	st.OwnedSessions[task.ID] = bound
}

func relayBindingGap(bound domain.RelayOwnedSession, reason string) domain.RelayOwnedSession {
	bound.SourceStatus, bound.Gap = "gap", reason
	bound.ArchiveFinalized, bound.ArchiveAppendOnly = "unknown", "unknown"
	bound.WorkspaceRoot, bound.SourceRoot, bound.SessionID, bound.NativeID, bound.Ownership = "", "", "", "", ""
	bound.ArchivePath, bound.ArchiveRoot, bound.ArchiveFormat = "", "", ""
	return bound
}

// Filesystem checks cover only the registered source and neutral root. No
// archive is opened and no ancestor is enumerated for session discovery.
func (s *relayService) existingOwnedRoot(root string) bool {
	if s.options.OwnedRoots == nil {
		return false
	}
	resolved, err := s.options.OwnedRoots.CanonicalRoot(root)
	return err == nil && sameRelayRoot(resolved, root)
}

func (s *relayService) ListOwnedRelaySessions(ctx context.Context) ([]domain.RelayOwnedSession, error) {
	st, err := s.store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.RelayOwnedSession, 0, len(st.OwnedSessions))
	for _, task := range st.Tasks {
		if task.TerminalID == "" {
			continue
		}
		bound, exists := st.OwnedSessions[task.ID]
		if !exists {
			result = append(result, domain.RelayOwnedSession{TaskID: task.ID, TerminalID: task.TerminalID, WorktreeID: task.WorktreeID, ProjectID: task.ProjectID, Agent: task.Agent, SourceStatus: "gap", Gap: "binding_not_recorded"})
			continue
		}
		p, permissionErr := s.permission(ctx, domain.Caller{}, task, "observe")
		if permissionErr != nil {
			result = append(result, relayBindingGap(bound, "permission_denied"))
			continue
		}
		workspace, ok := st.Workspaces[task.WorktreeID]
		terminal, terminalOK := st.Terminals[task.TerminalID]
		if !ok || !terminalOK || !workspace.Prepared || workspace.ID != bound.WorktreeID || terminal.WorktreeID != bound.WorktreeID || terminal.Agent != bound.Agent || bound.TaskID != task.ID || bound.TerminalID != task.TerminalID || bound.ProjectID != task.ProjectID || bound.Agent != task.Agent || !sameRelayRoot(p.Root, bound.SourceRoot) || !sameRelayRoot(workspace.SourceRoot, bound.SourceRoot) || !sameRelayRoot(workspace.Root, bound.WorkspaceRoot) {
			result = append(result, relayBindingGap(bound, "workspace_identity_changed"))
			continue
		}
		if !s.existingOwnedRoot(bound.SourceRoot) || !s.existingOwnedRoot(bound.WorkspaceRoot) {
			result = append(result, relayBindingGap(bound, "workspace_unavailable"))
			continue
		}
		if bound.Ownership != "owned" || bound.SessionID == "" || bound.NativeID == "" {
			result = append(result, relayBindingGap(bound, "session_ownership_unverified"))
			continue
		}
		// Closed bindings survive terminal reuse. Active bindings must still name
		// the exact controller session and writer captured by this Service.
		if bound.ExitedAt == nil {
			session := st.Sessions[bound.TerminalID]
			if workspace.ActiveTaskID != task.ID || session.ID != bound.SessionID || session.NativeID != bound.NativeID || session.Ownership != "owned" {
				result = append(result, relayBindingGap(bound, "session_identity_changed"))
				continue
			}
		}
		if bound.SourceStatus == "gap" {
			bound = relayBindingGap(bound, bound.Gap)
		} else if bound.SourceStatus != "available" {
			bound.ArchivePath, bound.ArchiveRoot, bound.ArchiveFormat = "", "", ""
			bound.ArchiveFinalized, bound.ArchiveAppendOnly = "unknown", "unknown"
		} else if !relayArchiveMatches(domain.RelaySessionArchive{SessionID: bound.SessionID, NativeID: bound.NativeID, WorkspaceRoot: bound.WorkspaceRoot, Path: bound.ArchivePath, Root: bound.ArchiveRoot, Format: bound.ArchiveFormat}, bound) {
			bound = relayBindingGap(bound, "archive_identity_changed")
		}
		result = append(result, bound)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].TaskID < result[j].TaskID })
	return result, nil
}
