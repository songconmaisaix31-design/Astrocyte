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

type repositoryContextConfig struct {
	EnvironmentID string
	Node          string
	SDK           string
	Script        string
	AOCI          string
	CacheRoot     string
	MaxFiles      int
	MaxBytes      int64
	MaxResults    int
	MaxDepth      int

	CleanupMaxGenerations int
}

// 每次清理最多删除 64 个 generation；可信环境可在 S4 规定的 1..1000 内配置。
const defaultRepositoryCleanupMaxGenerations = 64

// HTTP 只提交业务请求；程序、环境身份和预算由本地可信启动配置提供。
func resolveRepositoryContextConfig(ctx context.Context, dataDir string) (repositoryContextConfig, error) {
	c := repositoryContextConfig{
		EnvironmentID: os.Getenv("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID"),
		Node:          os.Getenv("ASTROCYTE_CODEGRAPH_NODE"),
		SDK:           os.Getenv("ASTROCYTE_CODEGRAPH_SDK"),
		Script:        os.Getenv("ASTROCYTE_CODEGRAPH_SCRIPT"),
		AOCI:          os.Getenv("ASTROCYTE_AOCI_EXECUTABLE"),
	}
	if c.EnvironmentID != strings.TrimSpace(c.EnvironmentID) {
		return c, fmt.Errorf("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID must be a canonical explicit identifier")
	}
	var err error
	if c.MaxFiles, err = configuredPositiveInt("ASTROCYTE_REPOSITORY_MAX_FILES", 20000, 20000); err != nil {
		return c, err
	}
	maxBytes, err := configuredPositiveInt("ASTROCYTE_REPOSITORY_MAX_BYTES", 134217728, 134217728)
	if err != nil {
		return c, err
	}
	c.MaxBytes = int64(maxBytes)
	if c.MaxResults, err = configuredPositiveInt("ASTROCYTE_REPOSITORY_MAX_RESULTS", 200, 200); err != nil {
		return c, err
	}
	if c.MaxDepth, err = configuredPositiveInt("ASTROCYTE_REPOSITORY_MAX_DEPTH", 5, 5); err != nil {
		return c, err
	}
	if c.CleanupMaxGenerations, err = configuredPositiveInt("ASTROCYTE_REPOSITORY_CLEANUP_MAX_GENERATIONS", defaultRepositoryCleanupMaxGenerations, 1000); err != nil {
		return c, err
	}
	graphConfigured := c.Node != "" || c.SDK != "" || c.Script != ""
	if graphConfigured && (c.Node == "" || c.SDK == "" || c.Script == "") {
		return c, fmt.Errorf("ASTROCYTE_CODEGRAPH_NODE, ASTROCYTE_CODEGRAPH_SDK and ASTROCYTE_CODEGRAPH_SCRIPT must be configured together")
	}
	if !filepath.IsAbs(dataDir) {
		return c, fmt.Errorf("repository context requires the actual absolute application data directory")
	}
	c.CacheRoot = filepath.Join(filepath.Clean(dataDir), "repository-context")
	// 缺少稳定环境身份时由装配注册 unavailable 能力，不启动任何外部程序。
	if c.EnvironmentID == "" {
		return c, nil
	}
	if !graphConfigured && c.AOCI == "" {
		return c, nil
	}
	runtimeRoot := filepath.Join(c.CacheRoot, "runtime")
	for _, directory := range []string{runtimeRoot, filepath.Join(runtimeRoot, "profile"), filepath.Join(runtimeRoot, "appdata"), filepath.Join(runtimeRoot, "localappdata"), filepath.Join(runtimeRoot, "temp")} {
		if err := os.MkdirAll(directory, 0700); err != nil {
			return c, fmt.Errorf("create repository tool runtime directory: %w", err)
		}
	}
	probe := func(executable string, args ...string) ([]byte, error) {
		probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		command := exec.CommandContext(probeCtx, executable, args...)
		command.Dir = runtimeRoot
		command.Env = []string{
			"SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("WINDIR"),
			"PATH=" + filepath.Dir(executable), "CODEGRAPH_TELEMETRY=0", "CODEGRAPH_NO_DOWNLOAD=1",
			"HOME=" + filepath.Join(runtimeRoot, "profile"), "USERPROFILE=" + filepath.Join(runtimeRoot, "profile"),
			"APPDATA=" + filepath.Join(runtimeRoot, "appdata"), "LOCALAPPDATA=" + filepath.Join(runtimeRoot, "localappdata"),
			"TEMP=" + filepath.Join(runtimeRoot, "temp"), "TMP=" + filepath.Join(runtimeRoot, "temp"),
		}
		output, err := command.Output()
		if err != nil {
			return nil, fmt.Errorf("verify configured repository executable %s: %w", executable, err)
		}
		return output, nil
	}
	if graphConfigured {
		for name, path := range map[string]string{"ASTROCYTE_CODEGRAPH_NODE": c.Node, "ASTROCYTE_CODEGRAPH_SCRIPT": c.Script} {
			if err := repositoryConfiguredPath(name, path, false); err != nil {
				return c, err
			}
		}
		if err := repositoryConfiguredPath("ASTROCYTE_CODEGRAPH_SDK", c.SDK, true); err != nil {
			return c, err
		}
		version, err := probe(c.Node, "--version")
		if err != nil {
			return c, err
		}
		if strings.TrimSpace(string(version)) != "v24.16.0" {
			return c, fmt.Errorf("repository CodeGraph requires Node v24.16.0")
		}
		if err := repositoryPackageVersion(filepath.Join(c.SDK, "package.json"), "@colbymchenry/codegraph"); err != nil {
			return c, err
		}
		// 让官方包实际解析其生产依赖；源码 SDK 与 npm wrapper 共用公开导出契约。
		if _, err := probe(c.Node, "-e", "const sdk=require(process.argv[1]);for(const key of ['CodeGraph','QueryBuilder','setLogger'])if(typeof sdk[key]!=='function')throw Error('codegraph_exports_unavailable');if(!sdk.silentLogger)throw Error('codegraph_logger_unavailable')", c.SDK); err != nil {
			return c, err
		}
		if _, err := probe(c.Node, "--check", c.Script); err != nil {
			return c, err
		}
	}
	if c.AOCI != "" {
		if err := repositoryConfiguredPath("ASTROCYTE_AOCI_EXECUTABLE", c.AOCI, false); err != nil {
			return c, err
		}
		version, err := probe(c.AOCI, "--version")
		if err != nil {
			return c, err
		}
		if !strings.HasPrefix(strings.TrimSpace(string(version)), "aoci version v0.1.0-rc20 (commit 45990a1dcfc4415efdc6fbfa13bc159a415f127c, built ") {
			return c, fmt.Errorf("repository AOCI requires v0.1.0-rc20 from commit 45990a1dcfc4415efdc6fbfa13bc159a415f127c")
		}
	}
	return c, nil
}

func repositoryConfiguredPath(name, path string, directory bool) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s must be an existing absolute path", name)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", name, err)
	}
	if (directory && !info.IsDir()) || (!directory && !info.Mode().IsRegular()) {
		return fmt.Errorf("%s has the wrong filesystem type", name)
	}
	return nil
}

func repositoryPackageVersion(path, expectedName string) error {
	if err := repositoryConfiguredPath(expectedName+" package.json", path, false); err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pkg struct{ Name, Version string }
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return fmt.Errorf("parse actual %s package metadata: %w", expectedName, err)
	}
	if pkg.Name != expectedName || pkg.Version != "1.0.1" {
		return fmt.Errorf("repository requires %s@1.0.1", expectedName)
	}
	return nil
}

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
