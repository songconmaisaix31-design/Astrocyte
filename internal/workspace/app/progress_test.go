package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

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

type progressStore struct {
	records map[string]domain.ProjectProgress
}

func (s *progressStore) LoadProjectProgress(_ context.Context, id string) (domain.ProjectProgress, error) {
	return s.records[id], nil
}
func (s *progressStore) SaveProjectProgress(_ context.Context, p domain.ProjectProgress, expected int) error {
	if p.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict}
	}
	s.records[p.ProjectID] = p
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

func TestProgressHumanSetAndCacheRead(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1}}

	percent := 40
	for _, caller := range []domain.Caller{{}, {Kind: "agent", ID: "a", ProjectID: "p1"}} {
		if _, err := service.SetProjectProgress(ctx, caller, "p1", domain.ProgressInput{Stage: "in_progress", Percent: &percent}); err == nil {
			t.Fatal("non-human progress write accepted")
		}
	}
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Stage: strings.Repeat("x", 201)}); err == nil {
		t.Fatal("oversized stage accepted")
	}
	bad := 101
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Stage: "ok", Percent: &bad}); err == nil {
		t.Fatal("out-of-range percent accepted")
	}
	record, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Stage: "in_progress", Percent: &percent})
	if err != nil || record.Source != "human" || record.Revision != 1 || record.Percent == nil || *record.Percent != 40 {
		t.Fatalf("human set %+v %v", record, err)
	}
	got, err := service.GetProjectProgress(ctx, human, "p1")
	if err != nil || got.Stage != "in_progress" || got.Source != "human" {
		t.Fatalf("cached read %+v %v", got, err)
	}
	if processor.calls != 0 {
		t.Fatal("cache GET started a processor")
	}
	if len(store.records) != 1 {
		t.Fatal("expected one stored record")
	}
}

func TestProgressInferRequiresConsentAndRecordsBasis(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, store, processor := newProgressService()
	projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1}}
	if _, err := service.InferProjectProgress(ctx, human, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("inference accepted without model consent")
	}
	projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "codex"}}
	if _, err := service.InferProjectProgress(ctx, human, "p1", domain.ProgressCommand{Files: []string{}}); err == nil {
		t.Fatal("inference accepted empty files")
	}
	model := "gpt-test"
	processor.result = domain.TextResult{Text: `{"stage":"review","percent":75}`, NativeID: "nid", Model: &model}
	record, err := service.InferProjectProgress(ctx, human, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || record.Source != "agent_inferred" || record.Stage != "review" || record.Percent == nil || *record.Percent != 75 || record.NativeID != "nid" {
		t.Fatalf("inferred record %+v %v", record, err)
	}
	if len(record.Basis) != 1 || record.Basis[0].Path != "TASK.md" || record.Basis[0].Version != "v1" {
		t.Fatalf("basis not recorded from observed files %+v", record.Basis)
	}
	if processor.calls != 1 {
		t.Fatalf("expected one processor call, got %d", processor.calls)
	}
	if _, err := service.InferProjectProgress(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: "p1"}, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("Agent-initiated inference accepted")
	}
	_ = store
}

func TestProgressInferRejectsInvalidModelOutput(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, _, processor := newProgressService()
	projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1, ExternalModelCLI: "codex"}}
	processor.result = domain.TextResult{Text: `not json at all`, NativeID: "nid"}
	if _, err := service.InferProjectProgress(ctx, human, "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("invalid model output accepted")
	} else if e, ok := err.(*apierrors.ServiceError); !ok || e.Code != apierrors.EvidenceMissing {
		t.Fatalf("wrong rejection %v", err)
	}
}

func TestProgressAgentReadUsesGrant(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	service, projects, _, _ := newProgressService()
	projects.projects["p1"] = domain.LocalProject{ID: "p1", Root: "r", Settings: domain.ProjectSettings{Revision: 1}}
	if _, err := service.SetProjectProgress(ctx, human, "p1", domain.ProgressInput{Stage: "planning"}); err != nil {
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
