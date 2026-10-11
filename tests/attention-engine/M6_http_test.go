package attention_engine_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type m6HTTP struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
	csrf   string
}

func newM6HTTP(t *testing.T, handler http.Handler) *m6HTTP {
	t.Helper()
	server := httptest.NewServer(handler)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	h := &m6HTTP{t: t, server: server, client: &http.Client{Jar: jar}}
	response, err := h.client.Get(server.URL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	if err = json.NewDecoder(response.Body).Decode(&session); err != nil || response.StatusCode != 200 || session.CSRF == "" {
		t.Fatalf("HTTP session: %d %v", response.StatusCode, err)
	}
	h.csrf = session.CSRF
	return h
}
func (h *m6HTTP) Close() { h.server.Close() }
func m6Meta(id string, version int) app.CommandMeta {
	return app.CommandMeta{SchemaVersion: 1, RequestID: id, ExpectedVersion: version}
}
func (h *m6HTTP) Command(path string, body any, kind string, want int) []byte {
	h.t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		h.t.Fatal(err)
	}
	var meta app.CommandMeta
	if err = json.Unmarshal(data, &meta); err != nil {
		h.t.Fatal(err)
	}
	request, err := http.NewRequest("POST", h.server.URL+path, bytes.NewReader(data))
	if err != nil {
		h.t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", h.csrf)
	request.Header.Set("Idempotency-Key", meta.RequestID)
	if kind == "agent" {
		request.Header.Set("Authorization", "Bearer fixture-agent")
	}
	response, err := h.client.Do(request)
	if err != nil {
		h.t.Fatal(err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		h.t.Fatal(err)
	}
	if response.StatusCode != want {
		h.t.Fatalf("%s %s: %d want %d: %s", kind, path, response.StatusCode, want, payload)
	}
	return payload
}
