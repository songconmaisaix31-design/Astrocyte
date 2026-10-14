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

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 显式目录模式使用应用级配置；未设置时保持现有注册来源。
func resolveProjectDiscovery() (*agents.RegisteredProjects, error) {
	raw, configured := os.LookupEnv("ASTROCYTE_PROJECT_DISCOVERY_ROOTS")
	if !configured {
		return agents.NewRegisteredProjects(), nil
	}
	var roots []string
	if err := json.Unmarshal([]byte(raw), &roots); err != nil {
		return nil, fmt.Errorf("ASTROCYTE_PROJECT_DISCOVERY_ROOTS must be a JSON array: %w", err)
	}
	if roots == nil {
		return nil, fmt.Errorf("ASTROCYTE_PROJECT_DISCOVERY_ROOTS must be an explicit JSON array")
	}
	return agents.NewScopedRegisteredProjects(roots)
}

// Reader 与 Service 共用现有处理期限。
const defaultJobTimeoutSeconds = 1800

// yt-dlp 只使用明确配置或固定安装的 Python 环境。
func resolveListingPython() (string, error) {
	if configured := os.Getenv("ASTROCYTE_LISTING_PYTHON"); configured != "" {
		info, err := os.Stat(configured)
		if !filepath.IsAbs(configured) || err != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("ASTROCYTE_LISTING_PYTHON must be an existing absolute file")
		}
		return configured, nil
	}
	var candidate string
	if ytDlp := os.Getenv("ASTROCYTE_YT_DLP_PATH"); filepath.IsAbs(ytDlp) {
		candidate = filepath.Join(filepath.Dir(ytDlp), "python.exe")
	}
	if candidate == "" {
		mediaRoot := os.Getenv("ASTROCYTE_MEDIA_DIR")
		if mediaRoot == "" && os.Getenv("LOCALAPPDATA") != "" {
			mediaRoot = filepath.Join(os.Getenv("LOCALAPPDATA"), "Astrocyte", "media")
		}
		if mediaRoot != "" {
			if !filepath.IsAbs(mediaRoot) {
				return "", fmt.Errorf("ASTROCYTE_MEDIA_DIR must be absolute")
			}
			candidate = filepath.Join(mediaRoot, "yt-dlp-2026.08.19", "Scripts", "python.exe")
		}
	}
	if candidate != "" {
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
	}
	// 收藏夹继续使用公开 HTTP；上传列表缺少 Python 时返回明确配置错误。
	return "", nil
}

// 启动配置明确控制来源范围与模型处理权限。
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
	timeoutSeconds, err := configuredPositiveInt("ASTROCYTE_CODEX_TIMEOUT_SECONDS", 1800, 86400)
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

// 提取使用项目固定安装，启动脚本提供绝对路径；缺少安装时明确失败。
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
	for _, path := range []string{cliPath, nodePath} {
		info, err := os.Stat(path)
		if !filepath.IsAbs(path) || err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("summarize dependency must be an existing absolute file: %s", path)
		}
	}
	// Node 按包管理器实际规则解析 pnpm junction，与媒体处理保持一致。
	resolveCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resolved, err := exec.CommandContext(resolveCtx, nodePath, "-e", "process.stdout.write(require('node:fs').realpathSync(process.argv[1]))", cliPath).Output()
	if err != nil {
		return nil, fmt.Errorf("resolve summarize CLI with Node: %w", err)
	}
	cliPath = string(resolved)
	timeoutSeconds, err := configuredPositiveInt("ASTROCYTE_JOB_TIMEOUT_SECONDS", defaultJobTimeoutSeconds, 86400)
	if err != nil {
		return nil, err
	}
	extractor, err := importers.NewSummarizeExtractorWithOptions(nodePath, cliPath, importers.SummarizeOptions{
		Version: "0.25.1", YtDlpPath: os.Getenv("ASTROCYTE_YT_DLP_PATH"),
		FFmpegPath: os.Getenv("ASTROCYTE_FFMPEG_PATH"), WhisperBinary: os.Getenv("ASTROCYTE_WHISPER_BINARY"),
		WhisperModel: os.Getenv("ASTROCYTE_WHISPER_MODEL"), UVXPath: os.Getenv("ASTROCYTE_UVX_PATH"),
		ProxyURL: os.Getenv("ASTROCYTE_SUMMARIZE_PROXY_URL"),
		Timeout:  time.Duration(timeoutSeconds) * time.Second,
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
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve server executable: %w", err)
	}
	starts = append(starts, filepath.Dir(executable))
	for _, start := range starts {
		for dir := start; ; dir = filepath.Dir(dir) {
			var project struct{ Name string }
			raw, readErr := os.ReadFile(filepath.Join(dir, "package.json"))
			if readErr != nil && !os.IsNotExist(readErr) {
				return "", fmt.Errorf("read project package: %w", readErr)
			}
			if readErr == nil {
				if err := json.Unmarshal(raw, &project); err != nil {
					return "", fmt.Errorf("parse project package: %w", err)
				}
			}
			if readErr == nil && project.Name == "astrocyte" {
				packageRoot := filepath.Join(dir, "node_modules", "@steipete", "summarize")
				var pkg struct{ Version string }
				raw, err := os.ReadFile(filepath.Join(packageRoot, "package.json"))
				if err != nil {
					return "", fmt.Errorf("read summarize package: %w", err)
				}
				if err := json.Unmarshal(raw, &pkg); err != nil {
					return "", fmt.Errorf("parse summarize package: %w", err)
				}
				if pkg.Version != "0.25.1" {
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

// 运行期限来自本地服务配置。
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
