//go:build !windows

package events

import (
	"os"
	"path/filepath"
)

func expandReplayNames(path string) (string, error) { return filepath.Clean(path), nil }

func replayNamesEqual(a, b string) bool { return filepath.Clean(a) == filepath.Clean(b) }

func replayPathRedirect(_ string, info os.FileInfo) (bool, error) {
	return info.Mode()&os.ModeSymlink != 0, nil
}
