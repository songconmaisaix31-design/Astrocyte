package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelayCompletionExperienceActualAssemblyAndRequestBoundaries(t *testing.T) {
	ctx := context.Background()
	parent := filepath.Join("..", "..", "dist", "completion-experience")
	if configured := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT"); configured != "" {
		if !filepath.IsAbs(configured) {
			t.Fatal("测试输出根必须为绝对目录")
		}
		parent = filepath.Join(configured, "completion-experience")
	}
	if os.MkdirAll(parent, 0700) != nil {
		t.Fatal("实际检查输出目录不可用")
	}
	root, err := os.MkdirTemp(parent, "owned-")
	if err != nil {
		t.Fatal("实际独占检查目录不可用")
	}
	root, err = filepath.Abs(root)
	if err != nil {
		t.Fatal("实际检查目录无法解析")
	}
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_RELAY_SESSION_SOURCE", "archive")
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "")
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil || db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("实际 SQLite 和迁移不可用")
	}
	defer db.Close()
	registry := agents.NewRegistry(filepath.Join(root, "native"))
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	human := attention.Principal{Kind: "human", ID: "local-human"}
	space, err := core.CreateProjectSpace(ctx, human, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public completion assembly permission check"})
	if err != nil {
		t.Fatal("实际项目空间创建失败")
	}
	access, err := relayRoot()
	if err != nil {
		t.Fatal("实际接入根不可用")
	}
	publicRoot := filepath.Join(access, "verification", "public-project")
	if os.MkdirAll(publicRoot, 0700) != nil {
		t.Fatal("公开检查目录创建失败")
	}
	files := []string{"README.md", "LICENSE", "go.mod", ".gitignore"}
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", name))
		if err != nil || os.WriteFile(filepath.Join(publicRoot, name), data, 0600) != nil {
			t.Fatal("公开仓库文件复制失败")
		}
	}
	git := func(arguments ...string) string {
		t.Helper()
		command := exec.CommandContext(ctx, "git", append([]string{"-C", publicRoot}, arguments...)...)
		command.Stderr = io.Discard
		output, err := command.Output()
		if err != nil {
			t.Fatal("公开检查 Git 操作失败")
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "--initial-branch=main")
	git("config", "core.autocrlf", "false")
	git("add", "--", "README.md", "LICENSE", "go.mod", ".gitignore")
	git("-c", "user.name=Relay Verification", "-c", "user.email=relay@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "Public completion assembly input")
	encoded, err := json.Marshal(files)
	if err != nil {
		t.Fatal("公开文件选择编码失败")
	}
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", publicRoot)
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT", git("rev-parse", "HEAD"))
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES", string(encoded))
	project, err := projects.RegisterProject(ctx, domain.Caller{Kind: human.Kind, ID: human.ID}, domain.RegisterProjectCommand{Name: "Public completion assembly", Root: publicRoot, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal("真实公开项目登记失败")
	}
	store, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal("实际对象库不可用")
	}
	bridge := &experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: projects}
	services := httpapi.Services{Attention: core, LocalProjects: projects, ExperienceGovernance: bridge, WorkspaceS2: workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)}
	if assembleSwarmLedger(&services, db, registry, store) != nil {
		t.Fatal("实际 Swarm 生产装配失败")
	}
	closeRelay, err := assembleRelay(&services, db, registry, store)
	if err != nil {
		t.Fatal("实际空 Relay 生产装配失败")
	}
	defer func() {
		if closeRelay() != nil {
			t.Error("实际 Relay 正常关闭失败")
		}
	}()
	if assembleRelayCompletionSource(&services) != nil || assembleRelayCompletionInterpreter(&services, registry) != nil || bridge.completionInterpreter != nil {
		t.Fatal("未配置 Interpreter 没有保留明确不可用状态")
	}
	readers, err := relayPublicLaunchContexts(projects)
	if err != nil || len(readers) != 1 {
		t.Fatal("实际公开固定上下文不可用")
	}
	services.ExperienceGovernance = bridge
	if assembleRelayCompletionInterpreter(&services, registry, readers...) == nil {
		t.Fatal("没有显式 deadline 的部分配置未拒绝")
	}
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "30")
	if assembleRelayCompletionInterpreter(&services, registry) == nil {
		t.Fatal("缺少固定上下文的部分配置未拒绝")
	}
	if assembleRelayCompletionInterpreter(&services, agents.NewRegistry(filepath.Join(root, "other-native")), readers...) == nil {
		t.Fatal("另一 Registry 没有被拒绝")
	}
	if assembleRelayCompletionInterpreter(&services, registry, readers...) != nil {
		t.Fatal("准确生产 Interpreter 依赖装配失败")
	}
	interpreter, ok := bridge.completionInterpreter.(*relayCompletionInterpretation)
	if !ok || interpreter.backend == nil || interpreter.guard != services.Relay.(*relaySessionRuntime).sourceGuard {
		t.Fatal("原 Interpreter 或同一来源保护没有连接")
	}
	assertCode := func(err error, code apierrors.Code) {
		t.Helper()
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != code || strings.Contains(failure.Message, root) {
			t.Fatal("实际安全拒绝没有保留准确错误")
		}
	}
	_, _, err = interpreter.ReadContext(ctx, attention.ExperienceCompletionEvidence{})
	assertCode(err, apierrors.ScopeDenied)
	assertCode(interpreter.BeforeFirstInput(ctx, domain.NativeSession{}, nil, ""), apierrors.ScopeDenied)
	_, err = interpreter.DraftCompletionExperience(ctx, attention.CompletionExperienceRequest{})
	assertCode(err, apierrors.ScopeDenied)
	// 这是缺少原 operation 的拒绝输入，不创建完成或来源事实。
	requestCtx := context.WithValue(ctx, relayCompletionRequestKey{}, &relayCompletionRequest{principal: human})
	_, _, err = interpreter.ReadContext(requestCtx, attention.ExperienceCompletionEvidence{})
	assertCode(err, apierrors.EvidenceMissing)
	assertCode(interpreter.BeforeFirstInput(requestCtx, domain.NativeSession{}, nil, ""), apierrors.EvidenceMissing)
	if sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		_, _, failure := interpreter.readCurrentContext(tx.Context(), human, attention.ExperienceCompletionEvidence{})
		assertCode(failure, apierrors.ScopeDenied)
		return nil
	}) != nil {
		t.Fatal("实际事务内读取拒绝检查失败")
	}
	assertMissingSourceHTTP := func(selected httpapi.Services) {
		t.Helper()
		api := httpapi.NewServer(httpapi.Config{Services: selected, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		server := httptest.NewServer(api.Handler())
		defer server.Close()
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal("实际 HTTP Cookie 容器不可用")
		}
		client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
		response, err := client.Get(server.URL + "/api/v1/auth/session")
		if err != nil {
			t.Fatal("实际人工会话请求失败")
		}
		var session struct {
			CSRF string `json:"csrf_token"`
		}
		decodeErr, closeErr := json.NewDecoder(response.Body).Decode(&session), response.Body.Close()
		if decodeErr != nil || closeErr != nil || response.StatusCode != http.StatusOK || session.CSRF == "" {
			t.Fatal("实际人工会话不可用")
		}
		command := attention.CaptureCompletionExperienceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), ExpectedVersion: 1}, LocalProjectID: project.ID, EventID: uuid.NewString()}
		body, err := json.Marshal(command)
		if err != nil {
			t.Fatal("实际拒绝请求编码失败")
		}
		post, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/experiences/completions", bytes.NewReader(body))
		if err != nil {
			t.Fatal("实际拒绝请求创建失败")
		}
		post.Header.Set("Content-Type", "application/json")
		post.Header.Set("Idempotency-Key", uuid.NewString())
		post.Header.Set("X-CSRF-Token", session.CSRF)
		response, err = client.Do(post)
		if err != nil {
			t.Fatal("实际 Capture HTTP 请求失败")
		}
		var envelope apierrors.ErrorEnvelope
		decodeErr, closeErr = json.NewDecoder(response.Body).Decode(&envelope), response.Body.Close()
		if decodeErr != nil || closeErr != nil || response.StatusCode != http.StatusNotFound || envelope.Error == nil || envelope.Error.Code != apierrors.NotFound || envelope.Error.RequestID == "" {
			t.Fatal("当前真实 HTTP 没有拒绝缺失的原完成来源")
		}
		if strings.Contains(envelope.Error.Message, root) {
			t.Fatal("实际 HTTP 错误泄漏本机目录")
		}
	}
	assertMissingSourceHTTP(services)
	sources := bridge.completionSources.(*relayCompletionSources)
	if sources.accepted == nil || sources.reader == nil {
		t.Fatal("两个实际来源没有同时连接")
	}
	_, err = sources.ReadExperienceCompletion(ctx, attention.Principal{Kind: "agent", ID: human.ID}, project.ID, uuid.NewString())
	assertCode(err, apierrors.ScopeDenied)
	_, err = sources.ReadExperienceCompletion(ctx, human, project.SpaceID, uuid.NewString())
	assertCode(err, apierrors.ScopeDenied)
	assertCode(sources.AuthorizeExperienceCompletion(ctx, human, attention.ExperienceCompletionEvidence{}), apierrors.ScopeDenied)
	if sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		_, failure := sources.ReadExperienceCompletion(tx.Context(), human, project.ID, uuid.NewString())
		assertCode(failure, apierrors.ScopeDenied)
		return nil
	}) != nil {
		t.Fatal("共享来源没有拒绝事务内外部读取")
	}
	noRelay := services
	noRelay.Relay = nil
	noRelayBridge := &experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: projects}
	noRelay.ExperienceGovernance = noRelayBridge
	if assembleRelayCompletionSource(&noRelay) != nil {
		t.Fatal("独立 Swarm 来源装配失败")
	}
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "")
	if assembleRelayCompletionInterpreter(&noRelay, registry) != nil || noRelayBridge.completionInterpreter != nil {
		t.Fatal("独立 Swarm 未配置模型时没有保留不可用状态")
	}
	assertMissingSourceHTTP(noRelay)
	noRelay.ExperienceGovernance = noRelayBridge
	t.Setenv("ASTROCYTE_RELAY_SESSION_DEADLINE_SECONDS", "30")
	if assembleRelayCompletionInterpreter(&noRelay, registry, readers...) != nil {
		t.Fatal("独立 Swarm 的原 Registry 和显式 Interpreter 装配失败")
	}
	assertMissingSourceHTTP(noRelay)
	for _, table := range []string{"attention_engine_experiences", "experience_completion_sources", "local_agent_sessions"} {
		var count int
		if db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count) != nil || count != 0 {
			t.Fatal("拒绝检查创建了经验来源或原生会话")
		}
	}
	var interpretationReceipts int
	if db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_receipts WHERE command LIKE 'Experience/InterpretCompletion/%'").Scan(&interpretationReceipts) != nil || interpretationReceipts != 0 {
		t.Fatal("缺少来源的请求创建了解释操作收据")
	}
	observed, err := registry.SnapshotNative(ctx)
	if err != nil {
		t.Fatal("原 Registry 只读缓存检查失败")
	}
	for _, item := range observed {
		if item.Startable.Status != "unknown" {
			t.Fatal("无模型装配检查执行了原生启动")
		}
	}
}
