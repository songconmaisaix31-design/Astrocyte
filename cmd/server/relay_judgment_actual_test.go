package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayJudgmentProductionAssemblyActualSQLiteHTTP(t *testing.T) {
	relayJudgmentActualAssembly(t, true)
}

func TestRelayHumanAdmissionsWithoutJudgmentSnapshotsActualSQLiteHTTP(t *testing.T) {
	relayJudgmentActualAssembly(t, false)
}

func relayJudgmentActualAssembly(t *testing.T, withSnapshots bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	root := t.TempDir()
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal("实际判断数据库打开失败")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("实际判断数据库迁移失败")
	}
	store, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal("实际判断对象库打开失败")
	}
	publicRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal("公开仓库目录不可用")
	}
	readme, err := os.ReadFile(filepath.Join(publicRoot, "README.md"))
	if err != nil {
		t.Fatal("公开仓库实际 README 读取失败")
	}
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader([]string{publicRoot}), store, attention.ServiceOptions{})
	references := &projectReferenceBridge{attention: core}
	registry := agents.NewRegistry(filepath.Join(root, "native"))
	projects := workspace.NewLocalProjectService(db, references, agents.ProjectFiles{}, registry)
	services := httpapi.Services{Attention: core, LocalProjects: projects, WorkspaceS2: workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, references)}
	services.ExperienceGovernance = &experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: projects, registered: projects}
	environment, err := os.Hostname()
	if err != nil {
		t.Fatal("实际部署身份读取失败")
	}
	provider, err := assembleConfiguredRepositoryContext(&services, db, store, repositoryContextConfig{EnvironmentID: environment, CacheRoot: filepath.Join(root, "repository-context"), MaxFiles: 20000, MaxBytes: 134217728, MaxResults: 200, MaxDepth: 5}, registry)
	if err != nil {
		t.Fatal("生产快照组装失败")
	}
	if err := projects.ConfigureRepositoryCleanup(provider); err != nil {
		t.Fatal("实际主数据库未连接提交后清理通知")
	}
	var judgmentSnapshots workspace.ProjectSnapshotProvider
	if withSnapshots {
		judgmentSnapshots = provider
	}
	closeJudgment, err := assembleRelayJudgment(ctx, &services, db, references, judgmentSnapshots, core, relayJudgmentConfig{})
	if err != nil {
		t.Fatal("生产判断组装失败")
	}
	defer func() {
		if closeJudgment() != nil {
			t.Error("生产判断关闭失败")
		}
	}()
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "archive")
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", "")
	closeRelay, err := assembleRelay(&services, db, registry, store)
	if err != nil {
		t.Fatal("实际空 Relay 生产组装失败")
	}
	defer func() {
		if closeRelay() != nil {
			t.Error("实际空 Relay 关闭失败")
		}
	}()
	if err := assembleRelayCompletionSource(&services); err != nil {
		t.Fatal("生产完成来源组装失败")
	}
	completionReader, ok := services.Relay.(attention.ExperienceCompletionReader)
	if !ok {
		t.Fatal("生产 Relay 未提供完成来源读取接口")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("真实判断 HTTP 监听失败")
	}
	server := httpapi.NewServer(httpapi.Config{Services: services, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	defer func() {
		if server.Shutdown(context.Background()) != nil || <-served != http.ErrServerClosed {
			t.Error("真实判断 HTTP 关闭失败")
		}
	}()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("真实判断 session 不可用")
	}
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}
	url := "http://" + listener.Addr().String()
	csrf := ""
	bearer := ""
	contractSamples := map[string][]json.RawMessage{}
	defer func() {
		if t.Failed() {
			return
		}
		var nativeSessions int
		if db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM local_agent_sessions").Scan(&nativeSessions) != nil || nativeSessions != 0 {
			t.Error("无模型组装检查产生了原生会话")
			return
		}
		if outputRoot := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT"); outputRoot != "" {
			if !filepath.IsAbs(outputRoot) || os.MkdirAll(outputRoot, 0700) != nil {
				t.Error("实际契约响应输出目录不可用")
				return
			}
			name := "relay-product-contracts.json"
			if !withSnapshots {
				name = "relay-human-contracts-without-snapshots.json"
			}
			writeActualJSON(t, filepath.Join(outputRoot, name), contractSamples)
		}
	}()
	requestStatus := func(method, path string, body, out any, status int) {
		t.Helper()
		var encoded []byte
		if body != nil {
			var err error
			encoded, err = json.Marshal(body)
			if err != nil {
				t.Fatal("真实判断请求编码失败")
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, url+path, bytes.NewReader(encoded))
		if err != nil {
			t.Fatal("真实判断请求创建失败")
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrf)
			req.Header.Set("Idempotency-Key", uuid.NewString())
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal("真实判断请求失败")
		}
		defer response.Body.Close()
		if response.StatusCode != status {
			t.Fatalf("真实判断 HTTP 返回 %d", response.StatusCode)
		}
		var payload json.RawMessage
		if json.NewDecoder(response.Body).Decode(&payload) != nil {
			t.Fatal("真实判断响应解码失败")
		}
		if out != nil && json.Unmarshal(payload, out) != nil {
			t.Fatal("真实判断响应结构不符")
		}
		name := ""
		switch {
		case response.StatusCode >= 400:
			name = "ErrorV1"
		case path == "/api/v1/attention-taste-profile":
			name = "TasteProfileResultV1"
		case strings.HasSuffix(path, "/judgment-snapshots"):
			name = "JudgmentSnapshotResultV1"
		case strings.Contains(path, "/judgment-policies/"):
			name = "JudgmentPolicyResultV1"
		case strings.HasSuffix(path, "/judgments"):
			name = "JudgmentResultEnvelopeV1"
		case strings.HasSuffix(path, "/feedback") && method == "GET":
			name = "JudgmentFeedbackListV1"
		case strings.HasSuffix(path, "/feedback"):
			name = "JudgmentFeedbackResultV1"
		case strings.HasPrefix(path, "/api/v1/admissions/") || strings.HasSuffix(path, "/admissions"):
			name = "AdmissionResultV1"
		}
		if name != "" {
			contractSamples[name] = append(contractSamples[name], payload)
		}
	}
	request := func(method, path string, body, out any) {
		t.Helper()
		requestStatus(method, path, body, out, http.StatusOK)
	}
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	request("GET", "/api/v1/auth/session", nil, &session)
	csrf = session.CSRF
	principal := attention.Principal{ID: session.ID, Kind: session.Kind}
	caller := domain.Caller{ID: session.ID, Kind: session.Kind}
	meta := func() attention.CommandMeta {
		return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
	}
	job, err := core.ImportMaterial(ctx, principal, attention.ImportMaterialCommand{CommandMeta: meta(), Kind: "text", Adapter: "manual", SourceLocator: "https://github.com/songconmaisaix31-design/Astrocyte/blob/main/README.md", Title: "Astrocyte public README", ExportText: string(readme), ContentDigest: judgment.DigestText(string(readme)), SourceSpans: []string{"README.md"}})
	if err != nil {
		t.Fatal("公开 README 实际导入失败")
	}
	if progressed, err := core.ProcessNextJob(ctx); err != nil || !progressed {
		t.Fatal("公开 README 实际导入未处理")
	}
	imported, err := core.GetJob(ctx, principal, job.JobID)
	if err != nil || imported.MaterialID == nil || imported.MaterialRevision == nil {
		t.Fatal("公开 README 实际导入身份不完整")
	}
	ref := judgment.EvidenceRef{MaterialID: *imported.MaterialID, Revision: *imported.MaterialRevision, Digest: judgment.DigestText(string(readme))}
	space, err := core.CreateProjectSpace(ctx, principal, attention.ProjectSpaceCommand{CommandMeta: meta(), Title: "Astrocyte public assembly check"})
	if err != nil {
		t.Fatal("实际资料空间创建失败")
	}
	space, err = core.ReferenceMaterial(ctx, principal, space.Space.ID, attention.ReferenceMaterialCommand{CommandMeta: meta(), MaterialID: ref.MaterialID, Revision: ref.Revision})
	if err != nil {
		t.Fatal("实际固定资料关联失败")
	}
	project, err := projects.RegisterProject(ctx, caller, domain.RegisterProjectCommand{Name: "Astrocyte public checkout", Root: publicRoot, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal("实际公开项目登记失败")
	}
	access := relayCompletionProjectAccess{projects: projects}
	if owner, err := access.currentOwner(ctx, principal, project.ID); err != nil || owner != project.ID {
		t.Fatal("实际完成来源未绑定当前项目身份")
	}
	if evidence, err := completionReader.ReadExperienceCompletion(ctx, principal, project.ID, uuid.NewString()); err == nil || evidence.EventID != "" {
		t.Fatal("真实空 Relay store 产生了完成证据")
	}
	for _, candidate := range []attention.Principal{{ID: principal.ID, Kind: "agent"}, {Kind: "human"}} {
		_, err := access.currentOwner(ctx, candidate, project.ID)
		var denied *apierrors.ServiceError
		if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
			t.Fatal("未授权完成来源未被拒绝")
		}
		if _, err := completionReader.ReadExperienceCompletion(ctx, candidate, project.ID, uuid.NewString()); !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
			t.Fatal("生产完成来源转发未拒绝未授权身份")
		}
	}
	if err := sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		owner, err := access.currentOwner(tx.Context(), principal, project.ID)
		if err != nil {
			return err
		}
		if owner != project.ID {
			t.Fatal("完成来源事务内项目身份不符")
		}
		return nil
	}); err != nil {
		t.Fatal("完成来源未复用实际主数据库事务")
	}
	settings := project.Settings
	settings.AllowDirectory, settings.AllowedSubdirs = true, []string{"."}
	project, err = projects.SetProjectSettings(ctx, caller, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal("实际公开目录授权失败")
	}
	refs := []attention.SourceRef{{MaterialID: ref.MaterialID, Revision: ref.Revision, Locator: "https://github.com/songconmaisaix31-design/Astrocyte/blob/main/README.md"}}
	topic, err := core.RecordDistillation(ctx, principal, attention.RecordDistillationCommand{CommandMeta: meta(), InputRefs: refs, Stage: "topic", ProcessingConfig: "manual", OutputText: "人工记录：核查当前公开 README 的实际内容与待补证据。", PendingQuestions: []string{"公开 README 有哪些尚待验证的说明？"}})
	if err != nil {
		t.Fatal("公开资料人工记录保存失败")
	}
	newOpportunity := func(title string) attention.OpportunityDetail {
		t.Helper()
		var value attention.OpportunityDetail
		requestStatus("POST", "/api/v1/opportunities", attention.OpportunityCommand{CommandMeta: meta(), Title: title, EvidenceRefs: refs, DistillationIDs: []string{topic.Distillation.ID}, Purpose: "人工核查选定公开资料的内容与证据。", NextStep: "检查原资料并保存待补问题。"}, &value, http.StatusCreated)
		return value
	}
	opportunity := newOpportunity("Review the actual public README")
	otherSpace, err := core.CreateProjectSpace(ctx, principal, attention.ProjectSpaceCommand{CommandMeta: meta(), Title: "Public review without a selected material reference"})
	if err != nil {
		t.Fatal("实际未选资料空间创建失败")
	}
	otherProject, err := projects.RegisterProject(ctx, caller, domain.RegisterProjectCommand{Name: "Public project without selected source", Root: publicRoot, SpaceID: otherSpace.Space.ID})
	if err != nil {
		t.Fatal("实际未选资料项目登记失败")
	}
	var denied apierrors.ErrorEnvelope
	requestStatus("POST", "/api/v1/opportunities/"+opportunity.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: opportunity.Opportunity.Revision, ProjectID: otherProject.ID}, &denied, http.StatusForbidden)
	if denied.Error == nil || denied.Error.Code != apierrors.ScopeDenied {
		t.Fatal("未选择资料的实际项目通过了来源授权")
	}
	missingProjectStatus := http.StatusForbidden
	if withSnapshots {
		missingProjectStatus = http.StatusNotFound
	}
	requestStatus("POST", "/api/v1/opportunities/"+opportunity.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: opportunity.Opportunity.Revision, ProjectID: uuid.NewString()}, &denied, missingProjectStatus)
	grant, err := projects.GrantProjectAgent(ctx, caller, project.ID, domain.ProjectGrant{AgentID: "public-admission-permission-reader", Actions: []string{"read_context"}})
	if err != nil {
		t.Fatal("实际项目读取授权创建失败")
	}
	token, err := projects.IssueProjectAgentToken(ctx, caller, project.ID, grant.AgentID)
	if err != nil {
		t.Fatal("实际项目鉴权令牌创建失败")
	}
	bearer = token.Token
	requestStatus("POST", "/api/v1/opportunities/"+opportunity.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: opportunity.Opportunity.Revision, ProjectID: project.ID}, &denied, http.StatusForbidden)
	bearer = ""
	if denied.Error == nil || denied.Error.Code != apierrors.ScopeDenied {
		t.Fatal("项目 Agent 通过了人工准入权限")
	}
	var admission attention.AdmissionResult
	requestStatus("POST", "/api/v1/opportunities/"+opportunity.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: opportunity.Opportunity.Revision, ProjectID: project.ID}, &admission, http.StatusCreated)
	if admission.Admission.Authority != "human" || admission.Admission.ProjectID != project.ID || admission.Draft == nil {
		t.Fatal("真实人工准入来源或草稿关系不符")
	}
	proposal, err := services.WorkspaceS2.ReadProposal(ctx, caller, admission.Draft.ID, 0)
	if err != nil || proposal.AdmissionID != admission.Admission.ID || proposal.AssignmentStatus != "unassigned" || proposal.GrantID != "" {
		t.Fatal("实际准入未保存未分配草稿或创建了执行授权")
	}
	request("GET", "/api/v1/admissions/"+admission.Admission.ID, nil, &admission)
	request("POST", "/api/v1/admissions/"+admission.Admission.ID+"/revocations", attention.RevokeAdmissionCommand{CommandMeta: meta(), Reason: "人工撤销本次公开资料核查准入。"}, &admission)
	if admission.Admission.Status != "revoked" {
		t.Fatal("实际准入撤销未持久化")
	}
	unassigned := newOpportunity("Review public README without assigning a project")
	var unassignedAdmission attention.AdmissionResult
	requestStatus("POST", "/api/v1/opportunities/"+unassigned.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: unassigned.Opportunity.Revision}, &unassignedAdmission, http.StatusCreated)
	if unassignedAdmission.Admission.ProjectID != "" || unassignedAdmission.Draft == nil {
		t.Fatal("未选项目的实际准入被自动绑定项目")
	}
	blocked := newOpportunity("Review the current selected source before removal")
	checkRemovedReference := func() {
		t.Helper()
		remove := meta()
		remove.ExpectedVersion = space.Space.Version
		if _, err := core.RemoveMaterialReference(ctx, principal, space.Space.ID, attention.RemoveMaterialReferenceCommand{CommandMeta: remove, MaterialID: ref.MaterialID}); err != nil {
			t.Fatal("实际项目固定引用撤回失败")
		}
		requestStatus("POST", "/api/v1/opportunities/"+blocked.Opportunity.ID+"/admissions", attention.AdmissionCommand{CommandMeta: meta(), ExpectedRevision: blocked.Opportunity.Revision, ProjectID: project.ID}, &denied, http.StatusForbidden)
		if denied.Error == nil || denied.Error.Code != apierrors.ScopeDenied {
			t.Fatal("当前项目已撤回的来源仍可准入")
		}
	}
	var missing apierrors.ErrorEnvelope
	requestStatus("POST", "/api/v1/experiences/completions", attention.CaptureCompletionExperienceCommand{CommandMeta: meta(), LocalProjectID: project.ID, EventID: uuid.NewString()}, &missing, http.StatusNotFound)
	if missing.Error == nil || missing.Error.Code != apierrors.NotFound {
		t.Fatal("生产完成经验桥未读取原空 Relay store")
	}
	var taste attention.TasteProfileDetail
	request("GET", "/api/v1/attention-taste-profile", nil, &taste)
	if taste.SchemaVersion != 1 || taste.Configured || taste.Profile != nil {
		t.Fatal("空库品味状态不符")
	}
	if !withSnapshots {
		checkRemovedReference()
		return
	}
	prefix := "/api/v1/local-projects/" + project.ID
	var snapshot struct {
		Snapshot domain.ProjectContextSnapshot `json:"snapshot"`
	}
	request("POST", prefix+"/judgment-snapshots", httpapi.JudgmentSnapshotCommand{Files: []string{"README.md"}}, &snapshot)
	if snapshot.Snapshot.Root != "" || snapshot.Snapshot.Generation == "" {
		t.Fatal("公开判断快照泄漏私有目录或缺少真实身份")
	}
	policy := judgment.Policy{Revision: 1, ProjectID: project.ID, ProjectSettingsRevision: project.Settings.Revision, WorktreeID: snapshot.Snapshot.WorktreeID, SnapshotID: snapshot.Snapshot.Generation, Snapshot: snapshot.Snapshot, SnapshotFiles: []string{"README.md"}, QuestionType: judgment.Truth, Backend: judgment.RuleBackend, Languages: []string{"en"}, Versions: judgment.Versions{Goal: "public-readme-presence", Taste: "human-selected-public-readme", Policy: "public-readme-policy", Model: "evidence-rules/v1", AccessEpoch: "actual-human-session"}, Budget: judgment.Budget{TimeoutMS: 10000, MaxInputBytes: 131072, MaxInputTokens: 2048}}
	request("PUT", prefix+"/judgment-policies/truth", httpapi.JudgmentPolicyCommand{Policy: policy}, nil)
	input := judgment.JudgmentRequest{SchemaVersion: "1.0", ProjectID: project.ID, WorktreeID: policy.WorktreeID, SnapshotID: policy.SnapshotID, Question: judgment.Question{ID: "public-readme-present", Type: judgment.Truth, Instructions: "Is the selected actual public README evidence present?", Rule: "all_evidence_present"}, EvidenceRefs: []judgment.EvidenceRef{ref}, Language: "en", Backend: judgment.RuleBackend, Versions: policy.Versions, Budget: policy.Budget}
	var result struct {
		Judgment judgment.JudgmentResult `json:"judgment"`
	}
	request("POST", prefix+"/judgments", input, &result)
	if result.Judgment.Decision != judgment.Recommend || result.Judgment.Value == nil || *result.Judgment.Value != 1 || result.Judgment.ProviderConfidence != nil || result.Judgment.Usage != nil {
		t.Fatal("实际规则判断、usage 或 confidence 不符")
	}
	feedbackPath := prefix + "/judgments/" + result.Judgment.ID + "/feedback"
	request("POST", feedbackPath, httpapi.JudgmentFeedbackCommand{ResultVersion: result.Judgment.Version, Kind: "adopt", Reason: "人工记录本次公开 README 规则判断。"}, nil)
	request("GET", feedbackPath, nil, nil)
	record, err := db.LoadResult(ctx, result.Judgment.ID)
	if err != nil || record.Result.ID != result.Judgment.ID {
		t.Fatal("实际判断未保存到原 SQLite")
	}
	checkRemovedReference()
}
