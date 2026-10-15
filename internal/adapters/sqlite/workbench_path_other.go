//go:build !windows

package sqlite

import (
	"os"
	"path/filepath"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

func workbenchPathUnredirected(path string) error {
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return workbench.ErrUnavailable
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}
