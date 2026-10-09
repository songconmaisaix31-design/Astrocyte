package app

import (
	"context"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestClassificationAndSpaceReferenceKeepOriginalAndFixedVersion(t *testing.T) {
	s, r, _ := fixture(t)
	material := importFixture(t, s, "first", "original")
	d, err := s.CreateMaterialDomain(context.Background(), human, MaterialDomainCommand{CommandMeta: meta("domain", 1), Title: "日常学习"})
	if err != nil {
		t.Fatal(err)
	}
	c := SetMaterialDomainsCommand{CommandMeta: meta("classify", material.Material.Version), DomainIDs: []string{d.Domain.ID, d.Domain.ID}}
	classified, err := s.SetMaterialDomains(context.Background(), human, material.Material.ID, c)
	if err != nil || len(classified.Material.DomainIDs) != 1 || classified.Material.HumanUsageCount != 1 || classified.Material.AttentionScore != 1 {
		t.Fatal("classification failed or duplicated signal")
	}
	c.RequestID = "retry-with-new-trace"
	if _, err = s.SetMaterialDomains(context.Background(), human, material.Material.ID, c); err != nil {
		t.Fatal(err)
	}
	same, err := s.SetMaterialDomains(context.Background(), human, material.Material.ID, SetMaterialDomainsCommand{CommandMeta: meta("same-membership", classified.Material.Version), DomainIDs: []string{d.Domain.ID}})
	if err != nil || same.Material.Version != classified.Material.Version || same.Material.HumanUsageCount != 1 {
		t.Fatal("same classification did work")
	}
	space, err := s.CreateProjectSpace(context.Background(), human, ProjectSpaceCommand{CommandMeta: meta("space", 1), Title: "当前研究"})
	if err != nil {
		t.Fatal(err)
	}
	reference := ReferenceMaterialCommand{CommandMeta: meta("mention", space.Space.Version), MaterialID: material.Material.ID, Revision: 1}
	space, err = s.ReferenceMaterial(context.Background(), human, space.Space.ID, reference)
	if err != nil || len(space.Space.MaterialRefs) != 1 || space.Space.MaterialRefs[0].Locator != material.Revisions[0].SourceLocator {
		t.Fatal("reference was not server-derived fixed source")
	}
	if _, err = s.ReferenceMaterial(context.Background(), human, space.Space.ID, reference); err != nil {
		t.Fatal(err)
	}
	still, err := s.GetMaterial(context.Background(), human, material.Material.ID)
	if err != nil || !slices.Equal(still.Material.DomainIDs, []string{d.Domain.ID}) || len(r.state.Materials) != 1 || len(still.Revisions) != 1 || still.Material.HumanUsageCount != 2 {
		t.Fatal("@ copied/moved original or duplicated signal")
	}
	updated := importFixture(t, s, "new", "changed bytes")
	pinned, _ := s.GetProjectSpace(context.Background(), human, space.Space.ID)
	if updated.Material.CurrentRevision != 2 || pinned.Space.MaterialRefs[0].Revision != 1 {
		t.Fatal("reference silently followed new source")
	}
	_, err = s.GetProjectSpace(context.Background(), agent, space.Space.ID)
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = s.GetContent(context.Background(), agent, material.Material.ID, 1)
	errorCode(t, err, apierrors.ScopeDenied)
	removed, err := s.RemoveMaterialReference(context.Background(), human, space.Space.ID, RemoveMaterialReferenceCommand{CommandMeta: meta("remove", pinned.Space.Version), MaterialID: material.Material.ID})
	if err != nil || len(removed.Space.MaterialRefs) != 0 || len(r.state.Materials) != 1 || len(r.state.Domains) != 1 || len(r.state.Materials[material.Material.ID].Revisions) != 2 {
		t.Fatal("reference removal discarded original history or domain")
	}
}

func TestSpaceAndClassificationFailureRollbackSignals(t *testing.T) {
	s, r, _ := fixture(t)
	material := importFixture(t, s, "import", "original")
	space, err := s.CreateProjectSpace(context.Background(), human, ProjectSpaceCommand{CommandMeta: meta("space", 1), Title: "研究"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SetMaterialDomains(context.Background(), human, material.Material.ID, SetMaterialDomainsCommand{CommandMeta: meta("missing-domain", 1), DomainIDs: []string{"absent"}})
	errorCode(t, err, apierrors.NotFound)
	r.failEvent = true
	_, err = s.ReferenceMaterial(context.Background(), human, space.Space.ID, ReferenceMaterialCommand{CommandMeta: meta("rollback", 1), MaterialID: material.Material.ID, Revision: 1})
	errorCode(t, err, apierrors.InternalError)
	if len(r.state.Spaces[space.Space.ID].MaterialRefs) != 0 || r.state.Spaces[space.Space.ID].Version != 1 || r.state.Materials[material.Material.ID].Material.HumanUsageCount != 0 {
		t.Fatal("failed reference committed selection or human heat")
	}
	r.failEvent = false
	space, err = s.ReferenceMaterial(context.Background(), human, space.Space.ID, ReferenceMaterialCommand{CommandMeta: meta("valid", 1), MaterialID: material.Material.ID, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.RemoveMaterialReference(context.Background(), human, space.Space.ID, RemoveMaterialReferenceCommand{CommandMeta: meta("stale", 1), MaterialID: material.Material.ID})
	errorCode(t, err, apierrors.VersionConflict)
	_, err = s.CreateMaterialDomain(context.Background(), agent, MaterialDomainCommand{CommandMeta: meta("agent-domain", 1), Title: "Forbidden"})
	errorCode(t, err, apierrors.ScopeDenied)
}

func TestMaterialListUsesConfiguredActivityRankingAndExplainsDefaults(t *testing.T) {
	s, r, o := fixture(t)
	old := importFixture(t, s, "old", "old original")
	c := importCommand("new", "new original")
	c.SourceKey = "second-source"
	c.SourceLocator = "https://example.test/second"
	now := s.options.Clock()
	s.options.Clock = func() time.Time { return now.Add(time.Hour) }
	response, err := s.ImportMaterial(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	newJob, _ := s.GetJob(context.Background(), human, response.JobID)
	configured := NewAttentionService(r, s.sources, o, ServiceOptions{Clock: s.options.Clock, AttentionHalfLife: time.Hour, AttentionWeights: map[string]float64{"reread": 4}})
	_, err = configured.RecordUse(context.Background(), human, old.Material.ID, RecordUseCommand{CommandMeta: meta("active", 1), Action: "reread"})
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := configured.ListMaterials(context.Background())
	if err != nil || ordered.Items[0].(Material).ID != old.Material.ID {
		t.Fatal("computed score ignored by list order")
	}
	first := ordered.Items[0].(Material)
	if first.AttentionScore != 4 || first.RankingStrategy != "attention-v1" || first.RankingReason == "" || first.AttentionHalfLifeSeconds != 3600 || first.AttentionWeights["reread"] != 4 {
		t.Fatal("ranking algorithm or actual configured policy missing")
	}
	first.AttentionWeights["reread"] = 999
	if configured.options.AttentionWeights["reread"] != 4 {
		t.Fatal("response mutates service policy")
	}
	_, err = configured.UpdateMaterial(context.Background(), human, *newJob.MaterialID, UpdateMaterialCommand{CommandMeta: meta("pin-new", 1), Pinned: true, Lifecycle: "active"})
	if err != nil {
		t.Fatal(err)
	}
	ordered, err = configured.ListMaterials(context.Background())
	if err != nil || ordered.Items[0].(Material).ID != *newJob.MaterialID || ordered.Items[0].(Material).AttentionScore != 0 {
		t.Fatal("cold fixed item fell out or gained implicit heat")
	}
	if err = ValidateAttentionWeights(map[string]float64{"reread": math.NaN()}); err == nil {
		t.Fatal("invalid configuration accepted")
	}
	invalid := NewAttentionService(r, s.sources, o, ServiceOptions{AttentionWeights: map[string]float64{"reread": -1}})
	_, err = invalid.ListMaterials(context.Background())
	errorCode(t, err, apierrors.ValidationFailed)
}

func TestImportReturningKnownContentRestoresHeadWithoutExtraWork(t *testing.T) {
	s, r, o := fixture(t)
	a := importFixture(t, s, "A", "A")
	b := importFixture(t, s, "B", "B")
	count := o.publishes.Load()
	returned := importFixture(t, s, "A-new-command", "A")
	if returned.Material.CurrentRevision != 1 || returned.Material.Version != 3 || len(returned.Revisions) != 2 || returned.Revisions[1].ObjectRef != b.Revisions[1].ObjectRef || o.publishes.Load() != count || returned.Material.ID != a.Material.ID {
		t.Fatal("A B A lost history, failed to restore head or repeated source work")
	}
	_, err := s.ImportMaterial(context.Background(), human, importCommand("A-repeat", "A"))
	if err != nil {
		t.Fatal(err)
	}
	if r.state.Materials[a.Material.ID].Material.Version != 3 || len(r.state.Jobs) != 2 {
		t.Fatal("unchanged current input created duplicate work")
	}
}
