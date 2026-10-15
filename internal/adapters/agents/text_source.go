package agents

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 本机诊断沿用原 controller 的身份和目录，源路径不进入产品 DTO 或数据库。
type selectedTextSource struct {
	ProjectID         string              `json:"project_id"`
	JobID             string              `json:"job_id"`
	OperationID       string              `json:"operation_id"`
	SessionID         string              `json:"session_id"`
	NativeID          string              `json:"native_id"`
	ControllerProject string              `json:"controller_project_id"`
	CLI               string              `json:"cli"`
	Phase             string              `json:"phase"`
	TurnID            string              `json:"turn_id,omitempty"`
	SourcePath        string              `json:"source_path,omitempty"`
	SourceCWD         string              `json:"source_cwd,omitempty"`
	SourceSnapshot    string              `json:"source_snapshot,omitempty"`
	SourceStatus      string              `json:"source_status"`
	Execution         RelayOwnedExecution `json:"execution"`
	Stop              *selectedTextStop   `json:"stop,omitempty"`
	ObservedAt        time.Time           `json:"observed_at"`
	file              string
}

type selectedTextStop struct {
	Status     string    `json:"status"`
	Confirmed  bool      `json:"confirmed"`
	ObservedAt time.Time `json:"observed_at"`
	Error      string    `json:"error,omitempty"`
}

// 始终停止原 controller，再保存同一进程的实际退出观测；工具集合保持原观测值。
func (n *Native) finishSelectedText(session domain.NativeSession, source *selectedTextSource) error {
	ctx := context.Background()
	observation, stopErr := n.Stop(ctx, session)
	var failure error
	if stopErr != nil || !observation.StopConfirmed {
		failure = errors.Join(nativeError(apierrors.DeliveryUnknown, "selected-text owned process stop remains unconfirmed"), stopErr)
	}
	if source == nil {
		return failure
	}
	source.Stop = &selectedTextStop{Status: observation.Status, Confirmed: observation.StopConfirmed, ObservedAt: time.Now().UTC()}
	if stopErr != nil {
		source.Stop.Error = stopErr.Error()
	}
	execution, err := n.OwnedRelayExecution(ctx, session)
	if err != nil {
		failure = errors.Join(failure, nativeError(apierrors.EvidenceMissing, "selected-text final execution observation failed"), err)
	} else {
		source.Execution = execution
	}
	// 即使原档保留失败，本次实际停止结果仍须先保存。
	if err := source.save(); err != nil {
		return errors.Join(failure, nativeError(apierrors.EvidenceMissing, "selected-text final source retention failed"), err)
	}
	if err := source.retainOriginal(); err != nil {
		return errors.Join(failure, nativeError(apierrors.EvidenceMissing, "selected-text original source retention failed"), err)
	}
	return failure
}

func (n *Native) selectedTextSource(ctx context.Context, session domain.NativeSession, input domain.TextRequest) (*selectedTextSource, error) {
	execution, err := n.OwnedRelayExecution(ctx, session)
	if err != nil {
		return nil, err
	}
	record := &selectedTextSource{ProjectID: input.ProjectID, JobID: input.JobID, OperationID: input.OperationID, SessionID: session.ID, NativeID: session.NativeID, ControllerProject: session.ProjectID, CLI: session.CLI, Phase: "started", SourceStatus: "unknown", Execution: execution}
	record.file = filepath.Join(n.stateRoot, session.ProjectID, n.id, "selected-text-source.json")
	if n.id == "codex" {
		p, err := n.get(session)
		if err != nil {
			return nil, err
		}
		response, err := p.call(ctx, "thread/read", map[string]any{"threadId": session.NativeID, "includeTurns": false}, false)
		if err != nil {
			return nil, err
		}
		// 仅保存当前精确 controller 返回的原 metadata，不扫描历史目录。
		raw, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		if err = os.WriteFile(filepath.Join(filepath.Dir(record.file), "selected-text-thread-read.json"), raw, 0600); err != nil {
			return nil, err
		}
		var result struct {
			Thread struct {
				ID   string  `json:"id"`
				CWD  string  `json:"cwd"`
				Path *string `json:"path"`
			} `json:"thread"`
		}
		if err = json.Unmarshal(response["result"], &result); err != nil {
			return nil, nativeError(apierrors.EvidenceMissing, "owned thread metadata cannot be decoded")
		}
		p.mu.Lock()
		cwd := p.cmd.Dir
		p.mu.Unlock()
		if result.Thread.ID != session.NativeID || filepath.Clean(result.Thread.CWD) != filepath.Clean(cwd) || result.Thread.Path == nil || !filepath.IsAbs(*result.Thread.Path) {
			return nil, nativeError(apierrors.EvidenceMissing, "owned thread metadata does not select its original source")
		}
		// 原生 writer 在首次 turn 后创建文件；当前回执只确认原 controller 的路径。
		record.SourcePath, record.SourceStatus = *result.Thread.Path, "owned_metadata_pointer_observed"
		record.SourceCWD = cwd
	}
	if err = record.save(); err != nil {
		return nil, err
	}
	return record, nil
}

func verifySelectedTextHeader(path, nativeID, cwd string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > 32*1024*1024 {
		return nativeError(apierrors.EvidenceMissing, "owned source must be an ordinary bounded native file")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	var header struct {
		Type    string `json:"type"`
		Payload struct {
			ID  string `json:"id"`
			CWD string `json:"cwd"`
		} `json:"payload"`
	}
	parseErr := json.NewDecoder(io.LimitReader(file, 32*1024*1024)).Decode(&header)
	if err = errors.Join(parseErr, file.Close()); err != nil {
		return nativeError(apierrors.EvidenceMissing, "owned source header cannot be decoded")
	}
	if header.Type != "session_meta" || header.Payload.ID != nativeID || filepath.Clean(header.Payload.CWD) != filepath.Clean(cwd) {
		return nativeError(apierrors.ScopeDenied, "native source header differs from its original owned thread")
	}
	return nil
}

func (s *selectedTextSource) save() error {
	s.ObservedAt = time.Now().UTC()
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, body, 0600)
}

func (s *selectedTextSource) retainOriginal() error {
	if s.SourcePath == "" || s.Phase == "started" {
		return s.save()
	}
	info, err := os.Lstat(s.SourcePath)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 32*1024*1024 {
		return nativeError(apierrors.EvidenceMissing, "owned native source became unavailable")
	}
	if err = verifySelectedTextHeader(s.SourcePath, s.NativeID, s.SourceCWD); err != nil {
		return err
	}
	body, err := os.ReadFile(s.SourcePath)
	if err != nil {
		return err
	}
	s.SourceSnapshot = filepath.Join(filepath.Dir(s.file), "selected-text-original.jsonl")
	if err = os.WriteFile(s.SourceSnapshot, body, 0600); err != nil {
		return err
	}
	// 原件只声明本次实际保留窗口，完整度由原生读取和具体消费者另行核查。
	s.SourceStatus = "owned_source_window_retained"
	return s.save()
}
