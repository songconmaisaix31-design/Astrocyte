package agents

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestOpencodeIsolationEnv(t *testing.T) {
	t.Setenv("OPENCODE_CONFIG_DIR", "caller-config-input")
	t.Setenv("XDG_CONFIG_HOME", "caller-xdg-input")
	t.Setenv("OPENAI_API_KEY", "caller-credential-input")
	t.Setenv("NODE_OPTIONS", "caller-injection-input")
	t.Setenv("LANG", "C.UTF-8")
	root := t.TempDir()
	env, err := opencodeIsolationEnv(root, "caller-user", "caller-password")
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, kv := range env {
		key, val, _ := strings.Cut(kv, "=")
		values[key] = val
	}
	for key, name := range map[string]string{
		"HOME": "home", "USERPROFILE": "home", "OPENCODE_TEST_HOME": "home",
		"OPENCODE_CONFIG_DIR": "config", "XDG_CONFIG_HOME": "config",
		"XDG_DATA_HOME": "data", "XDG_STATE_HOME": "state", "XDG_CACHE_HOME": "cache",
		"TEMP": "tmp", "TMP": "tmp", "TMPDIR": "tmp",
		"APPDATA": "appdata", "LOCALAPPDATA": "local-appdata", "ProgramData": "programdata",
		"OPENCODE_TEST_MANAGED_CONFIG_DIR": "managed",
	} {
		if values[key] != filepath.Join(root, name) {
			t.Fatalf("owned source path was not established for %s", key)
		}
	}
	if values["OPENCODE_DB"] != filepath.Join(root, "data", "opencode.db") {
		t.Fatal("native database did not remain within owned data")
	}
	for _, key := range []string{"OPENCODE_DISABLE_PROJECT_CONFIG", "OPENCODE_DISABLE_CLAUDE_CODE_PROMPT", "OPENCODE_PURE", "OPENCODE_DISABLE_DEFAULT_PLUGINS", "OPENCODE_DISABLE_AUTOUPDATE", "OPENCODE_DISABLE_MODELS_FETCH"} {
		if values[key] != "1" {
			t.Fatalf("native source restriction is missing: %s", key)
		}
	}
	if values["OPENCODE_PERMISSION"] != opencodeDenyPermission || values["OPENCODE_SERVER_USERNAME"] != "caller-user" || values["OPENCODE_SERVER_PASSWORD"] != "caller-password" {
		t.Fatal("owned permission or authentication input was not retained")
	}
	if values["OPENAI_API_KEY"] != "" || values["NODE_OPTIONS"] != "" || values["LANG"] != "C.UTF-8" {
		t.Fatal("child environment inherited credentials/injection or lost the allowed locale")
	}
	if _, err := opencodeIsolationEnv("relative-root", "caller-user", "caller-password"); err == nil {
		t.Fatal("relative isolation root was accepted")
	}
}

func actualOpencodeProcess(t *testing.T) (*opencodeNative, domain.NativeRequest, domain.NativeSession, *opencodeProcess) {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	verification := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if verification == "" {
		verification = filepath.Join(root, ".verification")
	}
	if !filepath.IsAbs(verification) {
		t.Fatal("OpenCode verification requires an absolute output root")
	}
	if err := os.MkdirAll(verification, 0700); err != nil {
		t.Fatal(err)
	}
	if hasRedirectedAncestor(verification) {
		t.Fatal("owned verification directory has a redirected ancestor")
	}
	stateRoot, err := os.MkdirTemp(verification, "opencode-owned-")
	if err != nil {
		t.Fatal(err)
	}
	native := newOpencodeNative(stateRoot, make(chan struct{}, 4))
	projectID, sessionID := uuid.NewString(), uuid.NewString()
	request := domain.NativeRequest{
		SessionID: sessionID,
		Project:   domain.LocalProject{ID: projectID, Root: root},
		Session:   domain.NativeSession{ID: sessionID, ProjectID: projectID},
		Command:   domain.NativeCommand{DeadlineSeconds: 120},
		Packet:    domain.ContextPacket{SchemaVersion: 1, ProjectID: projectID, Mode: "selected_context"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	session, err := native.Start(ctx, request)
	if err != nil {
		t.Fatal("actual isolated OpenCode start failed", err)
	}
	if session.NativeID == "" || session.StopConfirmed || session.Status != "idle" || session.Version != "1.18.35" {
		t.Fatal("actual native start did not retain its identity/version/lifecycle")
	}
	process, err := native.get(session)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		observation, err := native.Stop(ctx, session)
		if err != nil || !observation.StopConfirmed {
			t.Error("actual native process stop was not confirmed", err)
		}
	})
	data, status, err := process.do(context.Background(), http.MethodGet, "/path", nil)
	if err != nil || status != http.StatusOK {
		t.Fatal("actual native source path query failed", err)
	}
	if err := os.WriteFile(filepath.Join(stateRoot, "actual-native-paths.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	var paths map[string]string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatal("actual native path schema is unavailable", err)
	}
	for _, key := range []string{"home", "config", "state"} {
		if !filepath.IsAbs(paths[key]) || !inside(stateRoot, paths[key]) {
			t.Fatalf("actual native %s path did not remain within owned state", key)
		}
	}
	if !strings.EqualFold(filepath.Clean(paths["directory"]), filepath.Clean(root)) {
		t.Fatal("actual native project directory differs from the approved public worktree")
	}
	for _, endpoint := range []string{"/config", "/mcp"} {
		data, status, err := process.do(context.Background(), http.MethodGet, endpoint, nil)
		if err != nil || status != http.StatusOK {
			t.Fatal("actual native owned config query failed", err)
		}
		if err := os.WriteFile(filepath.Join(stateRoot, "actual-native"+strings.ReplaceAll(endpoint, "/", "-")+".json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal("actual native owned config schema is unavailable", err)
		}
		if endpoint == "/mcp" && len(fields) != 0 {
			t.Fatal("actual native instance loaded an MCP server")
		}
		if endpoint == "/config" {
			var plugins []json.RawMessage
			if value, ok := fields["plugin"]; ok {
				if err := json.Unmarshal(value, &plugins); err != nil {
					t.Fatal(err)
				}
			}
			if len(plugins) != 0 {
				t.Fatal("actual native instance loaded an external plugin")
			}
		}
	}
	t.Logf("actual native version=%s native_id=%s pid=%d; owned sources; model turns=0; evidence=%s", session.Version, session.NativeID, process.cmd.Process.Pid, stateRoot)
	return native, request, session, process
}

func TestOpencodeSessionLifecycleOverHTTP(t *testing.T) {
	native, request, session, process := actualOpencodeProcess(t)
	ctx := context.Background()
	events, err := native.ReadContext(ctx, session)
	if err != nil || events == nil || len(events) != 0 {
		t.Fatal("actual new native session was not empty", err)
	}
	if err := process.verifySession(ctx, session.NativeID); err != nil {
		t.Fatal("actual same-ID session did not verify", err)
	}
	if err := process.verifySession(ctx, session.NativeID+"-caller-unknown"); err == nil {
		t.Fatal("unknown session verified as resumable")
	}
	observation, err := native.Stop(ctx, session)
	if err != nil || !observation.StopConfirmed {
		t.Fatal("actual first process exit was not confirmed", err)
	}
	request.Session = session
	request.Session.StopConfirmed = true
	resumed, err := native.Resume(ctx, request)
	if err != nil || resumed.NativeID != session.NativeID || resumed.StopConfirmed {
		t.Fatal("actual persisted native identity did not resume", err)
	}
	events, err = native.ReadContext(ctx, resumed)
	if err != nil || events == nil || len(events) != 0 {
		t.Fatal("actual no-model history changed on resume", err)
	}
	if process.model != "" || process.provider != "" || native.Capabilities()["send"].Status != "unknown" {
		t.Fatal("no-model lifecycle invented model readiness or send capability")
	}
}

func TestOpencodeReplyProjectionInputs(t *testing.T) {
	// JSON 为投影函数的明确参数，不声明存在模型回复或 HTTP 来源。
	text, model, provider, err := decodeOpencodeReply([]byte(`{"info":{"role":"assistant","modelID":"input-model","providerID":"input-provider"},"parts":[{"type":"text","text":"reply"}]}`))
	if err != nil || text != "reply" || model != "input-model" || provider != "input-provider" {
		t.Fatal("explicit reply projection input was not preserved", err)
	}
	if _, _, _, err := decodeOpencodeReply([]byte(`{"info":`)); err == nil {
		t.Fatal("malformed reply projection input was accepted")
	}
}

func TestOpencodeSessionDenyPermissionFailClosed(t *testing.T) {
	_, _, _, process := actualOpencodeProcess(t)
	for _, permission := range [][]opencodePermissionRule{
		{},
		{{Permission: "*", Pattern: "specific", Action: "deny"}},
		{{Permission: "*", Pattern: "*", Action: "deny"}, {Permission: "*", Pattern: "*", Action: "allow"}},
		{{Permission: "*", Pattern: "*", Action: "deny"}, {Permission: "read", Pattern: "*", Action: "allow"}},
	} {
		data, status, err := process.do(context.Background(), http.MethodPost, "/session", map[string]any{"title": "owned permission verification", "permission": permission})
		if err != nil || status != http.StatusOK {
			t.Fatal("actual native permission session create failed", err)
		}
		var created struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(data, &created); err != nil || created.ID == "" {
			t.Fatal("actual native permission session identity is missing", err)
		}
		var service *apierrors.ServiceError
		if !errors.As(process.verifySession(context.Background(), created.ID), &service) || service.Code != apierrors.ScopeDenied {
			t.Fatal("actual session without an effective last deny-all rule verified as safe")
		}
	}
}

func TestOpencodeRejectsUnauthenticatedAndUnknownReply(t *testing.T) {
	_, _, session, process := actualOpencodeProcess(t)
	wrong := &opencodeProcess{client: process.client, baseURL: process.baseURL, username: "caller-wrong", password: "caller-wrong"}
	if _, err := wrong.createSession(context.Background(), "owned rejected session"); err == nil {
		t.Fatal("actual unauthenticated session create accepted")
	}
	var service *apierrors.ServiceError
	if _, err := process.readMessages(context.Background(), session.NativeID+"-caller-unknown"); !errors.As(err, &service) || service.Code != apierrors.ProviderUnavailable {
		t.Fatal("actual unknown session context was not a provider rejection")
	}
}

func TestOpencodeWaitHealth(t *testing.T) {
	_, _, _, process := actualOpencodeProcess(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := waitHealth(ctx, process); err != nil {
		t.Fatal("actual native health wait failed", err)
	}
	wrong := &opencodeProcess{client: process.client, baseURL: process.baseURL, username: "caller-wrong", password: "caller-wrong", done: process.done}
	var service *apierrors.ServiceError
	if !errors.As(waitHealth(ctx, wrong), &service) || service.Code != apierrors.ProviderUnavailable || service.Message != "OpenCode health endpoint rejected the request" {
		t.Fatal("actual unauthorized health response was suppressed or retried as readiness")
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
				t.Fatal("opencode configured observation was invented")
			}
		}
	}
	if !found {
		t.Fatal("opencode adapter not registered")
	}
	if _, err := registry.Adapter("opencode"); err != nil {
		t.Fatal("opencode adapter lookup failed", err)
	}
}
