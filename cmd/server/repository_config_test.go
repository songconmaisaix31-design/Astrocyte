package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repositoryConfigDirectory(t *testing.T) string {
	t.Helper()
	base, err := filepath.Abs(filepath.Join("..", "..", ".verification"))
	if err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(base, "repository-config-actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("actual owned repository configuration directory: %s", directory)
	return directory
}

func clearRepositoryConfig(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"ASTROCYTE_REPOSITORY_ENVIRONMENT_ID", "ASTROCYTE_CODEGRAPH_NODE", "ASTROCYTE_CODEGRAPH_SDK", "ASTROCYTE_CODEGRAPH_SCRIPT", "ASTROCYTE_AOCI_EXECUTABLE",
		"ASTROCYTE_REPOSITORY_MAX_FILES", "ASTROCYTE_REPOSITORY_MAX_BYTES", "ASTROCYTE_REPOSITORY_MAX_RESULTS", "ASTROCYTE_REPOSITORY_MAX_DEPTH",
		"ASTROCYTE_REPOSITORY_CLEANUP_MAX_GENERATIONS",
	} {
		t.Setenv(name, "")
	}
}

// 这里只核实明确配置值；环境身份输入不冒称任何实际仓库或模型来源。
func TestRepositoryContextConfigExplicitAndUnavailable(t *testing.T) {
	clearRepositoryConfig(t)
	directory := repositoryConfigDirectory(t)
	c, err := resolveRepositoryContextConfig(context.Background(), directory)
	if err != nil || c.EnvironmentID != "" || c.Node != "" || c.SDK != "" || c.Script != "" || c.AOCI != "" || c.CacheRoot != filepath.Join(directory, "repository-context") {
		t.Fatalf("unconfigured capabilities changed: %+v %v", c, err)
	}
	if c.MaxFiles != 20000 || c.MaxBytes != 134217728 || c.MaxResults != 200 || c.MaxDepth != 5 {
		t.Fatalf("trusted budget bounds changed: %+v", c)
	}
	if _, err := os.Stat(c.CacheRoot); !os.IsNotExist(err) {
		t.Fatal("unconfigured capability created a runtime or returned an unexpected filesystem error", err)
	}
	t.Setenv("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID", "explicit-rule-input")
	for _, name := range []string{"ASTROCYTE_REPOSITORY_MAX_FILES", "ASTROCYTE_REPOSITORY_MAX_BYTES", "ASTROCYTE_REPOSITORY_MAX_RESULTS", "ASTROCYTE_REPOSITORY_MAX_DEPTH"} {
		t.Setenv(name, "1")
	}
	for range 2 {
		c, err := resolveRepositoryContextConfig(context.Background(), directory)
		if err != nil || c.EnvironmentID != "explicit-rule-input" || c.MaxFiles != 1 || c.MaxBytes != 1 || c.MaxResults != 1 || c.MaxDepth != 1 {
			t.Fatalf("explicit stable Git-only configuration rejected: %+v %v", c, err)
		}
	}
}

func TestRepositoryContextConfigRejectsAmbiguousConfiguration(t *testing.T) {
	directory := repositoryConfigDirectory(t)
	for name, value := range map[string]string{
		"ASTROCYTE_REPOSITORY_ENVIRONMENT_ID": " padded ",
		"ASTROCYTE_CODEGRAPH_NODE":            "relative-node",
		"ASTROCYTE_CODEGRAPH_SDK":             "relative-sdk",
		"ASTROCYTE_CODEGRAPH_SCRIPT":          "relative-script",
		"ASTROCYTE_REPOSITORY_MAX_FILES":      "20001",
		"ASTROCYTE_REPOSITORY_MAX_BYTES":      "134217729",
		"ASTROCYTE_REPOSITORY_MAX_RESULTS":    "0",
		"ASTROCYTE_REPOSITORY_MAX_DEPTH":      "6",
	} {
		t.Run(name, func(t *testing.T) {
			clearRepositoryConfig(t)
			t.Setenv(name, value)
			if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil {
				t.Fatal("ambiguous or out-of-bounds trusted configuration was accepted")
			}
		})
	}
	clearRepositoryConfig(t)
	if _, err := resolveRepositoryContextConfig(context.Background(), "relative-data"); err == nil {
		t.Fatal("relative application data directory was accepted")
	}
}

func TestRepositoryContextConfigRejectsActualUnavailableDependencies(t *testing.T) {
	clearRepositoryConfig(t)
	directory := repositoryConfigDirectory(t)
	t.Setenv("ASTROCYTE_REPOSITORY_ENVIRONMENT_ID", "explicit-rule-input")
	missing := filepath.Join(directory, "missing-executable")
	t.Setenv("ASTROCYTE_AOCI_EXECUTABLE", missing)
	if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil {
		t.Fatal("missing actual external AOCI was accepted")
	}
	t.Setenv("ASTROCYTE_AOCI_EXECUTABLE", "")
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASTROCYTE_CODEGRAPH_NODE", node)
	t.Setenv("ASTROCYTE_CODEGRAPH_SDK", root)
	t.Setenv("ASTROCYTE_CODEGRAPH_SCRIPT", filepath.Join(root, "scripts", "check-dependencies.mjs"))
	// 实际 Astrocyte 包不是官方 SDK；保持其 package.json 原件，核实版本拒绝。
	if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil {
		t.Fatal("the actual application package was accepted as the CodeGraph SDK")
	}
	t.Setenv("ASTROCYTE_CODEGRAPH_NODE", missing)
	if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil {
		t.Fatal("missing actual Node executable was accepted")
	}
}

func TestRepositoryCleanupBatchTrustedConfiguration(t *testing.T) {
	clearRepositoryConfig(t)
	directory := repositoryConfigDirectory(t)
	c, err := resolveRepositoryContextConfig(context.Background(), directory)
	if err != nil || c.CleanupMaxGenerations != 64 {
		t.Fatal("program-defined repository cleanup batch changed", c.CleanupMaxGenerations, err)
	}
	for _, input := range []struct {
		value string
		want  int
	}{{"1", 1}, {"64", 64}, {"1000", 1000}} {
		t.Setenv("ASTROCYTE_REPOSITORY_CLEANUP_MAX_GENERATIONS", input.value)
		c, err := resolveRepositoryContextConfig(context.Background(), directory)
		if err != nil || c.CleanupMaxGenerations != input.want {
			t.Fatal("explicit cleanup batch changed or was rejected", input.value, c.CleanupMaxGenerations, err)
		}
	}
	for _, value := range []string{"0", "-1", "1001", "1.5", "unavailable"} {
		t.Setenv("ASTROCYTE_REPOSITORY_CLEANUP_MAX_GENERATIONS", value)
		if _, err := resolveRepositoryContextConfig(context.Background(), directory); err == nil {
			t.Fatal("invalid trusted cleanup batch accepted", value)
		}
	}
}
