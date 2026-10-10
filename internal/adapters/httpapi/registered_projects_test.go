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
	metadataWrites   int
	metadataID       string
	metadata         domain.ProjectHumanMetadata
}

func (s *registeredProjectProbe) SetRegisteredProjectMetadata(_ context.Context, c domain.Caller, id string, metadata domain.ProjectHumanMetadata) (domain.ProjectSummary, error) {
	s.metadataWrites++
	s.caller, s.metadataID, s.metadata = c, id, metadata
	return domain.ProjectSummary{ID: id, Human: metadata}, nil
}

func (s *registeredProjectProbe) ListRegisteredProjects(_ context.Context, c domain.Caller) (domain.ProjectDiscoverySnapshot, error) {
	s.reads++
	s.caller = c
	return domain.ProjectDiscoverySnapshot{Status: "unknown", Projects: []domain.RegisteredProject{}, Failures: []domain.ProjectDiscoveryFailure{}}, nil
}

func TestRegisteredProjectMetadataRejectsForgedAuthorityAndStaleSession(t *testing.T) {
	p := &registeredProjectProbe{}
	s := NewServer(Config{Services: Services{RegisteredProjects: p}, AgentToken: "unscoped-agent", Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	path := "/api/v1/local-projects/registered/observed-project/metadata"
	body := `{"schema_version":1,"request_id":"notes","expected_version":1,"notes":"人类备注","review":"复盘","group":"研究","intent":"继续检查","archived":true,"revision":0}`
	cookie, csrf := bootstrapSession(t, s)
	for _, denied := range []struct {
		cookie               *http.Cookie
		csrf, bearer, origin string
	}{
		{nil, csrf, "", ""},
		{cookie, "", "", ""},
		{cookie, csrf, "Bearer unscoped-agent", ""},
		{cookie, csrf, "", "https://untrusted.invalid"},
		{&http.Cookie{Name: sessionCookie, Value: "expired"}, csrf, "", ""},
	} {
		w := attentionRequest(s, "PUT", path, body, denied.cookie, denied.csrf, denied.bearer, denied.origin)
		if w.Code != http.StatusForbidden || p.metadataWrites != 0 {
			t.Fatalf("denied write reached domain: %d %+v", w.Code, p)
		}
	}
	for _, forged := range []string{
		`"actor_kind":"human"`, `"updated_at":"2026-10-10T00:00:00Z"`,
		`"progress":100`, `"actions":["start"]`, `"model_consent":true`,
	} {
		w := attentionRequest(s, "PUT", path, body[:len(body)-1]+","+forged+"}", cookie, csrf, "", "")
		if w.Code != http.StatusBadRequest || p.metadataWrites != 0 {
			t.Fatalf("forged field reached domain: %s %d", forged, w.Code)
		}
	}
	for _, malformed := range []string{
		`{"schema_version":1,"request_id":"notes","expected_version":1}`,
		`{"schema_version":1,"request_id":"notes","expected_version":1,"notes":"","review":"","group":"","intent":"","archived":false,"revision":-1}`,
	} {
		w := attentionRequest(s, "PUT", path, malformed, cookie, csrf, "", "")
		if w.Code != http.StatusBadRequest || p.metadataWrites != 0 {
			t.Fatalf("malformed metadata reached domain: %d", w.Code)
		}
	}
	w := attentionRequest(s, "PUT", path, body, cookie, csrf, "", "")
	if w.Code != http.StatusOK || p.metadataWrites != 1 || p.metadataID != "observed-project" || p.caller.Kind != "human" || p.caller.ID != "local-human" || p.caller.OperationID != "caller-request-key" || p.metadata.UpdatedAt != nil || p.metadata.Revision != 0 || !p.metadata.Archived || p.metadata.Notes != "人类备注" {
		t.Fatalf("manual metadata transport mismatch: %d %+v", w.Code, p)
	}
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
