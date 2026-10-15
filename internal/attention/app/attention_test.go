package app

import (
	"context"
	"encoding/json"
	"errors"
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
	Materials       map[string]MaterialDetail
	Distillations   map[string]Distillation
	Opportunities   map[string]OpportunityDetail
	Jobs            map[string]Job
	Receipts        map[string]Receipt
	Events          []OutboxEvent
	Domains         map[string]MaterialDomain
	Spaces          map[string]ProjectSpace
	Profile         RankingProfileDetail
	TrackingSources map[string]TrackingSource
	SourceItems     map[string]SourceItem
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
