package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

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
