package app

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestExplicitMentionAndProjectReuseReceiptAndHumanOnly(t *testing.T) {
	ctx := context.Background()
	s, r, _, m, _ := automaticFixture(t)
	mention := RecordUseCommand{CommandMeta: meta("explicit-mention", m.Material.Version), Action: "mention"}
	first, err := s.RecordUse(ctx, human, m.Material.ID, mention)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := s.RecordUse(ctx, human, m.Material.ID, mention)
	if err != nil || duplicate.Version != first.Version {
		t.Fatalf("mention replay: %+v %v", duplicate, err)
	}
	reuse := RecordUseCommand{CommandMeta: meta("explicit-reuse", first.Version), Action: "project_reuse"}
	second, err := s.RecordUse(ctx, human, m.Material.ID, reuse)
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := s.RecordUse(ctx, human, m.Material.ID, reuse)
	if err != nil || repeat.Version != second.Version {
		t.Fatal("project reuse not idempotent")
	}
	got, _ := s.GetMaterial(ctx, human, m.Material.ID)
	if got.Material.HumanUsageCount != 2 || got.Material.AgentUsageCount != 0 || got.Material.AttentionScore != 4 || got.Material.LongTermValue != 1 || len(got.Uses) != 2 {
		t.Fatalf("human signals: %+v", got)
	}
	if got.Uses[0].Action != "mention" || got.Uses[1].Action != "project_reuse" {
		t.Fatal("signal reason lost")
	}
	stale := RecordUseCommand{CommandMeta: meta("stale-project-reuse", first.Version), Action: "project_reuse"}
	_, err = s.RecordUse(ctx, human, m.Material.ID, stale)
	errorCode(t, err, apierrors.VersionConflict)
	reuse.CommandMeta = meta("agent-project-reuse", second.Version)
	_, err = s.RecordUse(ctx, agent, m.Material.ID, reuse)
	errorCode(t, err, apierrors.ScopeDenied)
	r.failEvent = true
	reuse.CommandMeta = meta("failed-project-reuse", second.Version)
	_, err = s.RecordUse(ctx, human, m.Material.ID, reuse)
	if err == nil {
		t.Fatal("failed outbox accepted")
	}
	r.failEvent = false
	after, _ := s.GetMaterial(ctx, human, m.Material.ID)
	if after.Material.Version != second.Version || after.Material.HumanUsageCount != 2 || after.Material.LongTermValue != 1 {
		t.Fatal("failed/rejected action changed attention")
	}
}
