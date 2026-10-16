package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// cloneProgress 镜像 SQLite 边界：加载返回全新值（无共享 map/slice 引用），
// 内存中修改不会泄漏到存储记录。
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

// progressStore 是 CAS 存储替身。failAt / conflictAt 在指定保存序号（1-based）
// 注入存储错误或版本冲突，以便测试定位 pending 保存与 accepted 发布保存。
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
	// 处理器在付费推理运行期间撤销许可，模拟并发设置变更。
	// 发布前重新检查必须丢弃结果，不发布过时的推断状态。
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

// 人工阶段写入必须保留持久操作回执；否则已受理推理的重试会重复计费模型。
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
	// 重试同一操作必须返回原已受理结果，不重新运行模型。
	calls := processor.calls
	again, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}})
	if err != nil || again.Status != "review" || again.Source != "agent_inferred" || again.Percent == nil || *again.Percent != 70 || processor.calls != calls {
		t.Fatalf("accepted retry returned the live human write instead of the original result: %+v %v calls %d->%d", again, err, calls, processor.calls)
	}
}

// 发布期间的瞬时 CAS 冲突必须通过重新加载并重新保存已计算结果恢复，
// 不重新运行模型推理。已计算快照先持久化（保存 2），因此接受发布（保存 3）
// 上的冲突可恢复。
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

// 持久发布失败必须返回错误，不能作为 nil 成功虚假报告存储从未接收的结果。
func TestProgressPublicationFailureIsNotReportedSuccess(t *testing.T) {
	ctx := context.Background()
	projects := &progressProjects{projects: map[string]domain.LocalProject{"p1": consentProject()}, grants: map[string]domain.ProjectGrant{}}
	store := &progressStore{records: map[string]domain.ProjectProgress{}, failAt: 2}
	processor := &progressProcessor{result: domain.TextResult{Text: `{"status":"review"}`, NativeID: "nid"}}
	service := NewProjectProgressService(projects, store, progressFiles{files: []domain.ContextFile{{Path: "TASK.md", Version: "v1", Text: "x"}}}, processor, progressRegistry{})

	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("failed computed persistence reported success")
	}
	// 模型推理之前持久化的 pending 回执仍然阻止重放。
	if _, err := service.InferProjectProgress(ctx, progressHuman(opA), "p1", domain.ProgressCommand{Files: []string{"TASK.md"}}); err == nil {
		t.Fatal("unpublished operation was replayed after storage recovered")
	}
}

// 持久化已计算结果与接受发布之间的崩溃必须可恢复：
// 重试找到已计算快照并重新发布，不运行新的模型推理。
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
