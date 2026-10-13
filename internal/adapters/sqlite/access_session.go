package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"io"
	"strings"
	"time"
)

// AccessSessionReader reads committed pending observations in the existing
// isolated app DB only. It never opens native archives or cognitive snapshots.
type AccessSessionReader struct{ db *DB }

func NewAccessSessionReader(db *DB) *AccessSessionReader { return &AccessSessionReader{db: db} }
func (s *AccessSessionReader) ReadCommittedSessionEvents(ctx context.Context, scopes []workbench.SessionAdmissionScope, q workbench.SessionEventsQuery) ([]workbench.Event, int, error) {
	result := []workbench.Event{}
	if len(scopes) == 0 {
		return result, 0, nil
	}
	if q.Page < 1 || q.Page > 1000000 || q.PageSize < 1 || q.PageSize > 100 {
		return nil, 0, errors.New("session admission query invalid")
	}
	clauses := []string{}
	args := []any{}
	selections := map[string]workbench.SessionAdmissionScope{}
	for _, scope := range scopes {
		task, e := uuid.Parse(scope.TaskID)
		terminal, terminalErr := uuid.Parse(scope.TerminalID)
		if e != nil || terminalErr != nil || task == uuid.Nil || terminal == uuid.Nil || task.String() != scope.TaskID || terminal.String() != scope.TerminalID || scope.ProjectID == "" {
			return nil, 0, errors.New("session admission scope invalid")
		}
		if scope.NativeSessionID != "" {
			id, e := uuid.Parse(scope.NativeSessionID)
			if e != nil || id == uuid.Nil || id.String() != scope.NativeSessionID {
				return nil, 0, errors.New("session admission scope invalid")
			}
		}
		topic := "relay:" + scope.TaskID
		selections[topic] = scope
		clauses = append(clauses, "(json_extract(raw_bytes,'$.topic')=? AND json_extract(observation,'$.scope')=? AND ((origin_key LIKE 'relay:%' AND json_extract(observation,'$.session')=?) OR (origin_key LIKE 'relay-session:%' AND json_extract(observation,'$.session')=?)))")
		args = append(args, topic, scope.ProjectID, scope.TerminalID, scope.NativeSessionID)
	}
	rows, e := s.db.Conn().QueryContext(ctx, "SELECT id,origin_key,ts,raw_bytes FROM attention_engine_events WHERE mapping_status='pending' AND ("+strings.Join(clauses, " OR ")+") ORDER BY ts DESC,id DESC", args...)
	if e != nil {
		return nil, 0, errors.New("session admissions unavailable")
	}
	defer rows.Close()
	start, end := time.Time{}, time.Time{}
	if q.Start != "" {
		start, e = time.Parse(time.RFC3339Nano, q.Start)
		if e != nil {
			return nil, 0, errors.New("session admission query invalid")
		}
	}
	if q.End != "" {
		end, e = time.Parse(time.RFC3339Nano, q.End)
		if e != nil {
			return nil, 0, errors.New("session admission query invalid")
		}
	}
	for rows.Next() {
		var id, origin string
		var ts int64
		var raw []byte
		if rows.Scan(&id, &origin, &ts, &raw) != nil {
			return nil, 0, errors.New("session admission unavailable")
		}
		var p workbench.Event
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&p) != nil || decoder.Decode(new(any)) != io.EOF {
			return nil, 0, errors.New("session admission projection invalid")
		}
		scope, matched := selections[p.Topic]
		if !matched || p.Scope != scope.ProjectID || id == "" || ts <= 0 || (p.Tokens != nil && *p.Tokens < 0) {
			return nil, 0, errors.New("session admission projection invalid")
		}
		parsed, parseErr := time.Parse(time.RFC3339Nano, p.TS)
		if parseErr != nil || parsed.UnixMilli() != ts {
			return nil, 0, errors.New("session admission timestamp invalid")
		}
		if (!start.IsZero() && parsed.Before(start)) || (!end.IsZero() && parsed.After(end)) {
			continue
		}
		switch p.Kind {
		case "native_session_header", "native_activity", "edit_file", "search", "read_paper", "commit", "git_activity", "turn_started", "turn_completed", "message", "agent_message", "user_message", "function_call_output", "custom_tool_call", "custom_tool_call_output", "reasoning", "agent_reasoning", "agent_reasoning_raw_content", "turn_context", "task_started", "task_complete", "turn_complete", "context_compacted", "error", "warning":
		default:
			p.Kind = "native_activity"
		}
		p.ID = id // EventRepository's committed receipt, never a cursor-generated ID.
		p.TS = parsed.UTC().Format(time.RFC3339Nano)
		result = append(result, p)
	}
	if rows.Err() != nil {
		return nil, 0, errors.New("session admissions unavailable")
	}
	total := len(result)
	offset := (q.Page - 1) * q.PageSize
	if offset >= total {
		return []workbench.Event{}, total, nil
	}
	last := offset + q.PageSize
	if last > total {
		last = total
	}
	return result[offset:last], total, nil
}
