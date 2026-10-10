package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// Contract-local transport probe; this verifies dispatch identity and guard
// behavior, not actual native CLI execution or material authorization effects.
type scopedProjectProbe struct {
	workspaceapp.LocalProjects
	revoked bool
	reads   int
	caller  domain.Caller
}

func (s *scopedProjectProbe) AuthenticateAgentToken(_ context.Context, token string) (domain.Caller, error) {
	if token != "project-scoped-token" || s.revoked {
		return domain.Caller{}, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	return domain.Caller{ID: "granted-agent", Kind: "agent", ProjectID: "project-A"}, nil
}
func (s *scopedProjectProbe) ReadProjectContext(_ context.Context, c domain.Caller, p string, _ domain.ContextRequest) (domain.ContextPacket, error) {
	s.reads++
	s.caller = c
	return domain.ContextPacket{SchemaVersion: 1, ProjectID: p, Materials: []domain.ContextMaterial{}, Files: []domain.ContextFile{}}, nil
}

func TestScopedAgentTransportHumanRoutesAndCurrentRevoke(t *testing.T) {
	p := &scopedProjectProbe{}
	s := NewServer(Config{Services: Services{LocalProjects: p}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	body := `{"schema_version":1,"request_id":"contract-request","expected_version":1,"references":[],"files":[]}`
	for range 2 {
		w := attentionRequest(s, "POST", "/api/v1/local-projects/project-A/context", body, nil, "", "Bearer project-scoped-token", "")
		if w.Code != http.StatusOK || p.caller.Kind != "agent" || p.caller.ID != "granted-agent" || p.caller.ProjectID != "project-A" || p.caller.OperationID != "caller-request-key" {
			t.Fatalf("scoped principal: %d %+v %s", w.Code, p.caller, w.Body.String())
		}
	}
	for _, tc := range []struct{ method, path string }{
		{"POST", "/api/v1/local-projects/project-B/context"},
		{"PUT", "/api/v1/local-projects/project-A/settings"},
		{"POST", "/api/v1/local-projects/project-A/grants"},
		{"POST", "/api/v1/local-projects/project-A/grants/revoke"},
		{"POST", "/api/v1/local-projects/project-A/agent-token"},
		{"GET", "/api/v1/local-projects"},
		{"GET", "/api/v1/materials"},
		{"GET", "/api/v1/auth/session"},
	} {
		w := attentionRequest(s, tc.method, tc.path, body, nil, "", "Bearer project-scoped-token", "")
		if w.Code != http.StatusForbidden || p.reads != 2 {
			t.Fatalf("Agent escaped scope %s: %d %s", tc.path, w.Code, w.Body.String())
		}
	}
	p.revoked = true
	w := attentionRequest(s, "POST", "/api/v1/local-projects/project-A/context", body, nil, "", "Bearer project-scoped-token", "")
	if w.Code != http.StatusForbidden || p.reads != 2 {
		t.Fatalf("revoked credential reached service: %d", w.Code)
	}
}
func TestLocalProjectWritesRequireHumanSessionCSRFAndRejectBodyIdentity(t *testing.T) {
	p := &scopedProjectProbe{}
	s := NewServer(Config{Services: Services{LocalProjects: p}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	cookie, csrf := bootstrapSession(t, s)
	body := `{"schema_version":1,"request_id":"contract-request","expected_version":1,"references":[],"files":[]}`
	w := attentionRequest(s, "POST", "/api/v1/local-projects/project-A/context", body, cookie, "", "", "")
	if w.Code != http.StatusForbidden || p.reads != 0 {
		t.Fatalf("missing csrf: %d", w.Code)
	}
	w = attentionRequest(s, "POST", "/api/v1/local-projects/project-A/context", body, cookie, csrf, "", "")
	if w.Code != http.StatusOK || p.caller.Kind != "human" || p.caller.ID != "local-human" {
		t.Fatalf("human identity: %d %+v", w.Code, p.caller)
	}
	for _, bad := range []string{
		`{"schema_version":1,"request_id":"x","expected_version":1,"references":[],"files":[],"actor_kind":"human"}`,
		`{"schema_version":1,"request_id":"x","expected_version":1,"references":[],"files":[],"OperationID":"fake"}`,
	} {
		w = attentionRequest(s, "POST", "/api/v1/local-projects/project-A/context", bad, cookie, csrf, "", "")
		if w.Code != http.StatusBadRequest || p.reads != 1 {
			t.Fatalf("body principal accepted: %d %s", w.Code, w.Body.String())
		}
	}
}
