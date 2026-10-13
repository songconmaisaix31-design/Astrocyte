package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

type workbenchStub struct{ err error }

func (s workbenchStub) Read(_ context.Context, _ string, _ workbench.Query) (workbench.View, error) {
	return workbench.View{Events: &workbench.Events{Events: []workbench.Event{{ID: "safe", Kind: "task", Topic: "topic", Scope: "scope", TS: "2026-10-13T00:00:00Z"}}, Total: 1, Page: 1, PageSize: 50}}, s.err
}
func TestWorkbenchUsesSharedSessionOriginAndErrorV1(t *testing.T) {
	server := NewServer(Config{Services: Services{Workbench: workbenchStub{}}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	h := server.Handler()
	makeRequest := func(method, path string, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"health", "today", "focus", "topics/evidence", "agents", "events", "filters"} {
		w := makeRequest("GET", "/api/v1/cognition/"+path, nil, "")
		if w.Code != 403 || !strings.Contains(w.Body.String(), `"schema_version":1`) {
			t.Fatal("cognition bypassed shared auth")
		}
	}
	bootstrap := makeRequest("GET", "/api/v1/auth/session", nil, "")
	cookie := bootstrap.Result().Cookies()[0]
	w := makeRequest("GET", "/api/v1/cognition/events", cookie, "")
	if w.Code != 200 {
		t.Fatal("session rejected")
	}
	var v struct {
		Events []map[string]any `json:"events"`
	}
	if json.Unmarshal(w.Body.Bytes(), &v) != nil || len(v.Events[0]) != 6 {
		t.Fatal("not six fields")
	}
	for _, test := range []struct {
		method, path, origin string
		status               int
	}{{"GET", "/api/v1/cognition/events", "http://localhost:9999", 403}, {"GET", "/api/v1/cognition/events?page=bad", "", 400}, {"HEAD", "/api/v1/cognition/events", "", 405}, {"GET", "/api/v1/cognition/unknown", "", 404}, {"GET", "/api/v1/cognition/events?topic=%zz", "", 400}} {
		w = makeRequest(test.method, test.path, cookie, test.origin)
		if w.Code != test.status || strings.Contains(w.Body.String(), "raw_unique") {
			t.Fatal("shared guard or fixed error mismatch")
		}
	}
	h = NewServer(Config{Services: Services{Workbench: workbenchStub{errors.New("PRIVATE_CANARY SQL/path/raw_unique")}}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler()
	bootstrap = makeRequest("GET", "/api/v1/auth/session", nil, "")
	cookie = bootstrap.Result().Cookies()[0]
	w = makeRequest("GET", "/api/v1/cognition/events", cookie, "")
	if w.Code != 503 || strings.Contains(w.Body.String(), "PRIVATE_CANARY") || !strings.Contains(w.Body.String(), `"provider_unavailable"`) {
		t.Fatal("adapter error leaked")
	}
}
