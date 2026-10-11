package attention_engine_test

import (
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM1Section52WeightSamples(t *testing.T) {
	for _, tt := range []struct {
		kind, outcome, depth string
		class                domain.SignalClass
		weight               float64
	}{
		{domain.EventKindReadPaper, "unknown", "deep", domain.SignalImplicitDepth, 1},
		{domain.EventKindWatchVideo, "unknown", "deep", domain.SignalImplicitDepth, 1},
		{domain.EventKindEditFile, "unknown", "unknown", domain.SignalImplicitRegular, .6},
		{domain.EventKindCommit, "unknown", "unknown", domain.SignalImplicitRegular, .6},
		{domain.EventKindSearch, "unknown", "unknown", domain.SignalImplicitRegular, .6},
		{domain.EventKindBookmark, "unknown", "unknown", domain.SignalImplicitLight, .3},
		{domain.EventKindCollect, "unknown", "unknown", domain.SignalImplicitLight, .3},
		{domain.EventKindUserApprove, "unknown", "unknown", domain.SignalExplicitPositive, 1.5},
		{domain.EventKindUserVeto, "unknown", "unknown", domain.SignalExplicitNegative, -.8},
		{domain.EventKindAgentTask, "success", "unknown", domain.SignalFeedbackSuccess, .8},
		{domain.EventKindAgentTask, "failure", "unknown", domain.SignalFeedbackFailure, -.1},
	} {
		s, err := domain.NormalizeSignal("source-receipt", domain.Event{Kind: tt.kind, Topic: "selected-topic", TS: 1791590400123}, tt.outcome, tt.depth, 1)
		if err != nil || s == nil || s.Class != tt.class || s.Weight != tt.weight || s.Strength != 1 || s.EventID != "source-receipt" || s.TS != 1791590400123 || s.ConfidenceOnly() != (tt.class == domain.SignalFeedbackFailure) {
			t.Fatalf("§5.2 sample %s/%s: %+v, %v", tt.kind, tt.outcome, s, err)
		}
	}
}

func TestM1UnknownDepthAndOutcomeDoNotInventSignals(t *testing.T) {
	for _, kind := range []string{domain.EventKindReadPaper, domain.EventKindWatchVideo, domain.EventKindAgentTask} {
		s, err := domain.NormalizeSignal("observed", domain.Event{Kind: kind}, "unknown", "unknown", 1)
		if err != nil || s != nil {
			t.Fatalf("unknown observation produced signal: %+v, %v", s, err)
		}
	}
}
