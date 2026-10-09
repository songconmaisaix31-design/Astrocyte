package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type cachedInventory struct {
	calls int
	err   error
}

func (s *cachedInventory) ListLocalAgents(_ context.Context) (apierrors.ListResult, error) {
	s.calls++
	if s.err != nil {
		return apierrors.ListResult{}, s.err
	}
	result := apierrors.EmptyList()
	result.Items = append(result.Items, domain.LocalAgent{
		ID: "contract-local-cli", DisplayName: "Contract local CLI",
		Installed:    domain.Observation{Status: "available", Reason: "path_entry_found"},
		Configured:   domain.Observation{Status: "unknown", Reason: "configuration_not_inspected"},
		Startable:    domain.Observation{Status: "unknown", Reason: "native_runtime_not_tested"},
		Capabilities: domain.UnknownNativeCapabilities(),
	})
	return result, nil
}

func TestLocalAgentInventoryHumanReadBoundary(t *testing.T) {
	inventory := &cachedInventory{}
	// Inventory alone must enable the existing session guard, even when the
	// Attention service is absent. Discovery never creates project authority.
	srv := NewServer(Config{
		Services: Services{LocalAgents: inventory}, AgentToken: "inventory-agent",
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	cookie, _ := bootstrapSession(t, srv)
	for _, denied := range []struct {
		name, bearer, origin string
		useCookie            bool
	}{
		{name: "no session"},
		{name: "Agent bearer", bearer: "Bearer inventory-agent", useCookie: true},
		{name: "invalid bearer", bearer: "Bearer invalid", useCookie: true},
		{name: "foreign origin", origin: "https://example.invalid", useCookie: true},
	} {
		t.Run(denied.name, func(t *testing.T) {
			var session *http.Cookie
			if denied.useCookie {
				session = cookie
			}
			response := attentionRequest(srv, "GET", "/api/v1/local-agents", "", session, "", denied.bearer, denied.origin)
			if response.Code != http.StatusForbidden || inventory.calls != 0 {
				t.Fatalf("denied read invoked inventory: %d calls=%d %s", response.Code, inventory.calls, response.Body.String())
			}
		})
	}
	for range 2 {
		response := attentionRequest(srv, "GET", "/api/v1/local-agents", "", cookie, "", "", "")
		if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("human read: %d %s", response.Code, response.Body.String())
		}
		body := readJSON(t, response.Result())
		item := body["items"].([]any)[0].(map[string]any)
		if item["configured"].(map[string]any)["status"] != "unknown" || item["version"] != nil {
			t.Fatalf("installation incorrectly promoted to readiness: %v", item)
		}
		capabilities := item["capabilities"].(map[string]any)
		if len(capabilities) != 8 {
			t.Fatalf("missing native observations: %v", capabilities)
		}
		for name, value := range capabilities {
			if value.(map[string]any)["status"] != "unknown" {
				t.Fatalf("%s incorrectly promoted to native support", name)
			}
		}
	}
	// Native control remains unassembled after successful discovery.
	response := attentionRequest(srv, "POST", "/api/v1/sessions/test/resume", "", cookie, "", "", "")
	if response.Code != http.StatusForbidden { // No CSRF, so even the future route cannot run.
		t.Fatalf("control bypass: %d %s", response.Code, response.Body.String())
	}
}

func TestLocalAgentInventoryMissingAndFailedAreNotEmptySuccess(t *testing.T) {
	missing := doRequest(newTestServer(), "GET", "/api/v1/local-agents")
	if missing.StatusCode != http.StatusNotImplemented {
		t.Fatalf("missing service returned %d", missing.StatusCode)
	}
	if body := readErrorBody(t, missing); body["code"] != "unsupported_capability" {
		t.Fatalf("missing inventory: %v", body)
	}
	inventory := &cachedInventory{err: errors.New("private-path-must-not-leak")}
	srv := NewServer(Config{Services: Services{LocalAgents: inventory}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	cookie, _ := bootstrapSession(t, srv)
	failed := attentionRequest(srv, "GET", "/api/v1/local-agents", "", cookie, "", "", "")
	if failed.Code != http.StatusInternalServerError || strings.Contains(failed.Body.String(), "private-path") {
		t.Fatalf("failed query swallowed or leaked: %d %s", failed.Code, failed.Body.String())
	}
	if body := readErrorBody(t, failed.Result()); body["code"] != "internal_error" || body["request_id"] == "" {
		t.Fatalf("missing structured error: %v", body)
	}
}
