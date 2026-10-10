package agents

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestOpencodeIsolationEnv(t *testing.T) {
	t.Setenv("OPENCODE_CONFIG_DIR", "inherited-config")
	t.Setenv("XDG_CONFIG_HOME", "inherited-xdg")
	t.Setenv("KEEP_ME", "kept")
	env := opencodeIsolationEnv("owned-config", "user", "pass")
	values := map[string]string{}
	for _, kv := range env {
		key, val, _ := strings.Cut(kv, "=")
		values[key] = val
	}
	if values["OPENCODE_CONFIG_DIR"] != "owned-config" {
		t.Fatalf("config dir not redirected: %v", values["OPENCODE_CONFIG_DIR"])
	}
	if values["XDG_CONFIG_HOME"] != "owned-config" {
		t.Fatalf("xdg config not redirected: %v", values["XDG_CONFIG_HOME"])
	}
	if values["OPENCODE_DISABLE_PROJECT_CONFIG"] != "1" || values["OPENCODE_DISABLE_CLAUDE_CODE_PROMPT"] != "1" {
		t.Fatalf("project/claude prompt not disabled: %v", values)
	}
	if values["OPENCODE_PURE"] != "1" || values["OPENCODE_DISABLE_DEFAULT_PLUGINS"] != "1" {
		t.Fatalf("plugins not disabled: %v", values)
	}
	if values["OPENCODE_PERMISSION"] == "" || !strings.Contains(values["OPENCODE_PERMISSION"], `"*":"deny"`) || !strings.Contains(values["OPENCODE_PERMISSION"], `"bash":"deny"`) {
		t.Fatalf("permission deny not set as final override: %v", values["OPENCODE_PERMISSION"])
	}
	if values["OPENCODE_SERVER_USERNAME"] != "user" || values["OPENCODE_SERVER_PASSWORD"] != "pass" {
		t.Fatalf("basic auth not set: %v", values)
	}
	if values["KEEP_ME"] != "kept" {
		t.Fatal("unrelated environment dropped")
	}
}

// opencodeTestServer mimics the subset of the OpenCode serve HTTP API the
// driver depends on, and asserts loopback Basic auth is present on every call.
func opencodeTestServer(t *testing.T, username, password string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	auth := func(r *http.Request) bool {
		u, p, ok := r.BasicAuth()
		return ok && u == username && p == password
	}
	mux.HandleFunc("/global/health", func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"healthy":true,"version":"1.18.35"}`))
	})
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write([]byte(`{"id":"sess-1"}`))
	})
	mux.HandleFunc("/session/sess-1", func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write([]byte(`{"id":"sess-1"}`))
	})
	mux.HandleFunc("/session/sess-1/message", func(w http.ResponseWriter, r *http.Request) {
		if !auth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost {
			var body struct {
				Parts []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"parts"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Parts) != 1 || body.Parts[0].Text != "hello" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"info":{"role":"assistant","modelID":"test-model","providerID":"test-provider"},"parts":[{"type":"text","text":"reply"}]}`))
			return
		}
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"info":{"role":"user"},"parts":[{"type":"text","text":"hello"}]},{"info":{"role":"assistant"},"parts":[{"type":"text","text":"reply"}]}]`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	return httptest.NewServer(mux)
}

func newTestOpencodeProcess(server *httptest.Server, username, password string) *opencodeProcess {
	return &opencodeProcess{
		done:     make(chan struct{}),
		client:   server.Client(),
		baseURL:  server.URL,
		username: username,
		password: password,
		status:   "idle",
	}
}

func TestOpencodeSessionLifecycleOverHTTP(t *testing.T) {
	server := opencodeTestServer(t, "user", "pass")
	defer server.Close()
	p := newTestOpencodeProcess(server, "user", "pass")

	id, err := p.createSession(context.Background(), "project")
	if err != nil || id != "sess-1" {
		t.Fatalf("session create %q %v", id, err)
	}
	text, err := p.sendMessage(context.Background(), "sess-1", "hello")
	if err != nil || text != "reply" {
		t.Fatalf("message send %q %v", text, err)
	}
	p.mu.Lock()
	model, provider := p.model, p.provider
	p.mu.Unlock()
	if model != "test-model" || provider != "test-provider" {
		t.Fatalf("model/provider not observed from reply: %q/%q", model, provider)
	}
	events, err := p.readMessages(context.Background(), "sess-1")
	if err != nil || len(events) != 2 || events[0].Kind != "user_text" || events[0].Text != "hello" || events[1].Text != "reply" {
		t.Fatalf("context read %+v %v", events, err)
	}
	if err := p.verifySession(context.Background(), "sess-1"); err != nil {
		t.Fatalf("same-ID session verify %v", err)
	}
	if err := p.verifySession(context.Background(), "sess-missing"); err == nil {
		t.Fatal("missing session verified as resumable")
	}
}

func TestOpencodeRejectsUnauthenticatedAndUnknownReply(t *testing.T) {
	server := opencodeTestServer(t, "user", "pass")
	defer server.Close()
	p := newTestOpencodeProcess(server, "wrong", "wrong")
	if _, err := p.createSession(context.Background(), "project"); err == nil {
		t.Fatal("unauthenticated session create accepted")
	}
	good := newTestOpencodeProcess(server, "user", "pass")
	_, err := good.sendMessage(context.Background(), "sess-unknown", "hello")
	if err == nil {
		t.Fatal("unknown session message accepted")
	}
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ProviderUnavailable {
		t.Fatalf("unknown session not a provider rejection: %v", err)
	}
}

func TestOpencodeWaitHealth(t *testing.T) {
	server := opencodeTestServer(t, "user", "pass")
	defer server.Close()
	p := newTestOpencodeProcess(server, "user", "pass")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := waitHealth(ctx, p); err != nil {
		t.Fatalf("health wait %v", err)
	}
}

func TestOpencodeAdapterSnapshotRegistered(t *testing.T) {
	registry := NewRegistry(t.TempDir())
	items, err := registry.SnapshotNative(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range items {
		if item.ID == "opencode" && item.NativeAdapterRegistered {
			found = true
			if item.Configured.Status != "unknown" {
				t.Fatalf("opencode configured observation fabricated: %+v", item.Configured)
			}
		}
	}
	if !found {
		t.Fatal("opencode adapter not registered")
	}
	if _, err := registry.Adapter("opencode"); err != nil {
		t.Fatalf("opencode adapter lookup %v", err)
	}
}
