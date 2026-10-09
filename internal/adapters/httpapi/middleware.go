package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
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
		start := w
		sw := &statusWriter{ResponseWriter: start, status: http.StatusOK}
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

// loopbackMiddleware rejects requests whose Host or Origin do not
// resolve to a loopback address. This protects the local-only API
// from non-local access.
func loopbackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Host header.
		host := r.Host
		if host == "" {
			host = r.RemoteAddr
		}
		if !isLoopbackHost(host) {
			http.Error(w, `{"code":"scope_denied","message":"non-loopback host rejected","retryable":false}`, http.StatusForbidden)
			return
		}

		// Check Origin header if present.
		if origin := r.Header.Get("Origin"); origin != "" {
			// Parse origin host from the full URL.
			originHost := origin
			if idx := findSchemeEnd(originHost); idx >= 0 {
				originHost = originHost[idx:]
			}
			if idx := findPathStart(originHost); idx >= 0 {
				originHost = originHost[:idx]
			}
			if !isLoopbackHost(originHost) {
				http.Error(w, `{"code":"scope_denied","message":"non-loopback origin rejected","retryable":false}`, http.StatusForbidden)
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
