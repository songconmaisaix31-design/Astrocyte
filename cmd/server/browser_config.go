package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Explicit host selection only. No cookies, credentials, browser profile paths,
// arbitrary expressions, or provider configuration are accepted.
type selectedBrowserConfig struct {
	NodePath  string   `json:"node_path"`
	CLIMain   string   `json:"cli_main"`
	Profile   string   `json:"profile"`
	Session   string   `json:"session"`
	OwnerID   string   `json:"owner_id"`
	FolderIDs []string `json:"folder_ids"`
}

var browserSelectionName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)
var browserOwnerID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,200}$`)
var browserFolderID = regexp.MustCompile(`^[0-9]{1,30}$`)

func resolveSelectedBrowserConfig() (*selectedBrowserConfig, error) {
	c := selectedBrowserConfig{
		NodePath: os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_NODE"), CLIMain: os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_MAIN"),
		Profile: os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_PROFILE"), Session: os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_SESSION"),
		OwnerID: os.Getenv("ASTROCYTE_DOUYIN_OWNER_ID"),
	}
	folders := os.Getenv("ASTROCYTE_DOUYIN_FOLDER_IDS")
	file := os.Getenv("ASTROCYTE_DOUYIN_CONFIG")
	hasEnv := c.NodePath != "" || c.CLIMain != "" || c.Profile != "" || c.Session != "" || c.OwnerID != "" || folders != ""
	if file == "" && !hasEnv {
		return nil, nil
	}
	if file != "" {
		if hasEnv || !filepath.IsAbs(file) {
			return nil, fmt.Errorf("ASTROCYTE_DOUYIN_CONFIG requires one explicit absolute config path without mixed browser environment options")
		}
		info, err := os.Stat(file)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("selected browser config must be one explicit regular JSON file")
		}
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open selected browser config: %w", err)
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, 16*1024+1))
		if err != nil || len(data) > 16*1024 {
			return nil, fmt.Errorf("selected browser config exceeds its bound or cannot be read")
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&c) != nil || decoder.Decode(new(any)) != io.EOF {
			return nil, fmt.Errorf("selected browser config must contain one supported noncredential JSON object")
		}
	} else if err := json.Unmarshal([]byte(folders), &c.FolderIDs); err != nil {
		return nil, fmt.Errorf("ASTROCYTE_DOUYIN_FOLDER_IDS must be an explicit JSON array of selected folder IDs")
	}
	for _, path := range []string{c.NodePath, c.CLIMain} {
		if !filepath.IsAbs(path) {
			return nil, fmt.Errorf("selected browser Node and OpenCLI entrypoint must be explicit absolute files")
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("selected browser Node or OpenCLI entrypoint is unavailable")
		}
	}
	if ext := strings.ToLower(filepath.Ext(c.NodePath)); ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		return nil, fmt.Errorf("selected browser requires a native Node executable")
	}
	if !browserSelectionName.MatchString(c.Profile) || !browserSelectionName.MatchString(c.Session) || !browserOwnerID.MatchString(c.OwnerID) || len(c.FolderIDs) == 0 || len(c.FolderIDs) > 100 {
		return nil, fmt.Errorf("selected browser requires a bound profile/session, stable owner and one to100 selected folders")
	}
	seen := map[string]bool{}
	for _, id := range c.FolderIDs {
		if !browserFolderID.MatchString(id) || seen[id] {
			return nil, fmt.Errorf("selected browser folder IDs must be unique stable decimal IDs")
		}
		seen[id] = true
	}
	return &c, nil
}
