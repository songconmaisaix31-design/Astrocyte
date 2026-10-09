package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// The source scope is startup-owned and has no implicit public/private defaults.
// An ordinary test or dev invocation never starts model processing unless opted in.
func resolveDistiller(ctx context.Context, dataDir string) (attentionapp.Distiller, []string, error) {
	raw := os.Getenv("ASTROCYTE_ENABLE_CODEX_DISTILLATION")
	if raw == "" || raw == "false" {
		return nil, nil, nil
	}
	if raw != "true" {
		return nil, nil, fmt.Errorf("ASTROCYTE_ENABLE_CODEX_DISTILLATION must be true or false")
	}
	var sourceKeys []string
	if err := json.Unmarshal([]byte(os.Getenv("ASTROCYTE_PROCESSING_SOURCE_KEYS")), &sourceKeys); err != nil || len(sourceKeys) == 0 {
		return nil, nil, fmt.Errorf("ASTROCYTE_PROCESSING_SOURCE_KEYS must be a nonempty JSON array of explicitly authorized source keys")
	}
	seen := make(map[string]bool)
	for _, key := range sourceKeys {
		if key == "" || key != strings.TrimSpace(key) || seen[key] {
			return nil, nil, fmt.Errorf("ASTROCYTE_PROCESSING_SOURCE_KEYS entries must be distinct nonempty canonical source keys")
		}
		seen[key] = true
	}
	timeoutSeconds, err := configuredPositiveInt("ASTROCYTE_CODEX_TIMEOUT_SECONDS", 180, 86400)
	if err != nil {
		return nil, nil, err
	}
	workRoot, err := filepath.Abs(filepath.Join(dataDir, "processor-work"))
	if err != nil {
		return nil, nil, fmt.Errorf("resolve processor work directory: %w", err)
	}
	processor, err := distillers.NewCodex(distillers.CodexOptions{
		Executable: os.Getenv("ASTROCYTE_CODEX_EXECUTABLE"),
		WorkRoot:   workRoot, Model: os.Getenv("ASTROCYTE_CODEX_MODEL"),
		Timeout: time.Duration(timeoutSeconds) * time.Second, AllowedSourceKeys: sourceKeys,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("configure Codex processor: %w", err)
	}
	if _, err := processor.ConfigurationID(ctx); err != nil {
		return nil, nil, fmt.Errorf("verify Codex processor configuration: %w", err)
	}
	return processor, sourceKeys, nil
}

// Extraction uses the project installation by default. Launch scripts supply
// absolute paths; direct binary launches can resolve a checkout beside the binary
// or above the current directory. A missing installation is an actionable error.
func resolveSourceReader(roots []string) (*importers.Reader, error) {
	reader := importers.NewReader(roots)
	enabled := os.Getenv("ASTROCYTE_ENABLE_SUMMARIZE")
	if enabled == "false" {
		return reader, nil
	}
	if enabled != "" && enabled != "true" {
		return nil, fmt.Errorf("ASTROCYTE_ENABLE_SUMMARIZE must be true or false")
	}
	cliPath := os.Getenv("ASTROCYTE_SUMMARIZE_CLI")
	if cliPath == "" {
		var err error
		cliPath, err = resolveProjectSummarize()
		if err != nil {
			return nil, err
		}
	}
	if !filepath.IsAbs(cliPath) {
		return nil, fmt.Errorf("ASTROCYTE_SUMMARIZE_CLI must be absolute")
	}
	// pnpm installs the CLI behind a junction. The media bridge resolves core
	// from this location, so it needs the real package directory.
	cliPath, err := filepath.EvalSymlinks(cliPath)
	if err != nil {
		return nil, fmt.Errorf("resolve summarize CLI: %w", err)
	}
	nodePath := os.Getenv("ASTROCYTE_NODE")
	if nodePath == "" {
		var err error
		nodePath, err = exec.LookPath("node")
		if err != nil {
			return nil, fmt.Errorf("find Node for ASTROCYTE_SUMMARIZE_CLI: %w", err)
		}
		nodePath, err = filepath.Abs(nodePath)
		if err != nil {
			return nil, fmt.Errorf("resolve Node: %w", err)
		}
	}
	timeoutSeconds, err := configuredPositiveInt("ASTROCYTE_JOB_TIMEOUT_SECONDS", 300, 86400)
	if err != nil {
		return nil, err
	}
	extractor, err := importers.NewSummarizeExtractorWithOptions(nodePath, cliPath, importers.SummarizeOptions{
		Version: "0.25.1", YtDlpPath: os.Getenv("ASTROCYTE_YT_DLP_PATH"),
		FFmpegPath: os.Getenv("ASTROCYTE_FFMPEG_PATH"), WhisperBinary: os.Getenv("ASTROCYTE_WHISPER_BINARY"),
		WhisperModel: os.Getenv("ASTROCYTE_WHISPER_MODEL"), Timeout: time.Duration(timeoutSeconds) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("configure summarize extraction: %w", err)
	}
	reader.Arxiv.TextExtractor = extractor
	reader.Summarize = extractor
	return reader, nil
}

func resolveProjectSummarize() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}
	starts := []string{cwd}
	if executable, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(executable))
	}
	for _, start := range starts {
		for dir := start; ; dir = filepath.Dir(dir) {
			var project struct{ Name string }
			raw, readErr := os.ReadFile(filepath.Join(dir, "package.json"))
			if readErr == nil && json.Unmarshal(raw, &project) == nil && project.Name == "astrocyte" {
				packageRoot := filepath.Join(dir, "node_modules", "@steipete", "summarize")
				var pkg struct{ Version string }
				raw, err := os.ReadFile(filepath.Join(packageRoot, "package.json"))
				if err != nil || json.Unmarshal(raw, &pkg) != nil || pkg.Version != "0.25.1" {
					return "", fmt.Errorf("project summarize 0.25.1 is missing or mismatched; run pnpm install --frozen-lockfile")
				}
				return filepath.Join(packageRoot, "dist", "cli.js"), nil
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return "", fmt.Errorf("project summarize not found; launch with pnpm dev/start from an installed checkout, or configure absolute ASTROCYTE_SUMMARIZE_CLI and ASTROCYTE_NODE")
}

// Runtime bounds are local service configuration, not client-supplied authority.
func configuredPositiveInt(name string, defaultValue, maxValue int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxValue {
		return 0, fmt.Errorf("%s must be an integer from 1 to %d", name, maxValue)
	}
	return n, nil
}

func resolveAttentionPolicy() (time.Duration, map[string]float64, error) {
	seconds, err := configuredPositiveInt("ASTROCYTE_ATTENTION_HALF_LIFE_SECONDS", 7*24*60*60, 315360000)
	if err != nil {
		return 0, nil, err
	}
	var weights map[string]float64
	if raw := os.Getenv("ASTROCYTE_ATTENTION_WEIGHTS"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &weights); err != nil || weights == nil {
			return 0, nil, fmt.Errorf("ASTROCYTE_ATTENTION_WEIGHTS must be a JSON object of event weights")
		}
		if err := attentionapp.ValidateAttentionWeights(weights); err != nil {
			return 0, nil, err
		}
	}
	return time.Duration(seconds) * time.Second, weights, nil
}
func resolveImportRoots() ([]string, error) {
	var roots []string
	for _, root := range filepath.SplitList(os.Getenv("ASTROCYTE_IMPORT_ROOTS")) {
		if !filepath.IsAbs(root) {
			return nil, fmt.Errorf("ASTROCYTE_IMPORT_ROOTS entries must be absolute directories")
		}
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("ASTROCYTE_IMPORT_ROOTS entry is not a directory")
		}
		resolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			return nil, fmt.Errorf("resolve import root: %w", err)
		}
		roots = append(roots, resolved)
	}
	return roots, nil
}
func resolveAllowedOrigins() ([]string, error) {
	webPort, err := configuredPositiveInt("ASTROCYTE_WEB_PORT", 5173, 65535)
	if err != nil {
		return nil, err
	}
	origins := []string{fmt.Sprintf("http://127.0.0.1:%d", webPort), fmt.Sprintf("http://localhost:%d", webPort)}
	for _, origin := range strings.Split(os.Getenv("ASTROCYTE_ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		u, err := url.Parse(origin)
		if err != nil || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") {
			return nil, fmt.Errorf("ASTROCYTE_ALLOWED_ORIGINS requires exact loopback HTTP origins")
		}
		origins = append(origins, origin)
	}
	return origins, nil
}
