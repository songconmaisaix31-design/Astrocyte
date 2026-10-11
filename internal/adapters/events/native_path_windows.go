package events

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// GetLongPathName expands directory-entry spelling, not a link target.
func expandReplayNames(path string) (string, error) {
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

func replayNamesEqual(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func replayPathRedirect(path string, info os.FileInfo) (bool, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attributes, err := windows.GetFileAttributes(p)
	// Go can leave a terminal junction unchanged in EvalSymlinks and mark it
	// ModeIrregular. Inspect Windows reparse attributes separately, fail closed.
	return info.Mode()&os.ModeSymlink != 0 || attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0, err
}
