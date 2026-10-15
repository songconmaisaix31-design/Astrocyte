package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 只在本机对象与事件中保存原执行关系，禁止作为 HTTP DTO。
type RelayGoalIdentity struct {
	TaskID, ProjectID, WorktreeID, TerminalID string
	ControllerID, StartupOperationID          string
	SessionID, NativeID, InputOperationID     string
	TurnID                                    string
	AcceptedAt                                time.Time
}

// 此输入只来自受信宿主的实际命令结果，没有 HTTP 写入口。
type RelayGoalHostCheck struct {
	Name                  string
	StartedAt, FinishedAt time.Time
	ExitCode              int
	Stdout, Stderr        []byte `json:"-"`
	Contract              []byte `json:"-"`
}

type relayGoalCheckRecord struct {
	Name                  string
	StartedAt, FinishedAt time.Time
	ExitCode              int
	StdoutRef, StderrRef  string
	ContractRef           string
}

type relayGoalRecord struct {
	SchemaVersion int
	Identity      RelayGoalIdentity
	OutputSource  string
	AppliedBy     string
	OutputEventID string
	OutputOrigin  string
	OutputRef     string
	AppliedFiles  map[string]string
	Checks        []relayGoalCheckRecord
}

type RelayGoalEvidence struct {
	Identity      RelayGoalIdentity `json:"-"`
	EvidenceID    string            `json:"-"`
	OutputEventID string            `json:"-"`
	OutputOrigin  string            `json:"-"`
	Output        []byte            `json:"-"`
	AppliedFiles  map[string][]byte `json:"-"`
	PassedAt      time.Time         `json:"-"`
}

// 沿原对象与事件口读写真实关系，不另建数据库或公共命令。
type RelayGoalEvidenceStore struct {
	objects  attention.ObjectStore
	recorder RelaySessionRecorder
	events   attention.EventReader
}

func NewRelayGoalEvidenceStore(objects attention.ObjectStore, recorder RelaySessionRecorder, events attention.EventReader) *RelayGoalEvidenceStore {
	return &RelayGoalEvidenceStore{objects: objects, recorder: recorder, events: events}
}

func goalEvidenceFailure() error {
	return errors.New("original relay output or host checks unavailable")
}

func validGoalIdentity(v RelayGoalIdentity) bool {
	for _, value := range []string{v.TaskID, v.ProjectID, v.WorktreeID, v.TerminalID, v.ControllerID, v.StartupOperationID, v.SessionID, v.InputOperationID} {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || id.String() != value {
			return false
		}
	}
	return v.NativeID != "" && v.TurnID != "" && !v.AcceptedAt.IsZero()
}

func validGoalChecks(checks []relayGoalCheckRecord, acceptedAt time.Time, requirePass bool) bool {
	names := []string{"go_test", "go_build", "git_diff_check"}
	if len(checks) == 0 || len(checks) > len(names) || requirePass && len(checks) != len(names) {
		return false
	}
	previous := acceptedAt
	for i, check := range checks {
		if check.Name != names[i] || check.ExitCode != 0 && (requirePass || i != len(checks)-1) || check.StartedAt.Before(previous) || check.FinishedAt.Before(check.StartedAt) || check.FinishedAt.After(time.Now()) || check.StdoutRef == "" || check.StderrRef == "" || i == 0 && check.ContractRef == "" {
			return false
		}
		previous = check.FinishedAt
	}
	return true
}

// 调用方先核原 SQL 输出 ID、用文件编辑工具应用原输出，再传入实际检查结果。
func (s *RelayGoalEvidenceStore) Record(ctx context.Context, identity RelayGoalIdentity, output RelayGoalOutput, outputEventID string, files map[string][]byte, checks []RelayGoalHostCheck) (string, error) {
	if s.objects == nil || s.recorder == nil || s.events == nil || !validGoalIdentity(identity) || len(output.Text) == 0 || output.Origin == "" || outputEventID == "" || len(files) != 2 || len(files["stats.go"]) == 0 || len(files["stats_test.go"]) == 0 {
		return "", goalEvidenceFailure()
	}
	original, err := s.events.GetEvent(ctx, outputEventID)
	projection, encodeErr := json.Marshal(output.Projection)
	if err != nil || encodeErr != nil || original.ID != outputEventID || original.OriginKey != output.Origin || original.Observation.Session != identity.SessionID || original.Observation.Scope != identity.ProjectID || original.Observation.Source != "codex" || original.Observation.SourceKind != output.Projection.Kind || original.Observation.Provenance != "workspace:relay_session" || !bytes.Equal(original.RawBytes, projection) {
		return "", goalEvidenceFailure()
	}
	record := relayGoalRecord{SchemaVersion: 1, Identity: identity, OutputSource: "owned_codex_archive", AppliedBy: "trusted_host_file_editor", OutputEventID: outputEventID, OutputOrigin: output.Origin, AppliedFiles: map[string]string{}}
	if record.OutputRef, err = s.objects.Publish(ctx, output.Text); err != nil {
		return "", goalEvidenceFailure()
	}
	for _, name := range []string{"stats.go", "stats_test.go"} {
		if record.AppliedFiles[name], err = s.objects.Publish(ctx, files[name]); err != nil {
			return "", goalEvidenceFailure()
		}
	}
	for _, check := range checks {
		row := relayGoalCheckRecord{Name: check.Name, StartedAt: check.StartedAt, FinishedAt: check.FinishedAt, ExitCode: check.ExitCode}
		if row.StdoutRef, err = s.objects.Publish(ctx, check.Stdout); err != nil {
			return "", goalEvidenceFailure()
		}
		if row.StderrRef, err = s.objects.Publish(ctx, check.Stderr); err != nil {
			return "", goalEvidenceFailure()
		}
		if check.Name == "go_test" {
			if len(check.Contract) == 0 {
				return "", goalEvidenceFailure()
			}
			if row.ContractRef, err = s.objects.Publish(ctx, check.Contract); err != nil {
				return "", goalEvidenceFailure()
			}
		}
		record.Checks = append(record.Checks, row)
	}
	if !validGoalChecks(record.Checks, identity.AcceptedAt, false) {
		return "", goalEvidenceFailure()
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return "", goalEvidenceFailure()
	}
	when := record.Checks[len(record.Checks)-1].FinishedAt
	outcome := "failure"
	if validGoalChecks(record.Checks, identity.AcceptedAt, true) {
		outcome = "success"
	}
	rows, err := s.recorder.Record(ctx, []attention.StoredEvent{{OriginKey: "relay-goal-check:" + uuid.NewString(), RawBytes: payload, ReceivedAt: time.Now().UnixMilli(), Observation: attention.Observation{Source: "relay_host", SourceKind: "goal_check", Session: identity.SessionID, Scope: identity.ProjectID, TS: when.UnixMilli(), Outcome: outcome, Depth: "unknown", Provenance: "workspace:relay_trusted_application_and_checks"}}})
	if err != nil || len(rows) != 1 || rows[0].ID == "" {
		return "", goalEvidenceFailure()
	}
	return rows[0].ID, nil
}

func (s *RelayGoalEvidenceStore) Read(ctx context.Context, id string) (RelayGoalEvidence, error) {
	var result RelayGoalEvidence
	if s.objects == nil || s.events == nil || id == "" {
		return result, goalEvidenceFailure()
	}
	row, err := s.events.GetEvent(ctx, id)
	if err != nil || row.ID != id || row.Event != nil || row.Observation.Source != "relay_host" || row.Observation.SourceKind != "goal_check" || row.Observation.Provenance != "workspace:relay_trusted_application_and_checks" || row.Observation.Outcome != "success" {
		return result, goalEvidenceFailure()
	}
	var record relayGoalRecord
	decoder := json.NewDecoder(bytes.NewReader(row.RawBytes))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&record) != nil || decoder.Decode(new(any)) != io.EOF || record.SchemaVersion != 1 || record.OutputSource != "owned_codex_archive" || record.AppliedBy != "trusted_host_file_editor" || !validGoalIdentity(record.Identity) || record.Identity.ProjectID != row.Observation.Scope || record.Identity.SessionID != row.Observation.Session || !validGoalChecks(record.Checks, record.Identity.AcceptedAt, true) || len(record.AppliedFiles) != 2 {
		return result, goalEvidenceFailure()
	}
	when := record.Checks[len(record.Checks)-1].FinishedAt
	if when.UnixMilli() != row.Observation.TS {
		return result, goalEvidenceFailure()
	}
	result = RelayGoalEvidence{Identity: record.Identity, EvidenceID: id, OutputEventID: record.OutputEventID, OutputOrigin: record.OutputOrigin, AppliedFiles: map[string][]byte{}, PassedAt: when}
	if result.Output, err = s.objects.Read(ctx, record.OutputRef); err != nil || len(result.Output) == 0 {
		return RelayGoalEvidence{}, goalEvidenceFailure()
	}
	for _, name := range []string{"stats.go", "stats_test.go"} {
		if result.AppliedFiles[name], err = s.objects.Read(ctx, record.AppliedFiles[name]); err != nil || len(result.AppliedFiles[name]) == 0 {
			return RelayGoalEvidence{}, goalEvidenceFailure()
		}
	}
	for _, check := range record.Checks {
		if check.Name == "go_test" {
			if _, err := s.objects.Read(ctx, check.ContractRef); err != nil {
				return RelayGoalEvidence{}, goalEvidenceFailure()
			}
		}
		if _, err := s.objects.Read(ctx, check.StdoutRef); err != nil {
			return RelayGoalEvidence{}, goalEvidenceFailure()
		}
		if _, err := s.objects.Read(ctx, check.StderrRef); err != nil {
			return RelayGoalEvidence{}, goalEvidenceFailure()
		}
	}
	return result, nil
}
