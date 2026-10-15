package main

import (
	"bytes"
	"context"
	"os"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayGoalVerificationReader struct {
	runtime  workspace.RelayRuntime
	owned    workspace.RelayOwnedSessionReader
	evidence *events.RelayGoalEvidenceStore
	admitted *sqlite.AccessSessionReader
	context  *relayPublicContext
	projects workspace.LocalProjectRepository
}

var _ workspace.RelayGoalVerificationReader = (*relayGoalVerificationReader)(nil)

// 读取原对象、已提交事件与当前权限；此端口不应用代码或执行检查命令。
func (r *relayGoalVerificationReader) ReadRelayGoalVerification(ctx context.Context, caller domain.Caller, taskID, evidenceID string) (domain.RelayGoalVerification, error) {
	var result domain.RelayGoalVerification
	denied := func() (domain.RelayGoalVerification, error) {
		return result, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "original owned output and fixed goal checks are unverified"}
	}
	if r.runtime == nil || r.owned == nil || r.evidence == nil || r.admitted == nil || r.context == nil || r.projects == nil {
		return denied()
	}
	view, err := r.runtime.GetTask(ctx, caller, taskID)
	if err != nil {
		return result, err
	}
	evidence, err := r.evidence.Read(ctx, evidenceID)
	if err != nil {
		return denied()
	}
	identity := evidence.Identity
	if identity.TaskID != taskID || identity.ProjectID != view.Task.ProjectID || identity.WorktreeID != view.Task.WorktreeID || identity.TerminalID != view.Task.TerminalID {
		return denied()
	}
	project, err := r.projects.LoadProject(ctx, identity.ProjectID)
	if err != nil {
		return denied()
	}
	currentRoot, err := os.Stat(project.Root)
	if err != nil || !os.SameFile(currentRoot, r.context.identity) {
		return denied()
	}
	packet, err := r.context.projects.ReadProjectContext(ctx, caller, identity.ProjectID, domain.ContextRequest{Files: []string{"stats.go", "stats_test.go"}})
	if err != nil {
		return denied()
	}
	if packet.ProjectID != identity.ProjectID || packet.SettingsRevision != view.Task.SettingsRevision || len(packet.Files) != 2 || len(packet.Materials) != 0 {
		return denied()
	}
	seenFiles := map[string]bool{}
	for _, file := range packet.Files {
		if !bytes.Equal([]byte(file.Text), evidence.AppliedFiles[file.Path]) || file.Version == "" {
			return denied()
		}
		if seenFiles[file.Path] {
			return denied()
		}
		seenFiles[file.Path] = true
	}
	bindings, err := r.owned.ListOwnedRelaySessions(ctx)
	if err != nil {
		return denied()
	}
	for _, binding := range bindings {
		if binding.TaskID != taskID {
			continue
		}
		if binding.ProjectID != identity.ProjectID || binding.WorktreeID != identity.WorktreeID || binding.TerminalID != identity.TerminalID || binding.ControllerID != identity.ControllerID || binding.StartupOperationID != identity.StartupOperationID || binding.SessionID != identity.SessionID || binding.NativeID != identity.NativeID || binding.Agent != "codex" || binding.SourceStatus != "available" || binding.Ownership != "owned" {
			return denied()
		}
		output, err := events.ReadRelayGoalOutput(ctx, relayCaptureBinding(binding), identity.TurnID, identity.AcceptedAt)
		if err != nil || output.Origin != evidence.OutputOrigin || !bytes.Equal(output.Text, evidence.Output) {
			return denied()
		}
		id, err := r.admitted.ReadCommittedRelayGoalOutput(ctx, output.Origin, identity.SessionID, identity.ProjectID, relayGoalProjection(output.Projection))
		if err != nil || id != evidence.OutputEventID {
			return denied()
		}
		return domain.RelayGoalVerification{EvidenceID: evidenceID, TaskID: taskID, WorktreeID: identity.WorktreeID, TerminalID: identity.TerminalID, ControllerID: identity.ControllerID, StartupOperationID: identity.StartupOperationID, SessionID: identity.SessionID, NativeID: identity.NativeID, InputOperationID: identity.InputOperationID, TurnID: identity.TurnID, OutputEventIDs: []string{id}, CheckEvidenceIDs: []string{evidenceID}, PassedAt: evidence.PassedAt}, nil
	}
	return denied()
}

func relayGoalProjection(p events.RelaySessionProjection) workbench.Event {
	var tokens *int64
	if p.Tokens != nil {
		value := int64(*p.Tokens)
		tokens = &value
	}
	return workbench.Event{ID: p.ID, Kind: p.Kind, Topic: p.Topic, Tokens: tokens, Scope: p.Scope, TS: p.TS}
}
