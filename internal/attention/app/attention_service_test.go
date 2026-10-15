package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestImportReceiptsContentVersionsAndTransactionIO(t *testing.T) {
	ctx := context.Background()
	_, db, objects := fixture(t)
	repo := &observedRepository{Repository: sqlite.NewAttentionRepository(db)}
	assertUnlocked := func() {
		if repo.active.Load() != 0 {
			t.Error("external I/O held production transaction")
		}
	}
	objects.beforePublish = assertUnlocked
	reader := &observedSource{Reader: importers.NewReader(nil), beforeRead: assertUnlocked}
	s := NewAttentionService(repo, reader, objects, ServiceOptions{})
	row := importFixture(t, s, "first", "original content")
	c := importCommand("first", "original content")
	c.RequestID = "different-trace"
	replayed, err := s.ImportMaterial(ctx, human, c)
	if err != nil || replayed.Status != "queued" {
		t.Fatal("receipt did not retain original complete queued result")
	}
	c.ExportText = "changed input with same idempotency key"
	_, err = s.ImportMaterial(ctx, human, c)
	errorCode(t, err, apierrors.VersionConflict)
	duplicate, err := s.ImportMaterial(ctx, human, importCommand("second-key", "original content"))
	if err != nil || duplicate.JobID != replayed.JobID || duplicate.Status != "succeeded" {
		t.Fatal("same input created duplicate job")
	}
	updated := importFixture(t, s, "changed", "updated content")
	if updated.Material.ID != row.Material.ID || len(updated.Revisions) != 2 || updated.Material.CurrentRevision != 2 {
		t.Fatal("source version history changed")
	}
	old, err := s.GetContent(ctx, human, row.Material.ID, 1)
	if err != nil || old.Text != "original content" {
		t.Fatal("old source overwritten")
	}
	var executions int
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_outbox WHERE type IN ('attention.opportunity_admitted','mission_created')").Scan(&executions); err != nil || executions != 0 {
		t.Fatal("import granted execution")
	}
}

func TestHumanQueriesMachineReadsPinAndValueAreIndependent(t *testing.T) {
	ctx := context.Background()
	s, db, objects := fixture(t)
	material := importFixture(t, s, "import", "manual source")
	id := material.Material.ID
	for i := 0; i < 3; i++ {
		if _, err := s.GetMaterial(ctx, human, id); err != nil {
			t.Fatal("human query failed")
		}
		if _, err := s.GetContent(ctx, human, id, 1); err != nil {
			t.Fatal("human content query failed")
		}
	}
	for i := 0; i < 3; i++ {
		_, err := s.GetMaterial(ctx, agent, id)
		errorCode(t, err, apierrors.ScopeDenied)
	}
	got, err := s.GetMaterial(ctx, human, id)
	if err != nil || got.Material.HumanUsageCount != 0 || got.Material.AttentionScore != 0 || got.Material.AgentUsageCount != 0 {
		t.Fatal("plain queries changed activity")
	}
	_, err = s.RecordUse(ctx, agent, id, RecordUseCommand{CommandMeta: meta("agent-use", got.Material.Version), Action: "adopt"})
	errorCode(t, err, apierrors.ScopeDenied)
	pinned, err := s.UpdateMaterial(ctx, human, id, UpdateMaterialCommand{CommandMeta: meta("pin", got.Material.Version), Pinned: true, Lifecycle: "active"})
	if err != nil || !pinned.Material.Pinned || pinned.Material.AttentionScore != 0 || len(pinned.Revisions) != 1 {
		t.Fatal("pin changed heat or content")
	}
	if _, err = s.RecordUse(ctx, human, id, RecordUseCommand{CommandMeta: meta("adopt", pinned.Material.Version), Action: "adopt"}); err != nil {
		t.Fatal("record human adoption")
	}
	before, err := s.GetMaterial(ctx, human, id)
	if err != nil {
		t.Fatal("load human activity before decay")
	}
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC).Add(7 * 24 * time.Hour)
	s = NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), objects, ServiceOptions{Clock: func() time.Time { return now }})
	after, err := s.GetMaterial(ctx, human, id)
	if err != nil || before.Material.AttentionScore != 3 || after.Material.AttentionScore != 1.5 || after.Material.LongTermValue != 1 || !after.Material.Pinned || after.Material.CurrentRevision != 1 {
		t.Fatal("decay affected pin, value or source version")
	}
}

func TestDistillationLayersReuseFixedInputsAndKeepPendingQuestion(t *testing.T) {
	ctx := context.Background()
	s, _, objects := fixture(t)
	material := importFixture(t, s, "import", "manual source")
	c := RecordDistillationCommand{CommandMeta: meta("content", 1), InputRefs: sourceRefs(material), Stage: "content", ProcessingConfig: "manual-v1", OutputText: "Actual human summary"}
	one, err := s.RecordDistillation(ctx, human, c)
	if err != nil || one.Distillation.Provenance.Mode != "manual" || one.Distillation.Status != "succeeded" {
		t.Fatal("manual provenance lost")
	}
	encoded, err := json.Marshal(one.Distillation)
	if err != nil {
		t.Fatal("encode actual manual distillation")
	}
	var wire map[string]json.RawMessage
	if err = json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal("decode actual manual distillation")
	}
	for _, field := range []string{"related_refs", "related_ideas", "conflicts", "pending_questions", "goal_refs", "existing_assets", "missing_evidence"} {
		if string(wire[field]) != "[]" {
			t.Fatalf("%s must remain an empty wire array", field)
		}
	}
	detail, err := s.GetMaterial(ctx, human, material.Material.ID)
	if err != nil || len(detail.Distillations) != 1 || detail.Distillations[0].ID != one.Distillation.ID {
		t.Fatal("stored distillation missing from material detail")
	}
	publishes := objects.publishes.Load()
	c.CommandMeta, c.OutputText = meta("repeat", 1), "Changed answer without new question"
	two, err := s.RecordDistillation(ctx, human, c)
	if err != nil || !two.Reused || two.Distillation.ID != one.Distillation.ID || objects.publishes.Load() != publishes {
		t.Fatal("repeated fixed input performed work")
	}
	c.Question, c.CommandMeta = "What new evidence contradicts this?", meta("question", 1)
	three, err := s.RecordDistillation(ctx, human, c)
	if err != nil || three.Reused || three.Distillation.ID == one.Distillation.ID {
		t.Fatal("new question did not create a new record")
	}
	c.CommandMeta, c.Stage, c.Question, c.OutputText = meta("theme", 1), "topic", "", "Unresolved thematic comparison"
	c.PendingQuestions = []string{"Need another paper"}
	topic, err := s.RecordDistillation(ctx, human, c)
	if err != nil || len(topic.Distillation.PendingQuestions) != 1 || topic.Distillation.Stage != "topic" {
		t.Fatal("public topic stage or pending evidence lost")
	}
	invalid := c
	invalid.CommandMeta, invalid.Stage = meta("invalid-theme", 1), "theme"
	_, err = s.RecordDistillation(ctx, human, invalid)
	errorCode(t, err, apierrors.ValidationFailed)
	c.CommandMeta, c.Stage, c.OutputText = meta("project", 1), "project", "Project association"
	c.GoalRefs, c.ExistingAssets = []string{"goal"}, []string{"existing repository"}
	c.ExpectedImprovement, c.MinimumArtifact, c.MissingEvidence = "reduce a measured failure", "one checked fix", []string{"baseline result"}
	project, err := s.RecordDistillation(ctx, human, c)
	if err != nil || len(project.Distillation.MissingEvidence) != 1 {
		t.Fatal("project assessment lost")
	}
	updated := importFixture(t, s, "updated", "source changed")
	c.InputRefs, c.CommandMeta = sourceRefs(updated), meta("new-version", 1)
	newVersion, err := s.RecordDistillation(ctx, human, c)
	if err != nil || newVersion.Reused {
		t.Fatal("new source version reused old work")
	}
}

func TestConcurrentImportCommandsShareDurableJob(t *testing.T) {
	s, db, _ := fixture(t)
	var wg sync.WaitGroup
	ids := make(chan string, 12)
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, err := s.ImportMaterial(context.Background(), human, importCommand(fmt.Sprint(i), "same fixed manual snapshot"))
			ids <- result.JobID
			errs <- err
		}(i)
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal("concurrent manual import failed")
		}
	}
	one := ""
	for id := range ids {
		if one == "" {
			one = id
		}
		if id != one {
			t.Fatal("concurrent import created duplicate job")
		}
	}
	if countRows(t, db, "attention_jobs") != 1 || countRows(t, db, "attention_receipts") != 12 {
		t.Fatal("durable dedupe or per-caller receipts missing")
	}
}

func TestOutboxFailureRollsBackAggregateAndReceipt(t *testing.T) {
	s, db, _ := fixture(t)
	material := importFixture(t, s, "import", "manual source")
	restore := refuseOutbox(t, db)
	_, err := s.CreateOpportunity(context.Background(), human, OpportunityCommand{CommandMeta: meta("failure", 1), Title: "Candidate", EvidenceRefs: sourceRefs(material), Purpose: "use", NextStep: "next"})
	errorCode(t, err, apierrors.InternalError)
	restore()
	if countRows(t, db, "attention_opportunities") != 0 {
		t.Fatal("aggregate survived outbox storage failure")
	}
	var receipts int
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_receipts WHERE caller=? AND command=? AND key=?", "human:"+human.ID, "CreateOpportunity", "failure").Scan(&receipts); err != nil || receipts != 0 {
		t.Fatal("success receipt survived rollback")
	}
}

func TestFeedbackCASHistoryAndUnknownDimensions(t *testing.T) {
	ctx := context.Background()
	s, db, _ := fixture(t)
	material := importFixture(t, s, "import", "manual source")
	d, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("theme", 1), InputRefs: sourceRefs(material), Stage: "topic", ProcessingConfig: "manual-v1", OutputText: "Two unresolved ideas", PendingQuestions: []string{"Need related evidence"}})
	if err != nil {
		t.Fatal("save manual topic")
	}
	c := OpportunityCommand{CommandMeta: meta("opportunity", 1), Title: "Candidate", EvidenceRefs: sourceRefs(material), Purpose: "Improve known asset", NextStep: "Read the missing baseline", DistillationIDs: []string{d.Distillation.ID}}
	opportunity, err := s.CreateOpportunity(ctx, human, c)
	if err != nil {
		t.Fatal("save manual candidate")
	}
	if opportunity.Opportunity.Dimensions.GoalProgress.Value != nil {
		t.Fatal("unknown filled with zero")
	}
	id := opportunity.Opportunity.ID
	review, err := s.ReviewOpportunity(ctx, human, id, ReviewOpportunityCommand{CommandMeta: meta("later", 1), Feedback: "later", Reason: "Not this week"})
	if err != nil || review.Version != 2 {
		t.Fatal("save later feedback")
	}
	got, err := s.GetOpportunity(ctx, human, id)
	if err != nil || got.Opportunity.State != "deferred" || got.Revisions[0].State != opportunity.Revisions[0].State || got.Reviews[0].Revision != 1 {
		t.Fatal("feedback destroyed history")
	}
	_, err = s.ReviewOpportunity(ctx, human, id, ReviewOpportunityCommand{CommandMeta: meta("stale", 1), Feedback: "later"})
	errorCode(t, err, apierrors.VersionConflict)
	got, err = s.GetOpportunity(ctx, human, id)
	if err != nil || len(got.Reviews) != 1 {
		t.Fatal("stale feedback partially persisted")
	}
	_, err = s.ReviewOpportunity(ctx, human, id, ReviewOpportunityCommand{CommandMeta: meta("reject-empty", 2), Feedback: "reject", Reason: "No"})
	errorCode(t, err, apierrors.EvidenceMissing)
	dims := Dimensions{Originality: DimensionScore{Reason: "Already known"}}
	if _, err = s.ReviewOpportunity(ctx, human, id, ReviewOpportunityCommand{CommandMeta: meta("reject", 2), Feedback: "reject", Reason: "Not novel", Dimensions: &dims}); err != nil {
		t.Fatal("save explicit dimension rejection")
	}
	got, err = s.GetOpportunity(ctx, human, id)
	if err != nil || got.Opportunity.State != "rejected" || got.Reviews[1].Dimensions.Originality.Reason != "Already known" {
		t.Fatal("dimension reason lost")
	}
	c.CommandMeta, c.NextStep = meta("revision", got.Opportunity.Version), "A different next step"
	revised, err := s.ReviseOpportunity(ctx, human, id, c)
	if err != nil || revised.Opportunity.Revision != 2 || len(revised.Revisions) != 2 || revised.Reviews[0].Revision != 1 {
		t.Fatal("revision overwrote review target")
	}
	persisted, err := s.GetMaterial(ctx, human, material.Material.ID)
	if err != nil || countRows(t, db, "attention_materials") != 1 || persisted.Material.Lifecycle != "active" {
		t.Fatal("later or reject discarded source material")
	}
}

func TestWithdrawnSourceCannotLeakThroughDerivedAgentQueries(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	material := importFixture(t, s, "import", "manual source")
	d, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("summary", 1), InputRefs: sourceRefs(material), Stage: "content", OutputText: "Original derived content", ProcessingConfig: "manual"})
	if err != nil {
		t.Fatal("save manual derived content")
	}
	opportunity, err := s.CreateOpportunity(ctx, human, OpportunityCommand{CommandMeta: meta("opportunity", 1), Title: "Derived candidate", EvidenceRefs: sourceRefs(material), Purpose: "use", NextStep: "next", DistillationIDs: []string{d.Distillation.ID}})
	if err != nil {
		t.Fatal("save candidate")
	}
	if _, err = s.UpdateMaterial(ctx, human, material.Material.ID, UpdateMaterialCommand{CommandMeta: meta("withdraw", 1), Lifecycle: "withdrawn"}); err != nil {
		t.Fatal("withdraw selected source")
	}
	_, err = s.GetContent(ctx, agent, material.Material.ID, 1)
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = s.GetOpportunity(ctx, agent, opportunity.Opportunity.ID)
	errorCode(t, err, apierrors.ScopeDenied)
	distillations, err := s.ListDistillations(ctx, agent)
	errorCode(t, err, apierrors.ScopeDenied)
	if len(distillations.Items) != 0 {
		t.Fatal("withdrawn derived content distributed")
	}
	opportunities, err := s.ListOpportunities(ctx)
	if err != nil || len(opportunities.Items) != 0 {
		t.Fatal("shared query distributed withdrawn source")
	}
	if _, err = s.GetOpportunity(ctx, human, opportunity.Opportunity.ID); err != nil {
		t.Fatal("human restricted history lost")
	}
}
