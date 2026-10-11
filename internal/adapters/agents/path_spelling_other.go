//go:build !windows

package agents

import (
	"os"
	"path/filepath"
)

func pathSpelling(path string) string { return filepath.Clean(path) }

func pathIsRedirect(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}
