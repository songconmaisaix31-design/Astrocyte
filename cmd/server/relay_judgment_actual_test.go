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
	environment, err := os.Hostname()
	if err != nil {
		t.Fatal("实际部署身份读取失败")
	}
	provider, err := assembleConfiguredRepositoryContext(&services, db, store, repositoryContextConfig{EnvironmentID: environment, CacheRoot: filepath.Join(root, "repository-context"), MaxFiles: 20000, MaxBytes: 134217728, MaxResults: 200, MaxDepth: 5}, registry)
	if err != nil {
		t.Fatal("生产快照组装失败")
	}
	closeJudgment, err := assembleRelayJudgment(ctx, &services, db, references, provider, core, relayJudgmentConfig{})
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
	request := func(method, path string, body, out any) {
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
		response, err := client.Do(req)
		if err != nil {
			t.Fatal("真实判断请求失败")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("真实判断 HTTP 返回 %d", response.StatusCode)
		}
		if out != nil && json.NewDecoder(response.Body).Decode(out) != nil {
			t.Fatal("真实判断响应解码失败")
		}
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
	if _, err := core.ReferenceMaterial(ctx, principal, space.Space.ID, attention.ReferenceMaterialCommand{CommandMeta: meta(), MaterialID: ref.MaterialID, Revision: ref.Revision}); err != nil {
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
	record, err := db.LoadResult(ctx, result.Judgment.ID)
	if err != nil || record.Result.ID != result.Judgment.ID {
		t.Fatal("实际判断未保存到原 SQLite")
	}
	var taste attention.TasteProfileDetail
	request("GET", "/api/v1/attention-taste-profile", nil, &taste)
	if taste.SchemaVersion != 1 || taste.Configured || taste.Profile != nil {
		t.Fatal("空库品味状态不符")
	}
	var nativeSessions int
	if db.Conn().QueryRowContext(ctx, "SELECT count(*) FROM local_agent_sessions").Scan(&nativeSessions) != nil || nativeSessions != 0 {
		t.Fatal("无模型组装检查产生了原生会话")
	}
}
