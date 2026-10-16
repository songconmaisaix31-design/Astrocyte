package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
)

type actualProgressHTTP struct {
	db       *sqlite.DB
	project  domain.LocalProject
	projects workspace.LocalProjects
	progress workspace.ProjectProgressService
	server   *httptest.Server
	client   *http.Client
	csrf     string
}

func openActualProgressHTTP(t *testing.T) *actualProgressHTTP {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
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
	human := domain.Caller{Kind: "human", ID: "local-human"}
	spaces := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := spaces.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "人工进度权限检查",
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	publicRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "公开工作树的人工进度", Root: publicRoot, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	// 人工与缓存路径不接模型处理器或原生 Registry。
	progress := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, nil, nil)
	api := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{LocalProjects: projects, Progress: progress}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	response, err := client.Get(server.URL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	decodeErr := json.NewDecoder(response.Body).Decode(&session)
	closeErr := response.Body.Close()
	if decodeErr != nil || closeErr != nil || response.StatusCode != http.StatusOK || session.CSRF == "" {
		t.Fatal("actual human session failed", response.StatusCode, decodeErr, closeErr)
	}
	return &actualProgressHTTP{db: db, project: project, projects: projects, progress: progress, server: server, client: client, csrf: session.CSRF}
}

func (s *actualProgressHTTP) request(t *testing.T, method string, body map[string]any, authenticated, csrf bool, bearer string, status int) map[string]any {
	t.Helper()
	var input []byte
	if body != nil {
		var err error
		input, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, s.server.URL+"/api/v1/local-projects/"+s.project.ID+"/progress", bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.NewString())
	if csrf {
		req.Header.Set("X-CSRF-Token", s.csrf)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	client := s.client
	if !authenticated {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	decodeErr := json.NewDecoder(response.Body).Decode(&result)
	closeErr := response.Body.Close()
	if decodeErr != nil || closeErr != nil || response.StatusCode != status {
		t.Fatalf("actual progress HTTP: got=%d expected=%d body=%+v decode=%v close=%v", response.StatusCode, status, result, decodeErr, closeErr)
	}
	return result
}

func (s *actualProgressHTTP) assertRows(t *testing.T, expected int) {
	t.Helper()
	for table, count := range map[string]int{"local_project_progress": expected, "local_agent_sessions": 0} {
		var got int
		if err := s.db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil || got != count {
			t.Fatal("actual row count differs", table, got, count, err)
		}
	}
}

// requestProjectPath 发送请求到任意项目 ID 和子路径，用于跨项目和 infer 拒绝检查。
func (s *actualProgressHTTP) requestProjectPath(t *testing.T, projectID, method, subpath string, body map[string]any, csrf bool, bearer string, status int) map[string]any {
	t.Helper()
	var input []byte
	if body != nil {
		var err error
		input, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, s.server.URL+"/api/v1/local-projects/"+projectID+subpath, bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.NewString())
	if csrf {
		req.Header.Set("X-CSRF-Token", s.csrf)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	response, err := s.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	decodeErr := json.NewDecoder(response.Body).Decode(&result)
	closeErr := response.Body.Close()
	if decodeErr != nil || closeErr != nil || response.StatusCode != status {
		t.Fatalf("actual progress HTTP: got=%d expected=%d body=%+v decode=%v close=%v", response.StatusCode, status, result, decodeErr, closeErr)
	}
	return result
}

func progressWrite(status string, percent int) map[string]any {
	return map[string]any{"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 1, "status": status, "percent": percent}
}

func TestProgressCacheGetFillsAuthorizedProjectIdentity(t *testing.T) {
	s := openActualProgressHTTP(t)
	got, err := s.progress.GetProjectProgress(context.Background(), domain.Caller{Kind: "human", ID: "local-human"}, s.project.ID)
	if err != nil {
		t.Fatalf("absent cache read %v", err)
	}
	if got.SchemaVersion != 1 || got.ProjectID != s.project.ID {
		t.Fatalf("absent read lost the authorized project identity %+v", got)
	}
	if got.Status != "" || got.Source != "" || got.Revision != 0 || !got.ObservedAt.IsZero() {
		t.Fatalf("absent read fabricated a stage %+v", got)
	}
	result := s.request(t, http.MethodGet, nil, true, false, "", http.StatusOK)
	wire, ok := result["progress"].(map[string]any)
	if !ok || wire["project_id"] != s.project.ID || wire["status"] != "unknown" || wire["source"] != nil || wire["observed_at"] != nil || wire["revision"] != float64(0) {
		t.Fatal("actual absent cache HTTP changed unobserved state", result)
	}
	s.assertRows(t, 0)
}

func TestProgressHumanSetAndCacheRead(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}
	percent := 40
	for _, caller := range []domain.Caller{{}, {Kind: "agent", ID: "progress-agent", ProjectID: s.project.ID}} {
		if _, err := s.progress.SetProjectProgress(ctx, caller, s.project.ID, domain.ProgressInput{Status: "in_progress", Percent: &percent}); err == nil {
			t.Fatal("non-human progress write accepted")
		}
	}
	s.request(t, http.MethodPut, progressWrite("in_progress", percent), false, false, "", http.StatusForbidden)
	s.request(t, http.MethodPut, progressWrite("in_progress", percent), true, false, "", http.StatusForbidden)
	s.request(t, http.MethodPut, progressWrite(strings.Repeat("x", 201), percent), true, true, "", http.StatusBadRequest)
	s.request(t, http.MethodPut, progressWrite("ok", 101), true, true, "", http.StatusBadRequest)
	s.assertRows(t, 0)
	s.request(t, http.MethodPut, progressWrite("in_progress", percent), true, true, "", http.StatusOK)
	got, err := s.progress.GetProjectProgress(ctx, human, s.project.ID)
	if err != nil || got.Status != "in_progress" || got.Source != "human" || got.Revision != 1 || got.Percent == nil || *got.Percent != percent || got.ObservedAt.IsZero() {
		t.Fatalf("human set and cached read %+v %v", got, err)
	}
	result := s.request(t, http.MethodGet, nil, true, false, "", http.StatusOK)
	wire, ok := result["progress"].(map[string]any)
	if !ok || wire["status"] != "in_progress" || wire["source"] != "human" || wire["percent"] != float64(percent) {
		t.Fatal("actual cached progress HTTP differs", result)
	}
	s.assertRows(t, 1)
}

func TestProgressAgentReadUsesGrant(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}
	agent := domain.Caller{Kind: "agent", ID: "progress-agent", ProjectID: s.project.ID}
	if _, err := s.progress.SetProjectProgress(ctx, human, s.project.ID, domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err == nil {
		t.Fatal("Agent read accepted without grant")
	}
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agent.ID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	if got, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err != nil || got.Status != "planning" {
		t.Fatalf("Agent read with actual grant failed %+v %v", got, err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agent.ID)
	if err != nil {
		t.Fatal(err)
	}
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusOK)
	s.request(t, http.MethodPut, progressWrite("done", 100), false, false, token.Token, http.StatusForbidden)
	if _, err := s.projects.RevokeProjectAgent(ctx, human, s.project.ID, agent.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err == nil {
		t.Fatal("revoked grant still readable")
	}
	result := s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusForbidden)
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var envelope apierrors.ErrorEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil || envelope.Error == nil || envelope.Error.Code != apierrors.ScopeDenied {
		t.Fatal("actual token revoke lost error identity", err)
	}
	s.assertRows(t, 1)
}

// 跨项目 GET 必须被会话中间件拒绝：Agent 持有项目 A 的有效 token，
// 访问项目 B 的进度端点时，scopedAgentRoute 检查 URL 中的项目 ID
// 与 token 绑定项目不匹配，返回 403。
func TestProgressCrossProjectGetRejected(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}

	// 创建第二个项目及独立空间
	spaceB, err := attention.NewAttentionService(sqlite.NewAttentionRepository(s.db), nil, nil, attention.ServiceOptions{}).
		CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{
			CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
			Title:       "跨项目拒绝检查",
		})
	if err != nil {
		t.Fatal(err)
	}
	publicRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	projectB, err := s.projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{
		Name: "跨项目拒绝目标", Root: publicRoot, SpaceID: spaceB.Space.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// 在项目 B 写入进度
	percent := 50
	if _, err := s.progress.SetProjectProgress(ctx, human, projectB.ID, domain.ProgressInput{Status: "in_progress", Percent: &percent}); err != nil {
		t.Fatal(err)
	}

	// 项目 A 的 Agent 授权与 token
	agentID := "cross-project-agent"
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}

	// 项目 A token 访问项目 A 端点 → 200
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusOK)

	// 同一 token 访问项目 B 端点 → 403
	s.requestProjectPath(t, projectB.ID, http.MethodGet, "/progress", nil, false, token.Token, http.StatusForbidden)

	// 项目 B 数据未受影响
	got, err := s.progress.GetProjectProgress(ctx, human, projectB.ID)
	if err != nil || got.Status != "in_progress" {
		t.Fatalf("project B progress modified by cross-project attempt %+v %v", got, err)
	}
}

// 合法 Grant 缺少 read_context 操作必须拒绝 GET。
// Grant 仅包含 "observe"，progressProject 检查 has(g.Actions, "read_context") 失败。
func TestProgressValidGrantWithoutReadContextRejected(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}

	if _, err := s.progress.SetProjectProgress(ctx, human, s.project.ID, domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}

	// Grant 只有 observe 操作，没有 read_context
	agentID := "no-read-agent"
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"observe"}}); err != nil {
		t.Fatal(err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}

	// GET 被拒绝 — token 通过认证但 scopedAgentRoute 放行后 progressProject 检查 action 失败
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusForbidden)

	// 直接服务层验证
	agent := domain.Caller{Kind: "agent", ID: agentID, ProjectID: s.project.ID}
	if _, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err == nil {
		t.Fatal("grant without read_context accepted for progress read")
	}

	s.assertRows(t, 1)
}

// 自然到期后的 GET 必须被拒绝。
// IssueProjectAgentToken 设置 ExpiresAt = now + 1h；通过直接 SQLite UPDATE
// 把 expires_at 改到过去，模拟到期。不修改系统时钟。
func TestProgressExpiredGrantGetRejected(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}

	if _, err := s.progress.SetProjectProgress(ctx, human, s.project.ID, domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}

	agentID := "expiry-agent"
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}

	// 到期前 GET 成功
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusOK)

	// 直接把 expires_at 改到过去
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	_, updateErr := s.db.Conn().Exec(
		`UPDATE local_agent_grants SET data = json_set(data, '$.expires_at', ?) WHERE project_id = ? AND agent_id = ?`,
		past, s.project.ID, agentID,
	)
	if updateErr != nil {
		t.Fatal(updateErr)
	}

	// 到期后 GET 被拒绝
	s.request(t, http.MethodGet, nil, false, false, token.Token, http.StatusForbidden)

	// 服务层也拒绝
	agent := domain.Caller{Kind: "agent", ID: agentID, ProjectID: s.project.ID}
	if _, err := s.progress.GetProjectProgress(ctx, agent, s.project.ID); err == nil {
		t.Fatal("expired grant still accepted for progress read")
	}

	s.assertRows(t, 1)
}

// Agent 不得发起模型推断：scopedAgentRoute 只允许 progress GET，
// POST /progress/infer 不在 Agent 路由白名单。
func TestProgressAgentInferRejected(t *testing.T) {
	s := openActualProgressHTTP(t)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}

	if _, err := s.progress.SetProjectProgress(ctx, human, s.project.ID, domain.ProgressInput{Status: "planning"}); err != nil {
		t.Fatal(err)
	}

	agentID := "infer-agent"
	if _, err := s.projects.GrantProjectAgent(ctx, human, s.project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := s.projects.IssueProjectAgentToken(ctx, human, s.project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}

	// Agent 的 POST /progress/infer 被拒绝
	inferBody := map[string]any{
		"schema_version": 1, "request_id": uuid.NewString(),
		"expected_version": 1, "files": []string{"TASK.md"},
	}
	s.requestProjectPath(t, s.project.ID, http.MethodPost, "/progress/infer", inferBody, false, token.Token, http.StatusForbidden)

	// Agent 的 PUT /progress 同样被拒绝
	s.request(t, http.MethodPut, progressWrite("done", 100), false, false, token.Token, http.StatusForbidden)

	// 原始进度未改变
	got, err := s.progress.GetProjectProgress(ctx, human, s.project.ID)
	if err != nil || got.Status != "planning" {
		t.Fatalf("progress modified by Agent infer attempt %+v %v", got, err)
	}
}

// 同一 SQLite 文件重启后，缓存进度与 Agent 权限必须仍然正确。
// 关闭原 DB、重新打开同一文件、重新创建服务和 HTTP 服务器、
// 验证人工缓存读取和 Agent token 授权均持久化。
func TestProgressRestartPreservesCacheAndPermissions(t *testing.T) {
	ctx := context.Background()
	dbDir, err := os.MkdirTemp("", "progress-restart-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dbDir) })
	dbPath := filepath.Join(dbDir, "state.sqlite")
	human := domain.Caller{Kind: "human", ID: "local-human"}

	// === 第一轮：打开、写入、关闭 ===
	db1, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := db1.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	spaces1 := attention.NewAttentionService(sqlite.NewAttentionRepository(db1), nil, nil, attention.ServiceOptions{})
	space1, err := spaces1.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "重启持久检查",
	})
	if err != nil {
		t.Fatal(err)
	}
	projects1 := workspace.NewLocalProjectService(db1, nil, agents.ProjectFiles{}, nil)
	publicRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects1.RegisterProject(ctx, human, domain.RegisterProjectCommand{
		Name: "重启持久项目", Root: publicRoot, SpaceID: space1.Space.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	progress1 := workspace.NewProjectProgressService(db1, db1, agents.ProjectFiles{}, nil, nil)
	percent := 55
	if _, err := progress1.SetProjectProgress(ctx, human, project.ID, domain.ProgressInput{Status: "in_progress", Percent: &percent}); err != nil {
		t.Fatal(err)
	}
	agentID := "restart-agent"
	if _, err := projects1.GrantProjectAgent(ctx, human, project.ID, domain.ProjectGrant{AgentID: agentID, Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := projects1.IssueProjectAgentToken(ctx, human, project.ID, agentID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db1.Close(); err != nil {
		t.Fatal(err)
	}

	// === 第二轮：重新打开同一文件 ===
	db2, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db2.Close() })
	projects2 := workspace.NewLocalProjectService(db2, nil, agents.ProjectFiles{}, nil)
	progress2 := workspace.NewProjectProgressService(db2, db2, agents.ProjectFiles{}, nil, nil)

	// 缓存读取仍然正确
	got, err := progress2.GetProjectProgress(ctx, human, project.ID)
	if err != nil || got.Status != "in_progress" || got.Percent == nil || *got.Percent != percent || got.Source != "human" {
		t.Fatalf("restart lost cached progress %+v %v", got, err)
	}

	// Agent 授权仍然有效
	agent := domain.Caller{Kind: "agent", ID: agentID, ProjectID: project.ID}
	if _, err := progress2.GetProjectProgress(ctx, agent, project.ID); err != nil {
		t.Fatalf("restart lost Agent grant: %v", err)
	}

	// Agent token 仍然可用
	api2 := httpapi.NewServer(httpapi.Config{
		Services: httpapi.Services{LocalProjects: projects2, Progress: progress2},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	server2 := httptest.NewServer(api2.Handler())
	t.Cleanup(server2.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, server2.URL+"/api/v1/local-projects/"+project.ID+"/progress", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if decodeErr := json.NewDecoder(response.Body).Decode(&result); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("restart lost Agent token validity: status=%d body=%+v", response.StatusCode, result)
	}
	wire, ok := result["progress"].(map[string]any)
	if !ok || wire["status"] != "in_progress" || wire["percent"] != float64(percent) {
		t.Fatalf("restart lost cached progress via HTTP %+v", result)
	}
}
