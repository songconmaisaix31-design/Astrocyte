//go:build repository_context_live

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 集成检查必须显式传入实际安装路径，缺少依赖立即失败，不替代程序或元数据。
func requireActualCodeGraphConfig(t *testing.T) string {
	t.Helper()
	clearRepositoryConfig(t)
	for _, name := range []string{"NODE", "SDK", "SCRIPT"} {
		path := os.Getenv("ASTROCYTE_VERIFY_CODEGRAPH_" + name)
		if path == "" {
			t.Fatal("actual installed dependency is required: ASTROCYTE_VERIFY_CODEGRAPH_" + name)
		}
		t.Setenv("ASTROCYTE_CODEGRAPH_"+name, path)
	}
	t.Setenv("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID", "explicit-installed-tools-validation")
	return repositoryConfigDirectory(t)
}

func TestRepositoryContextConfigActualInstalledTools(t *testing.T) {
	directory := requireActualCodeGraphConfig(t)
	aoci := os.Getenv("ASTROCYTE_VERIFY_AOCI_EXECUTABLE")
	if aoci == "" {
		t.Fatal("actual external AOCI is required: ASTROCYTE_VERIFY_AOCI_EXECUTABLE")
	}
	t.Setenv("ASTROCYTE_AOCI_EXECUTABLE", aoci)
	c, err := resolveRepositoryContextConfig(context.Background(), directory)
	if err != nil {
		t.Fatal("actual trusted dependency startup validation failed", err)
	}
	if c.Node != os.Getenv("ASTROCYTE_CODEGRAPH_NODE") || c.SDK != os.Getenv("ASTROCYTE_CODEGRAPH_SDK") || c.Script != os.Getenv("ASTROCYTE_CODEGRAPH_SCRIPT") || c.AOCI != aoci || c.EnvironmentID != "explicit-installed-tools-validation" {
		t.Fatalf("actual startup changed explicitly provided dependency identities: %+v", c)
	}
	t.Log("Actual Node, official SDK exports, bridge syntax and external AOCI version passed; model, indexing and HTTP operations were not run")
}

func TestRepositoryContextConfigActualNodeSyntaxRefusal(t *testing.T) {
	directory := requireActualCodeGraphConfig(t)
	// 原 Go 源文件触发实际 Node 格式拒绝，检查阶段与原错误都必须保留。
	invalidScript, err := filepath.Abs("config.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASTROCYTE_CODEGRAPH_SCRIPT", invalidScript)
	if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil || !strings.Contains(err.Error(), "CodeGraph bridge syntax") || !strings.Contains(err.Error(), "ERR_UNKNOWN_FILE_EXTENSION") {
		t.Fatal("actual Node syntax refusal lost its stage or original error", err)
	}
}

func TestVideoSourceConfigActualInstalledTools(t *testing.T) {
	path := os.Getenv("ASTROCYTE_VERIFY_VIDEO_CONFIG")
	if !filepath.IsAbs(path) {
		t.Fatal("select the absolute path of the actual public media configuration")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]string
	if err := json.Unmarshal(raw, &settings); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"ASTROCYTE_ENABLE_SUMMARIZE", "ASTROCYTE_NODE", "ASTROCYTE_MEDIA_DIR",
		"ASTROCYTE_SUMMARIZE_CLI", "ASTROCYTE_YT_DLP_PATH", "ASTROCYTE_LISTING_PYTHON",
		"ASTROCYTE_FFMPEG_PATH", "ASTROCYTE_WHISPER_BINARY", "ASTROCYTE_WHISPER_MODEL",
		"ASTROCYTE_UVX_PATH", "ASTROCYTE_SUMMARIZE_PROXY_URL", "ASTROCYTE_JOB_TIMEOUT_SECONDS",
	} {
		if settings[key] == "" {
			t.Fatal("actual explicit configuration is required", key)
		}
		t.Setenv(key, settings[key])
	}
	// 构造与生产启动相同的 Reader，只执行 Node 真实路径解析。
	reader, err := resolveSourceReader(nil)
	if err != nil || reader == nil || reader.Summarize == nil || reader.Arxiv.TextExtractor != reader.Summarize {
		t.Fatal("actual installed source reader configuration failed", err)
	}
	python, err := resolveListingPython()
	if err != nil || python != settings["ASTROCYTE_LISTING_PYTHON"] {
		t.Fatal("actual installed listing Python identity changed", err)
	}
	t.Setenv("ASTROCYTE_ENABLE_SUMMARIZE", "false")
	disabled, err := resolveSourceReader(nil)
	if err != nil || disabled == nil || disabled.Summarize != nil || disabled.Arxiv.TextExtractor != nil {
		t.Fatal("disabled extraction configuration changed", err)
	}
	t.Log("Actual production reader and listing configuration passed; extraction, ASR, model and shared service operations were not run")
}
