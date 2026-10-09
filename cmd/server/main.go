// Command server starts the Astrocyte API server.
//
// It assembles application services, opens the SQLite database,
// runs migrations and starts the HTTP server. On SIGINT/SIGTERM
// it performs a graceful shutdown.
//
// Environment variables:
//
//	ASTROCYTE_PORT      - HTTP listen port (default: 8787; invalid value fails startup)
//	ASTROCYTE_DATA_DIR  - Data directory for state.sqlite (default: system data dir)
//	ASTROCYTE_WEB_DIR   - Optional directory for static file serving (SPA fallback)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/foundation"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(logger); err != nil {
		logger.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	halfLife, weights, err := resolveAttentionPolicy()
	if err != nil {
		return err
	}
	importRoots, err := resolveImportRoots()
	if err != nil {
		return err
	}
	sourceReader, err := resolveSourceReader(importRoots)
	if err != nil {
		return err
	}
	origins, err := resolveAllowedOrigins()
	if err != nil {
		return err
	}
	concurrency, err := configuredPositiveInt("ASTROCYTE_JOB_CONCURRENCY", 2, 32)
	if err != nil {
		return err
	}
	maxAttempts, err := configuredPositiveInt("ASTROCYTE_JOB_MAX_ATTEMPTS", 3, 100)
	if err != nil {
		return err
	}
	jobSeconds, err := configuredPositiveInt("ASTROCYTE_JOB_TIMEOUT_SECONDS", 300, 86400)
	if err != nil {
		return err
	}
	// Resolve data directory.
	dataDir, err := resolveDataDir()
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir %q: %w", dataDir, err)
	}
	logger.Info("data directory", "path", dataDir)

	// Open database.
	dbPath := filepath.Join(dataDir, "state.sqlite")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	logger.Info("database opened", "path", dbPath)

	// Run embedded migrations.
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		return fmt.Errorf("run migrations (startup aborted): %w", err)
	}
	logger.Info("migrations applied")

	// Read schema version for foundation service.
	schemaVersion, err := db.SchemaVersion(context.Background())
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	objectStore, err := objects.New(filepath.Join(dataDir, "objects"))
	if err != nil {
		return fmt.Errorf("open objects: %w", err)
	}
	attention := attentionapp.NewAttentionService(sqlite.NewAttentionRepository(db), sourceReader, objectStore, attentionapp.ServiceOptions{WorkerConcurrency: concurrency, MaxAttempts: maxAttempts, JobTimeout: time.Duration(jobSeconds) * time.Second, AttentionHalfLife: halfLife, AttentionWeights: weights})
	// Assemble application services; Workspace and Swarm retain their S0 boundary.
	services := httpapi.Services{
		Attention:     attention,
		Foundation:    foundation.NewAttentionService(schemaVersion),
		Materials:     attention,
		Opportunities: attention,
		Projects:      workspaceapp.NewProjectService(),
		Proposals:     workspaceapp.NewProposalService(),
		Sessions:      workspaceapp.NewSessionService(),
		Missions:      swarmapp.NewMissionService(),
	}
	if automatic, ok := any(attention).(attentionapp.AutomaticDistillationService); ok {
		services.Automatic = automatic
	}
	if profile, ok := any(attention).(attentionapp.RankingProfileService); ok {
		services.RankingProfile = profile
	}

	// Resolve port — invalid value fails startup.
	port, err := resolvePort()
	if err != nil {
		return fmt.Errorf("invalid ASTROCYTE_PORT: %w", err)
	}
	logger.Info("configuring server", "port", port)

	// Resolve optional web dir for static serving.
	webDir := os.Getenv("ASTROCYTE_WEB_DIR")
	if webDir != "" {
		info, err := os.Stat(webDir)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("ASTROCYTE_WEB_DIR %q is not a valid directory", webDir)
		}
		logger.Info("static file serving enabled", "dir", webDir)
	}

	// Create and start HTTP server.
	srv := httpapi.NewServer(httpapi.Config{
		AgentToken:     os.Getenv("ASTROCYTE_AGENT_TOKEN"),
		AllowedOrigins: origins,
		Port:           port,
		Logger:         logger,
		Services:       services,
		WebDir:         webDir,
	})

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	workerErrors := make(chan error, 1)
	go func() { workerErrors <- attention.Run(workerCtx) }()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	var runError error
	workersStopped := false
	select {
	case err := <-errCh:
		if err != nil && err.Error() != "http: Server closed" {
			runError = fmt.Errorf("http server: %w", err)
		}
	case err := <-workerErrors:
		workersStopped = true
		if err != nil {
			runError = fmt.Errorf("attention workers: %w", err)
		} else {
			runError = fmt.Errorf("attention workers stopped unexpectedly")
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownErr := srv.Shutdown(shutdownCtx)
	cancelWorkers()
	if !workersStopped {
		// Adapters honor context; wait until checkpoints are persisted before closing DB.
		if err := <-workerErrors; err != nil && runError == nil {
			runError = fmt.Errorf("stop attention workers: %w", err)
		}
	}
	if shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}

	logger.Info("server stopped")
	return runError
}

func resolveDataDir() (string, error) {
	if dir := os.Getenv("ASTROCYTE_DATA_DIR"); dir != "" {
		return dir, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, "astrocyte"), nil
}

func resolvePort() (int, error) {
	p := os.Getenv("ASTROCYTE_PORT")
	if p == "" {
		return 8787, nil
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return 0, fmt.Errorf("not a number: %w", err)
	}
	if n <= 0 || n >= 65536 {
		return 0, fmt.Errorf("port %d out of range (1-65535)", n)
	}
	return n, nil
}
