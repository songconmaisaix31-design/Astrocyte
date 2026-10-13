package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func relayEnabled() (bool, error) {
	switch os.Getenv("ASTROCYTE_RELAY") {
	case "", "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, errors.New("ASTROCYTE_RELAY must be 0 or 1")
	}
}
func relayRoot() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if !filepath.IsAbs(local) {
		return "", errors.New("relay requires absolute LOCALAPPDATA")
	}
	expected := filepath.Join(local, "astrocyte-workbench", "v0.2-access")
	resolved, err := workbenchCanonical(expected)
	if err != nil || !strings.EqualFold(filepath.Clean(expected), filepath.Clean(resolved)) {
		return "", errors.New("relay access root must not redirect")
	}
	return resolved, nil
}
func relayInside(root, path string) bool {
	rel, err := filepath.Rel(strings.ToLower(root), strings.ToLower(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// Access mode always keeps the existing application's database under v0.2.
// The snapshot Reader is still configured independently against v0.1 publication.
func resolveRelayDataDir(previous string) (string, error) {
	enabled, err := relayEnabled()
	if err != nil || !enabled {
		return previous, err
	}
	root, err := relayRoot()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, "app-state")
	if os.Getenv("ASTROCYTE_DATA_DIR") != "" {
		path = previous
	}
	path, err = workbenchCanonical(path)
	if err != nil || !relayInside(root, path) {
		return "", errors.New("relay application state must stay under local access root")
	}
	return path, nil
}

func relayConfiguration() (string, []domain.AgentSpec, time.Duration, error) {
	root, err := relayRoot()
	if err != nil {
		return "", nil, 0, err
	}
	window := time.Duration(0)
	if raw := os.Getenv("ASTROCYTE_RELAY_LIVE_SECONDS"); raw != "" {
		seconds, e := strconv.Atoi(raw)
		if e != nil || seconds < 1 || seconds > 86400 {
			return "", nil, 0, errors.New("relay live window must be explicit 1..86400 seconds")
		}
		window = time.Duration(seconds) * time.Second
	}
	// No default 60s and no inferred Agent command from inventory. Registration
	// is trusted local data, loaded once; HTTP cannot edit it or supply a command.
	registryPath, err := workbenchCanonical(filepath.Join(root, "agents.json"))
	if err != nil || !relayInside(root, registryPath) {
		return "", nil, 0, errors.New("relay registry must stay under local access root")
	}
	f, err := os.Open(registryPath)
	if os.IsNotExist(err) {
		return root, []domain.AgentSpec{}, window, nil
	}
	if err != nil {
		return "", nil, 0, errors.New("relay registry unavailable")
	}
	defer f.Close()
	decoder := json.NewDecoder(io.LimitReader(f, 1<<20))
	decoder.DisallowUnknownFields()
	var agents []domain.AgentSpec
	if decoder.Decode(&agents) != nil || decoder.Decode(new(any)) != io.EOF {
		return "", nil, 0, errors.New("invalid relay registration data")
	}
	if agents == nil {
		agents = []domain.AgentSpec{}
	}
	seen := map[string]bool{}
	for _, a := range agents {
		if a.Name == "" || seen[a.Name] || len(a.LaunchCmd) == 0 || (a.DriveMode != "structured" && a.DriveMode != "terminal" && a.DriveMode != "both") {
			return "", nil, 0, errors.New("invalid relay registration data")
		}
		if a.InputMode != "" && a.InputMode != "line" && a.InputMode != "stdin_eof" {
			return "", nil, 0, errors.New("invalid relay registration input mode")
		}
		seen[a.Name] = true
	}
	return root, agents, window, nil
}
