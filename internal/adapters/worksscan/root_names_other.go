//go:build !windows

package worksscan

import "os"

func rootRedirect(info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }

func expandRootNames(path string) (string, error) { return path, nil }
