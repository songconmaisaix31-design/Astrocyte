package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// generateRequestID creates a random hex request ID.
func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// requestIDMiddleware adds a request ID to every request context.
// Must run before loopbackMiddleware so rejected requests get an ID.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware logs every request with method, path, status and duration.
func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"request_id", r.Context().Value(requestIDKey),
		)
	})
}

// statusWriter captures the HTTP status code for logging.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// writeRejectionError writes an ErrorV1 JSON response for middleware rejections.
func writeRejectionError(w http.ResponseWriter, r *http.Request, status int, message string) {
	rid := ""
	if v, ok := r.Context().Value(requestIDKey).(string); ok {
		rid = v
	}
	if rid == "" {
		rid = generateRequestID()
		w.Header().Set("X-Request-ID", rid)
	}

	env := apierrors.ErrorEnvelope{
		SchemaVersion: apierrors.SchemaVersion,
		Error: &apierrors.ServiceError{
			Code:           apierrors.ScopeDenied,
			Message:        message,
			Retryable:      false,
			RequestID:      rid,
			RequiredAction: "establish_local_session_and_use_allowed_origin",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(env) //nolint:errcheck
}

// loopbackMiddleware rejects requests whose Host or Origin do not
// resolve to a loopback address. Returns ErrorV1 JSON with
// nonempty request_id, retryable, and required_action.
func loopbackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Host header.
		host := r.Host
		if host == "" {
			host = r.RemoteAddr
		}
		if !isLoopbackHost(host) {
			writeRejectionError(w, r, http.StatusForbidden, "non-loopback host rejected")
			return
		}

		// Check Origin header if present.
		if origin := r.Header.Get("Origin"); origin != "" {
			originHost := origin
			if idx := findSchemeEnd(originHost); idx >= 0 {
				originHost = originHost[idx:]
			}
			if idx := findPathStart(originHost); idx >= 0 {
				originHost = originHost[:idx]
			}
			if !isLoopbackHost(originHost) {
				writeRejectionError(w, r, http.StatusForbidden, "non-loopback origin rejected")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// findSchemeEnd returns the index after "://" in a URL, or -1.
func findSchemeEnd(s string) int {
	for i := 0; i < len(s)-2; i++ {
		if s[i] == ':' && s[i+1] == '/' && s[i+2] == '/' {
			return i + 3
		}
	}
	return -1
}

// findPathStart returns the index of the first '/' after the host, or -1.
func findPathStart(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
