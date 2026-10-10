package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// cloneProgress mirrors the SQLite boundary: a load returns a fresh value (no
// shared map/slice references) so an in-memory mutation cannot leak into the
// stored record the way a JSON decode would not.
func cloneProgress(p domain.ProjectProgress) domain.ProjectProgress {
	data, _ := json.Marshal(p)
	var out domain.ProjectProgress
	_ = json.Unmarshal(data, &out)
	return out
}

type progressProjects struct {
	projects map[string]domain.LocalProject
	grants   map[string]domain.ProjectGrant
}

func (p *progressProjects) ListProjects(context.Context) ([]domain.LocalProject, error) {
	return nil, nil
}
func (p *progressProjects) LoadProject(_ context.Context, id string) (domain.LocalProject, error) {
	project, ok := p.projects[id]
	if !ok {
		return project, apierrors.NewNotFound("local project", id)
	}
	return project, nil
}
func (p *progressProjects) SaveProject(context.Context, domain.LocalProject, int) error { return nil }
func (p *progressProjects) LoadGrant(_ context.Context, projectID, agentID string) (domain.ProjectGrant, error) {
	g, ok := p.grants[projectID+"/"+agentID]
	if !ok {
		return g, apierrors.NewNotFound("project Agent grant", agentID)
	}
	return g, nil
}
func (p *progressProjects) SaveGrant(context.Context, domain.ProjectGrant) error { return nil }
func (p *progressProjects) ListSessions(context.Context, string) ([]domain.NativeSession, error) {
	return nil, nil
}
func (p *progressProjects) LoadSession(context.Context, string) (domain.NativeSession, error) {
	return domain.NativeSession{}, nil
}
func (p *progressProjects) SaveSession(context.Context, domain.NativeSession) error { return nil }

// progressStore is a CAS repository double. failAt / conflictAt inject a storage
// error or a version conflict on a specific save ordinal (1-based) so the test
// can target the pending save versus the accepted publication save.
type progressStore struct {
	records    map[string]domain.ProjectProgress
	saves      int
	failAt     int
	conflictAt int
}

func (s *progressStore) LoadProjectProgress(_ context.Context, id string) (domain.ProjectProgress, error) {
	return cloneProgress(s.records[id]), nil
}
func (s *progressStore) SaveProjectProgress(_ context.Context, p domain.ProjectProgress, expected int) error {
	s.saves++
	if s.failAt > 0 && s.saves == s.failAt {
		return errors.New("storage unavailable")
	}
	if s.conflictAt > 0 && s.saves == s.conflictAt {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "progress changed; reload before saving"}
	}
	if p.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict}
	}
	s.records[p.ProjectID] = cloneProgress(p)
	return nil
}

type progressFiles struct {
	files []domain.ContextFile
	err   error
}

func (f progressFiles) CanonicalRoot(root string) (string, error) { return root, nil }
func (f progressFiles) ValidateProjectPath(root, rel string) (string, error) {
	return root + "/" + rel, nil
}
func (f progressFiles) ReadProjectFiles(context.Context, domain.LocalProject, []string) ([]domain.ContextFile, error) {
	return f.files, f.err
}
func (f progressFiles) DiscoverProjects(context.Context, string) ([]domain.ProjectCandidate, error) {
	return nil, nil
}

type progressProcessor struct {
	result domain.TextResult
	err    error
	calls  int
}

func (p *progressProcessor) ConfigurationID(context.Context, string) (string, error) {
	return "cfg", nil
}
func (p *progressProcessor) ProcessSelectedText(_ context.Context, _ domain.TextRequest) (domain.TextResult, error) {
	p.calls++
	return p.result, p.err
}

type progressRegistry struct{}

func (progressRegistry) Adapter(id string) (NativeAdapter, error) {
	if id == "codex" {
		return nil, nil
	}
	return nil, errors.New("unsupported")
}
func (progressRegistry) List() []string { return []string{"codex"} }

func newProgressService() (*projectProgressService, *progressProjects, *progressStore, *progressProcessor) {
	projects := &progressProjects{projects: map[string]domain.LocalProject{}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}}
	processor := &progressProcessor{}
	return NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "stage: implement feature"}}}, processor, progressRegistry{}), projects, store, processor
}

func consentProject() domain.LocalProject {
	return domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "codex"}}
}

func progressHuman(op string) domain.Caller {
	return domain.Caller{Kind: "human", ID: "h", OperationID: op}
}

const opA = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

func TestProgressCacheGetFillsAuthorizedProjectIdentity(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = consentProject()

	got, err := service.GetProjectProgress(ctx, human, "p1")
	if err != nil {
		t.Fatalf("absent cache read %v", err)
	}
	if got.SchemaVersion != 1 || got.ProjectID != "p1" {
		t.Fatalf("absent read lost the authorized project identity %+v", got)
	}
	if got.Status != "" || got.Source != "" || got.Revision != 0 || got.ObservedAt != (time.Time{}) {
		t.Fatalf("absent read fabricated a stage %+v", got)
	}
	if processor.calls != 0 {
		t.Fatal("absent cache read triggered a model turn")
	}
	if len(store.records) != 0 {
		t.Fatal("absent cache read wrote a progress row")
	}
}

func TestProgressHumanSetAndCacheRead(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = consentProject()

	percent := 40
	for _, caller := range []domain.Caller{{}, {Kind: "agent", ID: "a", ProjectID: "p1"}} {
		if _, err := service.SetProjectProgress(ctx, caller, "p1", domain.ProgressInput{Status: "in_progress", Percent: &percent}); err == nil {
			t.Fatal("non-human progress write accepted")
		}
	}
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Status: strings.Repeat("x", 201)}); err == nil {
		t.Fatal("oversized status accepted")
	}
	bad := 101
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Status: "ok", Percent: &bad}); err == nil {
		t.Fatal("out-of-range percent accepted")
	}
	record, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Status: "in_progress", Percent: &percent})
	if err != nil || record.Source != "human" || record.Revision != 1 || record.Percent == nil || *record.Percent != 40 {
		t.Fatalf("human set %+v %v", record, err)
	}
	got, err := service.GetProjectProgress(ctx, human, "p1")
	if err != nil || got.Status != "in_progress" || got.Source != "human" {
		t.Fatalf("cached read %+v %v", got, err)
	}
	if processor.calls != 0 {
		t.Fatal("cache GET started a processor")
	}
	if len(store.records) != 1 {
		t.Fatal("expected one stored record")
	}
}

func TestProgressInferRecordsEvidenceAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	service, projects, _, processor := newProgressService()
	projects.projects["p1"] = consentProject()

	if _, err := service.InferProjectProgress(ctx, domain.Caller{Kind: "human", ID: "h", OperationID: "not-a-uuid"}, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("inference accepted a non-UUID operation identity")
	}
	model := "gpt-test"
	processor.result = domain.TextResult{Text: `{"status":"review","percent":75}`, NativeID: "nid", Model: &model}
	record, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || record.Source != "agent_inferred" || record.Status != "review" || record.Percent == nil || *record.Percent != 75 || record.NativeID != "nid" {
		t.Fatalf("inferred record %+v %v", record, err)
	}
	if len(record.Evidence) != 1 || record.Evidence[0].SourcePath != "TASK.md" || record.Evidence[0].Version != "v1" || record.Evidence[0].Kind != "task" {
		t.Fatalf("evidence not recorded from observed files %+v", record.Evidence)
	}
	if record.Operations[opA].Result == nil || record.Operations[opA].Result.Status != "review" {
		t.Fatalf("accepted receipt did not keep its result snapshot %+v", record.Operations[opA])
	}
	calls := processor.calls
	again, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || again.Status != "review" || processor.calls != calls {
		t.Fatalf("same operation identity re-inferred: %+v %v calls %d->%d", again, err, calls, processor.calls)
	}
	if _, err := service.InferProjectProgress(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: "p1", OperationID: "6ba7b810-9dad-11d1-80b4-00c04fd430c9"}, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("Agent-initiated inference accepted")
	}
}

func TestProgressInferRejectsInvalidModelOutput(t *testing.T) {
	ctx := context.Background()
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = consentProject()
	processor.result = domain.TextResult{Text: `not json at all`, NativeID: "nid"}
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("invalid model output accepted")
	} else if e, ok := err.(*apierrors.ServiceError); !ok || e.Code != apierrors.EvidenceMissing {
		t.Fatalf("wrong rejection %v", err)
	}
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("failed operation identity replayed")
	}
	if _, err := store.LoadProjectProgress(ctx, "p1"); err != nil {
		t.Fatal(err)
	}
}

func TestProgressInferPersistsPendingReceiptBeforeModel(t *testing.T) {
	ctx := context.Background()
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = consentProject()
	processor.err = &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "native delivery unknown"}
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("unknown delivery accepted")
	}
	record, _ := store.LoadProjectProgress(ctx, "p1")
	if record.Operations[opA].Status != "unknown" {
		t.Fatalf("unknown receipt not persisted %+v", record.Operations)
	}
	processor.err = nil
	processor.result = domain.TextResult{Text: `{"status":"done"}`}
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("unknown operation identity replayed after clearing error")
	}
}

func TestProgressInferRechecksConsentBeforePublication(t *testing.T) {
	ctx := context.Background()
	projects := &progressProjects{projects: map[string]domain.LocalProject{}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}}
	projects.projects["p1"] = consentProject()
	model := "gpt-test"
	// The processor revokes consent mid-flight, simulating a concurrent settings
	// change while the paid turn runs. The publication re-check must discard the
	// result rather than publish a stale inferred status.
	processor := &progressProcessor{result: domain.TextResult{Text: `{"status":"review"}`, NativeID: "nid", Model: &model}}
	service := NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "x"}}}, &revokingProcessor{next: processor, projects: projects}, progressRegistry{})
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("consent change mid-flight published a result")
	}
	record, _ := store.LoadProjectProgress(ctx, "p1")
	if record.Source == "agent_inferred" || record.Operations[opA].Status != "unknown" {
		t.Fatalf("revoked inference published %+v", record)
	}
}

type revokingProcessor struct {
	next     *progressProcessor
	projects *progressProjects
}

func (p *revokingProcessor) ConfigurationID(context.Context, string) (string, error) {
	return "cfg", nil
}
func (p *revokingProcessor) ProcessSelectedText(ctx context.Context, r domain.TextRequest) (domain.TextResult, error) {
	p.projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 2, ExternalModelCLI: "codex"}}
	return p.next.ProcessSelectedText(ctx, r)
}

func TestProgressAgentReadUsesGrant(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, _, _ := newProgressService()
	projects.projects["p1"] = consentProject()
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetProjectProgress(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: "p1"}, "p1"); err == nil {
		t.Fatal("Agent read accepted without grant")
	}
	projects.grants["p1/a"] = domain.ProjectGrant{ProjectID: "p1", AgentID: "a", Actions: []string{"read_context"}}
	if _, err := service.GetProjectProgress(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: "p1"}, "p1"); err != nil {
		t.Fatalf("Agent read with grant failed %v", err)
	}
	now := time.Now()
	projects.grants["p1/a"] = domain.ProjectGrant{ProjectID: "p1", AgentID: "a", Actions: []string{"read_context"}, RevokedAt: &now}
	if _, err := service.GetProjectProgress(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: "p1"}, "p1"); err == nil {
		t.Fatal("revoked grant still readable")
	}
}

// A human stage write must keep the durable operation receipts; otherwise a
// retry of an already accepted inference would re-charge the model.
func TestProgressHumanSetPreservesOperationReceipts(t *testing.T) {
	ctx := context.Background()
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = consentProject()
	model := "gpt-test"
	processor.result = domain.TextResult{Text: `{"status":"review","percent":70}`, NativeID: "nid", Model: &model}

	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err != nil {
		t.Fatalf("infer %v", err)
	}
	if _, err := service.SetProjectProgress(ctx, domain.Caller{Kind: "human", ID: "h"}, "p1", domain.ProgressInput{Status: "done"}); err != nil {
		t.Fatalf("human set after infer %v", err)
	}
	record, _ := store.LoadProjectProgress(ctx, "p1")
	if record.Source != "human" || record.Status != "done" {
		t.Fatalf("human set lost its own status %+v", record)
	}
	if record.Operations[opA].Status != "accepted" || record.Operations[opA].Result == nil {
		t.Fatalf("human set dropped the accepted receipt %+v", record.Operations)
	}
	// Retrying the same operation must return the original accepted result and
	// must not re-run the model.
	calls := processor.calls
	again, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || again.Status != "review" || again.Source != "agent_inferred" || again.Percent == nil || *again.Percent != 70 || processor.calls != calls {
		t.Fatalf("accepted retry returned the live human write instead of the original result: %+v %v calls %d->%d", again, err, calls, processor.calls)
	}
}

// A transient CAS conflict during accepted publication must be recovered by
// reloading and re-saving the already computed result, never by re-running the
// model turn. The computed snapshot is durably stored first (save 2) so the
// conflict on the accepted publish (save 3) is recoverable.
func TestProgressPublicationRecoversFromCASConflictWithoutNewModel(t *testing.T) {
	ctx := context.Background()
	projects := &progressProjects{projects: map[string]domain.LocalProject{"p1": consentProject()}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}, conflictAt: 3}
	processor := &progressProcessor{result: domain.TextResult{Text: `{"status":"review"}`, NativeID: "nid"}}
	service := NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "x"}}}, processor, progressRegistry{})

	record, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || record.Status != "review" || record.Source != "agent_inferred" {
		t.Fatalf("publication did not recover from a CAS conflict: %+v %v", record, err)
	}
	if processor.calls != 1 {
		t.Fatalf("CAS recovery re-ran the model turn %d times", processor.calls)
	}
	if got, _ := store.LoadProjectProgress(ctx, "p1"); got.Operations[opA].Status != "accepted" || got.Operations[opA].Result == nil {
		t.Fatalf("accepted result was not durably published %+v", got)
	}
}

// A persistent publication failure must surface as an error, never as a nil
// success that would falsely report a result the store never received.
func TestProgressPublicationFailureIsNotReportedSuccess(t *testing.T) {
	ctx := context.Background()
	projects := &progressProjects{projects: map[string]domain.LocalProject{"p1": consentProject()}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}, failAt: 2}
	processor := &progressProcessor{result: domain.TextResult{Text: `{"status":"review"}`, NativeID: "nid"}}
	service := NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "x"}}}, processor, progressRegistry{})

	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("failed computed persistence reported success")
	}
	// The pending receipt persisted before the model turn still blocks a replay.
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("unpublished operation was replayed after storage recovered")
	}
}

// A crash between the durable computed save and the accepted publish must be
// recoverable: the retry finds the computed snapshot and re-publishes it without
// running another model turn.
func TestProgressComputedResultRecoveredAfterPublicationFailure(t *testing.T) {
	ctx := context.Background()
	projects := &progressProjects{projects: map[string]domain.LocalProject{"p1": consentProject()}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}, failAt: 3}
	processor := &progressProcessor{result: domain.TextResult{Text: `{"status":"review"}`, NativeID: "nid"}}
	service := NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "x"}}}, processor, progressRegistry{})

	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("publication failure reported success")
	}
	computed, _ := store.LoadProjectProgress(ctx, "p1")
	if computed.Operations[opA].Status != "computed" || computed.Operations[opA].Result == nil {
		t.Fatalf("computed result was not durably stored %+v", computed.Operations)
	}
	calls := processor.calls
	again, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || again.Status != "review" || again.Source != "agent_inferred" || processor.calls != calls {
		t.Fatalf("computed result not re-published without a new model turn: %+v %v calls %d->%d", again, err, calls, processor.calls)
	}
	if got, _ := store.LoadProjectProgress(ctx, "p1"); got.Operations[opA].Status != "accepted" {
		t.Fatalf("recovered result not accepted %+v", got.Operations)
	}
}
