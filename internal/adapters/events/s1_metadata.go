package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
)

// ReadCachedProjectMetadata consumes S1's existing metadata-only cache for one
// authorized project root. It does not rediscover histories or read messages.
func ReadCachedProjectMetadata(ctx context.Context, db *sql.DB, projectRoot string, limit int) ([]json.RawMessage, bool, error) {
	if !filepath.IsAbs(projectRoot) || limit < 1 || limit > maxReplayRecords {
		return nil, false, fmt.Errorf("absolute approved project root and limit 1..128 required")
	}
	rows, err := db.QueryContext(ctx, `SELECT c.value
		FROM local_project_discovery d, json_each(d.data, '$.projects') p,
		json_each(p.value, '$.contributors') c
		WHERE d.id='registered' AND lower(json_extract(p.value, '$.root'))=lower(?)
		AND json_extract(c.value, '$.source')='native_session_header'
		AND json_extract(c.value, '$.cli') IN ('claude','codex')
		ORDER BY json_extract(c.value, '$.cli'), json_extract(c.value, '$.session_id') LIMIT ?`, filepath.Clean(projectRoot), limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	result := []json.RawMessage{}
	for rows.Next() {
		if len(result) == limit {
			return result, true, nil
		}
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, false, err
		}
		result = append(result, json.RawMessage(raw))
	}
	return result, false, rows.Err()
}

func cachedMetadata(raw json.RawMessage) (string, nativeObservation, error) {
	v := object(raw)
	source := str(v, "cli")
	if source == "claude" {
		source = "claude_code"
	}
	if source != "claude_code" && source != "codex" || str(v, "source") != "native_session_header" || str(v, "session_id") == "" {
		return "", nativeObservation{}, fmt.Errorf("unverified session metadata")
	}
	return source, nativeObservation{session: str(v, "session_id"), cwd: str(v, "root"), ts: millis(v["created_at"]), kind: "unknown", depth: "unknown", outcome: "unknown"}, nil
}
