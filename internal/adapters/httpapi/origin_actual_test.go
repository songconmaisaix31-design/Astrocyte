package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
)

func TestLoopbackActualHTTPRejectsMalformedOrigins(t *testing.T) {
	api := NewServer(Config{
		Services:       Services{Foundation: foundation.NewService(1)},
		AllowedOrigins: []string{"http://localhost:8787", "https://[::1]:8787"},
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	server := httptest.NewServer(api.Handler())
	defer server.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	for _, input := range []struct {
		origin string
		status int
	}{
		{"", http.StatusOK},
		{server.URL, http.StatusOK},
		{"http://localhost:8787", http.StatusOK},
		{"https://[::1]:8787", http.StatusOK},
		{"http://outside.invalid", http.StatusForbidden},
		{"http://127.0.0.1:8787/path", http.StatusForbidden},
		{"http://127.0.0.1:8787/", http.StatusForbidden},
		{"http://127.0.0.1:8787?", http.StatusForbidden},
		{"http://127.0.0.1:8787#", http.StatusForbidden},
		{"http://127.0.0.1:8787?query=value", http.StatusForbidden},
		{"http://127.0.0.1:8787#fragment", http.StatusForbidden},
		{"http://caller:input@127.0.0.1:8787", http.StatusForbidden},
		{"ftp://127.0.0.1:8787", http.StatusForbidden},
		{"127.0.0.1:8787", http.StatusForbidden},
		{"http://[::1", http.StatusForbidden},
		{"null", http.StatusForbidden},
	} {
		t.Run(input.origin, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/health", nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Origin", input.origin)
			response, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Error(err)
				}
			}()
			if response.StatusCode != input.status || response.Header.Get("X-Request-ID") == "" {
				t.Fatal("actual HTTP origin status or request identity changed", response.StatusCode, input.status)
			}
			if input.status == http.StatusForbidden {
				var result apierrors.ErrorEnvelope
				if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				if result.SchemaVersion != 1 || result.Error == nil || result.Error.Code != apierrors.ScopeDenied || result.Error.RequestID != response.Header.Get("X-Request-ID") || result.Error.RequiredAction == "" || result.Error.Retryable {
					t.Fatal("actual origin rejection lost its error envelope", result)
				}
				return
			}
			var health struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(response.Body).Decode(&health); err != nil || health.Status != "ok" {
				t.Fatal("actual foundation health failed", health.Status, err)
			}
		})
	}
}
