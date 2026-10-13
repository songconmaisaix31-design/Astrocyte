package httpapi

import (
	"context"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type sessionHTTPReadStub struct{ calls int }

func (s *sessionHTTPReadStub) ReadSessionEvents(_ context.Context, q workbench.SessionEventsQuery) (workbench.SessionEvents, error) {
	s.calls++
	return workbench.SessionEvents{Events: []workbench.Event{}, Coverage: []workbench.SessionCoverage{}, Page: q.Page, PageSize: q.PageSize}, nil
}

func TestRelaySessionHTTPGuardAndQuery(t *testing.T) {
	reader := &sessionHTTPReadStub{}
	server := httptest.NewServer(NewServer(Config{AgentToken: "public-fixture-token", Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: Services{RelaySessionEvents: reader}}).Handler())
	defer server.Close()
	request := func(method, path, origin, token string, cookie *http.Cookie) *http.Response {
		t.Helper()
		r, e := http.NewRequest(method, server.URL+path, nil)
		if e != nil {
			t.Fatal("public request unavailable")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		response, e := http.DefaultClient.Do(r)
		if e != nil {
			t.Fatal("transport unavailable")
		}
		return response
	}
	bootstrap := request("GET", "/api/v1/auth/session", "", "", nil)
	cookie := bootstrap.Cookies()[0]
	bootstrap.Body.Close()
	for _, c := range []struct {
		method, path, origin, token string
		cookie                      *http.Cookie
		status                      int
	}{
		{"GET", "/api/v1/relay/session-events", "", "", nil, 403},
		{"GET", "/api/v1/relay/session-events", "https://outside.invalid", "", cookie, 403},
		{"GET", "/api/v1/relay/session-events", "", "public-fixture-token", cookie, 403},
		{"HEAD", "/api/v1/relay/session-events", "", "", cookie, 405},
		{"GET", "/api/v1/relay/session-events?raw=true", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?page=1&page=2", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?page_size=101", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?task_id=PRIVATE_CANARY", "", "", cookie, 400},
		{"GET", "/api/v1/relay/session-events?start=2026-10-15T00:00:00Z&end=2026-10-14T00:00:00Z", "", "", cookie, 400},
	} {
		response := request(c.method, c.path, c.origin, c.token, c.cookie)
		response.Body.Close()
		if response.StatusCode != c.status {
			t.Fatalf("guard status=%d expected=%d", response.StatusCode, c.status)
		}
	}
	if reader.calls != 0 {
		t.Fatal("rejected request crossed read port")
	}
	response := request("GET", "/api/v1/relay/session-events?page_size=1", "", "", cookie)
	response.Body.Close()
	if response.StatusCode != 200 || reader.calls != 1 || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("authorized read unavailable")
	}
}
