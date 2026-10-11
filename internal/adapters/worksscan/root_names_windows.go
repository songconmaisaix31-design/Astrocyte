package worksscan

import (
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

func rootRedirect(info os.FileInfo) bool {
	attributes, ok := info.Sys().(*syscall.Win32FileAttributeData)
	// Junctions may be ModeIrregular rather than ModeSymlink in Go. Fail closed
	// for root/ancestor reparse points instead of following a directory redirect.
	return info.Mode()&os.ModeSymlink != 0 || !ok || attributes.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}

// GetLongPathName expands directory-entry aliases without resolving a symlink
// target. EvalSymlinks remains a separate check for actual redirection.
func expandRootNames(path string) (string, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	buf := make([]uint16, 260)
	for {
		n, err := windows.GetLongPathName(p, &buf[0], uint32(len(buf)))
		if err != nil {
			return "", err
		}
		if n < uint32(len(buf)) {
			return filepath.Clean(windows.UTF16ToString(buf[:n])), nil
		}
		buf = make([]uint16, n+1)
	}
}
