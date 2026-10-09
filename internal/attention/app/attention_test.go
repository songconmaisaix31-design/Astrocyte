package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// This transactional fake exercises use cases and failure injection, not SQLite
// or live importer acceptance. W2/W4 own those adapter/integration checks.
type memoryState struct {
	Materials     map[string]MaterialDetail
	Distillations map[string]Distillation
	Opportunities map[string]OpportunityDetail
	Jobs          map[string]Job
	Receipts      map[string]Receipt
	Events        []OutboxEvent
	Domains       map[string]MaterialDomain
	Spaces        map[string]ProjectSpace
	Profile       RankingProfileDetail
}
type memoryRepo struct {
	mu        sync.Mutex
	state     memoryState
	failEvent bool
}
type memoryTx struct {
	state     *memoryState
	failEvent bool
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{state: memoryState{Materials: map[string]MaterialDetail{}, Distillations: map[string]Distillation{}, Opportunities: map[string]OpportunityDetail{}, Jobs: map[string]Job{}, Receipts: map[string]Receipt{}, Events: []OutboxEvent{}, Domains: map[string]MaterialDomain{}, Spaces: map[string]ProjectSpace{}}}
}
func clone[T any](value T) T {
	bytes, _ := json.Marshal(value)
	var copy T
	_ = json.Unmarshal(bytes, &copy)
	return copy
}
func (r *memoryRepo) WithTx(ctx context.Context, fn func(AttentionTx) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	copy := clone(r.state)
	for id, j := range r.state.Jobs {
		copied := copy.Jobs[id]
		copied.Payload = slices.Clone(j.Payload)
		copied.Caller = j.Caller
		copy.Jobs[id] = copied
	}
	if err := fn(&memoryTx{state: &copy, failEvent: r.failEvent}); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r.state = copy
	return nil
}
func (t *memoryTx) ListMaterials() ([]MaterialDetail, error) {
	rows := []MaterialDetail{}
	for _, r := range t.state.Materials {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) LoadMaterial(id string) (MaterialDetail, error) {
	r, ok := t.state.Materials[id]
	if !ok {
		return r, apierrors.NewNotFound("material", id)
	}
	return r, nil
}
func (t *memoryTx) FindMaterialBySourceKey(key string) (MaterialDetail, error) {
	for id, r := range t.state.Materials {
		for _, v := range r.Revisions {
			if v.SourceKey == key {
				return t.LoadMaterial(id)
			}
		}
	}
	return MaterialDetail{}, apierrors.NewNotFound("source", key)
}
func (t *memoryTx) SaveMaterial(r MaterialDetail, v int) error {
	old := t.state.Materials[r.Material.ID]
	if old.Material.Version != v {
		return serviceError(apierrors.VersionConflict, "CAS", "reload")
	}
	t.state.Materials[r.Material.ID] = r
	return nil
}
func (t *memoryTx) ListDistillations() ([]Distillation, error) {
	rows := []Distillation{}
	for _, r := range t.state.Distillations {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) FindDistillationByReuseKey(key string) (Distillation, error) {
	for _, r := range t.state.Distillations {
		if r.ReuseKey == key {
			return r, nil
		}
	}
	return Distillation{}, apierrors.NewNotFound("distillation", key)
}
func (t *memoryTx) SaveDistillation(r Distillation) error {
	t.state.Distillations[r.ID] = r
	return nil
}
func (t *memoryTx) ListOpportunities() ([]OpportunityDetail, error) {
	rows := []OpportunityDetail{}
	for _, r := range t.state.Opportunities {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) LoadOpportunity(id string) (OpportunityDetail, error) {
	r, ok := t.state.Opportunities[id]
	if !ok {
		return r, apierrors.NewNotFound("opportunity", id)
	}
	return r, nil
}
func (t *memoryTx) SaveOpportunity(r OpportunityDetail, v int) error {
	old := t.state.Opportunities[r.Opportunity.ID]
	if old.Opportunity.Version != v {
		return serviceError(apierrors.VersionConflict, "CAS", "reload")
	}
	t.state.Opportunities[r.Opportunity.ID] = r
	return nil
}
func (t *memoryTx) ListJobs() ([]Job, error) {
	rows := []Job{}
	for _, r := range t.state.Jobs {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) LoadJob(id string) (Job, error) {
	r, ok := t.state.Jobs[id]
	if !ok {
		return r, apierrors.NewNotFound("job", id)
	}
	return r, nil
}
func (t *memoryTx) FindJobByDedupeKey(key string) (Job, error) {
	for _, r := range t.state.Jobs {
		if r.DedupeKey == key {
			return r, nil
		}
	}
	return Job{}, apierrors.NewNotFound("job", key)
}
func (t *memoryTx) SaveJob(r Job, v int) error {
	old := t.state.Jobs[r.JobID]
	if old.Version != v {
		return serviceError(apierrors.VersionConflict, "CAS", "reload")
	}
	t.state.Jobs[r.JobID] = r
	return nil
}
func receiptKey(caller, command, key string) string { return caller + "/" + command + "/" + key }
func (t *memoryTx) LoadReceipt(caller, command, key string) (Receipt, error) {
	r, ok := t.state.Receipts[receiptKey(caller, command, key)]
	if !ok {
		return r, apierrors.NewNotFound("receipt", key)
	}
	return r, nil
}
func (t *memoryTx) SaveReceipt(caller, command, key string, r Receipt) error {
	t.state.Receipts[receiptKey(caller, command, key)] = r
	return nil
}
func (t *memoryTx) AppendEvent(e OutboxEvent) error {
	if t.failEvent {
		return errors.New("injected outbox failure")
	}
	t.state.Events = append(t.state.Events, e)
	return nil
}

func (t *memoryTx) ListMaterialDomains() ([]MaterialDomain, error) {
	rows := []MaterialDomain{}
	for _, r := range t.state.Domains {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) LoadMaterialDomain(id string) (MaterialDomain, error) {
	r, ok := t.state.Domains[id]
	if !ok {
		return r, apierrors.NewNotFound("domain", id)
	}
	return r, nil
}
func (t *memoryTx) SaveMaterialDomain(r MaterialDomain, v int) error {
	old := t.state.Domains[r.ID]
	if old.Version != v {
		return serviceError(apierrors.VersionConflict, "CAS", "reload")
	}
	t.state.Domains[r.ID] = r
	return nil
}
func (t *memoryTx) ListProjectSpaces() ([]ProjectSpace, error) {
	rows := []ProjectSpace{}
	for _, r := range t.state.Spaces {
		rows = append(rows, r)
	}
	return rows, nil
}
func (t *memoryTx) LoadProjectSpace(id string) (ProjectSpace, error) {
	r, ok := t.state.Spaces[id]
	if !ok {
		return r, apierrors.NewNotFound("space", id)
	}
	return r, nil
}
func (t *memoryTx) SaveProjectSpace(r ProjectSpace, v int) error {
	old := t.state.Spaces[r.ID]
	if old.Version != v {
		return serviceError(apierrors.VersionConflict, "CAS", "reload")
	}
	t.state.Spaces[r.ID] = r
	return nil
}

type sourceFunc func(context.Context, ImportMaterialCommand) (ImportedSource, error)

func (f sourceFunc) ReadSource(ctx context.Context, c ImportMaterialCommand) (ImportedSource, error) {
	return f(ctx, c)
}

type memoryObjects struct {
	mu            sync.Mutex
	values        map[string][]byte
	publishes     atomic.Int32
	beforePublish func()
}

func (o *memoryObjects) Publish(ctx context.Context, b []byte) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if o.beforePublish != nil {
		o.beforePublish()
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.publishes.Add(1)
	ref := digestBytes(b)
	o.values[ref] = slices.Clone(b)
	return ref, nil
}
func (o *memoryObjects) Read(ctx context.Context, ref string) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	b, ok := o.values[ref]
	if !ok {
		return nil, apierrors.NewNotFound("object", ref)
	}
	return slices.Clone(b), ctx.Err()
}

var human = Principal{ID: "local-user", Kind: "human"}
var agent = Principal{ID: "agent-reader", Kind: "agent"}

func meta(key string, version int) CommandMeta {
	return CommandMeta{SchemaVersion: 1, RequestID: "request-" + key, IdempotencyKey: key, ExpectedVersion: version}
}

func fixture(t *testing.T) (*Service, *memoryRepo, *memoryObjects) {
	t.Helper()
	r := newMemoryRepo()
	o := &memoryObjects{values: map[string][]byte{}}
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	s := NewAttentionService(r, sourceFunc(func(_ context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		return ImportedSource{SourceKey: c.SourceKey, SourceLocator: c.SourceLocator, Kind: c.Kind, Title: c.Title, Text: c.ExportText, SourceSpans: []string{"section 1"}, Provenance: Provenance{Processor: "test-reader", Mode: "manual", Version: "1", Source: c.SourceLocator}}, nil
	}), o, ServiceOptions{Clock: func() time.Time { return now }})
	return s, r, o
}
func importCommand(key, text string) ImportMaterialCommand {
	return ImportMaterialCommand{CommandMeta: meta(key, 1), SourceLocator: "https://example.test/source", SourceKey: "canonical-source", Kind: "text", ExportText: text, Adapter: "manual", Title: "Original source"}
}
func importFixture(t *testing.T, s *Service, key, text string) MaterialDetail {
	t.Helper()
	response, err := s.ImportMaterial(context.Background(), human, importCommand(key, text))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	j, err := s.GetJob(context.Background(), human, response.JobID)
	if err != nil || j.Status != "succeeded" || j.MaterialID == nil {
		t.Fatalf("import job %+v %v", j, err)
	}
	r, err := s.GetMaterial(context.Background(), human, *j.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func sourceRefs(r MaterialDetail) []SourceRef {
	return []SourceRef{{MaterialID: r.Material.ID, Revision: r.Material.CurrentRevision, Locator: r.Material.SourceLocator}}
}
func errorCode(t *testing.T, err error, want apierrors.Code) {
	t.Helper()
	var e *apierrors.ServiceError
	if !errors.As(err, &e) || e.Code != want {
		t.Fatalf("want %s, got %v", want, err)
	}
}

func TestImportReceiptsContentVersionsAndTransactionIO(t *testing.T) {
	s, r, o := fixture(t)
	assertUnlocked := func() {
		if !r.mu.TryLock() {
			t.Error("external I/O held transaction")
		} else {
			r.mu.Unlock()
		}
	}
	o.beforePublish = assertUnlocked
	base := s.sources
	s.sources = sourceFunc(func(ctx context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		assertUnlocked()
		return base.ReadSource(ctx, c)
	})
	row := importFixture(t, s, "first", "original content")
	c := importCommand("first", "original content")
	c.RequestID = "different-trace"
	replayed, err := s.ImportMaterial(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Status != "queued" {
		t.Fatal("receipt did not retain complete original result")
	}
	c.ExportText = "changed input with same idempotency key"
	_, err = s.ImportMaterial(context.Background(), human, c)
	errorCode(t, err, apierrors.VersionConflict)
	duplicate, err := s.ImportMaterial(context.Background(), human, importCommand("second-key", "original content"))
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.JobID != replayed.JobID || duplicate.Status != "succeeded" {
		t.Fatal("duplicate job created")
	}
	updated := importFixture(t, s, "changed", "updated content")
	if updated.Material.ID != row.Material.ID || len(updated.Revisions) != 2 || updated.Material.CurrentRevision != 2 {
		t.Fatalf("version history %+v", updated)
	}
	old, err := s.GetContent(context.Background(), human, row.Material.ID, 1)
	if err != nil || old.Text != "original content" {
		t.Fatal("old source overwritten")
	}
	for _, e := range r.state.Events {
		if e.Type == "attention.opportunity_admitted" || e.Type == "mission_created" {
			t.Fatal("import granted execution")
		}
	}
}

func TestConcurrentImportCommandsShareDurableJob(t *testing.T) {
	s, r, _ := fixture(t)
	var wg sync.WaitGroup
	ids := make(chan string, 12)
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result, err := s.ImportMaterial(context.Background(), human, importCommand(fmt.Sprint(i), "same fixed snapshot"))
			ids <- result.JobID
			errs <- err
		}(i)
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	one := ""
	for id := range ids {
		if one == "" {
			one = id
		}
		if id != one {
			t.Fatal("duplicate job")
		}
	}
	if len(r.state.Jobs) != 1 || len(r.state.Receipts) != 12 {
		t.Fatal("durable dedupe or caller receipts missing")
	}
}

func TestDistillationLayersReuseFixedInputsAndKeepPendingQuestion(t *testing.T) {
	s, _, o := fixture(t)
	material := importFixture(t, s, "import", "source")
	c := RecordDistillationCommand{CommandMeta: meta("content", 1), InputRefs: sourceRefs(material), Stage: "content", ProcessingConfig: "manual-v1", OutputText: "Actual human summary"}
	one, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	if one.Distillation.Provenance.Mode != "manual" || one.Distillation.Status != "succeeded" {
		t.Fatal("manual provenance lost")
	}
	encoded, _ := json.Marshal(one.Distillation)
	var wire map[string]json.RawMessage
	_ = json.Unmarshal(encoded, &wire)
	for _, field := range []string{"related_refs", "related_ideas", "conflicts", "pending_questions", "goal_refs", "existing_assets", "missing_evidence"} {
		if string(wire[field]) != "[]" {
			t.Fatalf("%s must be an empty wire array, got %s", field, wire[field])
		}
	}
	detail, err := s.GetMaterial(context.Background(), human, material.Material.ID)
	if err != nil || len(detail.Distillations) != 1 || detail.Distillations[0].ID != one.Distillation.ID {
		t.Fatal("stored distillation omitted from material detail")
	}
	publishes := o.publishes.Load()
	c.CommandMeta = meta("repeat", 1)
	c.OutputText = "Changed answer without new question"
	two, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil || !two.Reused || two.Distillation.ID != one.Distillation.ID || o.publishes.Load() != publishes {
		t.Fatal("repeated input did work")
	}
	c.Question = "What new evidence contradicts this?"
	c.CommandMeta = meta("question", 1)
	three, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil || three.Reused || three.Distillation.ID == one.Distillation.ID {
		t.Fatal("new question did not create work")
	}
	c.CommandMeta = meta("theme", 1)
	c.Stage = "topic"
	c.Question = ""
	c.OutputText = "Unresolved thematic comparison"
	c.PendingQuestions = []string{"Need another paper"}
	theme, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil || len(theme.Distillation.PendingQuestions) != 1 {
		t.Fatal("pending evidence lost")
	}
	if theme.Distillation.Stage != "topic" {
		t.Fatal("public topic stage was not preserved")
	}
	invalid := c
	invalid.CommandMeta = meta("invalid-theme", 1)
	invalid.Stage = "theme"
	_, err = s.RecordDistillation(context.Background(), human, invalid)
	errorCode(t, err, apierrors.ValidationFailed)
	c.CommandMeta = meta("project", 1)
	c.Stage = "project"
	c.OutputText = "Project association"
	c.GoalRefs = []string{"goal"}
	c.ExistingAssets = []string{"existing repository"}
	c.ExpectedImprovement = "reduce a measured failure"
	c.MinimumArtifact = "one checked fix"
	c.MissingEvidence = []string{"baseline result"}
	project, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil || len(project.Distillation.MissingEvidence) != 1 {
		t.Fatal("project assessment lost")
	}
	updated := importFixture(t, s, "updated", "source changed")
	c.InputRefs = sourceRefs(updated)
	c.CommandMeta = meta("new-version", 1)
	newVersion, err := s.RecordDistillation(context.Background(), human, c)
	if err != nil || newVersion.Reused {
		t.Fatal("new input version reused old work")
	}
}

func TestFeedbackCASHistoryAndUnknownDimensions(t *testing.T) {
	s, r, _ := fixture(t)
	material := importFixture(t, s, "import", "source")
	d, err := s.RecordDistillation(context.Background(), human, RecordDistillationCommand{CommandMeta: meta("theme", 1), InputRefs: sourceRefs(material), Stage: "topic", ProcessingConfig: "manual-v1", OutputText: "Two unresolved ideas", PendingQuestions: []string{"Need related evidence"}})
	if err != nil {
		t.Fatal(err)
	}
	c := OpportunityCommand{CommandMeta: meta("opportunity", 1), Title: "Candidate", EvidenceRefs: sourceRefs(material), Purpose: "Improve known asset", NextStep: "Read the missing baseline", DistillationIDs: []string{d.Distillation.ID}}
	opportunity, err := s.CreateOpportunity(context.Background(), human, c)
	if err != nil {
		t.Fatal(err)
	}
	if opportunity.Opportunity.Dimensions.GoalProgress.Value != nil {
		t.Fatal("unknown filled with zero")
	}
	id := opportunity.Opportunity.ID
	review, err := s.ReviewOpportunity(context.Background(), human, id, ReviewOpportunityCommand{CommandMeta: meta("later", 1), Feedback: "later", Reason: "Not this week"})
	if err != nil || review.Version != 2 {
		t.Fatal(err)
	}
	got, err := s.GetOpportunity(context.Background(), human, id)
	if err != nil || got.Opportunity.State != "deferred" || got.Revisions[0].State != opportunity.Revisions[0].State || got.Reviews[0].Revision != 1 {
		t.Fatal("feedback destroyed history")
	}
	_, err = s.ReviewOpportunity(context.Background(), human, id, ReviewOpportunityCommand{CommandMeta: meta("stale", 1), Feedback: "later"})
	errorCode(t, err, apierrors.VersionConflict)
	if len(r.state.Opportunities[id].Reviews) != 1 {
		t.Fatal("stale feedback partially persisted")
	}
	_, err = s.ReviewOpportunity(context.Background(), human, id, ReviewOpportunityCommand{CommandMeta: meta("reject-empty", 2), Feedback: "reject", Reason: "No"})
	errorCode(t, err, apierrors.EvidenceMissing)
	dims := Dimensions{Originality: DimensionScore{Reason: "Already known"}}
	_, err = s.ReviewOpportunity(context.Background(), human, id, ReviewOpportunityCommand{CommandMeta: meta("reject", 2), Feedback: "reject", Reason: "Not novel", Dimensions: &dims})
	if err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetOpportunity(context.Background(), human, id)
	if got.Opportunity.State != "rejected" || got.Reviews[1].Dimensions.Originality.Reason != "Already known" {
		t.Fatal("dimension reason lost")
	}
	c.CommandMeta = meta("revision", got.Opportunity.Version)
	c.NextStep = "A different next step"
	revised, err := s.ReviseOpportunity(context.Background(), human, id, c)
	if err != nil || revised.Opportunity.Revision != 2 || len(revised.Revisions) != 2 || revised.Reviews[0].Revision != 1 {
		t.Fatal("revision overwrote review target")
	}
	if len(r.state.Materials) != 1 || r.state.Materials[material.Material.ID].Material.Lifecycle != "active" {
		t.Fatal("later/reject discarded material")
	}
}

func TestOutboxFailureRollsBackAggregateAndReceipt(t *testing.T) {
	s, r, _ := fixture(t)
	material := importFixture(t, s, "import", "source")
	r.failEvent = true
	_, err := s.CreateOpportunity(context.Background(), human, OpportunityCommand{CommandMeta: meta("failure", 1), Title: "Candidate", EvidenceRefs: sourceRefs(material), Purpose: "use", NextStep: "next"})
	errorCode(t, err, apierrors.InternalError)
	if len(r.state.Opportunities) != 0 {
		t.Fatal("aggregate survived failed outbox")
	}
	if _, ok := r.state.Receipts[receiptKey("human:"+human.ID, "CreateOpportunity", "failure")]; ok {
		t.Fatal("success receipt survived rollback")
	}
}

func TestHumanQueriesMachineReadsPinAndValueAreIndependent(t *testing.T) {
	s, _, _ := fixture(t)
	material := importFixture(t, s, "import", "source")
	id := material.Material.ID
	for i := 0; i < 3; i++ {
		if _, err := s.GetMaterial(context.Background(), human, id); err != nil {
			t.Fatal(err)
		}
		if _, err := s.GetContent(context.Background(), human, id, 1); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		_, err := s.GetMaterial(context.Background(), agent, id)
		errorCode(t, err, apierrors.ScopeDenied)
	}
	got, _ := s.GetMaterial(context.Background(), human, id)
	if got.Material.HumanUsageCount != 0 || got.Material.AttentionScore != 0 || got.Material.AgentUsageCount != 0 {
		t.Fatal("reads increased human activity")
	}
	_, err := s.RecordUse(context.Background(), agent, id, RecordUseCommand{CommandMeta: meta("agent-use", got.Material.Version), Action: "adopt"})
	errorCode(t, err, apierrors.ScopeDenied)
	pinned, err := s.UpdateMaterial(context.Background(), human, id, UpdateMaterialCommand{CommandMeta: meta("pin", got.Material.Version), Pinned: true, Lifecycle: "active"})
	if err != nil || !pinned.Material.Pinned || pinned.Material.AttentionScore != 0 || len(pinned.Revisions) != 1 {
		t.Fatal("pin changed heat or content")
	}
	_, err = s.RecordUse(context.Background(), human, id, RecordUseCommand{CommandMeta: meta("adopt", pinned.Material.Version), Action: "adopt"})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := s.GetMaterial(context.Background(), human, id)
	now := s.options.Clock()
	s.options.Clock = func() time.Time { return now.Add(s.options.AttentionHalfLife) }
	after, _ := s.GetMaterial(context.Background(), human, id)
	if before.Material.AttentionScore != 3 || after.Material.AttentionScore != 1.5 || after.Material.LongTermValue != 1 || !after.Material.Pinned || after.Material.CurrentRevision != 1 {
		t.Fatal("decay affected pin/value/version")
	}
}

func TestWithdrawnSourceCannotLeakThroughDerivedAgentQueries(t *testing.T) {
	s, _, _ := fixture(t)
	material := importFixture(t, s, "import", "source")
	d, err := s.RecordDistillation(context.Background(), human, RecordDistillationCommand{CommandMeta: meta("summary", 1), InputRefs: sourceRefs(material), Stage: "content", OutputText: "Original derived content", ProcessingConfig: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	opportunity, err := s.CreateOpportunity(context.Background(), human, OpportunityCommand{CommandMeta: meta("opportunity", 1), Title: "Derived candidate", EvidenceRefs: sourceRefs(material), Purpose: "use", NextStep: "next", DistillationIDs: []string{d.Distillation.ID}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.UpdateMaterial(context.Background(), human, material.Material.ID, UpdateMaterialCommand{CommandMeta: meta("withdraw", 1), Lifecycle: "withdrawn"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.GetContent(context.Background(), agent, material.Material.ID, 1)
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = s.GetOpportunity(context.Background(), agent, opportunity.Opportunity.ID)
	errorCode(t, err, apierrors.ScopeDenied)
	distillations, err := s.ListDistillations(context.Background(), agent)
	errorCode(t, err, apierrors.ScopeDenied)
	if len(distillations.Items) != 0 {
		t.Fatal("derived content distributed")
	}
	opportunities, err := s.ListOpportunities(context.Background())
	if err != nil || len(opportunities.Items) != 0 {
		t.Fatal("shared list distributed withdrawn source")
	}
	if _, err = s.GetOpportunity(context.Background(), human, opportunity.Opportunity.ID); err != nil {
		t.Fatal("human restricted history lost")
	}
}
