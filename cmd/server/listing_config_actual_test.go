package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListingPythonExplicitInputsFailAtConfiguration(t *testing.T) {
	for _, test := range []struct {
		name, variable, value, message string
		missing                        bool
	}{
		{"relative_python", "ASTROCYTE_LISTING_PYTHON", "python.exe", "must be an existing absolute file", false},
		{"missing_python", "ASTROCYTE_LISTING_PYTHON", filepath.Join(t.TempDir(), "python.exe"), "read ASTROCYTE_LISTING_PYTHON", true},
		{"python_directory", "ASTROCYTE_LISTING_PYTHON", t.TempDir(), "must be an existing absolute file", false},
		{"relative_yt_dlp", "ASTROCYTE_YT_DLP_PATH", "yt-dlp.exe", "must be absolute", false},
		{"missing_sibling_python", "ASTROCYTE_YT_DLP_PATH", filepath.Join(t.TempDir(), "yt-dlp.exe"), "read configured listing Python", true},
		{"relative_media", "ASTROCYTE_MEDIA_DIR", "media", "must be absolute", false},
		{"missing_media_python", "ASTROCYTE_MEDIA_DIR", t.TempDir(), "read configured listing Python", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, variable := range []string{"ASTROCYTE_LISTING_PYTHON", "ASTROCYTE_YT_DLP_PATH", "ASTROCYTE_MEDIA_DIR", "LOCALAPPDATA"} {
				t.Setenv(variable, "")
			}
			t.Setenv(test.variable, test.value)
			got, err := resolveListingPython()
			if got != "" || err == nil || !strings.Contains(err.Error(), test.message) || (test.missing && !errors.Is(err, os.ErrNotExist)) {
				t.Fatal("explicit invalid configuration did not preserve its actual error", got, err)
			}
		})
	}
}

func TestListingPythonUnconfiguredCapabilityRemainsAbsent(t *testing.T) {
	for _, variable := range []string{"ASTROCYTE_LISTING_PYTHON", "ASTROCYTE_YT_DLP_PATH", "ASTROCYTE_MEDIA_DIR"} {
		t.Setenv(variable, "")
	}
	for _, localAppData := range []string{"", t.TempDir()} {
		t.Setenv("LOCALAPPDATA", localAppData)
		if got, err := resolveListingPython(); err != nil || got != "" {
			t.Fatal("unconfigured optional listing capability changed", got, err)
		}
	}
}
