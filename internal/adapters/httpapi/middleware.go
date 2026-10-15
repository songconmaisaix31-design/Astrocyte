package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// 生成独立的随机十六进制请求 ID。
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// 请求 ID 在回环检查之前生成，拒绝响应也保留同一 ID。
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

// 记录请求方法、路径、状态和请求 ID。
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

// 保留实际发送的 HTTP 状态供日志使用。
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// 拒绝响应复用现有 JSON 写入和错误处理。
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

	writeJSON(w, status, env)
}

// Host 和完整 Origin 仅接受明确的回环地址，拒绝结果使用 ErrorV1。
func loopbackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HTTP Host 与没有 Host 的连接来源分别核实。
		host := r.Host
		if host == "" {
			host = r.RemoteAddr
		}
		if !isLoopbackHost(host) {
			writeRejectionError(w, r, http.StatusForbidden, "non-loopback host rejected")
			return
		}

		// 标准 URL 解析拒绝凭据、路径和非 origin 字段。
		if origin := r.Header.Get("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || origin != parsed.Scheme+"://"+parsed.Host || !isLoopbackHost(parsed.Host) {
				writeRejectionError(w, r, http.StatusForbidden, "non-loopback origin rejected")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
