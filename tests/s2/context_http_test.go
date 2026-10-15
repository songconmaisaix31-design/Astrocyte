package s2_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func verifyActualPacketHTTP(t *testing.T, service *workspace.S2Service, projects workspace.LocalProjects, packet domain.ImmutableContextPacket, grant domain.ApprovalGrant, memberToken, otherToken string) {
	t.Helper()
	server := httptest.NewUnstartedServer(nil)
	baseURL := "http://" + server.Listener.Addr().String()
	port := server.Listener.Addr().(*net.TCPAddr).Port
	production := httpapi.NewServer(httpapi.Config{Port: port, AllowedOrigins: []string{baseURL}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Services: httpapi.Services{WorkspaceS2: service, LocalProjects: projects}})
	server.Config.Handler = production.Handler()
	server.Start()
	defer server.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	humanClient := &http.Client{Jar: jar, Timeout: 10 * time.Second}
	agentClient := &http.Client{Timeout: 10 * time.Second}
	response, err := humanClient.Get(baseURL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err = json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if err = response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || session.CSRFToken == "" {
		t.Fatal("actual human bootstrap did not authenticate")
	}
	request := func(method, path, token string, body any, want int) []byte {
		t.Helper()
		var data []byte
		if body != nil {
			var err error
			data, err = json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequest(method, baseURL+"/api/v1"+path, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		client := humanClient
		if token != "" {
			client = agentClient
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if method != http.MethodGet {
			req.Header.Set("Origin", baseURL)
			req.Header.Set("X-CSRF-Token", session.CSRFToken)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", uuid.NewString())
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		result, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err = res.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Fatalf("actual %s %s status=%d expected=%d", method, path, res.StatusCode, want)
		}
		return result
	}
	assertSafe := func(body []byte) {
		t.Helper()
		var envelope struct {
			Packet httpapi.WorkspaceContextPacketDTO `json:"context_packet"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Packet.Snapshot.WorktreeID != packet.Snapshot.WorktreeID || envelope.Packet.Goal.Revision != packet.Goal.Revision {
			t.Fatal("HTTP context omitted fixed actual references")
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw["context_packet"], &fields); err != nil {
			t.Fatal(err)
		}
		var snapshot map[string]json.RawMessage
		if err := json.Unmarshal(fields["snapshot"], &snapshot); err != nil {
			t.Fatal(err)
		}
		if _, exists := snapshot["root"]; exists {
			t.Fatal("HTTP snapshot exposed an absolute local root")
		}
		encodedRoot, err := json.Marshal(packet.Snapshot.Root)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(body, encodedRoot[1:len(encodedRoot)-1]) {
			t.Fatal("HTTP context exposed the absolute project root")
		}
		if len(envelope.Packet.Scope.Roots) != len(packet.Scope.Roots) {
			t.Fatal("HTTP context lost approved roots")
		}
		for index, root := range packet.Scope.Roots {
			relative, err := filepath.Rel(packet.Snapshot.Root, root)
			if err != nil {
				t.Fatal(err)
			}
			if envelope.Packet.Scope.Roots[index] != filepath.ToSlash(relative) {
				t.Fatal("HTTP root differs from the fixed authorized relative root")
			}
		}
	}
	assertSafe(request(http.MethodGet, "/context-packets/"+packet.ID, "", nil, http.StatusOK))
	assertSafe(request(http.MethodGet, "/context-packets/"+packet.ID, memberToken, nil, http.StatusOK))
	request(http.MethodGet, "/context-packets/"+packet.ID, otherToken, nil, http.StatusForbidden)
	command := workspace.BuildContextPacketCommand{WorkspaceCommandMeta: workspaceMeta(0), GrantID: grant.ID, GrantEpoch: grant.Epoch, PredecessorID: packet.ID, MaxBytes: 128 * 1024}
	assertSafe(request(http.MethodPost, "/context-packets", "", command, http.StatusOK))
}
