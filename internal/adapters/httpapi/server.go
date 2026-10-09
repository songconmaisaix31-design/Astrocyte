// Package httpapi provides the HTTP transport adapter.
//
// It maps HTTP routes to application service calls, handles request
// parsing and response serialization, and enforces loopback-only
// access. The HTTP layer never imports database/sql or executes
// queries directly.
package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

// Server is the HTTP server that exposes the Astrocyte API.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// Config holds server configuration.
type Config struct {
	AgentToken     string
	AllowedOrigins []string
	Port           int
	Logger         *slog.Logger
	Services       Services
	WebDir         string // Optional: directory for static file serving
}

// Services bundles all application services the HTTP layer depends on.
type Services struct {
	Attention     attentionapp.AttentionService
	Foundation    foundation.Service
	Materials     attentionapp.MaterialService
	Opportunities attentionapp.OpportunityService
	Projects      workspaceapp.ProjectService
	Proposals     workspaceapp.ProposalService
	Sessions      workspaceapp.SessionService
	Missions      swarmapp.MissionService
}

// NewServer creates a new HTTP server with the given configuration.
func NewServer(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Port == 0 {
		cfg.Port = 8787
	}

	mux := http.NewServeMux()
	h := &handler{
		services: cfg.Services,
		logger:   cfg.Logger,
	}

	// Health and foundation
	mux.HandleFunc("GET /api/v1/health", h.handleHealth)
	mux.HandleFunc("GET /api/v1/foundation", h.handleFoundation)

	// List endpoints (empty reads in S0)
	mux.HandleFunc("GET /api/v1/materials", h.handleListMaterials)
	mux.HandleFunc("GET /api/v1/opportunities", h.handleListOpportunities)
	mux.HandleFunc("GET /api/v1/projects", h.handleListProjects)
	mux.HandleFunc("GET /api/v1/proposals", h.handleListProposals)
	mux.HandleFunc("GET /api/v1/sessions", h.handleListSessions)
	mux.HandleFunc("GET /api/v1/missions", h.handleListMissions)

	// Single-resource reads
	mux.HandleFunc("GET /api/v1/missions/{id}", h.handleGetMission)

	// Future write endpoints return 501 unsupported_capability
	if cfg.Services.Attention == nil {
		mux.HandleFunc("POST /api/v1/materials/imports", h.notImplemented("material_import"))
		mux.HandleFunc("POST /api/v1/opportunities/{id}/reviews", h.notImplemented("opportunity_review"))
	} else {
		h.registerAttention(mux)
	}
	mux.HandleFunc("POST /api/v1/opportunities/{id}/admissions", h.notImplemented("opportunity_admission"))
	mux.HandleFunc("POST /api/v1/projects", h.notImplemented("project_create"))
	mux.HandleFunc("POST /api/v1/sessions/{id}/resume", h.notImplemented("session_resume"))
	mux.HandleFunc("POST /api/v1/sessions/{id}/handoff", h.notImplemented("session_handoff"))
	mux.HandleFunc("POST /api/v1/proposals/{id}/submit", h.notImplemented("proposal_submit"))
	mux.HandleFunc("POST /api/v1/proposals/{id}/approvals", h.notImplemented("proposal_approve"))
	mux.HandleFunc("POST /api/v1/approvals/{id}/revoke", h.notImplemented("approval_revoke"))
	mux.HandleFunc("POST /api/v1/missions/{id}/pause", h.notImplemented("mission_pause"))
	mux.HandleFunc("POST /api/v1/missions/{id}/cancel", h.notImplemented("mission_cancel"))
	mux.HandleFunc("POST /api/v1/work-items/{id}/claims", h.notImplemented("workitem_claim"))
	mux.HandleFunc("POST /api/v1/work-items/{id}/artifacts", h.notImplemented("workitem_submit"))
	mux.HandleFunc("POST /api/v1/artifacts/{id}/acceptance", h.notImplemented("artifact_accept"))
	mux.HandleFunc("GET /api/v1/events", h.notImplemented("events_sse"))

	// Catch-all for unknown /api/v1/* routes — always JSON 404
	mux.HandleFunc("/api/v1/", h.handleAPINotFound)
	// Bare /api and /api/v1 (no trailing slash) → JSON 404
	mux.HandleFunc("/api/v1", h.handleAPINotFound)
	mux.HandleFunc("/api", h.handleAPINotFound)

	// Build handler stack with optional static file serving
	var innerHandler http.Handler = mux

	if cfg.WebDir != "" {
		// Wrap with SPA static file serving for non-/api routes
		innerHandler = spaHandler(cfg.WebDir, mux)
	}
	if cfg.Services.Attention != nil {
		guard := newSessionGuard(cfg)
		mux.HandleFunc("GET /api/v1/auth/session", guard.bootstrap)
		innerHandler = guard.middleware(innerHandler)
	}

	// Apply middleware stack: request ID first so all responses have it
	stack := requestIDMiddleware(
		loopbackMiddleware(
			loggingMiddleware(cfg.Logger, innerHandler),
		),
	)

	return &Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("127.0.0.1:%d", cfg.Port),
			Handler:           stack,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		logger: cfg.Logger,
	}
}

// ListenAndServe starts the HTTP server. It blocks until the server
// is shut down or encounters a fatal error. Returns http.ErrServerClosed
// on graceful shutdown.
func (s *Server) ListenAndServe() error {
	s.logger.Info("HTTP server starting", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the configured listen address (useful in tests).
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// spaHandler serves static files from webDir with SPA fallback.
// - /api/* routes always go to the API mux (never static)
// - Existing files in webDir are served directly
// - Non-existent files get index.html (SPA fallback)
// - No directory listing
func spaHandler(webDir string, apiHandler http.Handler) http.Handler {
	fileServer := http.FileServer(http.Dir(webDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// API routes always go to API handler (including bare /api)
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			apiHandler.ServeHTTP(w, r)
			return
		}

		// Try to serve the file directly
		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// Check if file exists (no directory listing)
		fullPath := webDir + "/" + filePath
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			// SPA fallback: serve index.html for non-existent paths
			http.ServeFile(w, r, webDir+"/index.html")
			return
		}

		// Serve the actual file
		fileServer.ServeHTTP(w, r)
	})
}

// isLoopbackHost checks whether the request host resolves to a loopback address.
func isLoopbackHost(host string) bool {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		h = host
	}

	switch h {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	}

	// Host names that merely resolve to loopback are not trusted (DNS rebinding).
	return false
}
