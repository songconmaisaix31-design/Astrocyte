package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"io"
	"sort"
	"strings"
	"time"
)

// AccessSessionReader selects only positive owned origins in the isolated app
// DB before decoding JSON. Unrelated plaintext never enters the projection.
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
	exact := map[string]workbench.SessionAdmissionScope{}
	prefixes := map[string]workbench.SessionAdmissionScope{}
	exactArgs := []any{}
	validID := func(value string) bool {
		id, e := uuid.Parse(value)
		return e == nil && id != uuid.Nil && id.String() == value
	}
	for _, scope := range scopes {
		if !validID(scope.TaskID) || !validID(scope.TerminalID) || scope.ProjectID == "" {
			return nil, 0, errors.New("session admission scope invalid")
		}
		if scope.NativeBindingID != "" {
			if !validID(scope.NativeBindingID) || !validID(scope.NativeSessionID) {
				return nil, 0, errors.New("session admission scope invalid")
			}
			prefix := "relay-session:" + scope.NativeBindingID + ":"
			if scope.NativeNotifications {
				prefix = "relay-native-session:" + scope.NativeBindingID + ":"
			}
			prefixes[prefix] = scope
			clauses = append(clauses, "origin_key LIKE ?")
			args = append(args, prefix+"%")
		}
		for _, id := range scope.RelayEventIDs {
			if !validID(id) {
				return nil, 0, errors.New("session admission scope invalid")
			}
			origin := "relay:" + id
			if _, exists := exact[origin]; !exists {
				exactArgs = append(exactArgs, origin)
			}
			exact[origin] = scope
		}
	}
	if len(exactArgs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(exactArgs)), ",")
		clauses = append(clauses, "origin_key IN ("+placeholders+")")
		args = append(args, exactArgs...)
	}
	if len(clauses) == 0 {
		return result, 0, nil
	}
	rows, e := s.db.Conn().QueryContext(ctx, "SELECT id,origin_key,ts,raw_bytes,observation FROM attention_engine_events WHERE mapping_status='pending' AND ("+strings.Join(clauses, " OR ")+") ORDER BY ts DESC,id DESC", args...)
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
		var id, origin, observation string
		var ts int64
		var raw []byte
		if rows.Scan(&id, &origin, &ts, &raw, &observation) != nil {
			return nil, 0, errors.New("session admission unavailable")
		}
		scope, matched := exact[origin]
		expectedSession := scope.TerminalID
		if !matched {
			for prefix, selection := range prefixes {
				if strings.HasPrefix(origin, prefix) {
					scope, matched, expectedSession = selection, true, selection.NativeSessionID
					break
				}
			}
		}
		if !matched {
			return nil, 0, errors.New("session admission binding invalid")
		}
		var metadata struct {
			Session string `json:"session"`
			Scope   string `json:"scope"`
		}
		if json.Unmarshal([]byte(observation), &metadata) != nil || metadata.Session != expectedSession || metadata.Scope != scope.ProjectID {
			return nil, 0, errors.New("session admission metadata invalid")
		}
		var p workbench.Event
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&p) != nil || decoder.Decode(new(any)) != io.EOF || p.Scope != scope.ProjectID || p.Topic != "relay:"+scope.TaskID || id == "" || ts <= 0 || (p.Tokens != nil && *p.Tokens < 0) {
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
		p.ID = id
		p.TS = parsed.UTC().Format(time.RFC3339Nano)
		result = append(result, p)
	}
	if rows.Err() != nil {
		return nil, 0, errors.New("session admissions unavailable")
	}
	// SQLite stores milliseconds; the public source timestamp can retain finer
	// precision. Resolve that order before applying a page boundary.
	sort.Slice(result, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, result[i].TS)
		b, _ := time.Parse(time.RFC3339Nano, result[j].TS)
		if a.Equal(b) {
			return result[i].ID > result[j].ID
		}
		return a.After(b)
	})
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
