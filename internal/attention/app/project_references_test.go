package app

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestScopedAgentReadsFixedOldVersionWithoutHumanHeat(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	a := importFixture(t, s, "reference-a", "A")
	space, err := s.CreateProjectSpace(ctx, human, ProjectSpaceCommand{CommandMeta: meta("space", 1), Title: "Project"})
	if err != nil {
		t.Fatal(err)
	}
	space, err = s.ReferenceMaterial(ctx, human, space.Space.ID, ReferenceMaterialCommand{CommandMeta: meta("reference", space.Space.Version), MaterialID: a.Material.ID, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	b := importFixture(t, s, "reference-b", "B")
	before, _ := s.GetMaterial(ctx, human, a.Material.ID)
	for i := 0; i < 2; i++ {
		snapshot, err := s.ReadProjectReference(ctx, human, space.Space.ID, SourceRef{MaterialID: a.Material.ID, Revision: 1}, false)
		if err != nil || snapshot.Text != "A" {
			t.Fatal("human scoped A context failed", err)
		}
	}
	humanRead, _ := s.GetMaterial(ctx, human, a.Material.ID)
	if humanRead.Material.AgentUsageCount != before.Material.AgentUsageCount || humanRead.Material.HumanUsageCount != before.Material.HumanUsageCount || humanRead.Material.AttentionScore != before.Material.AttentionScore {
		t.Fatal("plain human context query counted heat or impersonated agent")
	}
	agent := Principal{Kind: "agent", ID: "scoped-agent"}
	for i := 0; i < 3; i++ {
		snapshot, err := s.ReadProjectReference(ctx, agent, space.Space.ID, SourceRef{MaterialID: a.Material.ID, Revision: 1}, false)
		if err != nil || snapshot.Text != "A" || snapshot.Ref.Revision != 1 {
			t.Fatal("scoped old fixed version not readable", err)
		}
	}
	after, _ := s.GetMaterial(ctx, human, a.Material.ID)
	if after.Material.AgentUsageCount != before.Material.AgentUsageCount+3 || after.Material.HumanUsageCount != before.Material.HumanUsageCount || after.Material.AttentionScore != before.Material.AttentionScore || after.Material.CurrentRevision != b.Material.CurrentRevision {
		t.Fatal("Agent read changed human heat or current revision")
	}
	_, err = s.ReadProjectReference(ctx, agent, space.Space.ID, SourceRef{MaterialID: b.Material.ID, Revision: 2}, false)
	errorCode(t, err, apierrors.ScopeDenied)
	space, err = s.RemoveMaterialReference(ctx, human, space.Space.ID, RemoveMaterialReferenceCommand{CommandMeta: meta("remove-ref", space.Space.Version), MaterialID: a.Material.ID})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ReadProjectReference(ctx, agent, space.Space.ID, SourceRef{MaterialID: a.Material.ID, Revision: 1}, false)
	errorCode(t, err, apierrors.ScopeDenied)
}

func TestOptionalLinkedScopeIsOneHopAndBoundToActiveFixedRoot(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	root := importFixture(t, s, "C-root", "root source")
	c := importCommand("C-target", "related source")
	c.SourceKey = "related-source"
	c.SourceLocator = "https://example.test/related"
	queued, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	job, _ := s.GetJob(ctx, human, queued.JobID)
	target, err := s.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	space, err := s.CreateProjectSpace(ctx, human, ProjectSpaceCommand{CommandMeta: meta("C-space", 1), Title: "Project"})
	if err != nil {
		t.Fatal(err)
	}
	space, err = s.ReferenceMaterial(ctx, human, space.Space.ID, ReferenceMaterialCommand{CommandMeta: meta("C-reference", space.Space.Version), MaterialID: root.Material.ID, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("C-link", 1), Stage: "topic", InputRefs: sourceRefs(root), RelatedRefs: sourceRefs(target), OutputText: "Explicit relation", ProcessingConfig: "human"})
	if err != nil {
		t.Fatal(err)
	}
	agent := Principal{ID: "scoped-C-agent", Kind: "agent"}
	ref := sourceRefs(target)[0]
	_, err = s.ReadProjectReference(ctx, agent, space.Space.ID, ref, false)
	errorCode(t, err, apierrors.ScopeDenied)
	linked, err := s.LinkedProjectReferences(ctx, agent, space.Space.ID, sourceRefs(root)[0])
	if err != nil || len(linked) != 1 || linked[0].MaterialID != target.Material.ID {
		t.Fatal("explicit fixed link absent", err)
	}
	snapshot, err := s.ReadProjectReference(ctx, agent, space.Space.ID, ref, true)
	if err != nil || snapshot.Text != "related source" {
		t.Fatal("approved expanded fixed link unreadable", err)
	}
	current, err := s.GetMaterial(ctx, human, root.Material.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.UpdateMaterial(ctx, human, root.Material.ID, UpdateMaterialCommand{CommandMeta: meta("C-withdraw-root", current.Material.Version), Lifecycle: "withdrawn"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ReadProjectReference(ctx, agent, space.Space.ID, ref, true)
	errorCode(t, err, apierrors.ScopeDenied)
}
