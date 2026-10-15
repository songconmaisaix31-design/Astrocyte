//go:build repository_context_live

package main

import (
	"context"
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
