package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 只观察真实服务的调用参数；鉴权、权限和上下文全部由生产服务处理。
type scopedProjectProbe struct {
	workspaceapp.LocalProjects
	reads            int
	caller           domain.Caller
	projectID, token string
}

func (s *scopedProjectProbe) AuthenticateAgentToken(ctx context.Context, token string) (domain.Caller, error) {
	return s.LocalProjects.(workspaceapp.ScopedAgentTokens).AuthenticateAgentToken(ctx, token)
}

func (s *scopedProjectProbe) ReadProjectContext(ctx context.Context, c domain.Caller, p string, request domain.ContextRequest) (domain.ContextPacket, error) {
	s.reads++
	s.caller = c
	return s.LocalProjects.ReadProjectContext(ctx, c, p, request)
}

func scopedProjectServer(t *testing.T) (*Server, *scopedProjectProbe) {
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
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	service := workspaceapp.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	human := domain.Caller{ID: "local-human", Kind: "human"}
	project, err := service.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "受控上下文", Root: t.TempDir(), SpaceID: "transport-context"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.GrantProjectAgent(ctx, human, project.ID, domain.ProjectGrant{AgentID: "granted-agent", Actions: []string{"read_context"}}); err != nil {
		t.Fatal(err)
	}
	token, err := service.IssueProjectAgentToken(ctx, human, project.ID, "granted-agent")
	if err != nil {
		t.Fatal(err)
	}
	p := &scopedProjectProbe{LocalProjects: service, projectID: project.ID, token: token.Token}
	return NewServer(Config{Services: Services{LocalProjects: p}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}), p
}

func TestScopedAgentTransportHumanRoutesAndCurrentRevoke(t *testing.T) {
	s, p := scopedProjectServer(t)
	body := `{"schema_version":1,"request_id":"contract-request","expected_version":1,"references":[],"files":[]}`
	for range 2 {
		w := attentionRequest(s, "POST", "/api/v1/local-projects/"+p.projectID+"/context", body, nil, "", "Bearer "+p.token, "")
		if w.Code != http.StatusOK || p.caller.Kind != "agent" || p.caller.ID != "granted-agent" || p.caller.ProjectID != p.projectID || p.caller.OperationID != "caller-request-key" {
			t.Fatalf("scoped principal: %d %+v %s", w.Code, p.caller, w.Body.String())
		}
	}
	for _, tc := range []struct{ method, path string }{
		{"POST", "/api/v1/local-projects/project-B/context"},
		{"PUT", "/api/v1/local-projects/" + p.projectID + "/settings"},
		{"POST", "/api/v1/local-projects/" + p.projectID + "/grants"},
		{"POST", "/api/v1/local-projects/" + p.projectID + "/grants/revoke"},
		{"POST", "/api/v1/local-projects/" + p.projectID + "/agent-token"},
		{"POST", "/api/v1/local-projects/" + p.projectID + "/sessions/probe"},
		{"GET", "/api/v1/local-projects"},
		{"GET", "/api/v1/local-projects/registered"},
		{"POST", "/api/v1/local-projects/registered/refresh"},
		{"GET", "/api/v1/materials"},
		{"GET", "/api/v1/auth/session"},
	} {
		w := attentionRequest(s, tc.method, tc.path, body, nil, "", "Bearer "+p.token, "")
		if w.Code != http.StatusForbidden || p.reads != 2 {
			t.Fatalf("Agent escaped scope %s: %d %s", tc.path, w.Code, w.Body.String())
		}
	}
	if _, err := p.RevokeProjectAgent(context.Background(), domain.Caller{ID: "local-human", Kind: "human"}, p.projectID, "granted-agent"); err != nil {
		t.Fatal(err)
	}
	w := attentionRequest(s, "POST", "/api/v1/local-projects/"+p.projectID+"/context", body, nil, "", "Bearer "+p.token, "")
	if w.Code != http.StatusForbidden || p.reads != 2 {
		t.Fatalf("revoked credential reached service: %d", w.Code)
	}
}
func TestLocalProjectWritesRequireHumanSessionCSRFAndRejectBodyIdentity(t *testing.T) {
	s, p := scopedProjectServer(t)
	cookie, csrf := bootstrapSession(t, s)
	body := `{"schema_version":1,"request_id":"contract-request","expected_version":1,"references":[],"files":[]}`
	w := attentionRequest(s, "POST", "/api/v1/local-projects/"+p.projectID+"/context", body, cookie, "", "", "")
	if w.Code != http.StatusForbidden || p.reads != 0 {
		t.Fatalf("missing csrf: %d", w.Code)
	}
	w = attentionRequest(s, "POST", "/api/v1/local-projects/"+p.projectID+"/context", body, cookie, csrf, "", "")
	if w.Code != http.StatusOK || p.caller.Kind != "human" || p.caller.ID != "local-human" {
		t.Fatalf("human identity: %d %+v", w.Code, p.caller)
	}
	for _, bad := range []string{
		`{"schema_version":1,"request_id":"x","expected_version":1,"references":[],"files":[],"actor_kind":"human"}`,
		`{"schema_version":1,"request_id":"x","expected_version":1,"references":[],"files":[],"OperationID":"fake"}`,
	} {
		w = attentionRequest(s, "POST", "/api/v1/local-projects/"+p.projectID+"/context", bad, cookie, csrf, "", "")
		if w.Code != http.StatusBadRequest || p.reads != 1 {
			t.Fatalf("body principal accepted: %d %s", w.Code, w.Body.String())
		}
	}
}
