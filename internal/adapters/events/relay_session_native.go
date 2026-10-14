package events

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// 原生观察端口只交付派生字段；本结构不包含协议原文。
type RelaySessionNativeEvent struct {
	Sequence int64
	Kind     string
	Tokens   *int
	TS       time.Time
}

type RelaySessionNativeBatch struct {
	Binding                   RelaySessionBinding
	BindingID                 string
	Events                    []RelaySessionNativeEvent
	LastSequence              int64
	Errors, Unknown           int64
	PrefixComplete, Finalized string
	FirstGap                  string
}

// 游标仅保存已提交序号与覆盖计数，事件投影仍由 EventRepository 持有。
type RelaySessionNativeCursor struct {
	Version             int
	BindingID           string
	Sequence            int64
	Counts              RelaySessionCounts
	FirstError, Blocked string
}

func sessionNativeOrigin(binding string, sequence int64) (string, string) {
	id := uuid.NewSHA1(uuid.MustParse(binding), []byte(fmt.Sprintf("native-notification:%d", sequence))).String()
	return "relay-native-session:" + binding + ":" + id, id
}

func (s *RelaySessionCapture) CaptureNative(ctx context.Context, source RelaySessionNativeBatch, prior RelaySessionNativeCursor) (RelaySessionNativeCursor, error) {
	next := prior
	fail := func(code string) (RelaySessionNativeCursor, error) {
		failed := prior
		if id, err := uuid.Parse(source.BindingID); err == nil && id != uuid.Nil && id.String() == source.BindingID && prior.Version == 0 {
			failed.Version, failed.BindingID = 1, source.BindingID
		}
		if failed.FirstError == "" {
			failed.FirstError = code
		}
		failed.Blocked = code
		failed.Counts.Errors++
		return failed, sessionFailure(code)
	}
	if err := ctx.Err(); err != nil {
		return fail("capture_cancelled")
	}
	if prior.Blocked != "" {
		return prior, sessionFailure("capture_blocked_requires_reconciliation")
	}
	b := source.Binding
	if !b.Owned || !b.Authorized || b.NativeID == "" || b.ProjectID == "" {
		return fail("selection_denied")
	}
	for _, value := range []string{source.BindingID, b.TaskID, b.TerminalID, b.SessionID, b.WorktreeID} {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || id.String() != value {
			return fail("binding_invalid")
		}
	}
	if prior.Version != 0 && (prior.Version != 1 || prior.BindingID != source.BindingID) {
		return fail("checkpoint_binding_mismatch")
	}
	if prior.Version == 0 && (prior.Sequence != 0 || prior.BindingID != "" || prior.Counts != (RelaySessionCounts{})) {
		return fail("checkpoint_invalid")
	}
	if source.PrefixComplete != "confirmed" {
		return fail("native_prefix_unverified")
	}
	if source.LastSequence < prior.Sequence || source.Errors < 0 || source.Unknown < 0 {
		return fail("native_sequence_invalid")
	}
	if source.Errors != 0 {
		return fail("native_source_error")
	}
	next.Version, next.BindingID = 1, source.BindingID
	items := []RelaySessionItem{}
	seen := int64(0)
	for _, event := range source.Events {
		if err := ctx.Err(); err != nil {
			return fail("capture_cancelled")
		}
		if event.Sequence <= seen || event.Sequence <= 0 || event.Sequence > source.LastSequence {
			return fail("native_sequence_invalid")
		}
		seen = event.Sequence
		if event.Sequence <= prior.Sequence {
			continue
		}
		if event.Sequence != next.Sequence+1 {
			return fail("native_sequence_gap")
		}
		if event.TS.IsZero() || !sessionNativeKind(event.Kind) || (event.Tokens != nil && *event.Tokens < 0) {
			return fail("invalid_native_record")
		}
		origin, id := sessionNativeOrigin(source.BindingID, event.Sequence)
		items = append(items, RelaySessionItem{Origin: origin, Projection: RelaySessionProjection{ID: id, Kind: event.Kind, Topic: "relay:" + b.TaskID, Tokens: event.Tokens, Scope: b.ProjectID, TS: event.TS.UTC().Format(time.RFC3339Nano)}})
		next.Sequence = event.Sequence
		next.Counts.Projected++
		if event.Tokens == nil {
			next.Counts.UnknownTokens++
		}
	}
	if next.Sequence != source.LastSequence {
		return fail("native_sequence_gap")
	}
	next.Counts.Lines, next.Counts.Unknown = source.LastSequence, source.Unknown
	if len(items) != 0 {
		if err := s.admit(ctx, b, items); err != nil {
			return fail("admission_unknown")
		}
	}
	return next, nil
}

func sessionNativeKind(kind string) bool {
	switch kind {
	case "session_started", "session_closed", "turn_started", "turn_completed", "reasoning", "text", "token_usage", "model_response", "native_error", "native_warning", "native_activity":
		return true
	}
	return false
}
