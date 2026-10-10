package app

import (
	"context"
	"testing"
)

func TestCandidateReadyRequiresRelevantAssociationOrOpenQuestion(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	m := importFixture(t, s, "gate-material", "source bytes")
	refs := sourceRefs(m)
	c := OpportunityCommand{CommandMeta: meta("gate-empty", 1), Title: "Candidate", EvidenceRefs: refs, Purpose: "Use evidence", NextStep: "Compare baseline"}
	o, err := s.CreateOpportunity(ctx, human, c)
	if err != nil || o.Opportunity.State != "incubating" {
		t.Fatalf("bare source/use/next became ready: %+v %v", o.Opportunity, err)
	}
	content, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("gate-content", 1), InputRefs: refs, Stage: "content", OutputText: "Summary", ProcessingConfig: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	c.CommandMeta = meta("gate-content-candidate", 1)
	c.DistillationIDs = []string{content.Distillation.ID}
	o, err = s.CreateOpportunity(ctx, human, c)
	if err != nil || o.Opportunity.State != "incubating" {
		t.Fatal("content summary alone became ready", err)
	}
	topic, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("gate-topic", 1), InputRefs: refs, Stage: "topic", OutputText: "Unresolved relation", PendingQuestions: []string{"Which baseline supports this?"}, ProcessingConfig: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	c.CommandMeta = meta("gate-ready", 1)
	c.DistillationIDs = []string{topic.Distillation.ID}
	o, err = s.CreateOpportunity(ctx, human, c)
	if err != nil || o.Opportunity.State != "ready_for_review" {
		t.Fatal("explicit relevant question did not become ready", err)
	}
	// A record from an older revision cannot satisfy this revision's gate.
	updated := importFixture(t, s, "gate-updated", "changed source")
	c.CommandMeta = meta("gate-stale-link", 1)
	c.EvidenceRefs = sourceRefs(updated)
	o, err = s.CreateOpportunity(ctx, human, c)
	if err != nil || o.Opportunity.State != "incubating" {
		t.Fatal("older evidence question satisfied new revision", err)
	}
}
