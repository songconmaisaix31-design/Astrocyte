package app_test

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
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
	sqlitedriver "modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"
)

type progressPreflightActual struct {
	ctx        context.Context
	db         *sqlite.DB
	service    workspace.ProjectProgressService
	registry   *agents.Registry
	baseline   []domain.LocalAgent
	nativeRoot string
	origin     string
	client     *http.Client
	human      domain.Caller
	csrf       string
	close      func()
}

func progressPreflightDirectory(t *testing.T) string {
	t.Helper()
	root := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("实际进度前置检查需要明确自有输出绝对目录")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(root, "progress-preflight-")
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func openProgressPreflightActual(t *testing.T, ctx context.Context, directory string) *progressPreflightActual {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	nativeRoot := filepath.Join(directory, "unused-native")
	if err := os.MkdirAll(nativeRoot, 0700); err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(nativeRoot)
	baseline, err := registry.SnapshotNative(ctx)
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	service := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, registry, registry)
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	server := httpapi.NewServer(httpapi.Config{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: httpapi.Services{LocalProjects: projects, Progress: service, Attention: core},
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	closed := false
	closeActual := func() {
		if closed {
			return
		}
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			t.Fatal(err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("实际进度 HTTP 服务没有正常退出", err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		closed = true
	}
	t.Cleanup(closeActual)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	s := &progressPreflightActual{ctx: ctx, db: db, service: service, registry: registry, baseline: baseline, nativeRoot: nativeRoot, origin: "http://" + listener.Addr().String(), client: client, close: closeActual}
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	s.request(t, http.MethodGet, "/api/v1/auth/session", "", nil, http.StatusOK, &session)
	if session.ID == "" || session.Kind != "human" || session.CSRF == "" {
		t.Fatal("实际 bootstrap 没有完整人类身份")
	}
	s.human, s.csrf = domain.Caller{ID: session.ID, Kind: session.Kind}, session.CSRF
	return s
}

func (s *progressPreflightActual) request(t *testing.T, method, route, operation string, command map[string]any, expected int, output any) {
	t.Helper()
	var body []byte
	if command != nil {
		command["schema_version"], command["request_id"], command["expected_version"] = 1, uuid.NewString(), 1
		var err error
		body, err = json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequestWithContext(s.ctx, method, s.origin+route, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if command != nil {
		if operation == "" {
			operation = uuid.NewString()
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", s.csrf)
		req.Header.Set("Idempotency-Key", operation)
	}
	response, err := s.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil || response.StatusCode != expected {
		t.Fatal("实际进度 HTTP 状态不符", method, route, response.StatusCode, expected, err)
	}
	for _, private := range []string{"\"operations\"", "\"pending_operation\""} {
		if bytes.Contains(data, []byte(private)) {
			t.Fatal("实际进度 HTTP 暴露内部回执")
		}
	}
	if err := json.Unmarshal(data, output); err != nil {
		t.Fatal(err)
	}
}

func (s *progressPreflightActual) register(t *testing.T) domain.LocalProject {
	t.Helper()
	var space attention.ProjectSpaceResult
	s.request(t, http.MethodPost, "/api/v1/project-spaces", "", map[string]any{"title": "实际进度前置公开项目"}, http.StatusCreated, &space)
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Project domain.LocalProject `json:"project"`
	}
	s.request(t, http.MethodPost, "/api/v1/local-projects", "", map[string]any{"name": "公开 README 配置前置", "root": root, "space_id": space.Space.ID}, http.StatusOK, &result)
	project := result.Project
	if project.ID == "" || project.Root != root || project.SpaceID != space.Space.ID || project.Settings.ExternalModelCLI != "" || project.Settings.AllowDirectory {
		t.Fatal("实际项目注册没有保留身份和默认许可")
	}
	settings := project.Settings
	settings.ExternalModelCLI, settings.AllowDirectory, settings.AllowedSubdirs = "codex", true, []string{"."}
	s.request(t, http.MethodPut, "/api/v1/local-projects/"+project.ID+"/settings", "", map[string]any{"expected_revision": project.Settings.Revision, "settings": settings}, http.StatusOK, &result)
	if result.Project.Settings.Revision != 2 || result.Project.Settings.ExternalModelCLI != "codex" || !result.Project.Settings.AllowDirectory || result.Project.Settings.AllowAgentControl || len(result.Project.Settings.AllowedTools) != 0 {
		t.Fatal("实际配置扩大了模型和目录以外的权限")
	}
	project = result.Project
	path := filepath.Join(project.Root, "README.md")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	files, err := (agents.ProjectFiles{}).ReadProjectFiles(s.ctx, project, []string{"README.md"})
	if err != nil || len(files) != 1 || files[0].Path != "README.md" || files[0].Text != string(original) || files[0].Version == "" {
		t.Fatal("生产 ProjectFiles 不能完整读取明确获准的公开 README", err)
	}
	t.Cleanup(func() {
		after, err := os.Stat(path)
		if err != nil {
			t.Error(err)
			return
		}
		current, err := os.ReadFile(path)
		if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) || !bytes.Equal(original, current) {
			t.Error("公开 README 原文件发生变化", err)
		}
	})
	return project
}

func (s *progressPreflightActual) assertNoNative(t *testing.T) {
	t.Helper()
	current, err := s.registry.SnapshotNative(s.ctx)
	if err != nil || !reflect.DeepEqual(current, s.baseline) {
		t.Fatal("实际前置失败改变了未探测 Registry", err)
	}
	entries, err := os.ReadDir(s.nativeRoot)
	if err != nil || len(entries) != 0 {
		t.Fatal("实际前置失败创建了原生输入或控制器目录", err)
	}
	var sessions int
	if err := s.db.Conn().QueryRowContext(s.ctx, "SELECT COUNT(*) FROM local_agent_sessions").Scan(&sessions); err != nil || sessions != 0 {
		t.Fatal("实际前置失败创建了原生会话", err)
	}
}

func progressPreflightExpectError(t *testing.T, envelope apierrors.ErrorEnvelope, code apierrors.Code) {
	t.Helper()
	if envelope.SchemaVersion != 1 || envelope.Error == nil || envelope.Error.Code != code || envelope.Error.RequestID == "" || envelope.Error.RequiredAction == "" {
		t.Fatal("实际前置 HTTP 没有准确 ErrorV1", code)
	}
}

func TestProgressPreflightActualColdConfigurationAndRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	directory := progressPreflightDirectory(t)
	s := openProgressPreflightActual(t, ctx, directory)
	project := s.register(t)
	operation := uuid.NewString()
	route := "/api/v1/local-projects/" + project.ID + "/progress"
	var failure apierrors.ErrorEnvelope
	s.request(t, http.MethodPost, route+"/infer", operation, map[string]any{"files": []string{"README.md"}}, http.StatusBadRequest, &failure)
	progressPreflightExpectError(t, failure, apierrors.EvidenceMissing)
	current, err := s.db.LoadProjectProgress(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	receipt, ok := current.Operations[operation]
	if !ok || len(current.Operations) != 1 || receipt.Action != "infer" || receipt.Status != "unknown" || receipt.Result != nil || receipt.CreatedAt.IsZero() || receipt.FinishedAt == nil || receipt.FinishedAt.Before(receipt.CreatedAt) || current.PendingOperation != "" || current.Revision != 2 || current.ProjectID != project.ID || current.Source != "" || current.NativeID != "" || current.Model != nil || !current.ObservedAt.IsZero() || len(current.Evidence) != 0 {
		t.Fatal("实际配置失败的保守回执或未观察结果不符")
	}
	assertRetry := func() {
		s.request(t, http.MethodPost, route+"/infer", operation, map[string]any{"files": []string{"README.md"}}, http.StatusConflict, &failure)
		progressPreflightExpectError(t, failure, apierrors.DeliveryUnknown)
		after, err := s.db.LoadProjectProgress(ctx, project.ID)
		if err != nil || !reflect.DeepEqual(after, current) {
			t.Fatal("原配置失败回执被同一操作重试改写", err)
		}
		s.assertNoNative(t)
	}
	assertRetry()
	var response struct {
		Progress domain.ProjectProgress `json:"progress"`
	}
	s.request(t, http.MethodPut, route, "", map[string]any{"status": "review", "summary": "人工检查当前进度；配置前置失败没有模型结论"}, http.StatusOK, &response)
	manual, err := s.db.LoadProjectProgress(ctx, project.ID)
	if err != nil || manual.Source != "human" || manual.Status != "review" || manual.Revision != current.Revision+1 || !reflect.DeepEqual(manual.Operations, current.Operations) || manual.PendingOperation != current.PendingOperation {
		t.Fatal("实际人工写入丢失配置失败回执", err)
	}
	current = manual
	assertRetry()
	s.close()
	s = openProgressPreflightActual(t, ctx, directory)
	storedProject, err := s.db.LoadProject(ctx, project.ID)
	if err != nil || !reflect.DeepEqual(storedProject, project) {
		t.Fatal("同库重启改变了实际项目和许可", err)
	}
	assertRetry()
	s.close()
	t.Log("实际配置前置 evidence_missing、生产保守 unknown 回执、人工保留、原 operation 和同库重启拒绝重发已验证；没有发生原生交付，原在途 unknown/计算/发布正向 NOT_RUN")
}

func TestProgressPreflightActualSQLiteReadOnlyRefusesPending(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s := openProgressPreflightActual(t, ctx, progressPreflightDirectory(t))
	project := s.register(t)
	if _, err := s.db.Conn().ExecContext(ctx, "PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := s.db.Conn().ExecContext(context.Background(), "PRAGMA query_only=OFF"); err != nil {
			t.Error(err)
		}
	})
	var readOnly int
	if err := s.db.Conn().QueryRowContext(ctx, "PRAGMA query_only").Scan(&readOnly); err != nil || readOnly != 1 {
		t.Fatal("自有 SQLite 连接未进入实际只读状态", err)
	}
	caller := s.human
	caller.OperationID = uuid.NewString()
	pending, err := s.service.InferProjectProgress(ctx, caller, project.ID, domain.ProgressCommand{Files: []string{"README.md"}})
	var databaseError *sqlitedriver.Error
	if !errors.As(err, &databaseError) || databaseError.Code() != sqlitelib.SQLITE_READONLY {
		t.Fatal("实际 pending 保存没有返回 SQLite READONLY", err)
	}
	receipt, ok := pending.Operations[caller.OperationID]
	if pending.SchemaVersion != 1 || pending.ProjectID != project.ID || pending.Revision != 1 || pending.PendingOperation != caller.OperationID || len(pending.Operations) != 1 || !ok || receipt.Status != "pending" || receipt.Action != "infer" || receipt.CreatedAt.IsZero() || receipt.FinishedAt != nil || receipt.Result != nil || pending.Source != "" || pending.NativeID != "" || !pending.ObservedAt.IsZero() {
		t.Fatal("真实拒写返回的内存 pending 状态不符")
	}
	var failure apierrors.ErrorEnvelope
	s.request(t, http.MethodPost, "/api/v1/local-projects/"+project.ID+"/progress/infer", uuid.NewString(), map[string]any{"files": []string{"README.md"}}, http.StatusInternalServerError, &failure)
	progressPreflightExpectError(t, failure, apierrors.InternalError)
	var records int
	if err := s.db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM local_project_progress").Scan(&records); err != nil || records != 0 {
		t.Fatal("只读 pending 保存失败留下了进度或回执", err)
	}
	s.assertNoNative(t)
	if _, err := s.db.Conn().ExecContext(ctx, "PRAGMA query_only=OFF"); err != nil {
		t.Fatal(err)
	}
	if err := s.db.Conn().QueryRowContext(ctx, "PRAGMA query_only").Scan(&readOnly); err != nil || readOnly != 0 {
		t.Fatal("自有 SQLite 连接没有恢复实际可写状态", err)
	}
	var response map[string]any
	s.request(t, http.MethodPut, "/api/v1/local-projects/"+project.ID+"/progress", "", map[string]any{"status": "review", "summary": "人工验证自有连接恢复写入"}, http.StatusOK, &response)
	manual, err := s.db.LoadProjectProgress(ctx, project.ID)
	if err != nil || manual.Source != "human" || manual.Status != "review" || manual.Revision != 1 || len(manual.Operations) != 0 || manual.PendingOperation != "" {
		t.Fatal("实际恢复可写后出现伪造推断或残留回执", err)
	}
	s.assertNoNative(t)
	t.Log("自有 SQLite query_only 实际拒绝 pending 保存、原始 READONLY 和 HTTP500、零回执及恢复写入已验证；没有模型、原生交付或计算结果")
}
