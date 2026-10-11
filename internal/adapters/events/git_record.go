package events

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// GitLogRecord is populated from actual git log, never an executed commit command.
type GitLogRecord struct {
	Commit  string `json:"commit"`
	TS      int64  `json:"ts"`
	Subject string `json:"subject"`
}

func MapGitLogRecord(project string, raw json.RawMessage, receivedAt int64) (app.StoredEvent, error) {
	var v GitLogRecord
	if json.Unmarshal(raw, &v) != nil || project == "" || v.TS <= 0 || (len(v.Commit) != 40 && len(v.Commit) != 64) {
		return app.StoredEvent{}, fmt.Errorf("invalid git log record")
	}
	if _, err := hex.DecodeString(v.Commit); err != nil {
		return app.StoredEvent{}, fmt.Errorf("invalid git commit ID")
	}
	kind := "unknown"
	for _, rule := range nativeKindRules {
		if rule.selector == "git_record" && rule.value == "commit" {
			kind = rule.kind
		}
	}
	o := nativeObservation{cwd: project, ts: v.TS, kind: kind, outcome: "unknown", depth: "unknown"}
	row := receipt("git:"+strings.ToLower(v.Commit), domain.EventSourceManual, "git_commit", "git:authorized_log", raw, o, receivedAt)
	if row.Event != nil {
		row.Event.Topic = MatchSeedTopic(project, v.Subject).Topic
	}
	return row, nil
}
