package events

import (
	"context"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"github.com/olivere/ndjson"
)

// 正文只供本机受信应用和对象保存；不进入事件投影或 HTTP。
type RelayGoalOutput struct {
	Text       []byte                 `json:"-"`
	Origin     string                 `json:"-"`
	Projection RelaySessionProjection `json:"-"`
}

// 核对原档完整快照与准确已接受 turn，不写游标、不推断 started 或任务完成。
func ReadRelayGoalOutput(ctx context.Context, binding RelaySessionBinding, turnID string, acceptedAt time.Time) (RelayGoalOutput, error) {
	var result RelayGoalOutput
	if binding.Format != "codex" || turnID == "" || acceptedAt.IsZero() {
		return result, sessionFailure("goal_output_binding_unverified")
	}
	batch, err := ReadRelaySession(ctx, binding, RelaySessionCursor{})
	if err != nil {
		return result, err
	}
	if !batch.HeaderVerified || !batch.SnapshotComplete || batch.TailBytes != 0 || batch.Next.Counts.Errors != 0 {
		return result, sessionFailure("goal_output_snapshot_unverified")
	}
	file, err := os.Open(binding.Path)
	if err != nil {
		return result, sessionFailure("source_open_failed")
	}
	defer file.Close()
	identity, err := verifyRelaySessionSourceFile(file, binding.Path)
	if err != nil {
		return result, err
	}
	info, err := file.Stat()
	if err != nil || identity != batch.Next.FileIdentity || info.Size() != batch.SnapshotBytes {
		return result, sessionFailure("source_changed_during_read")
	}
	items := make(map[string]RelaySessionItem, len(batch.Items))
	for _, item := range batch.Items {
		items[item.Origin] = item
	}
	reader := ndjson.NewReaderSize(sessionContextReader{ctx, io.NewSectionReader(file, 0, info.Size())}, int(info.Size())+1)
	var offset int64
	for reader.Next() {
		start := offset
		end, complete, err := sessionLineEnd(file, start, int64(len(reader.Bytes())), info.Size())
		if err != nil || !complete {
			return RelayGoalOutput{}, sessionFailure("source_changed_during_read")
		}
		offset = end
		var record struct {
			Type    string `json:"type"`
			Payload struct {
				Type             string  `json:"type"`
				TurnID           string  `json:"turn_id"`
				LastAgentMessage *string `json:"last_agent_message"`
			} `json:"payload"`
		}
		if reader.Decode(&record) != nil {
			return RelayGoalOutput{}, sessionFailure("malformed_json")
		}
		if record.Type != "event_msg" || (record.Payload.Type != "task_complete" && record.Payload.Type != "turn_complete") || record.Payload.TurnID != turnID {
			continue
		}
		if result.Origin != "" {
			return RelayGoalOutput{}, sessionFailure("goal_output_ambiguous")
		}
		origin, _ := sessionOrigin(binding.BindingID(), identity, start)
		item, exists := items[origin]
		when, parseErr := time.Parse(time.RFC3339Nano, item.Projection.TS)
		if !exists || parseErr != nil || when.Before(acceptedAt) || record.Payload.LastAgentMessage == nil || *record.Payload.LastAgentMessage == "" || !utf8.ValidString(*record.Payload.LastAgentMessage) {
			return RelayGoalOutput{}, sessionFailure("goal_output_unverified")
		}
		result = RelayGoalOutput{Text: []byte(*record.Payload.LastAgentMessage), Origin: origin, Projection: item.Projection}
	}
	if reader.Err() != nil || offset != info.Size() {
		return RelayGoalOutput{}, sessionFailure("source_read_failed")
	}
	after, err := file.Stat()
	if err != nil || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return RelayGoalOutput{}, sessionFailure("source_changed_during_read")
	}
	current, err := verifyRelaySessionSourceFile(file, binding.Path)
	if err != nil {
		return RelayGoalOutput{}, err
	}
	if current != identity {
		return RelayGoalOutput{}, sessionFailure("source_replaced")
	}
	if result.Origin == "" {
		return result, sessionFailure("goal_output_not_recorded")
	}
	return result, nil
}
