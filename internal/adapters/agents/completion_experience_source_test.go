package agents

import (
	"context"
	"errors"
	"testing"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestCompletionExperienceSourceRejectsUnconfiguredAuthority(t *testing.T) {
	if source, err := NewRelayCompletionExperienceSource(nil, nil); source != nil || err == nil {
		t.Fatal("missing actual completion readers produced a configured source")
	}
	source := &RelayCompletionExperienceSource{}
	for _, principal := range []attention.Principal{{}, {ID: "unconfigured-actor", Kind: "human"}, {ID: "unconfigured-actor", Kind: "agent"}} {
		value, err := source.ReadExperienceCompletion(context.Background(), principal, "unconfigured-project", "unrecorded-completion")
		if err == nil || value.EventID != "" || value.CompletionChannel != "" || value.CompletionReceiptID != "" || len(value.OriginEventIDs) != 0 {
			t.Fatal("unconfigured authority produced completion source evidence")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.ReadExperienceCompletion(ctx, attention.Principal{ID: "unconfigured-actor", Kind: "human"}, "unconfigured-project", "unrecorded-completion"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled completion source read continued")
	}
	if err := source.AuthorizeExperienceCompletion(context.Background(), attention.Principal{ID: "unconfigured-actor", Kind: "human"}, attention.ExperienceCompletionEvidence{}); err == nil {
		t.Fatal("missing actual fixed completion acquired current authorization")
	}
}
