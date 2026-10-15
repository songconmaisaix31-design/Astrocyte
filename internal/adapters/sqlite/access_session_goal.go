package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/google/uuid"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

// 只查正选输出位置，返回事件库实际 ID；来源 UUID 不冒充已提交的事件 ID。
func (s *AccessSessionReader) ReadCommittedRelayGoalOutput(ctx context.Context, origin, sessionID, projectID string, expected workbench.Event) (string, error) {
	parts := strings.Split(origin, ":")
	if len(parts) != 3 || parts[0] != "relay-session" || parts[2] != expected.ID || expected.Scope != projectID || sessionID == "" || projectID == "" || (expected.Kind != "task_complete" && expected.Kind != "turn_complete") {
		return "", errors.New("goal output admission scope invalid")
	}
	for _, value := range []string{parts[1], parts[2], sessionID} {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || id.String() != value {
			return "", errors.New("goal output admission identity invalid")
		}
	}
	var id, observation string
	var raw []byte
	if err := s.db.Conn().QueryRowContext(ctx, "SELECT id,observation,raw_bytes FROM attention_engine_events WHERE mapping_status='pending' AND origin_key=?", origin).Scan(&id, &observation, &raw); err != nil {
		return "", errors.New("goal output has no original committed event")
	}
	var metadata attention.Observation
	if json.Unmarshal([]byte(observation), &metadata) != nil || metadata.Session != sessionID || metadata.Scope != projectID || metadata.Source != "codex" || metadata.SourceKind != expected.Kind || metadata.Provenance != "workspace:relay_session" {
		return "", errors.New("goal output original event binding invalid")
	}
	var got workbench.Event
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&got) != nil || decoder.Decode(new(any)) != io.EOF || !reflect.DeepEqual(got, expected) || id == "" {
		return "", errors.New("goal output original projection differs")
	}
	return id, nil
}
