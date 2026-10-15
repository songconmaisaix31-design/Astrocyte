//go:build repository_context_live

package main

import (
	"context"
	"os"
	"testing"
)

// 集成检查必须显式传入实际安装路径，缺少依赖立即失败，不替代程序或元数据。
func TestRepositoryContextConfigActualInstalledTools(t *testing.T) {
	clearRepositoryConfig(t)
	for _, name := range []string{"NODE", "SDK", "SCRIPT"} {
		path := os.Getenv("ASTROCYTE_VERIFY_CODEGRAPH_" + name)
		if path == "" {
			t.Fatal("actual installed dependency is required: ASTROCYTE_VERIFY_CODEGRAPH_" + name)
		}
		t.Setenv("ASTROCYTE_CODEGRAPH_"+name, path)
	}
	aoci := os.Getenv("ASTROCYTE_VERIFY_AOCI_EXECUTABLE")
	if aoci == "" {
		t.Fatal("actual external AOCI is required: ASTROCYTE_VERIFY_AOCI_EXECUTABLE")
	}
	t.Setenv("ASTROCYTE_AOCI_EXECUTABLE", aoci)
	t.Setenv("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID", "explicit-installed-tools-validation")
	directory := repositoryConfigDirectory(t)
	c, err := resolveRepositoryContextConfig(context.Background(), directory)
	if err != nil {
		t.Fatal("actual trusted dependency startup validation failed", err)
	}
	if c.Node != os.Getenv("ASTROCYTE_CODEGRAPH_NODE") || c.SDK != os.Getenv("ASTROCYTE_CODEGRAPH_SDK") || c.Script != os.Getenv("ASTROCYTE_CODEGRAPH_SCRIPT") || c.AOCI != aoci || c.EnvironmentID != "explicit-installed-tools-validation" {
		t.Fatalf("actual startup changed explicitly provided dependency identities: %+v", c)
	}
	t.Log("Actual Node, official SDK exports, bridge syntax and external AOCI version passed; model, indexing and HTTP operations were not run")
}
