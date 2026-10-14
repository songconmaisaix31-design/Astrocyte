package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 人工交接声明只引用同项目的安全历史，不读取旧原生上下文或输出。
func (s *relayService) validateHandoff(ctx context.Context, c domain.Caller, p domain.LocalProject, cmd RelayIntentCommand) error {
	if cmd.HandoffFromTaskID == "" && !cmd.HandoffContextConfirmed {
		return nil
	}
	if cmd.HandoffFromTaskID == "" || !cmd.HandoffContextConfirmed {
		return relayError(apierrors.ValidationFailed, "handoff source and human context confirmation are required together")
	}
	if _, err := uuid.Parse(cmd.HandoffFromTaskID); err != nil {
		return relayError(apierrors.ValidationFailed, "handoff source task must be a UUID")
	}
	current := domain.RelayTask{ProjectID: p.ID, Agent: cmd.Agent, SettingsRevision: p.Settings.Revision, PermissionLevel: RelayPermissionLevel(p.Settings)}
	if _, err := s.permission(ctx, c, current, "observe"); err != nil {
		return err
	}
	st, err := s.snapshot(ctx, c)
	if err != nil {
		return err
	}
	return relayHandoffState(st, p, cmd)
}

func relayHandoffState(st RelayState, p domain.LocalProject, cmd RelayIntentCommand) error {
	previous, ok := st.Tasks[cmd.HandoffFromTaskID]
	if !ok {
		return apierrors.NewNotFound("relay task", cmd.HandoffFromTaskID)
	}
	if previous.ProjectID != p.ID || previous.WorkID != cmd.WorkID || previous.Agent == cmd.Agent {
		return relayError(apierrors.ScopeDenied, "handoff requires the same project and work with a different Agent")
	}
	workspace, ownsWorkspace := st.Workspaces[previous.WorktreeID]
	terminal, ownsTerminal := st.Terminals[previous.TerminalID]
	if previous.State != domain.RelayReleased || !ownsWorkspace || !workspace.Prepared || workspace.ProjectID != p.ID || !sameRelayRoot(workspace.SourceRoot, p.Root) || workspace.ActiveTaskID != "" || !ownsTerminal || terminal.WorktreeID != previous.WorktreeID || terminal.Agent != previous.Agent || terminal.ExitedAt == nil || st.PendingActions[previous.ID] != "" {
		return relayError(apierrors.VersionConflict, "handoff source must be released with confirmed exit and no pending action")
	}
	for _, r := range st.Receipts {
		if r.TaskID == previous.ID && r.Status == "pending" {
			return relayError(apierrors.VersionConflict, "handoff source has an unfinished action")
		}
	}
	return nil
}
