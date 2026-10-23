package main

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

func TestCompletionExperienceHTTPActualProjectBoundariesAndRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: "local-human"}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: testAttentionSpaces(db)})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "Human-created completion permission check",
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	publicRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Public completion permission check", Root: publicRoot, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	grant, err := projects.GrantProjectAgent(ctx, human, project.ID, domain.ProjectGrant{AgentID: "completion-permission-check", Actions: []string{"discover"}})
	if err != nil {
		t.Fatal(err)
	}
	token, err := projects.IssueProjectAgentToken(ctx, human, project.ID, grant.AgentID)
	if err != nil {
		t.Fatal(err)
	}

	var server *httptest.Server
	var client *http.Client
	var csrf string
	start := func() {
		t.Helper()
		projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
		bridge := &experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: projects}
		api := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{ExperienceGovernance: bridge, LocalProjects: projects}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		server = httptest.NewServer(api.Handler())
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client = &http.Client{Jar: jar, Timeout: 5 * time.Second}
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
		csrf = session.CSRF
	}
	start()
	defer func() { server.Close() }()
	request := func(method, endpoint string, body any, authenticated, withCSRF bool, bearer string, status int, code apierrors.Code) {
		t.Helper()
		var input []byte
		if body != nil {
			var err error
			input, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, server.URL+endpoint, bytes.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", uuid.NewString())
		if withCSRF {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		selectedClient := client
		if !authenticated {
			selectedClient = &http.Client{Timeout: 5 * time.Second}
		}
		response, err := selectedClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var result apierrors.ErrorEnvelope
		decodeErr := json.NewDecoder(response.Body).Decode(&result)
		closeErr := response.Body.Close()
		if decodeErr != nil || closeErr != nil || response.StatusCode != status || result.Error == nil || result.Error.Code != code || result.Error.RequestID == "" {
			t.Fatalf("actual %s %s: status=%d expected=%d error=%+v decode=%v close=%v", method, endpoint, response.StatusCode, status, result.Error, decodeErr, closeErr)
		}
	}
	// 未观察到的 UUID 只作为拒绝输入，不创建完成来源或模型结果。
	command := attention.CaptureCompletionExperienceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), ExpectedVersion: 1}, LocalProjectID: project.ID, EventID: uuid.NewString()}
	const endpoint = "/api/v1/experiences/completions"
	request(http.MethodPost, endpoint, command, false, false, "", http.StatusForbidden, apierrors.ScopeDenied)
	request(http.MethodPost, endpoint, command, true, false, "", http.StatusForbidden, apierrors.ScopeDenied)
	request(http.MethodPost, endpoint, command, true, true, token.Token, http.StatusForbidden, apierrors.ScopeDenied)
	for _, wrongProject := range []string{project.Root, project.SpaceID, uuid.NewString()} {
		wrong := command
		wrong.LocalProjectID = wrongProject
		request(http.MethodPost, endpoint, wrong, true, true, "", http.StatusForbidden, apierrors.ScopeDenied)
	}
	wrongVersion := command
	wrongVersion.ExpectedVersion = 2
	request(http.MethodPost, endpoint, wrongVersion, true, true, "", http.StatusBadRequest, apierrors.ValidationFailed)
	request(http.MethodPost, endpoint, command, true, true, "", http.StatusNotImplemented, apierrors.UnsupportedCapability)
	evidencePath := "/api/v1/experiences/" + uuid.NewString() + "/evidence"
	request(http.MethodGet, evidencePath, nil, true, false, "", http.StatusBadRequest, apierrors.ValidationFailed)
	request(http.MethodGet, evidencePath+"?project_id="+project.SpaceID, nil, true, false, "", http.StatusForbidden, apierrors.ScopeDenied)
	request(http.MethodGet, evidencePath+"?project_id="+project.ID, nil, true, false, token.Token, http.StatusForbidden, apierrors.ScopeDenied)
	request(http.MethodGet, evidencePath+"?project_id="+project.ID, nil, true, false, "", http.StatusNotFound, apierrors.NotFound)
	for _, table := range []string{"attention_engine_experiences", "experience_completion_sources", "local_agent_sessions"} {
		var count int
		if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("actual rejection created completion evidence or a native session", table, count, err)
		}
	}
	server.Close()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	start()
	request(http.MethodPost, endpoint, command, true, true, "", http.StatusNotImplemented, apierrors.UnsupportedCapability)
	request(http.MethodGet, evidencePath+"?project_id="+project.ID, nil, true, false, "", http.StatusNotFound, apierrors.NotFound)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	request(http.MethodPost, endpoint, command, true, true, "", http.StatusInternalServerError, apierrors.InternalError)
}
