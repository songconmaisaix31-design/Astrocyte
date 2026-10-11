package domain

import (
	"fmt"
	"math"
)

// NormalizeSignal uses only observed semantics and explicitly supplied s_i.
// A nil result retains the Event but does not invent an attention signal.
func NormalizeSignal(eventID string, event Event, outcome, depth string, strength float64) (*Signal, error) {
	if eventID == "" || math.IsNaN(strength) || math.IsInf(strength, 0) || strength < 0 {
		return nil, fmt.Errorf("event ID and finite nonnegative strength required")
	}
	var class SignalClass
	switch event.Kind {
	case EventKindEditFile, EventKindCommit, EventKindSearch:
		class = SignalImplicitRegular
	case EventKindBookmark, EventKindCollect:
		class = SignalImplicitLight
	case EventKindUserApprove:
		class = SignalExplicitPositive
	case EventKindUserVeto:
		class = SignalExplicitNegative
	case EventKindReadPaper, EventKindWatchVideo:
		if depth != "deep" {
			return nil, nil
		}
		class = SignalImplicitDepth
	case EventKindAgentTask:
		switch outcome {
		case "success", "succeeded":
			class = SignalFeedbackSuccess
		case "failure", "failed":
			class = SignalFeedbackFailure
		default:
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("unsupported core event kind %q", event.Kind)
	}
	return &Signal{EventID: eventID, Topic: event.Topic, Class: class, Strength: strength, Weight: SignalWeight(class), TS: event.TS}, nil
}

// ConfidenceOnly is the §5.2 failure routing rule, not an attention penalty.
func (s Signal) ConfidenceOnly() bool { return s.Class == SignalFeedbackFailure }
