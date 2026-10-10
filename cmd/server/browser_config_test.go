package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectedBrowserConfigExplicitBoundScope(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"CONFIG", "OPENCLI_NODE", "OPENCLI_MAIN", "OPENCLI_PROFILE", "OPENCLI_SESSION", "OWNER_ID", "FOLDER_IDS"} {
		t.Setenv("ASTROCYTE_DOUYIN_"+name, "")
	}
	if c, err := resolveSelectedBrowserConfig(root); c != nil || err != nil {
		t.Fatal("absent configuration must disable the bridge", c, err)
	}
	config := selectedBrowserConfig{NodePath: filepath.Join(root, "node.exe"), CLIMain: filepath.Join(root, "opencli.mjs"), Profile: "selected-profile", Session: "selected-session", OwnerID: "selected-owner", FolderIDs: []string{"12345"}}
	for _, name := range []string{config.NodePath, config.CLIMain} {
		if err := os.WriteFile(name, []byte("configuration test; never executed"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(root, "selected-browser.json")
	write := func(value any) {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("ASTROCYTE_DOUYIN_CONFIG", file)
	write(config)
	if c, err := resolveSelectedBrowserConfig(root); err != nil || c.OwnerID != config.OwnerID || len(c.FolderIDs) != 1 || c.FolderIDs[0] != "12345" {
		t.Fatal("explicit configuration was not retained", c, err)
	}
	t.Setenv("ASTROCYTE_DOUYIN_CONFIG", "")
	for range 2 {
		if c, err := resolveSelectedBrowserConfig(root); err != nil || c.OwnerID != config.OwnerID || c.FolderIDs[0] != "12345" {
			t.Fatal("ordinary application restart lost its persisted selection", c, err)
		}
	}
	if err := os.WriteFile(file, []byte("{malformed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveSelectedBrowserConfig(root); err == nil {
		t.Fatal("malformed persisted selection silently disabled the bridge")
	}
	write(config)
	t.Setenv("ASTROCYTE_DOUYIN_CONFIG", file)
	t.Setenv("ASTROCYTE_DOUYIN_OWNER_ID", "another-owner")
	if _, err := resolveSelectedBrowserConfig(root); err == nil {
		t.Fatal("environment broadened file selection")
	}
	t.Setenv("ASTROCYTE_DOUYIN_OWNER_ID", "")
	write(map[string]any{"node_path": config.NodePath, "cookie": "unsupported-test-field"})
	if _, err := resolveSelectedBrowserConfig(root); err == nil {
		t.Fatal("credential field accepted")
	}
	config.FolderIDs = []string{"12345", "12345"}
	write(config)
	if _, err := resolveSelectedBrowserConfig(root); err == nil {
		t.Fatal("duplicate folder scope accepted")
	}
	config.FolderIDs = []string{"12345"}
	config.NodePath = "node.exe"
	write(config)
	if _, err := resolveSelectedBrowserConfig(root); err == nil {
		t.Fatal("implicit PATH executable accepted")
	}
}
