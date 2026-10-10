package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type registeredProjectProbe struct {
	reads, refreshes int
	caller           domain.Caller
}

func (s *registeredProjectProbe) ListRegisteredProjects(_ context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	s.reads++
	s.caller = c
	return domain.ProjectDiscoverySnapshot{Status: "unknown", Projects: []domain.RegisteredProject{}, Failures: []domain.ProjectDiscoveryFailure{}}, nil
}

func (s *registeredProjectProbe) RefreshRegisteredProjects(ctx context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	s.refreshes++
	return s.ListRegisteredProjects(ctx, c)
}

func TestRegisteredProjectsHumanCacheReadAndExplicitRefresh(t *testing.T) {
	p := &registeredProjectProbe{}
	s := NewServer(Config{Services: Services{RegisteredProjects: p}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	path := "/api/v1/local-projects/registered"
	for _, auth := range []string{"", "Bearer unscoped-agent"} {
		w := attentionRequest(s, "GET", path, "", nil, "", auth, "")
		if w.Code != http.StatusForbidden || p.reads != 0 {
			t.Fatalf("unauthorized discovery: %d", w.Code)
		}
	}
	cookie, csrf := bootstrapSession(t, s)
	for range 2 {
		w := attentionRequest(s, "GET", path, "", cookie, "", "", "")
		if w.Code != http.StatusOK || p.refreshes != 0 || p.caller.Kind != "human" {
			t.Fatalf("cache read triggered discovery: %d %+v", w.Code, p)
		}
	}
	body := `{"schema_version":1,"request_id":"refresh","expected_version":1}`
	w := attentionRequest(s, "POST", path+"/refresh", body, cookie, "", "", "")
	if w.Code != http.StatusForbidden || p.refreshes != 0 {
		t.Fatalf("refresh without CSRF: %d", w.Code)
	}
	w = attentionRequest(s, "POST", path+"/refresh", body, cookie, csrf, "", "")
	if w.Code != http.StatusOK || p.refreshes != 1 || p.caller.OperationID != "caller-request-key" {
		t.Fatalf("explicit refresh: %d %+v", w.Code, p)
	}
	w = attentionRequest(s, "POST", path+"/refresh", `{"schema_version":1,"request_id":"refresh","expected_version":1,"root":"C:/"}`, cookie, csrf, "", "")
	if w.Code != http.StatusBadRequest || p.refreshes != 1 {
		t.Fatalf("refresh accepted arbitrary roots: %d", w.Code)
	}
}
