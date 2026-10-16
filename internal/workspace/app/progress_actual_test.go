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
