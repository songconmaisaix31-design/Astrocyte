package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestLocalProjectActualHTTPPermissionsTokensAndRestartWithoutNative(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	directory := acceptedExperienceDirectory(t)
	path := filepath.Join(directory, "state.sqlite")
	db := openAttentionDB(t, path)
	nativeRoot := filepath.Join(directory, "unused-native")
	if err := os.Mkdir(nativeRoot, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	var service interface {
		workspace.LocalProjects
		workspace.ScopedAgentTokens
	}
	var registry *agents.Registry
	var baseline []domain.LocalAgent
	var origin, csrf string
	var client *http.Client
	var currentHuman domain.Caller
	start := func() func() {
		registry = agents.NewRegistry(nativeRoot)
		var err error
		baseline, err = registry.SnapshotNative(ctx)
		if err != nil {
			t.Fatal(err)
		}
		service = workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry).WithNativeMessageObjects(store)
		server := httpapi.NewServer(httpapi.Config{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			Services: httpapi.Services{
				LocalProjects: service,
				Attention:     attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: attentionSpaces(db)}),
				ProjectSpaces: workspace.NewProjectSpaceService(NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{}),
			},
		})
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		origin = "http://" + listener.Addr().String()
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		stopped := false
		stop := func() {
			if stopped {
				return
			}
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdown); err != nil {
				t.Fatal(err)
			}
			if err := <-done; !errors.Is(err, http.ErrServerClosed) {
				t.Fatal("actual permission server did not shut down normally", err)
			}
			stopped = true
		}
		t.Cleanup(stop)
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client = &http.Client{Jar: jar, Timeout: 5 * time.Second}
		t.Cleanup(client.CloseIdleConnections)
		response, err := client.Get(origin + "/api/v1/auth/session")
		if err != nil {
			t.Fatal(err)
		}
		var session struct {
			ID   string `json:"actor_id"`
			Kind string `json:"actor_kind"`
			CSRF string `json:"csrf_token"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&session)
		if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK || session.Kind != "human" || session.ID == "" || session.CSRF == "" {
			t.Fatal("real authenticated human session is required", err)
		}
		csrf = session.CSRF
		currentHuman = domain.Caller{ID: session.ID, Kind: session.Kind}
		return stop
	}
	stop := start()
	request := func(method, route, token, key string, command map[string]any, expected int, output any) []byte {
		var body []byte
		if command != nil {
			command["schema_version"], command["request_id"], command["expected_version"] = 1, uuid.NewString(), 1
			var err error
			body, err = json.Marshal(command)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, origin+route, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if command != nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", csrf)
			if key == "" {
				key = uuid.NewString()
			}
			req.Header.Set("Idempotency-Key", key)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil || response.StatusCode != expected {
			t.Fatal("actual permission HTTP result differs", method, route, response.StatusCode, expected, err)
		}
		if err := json.Unmarshal(data, output); err != nil {
			t.Fatal(err)
		}
		return data
	}
	assertError := func(actual apierrors.ErrorEnvelope, code apierrors.Code) {
		if actual.SchemaVersion != 1 || actual.Error == nil || actual.Error.Code != code || actual.Error.RequestID == "" || actual.Error.RequiredAction == "" {
			t.Fatal("real permission rejection lacks its exact ErrorV1 fields")
		}
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	register := func(name string) domain.LocalProject {
		var space attention.ProjectSpaceResult
		request(http.MethodPost, "/api/v1/project-spaces", "", "", map[string]any{"title": name}, http.StatusCreated, &space)
		var result struct {
			Project domain.LocalProject `json:"project"`
		}
		request(http.MethodPost, "/api/v1/local-projects", "", "", map[string]any{"name": name, "root": root, "space_id": space.Space.ID}, http.StatusOK, &result)
		if result.Project.ID == "" || result.Project.Root != root || result.Project.SpaceID != space.Space.ID {
			t.Fatal("actual registered project does not retain its human-selected public checkout")
		}
		return result.Project
	}
	project := register("Actual public file permission verification")
	other := register("Separate actual project permission scope")
	if project.Settings.AllowDirectory || project.Settings.ExpandReferences || project.Settings.ExternalModelCLI != "" || project.Settings.AllowAgentControl {
		t.Fatal("actual project registration expanded default permissions")
	}
	contextPath := "/api/v1/local-projects/" + project.ID + "/context"
	var rejection apierrors.ErrorEnvelope
	request(http.MethodPost, contextPath, "", "", map[string]any{"files": []string{"README.md"}}, http.StatusForbidden, &rejection)
	assertError(rejection, apierrors.ScopeDenied)
	request(http.MethodPost, "/api/v1/local-projects/"+project.ID+"/sessions", "", "", map[string]any{"cli": "codex"}, http.StatusForbidden, &rejection)
	assertError(rejection, apierrors.ApprovalRequired)
	settings := project.Settings
	settings.AllowDirectory, settings.AllowedSubdirs = true, []string{"."}
	settingsPath := "/api/v1/local-projects/" + project.ID + "/settings"
	var configured struct {
		Project domain.LocalProject `json:"project"`
	}
	request(http.MethodPut, settingsPath, "", "", map[string]any{"expected_revision": 1, "settings": settings}, http.StatusOK, &configured)
	project = configured.Project
	if project.Settings.Revision != 2 || !project.Settings.AllowDirectory || project.Settings.ExternalModelCLI != "" || project.Settings.AllowAgentControl {
		t.Fatal("actual settings changed unrelated model or control permissions")
	}
	request(http.MethodPut, settingsPath, "", "", map[string]any{"expected_revision": 1, "settings": settings}, http.StatusConflict, &rejection)
	assertError(rejection, apierrors.VersionConflict)
	var granted struct {
		Grant domain.ProjectGrant `json:"grant"`
	}
	agentID := uuid.NewString()
	grantPath := "/api/v1/local-projects/" + project.ID + "/grants"
	grantJSON := request(http.MethodPost, grantPath, "", "", map[string]any{"agent_id": agentID, "actions": []string{"read_context", "discover"}}, http.StatusOK, &granted)
	if granted.Grant.AgentID != agentID || granted.Grant.ProjectID != project.ID || strings.Contains(string(grantJSON), "private_token_digest") || strings.Contains(string(grantJSON), "TokenDigest") {
		t.Fatal("actual grant identity or private digest exposure differs")
	}
	var issued struct {
		Credential domain.AgentToken `json:"credential"`
	}
	request(http.MethodPost, "/api/v1/local-projects/"+project.ID+"/agent-token", "", "", map[string]any{"agent_id": agentID}, http.StatusOK, &issued)
	token := issued.Credential.Token
	caller, err := service.AuthenticateAgentToken(ctx, token)
	if err != nil || token == "" || caller.ID != agentID || caller.Kind != "agent" || caller.ProjectID != project.ID || !issued.Credential.ExpiresAt.After(time.Now()) {
		t.Fatal("actual scoped Agent token does not resolve to its persisted human grant", err)
	}
	assertContext := func() {
		var packet domain.ContextPacket
		request(http.MethodPost, contextPath, token, "", map[string]any{"files": []string{"README.md"}}, http.StatusOK, &packet)
		if packet.ProjectID != project.ID || packet.SettingsRevision != project.Settings.Revision || len(packet.Materials) != 0 || len(packet.Files) != 1 || packet.Files[0].Path != "README.md" || packet.Files[0].Text != string(original) {
			t.Fatal("actual scoped HTTP context differs from the selected public README")
		}
	}
	assertContext()
	for _, forbidden := range []struct {
		method string
		path   string
		body   map[string]any
	}{
		{http.MethodPost, "/api/v1/local-projects/" + other.ID + "/context", map[string]any{"files": []string{"README.md"}}},
		{http.MethodPut, settingsPath, map[string]any{"expected_revision": 2, "settings": settings}},
		{http.MethodPost, grantPath, map[string]any{"agent_id": agentID, "actions": []string{"read_context"}}},
		{http.MethodPost, "/api/v1/local-projects/" + project.ID + "/agent-token", map[string]any{"agent_id": agentID}},
	} {
		request(forbidden.method, forbidden.path, token, "", forbidden.body, http.StatusForbidden, &rejection)
		assertError(rejection, apierrors.ScopeDenied)
	}
	// 不存在的会话参数只验证真实消息依赖守卫，不声明有 unknown 原生回执。
	missingSession, operation := uuid.NewString(), uuid.NewString()
	for range 2 {
		request(http.MethodPost, "/api/v1/local-projects/"+project.ID+"/sessions/"+missingSession+"/send", "", operation, map[string]any{"message": string(original)}, http.StatusNotImplemented, &rejection)
		assertError(rejection, apierrors.UnsupportedCapability)
	}
	assertUnused := func() {
		observations, err := registry.SnapshotNative(ctx)
		if err != nil || !reflect.DeepEqual(observations, baseline) {
			t.Fatal("permission checks changed actual unprobed Registry observations", err)
		}
		for _, name := range []string{nativeRoot, filepath.Join(directory, "objects")} {
			entries, err := os.ReadDir(name)
			if err != nil || len(entries) != 0 {
				t.Fatal("permission refusal started a controller or published native input", err)
			}
		}
		var count int
		if err := db.conn.QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&count); err != nil || count != 0 {
			t.Fatal("permission refusal created a native session", err)
		}
	}
	assertUnused()
	stop()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	stop = start()
	saved, err := db.LoadProject(ctx, project.ID)
	if err != nil || !reflect.DeepEqual(saved, project) {
		t.Fatal("actual project settings did not survive same-file SQLite reopen", err)
	}
	authenticated, err := service.AuthenticateAgentToken(ctx, token)
	if err != nil || !reflect.DeepEqual(authenticated, caller) {
		t.Fatal("actual scoped token did not survive same-file SQLite reopen", err)
	}
	assertContext()
	request(http.MethodPost, grantPath+"/revoke", "", "", map[string]any{"agent_id": agentID}, http.StatusOK, &granted)
	if granted.Grant.RevokedAt == nil || granted.Grant.TokenDigest != "" {
		t.Fatal("actual human grant revocation was not persisted")
	}
	if _, err := service.AuthenticateAgentToken(ctx, token); err == nil {
		t.Fatal("revoked actual token still authenticates")
	} else {
		acceptedArtifactExpectCode(t, err, apierrors.ApprovalRevoked)
	}
	request(http.MethodPost, contextPath, token, "", map[string]any{"files": []string{"README.md"}}, http.StatusForbidden, &rejection)
	assertError(rejection, apierrors.ScopeDenied)
	if _, err := service.ReadProjectContext(ctx, caller, project.ID, domain.ContextRequest{Files: []string{"README.md"}}); err == nil {
		t.Fatal("revoked actual Agent identity still reads the public project context")
	} else {
		acceptedArtifactExpectCode(t, err, apierrors.ApprovalRevoked)
	}
	if _, err := service.SetProjectSettings(ctx, currentHuman, project.ID, domain.SettingsCommand{ExpectedRevision: 1, Settings: settings}); err == nil {
		t.Fatal("stale project settings became writable after HTTP and SQLite restart")
	} else {
		acceptedArtifactExpectCode(t, err, apierrors.VersionConflict)
	}
	assertUnused()
	stop()
	t.Log("real HTTP identity, settings, scoped grant/token, cross-project refusal, same-file restart, revoked reads and missing-message-dependency refusal verified; original unknown delivery/stop/resume acceptance NOT_RUN")
}
