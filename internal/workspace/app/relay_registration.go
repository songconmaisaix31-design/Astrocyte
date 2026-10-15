package app

import (
	"context"
	"reflect"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ RelayTrustedAgentRegistrar = (*relayService)(nil)

func cloneRelayAgents(agents []domain.AgentSpec) []domain.AgentSpec {
	result := make([]domain.AgentSpec, len(agents))
	for i, spec := range agents {
		result[i] = spec
		if spec.LaunchCmd != nil {
			result[i].LaunchCmd = append([]string{}, spec.LaunchCmd...)
		}
		if spec.Capabilities != nil {
			result[i].Capabilities = make(map[string]domain.CapabilityObservation, len(spec.Capabilities))
			for name, observation := range spec.Capabilities {
				if observation.CheckedAt != nil {
					when := *observation.CheckedAt
					observation.CheckedAt = &when
				}
				result[i].Capabilities[name] = observation
			}
		}
	}
	return result
}

func validateRelayAgents(agents []domain.AgentSpec) error {
	seen := make(map[string]bool, len(agents))
	for _, spec := range agents {
		if strings.TrimSpace(spec.Name) == "" || strings.TrimSpace(spec.Name) != spec.Name || seen[spec.Name] || len(spec.LaunchCmd) == 0 || strings.TrimSpace(spec.LaunchCmd[0]) == "" {
			return relayError(apierrors.ValidationFailed, "invalid trusted relay registration")
		}
		if spec.DriveMode != "structured" && spec.DriveMode != "terminal" && spec.DriveMode != "both" {
			return relayError(apierrors.ValidationFailed, "invalid trusted relay drive mode")
		}
		if spec.InputMode != "" && spec.InputMode != "line" && spec.InputMode != "stdin_eof" {
			return relayError(apierrors.ValidationFailed, "invalid trusted relay input mode")
		}
		for _, argument := range spec.LaunchCmd {
			if strings.ContainsRune(argument, '\x00') {
				return relayError(apierrors.ValidationFailed, "invalid trusted relay launch argument")
			}
		}
		seen[spec.Name] = true
	}
	return nil
}

func (s *relayService) RegisterTrustedAgents(ctx context.Context, agents []domain.AgentSpec) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	items := cloneRelayAgents(agents)
	if err := validateRelayAgents(items); err != nil {
		return err
	}
	if s.store == nil {
		return relayError(apierrors.UnsupportedCapability, "relay store is not connected")
	}
	s.agentsMu.Lock()
	defer s.agentsMu.Unlock()
	previous := make(map[string]domain.AgentSpec, len(s.options.Agents))
	for _, spec := range s.options.Agents {
		previous[spec.Name] = spec
	}
	changed := make(map[string]bool)
	for _, spec := range items {
		if old, ok := previous[spec.Name]; !ok || !reflect.DeepEqual(old, spec) {
			changed[spec.Name] = true
		}
		delete(previous, spec.Name)
	}
	for name := range previous {
		changed[name] = true
	}
	// 与 begin 使用同一事务边界，避免读取完成后才出现新的在途动作。
	if len(changed) > 0 {
		if err := s.store.Update(ctx, func(st *RelayState) error {
			for _, task := range st.Tasks {
				if !changed[task.Agent] {
					continue
				}
				terminal := st.Terminals[task.TerminalID]
				workspace := st.Workspaces[task.WorktreeID]
				if task.State != domain.RelayReleased || st.PendingActions[task.ID] != "" || workspace.ActiveTaskID == task.ID || (st.Sessions[task.TerminalID].Ownership == "owned" && terminal.ExitedAt == nil) {
					return relayError(apierrors.VersionConflict, "relay registration is reserved by an unfinished task")
				}
			}
			for _, receipt := range st.Receipts {
				if receipt.Status != "pending" && receipt.Status != "delivery_unknown" {
					continue
				}
				task, found := st.Tasks[receipt.TaskID]
				if !found || changed[task.Agent] {
					return relayError(apierrors.VersionConflict, "relay registration has an unresolved operation")
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.options.Agents = items
	return nil
}
