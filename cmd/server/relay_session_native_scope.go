package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 在原来源 UUID 下区分控制器连接，避免恢复后的序号套用旧游标。
func relayNativeConnectionBinding(binding, controller, operation string) (string, error) {
	for _, value := range []string{binding, controller, operation} {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || id.String() != value {
			return "", errors.New("native connection identity invalid")
		}
	}
	data, err := json.Marshal([]string{controller, operation})
	if err != nil {
		return "", errors.New("native connection identity unavailable")
	}
	return uuid.NewSHA1(uuid.MustParse(binding), data).String(), nil
}

func relayNativeCheckpointIdentity(binding, controller, operation string) bool {
	if controller == "" && operation == "" {
		return true
	}
	_, err := relayNativeConnectionBinding(binding, controller, operation)
	return err == nil
}

// 历史只选择已提交的来源；权限仍由原 runtime 重核，不授予进程所有权。
func (c *relaySessionCollector) nativeHistoricalBindings(ctx context.Context, owned domain.RelayOwnedSession, entry relaySessionEntry) ([]string, error) {
	result := []string{}
	if len(entry.NativeHistory) == 0 {
		return result, nil
	}
	reader, ok := c.owned.(workspace.RelayControllerHistoryReader)
	if !ok {
		return nil, errors.New("native history unavailable")
	}
	history, err := reader.ListOwnedControllerHistory(ctx, owned.TaskID)
	if err != nil {
		return nil, errors.New("native history unavailable")
	}
	for binding, checkpoint := range entry.NativeHistory {
		if entry.Native != nil && binding == entry.Native.BindingID {
			continue
		}
		// 旧记录仅保留原来源选择，缺字段不会推导出连接或当前归属。
		if checkpoint.ControllerID == "" && checkpoint.StartupOperationID == "" && binding == owned.BindingID {
			result = append(result, binding)
			continue
		}
		expected, err := relayNativeConnectionBinding(owned.BindingID, checkpoint.ControllerID, checkpoint.StartupOperationID)
		if err != nil || expected != binding {
			return nil, errors.New("native historical checkpoint invalid")
		}
		matched := false
		for _, evidence := range history {
			if evidence.ControllerID == checkpoint.ControllerID && evidence.StartupOperationID == checkpoint.StartupOperationID && evidence.SessionID == owned.SessionID && evidence.NativeID == owned.NativeID && evidence.Provider == owned.Agent && strings.EqualFold(filepath.Clean(evidence.WorkspaceRoot), filepath.Clean(owned.WorkspaceRoot)) {
				matched = true
			}
		}
		if !matched {
			return nil, errors.New("native historical source unverified")
		}
		result = append(result, binding)
	}
	sort.Strings(result)
	return result, nil
}
